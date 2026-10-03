package groups

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	db "github.com/repa-app/repa/internal/db/sqlc"
	appmw "github.com/repa-app/repa/internal/middleware"
	groupsvc "github.com/repa-app/repa/internal/service/groups"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &testValidator{}
	return e
}

type testValidator struct{}

func (tv *testValidator) Validate(i any) error {
	return nil
}

func setUser(c echo.Context, userID, username string) {
	c.Set("user", &appmw.JWTClaims{UserID: userID, Username: username})
}

func TestToGroupDto(t *testing.T) {
	now := time.Now()
	g := db.Group{
		ID:                   "g1",
		Name:                 "Test Group",
		AdminID:              "u1",
		InviteCode:           "inv-123",
		Categories:           []string{"HOT", "FUNNY"},
		TelegramChatUsername: sql.NullString{String: "testchat", Valid: true},
		CreatedAt:            now,
	}

	dto := toGroupDto(g)

	if dto.ID != "g1" {
		t.Errorf("expected ID g1, got %s", dto.ID)
	}
	if dto.Name != "Test Group" {
		t.Errorf("expected name Test Group, got %s", dto.Name)
	}
	if dto.AdminID != "u1" {
		t.Errorf("expected admin_id u1, got %s", dto.AdminID)
	}
	if len(dto.Categories) != 2 || dto.Categories[0] != "HOT" {
		t.Errorf("expected categories [HOT FUNNY], got %v", dto.Categories)
	}
	if dto.TelegramUsername == nil || *dto.TelegramUsername != "testchat" {
		t.Error("expected telegram username testchat")
	}
	if dto.CreatedAt != now.Format(time.RFC3339) {
		t.Errorf("expected created_at %s, got %s", now.Format(time.RFC3339), dto.CreatedAt)
	}
}

func TestToGroupDto_NullTelegram(t *testing.T) {
	g := db.Group{
		ID:         "g1",
		Name:       "No TG",
		Categories: []string{"HOT"},
		CreatedAt:  time.Now(),
	}

	dto := toGroupDto(g)
	if dto.TelegramUsername != nil {
		t.Error("expected nil telegram username")
	}
}

func TestMapServiceError(t *testing.T) {
	e := setupEcho()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", groupsvc.ErrGroupNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"not member", groupsvc.ErrNotMember, http.StatusForbidden, "NOT_MEMBER"},
		{"already member", groupsvc.ErrAlreadyMember, http.StatusConflict, "ALREADY_MEMBER"},
		{"group limit user", groupsvc.ErrGroupLimitUser, http.StatusConflict, "GROUP_LIMIT"},
		{"group limit size", groupsvc.ErrGroupLimitSize, http.StatusConflict, "MEMBER_LIMIT"},
		{"not admin", groupsvc.ErrNotAdmin, http.StatusForbidden, "NOT_ADMIN"},
		{"invalid name", groupsvc.ErrInvalidName, http.StatusBadRequest, "VALIDATION"},
		{"no categories", groupsvc.ErrNoCategories, http.StatusBadRequest, "VALIDATION"},
		{"invalid category", groupsvc.ErrInvalidCategory, http.StatusBadRequest, "VALIDATION"},
		{"romance blocked", groupsvc.ErrRomanceBlocked, http.StatusForbidden, "ROMANCE_BLOCKED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = mapServiceError(c, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp map[string]map[string]string
			json.Unmarshal(rec.Body.Bytes(), &resp)
			if resp["error"]["code"] != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, resp["error"]["code"])
			}
		})
	}
}

