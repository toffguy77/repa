package groups

import (
	"strings"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

func TestNormalizeInviteCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"already normalised", "AB2CD3", "AB2CD3"},
		{"lowercase", "ab2cd3", "AB2CD3"},
		{"mixed case", "Ab2Cd3", "AB2CD3"},
		{"leading and trailing space", "  AB2CD3 ", "AB2CD3"},
		{"grouped display", "AB2 CD3", "AB2CD3"},
		{"hyphenated", "AB2-CD3", "AB2CD3"},
		{"underscored", "AB2_CD3", "AB2CD3"},
		{"en dash", "AB2–CD3", "AB2CD3"},
		{"full url", "https://repa.app/join/AB2CD3", "AB2CD3"},
		{"full url lowercase", "https://repa.app/join/ab2cd3", "AB2CD3"},
		{"url with query", "https://repa.app/join/AB2CD3?utm_source=tg", "AB2CD3"},
		{"url with fragment", "https://repa.app/join/AB2CD3#x", "AB2CD3"},
		{"url with trailing space", " https://repa.app/join/AB2CD3 ", "AB2CD3"},
		// A legacy UUID keeps its hyphens: they are part of the stored value, and stripping
		// them would stop old invite links from resolving.
		{"legacy uuid", "0f1d4e6a-6b3c-4a1e-9f2e-123456789abc", "0F1D4E6A-6B3C-4A1E-9F2E-123456789ABC"},
		{"legacy uuid url", "https://repa.app/join/0f1d4e6a-6b3c-4a1e-9f2e-123456789abc", "0F1D4E6A-6B3C-4A1E-9F2E-123456789ABC"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeInviteCode(tt.in); got != tt.want {
				t.Errorf("NormalizeInviteCode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeInviteCode_IsIdempotent(t *testing.T) {
	for _, in := range []string{"ab2 cd3", "https://repa.app/join/ab2cd3", "AB2-CD3"} {
		once := NormalizeInviteCode(in)
		twice := NormalizeInviteCode(once)
		if once != twice {
			t.Errorf("normalising %q twice changed it: %q -> %q", in, once, twice)
		}
	}
}

func TestGenerateInviteCode_ShapeAndUniqueness(t *testing.T) {
	const n = 1000
	seen := make(map[string]bool, n)

	for i := 0; i < n; i++ {
		code, err := generateInviteCode()
		if err != nil {
			t.Fatalf("generateInviteCode: %v", err)
		}

		if len(code) != InviteCodeLength {
			t.Fatalf("code %q has length %d, want %d", code, len(code), InviteCodeLength)
		}
		for _, r := range code {
			if !strings.ContainsRune(inviteAlphabet, r) {
				t.Fatalf("code %q contains %q, which is not in the alphabet", code, r)
			}
		}
		if strings.ContainsAny(code, "01OIL") {
			t.Fatalf("code %q contains a confusable character", code)
		}
		if seen[code] {
			t.Fatalf("generated a duplicate code %q within %d draws", code, n)
		}
		seen[code] = true
	}
}

func TestInviteAlphabet_ExcludesConfusables(t *testing.T) {
	for _, r := range "01OIL" {
		if strings.ContainsRune(inviteAlphabet, r) {
			t.Errorf("alphabet must not contain the confusable %q", r)
		}
	}
	if len(inviteAlphabet) != 31 {
		t.Errorf("alphabet has %d symbols, want 31", len(inviteAlphabet))
	}
}

func TestFormatInviteCode(t *testing.T) {
	if got := FormatInviteCode("AB2CD3"); got != "AB2 CD3" {
		t.Errorf("FormatInviteCode(\"AB2CD3\") = %q, want \"AB2 CD3\"", got)
	}

	// A legacy code is shown as-is rather than mis-grouped.
	legacy := "0f1d4e6a-6b3c-4a1e-9f2e-123456789abc"
	if got := FormatInviteCode(legacy); got != legacy {
		t.Errorf("a legacy code should be unchanged, got %q", got)
	}

	// The grouped form must normalise back to the stored code.
	if got := NormalizeInviteCode(FormatInviteCode("AB2CD3")); got != "AB2CD3" {
		t.Errorf("grouped display did not round-trip: %q", got)
	}
}

func TestInviteURL(t *testing.T) {
	if got := InviteURL("AB2CD3"); got != "https://repa.app/join/AB2CD3" {
		t.Errorf("InviteURL = %q", got)
	}
	if got := NormalizeInviteCode(InviteURL("AB2CD3")); got != "AB2CD3" {
		t.Errorf("an invite URL did not round-trip: %q", got)
	}
}

// --- Join attribution ---

func TestParseJoinSource(t *testing.T) {
	tests := []struct {
		in   string
		want db.JoinSource
	}{
		{"LINK", db.JoinSourceLINK},
		{"CODE", db.JoinSourceCODE},
		{"CARD", db.JoinSourceCARD},
		{"TELEGRAM", db.JoinSourceTELEGRAM},
		{"UNKNOWN", db.JoinSourceUNKNOWN},
		// Case and padding are the client's business, not a reason to lose the attribution.
		{"link", db.JoinSourceLINK},
		{" card ", db.JoinSourceCARD},
		{"Telegram", db.JoinSourceTELEGRAM},
		// Version skew must not cost an acquisition.
		{"", db.JoinSourceUNKNOWN},
		{"something-new", db.JoinSourceUNKNOWN},
		{"'; DROP TABLE groups; --", db.JoinSourceUNKNOWN},
	}

	for _, tt := range tests {
		if got := ParseJoinSource(tt.in); got != tt.want {
			t.Errorf("ParseJoinSource(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseJoinSource_NeverReturnsAnInvalidEnumValue(t *testing.T) {
	valid := map[db.JoinSource]bool{
		db.JoinSourceLINK:     true,
		db.JoinSourceCODE:     true,
		db.JoinSourceCARD:     true,
		db.JoinSourceTELEGRAM: true,
		db.JoinSourceUNKNOWN:  true,
	}
	for _, in := range []string{"", "x", "LINKS", "0", "\u0000", "CARD\n"} {
		got := ParseJoinSource(in)
		if !valid[got] {
			t.Errorf("ParseJoinSource(%q) returned %q, which is not a valid enum value", in, got)
		}
	}
}
