package eligibility

import (
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name              string
		members, voted    int64
		wantMembersNeeded int64
		wantVotersNeeded  int64
		wantFloorMet      bool
		wantQuorumMet     bool
		wantEligible      bool
	}{
		{"empty group", 0, 0, 3, 3, false, true, false},
		{"solo founder", 1, 0, 2, 3, false, false, false},
		{"two members both voted", 2, 2, 1, 1, false, true, false},
		{"three members two voted", 3, 2, 0, 1, false, true, false},
		{"three members all voted", 3, 3, 0, 0, true, true, true},
		{"four members two voted (40% quorum, floor short)", 4, 2, 0, 1, false, true, false},
		{"four members three voted", 4, 3, 0, 0, true, true, true},
		{"eight members four voted (50% boundary)", 8, 4, 0, 0, true, true, true},
		{"eight members three voted (quorum short)", 8, 3, 0, 0, true, false, false},
		{"fifty members ten voted", 50, 10, 0, 0, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate(tt.members, tt.voted)
			if got.MembersNeeded != tt.wantMembersNeeded {
				t.Errorf("MembersNeeded = %d, want %d", got.MembersNeeded, tt.wantMembersNeeded)
			}
			if got.VotersNeeded != tt.wantVotersNeeded {
				t.Errorf("VotersNeeded = %d, want %d", got.VotersNeeded, tt.wantVotersNeeded)
			}
			if got.FloorMet != tt.wantFloorMet {
				t.Errorf("FloorMet = %v, want %v", got.FloorMet, tt.wantFloorMet)
			}
			if got.QuorumMet != tt.wantQuorumMet {
				t.Errorf("QuorumMet = %v, want %v", got.QuorumMet, tt.wantQuorumMet)
			}
			if got.Eligible != tt.wantEligible {
				t.Errorf("Eligible = %v, want %v", got.Eligible, tt.wantEligible)
			}
		})
	}
}

func TestEvaluate_QuorumThresholdBySize(t *testing.T) {
	if got := Evaluate(7, 0).QuorumThreshold; got != 0.4 {
		t.Errorf("groups under 8 use the 40%% threshold, got %v", got)
	}
	if got := Evaluate(8, 0).QuorumThreshold; got != 0.5 {
		t.Errorf("groups of 8+ use the 50%% threshold, got %v", got)
	}
}

func TestStateFor(t *testing.T) {
	tests := []struct {
		name          string
		status        db.SeasonStatus
		postponeCount int32
		members       int64
		voted         int64
		want          RevealState
	}{
		{"revealed season", db.SeasonStatusREVEALED, 0, 5, 5, StateRevealed},
		{"closed season", db.SeasonStatusCLOSED, 2, 5, 5, StateRevealed},
		{"solo group", db.SeasonStatusVOTING, 0, 1, 0, StateWaitingForMembers},
		{"two members", db.SeasonStatusVOTING, 0, 2, 2, StateWaitingForMembers},
		{"big enough, too few voters", db.SeasonStatusVOTING, 0, 6, 1, StateWaitingForVoters},
		{"postponed and now eligible", db.SeasonStatusVOTING, 1, 6, 4, StatePostponed},
		{"postponed but still short of voters", db.SeasonStatusVOTING, 1, 6, 2, StateWaitingForVoters},
		{"on track", db.SeasonStatusVOTING, 0, 6, 4, StateScheduled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StateFor(tt.status, tt.postponeCount, Evaluate(tt.members, tt.voted))
			if got != tt.want {
				t.Errorf("StateFor = %q, want %q", got, tt.want)
			}
		})
	}
}
