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

### 4.1. Миграции после MVP (004–010)

Схема выше — состояние на миграцию 003. Миграции 004–010 добавлены в ходе работы над виральностью;
SQL — в `backend/internal/db/migrations/`, требования — в `openspec/specs/`, поведение — в
`docs/features/`.

| Миграция | Что добавляет | Зачем |
| --- | --- | --- |
| `004_season_kickoff` | `season_kind` enum, `seasons.kind`, `seasons.postpone_count` | Первый сезон группы открывается сразу и раскрывается через час после того, как группа впервые стала подходящей — а не в следующую пятницу |
| `005_invite_code_short` | `UNIQUE INDEX groups_invite_code_upper_idx ON groups (upper(invite_code))` | Код приглашения из 6 символов, который можно продиктовать голосом; регистр не имеет значения |
| `006_share_attribution` | `join_source` enum, `group_members.join_source`, таблица `share_events` | Воронка «поделился карточкой → пришёл участник» измерима, а не угадывается |
| `007_referrals` | `group_members.invited_by` | Награда за приглашение достаётся тому, кто пригласил, а не админу группы |
| `008_detector_hints` | `detector_hints` с `UNIQUE(user_id, season_id, revealed_user_id)` | Детектор — лестница: бесплатная ступень (сколько голосовало), подсказки по одному, полный список |
| `009_member_safety` | `blocks`, `group_bans`, `user_reports` | Блокировка, безвозвратный выход, жалоба на человека (а не только на вопрос) |
| `010_question_tone` | `question_tone` enum, `questions.tone`, `groups.kind_only` | Группа может ограничиться добрыми вопросами; у школьников это включено по умолчанию |

Схемных изменений после 010 нет: хроника группы и рейтинг знатоков читают то, что уже хранится.

### 4.2. Статусы сезона: `CLOSED` — это раскрытый сезон

`seasons.status` идёт VOTING → REVEALED → CLOSED, причём последний переход происходит, когда
открывается **следующий** сезон (`groups.createSeasonForGroup` закрывает все REVEALED сезоны группы).
Значит в любой момент у группы не более одного REVEALED сезона, а вся более старая история — CLOSED.

**Любой запрос по истории должен брать `status IN ('REVEALED', 'CLOSED')`.** Фильтр только по
`REVEALED` возвращает ровно один сезон при любой длине истории — и выглядит при этом корректным, потому
что строки он всё-таки возвращает. Единственное исключение — `GetRevealedSeasonsForGroup`: это и есть
шаг закрытия.

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

Анонимность — про **выводимость**, а не только про поля. Убрать `voter_id` из ответа необходимо, но
недостаточно: **производное** число может восстановить голос в сочетании с тем, что API уже публикует.

- Победители по вопросам публичны (сводка раскрытия, хроника группы). Значит любая персональная
  метрика «как часто совпадал с группой», поставленная рядом, сужает или прямо задаёт голоса участника.
- Если метрика — скользящее среднее, её можно **вычесть по неделям** и получить число совпадений за
  одну неделю. Вес (`seasons_played`) при этом сам по себе безобиден, но только пока рядом нет метрики.
- Поэтому `stats.guess_accuracy` отдаётся **только владельцу профиля**, и отсутствует, а не ноль:
  `0` — это утверждение «ни разу не совпал», и оно ложное.
- **Публиковать порядок, а не измерение.** Рейтинг знатоков отдаёт место — место к голосу арифметикой
  не сводится. Прежде чем добавлять любую персональную метрику, нужно спросить, с чем её можно сложить.

- `votes` таблица хранит `voterId` — это нужно для детектора
- **Детектор возвращает только список `voterId`** — без привязки к конкретным вопросам или ответам
- API голосования **никогда** не возвращает `voterId` в результатах
- `SeasonResult` не содержит `voterId`

### Сезоны
Расписание сезонов живёт в `internal/schedule` — одно место и для воркера, и для API.

- Первый сезон группы — **kickoff** (`seasons.kind = 'KICKOFF'`): открыт сразу и раскрывается через час
  после того, как группа впервые стала подходящей, а не в следующую пятницу. Ждать до пятницы на пустой
  группе — это и есть «холодный старт», от которого новые группы умирали
- Еженедельные сезоны раскрываются в ближайшую пятницу 20:00 МСК, но не раньше чем через 48 часов после
  открытия (`schedule.MinVotingWindow`) — иначе сезон, созданный в четверг, раскрылся бы за сутки
- Последующие сезоны создаёт asynq job (не BullMQ — это Go, а не Node)
- Группа без сезона в статусе `VOTING` получает новый автоматически; с < 3 участников — нет

### Reveal
Правила раскрытия живут в `internal/eligibility` — на них гейтится воркер и ими же приложение объясняет
ожидание. Двух копий быть не должно: иначе объяснение разойдётся с поведением.

- Пятница 20:00 МСК (17:00 UTC) — целевое время, не гарантия
- **Абсолютные полы: ≥ 3 участника и ≥ 3 проголосовавших** (`MinMembers`, `MinVoters`). Ниже них сезон не
  раскрывается никогда, в том числе через повторные попытки: карточка была бы пустой, а проценты выдали
  бы конкретных голосующих
