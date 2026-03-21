# 🍆 РЕПА — Master Context для Claude Code

> Этот документ передаётся агенту вместе с каждой атомарной задачей.
> Он описывает всё, что нужно знать о продукте, архитектуре и соглашениях.
> Не реализовывай ничего сверх текущей задачи — только то, что в ней описано.

---

## 1. Что такое Репа

Мобильное приложение (Flutter, iOS + Android), в котором пользователи состоят в закрытых группах и еженедельно **анонимно голосуют** за участников по смешным вопросам («Кто первым убежит при пожаре?»). В пятницу в 20:00 — **Reveal**: каждый видит свою карточку репутации. Можно купить «детектор» за кристаллы — узнать, кто голосовал.

**Аудитория:** школьники и студенты 14–22 лет, Россия.

---

## 2. Стек

### Backend
| Слой | Технология |
|---|---|
| Язык | Go 1.22+ |
| Framework | Echo v4 |
| DB migrations | golang-migrate |
| DB queries | sqlc (type-safe генерация из SQL) |
| БД | PostgreSQL 16 |
| Кэш / очереди | Redis 7 (go-redis/v9) |
| Джобы | asynq (Redis-based task queue) |
| Push | firebase-admin-go (FCM) |
| Хранилище | Yandex Object Storage (aws-sdk-go-v2, S3-compatible) |
| Рендер карточек | chromedp (headless Chrome на Go) |
| AI модерация | Anthropic API (HTTP client, net/http) |
| Биллинг | ЮKassa REST API (net/http) |
| Telegram | go-telegram-bot-api/v5 |
| Валидация | go-playground/validator/v10 |
| JWT | golang-jwt/jwt/v5 |
| Логирование | zerolog |
| Конфигурация | os.Getenv + godotenv |
| Тесты | testify + httptest |

### Flutter (Mobile)
| Слой | Технология |
|---|---|
| SDK | Flutter 3.x, Dart 3 |
| State | Riverpod 2 |
| Навигация | go_router |
| HTTP | Dio + retrofit |
| Push | firebase_messaging |
| Deeplinks | app_links |
| Шеринг | share_plus |
| Telegram | url_launcher |
| Платежи | url_launcher → внешний браузер |
| Локальное хранилище | flutter_secure_storage |
| Анимации | flutter_animate |
| Кодогенерация | build_runner, freezed, json_serializable |

### Инфраструктура
- Монорепозиторий: `/backend` и `/mobile` в корне
- Docker Compose для локальной разработки (postgres, redis)
- Переменные окружения через `.env` (пример в `.env.example`)

---

## 3. Структура репозитория

```
repa/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # точка входа
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go         # загрузка env
│   │   ├── db/
│   │   │   ├── migrations/       # SQL файлы golang-migrate
│   │   │   ├── queries/          # SQL запросы для sqlc
│   │   │   └── sqlc/             # сгенерированный Go код (не редактировать)
│   │   ├── handler/              # Echo handlers (routing + validation)
│   │   │   └── {feature}/
│   │   │       ├── handler.go
│   │   │       └── handler_test.go
│   │   ├── service/              # бизнес-логика
│   │   │   └── {feature}/
│   │   │       ├── service.go
│   │   │       └── service_test.go
│   │   ├── worker/               # asynq workers
│   │   │   ├── worker.go         # регистрация всех handlers
│   │   │   └── tasks/            # task handlers по доменам
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── ratelimit.go
│   │   │   └── security.go
│   │   └── lib/                  # внешние клиенты-синглтоны
│   │       ├── redis.go
│   │       ├── firebase.go
│   │       ├── s3.go
│   │       ├── telegram.go
│   │       └── asynq.go
│   ├── sqlc.yaml
│   ├── .env.example
│   └── go.mod
├── mobile/
│   ├── lib/
│   │   ├── core/
│   │   │   ├── api/          # Dio клиент, retrofit интерфейсы
│   │   │   ├── router/       # go_router конфиг
│   │   │   ├── theme/        # Цвета, типографика, компоненты
│   │   │   └── providers/    # Глобальные Riverpod провайдеры
│   │   └── features/         # Фичи
│   │       └── {feature}/
│   │           ├── data/     # Repository, API модели
│   │           ├── domain/   # Use cases, entities (freezed)
│   │           └── presentation/  # Screens, Widgets, Notifiers
│   └── pubspec.yaml
└── docker-compose.yml
```

