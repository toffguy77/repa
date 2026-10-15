// Package safety is how a member gets away from another member.
//
// Required by App Store Guideline 1.2 — a store needs a way to filter objectionable content, a way to
// report it, and the ability to block abusive users — and by the PRD's own bullying risk, which until
// now was mitigated only for *questions*, never for people.
package safety

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	db "github.com/repa-app/repa/internal/db/sqlc"
)

var (
	ErrNotMember     = errors.New("user is not a member of this group")
	ErrNotAdmin      = errors.New("only the group admin can do this")
	ErrSelfTarget    = errors.New("cannot target yourself")
	ErrGroupNotFound = errors.New("group not found")
)

type Service struct {
	queries db.Querier
	sqlDB   *sql.DB
}

func NewService(queries db.Querier, sqlDB *sql.DB) *Service {
	return &Service{queries: queries, sqlDB: sqlDB}
}

// Block hides two members from each other.
//
// Only the blocker's action is recorded, so only they can undo it, but the *effect* is symmetric: a
// one-directional block would keep asking the blocked person to rate someone who withdrew, and would
// keep delivering their votes to that person — exactly the harm the block exists to stop.
//
// Idempotent: blocking twice succeeds and changes nothing.
func (s *Service) Block(ctx context.Context, blockerID, blockedID string) error {
	if blockerID == blockedID {
		return ErrSelfTarget
	}

	_, err := s.queries.CreateBlock(ctx, db.CreateBlockParams{
		ID:        uuid.New().String(),
		BlockerID: blockerID,
		BlockedID: blockedID,
	})
	return err
}

// Unblock reverses a block created by this member.
func (s *Service) Unblock(ctx context.Context, blockerID, blockedID string) error {
	return s.queries.DeleteBlock(ctx, db.DeleteBlockParams{
		BlockerID: blockerID,
		BlockedID: blockedID,
	})
}

// IsBlocked reports whether a block exists in either direction.
func (s *Service) IsBlocked(ctx context.Context, a, b string) (bool, error) {
	return s.queries.IsBlockedEitherWay(ctx, db.IsBlockedEitherWayParams{
		BlockerID: a,
		BlockedID: b,
	})
}

// ReportMember files a report with the administrative queue.
//
// Independent of blocking — someone may want a person looked at without cutting themselves off — and
// deliberately silent: the reported member is not notified, because telling them would make reporting
// an act of confrontation.
//
// Idempotent on (reported, reporter): a second report updates the reason rather than piling up rows.
func (s *Service) ReportMember(ctx context.Context, reporterID, reportedID, groupID, reason string) error {
	if reporterID == reportedID {
		return ErrSelfTarget
	}

	if err := s.requireMembership(ctx, reporterID, groupID); err != nil {
		return err
	}

	_, err := s.queries.CreateUserReport(ctx, db.CreateUserReportParams{
		ID:         uuid.New().String(),
		ReportedID: reportedID,
		ReporterID: reporterID,
		GroupID:    sql.NullString{String: groupID, Valid: groupID != ""},
		Reason:     sql.NullString{String: reason, Valid: reason != ""},
	})
	return err
}

// RemoveMember takes a member out of a group and bans them from re-joining it.
//
// The ban is what makes this different from the member leaving: an invite — including a regenerated
// one — cannot undo it, because the ban is on the person and the group rather than on the code.
func (s *Service) RemoveMember(ctx context.Context, adminID, groupID, memberID string) error {
	if adminID == memberID {
		// Removing yourself is leaving, and leaving has its own rules about admin transfer.
		return ErrSelfTarget
	}

	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupNotFound
		}
		return err
	}
	if group.AdminID != adminID {
		return ErrNotAdmin
	}

	if err := s.requireMembership(ctx, memberID, groupID); err != nil {
		return err
	}

	if err := s.queries.RemoveGroupMember(ctx, db.RemoveGroupMemberParams{
		UserID:  memberID,
		GroupID: groupID,
	}); err != nil {
		return err
	}

	_, err = s.queries.CreateGroupBan(ctx, db.CreateGroupBanParams{
		ID:       uuid.New().String(),
		GroupID:  groupID,
		UserID:   memberID,
		BannedBy: sql.NullString{String: adminID, Valid: true},
		Reason:   sql.NullString{String: "removed by admin", Valid: true},
	})
	return err
}

// BanSelf records that a member's departure is permanent, so an invite cannot bring them back.
//
// Ordinary leaving stays reversible on purpose — people leave groups by accident — so this is a
// separate, explicit choice.
func (s *Service) BanSelf(ctx context.Context, userID, groupID string) error {
	_, err := s.queries.CreateGroupBan(ctx, db.CreateGroupBanParams{
		ID:       uuid.New().String(),
		GroupID:  groupID,
		UserID:   userID,
		BannedBy: sql.NullString{String: userID, Valid: true},
		Reason:   sql.NullString{String: "left permanently", Valid: true},
	})
	return err
}

// IsBanned reports whether this person may not re-join this group.
func (s *Service) IsBanned(ctx context.Context, groupID, userID string) (bool, error) {
	return s.queries.IsGroupBanned(ctx, db.IsGroupBannedParams{
		GroupID: groupID,
		UserID:  userID,
	})
}

func (s *Service) requireMembership(ctx context.Context, userID, groupID string) error {
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
	return nil
}
