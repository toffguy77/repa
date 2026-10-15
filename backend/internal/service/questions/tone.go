package questions

import (
	"strings"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// Tone classification.
//
// Three tones rather than two because the interesting distinction is not "nice / not nice": most of the
// bank is neither a compliment nor an insult («Кто чаще всех опаздывает?»), and collapsing that into
// either bucket would either empty a kind-only group's bank or let real jabs into it.
//
// The property that matters is that a kind-only group cannot receive anything from the edgy set — so
// edgyMarkers must be *complete* rather than precise. A warm question labelled neutral costs nothing.

// warmMarkers mark a question whose answer is a compliment to receive.
var warmMarkers = []string{
	"поможет", "поддержит", "выручит", "обнимет", "подбодрит",
	"добр", "щедр", "надёжн", "надежн", "честн", "вежлив",
	"лучший друг", "можно доверить", "на кого можно положиться",
	"заступится", "утешит", "защитит", "первым поздравит",
	"самый спокойн", "самый весёл", "самый весел", "улыб",
	"талант", "умеет", "лучше всех", "научит", "объяснит",
}

// edgyMarkers mark a question whose answer is a jab. Deliberately broad.
var edgyMarkers = []string{
	"предаст", "наябедничает", "украд", "съест чужую", "сольёт", "сольет",
	"тайно", "притворя", "врёт", "врет", "обман", "завид", "бесит",
	"сплетн", "подставит", "бросит", "сдаст", "спалит", "подслушива",
	"читает чужие", "хуже всех", "не умеет", "позор", "неудач",
	"стыдн", "осудит", "высмеет", "подколет", "унизит",
	"скандал", "драк", "истери", "жад", "лениво", "лень", "опазд",
}

// ClassifyTone assigns a tone to a question.
//
// Phrase markers win over the category, because a category is a coarse signal: SKILLS is usually a
// compliment but «Кто хуже всех водит?» is not, and HOT is usually a jab but «Кто первым поздравит?»
// is not.
func ClassifyTone(text string, category db.QuestionCategory) db.QuestionTone {
	lower := strings.ToLower(text)

	// Edgy first: a question carrying both a warm and an edgy marker is the kind a kind-only group
	// should not receive.
	for _, m := range edgyMarkers {
		if strings.Contains(lower, m) {
			return db.QuestionToneEDGY
		}
	}
	for _, m := range warmMarkers {
		if strings.Contains(lower, m) {
			return db.QuestionToneWARM
		}
	}

	return toneForCategory(category)
}

// toneForCategory is the fallback when no phrase marker matched.
func toneForCategory(category db.QuestionCategory) db.QuestionTone {
	switch category {
	case db.QuestionCategorySKILLS:
		// An ability is a compliment by default.
		return db.QuestionToneWARM
	case db.QuestionCategoryHOT, db.QuestionCategorySECRETS:
		// Provocation and "what are you hiding" are edgy by construction — those categories exist to
		// be edgy, which is exactly why a kind-only group should not draw from them.
		return db.QuestionToneEDGY
	default:
		// FUNNY, STUDY, ROMANCE: jokes and facts, neither compliment nor insult.
		return db.QuestionToneNEUTRAL
	}
}

// IsKindTone reports whether a tone may be used by a kind-only group.
func IsKindTone(t db.QuestionTone) bool {
	return t != db.QuestionToneEDGY
}

// WarmMarkers and EdgyMarkers expose the phrase lists so a test can check they still agree with
// migration 010, which classifies databases that already exist. Copies, so a caller cannot mutate the
// lists the classifier reads.
func WarmMarkers() []string { return append([]string(nil), warmMarkers...) }
func EdgyMarkers() []string { return append([]string(nil), edgyMarkers...) }