---

## 4. Схема базы данных (SQL migrations)

Схема реализуется через SQL-миграции в `internal/db/migrations/`.
sqlc читает SQL-запросы из `internal/db/queries/` и генерирует типобезопасный Go-код.

```sql
-- 001_init.up.sql

CREATE TYPE season_status AS ENUM ('VOTING', 'REVEALED', 'CLOSED');
CREATE TYPE question_category AS ENUM ('HOT', 'FUNNY', 'SECRETS', 'SKILLS', 'ROMANCE', 'STUDY');
CREATE TYPE question_source AS ENUM ('SYSTEM', 'USER');
CREATE TYPE question_status AS ENUM ('ACTIVE', 'PENDING', 'REJECTED');
CREATE TYPE crystal_log_type AS ENUM ('PURCHASE', 'SPEND_DETECTOR', 'SPEND_ATTRIBUTES', 'SPEND_QUESTION', 'BONUS');
CREATE TYPE achievement_type AS ENUM (
  'SNIPER', 'ORACLE', 'TELEPATH', 'BLIND', 'RANDOM',
  'EXPERT_OF', 'BEST_FRIEND', 'DETECTIVE', 'STRANGER',
  'LEGEND', 'CHANGEABLE', 'MONOPOLIST', 'ENIGMA', 'RISING', 'PIONEER',
  'STREAK_VOTER', 'FIRST_VOTER', 'LAST_VOTER', 'NIGHT_OWL', 'ANALYST',
  'MEDIA', 'CONSPIRATOR', 'RECRUITER'
);
CREATE TYPE push_category AS ENUM ('SEASON_START', 'REMINDER', 'REVEAL', 'REACTION', 'NEXT_SEASON');

CREATE TABLE users (
  id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  phone         TEXT UNIQUE,
  apple_id      TEXT UNIQUE,
  google_id     TEXT UNIQUE,
  username      TEXT UNIQUE NOT NULL,
  avatar_url    TEXT,
  avatar_emoji  TEXT,
  birth_year    INT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE groups (
  id                      TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  name                    TEXT NOT NULL,
  invite_code             TEXT UNIQUE NOT NULL DEFAULT gen_random_uuid()::text,
  admin_id                TEXT NOT NULL REFERENCES users(id),
  telegram_chat_id        TEXT,
  telegram_chat_username  TEXT,
  telegram_connect_code   TEXT,
  telegram_connect_expiry TIMESTAMPTZ,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE group_members (
  id        TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id  TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, group_id)
);

CREATE TABLE seasons (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  group_id   TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  number     INT NOT NULL,
  status     season_status NOT NULL DEFAULT 'VOTING',
  starts_at  TIMESTAMPTZ NOT NULL,
  reveal_at  TIMESTAMPTZ NOT NULL,
  ends_at    TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  text       TEXT NOT NULL,
  category   question_category NOT NULL,
  source     question_source NOT NULL DEFAULT 'SYSTEM',
  group_id   TEXT REFERENCES groups(id) ON DELETE CASCADE,
  author_id  TEXT REFERENCES users(id),
  status     question_status NOT NULL DEFAULT 'ACTIVE',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE season_questions (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id   TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  question_id TEXT NOT NULL REFERENCES questions(id),
  ord         INT NOT NULL,
  UNIQUE(season_id, question_id)
);

CREATE TABLE votes (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id   TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  voter_id    TEXT NOT NULL REFERENCES users(id),
  target_id   TEXT NOT NULL REFERENCES users(id),
  question_id TEXT NOT NULL REFERENCES questions(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(season_id, voter_id, question_id)
);

CREATE TABLE season_results (
  id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id    TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  target_id    TEXT NOT NULL REFERENCES users(id),
  question_id  TEXT NOT NULL REFERENCES questions(id),
  vote_count   INT NOT NULL,
  total_voters INT NOT NULL,
  percentage   FLOAT NOT NULL,
  UNIQUE(season_id, target_id, question_id)
);

CREATE TABLE achievements (
  id               TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id         TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  season_id        TEXT REFERENCES seasons(id),
  achievement_type achievement_type NOT NULL,
  metadata         JSONB,
  earned_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_group_stats (
  id                  TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id            TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  seasons_played      INT NOT NULL DEFAULT 0,
  voting_streak       INT NOT NULL DEFAULT 0,
  max_voting_streak   INT NOT NULL DEFAULT 0,
  guess_accuracy      FLOAT NOT NULL DEFAULT 0,
  total_votes_cast    INT NOT NULL DEFAULT 0,
  total_votes_received INT NOT NULL DEFAULT 0,
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, group_id)
);

CREATE TABLE detectors (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  group_id   TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, season_id)
);

CREATE TABLE crystal_logs (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  delta       INT NOT NULL,
  balance     INT NOT NULL,
  type        crystal_log_type NOT NULL,
  description TEXT,
  external_id TEXT UNIQUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fcm_tokens (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token      TEXT UNIQUE NOT NULL,
  platform   TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE card_cache (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  image_url  TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, season_id)
);

CREATE TABLE reactions (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  reactor_id TEXT NOT NULL REFERENCES users(id),
  target_id  TEXT NOT NULL REFERENCES users(id),
  emoji      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(season_id, reactor_id, target_id)
);

CREATE TABLE reports (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  question_id TEXT NOT NULL REFERENCES questions(id),
  reporter_id TEXT NOT NULL REFERENCES users(id),
  reason      TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(question_id, reporter_id)
);

CREATE TABLE push_preferences (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  category   push_category NOT NULL,
  enabled    BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE(user_id, category)
);

CREATE TABLE next_season_votes (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  group_id    TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id     TEXT NOT NULL REFERENCES users(id),
  question_id TEXT NOT NULL REFERENCES questions(id),
  season_number INT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(group_id, user_id, season_number)
);

-- Индексы
CREATE INDEX idx_votes_season ON votes(season_id);
CREATE INDEX idx_votes_target_season ON votes(target_id, season_id);
CREATE INDEX idx_season_results_season ON season_results(season_id, target_id);
CREATE INDEX idx_achievements_user_group ON achievements(user_id, group_id);
CREATE INDEX idx_user_group_stats ON user_group_stats(user_id, group_id);
CREATE INDEX idx_seasons_group_status ON seasons(group_id, status);
```

