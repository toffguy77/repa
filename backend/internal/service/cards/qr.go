package cards

import (
	"encoding/base64"
	"fmt"

	"github.com/rs/zerolog/log"
	qrcode "github.com/skip2/go-qrcode"
)

// qrSizePx is the rendered QR edge length. Large enough to scan from a phone screen that has
// been re-photographed, small enough to sit in the card footer without competing with the
// attributes — the card's subject is still the person.
const qrSizePx = 240

// InviteQRDataURI renders inviteURL as a base64 PNG data URI for embedding in the card HTML.
//
// Rendering happens in-process rather than through a chart service: an outbound dependency in
// the render path would be a new failure mode, and it would hand every group's invite link to
// a third party.
//
// A failure returns an empty string, never an error. A decoration must not cost a member their
// Reveal card — the caller omits the image and the printed code still works.
func InviteQRDataURI(inviteURL string) string {
	if inviteURL == "" {
		return ""
	}

	// Medium error correction: a card gets viewed on screens and re-photographed, so some
	// redundancy is worth the slightly denser code.
	png, err := qrcode.Encode(inviteURL, qrcode.Medium, qrSizePx)
	if err != nil {
		log.Error().Err(err).Str("invite_url", inviteURL).Msg("failed to render invite QR; card will show the code only")
		return ""
	}

	return fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(png))
}
