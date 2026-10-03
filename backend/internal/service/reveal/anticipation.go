package reveal

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/schedule"
)

// TeaserFromWeekday is the first day the leading-category teaser is shown.
//
// The PRD's choice, and the right one for a reason worth recording: a teaser on Monday has four days
// to become boring, while a teaser on Thursday has one night.
const TeaserFromWeekday = time.Thursday

// AnticipationState is everything a member may learn about votes concerning them before the Reveal.
//
// The shape is deliberately *incapable* of leaking rather than merely not leaking today: there is no
// identity field and no attribute field, so no future addition can turn it into a leak. TeaserEmoji
// is the emoji itself, not a category code the client could map back to a name.
type AnticipationState struct {
	VotersAboutMe int64  `json:"voters_about_me"`
	TeaserEmoji   string `json:"teaser_emoji"`
	RevealAt      string `json:"reveal_at"`
}

// categoryEmoji maps a question category to the emoji used to tease it.
//
// Lives here beside the enum rather than in the database or on the client, so a client cannot render
// a category it was not told about.
var categoryEmoji = map[db.QuestionCategory]string{
	db.QuestionCategoryHOT:     "\U0001F525", // 🔥
	db.QuestionCategoryFUNNY:   "\U0001F602", // 😂
	db.QuestionCategorySECRETS: "\U0001F92B", // 🤫
	db.QuestionCategorySKILLS:  "\U0001F4AA", // 💪
	db.QuestionCategoryROMANCE: "❤️",         // ❤️
	db.QuestionCategorySTUDY:   "\U0001F393", // 🎓
}

// CategoryEmoji returns the teaser emoji for a category, or an empty string for an unknown one.
func CategoryEmoji(c db.QuestionCategory) string {
	return categoryEmoji[c]
}

// GetAnticipation returns what a member may know about votes concerning them mid-week.
//
// Served only while the season is open: after the Reveal the member has the real thing, and a partial
// signal next to a full result is noise.
func (s *Service) GetAnticipation(ctx context.Context, seasonID, userID string) (*AnticipationState, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSeasonNotFound
		}
		return nil, err
	}

	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: season.GroupID,
	})
	if err != nil {
		return nil, err
	}
	if isMember == 0 {
		return nil, ErrNotMember
	}

	if season.Status != db.SeasonStatusVOTING {
		return nil, ErrSeasonAlreadyRevealed
	}

	count, err := s.queries.CountVotersAboutTarget(ctx, db.CountVotersAboutTargetParams{
		SeasonID: seasonID,
		TargetID: userID,
	})
	if err != nil {
		return nil, err
	}

	state := &AnticipationState{
		VotersAboutMe: count,
		RevealAt:      season.RevealAt.Format(time.RFC3339),
	}

	// The restraint has to live where the data is: gating on the client would make the teaser a
	// client-version property and would ship the leading category to every device from Monday.
	if count > 0 && teaserVisible(time.Now()) {
		state.TeaserEmoji = s.leadingCategoryEmoji(ctx, seasonID, userID)
	}

	return state, nil
}

// teaserVisible reports whether the weekday in MSK has reached the teaser day.
func teaserVisible(now time.Time) bool {
	wd := now.In(schedule.MSK()).Weekday()
	// Sunday sorts as 0 in Go, so a plain >= comparison would exclude it.
	if wd == time.Sunday {
		return true
	}
	return wd >= TeaserFromWeekday
}

// leadingCategoryEmoji returns the emoji for the member's leading category, or "" when it cannot be
// determined. A missing teaser is not an error: the screen simply shows the count.
func (s *Service) leadingCategoryEmoji(ctx context.Context, seasonID, userID string) string {
	row, err := s.queries.GetLeadingCategoryForTarget(ctx, db.GetLeadingCategoryForTargetParams{
		SeasonID: seasonID,
		TargetID: userID,
	})
	if err != nil {
		return ""
	}
	return CategoryEmoji(row.Category)
}
