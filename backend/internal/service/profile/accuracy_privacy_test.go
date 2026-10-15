package profile

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// The figure is withheld from everyone but its owner. Each question's winner is published, so another
// member's figure at either extreme pins their individual votes — and because the figure is a rolling
// average weighted by seasons_played, reading the same profile two weeks running solves for that week's
// match count.

func TestGetProfile_AccuracyIsWithheldFromOtherMembers(t *testing.T) {
	svc := NewService(newMock())

	other, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatal(err)
	}
	if other.Stats.GuessAccuracy != nil {
		t.Errorf("another member's profile exposed an accuracy of %f", *other.Stats.GuessAccuracy)
	}

	own, err := svc.GetProfile(context.Background(), "g1", "u1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if own.Stats.GuessAccuracy == nil {
		t.Error("a member's own accuracy describes votes they cast themselves and should be present")
	}
}

// TestGetProfile_DecidedInTheServiceNotTheHandler calls the service directly, with no HTTP layer
// involved: a privacy decision made in a handler would hand the figure to the next caller of the
// service by default, which is the wrong direction to fail in.
func TestGetProfile_DecidedInTheServiceNotTheHandler(t *testing.T) {
	// "u2" rather than an arbitrary id: the viewer still has to be a member of the group, which is a
	// separate check this test is not about.
	svc := NewService(newMock())
	resp, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Stats.GuessAccuracy != nil {
		t.Error("the service itself must withhold the figure, not rely on a caller to strip it")
	}
}

func TestGetProfile_OtherStatisticsAreUnaffected(t *testing.T) {
	// The attack needs a count of *matches*. These count participation, not agreement: knowing a member
	// cast 40 votes says nothing about which targets they chose. They are left alone rather than
	// stripped on suspicion.
	svc := NewService(newMock())
	resp, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Stats.SeasonsPlayed == 0 {
		t.Error("seasons_played should still be visible")
	}
	if resp.Stats.VotingStreak == 0 {
		t.Error("voting_streak should still be visible")
	}
	if resp.Stats.MaxVotingStreak == 0 {
		t.Error("max_voting_streak should still be visible")
	}
	if resp.Stats.TotalVotesCast == 0 {
		t.Error("total_votes_cast should still be visible")
	}
	if resp.Stats.TotalVotesReceived == 0 {
		t.Error("total_votes_received should still be visible")
	}
}

// TestGetProfile_AccuracyKeyIsAbsentOnTheWire checks the serialised form, which is what reaches a
// client. A zeroed field would still be a `"guess_accuracy": 0` a client would render as "never matched
// anything" — a different and false claim.
func TestGetProfile_AccuracyKeyIsAbsentOnTheWire(t *testing.T) {
	svc := NewService(newMock())

	other, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatal(err)
	}
	blob, err := json.Marshal(other.Stats)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(blob), "guess_accuracy") {
		t.Errorf("another member's serialised stats carry the key: %s", blob)
	}

	own, err := svc.GetProfile(context.Background(), "g1", "u1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	ownBlob, err := json.Marshal(own.Stats)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(ownBlob), "guess_accuracy") {
		t.Errorf("one's own serialised stats should carry the key: %s", ownBlob)
	}
}
