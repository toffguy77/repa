package groups

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/repa-app/repa/internal/db/sqlc"
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
)

const (
	MaxGroupsPerUser = 10
	MaxMembersPerGroup = 50
	MinGroupName     = 3
	MaxGroupName     = 40
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
	inviteCode := uuid.New().String()

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	group, err := qtx.CreateGroup(ctx, db.CreateGroupParams{
		ID:                    groupID,
		Name:                  p.Name,
		InviteCode:            inviteCode,
		AdminID:               p.UserID,
		TelegramChatUsername:  sql.NullString{String: p.TelegramUsername, Valid: p.TelegramUsername != ""},
		Categories:            categoryStrings(categories),
	})
	if err != nil {
		return nil, err
	}

	_, err = qtx.AddGroupMember(ctx, db.AddGroupMemberParams{
		ID:      uuid.New().String(),
		UserID:  p.UserID,
		GroupID: groupID,
	})
	if err != nil {
		return nil, err
	}

	if err := s.createFirstSeasonTx(ctx, qtx, groupID, categories); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateGroupResult{
		Group:     group,
		InviteURL: "https://repa.app/join/" + inviteCode,
	}, nil
}

func (s *Service) createFirstSeasonTx(ctx context.Context, qtx *db.Queries, groupID string, categories []db.QuestionCategory) error {
	startsAt, revealAt, endsAt := getNextSeasonDates()

	season, err := qtx.CreateSeason(ctx, db.CreateSeasonParams{
		ID:       uuid.New().String(),
		GroupID:  groupID,
		Number:   1,
		StartsAt: startsAt,
		RevealAt: revealAt,
		EndsAt:   endsAt,
	})
	if err != nil {
		return err
	}

	return s.selectAndAssignQuestionsTx(ctx, qtx, season.ID, groupID, categories, 1)
}

func (s *Service) selectAndAssignQuestionsTx(ctx context.Context, qtx *db.Queries, seasonID, groupID string, categories []db.QuestionCategory, memberCount int) error {
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
}

func (s *Service) ListUserGroups(ctx context.Context, userID string) ([]GroupListItem, error) {
	rows, err := s.queries.GetUserGroupsWithStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]GroupListItem, 0, len(rows))
	for _, r := range rows {
		group := db.Group{
			ID:                   r.ID,
			Name:                 r.Name,
			InviteCode:           r.InviteCode,
			AdminID:              r.AdminID,
			TelegramChatID:       r.TelegramChatID,
			TelegramChatUsername: r.TelegramChatUsername,
			TelegramConnectCode:  r.TelegramConnectCode,
			TelegramConnectExpiry: r.TelegramConnectExpiry,
			CreatedAt:            r.CreatedAt,
			Categories:           r.Categories,
		}
		item := GroupListItem{
			Group:       group,
			MemberCount: r.MemberCount,
		}

		if r.ActiveSeasonID.Valid {
			season := db.Season{
				ID:       r.ActiveSeasonID.String,
				GroupID:  r.ID,
				Number:   r.ActiveSeasonNumber.Int32,
				Status:   r.ActiveSeasonStatus.SeasonStatus,
				StartsAt: r.ActiveSeasonStartsAt.Time,
				RevealAt: r.ActiveSeasonRevealAt.Time,
				EndsAt:   r.ActiveSeasonEndsAt.Time,
			}
			item.ActiveSeason = &season
			item.VotedCount = r.VotedCount
			item.UserVoted = r.UserVoteCount > 0
		}

		items = append(items, item)
	}

	return items, nil
}

type GroupDetail struct {
	Group        db.Group
	Members      []db.GetGroupMembersRow
	ActiveSeason *db.Season
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

	season, err := s.queries.GetActiveSeasonByGroup(ctx, groupID)
	if err == nil {
		detail.ActiveSeason = &season
	}

	return detail, nil
}

type JoinPreview struct {
	Name          string
	MemberCount   int64
	AdminUsername  string
}

func (s *Service) GetJoinPreview(ctx context.Context, inviteCode string) (*JoinPreview, error) {
	group, err := s.queries.GetGroupByInviteCode(ctx, inviteCode)
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
		Name:         group.Name,
		MemberCount:  memberCount,
		AdminUsername: adminUsername,
	}, nil
}

func (s *Service) JoinGroup(ctx context.Context, userID, inviteCode string) (*db.Group, error) {
	group, err := s.queries.GetGroupByInviteCode(ctx, inviteCode)
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

	_, err = s.queries.AddGroupMember(ctx, db.AddGroupMemberParams{
		ID:      uuid.New().String(),
		UserID:  userID,
		GroupID: group.ID,
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

	newCode := uuid.New().String()
	if err := s.queries.UpdateGroupInviteCode(ctx, db.UpdateGroupInviteCodeParams{
		ID:         groupID,
		InviteCode: newCode,
	}); err != nil {
		return "", err
	}

	return "https://repa.app/join/" + newCode, nil
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

func categoryStrings(cats []db.QuestionCategory) []string {
	result := make([]string, len(cats))
	for i, c := range cats {
		result[i] = string(c)
	}
	return result
}

func getNextSeasonDates() (startsAt, revealAt, endsAt time.Time) {
	msk := time.FixedZone("MSK", 3*60*60)
	now := time.Now().In(msk)

	// Next Monday 00:00 MSK
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	monday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, msk)
	startsAt = monday.UTC()

	// Friday same week 20:00 MSK = 17:00 UTC
	friday := monday.AddDate(0, 0, 4)
	revealAt = time.Date(friday.Year(), friday.Month(), friday.Day(), 20, 0, 0, 0, msk).UTC()

	// Sunday same week 23:59 MSK = 20:59 UTC
	sunday := monday.AddDate(0, 0, 6)
	endsAt = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 0, 0, msk).UTC()

	return startsAt, revealAt, endsAt
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

	startsAt, revealAt, endsAt := getNextSeasonDates()

	season, err := qtx.CreateSeason(ctx, db.CreateSeasonParams{
		ID:       uuid.New().String(),
		GroupID:  g.ID,
		Number:   int32(newNumber),
		StartsAt: startsAt,
		RevealAt: revealAt,
		EndsAt:   endsAt,
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

	if err := s.selectAndAssignQuestionsTx(ctx, qtx, season.ID, g.ID, categories, int(memberCount)); err != nil {
		return err
	}

	return tx.Commit()
}
