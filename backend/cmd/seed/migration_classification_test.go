package main

import (
	"database/sql"
	"os"
	"regexp"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestMigrationClassifiesTheRealBank runs migration 010's classification UPDATEs against the real
// 205-question bank and checks they agree with the Go classifier's distribution.
//
// This is the only test that exercises the migration's *logic* rather than its marker lists: the
// order of the four UPDATEs, the category fallbacks, and how Postgres' own `lower()` and regex
// operators behave on Russian text. None of that is checked by comparing phrase lists, and the
// migration is what classifies every database that already exists — the e2e harness applies
// migrations before seeding, so there it classifies nothing.
//
// Needs a database, so it is skipped without DATABASE_URL. Everything happens inside a transaction
// that is rolled back, so it leaves no trace:
//
//	DATABASE_URL="postgres://repa:repa@localhost:5432/repa?sslmode=disable" go test ./cmd/seed/
func TestMigrationClassifiesTheRealBank(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("set DATABASE_URL to run the migration classification check")
	}

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()

	updates := classificationUpdates(t)

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	// Rolled back unconditionally: this test reads the bank, it does not change it.
	defer tx.Rollback()

	// Insert the bank as it stood before migration 010 — every row at the column default.
	if _, err := tx.Exec(`DELETE FROM questions WHERE source = 'SYSTEM'`); err != nil {
		t.Fatalf("clear system questions: %v", err)
	}
	for _, q := range questions {
		_, err := tx.Exec(
			`INSERT INTO questions (text, category, source, status) VALUES ($1, $2, 'SYSTEM', 'ACTIVE')`,
			q.Text, q.Category)
		if err != nil {
			t.Fatalf("insert %q: %v", q.Text, err)
		}
	}

	for i, u := range updates {
		if _, err := tx.Exec(u); err != nil {
			t.Fatalf("classification update %d: %v", i+1, err)
		}
	}

	rows, err := tx.Query(`SELECT tone, count(*) FROM questions WHERE source = 'SYSTEM' GROUP BY tone`)
	if err != nil {
		t.Fatalf("count by tone: %v", err)
	}
	defer rows.Close()

	got := map[string]int{}
	for rows.Next() {
		var tone string
		var n int
		if err := rows.Scan(&tone, &n); err != nil {
			t.Fatal(err)
		}
		got[tone] = n
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	// The same numbers TestSeedBank_ToneDistribution pins for the Go classifier. Equal distributions
	// are not proof of identical classification, so the per-question comparison below follows.
	for tone, want := range map[string]int{"WARM": wantWarm, "NEUTRAL": wantNeutral, "EDGY": wantEdgy} {
		if got[tone] != want {
			t.Errorf("migration produced %s = %d, want %d (the Go classifier's count)", tone, got[tone], want)
		}
	}

	// Per question, which is what actually matters: a database classified by the migration must hold
	// the same tones as one classified by the seeder, question for question.
	perQuestion, err := tx.Query(`SELECT text, tone FROM questions WHERE source = 'SYSTEM'`)
	if err != nil {
		t.Fatalf("read tones: %v", err)
	}
	defer perQuestion.Close()

	byText := map[string]string{}
	for _, q := range questions {
		byText[q.Text] = q.tone()
	}
	var mismatches int
	for perQuestion.Next() {
		var text, tone string
		if err := perQuestion.Scan(&text, &tone); err != nil {
			t.Fatal(err)
		}
		if want, ok := byText[text]; ok && want != tone {
			mismatches++
			if mismatches <= 5 {
				t.Errorf("%q: migration says %s, the Go classifier says %s", text, tone, want)
			}
		}
	}
	if err := perQuestion.Err(); err != nil {
		t.Fatal(err)
	}
	if mismatches > 5 {
		t.Errorf("and %d more mismatches", mismatches-5)
	}
}

// classificationUpdates extracts the four tone UPDATEs from migration 010, so the test runs the
// shipped statements rather than a copy that could drift from them.
func classificationUpdates(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile("../../internal/db/migrations/010_question_tone.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	updates := regexp.MustCompile(`(?ms)^UPDATE questions SET tone.*?;`).FindAllString(string(data), -1)
	if len(updates) != 4 {
		t.Fatalf("found %d classification updates in migration 010, want 4", len(updates))
	}
	return updates
}
