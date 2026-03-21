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
	"image/jpeg"
	_ "image/png"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/disintegration/imaging"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/repa-app/repa/internal/middleware"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9_]{3,20}$`)

var (
	ErrInvalidOTP     = errors.New("invalid OTP code")
	ErrOTPBlocked     = errors.New("too many OTP attempts")
	ErrOTPRateLimit   = errors.New("OTP rate limit exceeded")
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameTaken  = errors.New("username already taken")
	ErrUsernameRecent = errors.New("username changed too recently")
	ErrInvalidUsername = errors.New("invalid username format")
	ErrInvalidToken   = errors.New("invalid token")
	ErrInvalidImage      = errors.New("invalid image file")
	ErrImageTooLarge     = errors.New("image too large")
	ErrAvatarUnavailable = errors.New("avatar uploads not configured")
)

type Service struct {
	queries   *db.Queries
	rdb       *redis.Client
	s3        *lib.S3Client
	jwtSecret string
	devMode   bool
}

func NewService(queries *db.Queries, rdb *redis.Client, s3 *lib.S3Client, jwtSecret string, devMode bool) *Service {
	return &Service{
		queries:   queries,
		rdb:       rdb,
		s3:        s3,
		jwtSecret: jwtSecret,
		devMode:   devMode,
	}
}

type AuthResult struct {
	Token string
	User  db.User
}

// --- OTP ---

func (s *Service) OTPSend(ctx context.Context, phone string) (string, error) {
	rlKey := fmt.Sprintf("rl:otp:%s", phone)
	count, err := s.rdb.Incr(ctx, rlKey).Result()
	if err != nil {
		return "", err
	}
	if count == 1 {
		s.rdb.Expire(ctx, rlKey, time.Hour)
	}
	if count > 3 {
		return "", ErrOTPRateLimit
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", n.Int64())

	otpKey := fmt.Sprintf("otp:%s", phone)
	s.rdb.Set(ctx, otpKey, code, 5*time.Minute)

	attKey := fmt.Sprintf("otp-attempts:%s", phone)
	s.rdb.Del(ctx, attKey)

	if s.devMode {
		return code, nil
	}
	return "", nil
}

func (s *Service) OTPVerify(ctx context.Context, phone, code string) (*AuthResult, error) {
	attKey := fmt.Sprintf("otp-attempts:%s", phone)
	attempts, err := s.rdb.Incr(ctx, attKey).Result()
	if err != nil {
		return nil, err
	}
	if attempts == 1 {
		s.rdb.Expire(ctx, attKey, 5*time.Minute)
	}
	if attempts > 5 {
		return nil, ErrOTPBlocked
	}

	otpKey := fmt.Sprintf("otp:%s", phone)
	stored, err := s.rdb.Get(ctx, otpKey).Result()
	if err != nil || stored != code {
		return nil, ErrInvalidOTP
	}

	s.rdb.Del(ctx, otpKey)
	s.rdb.Del(ctx, attKey)

	return s.upsertByPhone(ctx, phone)
}

func (s *Service) upsertByPhone(ctx context.Context, phone string) (*AuthResult, error) {
	user, err := s.queries.GetUserByPhone(ctx, sql.NullString{String: phone, Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		user, err = s.queries.CreateUser(ctx, db.CreateUserParams{
			ID:       uuid.New().String(),
			Phone:    sql.NullString{String: phone, Valid: true},
			Username: generateUsername(),
		})
		if err != nil {
			return nil, err
		}
	}
	token, err := s.signToken(user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

// --- Apple Auth ---

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwkKeySet struct {
	Keys []jwkKey `json:"keys"`
}

func (s *Service) AppleAuth(ctx context.Context, idToken string) (*AuthResult, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	unverified, _, err := parser.ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	kid, ok := unverified.Header["kid"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://appleid.apple.com/auth/keys", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch apple keys: %w", err)
	}
	defer resp.Body.Close()

	var keySet jwkKeySet
	if err := json.NewDecoder(resp.Body).Decode(&keySet); err != nil {
		return nil, fmt.Errorf("decode apple keys: %w", err)
	}

	var matched *jwkKey
	for _, k := range keySet.Keys {
		if k.Kid == kid {
			matched = &k
			break
		}
	}
	if matched == nil {
		return nil, ErrInvalidToken
	}

	pubKey, err := jwkToRSAPublicKey(matched)
	if err != nil {
		return nil, ErrInvalidToken
	}

	token, err := jwt.Parse(idToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, ErrInvalidToken
	}

	return s.upsertByAppleID(ctx, sub)
}

func jwkToRSAPublicKey(k *jwkKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

func (s *Service) upsertByAppleID(ctx context.Context, appleID string) (*AuthResult, error) {
	user, err := s.queries.GetUserByAppleID(ctx, sql.NullString{String: appleID, Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		user, err = s.queries.CreateUser(ctx, db.CreateUserParams{
			ID:       uuid.New().String(),
			AppleID:  sql.NullString{String: appleID, Valid: true},
			Username: generateUsername(),
		})
		if err != nil {
			return nil, err
		}
	}
	token, err := s.signToken(user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

// --- Google Auth ---

type googleTokenInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func (s *Service) GoogleAuth(ctx context.Context, idToken string) (*AuthResult, error) {
	googleURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verify google token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidToken
	}

	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, ErrInvalidToken
	}
	if info.Sub == "" {
		return nil, ErrInvalidToken
	}

	return s.upsertByGoogleID(ctx, info.Sub)
}

func (s *Service) upsertByGoogleID(ctx context.Context, googleID string) (*AuthResult, error) {
	user, err := s.queries.GetUserByGoogleID(ctx, sql.NullString{String: googleID, Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		user, err = s.queries.CreateUser(ctx, db.CreateUserParams{
			ID:       uuid.New().String(),
			GoogleID: sql.NullString{String: googleID, Valid: true},
			Username: generateUsername(),
		})
		if err != nil {
			return nil, err
		}
	}
	token, err := s.signToken(user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

// --- Profile ---

func (s *Service) GetMe(ctx context.Context, userID string) (db.User, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, err
	}
	return user, nil
}

func (s *Service) CheckUsername(ctx context.Context, username string) (bool, error) {
	if !usernameRegex.MatchString(username) {
		return false, ErrInvalidUsername
	}
	_, err := s.queries.GetUserByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

type UpdateProfileParams struct {
	Username    *string
	AvatarEmoji *string
	BirthYear   *int
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, params UpdateProfileParams) (db.User, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, err
	}

	newUsername := user.Username
	if params.Username != nil && *params.Username != user.Username {
		if !usernameRegex.MatchString(*params.Username) {
			return db.User{}, ErrInvalidUsername
		}
		if time.Since(user.UpdatedAt) < 30*24*time.Hour {
			return db.User{}, ErrUsernameRecent
		}
		existing, err := s.queries.GetUserByUsername(ctx, *params.Username)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return db.User{}, err
		}
		if err == nil && existing.ID != userID {
			return db.User{}, ErrUsernameTaken
		}
		newUsername = *params.Username
	}

	newEmoji := user.AvatarEmoji
	if params.AvatarEmoji != nil {
		newEmoji = sql.NullString{String: *params.AvatarEmoji, Valid: *params.AvatarEmoji != ""}
	}

	newBirthYear := user.BirthYear
	if params.BirthYear != nil {
		newBirthYear = sql.NullInt32{Int32: int32(*params.BirthYear), Valid: true}
	}

	return s.queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:          userID,
		Username:    newUsername,
		AvatarEmoji: newEmoji,
		AvatarUrl:   user.AvatarUrl,
		BirthYear:   newBirthYear,
	})
}

// --- Avatar upload ---

const maxAvatarSize = 5 * 1024 * 1024

func (s *Service) UploadAvatar(ctx context.Context, userID string, fileData io.Reader, fileSize int64) (db.User, error) {
	if s.s3 == nil {
		return db.User{}, ErrAvatarUnavailable
	}
	if fileSize > maxAvatarSize {
		return db.User{}, ErrImageTooLarge
	}

	data, err := io.ReadAll(io.LimitReader(fileData, maxAvatarSize+1))
	if err != nil {
		return db.User{}, err
	}
	if int64(len(data)) > maxAvatarSize {
		return db.User{}, ErrImageTooLarge
	}

	if !isValidImage(data) {
		return db.User{}, ErrInvalidImage
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return db.User{}, ErrInvalidImage
	}

	resized := imaging.Fill(img, 256, 256, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85}); err != nil {
		return db.User{}, err
	}

	key := fmt.Sprintf("avatars/%s.jpg", userID)
	avatarURL, err := s.s3.Upload(ctx, key, buf.Bytes(), "image/jpeg")
	if err != nil {
		return db.User{}, err
	}

	return s.queries.UpdateUserAvatarURL(ctx, db.UpdateUserAvatarURLParams{
		ID:        userID,
		AvatarUrl: sql.NullString{String: avatarURL, Valid: true},
	})
}

func isValidImage(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}
	return false
}

// --- Delete account ---

func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
	return s.queries.DeleteUser(ctx, userID)
}

// --- Push preferences ---

func (s *Service) UpsertPushPreference(ctx context.Context, userID, category string, enabled bool) (db.PushPreference, error) {
	return s.queries.UpsertPushPreference(ctx, db.UpsertPushPreferenceParams{
		ID:       uuid.New().String(),
		UserID:   userID,
		Category: db.PushCategory(category),
		Enabled:  enabled,
	})
}

// --- JWT ---

func (s *Service) signToken(user db.User) (string, error) {
	claims := middleware.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(90 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func generateUsername() string {
	return "user_" + uuid.New().String()[:8]
}