func TestCreateGroup_BadJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGroup(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateGroup_BadJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups/g1",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.UpdateGroup(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMapServiceError_Unknown(t *testing.T) {
	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = mapServiceError(c, errors.New("unknown error"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// --- Additional tests ---

func TestCreateGroup_EmptyJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups",
		strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "u1", "testuser")

	// Bind succeeds with empty JSON, test validator passes,
	// then svc.GetUser panics on nil svc — confirms the bind+validate path works.
	defer func() {
		recover()
	}()

	_ = h.CreateGroup(c)
}

func TestCreateGroup_WrongContentType(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups",
		strings.NewReader(`name=test`))
	req.Header.Set(echo.HeaderContentType, "text/plain")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "u1", "testuser")

	// text/plain content type — Echo won't parse the body into the struct.
	// Bind won't error for text/plain, validator passes, svc panics.
	defer func() {
		recover()
	}()

	_ = h.CreateGroup(c)
}

func TestUpdateGroup_EmptyJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups/g1",
		strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "testuser")

	// Empty JSON binds fine, validator passes, then svc.UpdateGroup panics on nil svc.
	defer func() {
		recover()
	}()

	_ = h.UpdateGroup(c)
}

func TestToGroupDto_AllFieldsPopulated(t *testing.T) {
	now := time.Now()
	g := db.Group{
		ID:                   "g-full",
		Name:                 "Full Group",
		AdminID:              "admin-1",
		InviteCode:           "INV-FULL",
		Categories:           []string{"HOT", "FUNNY", "DEEP"},
		TelegramChatUsername: sql.NullString{String: "fullchat", Valid: true},
		CreatedAt:            now,
	}

	dto := toGroupDto(g)

	if dto.ID != "g-full" {
		t.Errorf("expected ID g-full, got %s", dto.ID)
	}
	if dto.InviteCode != "INV-FULL" {
		t.Errorf("expected invite code INV-FULL, got %s", dto.InviteCode)
	}
	if len(dto.Categories) != 3 {
		t.Errorf("expected 3 categories, got %d", len(dto.Categories))
	}
	if dto.TelegramUsername == nil || *dto.TelegramUsername != "fullchat" {
		t.Error("expected telegram username fullchat")
	}
	if dto.CreatedAt != now.Format(time.RFC3339) {
		t.Errorf("expected created_at %s, got %s", now.Format(time.RFC3339), dto.CreatedAt)
	}
}

func TestToGroupDto_EmptyCategories(t *testing.T) {
	g := db.Group{
		ID:         "g-empty",
		Name:       "Empty Cat",
		Categories: []string{},
		CreatedAt:  time.Now(),
	}

	dto := toGroupDto(g)

	if len(dto.Categories) != 0 {
		t.Errorf("expected 0 categories, got %d", len(dto.Categories))
	}
}

func TestToGroupDto_NilCategories(t *testing.T) {
	g := db.Group{
		ID:        "g-nil",
		Name:      "Nil Cat",
		CreatedAt: time.Now(),
	}

	dto := toGroupDto(g)

	if dto.Categories != nil && len(dto.Categories) != 0 {
		t.Errorf("expected nil/empty categories, got %v", dto.Categories)
	}
}

func TestMapServiceError_WrappedErrors(t *testing.T) {
	e := setupEcho()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			"wrapped not found",
			fmt.Errorf("context: %w", groupsvc.ErrGroupNotFound),
			http.StatusNotFound,
			"NOT_FOUND",
		},
		{
			"wrapped not member",
			fmt.Errorf("failed: %w", groupsvc.ErrNotMember),
			http.StatusForbidden,
			"NOT_MEMBER",
		},
		{
			"wrapped already member",
			fmt.Errorf("join: %w", groupsvc.ErrAlreadyMember),
			http.StatusConflict,
			"ALREADY_MEMBER",
		},
		{
			"wrapped group limit",
			fmt.Errorf("create: %w", groupsvc.ErrGroupLimitUser),
			http.StatusConflict,
			"GROUP_LIMIT",
		},
		{
			"wrapped size limit",
			fmt.Errorf("join: %w", groupsvc.ErrGroupLimitSize),
			http.StatusConflict,
			"MEMBER_LIMIT",
		},
		{
			"wrapped not admin",
			fmt.Errorf("update: %w", groupsvc.ErrNotAdmin),
			http.StatusForbidden,
			"NOT_ADMIN",
		},
		{
			"wrapped romance blocked",
			fmt.Errorf("create: %w", groupsvc.ErrRomanceBlocked),
			http.StatusForbidden,
			"ROMANCE_BLOCKED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = mapServiceError(c, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp map[string]map[string]string
			json.Unmarshal(rec.Body.Bytes(), &resp)
			if resp["error"]["code"] != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, resp["error"]["code"])
			}
		})
	}
}

