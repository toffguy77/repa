package groups

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// inviteAlphabet excludes the characters people actually confuse when reading a code aloud
// or off a screen: the digits 0 and 1, and the letters O, I and L. 31 symbols at
// InviteCodeLength give ~887 million codes.
const inviteAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

// InviteCodeLength is the trade-off point between guessability and what fits in working
// memory while walking to the next desk. See the change's design notes.
const InviteCodeLength = 6

// InviteURLPrefix is the public form of an invite link.
const InviteURLPrefix = "https://repa.app/join/"

// separatorStripper removes the characters a user might type between code groups.
var separatorStripper = strings.NewReplacer(
	" ", "",
	"-", "",
	"\u2013", "", // en dash
	"\u2014", "", // em dash
	"_", "",
)

// maxInviteCodeAttempts bounds collision retries. With 887M codes, exhausting this means
// something is wrong, not that the space is full — so creation fails loudly instead of
// issuing a duplicate.
const maxInviteCodeAttempts = 10

// generateInviteCode returns a fresh code from inviteAlphabet.
//
// It uses crypto/rand rather than math/rand: a predictable sequence would let someone
// enumerate newly created groups, which is the property the alphabet choice protects.
func generateInviteCode() (string, error) {
	var b strings.Builder
	b.Grow(InviteCodeLength)
	max := big.NewInt(int64(len(inviteAlphabet)))
	for i := 0; i < InviteCodeLength; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(inviteAlphabet[n.Int64()])
	}
	return b.String(), nil
}

// NormalizeInviteCode turns whatever a user typed or pasted into the stored form.
//
// Every entry point goes through this one function — handler, deep link, preview — so they
// cannot disagree about what a valid code is. It accepts a bare code in any case, a code with
// separators the user copied from the grouped display, and a full invite URL.
func NormalizeInviteCode(raw string) string {
	s := strings.TrimSpace(raw)

	// A pasted link: keep whatever follows the last slash.
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	// Drop a query string or fragment a share sheet may have appended.
	if i := strings.IndexAny(s, "?#"); i >= 0 {
		s = s[:i]
	}

	// Separators exist only in the grouped display of a *short* code (`AB2 CD3`). Stripping
	// them unconditionally would mangle a legacy UUID code, whose hyphens are part of the
	// stored value — so only strip when what is left could still be a short code.
	if stripped := separatorStripper.Replace(s); len(stripped) <= InviteCodeLength+2 {
		s = stripped
	}

	return strings.ToUpper(s)
}

// FormatInviteCode groups a short code for display, so six characters read as two chunks
// instead of one blur. Purely cosmetic: NormalizeInviteCode strips the gap back out.
func FormatInviteCode(code string) string {
	if len(code) != InviteCodeLength {
		return code // a legacy code is shown as-is
	}
	return code[:3] + " " + code[3:]
}

// InviteURL is the shareable link for a code.
func InviteURL(code string) string {
	return InviteURLPrefix + code
}
