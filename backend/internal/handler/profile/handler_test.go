package profile

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	db "github.com/repa-app/repa/internal/db/sqlc"
	appmw "github.com/repa-app/repa/internal/middleware"
	profilesvc "github.com/repa-app/repa/internal/service/profile"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &testValidator{}
	return e
}

type testValidator struct{}

func (tv *testValidator) Validate(i any) error { return nil }

func setUser(c echo.Context, userID, username string) {
	c.Set("user", &appmw.JWTClaims{UserID: userID, Username: username})
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(nil)
	if h == nil {
		t.Fatal("expected non-nil handler from NewHandler(nil)")
	}
}

func TestNewHandler_WithService(t *testing.T) {
	svc := profilesvc.NewService(nil)
	h := NewHandler(svc)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.svc != svc {
		t.Error("handler service does not match the provided service")
	}
}

// ---------------------------------------------------------------------------
// Mock querier for profile handler tests
// ---------------------------------------------------------------------------

type profileMockQuerier struct {
	db.Querier
	members       map[string]map[string]bool // groupID -> userID -> bool
	users         map[string]db.GetUserProfileInfoRow
	stats         map[string]db.UserGroupStat // key = userID:groupID
	topAttr       map[string]db.GetTopAttributeAllTimeRow
	achievements  map[string][]db.Achievement
	seasonHistory map[string][]db.GetUserSeasonHistoryRow
	forceErr      error
}

func (m *profileMockQuerier) IsGroupMember(_ context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if m.forceErr != nil {
		return 0, m.forceErr
	}
	if g, ok := m.members[arg.GroupID]; ok && g[arg.UserID] {
		return 1, nil
	}
	return 0, nil
}

func (m *profileMockQuerier) GetUserProfileInfo(_ context.Context, id string) (db.GetUserProfileInfoRow, error) {
	if m.forceErr != nil {
		return db.GetUserProfileInfoRow{}, m.forceErr
	}
	u, ok := m.users[id]
	if !ok {
		return db.GetUserProfileInfoRow{}, sql.ErrNoRows
	}
	return u, nil
}

func (m *profileMockQuerier) GetUserGroupStats(_ context.Context, arg db.GetUserGroupStatsParams) (db.UserGroupStat, error) {
	key := arg.UserID + ":" + arg.GroupID
	s, ok := m.stats[key]
	if !ok {
		return db.UserGroupStat{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *profileMockQuerier) GetTopAttributeAllTime(_ context.Context, arg db.GetTopAttributeAllTimeParams) (db.GetTopAttributeAllTimeRow, error) {
	key := arg.TargetID + ":" + arg.GroupID
	r, ok := m.topAttr[key]
	if !ok {
		return db.GetTopAttributeAllTimeRow{}, sql.ErrNoRows
	}
	return r, nil
}

func (m *profileMockQuerier) GetUserAchievements(_ context.Context, arg db.GetUserAchievementsParams) ([]db.Achievement, error) {
	return m.achievements[arg.UserID+":"+arg.GroupID], nil
}

func (m *profileMockQuerier) GetUserSeasonHistory(_ context.Context, arg db.GetUserSeasonHistoryParams) ([]db.GetUserSeasonHistoryRow, error) {
	return m.seasonHistory[arg.TargetID+":"+arg.GroupID], nil
}

func newProfileMock() *profileMockQuerier {
	return &profileMockQuerier{
		members: map[string]map[string]bool{
			"g1": {"u1": true, "u2": true, "requester": true},
		},
		users: map[string]db.GetUserProfileInfoRow{
			"u1": {ID: "u1", Username: "alice"},
			"u2": {ID: "u2", Username: "bob"},
		},
		stats:         map[string]db.UserGroupStat{},
		topAttr:       map[string]db.GetTopAttributeAllTimeRow{},
		achievements:  map[string][]db.Achievement{},
		seasonHistory: map[string][]db.GetUserSeasonHistoryRow{},
	}
}

func buildProfileHandler(mq *profileMockQuerier) *Handler {
	svc := profilesvc.NewService(mq)
	return NewHandler(svc)
}

// ---------------------------------------------------------------------------
// GetProfile tests
// ---------------------------------------------------------------------------

func TestGetProfile_Success(t *testing.T) {
	mq := newProfileMock()
	h := buildProfileHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "userId")
	c.SetParamValues("g1", "u1")
	setUser(c, "requester", "requester")

	if err := h.GetProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatal("missing data envelope")
	}
	user, ok := data["user"].(map[string]any)
	if !ok {
		t.Fatal("missing user field")
	}
	if user["username"] != "alice" {
		t.Errorf("expected username alice, got %v", user["username"])
	}
}

func TestGetProfile_UserNotFound(t *testing.T) {
	mq := newProfileMock()
	h := buildProfileHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "userId")
	c.SetParamValues("g1", "nonexistent")
	setUser(c, "requester", "requester")

	_ = h.GetProfile(c)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	errObj := body["error"].(map[string]any)
	if errObj["code"] != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %v", errObj["code"])
	}
}

func TestGetProfile_RequesterNotMember(t *testing.T) {
	mq := newProfileMock()
	h := buildProfileHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "userId")
	c.SetParamValues("g1", "u1")
	setUser(c, "outsider", "outsider")

	_ = h.GetProfile(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	errObj := body["error"].(map[string]any)
	if errObj["code"] != "NOT_MEMBER" {
		t.Errorf("expected NOT_MEMBER, got %v", errObj["code"])
	}
}

func TestGetProfile_GenericError(t *testing.T) {
	mq := newProfileMock()
	mq.forceErr = errors.New("database is down")
	h := buildProfileHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "userId")
	c.SetParamValues("g1", "u1")
	setUser(c, "requester", "requester")

	_ = h.GetProfile(c)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	errObj := body["error"].(map[string]any)
	if errObj["code"] != "INTERNAL" {
		t.Errorf("expected INTERNAL, got %v", errObj["code"])
	}
}
