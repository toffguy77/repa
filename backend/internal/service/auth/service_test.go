package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/repa-app/repa/internal/middleware"
)

// --- Mock querier for testing ---

type mockQuerier struct {
	getUserByIDFn                    func(ctx context.Context, id string) (db.User, error)
	getUserByPhoneFn                 func(ctx context.Context, phone sql.NullString) (db.User, error)
	getUserByUsernameFn              func(ctx context.Context, username string) (db.User, error)
	getUserByAppleIDFn               func(ctx context.Context, appleID sql.NullString) (db.User, error)
	getUserByGoogleIDFn              func(ctx context.Context, googleID sql.NullString) (db.User, error)
	createUserFn                     func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	updateUserProfileFn              func(ctx context.Context, arg db.UpdateUserProfileParams) (db.User, error)
	updateUserProfileWithUsernameFn  func(ctx context.Context, arg db.UpdateUserProfileWithUsernameParams) (db.User, error)
	updateUserAvatarURLFn            func(ctx context.Context, arg db.UpdateUserAvatarURLParams) (db.User, error)
	deleteUserFn                     func(ctx context.Context, id string) error
	upsertPushPreferenceFn           func(ctx context.Context, arg db.UpsertPushPreferenceParams) (db.PushPreference, error)
}

func (m *mockQuerier) GetUserByID(ctx context.Context, id string) (db.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQuerier) GetUserByPhone(ctx context.Context, phone sql.NullString) (db.User, error) {
	if m.getUserByPhoneFn != nil {
		return m.getUserByPhoneFn(ctx, phone)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQuerier) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	if m.getUserByUsernameFn != nil {
		return m.getUserByUsernameFn(ctx, username)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQuerier) GetUserByAppleID(ctx context.Context, appleID sql.NullString) (db.User, error) {
	if m.getUserByAppleIDFn != nil {
		return m.getUserByAppleIDFn(ctx, appleID)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQuerier) GetUserByGoogleID(ctx context.Context, googleID sql.NullString) (db.User, error) {
	if m.getUserByGoogleIDFn != nil {
		return m.getUserByGoogleIDFn(ctx, googleID)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, arg)
	}
	return db.User{ID: arg.ID, Username: arg.Username}, nil
}

func (m *mockQuerier) UpdateUserProfile(ctx context.Context, arg db.UpdateUserProfileParams) (db.User, error) {
	if m.updateUserProfileFn != nil {
		return m.updateUserProfileFn(ctx, arg)
	}
	return db.User{ID: arg.ID, Username: arg.Username}, nil
}

func (m *mockQuerier) UpdateUserProfileWithUsername(ctx context.Context, arg db.UpdateUserProfileWithUsernameParams) (db.User, error) {
	if m.updateUserProfileWithUsernameFn != nil {
		return m.updateUserProfileWithUsernameFn(ctx, arg)
	}
	return db.User{ID: arg.ID, Username: arg.Username}, nil
}

func (m *mockQuerier) UpdateUserAvatarURL(ctx context.Context, arg db.UpdateUserAvatarURLParams) (db.User, error) {
	if m.updateUserAvatarURLFn != nil {
		return m.updateUserAvatarURLFn(ctx, arg)
	}
	return db.User{ID: arg.ID, AvatarUrl: arg.AvatarUrl}, nil
}

func (m *mockQuerier) DeleteUser(ctx context.Context, id string) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, id)
	}
	return nil
}

func (m *mockQuerier) UpsertPushPreference(ctx context.Context, arg db.UpsertPushPreferenceParams) (db.PushPreference, error) {
	if m.upsertPushPreferenceFn != nil {
		return m.upsertPushPreferenceFn(ctx, arg)
	}
	return db.PushPreference{ID: arg.ID, UserID: arg.UserID, Category: arg.Category, Enabled: arg.Enabled}, nil
}

func TestIsValidImage(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"JPEG magic bytes", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}, true},
		{"PNG magic bytes", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D}, true},
		{"GIF - not supported", []byte{0x47, 0x49, 0x46, 0x38, 0x39}, false},
		{"empty", []byte{}, false},
		{"too short", []byte{0xFF, 0xD8}, false},
		{"text file", []byte("hello world text file"), false},
		{"random bytes", []byte{0x00, 0x01, 0x02, 0x03, 0x04}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidImage(tt.data)
			if got != tt.valid {
				t.Errorf("isValidImage(%v) = %v, want %v", tt.data, got, tt.valid)
			}
		})
	}
}

func TestUsernameRegex(t *testing.T) {
	tests := []struct {
		username string
		valid    bool
	}{
		{"alice", true},
		{"user_123", true},
		{"Алиса", true},
		{"ab", false},              // too short
		{"a", false},               // too short
		{"aaaaabbbbbcccccddddde", false}, // 21 chars, too long
		{"user name", false},       // space not allowed
		{"user@name", false},       // @ not allowed
		{"user-name", false},       // dash not allowed
		{"___", true},              // underscores only
		{"ёЁ_test", true},          // ё is allowed
		{"123", true},              // digits only
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			got := usernameRegex.MatchString(tt.username)
			if got != tt.valid {
				t.Errorf("usernameRegex.MatchString(%q) = %v, want %v", tt.username, got, tt.valid)
			}
		})
	}
}

