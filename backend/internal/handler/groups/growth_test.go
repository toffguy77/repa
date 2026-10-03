package groups

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/repa-app/repa/internal/eligibility"
)

// memberRow columns, matching GetGroupMembers' scan order.
var growthMemberCols = []string{"id", "username", "avatar_emoji", "avatar_url"}

// getGroupThreshold drives GET /groups/:id with the given membership and block list, and returns the
// response's next_threshold.
func getGroupThreshold(t *testing.T, memberCount int, blockedIDs []string) (map[string]any, bool) {
	t.Helper()
	h, mock, e := newHandlerWithMock(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1", "g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(handlerMockGroupRow("g1", "Группа", "ABC123", "u1"))

	members := sqlmock.NewRows(growthMemberCols)
	for i := 0; i < memberCount; i++ {
		members.AddRow("member-"+string(rune('a'+i)), "user", nil, nil)
	}
	mock.ExpectQuery("SELECT .+ FROM users").
		WithArgs("g1").
		WillReturnRows(members)

	// Scoped to this group: a count over group_members, not the global block list.
	mock.ExpectQuery("SELECT COUNT.+ FROM group_members").
		WithArgs("g1", "u1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(len(blockedIDs))))

	// No active season: the threshold is about the group, not the season.
	seasonCols := []string{"id", "group_id", "number", "status", "starts_at", "reveal_at", "ends_at", "created_at", "kind", "postpone_count"}
	mock.ExpectQuery("SELECT .+ FROM seasons").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows(seasonCols))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/g1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "testuser")

	if err := h.GetGroup(c); err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	raw, present := resp.Data["next_threshold"]
	if !present || raw == nil {
		return nil, false
	}
	return raw.(map[string]any), true
}

func TestGetGroup_NextThreshold(t *testing.T) {
	t.Run("a group of two is told the Reveal needs one more", func(t *testing.T) {
		got, ok := getGroupThreshold(t, 2, nil)
		if !ok {
			t.Fatal("expected a threshold for a group of two")
		}
		if got["unlocks"] != eligibility.UnlockReveal {
			t.Errorf("unlocks = %v, want %s", got["unlocks"], eligibility.UnlockReveal)
		}
		if got["needed"].(float64) != 1 {
			t.Errorf("needed = %v, want 1", got["needed"])
		}
	})

	t.Run("a group of three is told the detector needs two more", func(t *testing.T) {
		got, ok := getGroupThreshold(t, 3, nil)
		if !ok {
			t.Fatal("expected a threshold for a group of three")
		}
		if got["unlocks"] != eligibility.UnlockDetector {
			t.Errorf("unlocks = %v, want %s", got["unlocks"], eligibility.UnlockDetector)
		}
		if got["needed"].(float64) != 2 {
			t.Errorf("needed = %v, want 2", got["needed"])
		}
	})

	t.Run("a group of twelve is given no target", func(t *testing.T) {
		if _, ok := getGroupThreshold(t, 12, nil); ok {
			t.Error("a group past every threshold must not be given a target")
		}
	})

	t.Run("blocks count against the threshold", func(t *testing.T) {
		// Five members, two blocked: the reader can be rated by three, which is exactly the size at
		// which the detector is still withheld. Counting raw membership would have said "big enough"
		// while the anonymity warning said the opposite.
		got, ok := getGroupThreshold(t, 5, []string{"member-b", "member-c"})
		if !ok {
			t.Fatal("expected a threshold once blocks are counted")
		}
		if got["unlocks"] != eligibility.UnlockDetector {
			t.Errorf("unlocks = %v, want %s", got["unlocks"], eligibility.UnlockDetector)
		}
		if got["needed"].(float64) != 2 {
			t.Errorf("needed = %v, want 2 (three effective members)", got["needed"])
		}
	})
}
