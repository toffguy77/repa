package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/config"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	authsvc "github.com/repa-app/repa/internal/service/auth"
)

type Handler struct {
	svc *authsvc.Service
	cfg *config.Config
}

func NewHandler(svc *authsvc.Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// --- DTOs ---

type UserDto struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	AvatarURL   *string `json:"avatar_url"`
	AvatarEmoji *string `json:"avatar_emoji"`
	BirthYear   *int    `json:"birth_year"`
	CreatedAt   string  `json:"created_at"`
}

func toUserDto(u db.User) UserDto {
	dto := UserDto{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
	if u.AvatarUrl.Valid {
		dto.AvatarURL = &u.AvatarUrl.String
	}
	if u.AvatarEmoji.Valid {
		dto.AvatarEmoji = &u.AvatarEmoji.String
	}
	if u.BirthYear.Valid {
		v := int(u.BirthYear.Int32)
		dto.BirthYear = &v
	}
	return dto
}

type authResponse struct {
	Token string  `json:"token"`
	User  UserDto `json:"user"`
}

// --- Apple Auth ---

type appleAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

func (h *Handler) AppleAuth(c echo.Context) error {
	var req appleAuthRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	result, err := h.svc.AppleAuth(c.Request().Context(), req.IDToken)
	if err != nil {
		if errors.Is(err, authsvc.ErrInvalidToken) {
			return handler.ErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid Apple ID token")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Authentication failed")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": authResponse{Token: result.Token, User: toUserDto(result.User)},
	})
}

// --- Google Auth ---

type googleAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

func (h *Handler) GoogleAuth(c echo.Context) error {
	var req googleAuthRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	result, err := h.svc.GoogleAuth(c.Request().Context(), req.IDToken)
	if err != nil {
		if errors.Is(err, authsvc.ErrInvalidToken) {
			return handler.ErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid Google ID token")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Authentication failed")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": authResponse{Token: result.Token, User: toUserDto(result.User)},
	})
}

// --- OTP Send ---

type otpSendRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
}

func (h *Handler) OTPSend(c echo.Context) error {
	var req otpSendRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	code, err := h.svc.OTPSend(c.Request().Context(), req.Phone)
	if err != nil {
		if errors.Is(err, authsvc.ErrOTPRateLimit) {
			return handler.ErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT", "Too many OTP requests")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to send OTP")
	}

	resp := map[string]any{"sent": true}
	if code != "" {
		resp["code"] = code // dev mode only
	}

	return c.JSON(http.StatusOK, map[string]any{"data": resp})
}

// --- OTP Verify ---

type otpVerifyRequest struct {
	Phone string `json:"phone" validate:"required"`
	Code  string `json:"code" validate:"required,len=6"`
}

func (h *Handler) OTPVerify(c echo.Context) error {
	var req otpVerifyRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	result, err := h.svc.OTPVerify(c.Request().Context(), req.Phone, req.Code)
	if err != nil {
		if errors.Is(err, authsvc.ErrInvalidOTP) {
			return handler.ErrorResponse(c, http.StatusUnauthorized, "INVALID_OTP", "Invalid verification code")
		}
		if errors.Is(err, authsvc.ErrOTPBlocked) {
			return handler.ErrorResponse(c, http.StatusTooManyRequests, "OTP_BLOCKED", "Too many attempts, try again later")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Verification failed")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": authResponse{Token: result.Token, User: toUserDto(result.User)},
	})
}

// --- Get Me ---

func (h *Handler) GetMe(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)
	user, err := h.svc.GetMe(c.Request().Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, authsvc.ErrUserNotFound) {
			return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to get user")
	}

	return c.JSON(http.StatusOK, map[string]any{"data": toUserDto(user)})
}

// --- Username Check ---

func (h *Handler) UsernameCheck(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "username parameter required")
	}

	available, err := h.svc.CheckUsername(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, authsvc.ErrInvalidUsername) {
			return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid username format")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Check failed")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]bool{"available": available},
	})
}

// --- Update Profile ---

