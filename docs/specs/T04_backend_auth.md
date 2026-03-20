# T04 — Backend: Auth API

## Цель
Реализовать три способа аутентификации: Sign in with Apple, Sign in with Google, OTP по номеру телефона. Выдавать JWT.

## Эндпоинты

### POST /api/v1/auth/apple
```typescript
// Request
{ idToken: string }

// Логика
// 1. Верифицировать Apple ID Token через Apple Public Keys (jwks)
// 2. Извлечь appleId (sub) и email
// 3. Найти или создать User (upsert по appleId)
// 4. Вернуть JWT + user

// Response
{ data: { token: string, user: UserDto, isNew: boolean } }
```

### POST /api/v1/auth/google
```typescript
// Request
{ idToken: string }

// Логика
// 1. Верифицировать через Google Token Info API
//    GET https://oauth2.googleapis.com/tokeninfo?id_token={token}
// 2. Извлечь googleId (sub) и email
// 3. Найти или создать User
// 4. Вернуть JWT + user

// Response
{ data: { token: string, user: UserDto, isNew: boolean } }
```

### POST /api/v1/auth/otp/send
```typescript
// Request
{ phone: string }  // формат: +7XXXXXXXXXX

// Логика
// 1. Валидировать формат телефона (regex)
// 2. Сгенерировать 6-значный код
// 3. Сохранить в Redis: key=`otp:{phone}`, value=code, TTL=300 сек
// 4. Rate limit: не более 3 запросов на телефон в час (Redis counter)
// В dev режиме: вернуть код в ответе (для тестирования)
// В prod: интеграция с SMS-провайдером (заглушка с логированием)

// Response
{ data: { sent: true, expiresIn: 300 } }
```

### POST /api/v1/auth/otp/verify
```typescript
// Request
{ phone: string, code: string }

// Логика
// 1. Проверить код в Redis
// 2. Если неверный: счётчик попыток (max 5), потом блок
// 3. Удалить код из Redis
// 4. Найти или создать User по phone
// 5. Вернуть JWT + user

// Response
{ data: { token: string, user: UserDto, isNew: boolean } }
```

### GET /api/v1/auth/me
```typescript
// Headers: Authorization: Bearer {token}
// Response
{ data: { user: UserDto } }
```

### PATCH /api/v1/auth/profile
```typescript
// Headers: Authorization: Bearer {token}
// Request
{ username?: string, avatarEmoji?: string, birthYear?: number }

// Логика
// username: уникальность, 3-20 символов, только a-z0-9_ и кириллица
// username можно менять раз в 30 дней (проверять updatedAt)
// birthYear: 1990–2015 (14–35 лет)

// Response
{ data: { user: UserDto } }
```

### DELETE /api/v1/auth/account
```typescript
// Удалить аккаунт и все связанные данные
// Prisma cascade delete через транзакцию
// Response: 200 { data: { deleted: true } }
```

## Типы

```typescript
type UserDto = {
  id: string
  username: string
  avatarUrl: string | null
  avatarEmoji: string | null
  birthYear: number | null
  createdAt: string
}
```

## Структура файлов
```
src/modules/auth/
├── auth.router.ts
├── auth.service.ts
├── auth.schema.ts
└── auth.test.ts
```

## Тесты (auth.test.ts)
- OTP send → verify → получить токен → GET /me
- Повторный вход по Apple ID возвращает того же пользователя
- Неверный OTP-код → ошибка
- Истёкший OTP-код → ошибка
- Смена username чаще раза в 30 дней → 409

## Критерии готовности
- [ ] Все 6 эндпоинтов работают
- [ ] JWT проверяется в `/me`
- [ ] OTP хранится в Redis с TTL
- [ ] Rate limiting на OTP send
- [ ] Тесты проходят

## Не делать
- Не интегрировать реальный SMS-провайдер (заглушка в prod)
- Не реализовывать refresh token в MVP

---

## Дополнение: недостающие эндпоинты

### GET /api/v1/auth/username-check
```typescript
// Не требует авторизации — нужен на экране регистрации до получения JWT
// Query: ?username=xxx
// Валидация: 3-20 символов, только [a-zA-Z0-9_] и кириллица (regex)

{ data: { available: boolean } }

// Rate limit: 20 запросов в минуту по IP (защита от перебора)
```

### POST /api/v1/auth/avatar
```typescript
// Требует авторизации
// multipart/form-data, поле: file
// Ограничения: только image/jpeg и image/png, максимум 5MB

// Логика:
// 1. Валидировать MIME тип через magic bytes (не только Content-Type)
// 2. Ресайз до 256×256 через sharp (добавить в dependencies)
// 3. Загрузить в S3: key = `avatars/{userId}.jpg`, ContentType: image/jpeg
// 4. Обновить user.avatarUrl

{ data: { avatarUrl: string } }

// Dependencies: добавить sharp в package.json
// npm install sharp @types/sharp
```

### PATCH /api/v1/push/preferences
```typescript
// Настройки категорий push-уведомлений
// Request: { category: PushCategory, enabled: boolean }

enum PushCategory {
  SEASON_START       // понедельник, старт сезона
  REMINDER           // среда-четверг, напоминания
  REVEAL             // пятница, reveal готов
  REACTION           // реакция на карточку
  NEXT_SEASON        // воскресенье, голосование за вопросы
}

// Хранить в Redis: key = `push-prefs:{userId}`, value = JSON map категорий
// TTL: без TTL (постоянно)

{ data: { updated: true } }
```

### GET /api/v1/app/version
```typescript
// Не требует авторизации
// Возвращает версионирование для force update
// Значения берутся из env: APP_MIN_VERSION, APP_LATEST_VERSION

{
  data: {
    minVersion: string,      // '1.0.0' — ниже этой версии — force update
    latestVersion: string,   // '1.2.0' — мягкое предложение обновиться
    forceUpdate: boolean     // true если клиент ниже minVersion
  }
}
// Клиент передаёт свою версию в заголовке: X-App-Version: 1.0.0
```

### Обновить тесты (auth.test.ts)
- username-check: занятый username → `{ available: false }`
- username-check: свободный → `{ available: true }`
- username-check: невалидный формат → 400
- avatar upload: не-image файл → 400
- avatar upload: файл > 5MB → 400
- push preferences: установить категорию → учитывается при следующей отправке

### Новые переменные в .env.example
```env
APP_MIN_VERSION=1.0.0
APP_LATEST_VERSION=1.0.0
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change_me_in_production
```