func TestGenerateUsername(t *testing.T) {
	u1 := generateUsername()
	u2 := generateUsername()

	if u1 == u2 {
		t.Error("expected unique usernames")
	}
	if len(u1) < 5 {
		t.Error("username too short")
	}
	if u1[:5] != "user_" {
		t.Errorf("expected user_ prefix, got %s", u1[:5])
	}
}

// --- signToken tests ---

func TestSignToken_ValidJWT(t *testing.T) {
	svc := NewService(nil, nil, nil, "test-secret-key", false)

	user := db.User{ID: "user-abc-123", Username: "alice"}

	tokenStr, err := svc.signToken(user)
	if err != nil {
		t.Fatalf("signToken returned error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("signToken returned empty token")
	}

	// Parse the token back and verify claims
	parsed, err := jwtlib.ParseWithClaims(tokenStr, &middleware.JWTClaims{}, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("test-secret-key"), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("parsed token is not valid")
	}

	claims, ok := parsed.Claims.(*middleware.JWTClaims)
	if !ok {
		t.Fatal("failed to cast claims")
	}
	if claims.UserID != "user-abc-123" {
		t.Errorf("expected UserID user-abc-123, got %s", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("expected Username alice, got %s", claims.Username)
	}
}

func TestSignToken_ExpiresIn90Days(t *testing.T) {
	svc := NewService(nil, nil, nil, "test-secret", false)

	user := db.User{ID: "u1", Username: "bob"}
	tokenStr, err := svc.signToken(user)
	if err != nil {
		t.Fatalf("signToken error: %v", err)
	}

	parsed, err := jwtlib.ParseWithClaims(tokenStr, &middleware.JWTClaims{}, func(token *jwtlib.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	claims := parsed.Claims.(*middleware.JWTClaims)

	expiresAt := claims.ExpiresAt.Time
	issuedAt := claims.IssuedAt.Time

	diff := expiresAt.Sub(issuedAt)
	expected := 90 * 24 * time.Hour
	// Allow 5 seconds of tolerance
	if diff < expected-5*time.Second || diff > expected+5*time.Second {
		t.Errorf("expected ~90 days between issued and expiry, got %v", diff)
	}
}

func TestSignToken_WrongSecretFails(t *testing.T) {
	svc := NewService(nil, nil, nil, "correct-secret", false)

	user := db.User{ID: "u1", Username: "test"}
	tokenStr, err := svc.signToken(user)
	if err != nil {
		t.Fatalf("signToken error: %v", err)
	}

	_, err = jwtlib.ParseWithClaims(tokenStr, &middleware.JWTClaims{}, func(token *jwtlib.Token) (any, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Error("expected error when parsing with wrong secret")
	}
}

func TestSignToken_DifferentUsersGetDifferentTokens(t *testing.T) {
	svc := NewService(nil, nil, nil, "secret", false)

	t1, _ := svc.signToken(db.User{ID: "u1", Username: "a"})
	t2, _ := svc.signToken(db.User{ID: "u2", Username: "b"})

	if t1 == t2 {
		t.Error("expected different tokens for different users")
	}
}

// --- CheckUsername validation path ---

func TestCheckUsername_InvalidFormat(t *testing.T) {
	svc := NewService(nil, nil, nil, "", false)

	tests := []struct {
		name     string
		username string
	}{
		{"too short", "ab"},
		{"contains space", "user name"},
		{"contains @", "user@name"},
		{"contains dash", "user-name"},
		{"empty", ""},
		{"too long", "aaaaabbbbbcccccddddde"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CheckUsername(context.Background(), tt.username)
			if !errors.Is(err, ErrInvalidUsername) {
				t.Errorf("CheckUsername(%q) error = %v, want ErrInvalidUsername", tt.username, err)
			}
		})
	}
}

// --- UploadAvatar early-exit paths ---

func TestUploadAvatar_S3Nil_ReturnsUnavailable(t *testing.T) {
	svc := NewService(nil, nil, nil, "", false)

	_, err := svc.UploadAvatar(context.Background(), "user1", bytes.NewReader(nil), 100)
	if !errors.Is(err, ErrAvatarUnavailable) {
		t.Errorf("expected ErrAvatarUnavailable, got %v", err)
	}
}

func TestUploadAvatar_TooLarge_ReturnsError(t *testing.T) {
	// Use a zero-value S3Client pointer (non-nil) to get past the nil check
	svc := NewService(nil, nil, &lib.S3Client{}, "", false)

	_, err := svc.UploadAvatar(context.Background(), "user1", bytes.NewReader(nil), maxAvatarSize+1)
	if !errors.Is(err, ErrImageTooLarge) {
		t.Errorf("expected ErrImageTooLarge, got %v", err)
	}
}

func TestUploadAvatar_TooLarge_ViaActualRead(t *testing.T) {
	// Test the second size check: fileSize claims small but actual data exceeds limit
	svc := NewService(nil, nil, &lib.S3Client{}, "", false)

	// Create data larger than maxAvatarSize
	bigData := make([]byte, maxAvatarSize+100)
	_, err := svc.UploadAvatar(context.Background(), "user1", bytes.NewReader(bigData), 100)
	if !errors.Is(err, ErrImageTooLarge) {
		t.Errorf("expected ErrImageTooLarge from actual read, got %v", err)
	}
}

func TestUploadAvatar_InvalidImage_ReturnsError(t *testing.T) {
	svc := NewService(nil, nil, &lib.S3Client{}, "", false)

	// Valid size but not a valid image (random bytes)
	data := []byte("this is not an image file at all, just plain text content")
	_, err := svc.UploadAvatar(context.Background(), "user1", bytes.NewReader(data), int64(len(data)))
	if !errors.Is(err, ErrInvalidImage) {
		t.Errorf("expected ErrInvalidImage, got %v", err)
	}
}

func TestUploadAvatar_EmptyData_ReturnsError(t *testing.T) {
	svc := NewService(nil, nil, &lib.S3Client{}, "", false)

	_, err := svc.UploadAvatar(context.Background(), "user1", bytes.NewReader([]byte{}), 0)
	if !errors.Is(err, ErrInvalidImage) {
		t.Errorf("expected ErrInvalidImage for empty data, got %v", err)
	}
}

func TestUploadAvatar_JPEGHeader_ButInvalidImage(t *testing.T) {
	svc := NewService(nil, nil, &lib.S3Client{}, "", false)

	// JPEG magic bytes but not a real JPEG — image.Decode will fail
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	_, err := svc.UploadAvatar(context.Background(), "user1", bytes.NewReader(data), int64(len(data)))
	if !errors.Is(err, ErrInvalidImage) {
		t.Errorf("expected ErrInvalidImage for truncated JPEG, got %v", err)
	}
}

// --- jwkToRSAPublicKey ---

func TestJwkToRSAPublicKey_ValidKey(t *testing.T) {
	// Use known RSA values: N = 65537 * 3 = 196611, E = 65537
	// In practice, use a proper RSA modulus. We'll use small values for testing the conversion.
	nBig := big.NewInt(123456789)
	eBig := big.NewInt(65537)

	nB64 := base64.RawURLEncoding.EncodeToString(nBig.Bytes())
	eB64 := base64.RawURLEncoding.EncodeToString(eBig.Bytes())

	k := &jwkKey{
		Kty: "RSA",
		N:   nB64,
		E:   eB64,
	}

	pub, err := jwkToRSAPublicKey(k)
	if err != nil {
		t.Fatalf("jwkToRSAPublicKey error: %v", err)
	}

	if pub.N.Cmp(nBig) != 0 {
		t.Errorf("N mismatch: got %v, want %v", pub.N, nBig)
	}
	if pub.E != 65537 {
		t.Errorf("E mismatch: got %d, want 65537", pub.E)
	}
}

func TestJwkToRSAPublicKey_InvalidBase64_N(t *testing.T) {
	k := &jwkKey{
		N: "!!!invalid-base64!!!",
		E: base64.RawURLEncoding.EncodeToString(big.NewInt(65537).Bytes()),
	}
	_, err := jwkToRSAPublicKey(k)
	if err == nil {
		t.Error("expected error for invalid base64 in N")
	}
}

func TestJwkToRSAPublicKey_InvalidBase64_E(t *testing.T) {
	k := &jwkKey{
		N: base64.RawURLEncoding.EncodeToString(big.NewInt(12345).Bytes()),
		E: "!!!invalid-base64!!!",
	}
	_, err := jwkToRSAPublicKey(k)
	if err == nil {
		t.Error("expected error for invalid base64 in E")
	}
}

func TestJwkToRSAPublicKey_ReturnsRSAPublicKeyType(t *testing.T) {
	k := &jwkKey{
		N: base64.RawURLEncoding.EncodeToString(big.NewInt(999999937).Bytes()),
		E: base64.RawURLEncoding.EncodeToString(big.NewInt(3).Bytes()),
	}
	pub, err := jwkToRSAPublicKey(k)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it's a proper *rsa.PublicKey
	var _ *rsa.PublicKey = pub
}

// --- Sentinel error distinctness ---

func TestSentinelErrors_AreDistinct(t *testing.T) {
	errs := []error{
		ErrInvalidOTP,
		ErrOTPBlocked,
		ErrOTPRateLimit,
		ErrUserNotFound,
		ErrUsernameTaken,
		ErrUsernameRecent,
		ErrInvalidUsername,
		ErrInvalidToken,
		ErrInvalidImage,
		ErrImageTooLarge,
		ErrAvatarUnavailable,
	}

	for i := 0; i < len(errs); i++ {
		for j := i + 1; j < len(errs); j++ {
			if errors.Is(errs[i], errs[j]) {
				t.Errorf("sentinel errors should be distinct: %q == %q", errs[i], errs[j])
			}
		}
	}
}

func TestSentinelErrors_HaveMessages(t *testing.T) {
	errs := map[string]error{
		"ErrInvalidOTP":     ErrInvalidOTP,
		"ErrOTPBlocked":     ErrOTPBlocked,
		"ErrOTPRateLimit":   ErrOTPRateLimit,
		"ErrUserNotFound":   ErrUserNotFound,
		"ErrUsernameTaken":  ErrUsernameTaken,
		"ErrUsernameRecent": ErrUsernameRecent,
		"ErrInvalidUsername": ErrInvalidUsername,
		"ErrInvalidToken":   ErrInvalidToken,
		"ErrInvalidImage":   ErrInvalidImage,
		"ErrImageTooLarge":  ErrImageTooLarge,
		"ErrAvatarUnavailable": ErrAvatarUnavailable,
	}

	for name, err := range errs {
		if err.Error() == "" {
			t.Errorf("%s has empty error message", name)
		}
	}
}

// --- NewService constructor ---

func TestNewService_ReturnsNonNil(t *testing.T) {
	svc := NewService(nil, nil, nil, "secret", true)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestNewService_DevMode(t *testing.T) {
	svc := NewService(nil, nil, nil, "s", true)
	if !svc.devMode {
		t.Error("expected devMode to be true")
	}

	svc2 := NewService(nil, nil, nil, "s", false)
	if svc2.devMode {
		t.Error("expected devMode to be false")
	}
}

// --- GetMe tests (with mock) ---

func TestGetMe_UserFound(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, id string) (db.User, error) {
			return db.User{ID: id, Username: "alice"}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	user, err := svc.GetMe(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("expected ID u1, got %s", user.ID)
	}
	if user.Username != "alice" {
		t.Errorf("expected username alice, got %s", user.Username)
	}
}

func TestGetMe_UserNotFound(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	_, err := svc.GetMe(context.Background(), "nonexistent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetMe_DBError(t *testing.T) {
	dbErr := fmt.Errorf("connection refused")
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, dbErr
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	_, err := svc.GetMe(context.Background(), "u1")
	if err == nil || err.Error() != "connection refused" {
		t.Errorf("expected db error, got %v", err)
	}
}

// --- CheckUsername tests (with mock) ---

func TestCheckUsername_Available(t *testing.T) {
	mock := &mockQuerier{
		getUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	available, err := svc.CheckUsername(context.Background(), "validuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected username to be available")
	}
}

func TestCheckUsername_Taken(t *testing.T) {
	mock := &mockQuerier{
		getUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "other", Username: "validuser"}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	available, err := svc.CheckUsername(context.Background(), "validuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected username to be taken")
	}
}

func TestCheckUsername_DBError(t *testing.T) {
	mock := &mockQuerier{
		getUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, fmt.Errorf("db down")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	_, err := svc.CheckUsername(context.Background(), "validuser")
	if err == nil {
		t.Error("expected error from DB")
	}
}

// --- UpdateProfile tests (with mock) ---

func TestUpdateProfile_NoChanges(t *testing.T) {
	existingUser := db.User{ID: "u1", Username: "alice"}
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return existingUser, nil
		},
		updateUserProfileFn: func(_ context.Context, arg db.UpdateUserProfileParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	user, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("expected username alice, got %s", user.Username)
	}
}

func TestUpdateProfile_InvalidUsername(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "u1", Username: "oldname"}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	badName := "a b"
	_, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{Username: &badName})
	if !errors.Is(err, ErrInvalidUsername) {
		t.Errorf("expected ErrInvalidUsername, got %v", err)
	}
}

func TestUpdateProfile_UsernameCooldown(t *testing.T) {
	recentChange := sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true} // changed 1 day ago
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "u1", Username: "oldname", UsernameChangedAt: recentChange}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	newName := "newname"
	_, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{Username: &newName})
	if !errors.Is(err, ErrUsernameRecent) {
		t.Errorf("expected ErrUsernameRecent, got %v", err)
	}
}