-- (остальная схема — в миграционных файлах)
Схема реализуется через SQL-миграции в `internal/db/migrations/`.
sqlc читает SQL-запросы из `internal/db/queries/` и генерирует типобезопасный Go-код.

```sql
-- 001_init.up.sql

CREATE TYPE season_status AS ENUM ('VOTING', 'REVEALED', 'CLOSED');
CREATE TYPE question_category AS ENUM ('HOT', 'FUNNY', 'SECRETS', 'SKILLS', 'ROMANCE', 'STUDY');
CREATE TYPE question_source AS ENUM ('SYSTEM', 'USER');
CREATE TYPE question_status AS ENUM ('ACTIVE', 'PENDING', 'REJECTED');
CREATE TYPE crystal_log_type AS ENUM ('PURCHASE', 'SPEND_DETECTOR', 'SPEND_ATTRIBUTES', 'SPEND_QUESTION', 'BONUS');
CREATE TYPE achievement_type AS ENUM (
  'SNIPER', 'ORACLE', 'TELEPATH', 'BLIND', 'RANDOM',
  'EXPERT_OF', 'BEST_FRIEND', 'DETECTIVE', 'STRANGER',
  'LEGEND', 'CHANGEABLE', 'MONOPOLIST', 'ENIGMA', 'RISING', 'PIONEER',
  'STREAK_VOTER', 'FIRST_VOTER', 'LAST_VOTER', 'NIGHT_OWL', 'ANALYST',
  'MEDIA', 'CONSPIRATOR', 'RECRUITER'
);
CREATE TYPE push_category AS ENUM ('SEASON_START', 'REMINDER', 'REVEAL', 'REACTION', 'NEXT_SEASON');

CREATE TABLE users (
  id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  phone         TEXT UNIQUE,
  apple_id      TEXT UNIQUE,
  google_id     TEXT UNIQUE,
  username      TEXT UNIQUE NOT NULL,
  avatar_url    TEXT,
  avatar_emoji  TEXT,
  birth_year    INT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE groups (
  id                      TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  name                    TEXT NOT NULL,
  invite_code             TEXT UNIQUE NOT NULL DEFAULT gen_random_uuid()::text,
  admin_id                TEXT NOT NULL REFERENCES users(id),
  telegram_chat_id        TEXT,
  telegram_chat_username  TEXT,
  telegram_connect_code   TEXT,
  telegram_connect_expiry TIMESTAMPTZ,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE group_members (
  id        TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id  TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, group_id)
);

CREATE TABLE seasons (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  group_id   TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  number     INT NOT NULL,
  status     season_status NOT NULL DEFAULT 'VOTING',
  starts_at  TIMESTAMPTZ NOT NULL,
  reveal_at  TIMESTAMPTZ NOT NULL,
  ends_at    TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  text       TEXT NOT NULL,
  category   question_category NOT NULL,
  source     question_source NOT NULL DEFAULT 'SYSTEM',
  group_id   TEXT REFERENCES groups(id) ON DELETE CASCADE,
  author_id  TEXT REFERENCES users(id),
  status     question_status NOT NULL DEFAULT 'ACTIVE',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE season_questions (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id   TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  question_id TEXT NOT NULL REFERENCES questions(id),
  ord         INT NOT NULL,
  UNIQUE(season_id, question_id)
);

CREATE TABLE votes (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id   TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  voter_id    TEXT NOT NULL REFERENCES users(id),
  target_id   TEXT NOT NULL REFERENCES users(id),
  question_id TEXT NOT NULL REFERENCES questions(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(season_id, voter_id, question_id)
);

CREATE TABLE season_results (
  id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id    TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  target_id    TEXT NOT NULL REFERENCES users(id),
  question_id  TEXT NOT NULL REFERENCES questions(id),
  vote_count   INT NOT NULL,
  total_voters INT NOT NULL,
  percentage   FLOAT NOT NULL,
  UNIQUE(season_id, target_id, question_id)
);

CREATE TABLE achievements (
  id               TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id         TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  season_id        TEXT REFERENCES seasons(id),
  achievement_type achievement_type NOT NULL,
  metadata         JSONB,
  earned_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_group_stats (
  id                  TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id            TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  seasons_played      INT NOT NULL DEFAULT 0,
  voting_streak       INT NOT NULL DEFAULT 0,
  max_voting_streak   INT NOT NULL DEFAULT 0,
  guess_accuracy      FLOAT NOT NULL DEFAULT 0,
  total_votes_cast    INT NOT NULL DEFAULT 0,
  total_votes_received INT NOT NULL DEFAULT 0,
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, group_id)
);

CREATE TABLE detectors (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  group_id   TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, season_id)
);

CREATE TABLE crystal_logs (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  delta       INT NOT NULL,
  balance     INT NOT NULL,
  type        crystal_log_type NOT NULL,
  description TEXT,
  external_id TEXT UNIQUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fcm_tokens (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token      TEXT UNIQUE NOT NULL,
  platform   TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE card_cache (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  image_url  TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, season_id)
);

CREATE TABLE reactions (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  reactor_id TEXT NOT NULL REFERENCES users(id),
  target_id  TEXT NOT NULL REFERENCES users(id),
  emoji      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(season_id, reactor_id, target_id)
);

CREATE TABLE reports (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  question_id TEXT NOT NULL REFERENCES questions(id),
  reporter_id TEXT NOT NULL REFERENCES users(id),
  reason      TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(question_id, reporter_id)
);

CREATE TABLE push_preferences (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  category   push_category NOT NULL,
  enabled    BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE(user_id, category)
);

CREATE TABLE next_season_votes (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  group_id    TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id     TEXT NOT NULL REFERENCES users(id),
  question_id TEXT NOT NULL REFERENCES questions(id),
  season_number INT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(group_id, user_id, season_number)
);

-- Индексы
CREATE INDEX idx_votes_season ON votes(season_id);
CREATE INDEX idx_votes_target_season ON votes(target_id, season_id);
CREATE INDEX idx_season_results_season ON season_results(season_id, target_id);
CREATE INDEX idx_achievements_user_group ON achievements(user_id, group_id);
CREATE INDEX idx_user_group_stats ON user_group_stats(user_id, group_id);
CREATE INDEX idx_seasons_group_status ON seasons(group_id, status);
```

