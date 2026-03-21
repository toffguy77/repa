# T04 — Backend (Go): Auth API

## Цель
Три способа аутентификации (Apple, Google, OTP), JWT, профиль, загрузка аватара.

## Структура
```
internal/
├── handler/auth/
│   ├── handler.go
│   └── handler_test.go
├── service/auth/
│   ├── service.go
│   └── service_test.go
```

## Эндпоинты

### POST /api/v1/auth/apple
```go
// Request
type AppleAuthRequest struct {
    IDToken string `json:"id_token" validate:"required"`
}
// 1. Fetch Apple public keys: GET https://appleid.apple.com/auth/keys
// 2. Валидировать JWT (iss, aud, exp)
// 3. Извлечь sub (appleId) и email
// 4. UpsertUserByAppleID → вернуть JWT + UserDto
```

### POST /api/v1/auth/google
```go
type GoogleAuthRequest struct {
    IDToken string `json:"id_token" validate:"required"`
}
// GET https://oauth2.googleapis.com/tokeninfo?id_token={token}
// Извлечь sub (googleId)
// UpsertUserByGoogleID → JWT + UserDto
```

### POST /api/v1/auth/otp/send
```go
type OTPSendRequest struct {
    Phone string `json:"phone" validate:"required,e164"` // +7XXXXXXXXXX
}
// Rate limit: Redis key `rl:otp:{phone}` limit=3 window=3600s
// Сгенерировать 6-значный код: fmt.Sprintf("%06d", rand.Intn(1000000))
// Сохранить: SET otp:{phone} {code} EX 300
// В dev: вернуть код в ответе (env DEV_MODE=true)
// В prod: заглушка с логом (SMS интеграция — отдельная задача)
```

### POST /api/v1/auth/otp/verify
```go
type OTPVerifyRequest struct {
    Phone string `json:"phone" validate:"required"`
    Code  string `json:"code" validate:"required,len=6"`
}
// GET otp:{phone} → сравнить
// Счётчик попыток: INCR otp-attempts:{phone} EX 300, при > 5 → блок
// DEL otp:{phone}
// UpsertUserByPhone → JWT + UserDto
```

### GET /api/v1/auth/me (protected)
```go
// Извлечь userId из JWT claims
// GetUserByID → UserDto
```

### GET /api/v1/auth/username-check
```go
// Query param: ?username=xxx
// НЕ требует JWT
// Rate limit по IP: 20/min
// Regex: ^[a-zA-Zа-яА-ЯёЁ0-9_]{3,20}$
// GetUserByUsername → { data: { available: bool } }
```

### PATCH /api/v1/auth/profile (protected)
```go
type UpdateProfileRequest struct {
    Username    *string `json:"username" validate:"omitempty,min=3,max=20"`
    AvatarEmoji *string `json:"avatar_emoji"`
    BirthYear   *int    `json:"birth_year" validate:"omitempty,min=1990,max=2012"`
}
// Username: проверить уникальность, проверить updated_at (не чаще раза в 30 дней)
// UpdateUserProfile → UserDto
```

### POST /api/v1/auth/avatar (protected, multipart)
```go
// Поле: file (image/jpeg или image/png, max 5MB)
// Валидировать magic bytes: JPEG = FF D8 FF, PNG = 89 50 4E 47
// Ресайз до 256×256 через github.com/disintegration/imaging
// Загрузить в S3: key = avatars/{userId}.jpg
// UpdateUserProfile(avatarUrl) → UserDto

// Dependencies:
// go get github.com/disintegration/imaging
```

### PATCH /api/v1/push/preferences (protected)
```go
type PushPrefRequest struct {
    Category string `json:"category" validate:"required,oneof=SEASON_START REMINDER REVEAL REACTION NEXT_SEASON"`
    Enabled  bool   `json:"enabled"`
}
// UPSERT push_preferences (user_id, category, enabled)
```

### GET /api/v1/app/version
```go
// Не требует JWT
// Читать X-App-Version из заголовка
// Сравнить с APP_MIN_VERSION, APP_LATEST_VERSION из env
// Вернуть { minVersion, latestVersion, forceUpdate }
```

### DELETE /api/v1/auth/account (protected)
```go
// DELETE FROM users WHERE id = $1 (cascade удалит всё связанное)
// 200 { data: { deleted: true } }
```

## JWT генерация
```go
func SignToken(userID, username, secret string) (string, error) {
    claims := JWTClaims{
        UserID:   userID,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(90 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
```

## UserDto
```go
type UserDto struct {
    ID          string  `json:"id"`
    Username    string  `json:"username"`
    AvatarURL   *string `json:"avatar_url"`
    AvatarEmoji *string `json:"avatar_emoji"`
    BirthYear   *int    `json:"birth_year"`
    CreatedAt   string  `json:"created_at"`
}
```

## Тесты (handler_test.go)
```go
// httptest.NewRecorder + echo инстанс
// OTP send → verify → /me: получить UserDto
// Apple auth: заглушить HTTP клиент для Apple Keys API
// Повторный вход по Apple ID → тот же user.id
// username-check: занятый → false, свободный → true
// Смена username < 30 дней → 409
// Avatar: не-image → 400, > 5MB → 400
```

## Критерии готовности
- [ ] Все эндпоинты реализованы
- [ ] OTP хранится в Redis с TTL 300s
- [ ] Rate limit OTP: 3/час на телефон
- [ ] username-check без JWT работает
- [ ] Фото аватара ресайзится и загружается в S3
- [ ] Тесты проходят (`go test ./internal/...`)