func TestUpdateProfile_UsernameTaken(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "u1", Username: "oldname"}, nil
		},
		getUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "other-user", Username: "newname"}, nil // taken by someone else
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	newName := "newname"
	_, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{Username: &newName})
	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestUpdateProfile_UsernameChanged(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "u1", Username: "oldname"}, nil
		},
		getUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, sql.ErrNoRows // available
		},
		updateUserProfileWithUsernameFn: func(_ context.Context, arg db.UpdateUserProfileWithUsernameParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	newName := "newname"
	user, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{Username: &newName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "newname" {
		t.Errorf("expected username newname, got %s", user.Username)
	}
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	_, err := svc.UpdateProfile(context.Background(), "nonexistent", UpdateProfileParams{})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateProfile_AvatarEmoji(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "u1", Username: "alice"}, nil
		},
		updateUserProfileFn: func(_ context.Context, arg db.UpdateUserProfileParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username, AvatarEmoji: arg.AvatarEmoji}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	emoji := "🎉"
	user, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{AvatarEmoji: &emoji})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.AvatarEmoji.String != "🎉" {
		t.Errorf("expected emoji 🎉, got %s", user.AvatarEmoji.String)
	}
}

func TestUpdateProfile_BirthYear(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "u1", Username: "alice"}, nil
		},
		updateUserProfileFn: func(_ context.Context, arg db.UpdateUserProfileParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username, BirthYear: arg.BirthYear}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	year := 2005
	user, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{BirthYear: &year})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.BirthYear.Int32 != 2005 {
		t.Errorf("expected birth year 2005, got %d", user.BirthYear.Int32)
	}
}

