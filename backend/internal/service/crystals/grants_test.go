package crystals

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

func newGrantsService() (*Service, *mockQuerier) {
	m := &mockQuerier{balance: map[string]int32{}}
	return NewService(m, nil, nil, nil), m
}

func grantsFor(m *mockQuerier, userID string) []db.CreateCrystalLogParams {
	var out []db.CreateCrystalLogParams
	for _, l := range m.createdLogs {
		if l.UserID == userID && l.Type == db.CrystalLogTypeBONUS {
			out = append(out, l)
		}
	}
	return out
}

// --- Welcome ---

func TestGrantWelcome_PaysExactlyOneDetector(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantWelcome(context.Background(), "u1")

	logs := grantsFor(m, "u1")
	if len(logs) != 1 {
		t.Fatalf("expected 1 grant, got %d", len(logs))
	}
	if logs[0].Delta != WelcomeGrant {
		t.Errorf("welcome grant = %d, want %d", logs[0].Delta, WelcomeGrant)
	}
	// The point of the amount: enough to try the hook once.
	if WelcomeGrant != 10 {
		t.Errorf("the welcome grant should equal one detector (10), got %d", WelcomeGrant)
	}
	if !logs[0].Description.Valid || logs[0].Description.String == "" {
		t.Error("a grant must carry a reason the user can read")
	}
}

func TestGrantWelcome_PaysOnce(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantWelcome(context.Background(), "u1")
	svc.GrantWelcome(context.Background(), "u1")
	svc.GrantWelcome(context.Background(), "u1")

	if got := len(grantsFor(m, "u1")); got != 1 {
		t.Errorf("expected the welcome grant to pay once, got %d payments", got)
	}
}

func TestGrantWelcome_SpendingDoesNotReopenTheGrant(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantWelcome(context.Background(), "u1")
	// Simulate spending it on a detector.
	m.balance["u1"] = 0

	svc.GrantWelcome(context.Background(), "u1")

	if got := len(grantsFor(m, "u1")); got != 1 {
		t.Errorf("a spent grant must not be re-paid, got %d payments", got)
	}
}

// --- Referral ---

func TestGrantReferral_PaysTheInviter(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantReferral(context.Background(), "inviter", "invitee")

	logs := grantsFor(m, "inviter")
	if len(logs) != 1 {
		t.Fatalf("expected the inviter to be paid once, got %d", len(logs))
	}
	if logs[0].Delta != ReferralGrant {
		t.Errorf("referral grant = %d, want %d", logs[0].Delta, ReferralGrant)
	}
	if len(grantsFor(m, "invitee")) != 0 {
		t.Error("the invitee is not the one being rewarded")
	}
}

func TestGrantReferral_PaysOncePerInvitedMember(t *testing.T) {
	svc, m := newGrantsService()

	// The invitee completing further sessions must not pay again.
	svc.GrantReferral(context.Background(), "inviter", "invitee")
	svc.GrantReferral(context.Background(), "inviter", "invitee")

	if got := len(grantsFor(m, "inviter")); got != 1 {
		t.Errorf("expected one payment per invited member, got %d", got)
	}
}

func TestGrantReferral_DifferentInviteesBothPay(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantReferral(context.Background(), "inviter", "a")
	svc.GrantReferral(context.Background(), "inviter", "b")

	if got := len(grantsFor(m, "inviter")); got != 2 {
		t.Errorf("two different invitees should pay twice, got %d", got)
	}
}

func TestGrantReferral_IgnoresMissingAndSelfReferral(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantReferral(context.Background(), "", "invitee")
	svc.GrantReferral(context.Background(), "inviter", "")
	svc.GrantReferral(context.Background(), "same", "same")

	if len(m.createdLogs) != 0 {
		t.Errorf("expected no grants, got %d", len(m.createdLogs))
	}
}

// --- Achievements ---

func TestGrantForAchievement_PaysStreakAndRecruiter(t *testing.T) {
	cases := []struct {
		achType db.AchievementType
		want    int32
	}{
		{db.AchievementTypeSTREAKVOTER, StreakGrant},
		{db.AchievementTypeRECRUITER, RecruiterGrant},
	}

	for _, c := range cases {
		svc, m := newGrantsService()
		svc.GrantForAchievement(context.Background(), "u1", "ach-1", c.achType)

		logs := grantsFor(m, "u1")
		if len(logs) != 1 {
			t.Fatalf("%s: expected 1 grant, got %d", c.achType, len(logs))
		}
		if logs[0].Delta != c.want {
			t.Errorf("%s grant = %d, want %d", c.achType, logs[0].Delta, c.want)
		}
	}
}

