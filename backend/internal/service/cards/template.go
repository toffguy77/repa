package cards

import (
	"fmt"
	"html"
	"strings"
)

// CardData holds the data needed to render a reputation card.
type CardData struct {
	Username        string
	AvatarEmoji     string
	TopAttributes   []CardAttribute
	ReputationTitle string
	GroupName       string
	SeasonNumber    int

	// InviteCode and InviteQRDataURI make the card a standalone invitation: a viewer can act
	// on it without any other message. An empty QR degrades to the printed code (see qr.go).
	InviteCode      string
	InviteQRDataURI string
}

// CardAttribute represents a single attribute on the card.
type CardAttribute struct {
	QuestionText string
	Percentage   float64
}

// cardPalette mirrors the mobile app's **dark** palette so the artifact a user shares and
// the app they came from read as one product.
//
// There is no build step shared between Dart and Go, so these values and
// mobile/lib/core/theme/app_tokens.dart (AppColorTokens.dark) must be changed together.
// The pairing is documented in docs/features/design-system.md; TestCardPalette below pins
// the values so at least an accidental edit on this side is caught.
var cardPalette = struct {
	// Canvas, mid and deep form the background gradient: canvas -> a lifted mid tone ->
	// back down, which is what gives the 9:16 card depth without any imagery.
	Canvas string
	Mid    string
	Deep   string

	// Accent is the brand foreground; AccentFill is the darker fill the bar ends on.
	Accent     string
	AccentFill string

	TextPrimary   string
	TextSecondary string
}{
	Canvas:        "#0b0712", // AppColorTokens.dark.canvas
	Mid:           "#1f1733", // AppColorTokens.dark.surfaceRaised
	Deep:          "#161022", // AppColorTokens.dark.surface
	Accent:        "#9b6dff", // AppColorTokens.dark.accent
	AccentFill:    "#7c3aed", // AppColorTokens.dark.accentFill
	TextPrimary:   "#f6f3ff", // AppColorTokens.dark.textPrimary
	TextSecondary: "#afa3cc", // AppColorTokens.dark.textSecondary
}