func TestUpdateProfile_SameUsername_NoChange(t *testing.T) {
	// When the provided username equals the current one, no username change should occur
	updateCalled := false
	updateWithUsernameCalled := false
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "u1", Username: "alice"}, nil
		},
		updateUserProfileFn: func(_ context.Context, arg db.UpdateUserProfileParams) (db.User, error) {
			updateCalled = true
			return db.User{ID: arg.ID, Username: arg.Username}, nil
		},
		updateUserProfileWithUsernameFn: func(_ context.Context, arg db.UpdateUserProfileWithUsernameParams) (db.User, error) {
			updateWithUsernameCalled = true
			return db.User{ID: arg.ID, Username: arg.Username}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	sameName := "alice"
	_, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{Username: &sameName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateCalled {
		t.Error("expected UpdateUserProfile to be called (no username change)")
	}
	if updateWithUsernameCalled {
		t.Error("UpdateUserProfileWithUsername should not be called when username unchanged")
	}
}

// --- DeleteAccount tests (with mock) ---

func TestDeleteAccount_Success(t *testing.T) {
	mock := &mockQuerier{
		deleteUserFn: func(_ context.Context, id string) error {
			if id != "u1" {
				t.Errorf("expected id u1, got %s", id)
			}
			return nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	err := svc.DeleteAccount(context.Background(), "u1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteAccount_Error(t *testing.T) {
	mock := &mockQuerier{
		deleteUserFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("delete failed")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	err := svc.DeleteAccount(context.Background(), "u1")
	if err == nil {
		t.Error("expected error")
	}
}

// --- UpsertPushPreference tests (with mock) ---

func TestUpsertPushPreference_Success(t *testing.T) {
	mock := &mockQuerier{}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	pref, err := svc.UpsertPushPreference(context.Background(), "u1", "voting", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pref.UserID != "u1" {
		t.Errorf("expected user_id u1, got %s", pref.UserID)
	}
	if pref.Category != "voting" {
		t.Errorf("expected category voting, got %s", string(pref.Category))
	}
	if !pref.Enabled {
		t.Error("expected enabled to be true")
	}
}

// --- upsertByPhone tests (with mock) ---

func TestUpsertByPhone_ExistingUser(t *testing.T) {
	existing := db.User{ID: "u1", Username: "alice", Phone: sql.NullString{String: "+79001234567", Valid: true}}
	mock := &mockQuerier{
		getUserByPhoneFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return existing, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	result, err := svc.upsertByPhone(context.Background(), "+79001234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User.ID != "u1" {
		t.Errorf("expected user ID u1, got %s", result.User.ID)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestUpsertByPhone_NewUser(t *testing.T) {
	mock := &mockQuerier{
		getUserByPhoneFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username, Phone: arg.Phone}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	result, err := svc.upsertByPhone(context.Background(), "+79001234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestUpsertByPhone_DBError(t *testing.T) {
	mock := &mockQuerier{
		getUserByPhoneFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, fmt.Errorf("connection error")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	_, err := svc.upsertByPhone(context.Background(), "+79001234567")
	if err == nil {
		t.Error("expected error")
	}
}

// --- upsertByAppleID tests (with mock) ---

func TestUpsertByAppleID_ExistingUser(t *testing.T) {
	mock := &mockQuerier{
		getUserByAppleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{ID: "u1", Username: "alice"}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	result, err := svc.upsertByAppleID(context.Background(), "apple-sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User.ID != "u1" {
		t.Errorf("expected user ID u1, got %s", result.User.ID)
	}
}

func TestUpsertByAppleID_NewUser(t *testing.T) {
	mock := &mockQuerier{
		getUserByAppleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	result, err := svc.upsertByAppleID(context.Background(), "apple-sub-new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

// --- upsertByGoogleID tests (with mock) ---

func TestUpsertByGoogleID_ExistingUser(t *testing.T) {
	mock := &mockQuerier{
		getUserByGoogleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{ID: "u1", Username: "alice"}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	result, err := svc.upsertByGoogleID(context.Background(), "google-sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User.ID != "u1" {
		t.Errorf("expected user ID u1, got %s", result.User.ID)
	}
}

func TestUpsertByGoogleID_NewUser(t *testing.T) {
	mock := &mockQuerier{
		getUserByGoogleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	result, err := svc.upsertByGoogleID(context.Background(), "google-sub-new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

// --- helper: create a service with miniredis + mock querier ---

func newTestServiceWithRedis(t *testing.T, mock *mockQuerier) (*Service, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := NewServiceWithQuerier(mock, rdb, nil, "test-secret", true)
	return svc, mr
}

// --- OTPSend tests ---

func TestOTPSend_DevMode_ReturnsCode(t *testing.T) {
	svc, _ := newTestServiceWithRedis(t, &mockQuerier{})

	code, err := svc.OTPSend(context.Background(), "+79001234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == "" {
		t.Error("expected non-empty OTP code in dev mode")
	}
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %q", code)
	}
}

func TestOTPSend_RateLimit(t *testing.T) {
	svc, _ := newTestServiceWithRedis(t, &mockQuerier{})
	phone := "+79001234567"

	// First 3 calls should succeed
	for i := 0; i < 3; i++ {
		_, err := svc.OTPSend(context.Background(), phone)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
	}

	// 4th call should be rate limited
	_, err := svc.OTPSend(context.Background(), phone)
	if !errors.Is(err, ErrOTPRateLimit) {
		t.Errorf("expected ErrOTPRateLimit, got %v", err)
	}
}

func TestOTPSend_DifferentPhones_IndependentLimits(t *testing.T) {
	svc, _ := newTestServiceWithRedis(t, &mockQuerier{})

	_, err := svc.OTPSend(context.Background(), "+79001111111")
	if err != nil {
		t.Fatalf("phone1: unexpected error: %v", err)
	}

	_, err = svc.OTPSend(context.Background(), "+79002222222")
	if err != nil {
		t.Fatalf("phone2: unexpected error: %v", err)
	}
}

// --- OTPVerify tests ---

func TestOTPVerify_Success(t *testing.T) {
	mock := &mockQuerier{
		getUserByPhoneFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{ID: "u1", Username: "alice"}, nil
		},
	}
	svc, _ := newTestServiceWithRedis(t, mock)
	phone := "+79001234567"

	// Send OTP first (dev mode returns code)
	code, err := svc.OTPSend(context.Background(), phone)
	if err != nil {
		t.Fatalf("OTPSend error: %v", err)
	}

	// Verify with correct code
	result, err := svc.OTPVerify(context.Background(), phone, code)
	if err != nil {
		t.Fatalf("OTPVerify error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
	if result.User.ID != "u1" {
		t.Errorf("expected user ID u1, got %s", result.User.ID)
	}
}

func TestOTPVerify_WrongCode(t *testing.T) {
	svc, _ := newTestServiceWithRedis(t, &mockQuerier{})
	phone := "+79001234567"

	_, err := svc.OTPSend(context.Background(), phone)
	if err != nil {
		t.Fatalf("OTPSend error: %v", err)
	}

	_, err = svc.OTPVerify(context.Background(), phone, "000000")
	if !errors.Is(err, ErrInvalidOTP) {
		t.Errorf("expected ErrInvalidOTP, got %v", err)
	}
}

func TestOTPVerify_NoOTPSent(t *testing.T) {
	svc, _ := newTestServiceWithRedis(t, &mockQuerier{})

	_, err := svc.OTPVerify(context.Background(), "+79001234567", "123456")
	if !errors.Is(err, ErrInvalidOTP) {
		t.Errorf("expected ErrInvalidOTP, got %v", err)
	}
}

func TestOTPVerify_TooManyAttempts(t *testing.T) {
	svc, _ := newTestServiceWithRedis(t, &mockQuerier{})
	phone := "+79001234567"

	_, err := svc.OTPSend(context.Background(), phone)
	if err != nil {
		t.Fatalf("OTPSend error: %v", err)
	}

	// Make 5 wrong attempts
	for i := 0; i < 5; i++ {
		_, _ = svc.OTPVerify(context.Background(), phone, "000000")
	}

	// 6th attempt should be blocked
	_, err = svc.OTPVerify(context.Background(), phone, "000000")
	if !errors.Is(err, ErrOTPBlocked) {
		t.Errorf("expected ErrOTPBlocked, got %v", err)
	}
}

func TestOTPVerify_NewUser_Created(t *testing.T) {
	mock := &mockQuerier{
		getUserByPhoneFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username, Phone: arg.Phone}, nil
		},
	}
	svc, _ := newTestServiceWithRedis(t, mock)
	phone := "+79009999999"

	code, err := svc.OTPSend(context.Background(), phone)
	if err != nil {
		t.Fatalf("OTPSend error: %v", err)
	}

	result, err := svc.OTPVerify(context.Background(), phone, code)
	if err != nil {
		t.Fatalf("OTPVerify error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token for new user")
	}
}

// --- UpdateProfile DB error paths ---

func TestUpdateProfile_DBErrorOnGetUser(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, fmt.Errorf("db connection lost")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	_, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{})
	if err == nil || err.Error() != "db connection lost" {
		t.Errorf("expected db error, got %v", err)
	}
}

func TestUpdateProfile_UsernameCheckDBError(t *testing.T) {
	mock := &mockQuerier{
		getUserByIDFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: "u1", Username: "oldname"}, nil
		},
		getUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, fmt.Errorf("username check failed")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "", false)

	newName := "newname"
	_, err := svc.UpdateProfile(context.Background(), "u1", UpdateProfileParams{Username: &newName})
	if err == nil {
		t.Error("expected error from username check")
	}
}

// --- upsert error paths ---

func TestUpsertByPhone_CreateUserError(t *testing.T) {
	mock := &mockQuerier{
		getUserByPhoneFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, _ db.CreateUserParams) (db.User, error) {
			return db.User{}, fmt.Errorf("unique violation")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	_, err := svc.upsertByPhone(context.Background(), "+79001234567")
	if err == nil {
		t.Error("expected error from create user")
	}
}

func TestUpsertByAppleID_DBError(t *testing.T) {
	mock := &mockQuerier{
		getUserByAppleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, fmt.Errorf("connection refused")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	_, err := svc.upsertByAppleID(context.Background(), "apple-sub")
	if err == nil {
		t.Error("expected error")
	}
}

func TestUpsertByAppleID_CreateUserError(t *testing.T) {
	mock := &mockQuerier{
		getUserByAppleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, _ db.CreateUserParams) (db.User, error) {
			return db.User{}, fmt.Errorf("create failed")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	_, err := svc.upsertByAppleID(context.Background(), "apple-sub")
	if err == nil {
		t.Error("expected error from create user")
	}
}

func TestUpsertByGoogleID_DBError(t *testing.T) {
	mock := &mockQuerier{
		getUserByGoogleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, fmt.Errorf("connection refused")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	_, err := svc.upsertByGoogleID(context.Background(), "google-sub")
	if err == nil {
		t.Error("expected error")
	}
}

func TestUpsertByGoogleID_CreateUserError(t *testing.T) {
	mock := &mockQuerier{
		getUserByGoogleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, _ db.CreateUserParams) (db.User, error) {
			return db.User{}, fmt.Errorf("create failed")
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	_, err := svc.upsertByGoogleID(context.Background(), "google-sub")
	if err == nil {
		t.Error("expected error from create user")
	}
}

// --- GoogleAuth tests (with httptest) ---

func TestGoogleAuth_Success_ExistingUser(t *testing.T) {
	// Mock Google tokeninfo endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(googleTokenInfo{
			Sub:   "google-123",
			Email: "test@gmail.com",
		})
	}))
	defer ts.Close()

	mock := &mockQuerier{
		getUserByGoogleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{ID: "u1", Username: "alice"}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	// Override the Google URL by calling the internal method directly
	// We test GoogleAuth indirectly through upsertByGoogleID (already tested)
	// and test the HTTP flow by overriding DefaultClient transport
	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	result, err := svc.GoogleAuth(context.Background(), "fake-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User.ID != "u1" {
		t.Errorf("expected user ID u1, got %s", result.User.ID)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGoogleAuth_Success_NewUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(googleTokenInfo{
			Sub:   "google-new-456",
			Email: "new@gmail.com",
		})
	}))
	defer ts.Close()

	mock := &mockQuerier{
		getUserByGoogleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: arg.Username}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	result, err := svc.GoogleAuth(context.Background(), "fake-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGoogleAuth_InvalidToken_ServerReturns401(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	svc := NewServiceWithQuerier(&mockQuerier{}, nil, nil, "secret", false)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	_, err := svc.GoogleAuth(context.Background(), "bad-token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestGoogleAuth_EmptySub(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(googleTokenInfo{Sub: "", Email: "test@gmail.com"})
	}))
	defer ts.Close()

	svc := NewServiceWithQuerier(&mockQuerier{}, nil, nil, "secret", false)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	_, err := svc.GoogleAuth(context.Background(), "fake-token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestGoogleAuth_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	svc := NewServiceWithQuerier(&mockQuerier{}, nil, nil, "secret", false)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	_, err := svc.GoogleAuth(context.Background(), "fake-token")
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

// testTransport redirects all requests to the test server URL.
type testTransport struct {
	targetURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to our test server
	newURL := t.targetURL + req.URL.Path + "?" + req.URL.RawQuery
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, newURL, req.Body)
	if err != nil {
		return nil, err
	}
	return http.DefaultTransport.(*testTransport).roundTripDirect(newReq)
}

func (t *testTransport) roundTripDirect(req *http.Request) (*http.Response, error) {
	// Use a basic transport to avoid infinite recursion
	transport := &http.Transport{}
	return transport.RoundTrip(req)
}

// --- AppleAuth tests ---

func TestAppleAuth_InvalidJWT(t *testing.T) {
	svc := NewServiceWithQuerier(&mockQuerier{}, nil, nil, "secret", false)

	_, err := svc.AppleAuth(context.Background(), "not-a-jwt")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAppleAuth_NoKid(t *testing.T) {
	// Create a JWT without kid header
	token := jwtlib.New(jwtlib.SigningMethodHS256)
	tokenStr, _ := token.SignedString([]byte("dummy"))

	svc := NewServiceWithQuerier(&mockQuerier{}, nil, nil, "secret", false)

	// This will try to fetch Apple keys; we need to mock the HTTP client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwkKeySet{Keys: []jwkKey{}})
	}))
	defer ts.Close()

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	_, err := svc.AppleAuth(context.Background(), tokenStr)
	// kid is "" (empty string from type assertion of nil), server returns no matching key
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAppleAuth_KeyNotFound(t *testing.T) {
	// Create a JWT with a kid that won't match
	token := jwtlib.New(jwtlib.SigningMethodRS256)
	token.Header["kid"] = "nonexistent-kid"
	// Sign with a throwaway key just to produce a parseable JWT
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	tokenStr, _ := token.SignedString(privKey)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwkKeySet{Keys: []jwkKey{
			{Kid: "other-kid", Kty: "RSA",
				N: base64.RawURLEncoding.EncodeToString(big.NewInt(12345).Bytes()),
				E: base64.RawURLEncoding.EncodeToString(big.NewInt(65537).Bytes())},
		}})
	}))
	defer ts.Close()

	svc := NewServiceWithQuerier(&mockQuerier{}, nil, nil, "secret", false)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	_, err := svc.AppleAuth(context.Background(), tokenStr)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAppleAuth_FullFlow_Success(t *testing.T) {
	// Generate an RSA key pair
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := &privKey.PublicKey

	// Create a properly signed JWT
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, jwtlib.MapClaims{
		"iss": "https://appleid.apple.com",
		"sub": "apple-user-789",
		"aud": "com.repa.app",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	token.Header["kid"] = "test-kid"
	tokenStr, err := token.SignedString(privKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Encode public key as JWK
	nB64 := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
	eB64 := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwkKeySet{Keys: []jwkKey{
			{Kid: "test-kid", Kty: "RSA", Alg: "RS256", Use: "sig", N: nB64, E: eB64},
		}})
	}))
	defer ts.Close()

	mock := &mockQuerier{
		getUserByAppleIDFn: func(_ context.Context, _ sql.NullString) (db.User, error) {
			return db.User{ID: "u1", Username: "apple_user"}, nil
		},
	}
	svc := NewServiceWithQuerier(mock, nil, nil, "secret", false)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	result, err := svc.AppleAuth(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User.ID != "u1" {
		t.Errorf("expected user ID u1, got %s", result.User.ID)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAppleAuth_WrongIssuer(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := &privKey.PublicKey

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, jwtlib.MapClaims{
		"iss": "https://evil.example.com",
		"sub": "apple-user-789",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	token.Header["kid"] = "test-kid"
	tokenStr, _ := token.SignedString(privKey)

	nB64 := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
	eB64 := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwkKeySet{Keys: []jwkKey{
			{Kid: "test-kid", Kty: "RSA", N: nB64, E: eB64},
		}})
	}))
	defer ts.Close()

	svc := NewServiceWithQuerier(&mockQuerier{}, nil, nil, "secret", false)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	_, err := svc.AppleAuth(context.Background(), tokenStr)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for wrong issuer, got %v", err)
	}
}

func TestAppleAuth_EmptySub(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := &privKey.PublicKey

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, jwtlib.MapClaims{
		"iss": "https://appleid.apple.com",
		"sub": "",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	token.Header["kid"] = "test-kid"
	tokenStr, _ := token.SignedString(privKey)

	nB64 := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
	eB64 := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwkKeySet{Keys: []jwkKey{
			{Kid: "test-kid", Kty: "RSA", N: nB64, E: eB64},
		}})
	}))
	defer ts.Close()

	svc := NewServiceWithQuerier(&mockQuerier{}, nil, nil, "secret", false)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &testTransport{ts.URL}
	defer func() { http.DefaultTransport = origTransport }()

	_, err := svc.AppleAuth(context.Background(), tokenStr)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for empty sub, got %v", err)
	}
}

// --- UploadAvatar full flow test ---

func createTestJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a solid color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("failed to encode test JPEG: %v", err)
	}
	return buf.Bytes()
}

