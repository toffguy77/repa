package groups

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/eligibility"
	"github.com/repa-app/repa/internal/schedule"
	"github.com/rs/zerolog/log"
)

var (
	ErrGroupNotFound   = errors.New("group not found")
	ErrNotMember       = errors.New("user is not a member of this group")
	ErrAlreadyMember   = errors.New("user is already a member")
	ErrGroupLimitUser  = errors.New("user has reached maximum number of groups")
	ErrGroupLimitSize  = errors.New("group has reached maximum number of members")
	ErrNotAdmin        = errors.New("only admin can perform this action")
	ErrInvalidName     = errors.New("group name must be 3-40 characters")
	ErrNoCategories    = errors.New("at least one category is required")
	ErrInvalidCategory = errors.New("invalid question category")
	ErrRomanceBlocked  = errors.New("ROMANCE category is not available for users under 18")
	// ErrNoKindCategories guards the combination that would otherwise produce an empty season: a
	// kind-only group whose categories are all edgy (HOT, SECRETS). Refused rather than silently
	// relaxed — a group with no questions cannot vote, and the setting the member chose would have
	// been ignored.
	ErrNoKindCategories = errors.New("the chosen categories have no questions for a kind-only group")
	// ErrInviteCodeExhausted means the generator could not find a free code. With ~887M
	// codes this signals a fault, not a full code space, so it fails loudly.
	ErrInviteCodeExhausted = errors.New("could not generate a unique invite code")
	// ErrGroupBanned means this person was removed from the group, or left it permanently.
	ErrGroupBanned = errors.New("you cannot rejoin this group")
)

const (
	MaxGroupsPerUser   = 10
	MaxMembersPerGroup = 50
	MinGroupName       = 3
	MaxGroupName       = 40
	MinSeasonQuestions = 5
	MaxSeasonQuestions = 10
)

var validCategories = map[string]db.QuestionCategory{
	"HOT":     db.QuestionCategoryHOT,
	"FUNNY":   db.QuestionCategoryFUNNY,
	"SECRETS": db.QuestionCategorySECRETS,
	"SKILLS":  db.QuestionCategorySKILLS,
	"ROMANCE": db.QuestionCategoryROMANCE,
	"STUDY":   db.QuestionCategorySTUDY,
}

type Service struct {
	queries *db.Queries
	sqlDB   *sql.DB
}

func NewService(queries *db.Queries, sqlDB *sql.DB) *Service {
	return &Service{queries: queries, sqlDB: sqlDB}
}

