package cards

import (
	"strings"
	"testing"
)

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
