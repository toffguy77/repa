package cards

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestInviteQRDataURI_RendersAScannablePNG(t *testing.T) {
	uri := InviteQRDataURI("https://repa.app/join/AB2CD3")

	if uri == "" {
		t.Fatal("expected a data URI for a valid invite link")
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(uri, prefix) {
		t.Fatalf("data URI has the wrong prefix: %.40s", uri)
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(uri, prefix))
	if err != nil {
		t.Fatalf("payload is not valid base64: %v", err)
	}
	// PNG magic number.
	if len(raw) < 8 || string(raw[1:4]) != "PNG" {
		t.Fatalf("payload is not a PNG, first bytes: %v", raw[:min(8, len(raw))])
	}
}

func TestInviteQRDataURI_EmptyInputYieldsEmptyString(t *testing.T) {
	if got := InviteQRDataURI(""); got != "" {
		t.Errorf("expected an empty string for empty input, got %.40s", got)
	}
}

func TestInviteQRDataURI_DifferentLinksDifferentImages(t *testing.T) {
	a := InviteQRDataURI("https://repa.app/join/AB2CD3")
	b := InviteQRDataURI("https://repa.app/join/XY9ZW8")

	if a == b {
		t.Error("two different invite links produced the same QR image")
	}
}

func TestInviteQRDataURI_LongLinkStillRenders(t *testing.T) {
	long := "https://repa.app/join/AB2CD3?s=card&extra=" + strings.Repeat("x", 200)

	if got := InviteQRDataURI(long); got == "" {
		t.Error("a long but valid link should still render")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