func TestGrantForAchievement_IgnoresNonPayingAchievements(t *testing.T) {
	svc, m := newGrantsService()

	// BLIND is a joke badge for accuracy under 20%; paying for it would make the currency
	// meaningless.
	for _, t2 := range []db.AchievementType{
		db.AchievementTypeBLIND,
		db.AchievementTypeSNIPER,
		db.AchievementTypeMONOPOLIST,
		db.AchievementTypeNIGHTOWL,
	} {
		svc.GrantForAchievement(context.Background(), "u1", "ach-"+string(t2), t2)
	}

	if len(m.createdLogs) != 0 {
		t.Errorf("expected no grants for non-paying achievements, got %d", len(m.createdLogs))
	}
}

func TestGrantForAchievement_PaysOncePerAchievement(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantForAchievement(context.Background(), "u1", "ach-1", db.AchievementTypeSTREAKVOTER)
	svc.GrantForAchievement(context.Background(), "u1", "ach-1", db.AchievementTypeSTREAKVOTER)

	if got := len(grantsFor(m, "u1")); got != 1 {
		t.Errorf("re-running the achievement job must not double-pay, got %d", got)
	}
}

func TestGrantForAchievement_LaterMilestonePaysAgain(t *testing.T) {
	svc, m := newGrantsService()

	// A second streak milestone is a different achievement row, so it pays.
	svc.GrantForAchievement(context.Background(), "u1", "ach-5", db.AchievementTypeSTREAKVOTER)
	svc.GrantForAchievement(context.Background(), "u1", "ach-10", db.AchievementTypeSTREAKVOTER)

	if got := len(grantsFor(m, "u1")); got != 2 {
		t.Errorf("expected both milestones to pay, got %d", got)
	}
}

// --- Idempotency and balance ---

func TestGrant_DifferentGrantsToOneUserBothPay(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantWelcome(context.Background(), "u1")
	svc.GrantReferral(context.Background(), "u1", "invitee")

	logs := grantsFor(m, "u1")
	if len(logs) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(logs))
	}
	if want := int32(WelcomeGrant + ReferralGrant); m.balance["u1"] != want {
		t.Errorf("balance = %d, want %d", m.balance["u1"], want)
	}
}

func TestGrant_StorageFailureIsSwallowed(t *testing.T) {
	svc, m := newGrantsService()
	m.failOnCreate = true

	// A payout failure must not be able to roll back registering, voting, or awarding.
	svc.GrantWelcome(context.Background(), "u1")

	if len(m.createdLogs) != 0 {
		t.Error("nothing should have been written")
	}
}

func TestIsDuplicateGrant(t *testing.T) {
	dupes := []error{
		errors.New(`pq: duplicate key value violates unique constraint "crystal_logs_external_id_key"`),
		errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)"),
		errors.New("UNIQUE constraint failed: crystal_logs.external_id"),
	}
	for _, err := range dupes {
		if !isDuplicateGrant(err) {
			t.Errorf("expected %v to be recognised as a duplicate", err)
		}
	}

	others := []error{nil, errors.New("connection refused"), errors.New("deadlock detected")}
	for _, err := range others {
		if isDuplicateGrant(err) {
			t.Errorf("did not expect %v to be treated as a duplicate", err)
		}
	}
}

// --- History ---

func TestGetHistory_DistinguishesGrantsFromPurchases(t *testing.T) {
	svc, m := newGrantsService()

	svc.GrantWelcome(context.Background(), "u1")
	// A purchase, written the way the billing path does.
	_, err := m.CreateCrystalLog(context.Background(), db.CreateCrystalLogParams{
		ID:          "buy-1",
		UserID:      "u1",
		Delta:       30,
		Balance:     40,
		Type:        db.CrystalLogTypePURCHASE,
		Description: toNull("Purchase: 30 crystals"),
		ExternalID:  toNull("yk-1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	entries, err := svc.GetHistory(context.Background(), "u1", 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	var grants, purchases int
	for _, e := range entries {
		if e.IsGrant {
			grants++
			if e.Reason == "" {
				t.Error("a grant must carry a reason the user can understand")
			}
		} else {
			purchases++
		}
	}
	if grants != 1 || purchases != 1 {
		t.Errorf("grants = %d, purchases = %d; want 1 and 1", grants, purchases)
	}
}

func TestGetHistory_ClampsTheLimit(t *testing.T) {
	svc, _ := newGrantsService()

	// A caller asking for nothing, or for too much, gets a sane page rather than an error.
	for _, limit := range []int32{0, -1, 1000} {
		if _, err := svc.GetHistory(context.Background(), "u1", limit, 0); err != nil {
			t.Errorf("limit %d: unexpected error %v", limit, err)
		}
	}
}

func toNull(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