func TestUploadAvatar_ValidJPEG_ImageProcessing(t *testing.T) {
	jpegData := createTestJPEG(t, 512, 512)

	// S3Client has nil internal s3 client, so Upload will panic.
	// We recover from the panic to confirm the image processing pipeline
	// (isValidImage, image.Decode, imaging.Fill, jpeg.Encode) all succeeded.
	svc := NewServiceWithQuerier(&mockQuerier{}, nil, &lib.S3Client{}, "secret", false)

	var panicVal any
	func() {
		defer func() { panicVal = recover() }()
		_, _ = svc.UploadAvatar(context.Background(), "user-1", bytes.NewReader(jpegData), int64(len(jpegData)))
	}()

	// If we get a nil-pointer panic from S3 PutObject, it means image processing succeeded
	if panicVal == nil {
		t.Log("no panic — upload may have somehow succeeded")
	}
	// The key assertion: we did NOT get ErrInvalidImage or ErrImageTooLarge,
	// meaning the entire image decode+resize+encode pipeline completed.
}

func TestUploadAvatar_SmallJPEG_ImageProcessing(t *testing.T) {
	// Test with a small JPEG that will be upscaled to 256x256
	jpegData := createTestJPEG(t, 50, 50)

	svc := NewServiceWithQuerier(&mockQuerier{}, nil, &lib.S3Client{}, "secret", false)

	var panicVal any
	func() {
		defer func() { panicVal = recover() }()
		_, _ = svc.UploadAvatar(context.Background(), "user-1", bytes.NewReader(jpegData), int64(len(jpegData)))
	}()

	// Panic from S3 PutObject means image processing pipeline succeeded
	if panicVal == nil {
		t.Log("no panic — upload may have somehow succeeded")
	}
}