func (s *Service) GetUser(ctx context.Context, userID string) (*db.User, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type CreateGroupParams struct {
	UserID           string
	Name             string
	Categories       []string
	TelegramUsername string
	UserBirthYear    sql.NullInt32

	// KindOnly restricts the group's seasons to warm and neutral questions. Nil means "use the
	// age-based default" — an explicit value always wins, including an adult turning it on and a minor
	// turning it off. The default is a starting point, not a restriction they cannot see or change.
	KindOnly *bool
}

// defaultKindOnly derives the setting from the creator's age, reusing the same birth-year signal the
// ROMANCE restriction already uses so there is one notion of "under 18" in the product.
func defaultKindOnly(explicit *bool, birthYear sql.NullInt32) bool {
	if explicit != nil {
		return *explicit
	}
	if !birthYear.Valid {
		// Unknown age is treated as under 18, matching how ROMANCE is handled.
		return true
	}
	return time.Now().Year()-int(birthYear.Int32) < 18
}

type CreateGroupResult struct {
	Group     db.Group
	InviteURL string
}

func (s *Service) CreateGroup(ctx context.Context, p CreateGroupParams) (*CreateGroupResult, error) {
	if len(p.Name) < MinGroupName || len(p.Name) > MaxGroupName {
		return nil, ErrInvalidName
	}
	if len(p.Categories) == 0 {
		return nil, ErrNoCategories
	}

	categories, err := validateCategories(p.Categories, p.UserBirthYear)
	if err != nil {
		return nil, err
	}

	count, err := s.queries.CountUserGroups(ctx, p.UserID)
	if err != nil {
		return nil, err
	}
	if count >= MaxGroupsPerUser {
		return nil, ErrGroupLimitUser
	}

	groupID := uuid.New().String()
	kindOnly := defaultKindOnly(p.KindOnly, p.UserBirthYear)

	if kindOnly {
		if err := s.assertKindCategories(ctx, categories); err != nil {
			return nil, err
		}
	}

	inviteCode, err := s.freshInviteCode(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	group, err := qtx.CreateGroup(ctx, db.CreateGroupParams{
		ID:                   groupID,
		Name:                 p.Name,
		InviteCode:           inviteCode,
		AdminID:              p.UserID,
		TelegramChatUsername: sql.NullString{String: p.TelegramUsername, Valid: p.TelegramUsername != ""},
		Categories:           categoryStrings(categories),
		KindOnly:             kindOnly,
	})
	if err != nil {
		return nil, err
	}

	_, err = qtx.AddGroupMember(ctx, db.AddGroupMemberParams{
		ID:      uuid.New().String(),
		UserID:  p.UserID,
		GroupID: groupID,
		// The founder did not arrive through an invite; recording that explicitly keeps the
		// funnel honest.
		JoinSource: db.JoinSourceUNKNOWN,
	})
	if err != nil {
		return nil, err
	}

	if err := s.createFirstSeasonTx(ctx, qtx, groupID, categories, kindOnly); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateGroupResult{
		Group:     group,
		InviteURL: InviteURL(inviteCode),
	}, nil
}

func (s *Service) createFirstSeasonTx(ctx context.Context, qtx *db.Queries, groupID string, categories []db.QuestionCategory, kindOnly bool) error {
	startsAt, revealAt, endsAt := schedule.KickoffSeason(time.Now())

	season, err := qtx.CreateSeason(ctx, db.CreateSeasonParams{
		ID:       uuid.New().String(),
		GroupID:  groupID,
		Number:   1,
		StartsAt: startsAt,
		RevealAt: revealAt,
		EndsAt:   endsAt,
		Kind:     db.SeasonKindKICKOFF,
	})
	if err != nil {
		return err
	}

	return s.selectAndAssignQuestionsTx(ctx, qtx, season.ID, groupID, categories, 1, kindOnly)
}

func (s *Service) selectAndAssignQuestionsTx(ctx context.Context, qtx *db.Queries, seasonID, groupID string, categories []db.QuestionCategory, memberCount int, kindOnly bool) error {
	questionCount := memberCount * 2
	if questionCount < MinSeasonQuestions {
		questionCount = MinSeasonQuestions
	}
	if questionCount > MaxSeasonQuestions {
		questionCount = MaxSeasonQuestions
	}

	questions, err := qtx.GetRandomSystemQuestionsByCategories(ctx, db.GetRandomSystemQuestionsByCategoriesParams{
		Column1: categories,
		GroupID: groupID,
		Limit:   int32(questionCount),
		// A kind-only group never draws an edgy question.
		Column4: kindOnly,
	})
	if err != nil {
		return err
	}

	for i, q := range questions {
		_, err := qtx.AddSeasonQuestion(ctx, db.AddSeasonQuestionParams{
			ID:         uuid.New().String(),
			SeasonID:   seasonID,
			QuestionID: q.ID,
			Ord:        int32(i),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type GroupListItem struct {
	Group        db.Group
	MemberCount  int64
	ActiveSeason *db.Season
	VotedCount   int64
	UserVoted    bool
	// RevealStatus explains what the pending Reveal is waiting for. Nil when the group
	// has no open season.
	RevealStatus *eligibility.Status
}

func (s *Service) ListUserGroups(ctx context.Context, userID string) ([]GroupListItem, error) {
	rows, err := s.queries.GetUserGroupsWithStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]GroupListItem, 0, len(rows))
	for _, r := range rows {
		group := db.Group{
			ID:                    r.ID,
			Name:                  r.Name,
			InviteCode:            r.InviteCode,
			AdminID:               r.AdminID,
			TelegramChatID:        r.TelegramChatID,
			TelegramChatUsername:  r.TelegramChatUsername,
			TelegramConnectCode:   r.TelegramConnectCode,
			TelegramConnectExpiry: r.TelegramConnectExpiry,
			CreatedAt:             r.CreatedAt,
			Categories:            r.Categories,
		}
		item := GroupListItem{
			Group:       group,
			MemberCount: r.MemberCount,
		}

		if r.ActiveSeasonID.Valid {
			season := db.Season{
				ID:            r.ActiveSeasonID.String,
				GroupID:       r.ID,
				Number:        r.ActiveSeasonNumber.Int32,
				Status:        r.ActiveSeasonStatus.SeasonStatus,
				StartsAt:      r.ActiveSeasonStartsAt.Time,
				RevealAt:      r.ActiveSeasonRevealAt.Time,
				EndsAt:        r.ActiveSeasonEndsAt.Time,
				Kind:          r.ActiveSeasonKind.SeasonKind,
				PostponeCount: r.ActiveSeasonPostponeCount.Int32,
			}
			item.ActiveSeason = &season
			item.VotedCount = r.VotedCount
			item.UserVoted = r.UserVoteCount > 0
			status := eligibility.Evaluate(r.MemberCount, r.VotedCount)
			item.RevealStatus = &status
		}

		items = append(items, item)
	}

	return items, nil
}

type GroupDetail struct {
	// RevealStatus explains what the pending Reveal is waiting for. Nil when the group
	// has no open season.
	RevealStatus *eligibility.Status

	// EffectiveMemberCount is how many members this viewer can actually be rated by: the membership
	// minus anyone blocked in either direction. The anonymity thresholds are judged against this,
	// because a group of five with three blocks presents itself as safely anonymous while behaving
	// like a group of two.
	EffectiveMemberCount int
	Group                db.Group
	Members              []db.GetGroupMembersRow
	ActiveSeason         *db.Season
}

func (s *Service) GetGroup(ctx context.Context, groupID, userID string) (*GroupDetail, error) {
	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		return nil, err
	}
	if isMember == 0 {
		return nil, ErrNotMember
	}

	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	members, err := s.queries.GetGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	detail := &GroupDetail{
		Group:   group,
		Members: members,
	}

	detail.EffectiveMemberCount = len(members)
	// Scoped to this group: the global block list would also subtract people this member blocked in
	// other groups, which has nothing to do with how anonymous this group's percentages are.
	if blocked, err := s.queries.CountBlockedGroupMembers(ctx, db.CountBlockedGroupMembersParams{
		GroupID:   groupID,
		BlockerID: userID,
	}); err == nil {
		detail.EffectiveMemberCount -= int(blocked)
		if detail.EffectiveMemberCount < 1 {
			detail.EffectiveMemberCount = 1 // the viewer always counts
		}
	}

	season, err := s.queries.GetActiveSeasonByGroup(ctx, groupID)
	if err == nil {
		detail.ActiveSeason = &season

		totalQuestions, qErr := s.queries.CountSeasonQuestions(ctx, season.ID)
		if qErr != nil {
			return nil, qErr
		}
		votedCount, vErr := s.queries.CountCompletedVoters(ctx, db.CountCompletedVotersParams{
			SeasonID: season.ID,
			Column2:  totalQuestions,
		})
		if vErr != nil {
			return nil, vErr
		}
		status := eligibility.Evaluate(int64(len(members)), votedCount)
		detail.RevealStatus = &status
	}

	return detail, nil
}

type JoinPreview struct {
	Name          string
	MemberCount   int64
	AdminUsername string
}

func (s *Service) GetJoinPreview(ctx context.Context, inviteCode string) (*JoinPreview, error) {
	group, err := s.queries.GetGroupByInviteCode(ctx, NormalizeInviteCode(inviteCode))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	memberCount, err := s.queries.CountGroupMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	adminUsername, err := s.queries.GetAdminUsername(ctx, group.AdminID)
	if err != nil {
		return nil, err
	}

	return &JoinPreview{
		Name:          group.Name,
		MemberCount:   memberCount,
		AdminUsername: adminUsername,
	}, nil
}

// ParseJoinSource maps a client-supplied source to the enum.
//
// An unrecognised value becomes UNKNOWN rather than an error: a client sending a source the
// server does not know is version skew, and refusing the join over a telemetry field would
// trade an acquisition for a datum.
func ParseJoinSource(raw string) db.JoinSource {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "LINK":
		return db.JoinSourceLINK
	case "CODE":
		return db.JoinSourceCODE
	case "CARD":
		return db.JoinSourceCARD
	case "TELEGRAM":
		return db.JoinSourceTELEGRAM
	default:
		return db.JoinSourceUNKNOWN
	}
}

func (s *Service) JoinGroup(ctx context.Context, userID, inviteCode string, source db.JoinSource, referrerID string) (*db.Group, error) {
	group, err := s.queries.GetGroupByInviteCode(ctx, NormalizeInviteCode(inviteCode))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: group.ID,
	})
	if err != nil {
		return nil, err
	}
	if isMember > 0 {
		return nil, ErrAlreadyMember
	}

	memberCount, err := s.queries.CountGroupMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	if memberCount >= MaxMembersPerGroup {
		return nil, ErrGroupLimitSize
	}

	userGroupCount, err := s.queries.CountUserGroups(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userGroupCount >= MaxGroupsPerUser {
		return nil, ErrGroupLimitUser
	}

	// A removal must not be reversible with a link — not even a regenerated one, because the ban is
	// on the person and the group rather than on the code.
	banned, err := s.queries.IsGroupBanned(ctx, db.IsGroupBannedParams{
		GroupID: group.ID,
		UserID:  userID,
	})
	if err != nil {
		return nil, err
	}
	if banned {
		return nil, ErrGroupBanned
	}

	// A referrer is dropped rather than refused when it is not a real member of this group or
	// is the joining user: losing an acquisition over an attribution field is the worse trade.
	invitedBy := s.validReferrer(ctx, group.ID, userID, referrerID)

	_, err = s.queries.AddGroupMember(ctx, db.AddGroupMemberParams{
		ID:         uuid.New().String(),
		UserID:     userID,
		GroupID:    group.ID,
		JoinSource: source,
		InvitedBy:  invitedBy,
	})
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (s *Service) LeaveGroup(ctx context.Context, userID, groupID string) error {
	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		return err
	}
	if isMember == 0 {
		return ErrNotMember
	}

	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupNotFound
		}
		return err
	}

	memberCount, err := s.queries.CountGroupMembers(ctx, groupID)
	if err != nil {
		return err
	}

	if memberCount <= 1 {
		return s.queries.DeleteGroup(ctx, groupID)
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	if err := qtx.RemoveGroupMember(ctx, db.RemoveGroupMemberParams{
		UserID:  userID,
		GroupID: groupID,
	}); err != nil {
		return err
	}

	if group.AdminID == userID {
		nextAdmin, err := qtx.GetNextAdmin(ctx, db.GetNextAdminParams{
			GroupID: groupID,
			ID:      userID,
		})
		if err != nil {
			return err
		}
		if err := qtx.UpdateGroupAdmin(ctx, db.UpdateGroupAdminParams{
			ID:      groupID,
			AdminID: nextAdmin,
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

type UpdateGroupParams struct {
	UserID           string
	GroupID          string
	Name             *string
	TelegramUsername *string
}

func (s *Service) UpdateGroup(ctx context.Context, p UpdateGroupParams) (*db.Group, error) {
	group, err := s.queries.GetGroupByID(ctx, p.GroupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	if group.AdminID != p.UserID {
		return nil, ErrNotAdmin
	}

	if p.Name != nil {
		if len(*p.Name) < MinGroupName || len(*p.Name) > MaxGroupName {
			return nil, ErrInvalidName
		}
		if err := s.queries.UpdateGroupName(ctx, db.UpdateGroupNameParams{
			ID:   p.GroupID,
			Name: *p.Name,
		}); err != nil {
			return nil, err
		}
	}

	if p.TelegramUsername != nil {
		if err := s.queries.UpdateGroupTelegramUsername(ctx, db.UpdateGroupTelegramUsernameParams{
			ID:                   p.GroupID,
			TelegramChatUsername: sql.NullString{String: *p.TelegramUsername, Valid: *p.TelegramUsername != ""},
		}); err != nil {
			return nil, err
		}
	}

	updated, err := s.queries.GetGroupByID(ctx, p.GroupID)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (s *Service) RegenerateInviteLink(ctx context.Context, userID, groupID string) (string, error) {
	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrGroupNotFound
		}
		return "", err
	}

	if group.AdminID != userID {
		return "", ErrNotAdmin
	}

	newCode, err := s.freshInviteCode(ctx)
	if err != nil {
		return "", err
	}
	// Overwriting the column is what revokes the old code: the lookup only ever matches the
	// current value.
	if err := s.queries.UpdateGroupInviteCode(ctx, db.UpdateGroupInviteCodeParams{
		ID:         groupID,
		InviteCode: newCode,
	}); err != nil {
		return "", err
	}

	return InviteURL(newCode), nil
}

// validReferrer returns the referrer to record, or an empty value when the claim does not hold.
//
// A referral reward is money, so the claim is checked rather than trusted: the referrer must be
// an existing member of the group being joined, and must not be the joining user.
func (s *Service) validReferrer(ctx context.Context, groupID, joiningUserID, referrerID string) sql.NullString {
	none := sql.NullString{}

	if referrerID == "" || referrerID == joiningUserID {
		return none
	}

	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  referrerID,
		GroupID: groupID,
	})
	if err != nil {
		log.Warn().Err(err).Str("referrer_id", referrerID).Msg("could not verify referrer; dropping it")
		return none
	}
	if isMember == 0 {
		return none
	}

	return sql.NullString{String: referrerID, Valid: true}
}

// freshInviteCode returns a code no group currently holds.
//
// Uniqueness is also enforced by groups_invite_code_upper_idx, so this check is about giving
// a clear error instead of a constraint violation — and about not handing out a code that is
// already taken when the caller is inside a transaction that would fail later.
func (s *Service) freshInviteCode(ctx context.Context) (string, error) {
	for attempt := 0; attempt < maxInviteCodeAttempts; attempt++ {
		code, err := generateInviteCode()
		if err != nil {
			return "", err
		}

		_, err = s.queries.GetGroupByInviteCode(ctx, code)
		if errors.Is(err, sql.ErrNoRows) {
			return code, nil
		}
		if err != nil {
			return "", err
		}
		// Taken — draw again.
	}
	return "", ErrInviteCodeExhausted
}

// --- Helpers ---

func validateCategories(cats []string, userBirthYear sql.NullInt32) ([]db.QuestionCategory, error) {
	result := make([]db.QuestionCategory, 0, len(cats))
	for _, c := range cats {
		qc, ok := validCategories[c]
		if !ok {
			return nil, ErrInvalidCategory
		}
		if qc == db.QuestionCategoryROMANCE {
			if !userBirthYear.Valid || time.Now().Year()-int(userBirthYear.Int32) < 18 {
				return nil, ErrRomanceBlocked
			}
		}
		result = append(result, qc)
	}
	return result, nil
}

// assertKindCategories refuses a kind-only group whose categories offer no non-edgy question.
//
// Asked of the bank rather than decided from a hardcoded list of categories: the thing that makes the
// combination unusable is the absence of questions, and that is a property of the bank's contents.
func (s *Service) assertKindCategories(ctx context.Context, categories []db.QuestionCategory) error {
	n, err := s.queries.CountKindQuestionsByCategories(ctx, categories)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoKindCategories
	}
	return nil
}

func categoryStrings(cats []db.QuestionCategory) []string {
	result := make([]string, len(cats))
	for i, c := range cats {
		result[i] = string(c)
	}
	return result
}

// CreateNewSeasons creates a new VOTING season for all active groups (>=3 members)
// that don't currently have one. It also closes any previously REVEALED seasons.
func (s *Service) CreateNewSeasons(ctx context.Context) error {
	groups, err := s.queries.GetGroupsNeedingNewSeason(ctx)
	if err != nil {
		return err
	}

	var failed int
	for _, g := range groups {
		if err := s.createSeasonForGroup(ctx, g); err != nil {
			log.Error().Err(err).Str("group_id", g.ID).Msg("failed to create new season for group")
			failed++
			continue
		}
	}

	if failed > 0 && failed == len(groups) {
		return fmt.Errorf("season creator: all %d groups failed", failed)
	}
	if failed > 0 {
		log.Warn().Int("failed", failed).Int("total", len(groups)).Msg("season creator: some groups failed")
	}
	return nil
}

// CreateWeeklySeasonForGroup opens a weekly season for a single group. It is the one
// entry point shared by every creation path: the Sunday cron (CreateNewSeasons), the
// post-kickoff follow-up, and the hourly maintenance sweep. It is a no-op when the
// group already has an open season, so the paths cannot double-create.
func (s *Service) CreateWeeklySeasonForGroup(ctx context.Context, groupID string) error {
	g, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}

	if _, err := s.queries.GetActiveSeasonByGroup(ctx, groupID); err == nil {
		log.Debug().Str("group_id", groupID).Msg("group already has an open season, skipping creation")
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	return s.createSeasonForGroup(ctx, g)
}

// MaintainSeasons is the hourly safety net: it opens a weekly season for any eligible
// group that has none and whose most recent season ended more than an hour ago. The
// grace window keeps it clear of the Sunday cron and the post-kickoff follow-up.
func (s *Service) MaintainSeasons(ctx context.Context) error {
	groups, err := s.queries.GetGroupsNeedingSeasonMaintenance(ctx)
	if err != nil {
		return err
	}

	var failed int
	for _, g := range groups {
		if err := s.createSeasonForGroup(ctx, g); err != nil {
			log.Error().Err(err).Str("group_id", g.ID).Msg("season maintenance: failed to create season")
			failed++
		}
	}

	if failed > 0 && failed == len(groups) {
		return fmt.Errorf("season maintenance: all %d groups failed", failed)
	}
	if failed > 0 {
		log.Warn().Int("failed", failed).Int("total", len(groups)).Msg("season maintenance: some groups failed")
	}
	return nil
}

func (s *Service) createSeasonForGroup(ctx context.Context, g db.Group) error {
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	// Close any REVEALED seasons for this group
	revealedSeasons, err := qtx.GetRevealedSeasonsForGroup(ctx, g.ID)
	if err != nil {
		return err
	}
	for _, rs := range revealedSeasons {
		if err := qtx.UpdateSeasonStatus(ctx, db.UpdateSeasonStatusParams{
			ID:     rs.ID,
			Status: db.SeasonStatusCLOSED,
		}); err != nil {
			return err
		}
	}

	// Get next season number
	lastNum, err := qtx.GetLastSeasonNumber(ctx, g.ID)
	if err != nil {
		return err
	}
	newNumber := lastNum + 1

	startsAt, revealAt, endsAt := schedule.WeeklySeason(time.Now())

	season, err := qtx.CreateSeason(ctx, db.CreateSeasonParams{
		ID:       uuid.New().String(),
		GroupID:  g.ID,
		Number:   int32(newNumber),
		StartsAt: startsAt,
		RevealAt: revealAt,
		EndsAt:   endsAt,
		Kind:     db.SeasonKindWEEKLY,
	})
	if err != nil {
		return err
	}

	categories := make([]db.QuestionCategory, 0, len(g.Categories))
	for _, c := range g.Categories {
		categories = append(categories, db.QuestionCategory(c))
	}

	memberCount, err := qtx.CountGroupMembers(ctx, g.ID)
	if err != nil {
		return err
	}

	if err := s.selectAndAssignQuestionsTx(ctx, qtx, season.ID, g.ID, categories, int(memberCount), g.KindOnly); err != nil {
		return err
	}

	return tx.Commit()
}

// SetKindOnly changes the group's question-tone setting. Admin only.
//
// Applies from the *next* season: rewriting an open season would change the questions out from under
// members who already answered some of them, and their votes would reference questions no longer in it.
func (s *Service) SetKindOnly(ctx context.Context, userID, groupID string, kindOnly bool) error {
	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupNotFound
		}
		return err
	}
	if group.AdminID != userID {
		return ErrNotAdmin
	}

	if kindOnly {
		cats := make([]db.QuestionCategory, 0, len(group.Categories))
		for _, c := range group.Categories {
			cats = append(cats, db.QuestionCategory(c))
		}
		if err := s.assertKindCategories(ctx, cats); err != nil {
			return err
		}
	}

	return s.queries.UpdateGroupKindOnly(ctx, db.UpdateGroupKindOnlyParams{
		ID:       groupID,
		KindOnly: kindOnly,
	})
}
