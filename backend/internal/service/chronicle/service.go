// Package chronicle builds a group's shared record of its past seasons.
//
// Everything it returns is already stored: the chronicle is a different ordering of `season_results`
// rather than a new fact about the group. See openspec/specs/group-chronicle for the requirements.
package chronicle

import (
	"context"
	"errors"
	"sort"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/eligibility"
)

var (
	ErrNotMember = errors.New("user is not a member of this group")
)

// MaxSeasons caps a chronicle page. A group that plays for a year accumulates ~50 seasons, and the
// whole history is not what anyone opens the screen for. A cursor is a later change; this is a limit,
// so adding one does not alter what a page contains.
const MaxSeasons = 20

// MinSeasonsForStanding is how much revealed history the guess standing needs before it is shown.
//
// Two, because the stored accuracy is a rolling average and after one season it has averaged nothing:
// a leaderboard built on a single season's questions ranks luck and states it as a fact about who
// knows the group.
const MinSeasonsForStanding = 2

type Service struct {
	queries db.Querier
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

// EntryDto is one standout result: a question, and who led it.
type EntryDto struct {
	QuestionID   string  `json:"question_id"`
	QuestionText string  `json:"question_text"`
	Category     string  `json:"category"`
	UserID       string  `json:"user_id"`
	Username     string  `json:"username"`
	AvatarEmoji  *string `json:"avatar_emoji"`
	Percentage   float64 `json:"percentage"`
	VoteCount    int32   `json:"vote_count"`
	TotalVoters  int32   `json:"total_voters"`

	// SmallSample marks a result whose percentages were computed over too few voters to hide who
	// voted. Judged on the voters the share was computed over rather than the group's size today: a
	// share of two voters identifies them whatever the group was, and the historic membership is not
	// recoverable — group_members records joined_at and no departure.
	SmallSample bool `json:"small_sample"`
}

// SeasonDto is one season's worth of the record.
type SeasonDto struct {
	SeasonID string     `json:"season_id"`
	Number   int32      `json:"number"`
	RevealAt string     `json:"reveal_at"`
	Entries  []EntryDto `json:"entries"`
}

// StandingEntryDto is one member's place in the guess standing.
type StandingEntryDto struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	AvatarEmoji *string `json:"avatar_emoji"`

	// Rank is 1-based, or 0 for a member who has not been through a reveal yet. Assigned before
	// blocked members are filtered out, so removing a row does not renumber the rest into positions
	// other members do not see.
	Rank int `json:"rank"`

	// Ranked distinguishes "has not played here yet" from "placed last" — the two are different facts,
	// and reporting the first as the second says something untrue about a new member.
	Ranked bool `json:"ranked"`

	// SeasonsPlayed is how much history the place is based on: a standing over two seasons and one
	// over ten are not comparable, so the basis travels with the figure.
	SeasonsPlayed int32 `json:"seasons_played"`

	// The accuracy *figure* is deliberately not here, only the ordering it produces.
	//
	// A published accuracy is an inference channel, not just a score. The chronicle publishes who won
	// each question, so a member at 100% (or 0%) has their vote on every one of those questions
	// recovered exactly. Worse, the figure is a rolling average whose weight is seasons_played, so
	// diffing one week against the next solves for that season's match count:
	// seasonAccuracy = new*(w+1) - old*w. Handing out every member's figure in one response makes that
	// cheap to do for the whole group.
	//
	// A rank cannot be diffed into a count, and ordering is all the feature needs.
}

// StandingDto is the group's guess standing, or the reason it is not being shown.
type StandingDto struct {
	Available bool               `json:"available"`
	Reason    string             `json:"reason,omitempty"`
	Members   []StandingEntryDto `json:"members"`
}

type Chronicle struct {
	Seasons      []SeasonDto `json:"seasons"`
	Standing     StandingDto `json:"standing"`
	SeasonsTotal int64       `json:"seasons_total"`
	SeasonsShown int         `json:"seasons_shown"`
}

