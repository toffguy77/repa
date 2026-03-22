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
}

// CardAttribute represents a single attribute on the card.
type CardAttribute struct {
	QuestionText string
	Percentage   float64
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
	color: #fff;
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
	background: linear-gradient(90deg, #a78bfa, #7c3aed);
}
.footer {
	font-size: 30px;
	opacity: 0.6;
	text-align: center;
}
</style>
</head>
<body>
<div class="card">
	<svg class="bg" viewBox="0 0 1080 1920" xmlns="http://www.w3.org/2000/svg">
		<defs>
			<linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
				<stop offset="0%%" stop-color="#1e1033"/>
				<stop offset="50%%" stop-color="#2d1a4e"/>
				<stop offset="100%%" stop-color="#1a0d2e"/>
			</linearGradient>
		</defs>
		<rect width="1080" height="1920" fill="url(#g)"/>
		<circle cx="200" cy="300" r="400" fill="rgba(124,58,237,0.08)"/>
		<circle cx="900" cy="1600" r="500" fill="rgba(124,58,237,0.06)"/>
	</svg>
	<div class="content">
		<div class="logo"><span class="logo-emoji">%s</span>РЕПА</div>
		<div class="avatar-circle">%s</div>
		<div class="username">%s</div>
		<div class="title">%s</div>
		<div class="attrs">%s</div>
		<div class="footer">%s · Сезон %d</div>
	</div>
</div>
</body>
</html>`,
		"\U0001F346",
		avatar,
		html.EscapeString(data.Username),
		html.EscapeString(data.ReputationTitle),
		attrs.String(),
		html.EscapeString(data.GroupName),
		data.SeasonNumber,
	)
}