type updateProfileRequest struct {
	Username    *string `json:"username" validate:"omitempty,min=3,max=20"`
	AvatarEmoji *string `json:"avatar_emoji"`
	BirthYear   *int    `json:"birth_year" validate:"omitempty,min=1990,max=2012"`
}

func (h *Handler) UpdateProfile(c echo.Context) error {
	var req updateProfileRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	claims := appmw.GetCurrentUser(c)
	user, err := h.svc.UpdateProfile(c.Request().Context(), claims.UserID, authsvc.UpdateProfileParams{
		Username:    req.Username,
		AvatarEmoji: req.AvatarEmoji,
		BirthYear:   req.BirthYear,
	})
	if err != nil {
		if errors.Is(err, authsvc.ErrUserNotFound) {
			return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
		}
		if errors.Is(err, authsvc.ErrUsernameRecent) {
			return handler.ErrorResponse(c, http.StatusConflict, "USERNAME_COOLDOWN", "Username can be changed once every 30 days")
		}
		if errors.Is(err, authsvc.ErrUsernameTaken) {
			return handler.ErrorResponse(c, http.StatusConflict, "USERNAME_TAKEN", "Username is already taken")
		}
		if errors.Is(err, authsvc.ErrInvalidUsername) {
			return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid username format")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to update profile")
	}

	return c.JSON(http.StatusOK, map[string]any{"data": toUserDto(user)})
}

// --- Avatar Upload ---

func (h *Handler) UploadAvatar(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "File is required")
	}

	src, err := file.Open()
	if err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Cannot read file")
	}
	defer src.Close()

	claims := appmw.GetCurrentUser(c)
	user, err := h.svc.UploadAvatar(c.Request().Context(), claims.UserID, src, file.Size)
	if err != nil {
		if errors.Is(err, authsvc.ErrImageTooLarge) {
			return handler.ErrorResponse(c, http.StatusBadRequest, "FILE_TOO_LARGE", "Image must be under 5MB")
		}
		if errors.Is(err, authsvc.ErrInvalidImage) {
			return handler.ErrorResponse(c, http.StatusBadRequest, "INVALID_IMAGE", "File must be JPEG or PNG")
		}
		if errors.Is(err, authsvc.ErrAvatarUnavailable) {
			return handler.ErrorResponse(c, http.StatusServiceUnavailable, "UNAVAILABLE", "Avatar uploads are not available")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to upload avatar")
	}

	return c.JSON(http.StatusOK, map[string]any{"data": toUserDto(user)})
}

// --- Push Preferences ---

type pushPrefRequest struct {
	Category string `json:"category" validate:"required,oneof=SEASON_START REMINDER REVEAL REACTION NEXT_SEASON"`
	Enabled  bool   `json:"enabled"`
}

func (h *Handler) UpdatePushPreferences(c echo.Context) error {
	var req pushPrefRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	claims := appmw.GetCurrentUser(c)
	pref, err := h.svc.UpsertPushPreference(c.Request().Context(), claims.UserID, req.Category, req.Enabled)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to update preferences")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"category": string(pref.Category),
			"enabled":  pref.Enabled,
		},
	})
}

// --- App Version ---

func (h *Handler) AppVersion(c echo.Context) error {
	appVersion := c.Request().Header.Get("X-App-Version")
	forceUpdate := false
	if appVersion != "" && compareSemver(appVersion, h.cfg.AppMinVersion) < 0 {
		forceUpdate = true
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"min_version":    h.cfg.AppMinVersion,
			"latest_version": h.cfg.AppLatestVersion,
			"force_update":   forceUpdate,
		},
	})
}

// compareSemver compares two semver strings (e.g. "1.2.3").
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareSemver(a, b string) int {
	aParts := strings.SplitN(a, ".", 3)
	bParts := strings.SplitN(b, ".", 3)
	for i := 0; i < 3; i++ {
		var av, bv int
		if i < len(aParts) {
			av, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bv, _ = strconv.Atoi(bParts[i])
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

// --- Delete Account ---

func (h *Handler) DeleteAccount(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)
	if err := h.svc.DeleteAccount(c.Request().Context(), claims.UserID); err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to delete account")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]bool{"deleted": true},
	})
}