// Get returns the group's chronicle for a reader who must be a member of it.
func (s *Service) Get(ctx context.Context, userID, groupID string) (*Chronicle, error) {
	member, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		return nil, err
	}
	if member == 0 {
		return nil, ErrNotMember
	}

	// A block hides the other person from this reader, in the record as well as on the live cards.
	blocked := map[string]bool{}
	if ids, err := s.queries.ListBlockedUserIDs(ctx, userID); err == nil {
		for _, id := range ids {
			blocked[id] = true
		}
	}

	rows, err := s.queries.GetGroupChronicle(ctx, db.GetGroupChronicleParams{
		GroupID: groupID,
		Limit:   int32(MaxSeasons),
	})
	if err != nil {
		return nil, err
	}

	total, err := s.queries.CountRevealedSeasons(ctx, groupID)
	if err != nil {
		return nil, err
	}

	standing, err := s.standing(ctx, groupID, total, blocked)
	if err != nil {
		return nil, err
	}

	seasons := groupSeasons(rows, blocked)
	return &Chronicle{
		Seasons:      seasons,
		Standing:     standing,
		SeasonsTotal: total,
		SeasonsShown: len(seasons),
	}, nil
}

// groupSeasons folds the flat query rows into seasons, newest first.
//
// The query cannot return them in that order: DISTINCT ON requires the ORDER BY to lead with its own
// expressions, so the rows arrive grouped by season id. Sorting here rather than wrapping the query in
// another subquery keeps the query readable and costs nothing at these sizes.
func groupSeasons(rows []db.GetGroupChronicleRow, blocked map[string]bool) []SeasonDto {
	bySeason := map[string]*SeasonDto{}
	order := make([]string, 0, len(rows))

	for _, r := range rows {
		season, ok := bySeason[r.SeasonID]
		if !ok {
			season = &SeasonDto{
				SeasonID: r.SeasonID,
				Number:   r.SeasonNumber,
				RevealAt: r.RevealAt.Format(timeFormat),
				Entries:  []EntryDto{},
			}
			bySeason[r.SeasonID] = season
			order = append(order, r.SeasonID)
		}
		if blocked[r.TargetID] {
			// The entry is withheld, but the season stays: a season reduced to nothing is still part
			// of the group's history, and dropping it would tell the reader that a block happened.
			continue
		}
		var emoji *string
		if r.AvatarEmoji.Valid {
			emoji = &r.AvatarEmoji.String
		}
		season.Entries = append(season.Entries, EntryDto{
			QuestionID:   r.QuestionID,
			QuestionText: r.QuestionText,
			Category:     string(r.QuestionCategory),
			UserID:       r.TargetID,
			Username:     r.Username,
			AvatarEmoji:  emoji,
			Percentage:   r.Percentage,
			VoteCount:    r.VoteCount,
			TotalVoters:  r.TotalVoters,
			SmallSample:  int64(r.TotalVoters) < eligibility.MinDetectorMembers,
		})
	}

	out := make([]SeasonDto, 0, len(order))
	for _, id := range order {
		out = append(out, *bySeason[id])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Number > out[j].Number })
	return out
}

func (s *Service) standing(ctx context.Context, groupID string, revealedSeasons int64, blocked map[string]bool) (StandingDto, error) {
	if revealedSeasons < MinSeasonsForStanding {
		return StandingDto{
			Available: false,
			Reason:    "Нужно ещё хотя бы одну репу, чтобы сравнивать",
			Members:   []StandingEntryDto{},
		}, nil
	}

	rows, err := s.queries.GetGroupGuessStanding(ctx, groupID)
	if err != nil {
		return StandingDto{}, err
	}

	// Ranks are assigned over the whole group, then blocked members are dropped. Ranking after the
	// filter would give each reader their own numbering, so two members comparing screens would see
	// different places for the same person.
	entries := make([]StandingEntryDto, 0, len(rows))
	rank := 0
	for _, r := range rows {
		entry := StandingEntryDto{
			UserID:   r.ID,
			Username: r.Username,
		}
		if r.AvatarEmoji.Valid {
			entry.AvatarEmoji = &r.AvatarEmoji.String
		}
		// A member with no stats row has not been through a reveal here yet: unranked, not last.
		if r.GuessAccuracy.Valid {
			rank++
			entry.Rank = rank
			entry.Ranked = true
			entry.SeasonsPlayed = r.SeasonsPlayed.Int32
		}
		if blocked[r.ID] {
			continue
		}
		entries = append(entries, entry)
	}

	return StandingDto{Available: true, Members: entries}, nil
}

const timeFormat = "2006-01-02T15:04:05Z07:00"