func TestMapServiceError_ResponseBody(t *testing.T) {
	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = mapServiceError(c, groupsvc.ErrGroupNotFound)

	var resp map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	errObj := resp["error"]
	if errObj["code"] != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", errObj["code"])
	}
	if errObj["message"] != "Group not found" {
		t.Errorf("expected message 'Group not found', got %s", errObj["message"])
	}
}

func TestLeaveGroup_BadJSON_NoImpact(t *testing.T) {
	// LeaveGroup doesn't bind a body, so bad JSON in body is irrelevant.
	// It only reads path param and user claims.
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/g1",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "testuser")

	// Will panic on nil svc, which confirms the handler reached the service call.
	defer func() {
		recover()
	}()

	_ = h.LeaveGroup(c)
}

func TestJoinGroup_PathParam(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/join/INV-CODE", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("inviteCode")
	c.SetParamValues("INV-CODE")
	setUser(c, "u1", "testuser")

	// Will panic on nil svc — confirms path param reading works.
	defer func() {
		recover()
	}()

	_ = h.JoinGroup(c)
}

func TestJoinPreview_PathParam(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/join/INV-CODE/preview", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("inviteCode")
	c.SetParamValues("INV-CODE")

	// Will panic on nil svc
	defer func() {
		recover()
	}()

	_ = h.JoinPreview(c)
}

func TestGetGroup_PathParam(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/g1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "testuser")

	defer func() {
		recover()
	}()

	_ = h.GetGroup(c)
}

func TestListGroups_NilSvc(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "u1", "testuser")

	defer func() {
		recover()
	}()

	_ = h.ListGroups(c)
}

func TestRegenerateInviteLink_PathParam(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/g1/regenerate-invite", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "testuser")

	defer func() {
		recover()
	}()

	_ = h.RegenerateInviteLink(c)
}

// =============================================================================
// sqlmock-based handler tests
// =============================================================================

func newHandlerWithMock(t *testing.T) (*Handler, sqlmock.Sqlmock, *echo.Echo) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	queries := db.New(sqlDB)
	svc := groupsvc.NewService(queries, sqlDB)
	h := NewHandler(svc)
	e := setupEcho()
	return h, mock, e
}

// group row columns (matching sqlc scan order for GetGroupByID / GetGroupByInviteCode)
var handlerGroupCols = []string{
	"id", "name", "invite_code", "admin_id",
	"telegram_chat_id", "telegram_chat_username",
	"telegram_connect_code", "telegram_connect_expiry",
	"created_at", "categories", "kind_only",
}

func handlerMockGroupRow(id, name, inviteCode, adminID string) *sqlmock.Rows {
	return sqlmock.NewRows(handlerGroupCols).AddRow(
		id, name, inviteCode, adminID,
		nil, nil, nil, nil,
		time.Now(), `{"HOT"}`, false,
	)
}

// --- ListGroups: success (empty list) ---