---

## 5. API соглашения

- Язык: Go, Echo v4 router
- Базовый URL: `/api/v1`
- Аутентификация: Bearer JWT в заголовке `Authorization`
- Формат ответа всегда:
  ```json
  { "data": { ... } }           // успех
  { "error": { "code": "...", "message": "..." } }  // ошибка
  ```
- HTTP коды: 200 успех, 201 создание, 400 валидация, 401 не авторизован, 403 запрещено, 404 не найден, 409 конфликт, 500 сервер
- Пагинация: `?page=1&limit=20`, ответ: `{ data: [], meta: { total, page, limit } }`
- Все даты в ISO 8601 UTC

---

## 6. Ключевые бизнес-правила

### Анонимность (КРИТИЧНО)
- `votes` таблица хранит `voterId` — это нужно для детектора
- **Детектор возвращает только список `voterId`** — без привязки к конкретным вопросам или ответам
- API голосования **никогда** не возвращает `voterId` в результатах
- `SeasonResult` не содержит `voterId`

### Сезоны
- Первый сезон создаётся при создании группы (T06)
- Последующие сезоны создаются BullMQ job `season-creator` каждое воскресенье в 18:00 UTC
- Группа без сезона в статусе `VOTING` — получает новый сезон автоматически
- Группа с < 3 участников — новый сезон не создаётся

