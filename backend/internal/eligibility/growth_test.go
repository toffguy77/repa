package eligibility

import "testing"

func TestNextGrowthThreshold(t *testing.T) {
	tests := []struct {
		name        string
		memberCount int64
		effective   int64
		wantSize    int64
		wantNeeded  int64
		wantUnlocks string
		wantNil     bool
	}{
		{name: "a group of one needs two more before anything reveals", memberCount: 1, effective: 1, wantSize: MinMembers, wantNeeded: 2, wantUnlocks: UnlockReveal},
		{name: "a group of two needs one more", memberCount: 2, effective: 2, wantSize: MinMembers, wantNeeded: 1, wantUnlocks: UnlockReveal},
		{name: "a group of three can reveal but has no detector", memberCount: 3, effective: 3, wantSize: MinDetectorMembers, wantNeeded: 2, wantUnlocks: UnlockDetector},
		{name: "a group of four needs one more for the detector", memberCount: 4, effective: 4, wantSize: MinDetectorMembers, wantNeeded: 1, wantUnlocks: UnlockDetector},
		{name: "a group of five has crossed everything", memberCount: 5, effective: 5, wantNil: true},
		{name: "a group of twelve has crossed everything", memberCount: 12, effective: 12, wantNil: true},
		// An empty group is a transient state, not a special case: the rule still answers.
		{name: "an empty group", memberCount: 0, effective: 0, wantSize: MinMembers, wantNeeded: 3, wantUnlocks: UnlockReveal},

		// The Reveal is a group-level event, gated by Evaluate on actual membership. A reader who has
		// blocked someone must not be told the Reveal cannot happen while it is about to: that would
		// contradict the reveal_state shown on the same screen.
		{name: "three members with one block is short of the detector, not the Reveal", memberCount: 3, effective: 2, wantSize: MinDetectorMembers, wantNeeded: 3, wantUnlocks: UnlockDetector},
		{name: "five members with two blocks is still short of the detector", memberCount: 5, effective: 3, wantSize: MinDetectorMembers, wantNeeded: 2, wantUnlocks: UnlockDetector},
		{name: "twelve members with one block has crossed everything", memberCount: 12, effective: 11, wantNil: true},
		// A group below the Reveal floor is told about the Reveal even if blocks make it look smaller
		// still: the nearer, more actionable fact wins.
		{name: "two members with one block still reports the Reveal", memberCount: 2, effective: 1, wantSize: MinMembers, wantNeeded: 1, wantUnlocks: UnlockReveal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextGrowthThreshold(tt.memberCount, tt.effective)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected no threshold, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected a threshold, got nil")
			}
			if got.Size != tt.wantSize || got.Needed != tt.wantNeeded || got.Unlocks != tt.wantUnlocks {
				t.Errorf("got %+v, want size=%d needed=%d unlocks=%s",
					got, tt.wantSize, tt.wantNeeded, tt.wantUnlocks)
			}
		})
	}
}

// TestNextGrowthThreshold_DoesNotSellTheQuorumIncrease pins the one rule that is about honesty rather
// than correctness: at 8 members the share of members required to vote rises from 40% to 50%, which
// makes revealing harder. It must never appear as something to unlock.
func TestNextGrowthThreshold_DoesNotSellTheQuorumIncrease(t *testing.T) {
	// Confirm the quorum change is real, so this test fails if the rule it guards disappears.
	if Evaluate(7, 0).QuorumThreshold == Evaluate(8, 0).QuorumThreshold {
		t.Fatal("the quorum threshold no longer changes at 8 members; update this test deliberately")
	}
	for _, count := range []int64{5, 6, 7, 8, 9} {
		if got := NextGrowthThreshold(count, count); got != nil {
			t.Errorf("member count %d reported threshold %+v; nothing above %d is an unlock",
				count, got, MinDetectorMembers)
		}
	}
}

// TestNextGrowthThreshold_SizesComeFromTheRules asserts the thresholds are the constants the reveal
// and detector rules use, not literals typed twice.
func TestNextGrowthThreshold_SizesComeFromTheRules(t *testing.T) {
	below := NextGrowthThreshold(MinMembers-1, MinMembers-1)
	if below == nil || below.Size != MinMembers {
		t.Errorf("the first threshold should be MinMembers (%d), got %+v", MinMembers, below)
	}
	atMin := NextGrowthThreshold(MinMembers, MinMembers)
	if atMin == nil || atMin.Size != MinDetectorMembers {
		t.Errorf("the second threshold should be MinDetectorMembers (%d), got %+v", MinDetectorMembers, atMin)
	}
	if NextGrowthThreshold(MinDetectorMembers, MinDetectorMembers) != nil {
		t.Error("there should be no threshold at or above MinDetectorMembers")
	}
}
