package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const testSecret = "test-secret-key"

func makeToken(secret string, claims JWTClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return s
}

func makeValidClaims(userID, username string) JWTClaims {
	return JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func parseErrorResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var body struct {
		Error map[string]string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return body.Error
}

func TestJWTAuth_NoAuthorizationHeader(t *testing.T) {
	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	errBody := parseErrorResponse(t, rec)
	if errBody["message"] != "Missing authorization header" {
		t.Errorf("expected 'Missing authorization header', got %q", errBody["message"])
	}
	if errBody["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code 'UNAUTHORIZED', got %q", errBody["code"])
	}
}

func TestJWTAuth_InvalidFormat_NoBearerPrefix(t *testing.T) {
	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic some-token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	errBody := parseErrorResponse(t, rec)
	if errBody["message"] != "Invalid authorization format" {
		t.Errorf("expected 'Invalid authorization format', got %q", errBody["message"])
	}
}

func TestJWTAuth_InvalidFormat_BearerWithoutToken(t *testing.T) {
	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	errBody := parseErrorResponse(t, rec)
	if errBody["message"] != "Invalid authorization format" {
		t.Errorf("expected 'Invalid authorization format', got %q", errBody["message"])
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	claims := makeValidClaims("user-123", "testuser")
	tokenStr := makeToken(testSecret, claims)

	var capturedClaims *JWTClaims
	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		capturedClaims = GetCurrentUser(c)
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedClaims == nil {
		t.Fatal("expected claims to be set in context, got nil")
	}
	if capturedClaims.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got %q", capturedClaims.UserID)
	}
	if capturedClaims.Username != "testuser" {
		t.Errorf("expected Username 'testuser', got %q", capturedClaims.Username)
	}
}

func TestJWTAuth_ValidToken_CaseInsensitiveBearer(t *testing.T) {
	claims := makeValidClaims("user-456", "otheruser")
	tokenStr := makeToken(testSecret, claims)

	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "BEARER "+tokenStr)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with uppercase BEARER, got %d", rec.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	claims := JWTClaims{
		UserID:   "user-123",
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tokenStr := makeToken(testSecret, claims)

	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", rec.Code)
	}
	errBody := parseErrorResponse(t, rec)
	if errBody["message"] != "Invalid or expired token" {
		t.Errorf("expected 'Invalid or expired token', got %q", errBody["message"])
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	claims := makeValidClaims("user-123", "testuser")
	tokenStr := makeToken("wrong-secret", claims)

	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong secret, got %d", rec.Code)
	}
	errBody := parseErrorResponse(t, rec)
	if errBody["message"] != "Invalid or expired token" {
		t.Errorf("expected 'Invalid or expired token', got %q", errBody["message"])
	}
}

func TestJWTAuth_NonHMACSigningMethod(t *testing.T) {
	// Create a token with "none" signing method (alg: none attack)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, makeValidClaims("user-123", "testuser"))
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for non-HMAC signing method, got %d", rec.Code)
	}
}

func TestJWTAuth_MalformedToken(t *testing.T) {
	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for malformed token, got %d", rec.Code)
	}
	errBody := parseErrorResponse(t, rec)
	if errBody["message"] != "Invalid or expired token" {
		t.Errorf("expected 'Invalid or expired token', got %q", errBody["message"])
	}
}

func TestJWTAuth_EmptyAuthorizationHeader(t *testing.T) {
	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, JWTAuth(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Empty string header is treated as missing by Go's http
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetCurrentUser_ReturnsClaims(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expected := &JWTClaims{
		UserID:   "user-789",
		Username: "alice",
	}
	c.Set("user", expected)

	result := GetCurrentUser(c)
	if result == nil {
		t.Fatal("expected claims, got nil")
	}
	if result.UserID != "user-789" {
		t.Errorf("expected UserID 'user-789', got %q", result.UserID)
	}
	if result.Username != "alice" {
		t.Errorf("expected Username 'alice', got %q", result.Username)
	}
}

func TestGetCurrentUser_ReturnsNilWhenNotSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	result := GetCurrentUser(c)
	if result != nil {
		t.Errorf("expected nil when no claims set, got %+v", result)
	}
}

func TestGetCurrentUser_ReturnsNilForWrongType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("user", "not-a-claims-struct")

	result := GetCurrentUser(c)
	if result != nil {
		t.Errorf("expected nil for wrong type, got %+v", result)
	}
}