- Кворум: ≥ 50% участников (≥ 40% для групп < 8). Это сверх полов, а не вместо них
- Сезон, который дошёл до своего времени и не может раскрыться, **переносится** на следующую пятницу
  (`seasons.postpone_count`), а голоса сохраняются — выбросить их значит наказать тех, кто проголосовал
- Состояние ожидания приложение показывает через `reveal_state` (`eligibility.StateFor`): отсутствующий
  пол важнее переноса, потому что «нужно ещё 2 голоса» — действие, а «перенесено» — нет
- После раскрытия: статус → `REVEALED`, пишутся `season_results`, считаются ачивки, уходят пуши,
  ставится job генерации карточек

### Группы
- Один пользователь: не более 10 групп (MVP)
- Размер группы: 5–50 участников (MVP)
- Группа активируется при ≥ 3 участниках
- Администратор = создатель. Дополнительные привилегии: название группы, привязка Telegram, добавление
  вопросов, настройка «только добрые вопросы», бан участника
- **Пороги размера группы** живут в `internal/eligibility` и больше нигде: 3 участника — сезон вообще
  может раскрыться (`MinMembers`), 5 — открывается детектор и проценты перестают выдавать голосующих
  (`MinDetectorMembers`). Выше 5 ничего не меняется, и приложение так и говорит — придумывать цель
  нельзя. Рост квоты с 40% до 50% на 8 участниках порогом **не считается**: он делает раскрытие
  сложнее, и продавать его как награду — ложь в пользу пользователя.
- Порог детектора считается от участников, которых читатель реально может оценивать (membership минус
  блокировки в обе стороны **внутри этой группы**); порог раскрытия — от фактического состава, как его
  считает `eligibility.Evaluate`. Путать их нельзя: иначе на одном экране окажутся два противоречащих
  утверждения.

### Тон вопросов
- У каждого вопроса есть тон: `WARM`, `NEUTRAL`, `EDGY` (`internal/service/questions/tone.go`)
- `groups.kind_only` исключает `EDGY` из выбора вопросов сезона. По умолчанию включено для создателя
  младше 18 (неизвестный год рождения считается «младше 18», как и для `ROMANCE`); явное значение в
  запросе всегда побеждает
- Настройка применяется **со следующего сезона**: переписать открытый сезон значит сменить вопросы под
  теми, кто уже ответил
- Категории `HOT` и `SECRETS` целиком `EDGY`, поэтому `kind_only` вместе только с ними отклоняется
  (`NO_KIND_CATEGORIES`) — иначе сезон был бы пустым

### История группы
- Хроника (`GET /groups/:id/chronicle`) — что происходило в прошлых сезонах: по каждому вопросу тот, кто
  его возглавил. Считается из `season_results`, отдельной таблицы нет
- Рейтинг знатоков отдаёт **место и число сезонов, но не саму точность**. Хроника публикует победителей
  по вопросам, поэтому крайняя точность восстанавливает конкретные голоса участника; а поскольку
  хранится скользящее среднее с весом `seasons_played`, две соседние недели разностью дают число
  совпадений за неделю. Место арифметике не поддаётся
- Рейтинг скрыт, пока у группы меньше двух раскрытых сезонов: после одного скользящее среднее ещё ничего
  не усреднило

### Кристаллы
- Баланс — сумма всех `crystal_logs.delta` пользователя (отдельного поля нет)
- Проверка баланса и списание атомарны в одной SQL-транзакции (`database/sql`, не Prisma — это Go)
- Начисления идемпотентны через `crystal_logs.external_id`: повторный вызов не удваивает начисление
- **Детектор — лестница, а не одна покупка.** Бесплатная ступень: сколько человек голосовало, без имён —
  её задача создать вопрос. Подсказка: один голосующий за `DetectorHintCost`. Полный список:
  `DetectorFullCost`. Раскрытые подсказки лежат в `detector_hints` с `UNIQUE(user_id, season_id,
  revealed_user_id)`, поэтому дважды за одного человека не спишется
- Ни одна ступень не продаётся в группе меньше `eligibility.MinDetectorMembers` — список «все кроме
  тебя» анонимности не даёт, и продавать его нельзя
- Открытие скрытых атрибутов: 5 💎
- Есть бесплатные источники: приветственное начисление и награда за приглашение (`007_referrals` —
  награда идёт тому, кто пригласил). Без них у школьника без карты кристаллов не появится никогда, а
  платная механика, которую никто не может попробовать, не монетизирует

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
- **Цветовая схема:** дизайн-система с токенами, тёмная тема — референсная.
  Акцент `#7C3AED` (заливка контролов) / `#9B6DFF` (акцент на тёмном фоне), второй
  голос — кислотный лайм для стриков и кристаллов. Полная палитра, шкалы отступов,
  радиусов, типографики и моушена: `docs/features/design-system.md`.
  Экраны не содержат hex-литералов — только токены.
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
  T02 — схема БД (golang-migrate + sqlc) и seed вопросов
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
