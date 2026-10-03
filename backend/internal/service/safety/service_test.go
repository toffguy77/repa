package safety

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

type mockQuerier struct {
	db.Querier
	groups      map[string]db.Group
	members     map[string]map[string]bool // groupID -> userID -> true
	blocks      map[string]bool            // "blocker:blocked"
	bans        map[string]bool            // "group:user"
	reports     []db.CreateUserReportParams
	removed     []db.RemoveGroupMemberParams
	createdBans []db.CreateGroupBanParams
}

func newMock() *mockQuerier {
	return &mockQuerier{
		groups: map[string]db.Group{
			"g1": {ID: "g1", Name: "9Б", AdminID: "admin"},
		},
		members: map[string]map[string]bool{
			"g1": {"admin": true, "u1": true, "u2": true},
		},
		blocks: map[string]bool{},
		bans:   map[string]bool{},
	}
}

func (m *mockQuerier) GetGroupByID(_ context.Context, id string) (db.Group, error) {
	g, ok := m.groups[id]
	if !ok {
		return db.Group{}, sql.ErrNoRows
	}
	return g, nil
}

func (m *mockQuerier) IsGroupMember(_ context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if m.members[arg.GroupID][arg.UserID] {
		return 1, nil
	}
	return 0, nil
}

func (m *mockQuerier) CreateBlock(_ context.Context, arg db.CreateBlockParams) (db.Block, error) {
	m.blocks[arg.BlockerID+":"+arg.BlockedID] = true
	return db.Block{ID: arg.ID, BlockerID: arg.BlockerID, BlockedID: arg.BlockedID}, nil
}

func (m *mockQuerier) DeleteBlock(_ context.Context, arg db.DeleteBlockParams) error {
	delete(m.blocks, arg.BlockerID+":"+arg.BlockedID)
	return nil
}

func (m *mockQuerier) IsBlockedEitherWay(_ context.Context, arg db.IsBlockedEitherWayParams) (bool, error) {
	return m.blocks[arg.BlockerID+":"+arg.BlockedID] || m.blocks[arg.BlockedID+":"+arg.BlockerID], nil
}

func (m *mockQuerier) CreateUserReport(_ context.Context, arg db.CreateUserReportParams) (db.UserReport, error) {
	// Mirror the UNIQUE(reported_id, reporter_id) upsert: a second report updates rather than adds.
	for i, r := range m.reports {
		if r.ReportedID == arg.ReportedID && r.ReporterID == arg.ReporterID {
			m.reports[i] = arg
			return db.UserReport{ID: r.ID}, nil
		}
	}
	m.reports = append(m.reports, arg)
	return db.UserReport{ID: arg.ID}, nil
}

func (m *mockQuerier) RemoveGroupMember(_ context.Context, arg db.RemoveGroupMemberParams) error {
	m.removed = append(m.removed, arg)
	delete(m.members[arg.GroupID], arg.UserID)
	return nil
}

func (m *mockQuerier) CreateGroupBan(_ context.Context, arg db.CreateGroupBanParams) (db.GroupBan, error) {
	m.createdBans = append(m.createdBans, arg)
	m.bans[arg.GroupID+":"+arg.UserID] = true
	return db.GroupBan{ID: arg.ID}, nil
}

func (m *mockQuerier) IsGroupBanned(_ context.Context, arg db.IsGroupBannedParams) (bool, error) {
	return m.bans[arg.GroupID+":"+arg.UserID], nil
}

// --- Block ---

func TestBlock_TakesEffectImmediatelyAndBothWays(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.Block(context.Background(), "u1", "u2"); err != nil {
		t.Fatal(err)
	}

	// Only u1 created it, but the effect is symmetric — otherwise the product keeps asking u2 to rate
	// someone who withdrew, and keeps delivering u2's votes to u1.
	forward, _ := svc.IsBlocked(context.Background(), "u1", "u2")
	backward, _ := svc.IsBlocked(context.Background(), "u2", "u1")
	if !forward || !backward {
		t.Errorf("a block must hold in both directions, got %v/%v", forward, backward)
	}
}

func TestBlock_IsIdempotent(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	for i := 0; i < 3; i++ {
		if err := svc.Block(context.Background(), "u1", "u2"); err != nil {
			t.Fatalf("block %d: %v", i, err)
		}
	}
	if len(m.blocks) != 1 {
		t.Errorf("expected one block, got %d", len(m.blocks))
	}
}

func TestBlock_CannotBlockYourself(t *testing.T) {
	svc := NewService(newMock(), nil)

	if err := svc.Block(context.Background(), "u1", "u1"); !errors.Is(err, ErrSelfTarget) {
		t.Errorf("expected ErrSelfTarget, got %v", err)
	}
}

