package telegram

import (
	"testing"
)

func TestRandomAlphaNum(t *testing.T) {
	code := randomAlphaNum(4)
	if len(code) != 4 {
		t.Errorf("expected length 4, got %d", len(code))
	}

	// Should not contain ambiguous characters (0, O, 1, I)
	for _, c := range code {
		switch c {
		case '0', 'O', '1', 'I':
			t.Errorf("code contains ambiguous character: %c", c)
		}
	}

	// Two codes should be different (probabilistic but extremely unlikely to fail)
	code2 := randomAlphaNum(4)
	if code == code2 {
		t.Log("warning: two random codes were identical (extremely unlikely)")
	}
}

func TestRandomAlphaNum_Length(t *testing.T) {
	for _, n := range []int{1, 4, 8, 16} {
		code := randomAlphaNum(n)
		if len(code) != n {
			t.Errorf("expected length %d, got %d", n, len(code))
		}
	}
}