// BuildCardHTML generates an HTML string for a 1080x1920 reputation card.
func BuildCardHTML(data CardData) string {
	var attrs strings.Builder
	for _, a := range data.TopAttributes {
		pct := fmt.Sprintf("%.0f", a.Percentage)
		attrs.WriteString(fmt.Sprintf(`
		<div class="attr">
			<div class="attr-header">
				<span class="attr-text">%s</span>
				<span class="attr-pct">%s%%</span>
			</div>
			<div class="bar-bg"><div class="bar-fill" style="width:%s%%"></div></div>
		</div>`, html.EscapeString(a.QuestionText), pct, pct))
	}

	avatar := data.AvatarEmoji
	if avatar == "" {
		avatar = "\U0001F346" // eggplant default
	}

	cta := buildCTABlock(data)

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
	width: 1080px;
	height: 1920px;
	font-family: -apple-system, 'Segoe UI', Roboto, sans-serif;
	color: %s;
	overflow: hidden;
}
.card {
	width: 1080px;
	height: 1920px;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	position: relative;
}
.bg {
	position: absolute;
	top: 0; left: 0; width: 100%%; height: 100%%;
	z-index: 0;
}
.content {
	position: relative;
	z-index: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	width: 100%%;
	padding: 0 80px;
}
.logo {
	font-size: 48px;
	letter-spacing: 8px;
	text-transform: uppercase;
	margin-bottom: 60px;
	opacity: 0.9;
}
.logo-emoji {
	font-size: 56px;
	margin-right: 12px;
}
.avatar-circle {
	width: 220px;
	height: 220px;
	border-radius: 50%%;
	background: rgba(255,255,255,0.12);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 110px;
	margin-bottom: 40px;
	border: 4px solid rgba(255,255,255,0.2);
}
.username {
	font-size: 64px;
	font-weight: 700;
	margin-bottom: 16px;
	text-align: center;
}
.title {
	font-size: 40px;
	font-weight: 500;
	opacity: 0.85;
	margin-bottom: 80px;
	text-align: center;
}
.attrs {
	width: 100%%;
	display: flex;
	flex-direction: column;
	gap: 36px;
	margin-bottom: 100px;
}
.attr-header {
	display: flex;
	justify-content: space-between;
	margin-bottom: 12px;
}
.attr-text {
	font-size: 34px;
	font-weight: 500;
	max-width: 750px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
.attr-pct {
	font-size: 34px;
	font-weight: 700;
}
.bar-bg {
	width: 100%%;
	height: 28px;
	border-radius: 14px;
	background: rgba(255,255,255,0.15);
}
.bar-fill {
	height: 100%%;
	border-radius: 14px;
	background: linear-gradient(90deg, %s, %s);
}
.footer {
	font-size: 30px;
	opacity: 0.6;
	text-align: center;
}
.cta {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 36px;
	margin-top: 48px;
}
.cta-qr {
	/* A light plate with a quiet zone: scanners are far more reliable with normal polarity
	   than with an inverted one, so this is the one place the dark palette gives way to
	   function. */
	background: #ffffff;
	padding: 16px;
	border-radius: 24px;
	line-height: 0;
}
.cta-qr img { width: 180px; height: 180px; display: block; }
.cta-text { text-align: left; }
.cta-label {
	font-size: 26px;
	letter-spacing: 3px;
	text-transform: uppercase;
	opacity: 0.65;
}
.cta-code {
	font-size: 64px;
	font-weight: 800;
	letter-spacing: 6px;
	font-variant-numeric: tabular-nums;
}
.cta-hint {
	font-size: 26px;
	opacity: 0.7;
	margin-top: 8px;
}
</style>
</head>
<body>
<div class="card">
	<svg class="bg" viewBox="0 0 1080 1920" xmlns="http://www.w3.org/2000/svg">
		<defs>
			<linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
				<stop offset="0%%" stop-color="%s"/>
				<stop offset="50%%" stop-color="%s"/>
				<stop offset="100%%" stop-color="%s"/>
			</linearGradient>
		</defs>
		<rect width="1080" height="1920" fill="url(#g)"/>
		<circle cx="200" cy="300" r="400" fill="%s" fill-opacity="0.08"/>
		<circle cx="900" cy="1600" r="500" fill="%s" fill-opacity="0.06"/>
	</svg>
	<div class="content">
		<div class="logo"><span class="logo-emoji">%s</span>РЕПА</div>
		<div class="avatar-circle">%s</div>
		<div class="username">%s</div>
		<div class="title">%s</div>
		<div class="attrs">%s</div>
		%s
		<div class="footer">%s · Сезон %d</div>
	</div>
</div>
</body>
</html>`,
		// Order follows the %s placeholders above: body text, the attribute bar gradient,
		// the three background gradient stops, then the two decorative circles.
		cardPalette.TextPrimary,
		cardPalette.Accent,
		cardPalette.AccentFill,
		cardPalette.Canvas,
		cardPalette.Mid,
		cardPalette.Deep,
		cardPalette.AccentFill,
		cardPalette.AccentFill,
		"\U0001F346",
		avatar,
		html.EscapeString(data.Username),
		html.EscapeString(data.ReputationTitle),
		attrs.String(),
		cta,
		html.EscapeString(data.GroupName),
		data.SeasonNumber,
	)
}

// buildCTABlock renders the call to action that turns a card into an invitation.
//
// Returns an empty string when there is no invite code — a card for a group whose code is
// somehow missing should still render, just without the invitation.
func buildCTABlock(data CardData) string {
	if data.InviteCode == "" {
		return ""
	}

	qr := ""
	if data.InviteQRDataURI != "" {
		// The data URI is produced by us, not by user input, so it is embedded as-is.
		qr = fmt.Sprintf(`<div class="cta-qr"><img src="%s" alt=""></div>`, data.InviteQRDataURI)
	}

	return fmt.Sprintf(`<div class="cta">
			%s
			<div class="cta-text">
				<div class="cta-label">Код группы</div>
				<div class="cta-code">%s</div>
				<div class="cta-hint">Скачай Репу и введи код</div>
			</div>
		</div>`, qr, html.EscapeString(formatCardCode(data.InviteCode)))
}

// formatCardCode groups a short code so it reads as two chunks. Mirrors FormatInviteCode in
// the groups service; a legacy code is shown as-is.
func formatCardCode(code string) string {
	if len(code) != 6 {
		return code
	}
	return code[:3] + " " + code[3:]
}
