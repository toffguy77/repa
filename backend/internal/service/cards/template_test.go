package cards

import (
	"context"
	"fmt"
	"strings"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// mockCardQuerier implements the subset of db.Querier used by GetCardURL.
type mockCardQuerier struct {
	db.Querier
	cards map[string]string // key: "seasonID:userID" -> imageURL
}

func (m *mockCardQuerier) GetCardCache(_ context.Context, arg db.GetCardCacheParams) (db.CardCache, error) {
	key := arg.SeasonID + ":" + arg.UserID
	url, ok := m.cards[key]
	if !ok {
		return db.CardCache{}, fmt.Errorf("card cache not found")
	}
	return db.CardCache{
		ID:       "cache-id",
		UserID:   arg.UserID,
		SeasonID: arg.SeasonID,
		ImageUrl: url,
	}, nil
}

func TestBuildCardHTML_ContainsUsername(t *testing.T) {
	data := CardData{
		Username:        "alice",
		AvatarEmoji:     "\U0001F60E",
		TopAttributes:   []CardAttribute{{QuestionText: "Who is funniest?", Percentage: 80}},
		ReputationTitle: "Душа компании",
		GroupName:       "Test Group",
		SeasonNumber:    3,
	}

	html := BuildCardHTML(data)

	if !strings.Contains(html, "alice") {
		t.Error("expected HTML to contain username")
	}
	if !strings.Contains(html, "Душа компании") {
		t.Error("expected HTML to contain reputation title")
	}
	if !strings.Contains(html, "Test Group") {
		t.Error("expected HTML to contain group name")
	}
	if !strings.Contains(html, "Сезон 3") {
		t.Error("expected HTML to contain season number")
	}
	if !strings.Contains(html, "80%") {
		t.Error("expected HTML to contain percentage")
	}
	if !strings.Contains(html, "\U0001F60E") {
		t.Error("expected HTML to contain avatar emoji")
	}
}

func TestBuildCardHTML_DefaultAvatar(t *testing.T) {
	data := CardData{
		Username:        "bob",
		AvatarEmoji:     "",
		TopAttributes:   nil,
		ReputationTitle: "Загадка века",
		GroupName:       "Group",
		SeasonNumber:    1,
	}

	html := BuildCardHTML(data)

	// Should use default eggplant emoji
	if !strings.Contains(html, "\U0001F346") {
		t.Error("expected HTML to contain default eggplant emoji")
	}
}

func TestBuildCardHTML_MultipleAttributes(t *testing.T) {
	data := CardData{
		Username:    "charlie",
		AvatarEmoji: "\U0001F525",
		TopAttributes: []CardAttribute{
			{QuestionText: "Who is hottest?", Percentage: 90},
			{QuestionText: "Who is funniest?", Percentage: 60},
			{QuestionText: "Best student?", Percentage: 30.5},
		},
		ReputationTitle: "Горячая штучка",
		GroupName:       "Friends",
		SeasonNumber:    5,
	}

	html := BuildCardHTML(data)

	if !strings.Contains(html, "Who is hottest?") {
		t.Error("expected first attribute question")
	}
	if !strings.Contains(html, "Who is funniest?") {
		t.Error("expected second attribute question")
	}
	if !strings.Contains(html, "Best student?") {
		t.Error("expected third attribute question")
	}
	if !strings.Contains(html, "90%") {
		t.Error("expected 90% in output")
	}
}

func TestBuildCardHTML_EscapesHTML(t *testing.T) {
	data := CardData{
		Username:        "<script>alert('xss')</script>",
		AvatarEmoji:     "\U0001F346",
		TopAttributes:   nil,
		ReputationTitle: "Title",
		GroupName:       "Group & Friends",
		SeasonNumber:    1,
	}

	html := BuildCardHTML(data)

	if strings.Contains(html, "<script>") {
		t.Error("expected HTML-escaped username, found raw <script>")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("expected escaped script tag")
	}
	if !strings.Contains(html, "Group &amp; Friends") {
		t.Error("expected escaped ampersand in group name")
	}
}

func TestTitleForCategory(t *testing.T) {
	tests := []struct {
		cat   string
		title string
	}{
		{"HOT", "Горячая штучка"},
		{"FUNNY", "Душа компании"},
		{"SECRETS", "Хранитель тайн"},
		{"SKILLS", "Мастер на все руки"},
		{"ROMANCE", "Сердцеед"},
		{"STUDY", "Ботан года"},
		{"UNKNOWN", "Загадка века"},
	}
	for _, tt := range tests {
		got := titleForCategory(tt.cat)
		if got != tt.title {
			t.Errorf("titleForCategory(%q) = %q, want %q", tt.cat, got, tt.title)
		}
	}
}

func TestTitleForCategory_AllCases(t *testing.T) {
	// Verify each category individually, including empty string and default
	cases := []struct {
		input    string
		expected string
	}{
		{"HOT", "Горячая штучка"},
		{"FUNNY", "Душа компании"},
		{"SECRETS", "Хранитель тайн"},
		{"SKILLS", "Мастер на все руки"},
		{"ROMANCE", "Сердцеед"},
		{"STUDY", "Ботан года"},
		{"", "Загадка века"},
		{"hot", "Загадка века"},         // lowercase should not match
		{"NONEXISTENT", "Загадка века"}, // random string
	}
	for _, tc := range cases {
		got := titleForCategory(tc.input)
		if got != tc.expected {
			t.Errorf("titleForCategory(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestBuildCardHTML_ContainsExpectedElements(t *testing.T) {
	data := CardData{
		Username:        "testuser",
		AvatarEmoji:     "\U0001F680",
		TopAttributes:   []CardAttribute{{QuestionText: "Best coder?", Percentage: 75}},
		ReputationTitle: "Мастер на все руки",
		GroupName:       "DevTeam",
		SeasonNumber:    2,
	}

	html := BuildCardHTML(data)

	// Verify essential HTML structure elements
	requiredElements := []string{
		"<!DOCTYPE html>",
		"<html>",
		"<head>",
		"<meta charset=\"utf-8\">",
		"<style>",
		"</style>",
		"<body>",
		"</body>",
		"</html>",
		`class="card"`,
		`class="content"`,
		`class="username"`,
		`class="title"`,
		`class="attrs"`,
		`class="footer"`,
		`class="avatar-circle"`,
		`class="logo"`,
		"РЕПА",
	}
	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("expected HTML to contain %q", elem)
		}
	}
}

func TestBuildCardHTML_EscapesSpecialCharsInAttributes(t *testing.T) {
	data := CardData{
		Username:    "user",
		AvatarEmoji: "\U0001F346",
		TopAttributes: []CardAttribute{
			{QuestionText: "Who is <b>best</b> & greatest?", Percentage: 55},
		},
		ReputationTitle: "Title with \"quotes\"",
		GroupName:       "Group <> Test",
		SeasonNumber:    1,
	}

	html := BuildCardHTML(data)

	// Question text should be escaped
	if strings.Contains(html, "<b>best</b>") {
		t.Error("expected question text HTML tags to be escaped")
	}
	if !strings.Contains(html, "&lt;b&gt;best&lt;/b&gt;") {
		t.Error("expected escaped HTML tags in question text")
	}
	if !strings.Contains(html, "&amp; greatest") {
		t.Error("expected escaped ampersand in question text")
	}

	// Reputation title should be escaped
	if !strings.Contains(html, "Title with &#34;quotes&#34;") {
		t.Error("expected escaped quotes in reputation title")
	}

	// Group name should be escaped
	if !strings.Contains(html, "Group &lt;&gt; Test") {
		t.Error("expected escaped angle brackets in group name")
	}
}

func TestBuildCardHTML_EmptyAttributes(t *testing.T) {
	data := CardData{
		Username:        "emptyuser",
		AvatarEmoji:     "\U0001F346",
		TopAttributes:   []CardAttribute{},
		ReputationTitle: "Загадка века",
		GroupName:       "EmptyGroup",
		SeasonNumber:    1,
	}

	html := BuildCardHTML(data)

	// Should still produce valid HTML
	if !strings.Contains(html, "emptyuser") {
		t.Error("expected username in HTML")
	}
	if !strings.Contains(html, "Загадка века") {
		t.Error("expected reputation title in HTML")
	}
	// The attrs div should be present but empty
	if !strings.Contains(html, `class="attrs"`) {
		t.Error("expected attrs container in HTML")
	}
	// Should NOT contain any attr-header divs
	if strings.Contains(html, `class="attr-header"`) {
		t.Error("expected no attribute entries for empty attributes")
	}
}

func TestBuildCardHTML_NilAttributes(t *testing.T) {
	data := CardData{
		Username:        "niluser",
		AvatarEmoji:     "\U0001F346",
		TopAttributes:   nil,
		ReputationTitle: "Загадка века",
		GroupName:       "NilGroup",
		SeasonNumber:    1,
	}

	html := BuildCardHTML(data)

	if !strings.Contains(html, "niluser") {
		t.Error("expected username in HTML with nil attributes")
	}
	if strings.Contains(html, `class="attr-header"`) {
		t.Error("expected no attribute entries for nil attributes")
	}
}

func TestGetCardURL(t *testing.T) {
	mock := &mockCardQuerier{
		cards: map[string]string{
			"season1:user1": "https://s3.example.com/cards/season1/user1.png",
		},
	}
	svc := NewService(mock, nil)

	url, err := svc.GetCardURL(context.Background(), "season1", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://s3.example.com/cards/season1/user1.png" {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestGetCardURL_NotFound(t *testing.T) {
	mock := &mockCardQuerier{
		cards: map[string]string{},
	}
	svc := NewService(mock, nil)

	_, err := svc.GetCardURL(context.Background(), "season1", "unknown_user")
	if err == nil {
		t.Error("expected error for missing card cache entry")
	}
}

// --- Shared palette ---

// TestCardPalette pins the values the card mirrors from the mobile app's dark palette
// (AppColorTokens.dark in mobile/lib/core/theme/app_tokens.dart). There is no shared build
// step between Dart and Go, so this test catches an accidental edit on this side; the
// pairing itself is documented in docs/features/design-system.md.
func TestCardPalette(t *testing.T) {
	cases := map[string]struct {
		got, want string
	}{
		"canvas":        {cardPalette.Canvas, "#0b0712"},
		"mid":           {cardPalette.Mid, "#1f1733"},
		"deep":          {cardPalette.Deep, "#161022"},
		"accent":        {cardPalette.Accent, "#9b6dff"},
		"accentFill":    {cardPalette.AccentFill, "#7c3aed"},
		"textPrimary":   {cardPalette.TextPrimary, "#f6f3ff"},
		"textSecondary": {cardPalette.TextSecondary, "#afa3cc"},
	}
	for name, c := range cases {
		if c.got != c.want {
			t.Errorf("cardPalette.%s = %q, want %q (mirrors AppColorTokens.dark)", name, c.got, c.want)
		}
	}
}

func TestBuildCardHTML_UsesPaletteNotLiterals(t *testing.T) {
	out := BuildCardHTML(CardData{
		Username:        "alice",
		ReputationTitle: "Хранитель Тайн",
		GroupName:       "9Б",
		SeasonNumber:    3,
		TopAttributes: []CardAttribute{
			{QuestionText: "Кто первым побежит при пожаре?", Percentage: 67},
		},
	})

	for _, want := range []string{
		cardPalette.Canvas,
		cardPalette.Mid,
		cardPalette.Deep,
		cardPalette.Accent,
		cardPalette.AccentFill,
		cardPalette.TextPrimary,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered card does not contain palette value %q", want)
		}
	}

	// The pre-redesign hardcoded values must be gone.
	for _, stale := range []string{"#1e1033", "#2d1a4e", "#1a0d2e", "#a78bfa", "rgba(124,58,237"} {
		if strings.Contains(out, stale) {
			t.Errorf("rendered card still contains the pre-redesign value %q", stale)
		}
	}
}

func TestBuildCardHTML_StillEscapesUserContent(t *testing.T) {
	out := BuildCardHTML(CardData{
		Username:        `<script>alert(1)</script>`,
		ReputationTitle: `"><img onerror=x>`,
		GroupName:       `a&b`,
		SeasonNumber:    1,
		TopAttributes: []CardAttribute{
			{QuestionText: `<b>bold</b>`, Percentage: 50},
		},
	})

	for _, unsafe := range []string{"<script>", "<img onerror", "<b>bold</b>"} {
		if strings.Contains(out, unsafe) {
			t.Errorf("rendered card contains unescaped %q", unsafe)
		}
	}
	if !strings.Contains(out, "&amp;") {
		t.Error("expected the group name's ampersand to be escaped")
	}
}

// --- Call to action ---

func TestBuildCardHTML_ShowsInviteCodeAndInstruction(t *testing.T) {
	out := BuildCardHTML(CardData{
		Username:        "alice",
		GroupName:       "9Б",
		SeasonNumber:    2,
		InviteCode:      "AB2CD3",
		InviteQRDataURI: "data:image/png;base64,AAAA",
	})

	if !strings.Contains(out, "AB2 CD3") {
		t.Error("card should show the invite code grouped for transcription")
	}
	if !strings.Contains(out, "Код группы") {
		t.Error("card should label the code")
	}
	if !strings.Contains(out, "Скачай Репу и введи код") {
		t.Error("card should tell the viewer what to do")
	}
	if !strings.Contains(out, `<img src="data:image/png;base64,AAAA"`) {
		t.Error("card should embed the QR image")
	}
}

func TestBuildCardHTML_OmitsQRWhenRenderingFailed(t *testing.T) {
	out := BuildCardHTML(CardData{
		Username:     "alice",
		GroupName:    "9Б",
		SeasonNumber: 2,
		InviteCode:   "AB2CD3",
		// Empty: InviteQRDataURI returns "" rather than an error when rendering fails.
	})

	// The class exists in the stylesheet either way; assert on the markup, not the CSS.
	if strings.Contains(out, `<div class="cta-qr">`) {
		t.Error("the QR plate markup should be omitted entirely when there is no image")
	}
	if strings.Contains(out, "<img src=") {
		t.Error("no image element should be emitted without a data URI")
	}
	// The card must still be a usable invitation in text form.
	if !strings.Contains(out, "AB2 CD3") {
		t.Error("the printed code must survive a QR failure")
	}
	if !strings.Contains(out, "Скачай Репу и введи код") {
		t.Error("the instruction must survive a QR failure")
	}
}

func TestBuildCardHTML_NoInviteCodeStillRenders(t *testing.T) {
	out := BuildCardHTML(CardData{
		Username:     "alice",
		GroupName:    "9Б",
		SeasonNumber: 2,
	})

	if strings.Contains(out, "Код группы") {
		t.Error("there is no invitation to show without a code")
	}
	if !strings.Contains(out, "alice") {
		t.Error("the card itself must still render")
	}
}

func TestFormatCardCode(t *testing.T) {
	if got := formatCardCode("AB2CD3"); got != "AB2 CD3" {
		t.Errorf("formatCardCode = %q, want \"AB2 CD3\"", got)
	}

	legacy := "0f1d4e6a-6b3c-4a1e-9f2e-123456789abc"
	if got := formatCardCode(legacy); got != legacy {
		t.Errorf("a legacy code should be shown as-is, got %q", got)
	}
}

func TestBuildCardHTML_EscapesTheInviteCode(t *testing.T) {
	out := BuildCardHTML(CardData{
		Username:     "alice",
		GroupName:    "9Б",
		SeasonNumber: 1,
		InviteCode:   `<script>x</script>`,
	})

	if strings.Contains(out, "<script>") {
		t.Error("the invite code must be escaped like any other interpolated value")
	}
}

func TestCardInviteLink_CarriesTheSharer(t *testing.T) {
	// The card is generated per member, so its link can say who shared it — a group's code
	// identifies the group, not the inviter.
	uri := InviteQRDataURI("https://repa.app/join/AB2CD3?s=card&ref=member-7")
	if uri == "" {
		t.Fatal("expected the referrer-bearing link to render")
	}

	// A link without a referrer must still render: not every share has one.
	if InviteQRDataURI("https://repa.app/join/AB2CD3?s=card") == "" {
		t.Error("a link without a referrer should still render")
	}
}
