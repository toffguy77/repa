package voting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	revealsvc "github.com/repa-app/repa/internal/service/reveal"
	votingsvc "github.com/repa-app/repa/internal/service/voting"
)

// TestAnonymity_HandlerSourceNoVoterLeak scans actual handler source files
// for voter_id in JSON response constructions. This catches real regressions
// unlike mock-based tests.
func TestAnonymity_HandlerSourceNoVoterLeak(t *testing.T) {
	// Scan all handler Go files (non-test) for voter_id leaks
	handlerDirs := []string{
		".",                  // voting handler
		"../reveal",         // reveal handler
		"../crystals",       // crystals handler
		"../reactions",      // reactions handler
		"../groups",         // groups handler
		"../profile",        // profile handler
	}

	for _, dir := range handlerDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			t.Fatalf("failed to resolve %s: %v", dir, err)
		}

		entries, err := os.ReadDir(absDir)
		if err != nil {
			t.Fatalf("failed to read dir %s: %v", absDir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}

			path := filepath.Join(absDir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			src := string(content)
			relPath := filepath.Join(filepath.Base(absDir), entry.Name())

			// Check for voter_id in JSON map literals and struct tags
			for _, pattern := range []string{
				`"voter_id"`,
				`"voterId"`,
				`json:"voter_id"`,
			} {
				if strings.Contains(src, pattern) {
					t.Errorf("ANONYMITY VIOLATION in %s: contains %s — voter identity must not be exposed in API responses", relPath, pattern)
				}
			}
		}
	}
}

// TestAnonymity_RevealDTONoVoterField checks the actual reveal service DTO
// structs via reflection to ensure no voter_id field leaks into API responses.
func TestAnonymity_RevealDTONoVoterField(t *testing.T) {
	dtoTypes := []struct {
		name string
		typ  reflect.Type
	}{
		{"RevealData", reflect.TypeOf(revealsvc.RevealData{})},
		{"MyCard", reflect.TypeOf(revealsvc.MyCard{})},
		{"AttributeDto", reflect.TypeOf(revealsvc.AttributeDto{})},
		{"TopAttributeDto", reflect.TypeOf(revealsvc.TopAttributeDto{})},
		{"GroupSummary", reflect.TypeOf(revealsvc.GroupSummary{})},
		{"VoterProfile", reflect.TypeOf(revealsvc.VoterProfile{})},
		{"DetectorResult", reflect.TypeOf(revealsvc.DetectorResult{})},
		{"MemberCardDto", reflect.TypeOf(revealsvc.MemberCardDto{})},
	}

	forbiddenTags := []string{"voter_id", "voterId"}

	for _, dt := range dtoTypes {
		for i := range dt.typ.NumField() {
			field := dt.typ.Field(i)
			jsonTag := field.Tag.Get("json")
			for _, forbidden := range forbiddenTags {
				if strings.Contains(jsonTag, forbidden) {
					t.Errorf("ANONYMITY VIOLATION: %s.%s has json tag containing %q", dt.name, field.Name, forbidden)
				}
			}
		}
	}
}

// TestAnonymity_DetectorNoQuestionBinding verifies the DetectorResult DTO
// has no fields that could bind a voter to a specific question or target.
func TestAnonymity_DetectorNoQuestionBinding(t *testing.T) {
	dt := reflect.TypeOf(revealsvc.DetectorResult{})
	vp := reflect.TypeOf(revealsvc.VoterProfile{})

	forbiddenFields := []string{"question_id", "questionId", "target_id", "targetId"}

	for _, typ := range []reflect.Type{dt, vp} {
		for i := range typ.NumField() {
			field := typ.Field(i)
			jsonTag := field.Tag.Get("json")
			for _, forbidden := range forbiddenFields {
				if strings.Contains(jsonTag, forbidden) {
					t.Errorf("ANONYMITY VIOLATION: %s.%s has json tag %q — detector must not bind voters to questions/targets", typ.Name(), field.Name, jsonTag)
				}
			}
		}
	}

	// Also verify by serializing a zero-value DetectorResult
	result := revealsvc.DetectorResult{
		Purchased: true,
		Voters: []revealsvc.VoterProfile{
			{ID: "u1", Username: "test"},
		},
		CrystalBalance: 10,
	}
	body, _ := json.Marshal(result)
	bodyStr := string(body)

	for _, forbidden := range forbiddenFields {
		if strings.Contains(bodyStr, forbidden) {
			t.Errorf("ANONYMITY VIOLATION: serialized DetectorResult contains %q", forbidden)
		}
	}
}

// TestAnonymity_ProgressEndpointNoVoterLeak verifies the progress handler DTO
// and VotingProgress service struct expose no individual voter identities.
func TestAnonymity_ProgressEndpointNoVoterLeak(t *testing.T) {
	// Handler-level DTOs
	handlerDTO := progressDto{Answered: 3, Total: 5}
	body, _ := json.Marshal(handlerDTO)
	bodyStr := string(body)

	for _, forbidden := range []string{"voter_id", "voterId", "voter"} {
		if strings.Contains(bodyStr, forbidden) {
			t.Errorf("ANONYMITY VIOLATION: serialized progressDto contains %q", forbidden)
		}
	}

	// Service-level DTO
	svcDTO := votingsvc.VotingProgress{
		VotedCount:      3,
		TotalCount:      5,
		QuorumReached:   true,
		QuorumThreshold: 0.5,
		UserVoted:       true,
	}

	forbiddenFields := []string{"voter_id", "voterId"}
	typ := reflect.TypeOf(svcDTO)
	for i := range typ.NumField() {
		field := typ.Field(i)
		name := strings.ToLower(field.Name)
		if strings.Contains(name, "voterid") || strings.Contains(name, "voter_id") {
			t.Errorf("ANONYMITY VIOLATION: VotingProgress.%s exposes voter identity", field.Name)
		}
		jsonTag := field.Tag.Get("json")
		for _, forbidden := range forbiddenFields {
			if strings.Contains(jsonTag, forbidden) {
				t.Errorf("ANONYMITY VIOLATION: VotingProgress.%s has json tag containing %q", field.Name, forbidden)
			}
		}
	}
}

// TestAnonymity_ServiceSourceNoVoterInResponse scans service-level source files
// for voter_id appearing in response struct json tags (not in DB query params).
func TestAnonymity_ServiceSourceNoVoterInResponse(t *testing.T) {
	serviceDirs := []string{
		"../../service/reveal",
		"../../service/voting",
		"../../service/achievements",
		"../../service/profile",
		"../../service/reactions",
	}

	for _, dir := range serviceDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			t.Fatalf("failed to resolve %s: %v", dir, err)
		}

		entries, err := os.ReadDir(absDir)
		if err != nil {
			continue // service may not exist yet
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}

			path := filepath.Join(absDir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			src := string(content)
			relPath := filepath.Join(filepath.Base(absDir), entry.Name())

			// voter_id in json struct tags means it would be serialized to API responses
			if strings.Contains(src, `json:"voter_id"`) {
				t.Errorf("ANONYMITY VIOLATION in %s: contains json:\"voter_id\" struct tag — voter identity must not be exposed in API responses", relPath)
			}
			if strings.Contains(src, `json:"voterId"`) {
				t.Errorf("ANONYMITY VIOLATION in %s: contains json:\"voterId\" struct tag — voter identity must not be exposed in API responses", relPath)
			}
		}
	}
}
