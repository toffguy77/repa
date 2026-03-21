package auth

import (
	"testing"
)

func TestIsValidImage(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"JPEG magic bytes", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}, true},
		{"PNG magic bytes", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D}, true},
		{"GIF - not supported", []byte{0x47, 0x49, 0x46, 0x38, 0x39}, false},
		{"empty", []byte{}, false},
		{"too short", []byte{0xFF, 0xD8}, false},
		{"text file", []byte("hello world text file"), false},
		{"random bytes", []byte{0x00, 0x01, 0x02, 0x03, 0x04}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidImage(tt.data)
			if got != tt.valid {
				t.Errorf("isValidImage(%v) = %v, want %v", tt.data, got, tt.valid)
			}
		})
	}
}

func TestUsernameRegex(t *testing.T) {
	tests := []struct {
		username string
		valid    bool
	}{
		{"alice", true},
		{"user_123", true},
		{"Алиса", true},
		{"ab", false},              // too short
		{"a", false},               // too short
		{"aaaaabbbbbcccccddddde", false}, // 21 chars, too long
		{"user name", false},       // space not allowed
		{"user@name", false},       // @ not allowed
		{"user-name", false},       // dash not allowed
		{"___", true},              // underscores only
		{"ёЁ_test", true},          // ё is allowed
		{"123", true},              // digits only
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			got := usernameRegex.MatchString(tt.username)
			if got != tt.valid {
				t.Errorf("usernameRegex.MatchString(%q) = %v, want %v", tt.username, got, tt.valid)
			}
		})
	}
}

func TestGenerateUsername(t *testing.T) {
	u1 := generateUsername()
	u2 := generateUsername()

	if u1 == u2 {
		t.Error("expected unique usernames")
	}
	if len(u1) < 5 {
		t.Error("username too short")
	}
	if u1[:5] != "user_" {
		t.Errorf("expected user_ prefix, got %s", u1[:5])
	}
}