### Reveal
- Происходит строго в пятницу в 20:00 МСК (UTC+3 = 17:00 UTC)
- Только если кворум выполнен: ≥ 50% участников группы проголосовало (≥ 40% для групп < 8 человек)
- BullMQ job `reveal-checker` запускается каждую минуту, проверяет все сезоны со статусом `VOTING` и `revealAt <= now`
- После Reveal: статус → `REVEALED`, записываются `SeasonResult`, рассчитываются ачивки, отправляются пуши

### Группы
- Один пользователь: не более 10 групп (MVP)
- Размер группы: 5–50 участников (MVP)
- Группа активируется при ≥ 3 участниках
- Администратор = создатель. Никаких дополнительных привилегий кроме: название группы, привязка Telegram, добавление вопросов

### Кристаллы
- Детектор: 10 💎
- Открытие скрытых атрибутов: 5 💎
- Баланс хранится как сумма всех `CrystalLog.delta` для пользователя (не отдельное поле)
- Транзакция: проверка баланса + списание атомарно через Prisma transaction

### Возраст
- Пользователям до 18 лет категория `ROMANCE` недоступна при создании группы

### Push
- Не более 3 пушей в сутки на пользователя (счётчик в Redis с TTL до полуночи)
- Не отправлять в 23:00–09:00 МСК

---

## 7. Переменные окружения (`.env.example`)

