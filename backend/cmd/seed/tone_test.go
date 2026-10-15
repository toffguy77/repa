package main

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	questionsvc "github.com/repa-app/repa/internal/service/questions"
)

// The shipped bank's tone distribution, pinned. These numbers are not a target — they are a tripwire:
// moving a question between the marker lists, or adding one to the bank, changes them, and the change
// should be deliberate rather than noticed in production by a kind-only group.
const (
	wantTotal   = 205
	wantWarm    = 35
	wantNeutral = 89
	wantEdgy    = 81
)

func TestSeedBank_ToneDistribution(t *testing.T) {
	counts := map[string]int{}
	for _, q := range questions {
		counts[q.tone()]++
	}

	if len(questions) != wantTotal {
		t.Errorf("bank has %d questions, want %d — update the pinned counts deliberately", len(questions), wantTotal)
	}
	for tone, want := range map[string]int{"WARM": wantWarm, "NEUTRAL": wantNeutral, "EDGY": wantEdgy} {
		if counts[tone] != want {
			t.Errorf("%s = %d, want %d", tone, counts[tone], want)
		}
	}

	// The reason the distribution matters: a kind-only group draws from WARM+NEUTRAL only, and the
	// season takes up to 10 questions while avoiding recent ones. A bank that shrank below this would
	// start repeating questions week after week for those groups.
	if kind := counts["WARM"] + counts["NEUTRAL"]; kind < 60 {
		t.Errorf("only %d questions available to a kind-only group; too few to avoid weekly repeats", kind)
	}
}

func TestSeedBank_EveryQuestionHasAValidTone(t *testing.T) {
	valid := map[string]bool{"WARM": true, "NEUTRAL": true, "EDGY": true}
	for _, q := range questions {
		if !valid[q.tone()] {
			t.Errorf("%q got tone %q", q.Text, q.tone())
		}
	}
}

func TestSeedBank_SpotChecks(t *testing.T) {
	tones := map[string]string{}
	for _, q := range questions {
		tones[q.Text] = q.tone()
	}

	// Named questions on both lists. If one moves between lists this fails by name, which is more
	// useful than a shifted count.
	cases := map[string]string{
		"Кто наябедничает учителю первым?":                      "EDGY",
		"Кто тайно читает чужие переписки?":                     "EDGY",
		"Кто скорее всего съест чужую еду из холодильника?":      "EDGY",
		"Кто притворяется, что всё знает, но на самом деле нет?": "EDGY",
	}
	for text, want := range cases {
		got, ok := tones[text]
		if !ok {
			t.Errorf("%q is no longer in the bank — update this test with its replacement", text)
			continue
		}
		if got != want {
			t.Errorf("%q = %s, want %s", text, got, want)
		}
	}

	// Every SKILLS question is warm unless it carries a jab, and the bank has no SKILLS jabs.
	for _, q := range questions {
		if q.Category == "SKILLS" && q.tone() != "WARM" {
			t.Errorf("SKILLS question %q is %s, want WARM", q.Text, q.tone())
		}
	}
	// HOT and SECRETS exist to provoke; none of them may reach a kind-only group.
	for _, q := range questions {
		if (q.Category == "HOT" || q.Category == "SECRETS") && q.tone() != "EDGY" {
			t.Errorf("%s question %q is %s, want EDGY", q.Category, q.Text, q.tone())
		}
	}
}

// TestMigrationMarkersMatchClassifier is the agreement check between the two places tone is decided:
// migration 010 classifies databases that already exist, the Go classifier classifies fresh seeds and
// user submissions. They are written in different languages against the same phrase lists, so nothing
// but a test keeps them equal — and a drift would show up as one group's bank differing from another's
// purely by when its database was created.
func TestMigrationMarkersMatchClassifier(t *testing.T) {
	sql, err := os.ReadFile("../../internal/db/migrations/010_question_tone.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	// The first two alternations in the file are the edgy and warm lists, in that order — the same
	// order the classifier applies them in.
	alternations := regexp.MustCompile(`~ '\(([^']+)\)'`).FindAllStringSubmatch(string(sql), -1)
	if len(alternations) < 2 {
		t.Fatalf("found %d marker alternations in the migration, want at least 2", len(alternations))
	}

	for i, want := range [][]string{questionsvc.EdgyMarkers(), questionsvc.WarmMarkers()} {
		label := []string{"edgy", "warm"}[i]
		got := strings.Split(alternations[i][1], "|")
		sort.Strings(got)
		sorted := append([]string(nil), want...)
		sort.Strings(sorted)
		if strings.Join(got, "|") != strings.Join(sorted, "|") {
			t.Errorf("%s markers differ between migration 010 and the Go classifier\n migration: %v\n  go:        %v", label, got, sorted)
		}
	}
}