func TestUnblock_RestoresThePreviousState(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.Block(context.Background(), "u1", "u2"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Unblock(context.Background(), "u1", "u2"); err != nil {
		t.Fatal(err)
	}

	blocked, _ := svc.IsBlocked(context.Background(), "u1", "u2")
	if blocked {
		t.Error("unblocking should restore the previous state in both directions")
	}
}

// --- Report ---

func TestReportMember_ReachesTheQueueWithoutBlocking(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.ReportMember(context.Background(), "u1", "u2", "g1", "травит"); err != nil {
		t.Fatal(err)
	}

	if len(m.reports) != 1 {
		t.Fatalf("expected one report, got %d", len(m.reports))
	}
	r := m.reports[0]
	if r.ReportedID != "u2" || r.ReporterID != "u1" {
		t.Errorf("unexpected report %+v", r)
	}
	if !r.Reason.Valid || r.Reason.String != "травит" {
		t.Error("the reason should reach the queue")
	}
	// Reporting must not imply blocking: someone may want a person looked at without cutting
	// themselves off.
	if len(m.blocks) != 0 {
		t.Error("reporting must not create a block")
	}
}

func TestReportMember_IsIdempotentPerPerson(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.ReportMember(context.Background(), "u1", "u2", "g1", "раз"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ReportMember(context.Background(), "u1", "u2", "g1", "два"); err != nil {
		t.Fatal(err)
	}

	if len(m.reports) != 1 {
		t.Errorf("expected one report per person per reporter, got %d", len(m.reports))
	}
}

func TestReportMember_RequiresMembershipAndRefusesSelf(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.ReportMember(context.Background(), "u1", "u1", "g1", ""); !errors.Is(err, ErrSelfTarget) {
		t.Errorf("expected ErrSelfTarget, got %v", err)
	}
	if err := svc.ReportMember(context.Background(), "stranger", "u2", "g1", ""); !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
	if len(m.reports) != 0 {
		t.Error("no report should be filed for a refused request")
	}
}

// --- Removal ---

func TestRemoveMember_RemovesAndBans(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.RemoveMember(context.Background(), "admin", "g1", "u1"); err != nil {
		t.Fatal(err)
	}

	if len(m.removed) != 1 {
		t.Fatalf("expected the member to be removed, got %d removals", len(m.removed))
	}
	// The ban is what an invite cannot undo.
	banned, _ := svc.IsBanned(context.Background(), "g1", "u1")
	if !banned {
		t.Error("a removed member must be banned from re-joining")
	}
}

func TestRemoveMember_OnlyAdmin(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.RemoveMember(context.Background(), "u1", "g1", "u2"); !errors.Is(err, ErrNotAdmin) {
		t.Errorf("expected ErrNotAdmin, got %v", err)
	}
	if len(m.removed) != 0 {
		t.Error("the member must remain")
	}
}

func TestRemoveMember_AdminCannotRemoveThemselves(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	// Removing yourself is leaving, and leaving has its own admin-transfer rules.
	if err := svc.RemoveMember(context.Background(), "admin", "g1", "admin"); !errors.Is(err, ErrSelfTarget) {
		t.Errorf("expected ErrSelfTarget, got %v", err)
	}
}

func TestRemoveMember_NonMemberIsRefused(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.RemoveMember(context.Background(), "admin", "g1", "stranger"); !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestRemoveMember_UnknownGroup(t *testing.T) {
	svc := NewService(newMock(), nil)

	if err := svc.RemoveMember(context.Background(), "admin", "nope", "u1"); !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

// --- Permanent departure ---

func TestBanSelf_MakesLeavingIrreversible(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.BanSelf(context.Background(), "u1", "g1"); err != nil {
		t.Fatal(err)
	}

	banned, _ := svc.IsBanned(context.Background(), "g1", "u1")
	if !banned {
		t.Error("a permanent departure must be irreversible by invite")
	}
	// Ordinary leaving stays reversible: people leave groups by accident, so the two are separate
	// choices.
	other, _ := svc.IsBanned(context.Background(), "g1", "u2")
	if other {
		t.Error("banning one member must not ban another")
	}
}

func TestIsBanned_OtherGroupsUnaffected(t *testing.T) {
	m := newMock()
	m.groups["g2"] = db.Group{ID: "g2", Name: "Другая", AdminID: "admin"}
	svc := NewService(m, nil)

	if err := svc.BanSelf(context.Background(), "u1", "g1"); err != nil {
		t.Fatal(err)
	}

	banned, _ := svc.IsBanned(context.Background(), "g2", "u1")
	if banned {
		t.Error("a ban is per group; another group must be unaffected")
	}
}