```env
# Server
PORT=3000
NODE_ENV=development
JWT_SECRET=change_me_in_production

# Database
DATABASE_URL=postgresql://repa:repa@localhost:5432/repa

# Redis
REDIS_URL=redis://localhost:6379

# Firebase (FCM)
FIREBASE_PROJECT_ID=
FIREBASE_PRIVATE_KEY=
FIREBASE_CLIENT_EMAIL=

# Yandex Object Storage
S3_ENDPOINT=https://storage.yandexcloud.net
S3_BUCKET=repa-media
S3_ACCESS_KEY=
S3_SECRET_KEY=
S3_REGION=ru-central1

# Anthropic
ANTHROPIC_API_KEY=

# ЮKassa
YUKASSA_SHOP_ID=
YUKASSA_SECRET_KEY=
YUKASSA_RETURN_URL=https://repa.app/payment/return

# Telegram
TELEGRAM_BOT_TOKEN=
TELEGRAM_WEBHOOK_SECRET=

# App
APP_BASE_URL=https://repa.app
REVEAL_CRON=0 17 * * 5  # каждую пятницу в 17:00 UTC = 20:00 МСК
```

---

## 8. Flutter соглашения

- **Язык UI:** русский
- **Цветовая схема:** фиолетовый акцент `#7C3AED`, фон белый / тёмный системный
- **Шрифт:** системный (SF Pro на iOS, Roboto на Android)
- **Именование файлов:** `snake_case.dart`
- **Именование классов:** `PascalCase`
- **Провайдеры Riverpod:** `final xProvider = ...Provider((ref) => ...)` в отдельных файлах
- **Навигация:** именованные маршруты через go_router, deeplink-aware
- **Обработка ошибок:** `AsyncValue` от Riverpod, UI показывает `ErrorWidget` с кнопкой retry
- **Freezed модели** для всех domain entities и API ответов
- **Локализация:** хардкод русских строк в MVP (без arb-файлов)

---

## 9. Порядок реализации задач

```
Фаза 1: Фундамент
  T01 — Монорепо, Docker, конфиги
  T02 — Prisma схема и seed вопросов
  T03 — Echo приложение, middleware, базовая структура

Фаза 2: Auth
  T04 — Backend: Auth API (Apple, Google, OTP)
  T05 — Flutter: Auth screens

Фаза 3: Группы
  T06 — Backend: Groups API
  T07 — Flutter: Группы (список, создание, вступление)

Фаза 4: Голосование
  T08 — Backend: Voting API
  T09 — Flutter: Экран голосования

Фаза 5: Reveal
  T10 — Backend: Reveal engine (scheduler, агрегация)
  T11 — Backend: Ачивки — движок расчёта
  T12 — Backend: Генерация PNG карточек (Puppeteer)
  T13 — Flutter: Reveal Screen и анимации
  T14 — Flutter: Профиль участника и коллекция ачивок

Фаза 6: Монетизация
  T15 — Backend: Кристаллы и ЮKassa
  T16 — Flutter: Магазин кристаллов и платёжный flow

Фаза 7: Retention
  T17 — Backend: Push-уведомления (FCM) и недельный scheduler
  T18 — Flutter: Push handling, deeplinks, реакции на карточки

Фаза 8: Telegram
  T19 — Backend: Telegram bot и интеграция
  T20 — Flutter: Telegram UI (привязка, шеринг, кнопка чата)

Фаза 9: Модерация и безопасность
  T21 — Backend: AI-модерация пользовательских вопросов
  T22 — Backend: Rate limiting, security hardening

Фаза 10: Финал
  T23 — Flutter: Design polish, анимации, edge cases
  T24 — E2E чеклист и подготовка к релизу
  T25 — Flutter: Онбординг + экран голосования за вопросы следующей недели
  T26 — Flutter: Настройки, аналитика, иконка, splash, staging конфиги
```

---

## 10. Что не делать

- Не реализовывать Phase 2 фичи (Android-only фичи, множественные часовые пояса, расширенный банк ачивок, скины)
- Не использовать Apple IAP или Google Play Billing — только ЮKassa через браузер
- Не читать сообщения Telegram-чата в боте — только писать
- Не возвращать `voterId` в привязке к конкретному голосу в любом API-ответе
- Не хранить баланс кристаллов как отдельное поле — только через `crystal_logs`
- Не использовать GORM — только sqlc для типобезопасных запросов
- Не редактировать файлы в `internal/db/sqlc/` — они генерируются автоматически