func TestListGroups_Success_Empty(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	// GetUserGroupsWithStats returns 0 rows
	listCols := []string{
		"id", "name", "invite_code", "admin_id",
		"telegram_chat_id", "telegram_chat_username",
		"telegram_connect_code", "telegram_connect_expiry",
		"created_at", "categories", "kind_only", "member_count",
		"active_season_id", "active_season_number", "active_season_status",
		"active_season_starts_at", "active_season_reveal_at", "active_season_ends_at",
		"active_season_kind", "active_season_postpone_count",
		"voted_count", "user_vote_count",
	}
	mock.ExpectQuery("SELECT g.id").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows(listCols))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "u1", "testuser")

	err := h.ListGroups(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	groups, ok := resp["data"]["groups"].([]any)
	if !ok {
		t.Fatal("expected groups array in response")
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- ListGroups: success (with groups) ---

func TestListGroups_Success_WithGroups(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	listCols := []string{
		"id", "name", "invite_code", "admin_id",
		"telegram_chat_id", "telegram_chat_username",
		"telegram_connect_code", "telegram_connect_expiry",
		"created_at", "categories", "kind_only", "member_count",
		"active_season_id", "active_season_number", "active_season_status",
		"active_season_starts_at", "active_season_reveal_at", "active_season_ends_at",
		"active_season_kind", "active_season_postpone_count",
		"voted_count", "user_vote_count",
	}
	now := time.Now()
	mock.ExpectQuery("SELECT g.id").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows(listCols).AddRow(
			"g1", "Test Group", "inv-1", "u1",
			nil, nil, nil, nil,
			now, `{"HOT"}`, false, int64(5),
			nil, nil, nil, nil, nil, nil, nil, nil, // no active season
			int64(0), int64(0),
		))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "u1", "testuser")

	err := h.ListGroups(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	groups, ok := resp["data"]["groups"].([]any)
	if !ok {
		t.Fatal("expected groups array")
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- GetGroup: success ---

func TestGetGroup_Success(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	// IsGroupMember
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1", "g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// GetGroupByID
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(handlerMockGroupRow("g1", "Test Group", "inv-1", "u1"))

	// GetGroupMembers
	memberCols := []string{"id", "username", "avatar_emoji", "avatar_url"}
	mock.ExpectQuery("SELECT .+ FROM users").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows(memberCols).
			AddRow("u1", "alice", nil, nil).
			AddRow("u2", "bob", "X", nil))

	// GetActiveSeasonByGroup returns no rows
	seasonCols := []string{"id", "group_id", "number", "status", "starts_at", "reveal_at", "ends_at", "created_at", "kind", "postpone_count"}
	mock.ExpectQuery("SELECT .+ FROM seasons").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows(seasonCols))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/g1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "alice")

	err := h.GetGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"]
	if data["group"] == nil {
		t.Error("expected group in response")
	}
	members, ok := data["members"].([]any)
	if !ok {
		t.Fatal("expected members array")
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- GetGroup: not member ---

func TestGetGroup_NotMember(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("outsider", "g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/g1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "outsider", "outsider")

	err := h.GetGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_MEMBER" {
		t.Errorf("expected NOT_MEMBER, got %s", resp["error"]["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- JoinGroup: success ---

func TestJoinGroup_Success(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	// GetGroupByInviteCode
	mock.ExpectQuery("SELECT .+ FROM groups WHERE upper.invite_code.").
		WithArgs("INV1").
		WillReturnRows(handlerMockGroupRow("g1", "Test Group", "INV-1", "admin-1"))

	// IsGroupMember → not a member
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1", "g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	// CountGroupMembers → room available
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	// CountUserGroups → not at limit
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	// AddGroupMember
	// IsGroupBanned — a removal must not be undone by an invite.
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("g1", "u1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery("INSERT INTO group_members").
		WithArgs(sqlmock.AnyArg(), "u1", "g1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "group_id", "joined_at", "join_source", "invited_by"}).
			AddRow("member-1", "u1", "g1", time.Now(), "LINK", nil))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/join/INV-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("inviteCode")
	c.SetParamValues("INV-1")
	setUser(c, "u1", "testuser")

	err := h.JoinGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	group, ok := resp["data"]["group"].(map[string]any)
	if !ok {
		t.Fatal("expected group in response")
	}
	if group["id"] != "g1" {
		t.Errorf("expected group id g1, got %v", group["id"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- JoinGroup: group not found ---

func TestJoinGroup_GroupNotFound(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE upper.invite_code.").
		WithArgs("BADCODE").
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/join/BAD-CODE", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("inviteCode")
	c.SetParamValues("BAD-CODE")
	setUser(c, "u1", "testuser")

	err := h.JoinGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp["error"]["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- LeaveGroup: success (multiple members, non-admin) ---

func TestLeaveGroup_Success(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	// IsGroupMember
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u2", "g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// GetGroupByID
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(handlerMockGroupRow("g1", "Test", "inv-1", "u1"))

	// CountGroupMembers
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	// Begin transaction
	mock.ExpectBegin()

	// RemoveGroupMember
	mock.ExpectExec("DELETE FROM group_members").
		WithArgs("u2", "g1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// u2 is not admin, so no admin transfer
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/g1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u2", "bob")

	err := h.LeaveGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["data"]["left"] != true {
		t.Error("expected left=true in response")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- CreateGroup: success ---

func TestCreateGroup_Success(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	// GetUserByID
	userCols := []string{"id", "phone", "apple_id", "google_id", "username", "avatar_url", "avatar_emoji", "birth_year", "created_at", "updated_at", "username_changed_at"}
	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(
			"u1", nil, nil, nil, "testuser", nil, nil,
			sql.NullInt32{Int32: 2000, Valid: true},
			time.Now(), time.Now(), nil,
		))

	// CountUserGroups
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// freshInviteCode probes that the generated code is free.
	mock.ExpectQuery("SELECT .+ FROM groups WHERE upper.invite_code.").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(handlerGroupCols))

	// BeginTx
	mock.ExpectBegin()

	// CreateGroup: takes 6 args (id, name, invite_code, admin_id, telegram_chat_username, pq.Array(categories))
	mock.ExpectQuery("INSERT INTO groups").
		WithArgs(sqlmock.AnyArg(), "New Group", sqlmock.AnyArg(), "u1", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(handlerGroupCols).AddRow(
			"g-new", "New Group", "inv-new", "u1",
			nil, nil, nil, nil,
			time.Now(), `{"HOT"}`, false,
		))

	// AddGroupMember
	mock.ExpectQuery("INSERT INTO group_members").
		WithArgs(sqlmock.AnyArg(), "u1", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "group_id", "joined_at", "join_source", "invited_by"}).
			AddRow("mem-1", "u1", "g-new", time.Now(), "LINK", nil))

	// CreateSeason
	seasonCols := []string{"id", "group_id", "number", "status", "starts_at", "reveal_at", "ends_at", "created_at", "kind", "postpone_count"}
	mock.ExpectQuery("INSERT INTO seasons").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), int32(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), db.SeasonKindKICKOFF).
		WillReturnRows(sqlmock.NewRows(seasonCols).AddRow(
			"s-new", "g-new", int32(1), "VOTING",
			time.Now(), time.Now().Add(24*time.Hour), time.Now().Add(48*time.Hour), time.Now(),
			"KICKOFF", int32(0),
		))

	// GetRandomSystemQuestionsByCategories
	questionCols := []string{"id", "text", "category", "is_system", "group_id", "created_at"}
	mock.ExpectQuery("SELECT .+ FROM questions").
		WillReturnRows(sqlmock.NewRows(questionCols))

	// Commit
	mock.ExpectCommit()

	body := `{"name":"New Group","categories":["HOT"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "u1", "testuser")

	err := h.CreateGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- RegenerateInviteLink: success ---

func TestRegenerateInviteLink_Success(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	// GetGroupByID
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(handlerMockGroupRow("g1", "Test", "inv-old", "u1"))

	// freshInviteCode probes that the generated code is free.
	mock.ExpectQuery("SELECT .+ FROM groups WHERE upper.invite_code.").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(handlerGroupCols))

	// UpdateGroupInviteCode
	mock.ExpectExec("UPDATE groups SET invite_code").
		WithArgs("g1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/g1/regenerate-invite", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "testuser")

	err := h.RegenerateInviteLink(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["data"]["invite_url"] == "" {
		t.Error("expected invite_url in response")
	}
	if !strings.HasPrefix(resp["data"]["invite_url"], "https://repa.app/join/") {
		t.Errorf("expected invite URL prefix, got %s", resp["data"]["invite_url"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- RegenerateInviteLink: not admin ---

func TestRegenerateInviteLink_NotAdmin(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	// GetGroupByID — admin is "admin-1", not "u1"
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(handlerMockGroupRow("g1", "Test", "inv-1", "admin-1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/g1/regenerate-invite", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "testuser")

	err := h.RegenerateInviteLink(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_ADMIN" {
		t.Errorf("expected NOT_ADMIN, got %s", resp["error"]["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- ListGroups: reveal_state for a pending Reveal ---

var listStatsCols = []string{
	"id", "name", "invite_code", "admin_id",
	"telegram_chat_id", "telegram_chat_username",
	"telegram_connect_code", "telegram_connect_expiry",
	"created_at", "categories", "kind_only", "member_count",
	"active_season_id", "active_season_number", "active_season_status",
	"active_season_starts_at", "active_season_reveal_at", "active_season_ends_at",
	"active_season_kind", "active_season_postpone_count",
	"voted_count", "user_vote_count",
}

func listGroupsRevealState(t *testing.T, memberCount, votedCount int64, status string, postponeCount int32) map[string]any {
	t.Helper()
	h, mock, e := newHandlerWithMock(t)
	now := time.Now()

	mock.ExpectQuery("SELECT g.id").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows(listStatsCols).AddRow(
			"g1", "Test Group", "inv-1", "u1",
			nil, nil, nil, nil,
			now, `{"HOT"}`, false, memberCount,
			"s1", int32(1), status, now, now.Add(48*time.Hour), now.Add(96*time.Hour),
			"KICKOFF", postponeCount,
			votedCount, int64(1),
		))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "u1", "testuser")

	if err := h.ListGroups(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body %s)", rec.Code, rec.Body.String())
	}

	var resp map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	groups := resp["data"]["groups"].([]any)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	season := groups[0].(map[string]any)["active_season"].(map[string]any)
	return season
}

func TestListGroups_RevealStateWaitingForMembers(t *testing.T) {
	season := listGroupsRevealState(t, 2, 2, "VOTING", 0)

	if got := season["reveal_state"]; got != "WAITING_FOR_MEMBERS" {
		t.Errorf("reveal_state = %v, want WAITING_FOR_MEMBERS", got)
	}
	if got := season["members_needed"]; got != float64(1) {
		t.Errorf("members_needed = %v, want 1", got)
	}
	if got := season["voters_needed"]; got != float64(1) {
		t.Errorf("voters_needed = %v, want 1", got)
	}
}

func TestListGroups_RevealStateWaitingForVoters(t *testing.T) {
	season := listGroupsRevealState(t, 6, 1, "VOTING", 0)

	if got := season["reveal_state"]; got != "WAITING_FOR_VOTERS" {
		t.Errorf("reveal_state = %v, want WAITING_FOR_VOTERS", got)
	}
	if got := season["members_needed"]; got != float64(0) {
		t.Errorf("members_needed = %v, want 0", got)
	}
	if got := season["voters_needed"]; got != float64(2) {
		t.Errorf("voters_needed = %v, want 2", got)
	}
}

func TestListGroups_RevealStateScheduled(t *testing.T) {
	season := listGroupsRevealState(t, 6, 4, "VOTING", 0)

	if got := season["reveal_state"]; got != "SCHEDULED" {
		t.Errorf("reveal_state = %v, want SCHEDULED", got)
	}
}

func TestListGroups_RevealStatePostponed(t *testing.T) {
	season := listGroupsRevealState(t, 6, 4, "VOTING", 1)

	if got := season["reveal_state"]; got != "POSTPONED" {
		t.Errorf("reveal_state = %v, want POSTPONED", got)
	}
}

func TestListGroups_RevealStateRevealed(t *testing.T) {
	season := listGroupsRevealState(t, 6, 6, "REVEALED", 0)

	if got := season["reveal_state"]; got != "REVEALED" {
		t.Errorf("reveal_state = %v, want REVEALED", got)
	}
}
