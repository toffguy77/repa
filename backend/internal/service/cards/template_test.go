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
		{"hot", "Загадка века"},       // lowercase should not match
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
