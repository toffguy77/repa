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
| Runtime | Node.js 22 LTS |
| Framework | Fastify 4 |
| ORM | Prisma |
| БД | PostgreSQL 16 |
| Кэш / очереди | Redis 7 (ioredis) |
| Джобы | BullMQ (поверх Redis) |
| Push | Firebase Admin SDK (FCM) |
| Хранилище | Yandex Object Storage (S3-compatible, aws-sdk v3) |
| Рендер карточек | Puppeteer (headless Chrome) |
| AI модерация | Anthropic SDK (`@anthropic-ai/sdk`) |
| Биллинг | ЮKassa REST API |
| Telegram | node-telegram-bot-api |
| Валидация | Zod |
| Тесты | Vitest |
| Язык | TypeScript strict |

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
│   ├── src/
│   │   ├── modules/          # Фичи (auth, groups, voting, reveal, ...)
│   │   │   └── {module}/
│   │   │       ├── {module}.router.ts
│   │   │       ├── {module}.service.ts
│   │   │       ├── {module}.schema.ts   # Zod схемы
│   │   │       └── {module}.test.ts
│   │   ├── lib/              # Инфраструктурные клиенты
│   │   │   ├── prisma.ts
│   │   │   ├── redis.ts
│   │   │   ├── bullmq.ts
│   │   │   ├── firebase.ts
│   │   │   ├── s3.ts
│   │   │   ├── anthropic.ts
│   │   │   └── telegram.ts
│   │   ├── jobs/             # BullMQ workers
│   │   ├── plugins/          # Fastify plugins
│   │   └── app.ts            # Fastify instance
│   ├── prisma/
│   │   ├── schema.prisma
│   │   └── seed.ts
│   ├── .env.example
│   └── package.json
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

## 4. Схема базы данных (Prisma)

```prisma
model User {
  id            String   @id @default(cuid())
  phone         String?  @unique
  appleId       String?  @unique
  googleId      String?  @unique
  username      String   @unique
  avatarUrl     String?
  avatarEmoji   String?
  birthYear     Int?
  createdAt     DateTime @default(now())
  updatedAt     DateTime @updatedAt

  memberships   GroupMember[]
  votes         Vote[]
  detectors     Detector[]
  crystalLogs   CrystalLog[]
  achievements  Achievement[]
  stats         UserGroupStat[]
  fcmTokens     FcmToken[]
}

model Group {
  id                    String   @id @default(cuid())
  name                  String
  inviteCode            String   @unique @default(cuid())
  adminId               String
  telegramChatId        String?
  telegramChatUsername  String?
  telegramConnectCode   String?
  telegramConnectExpiry DateTime?
  createdAt             DateTime @default(now())

  members   GroupMember[]
  seasons   Season[]
}

model GroupMember {
  id       String   @id @default(cuid())
  userId   String
  groupId  String
  joinedAt DateTime @default(now())

  user  User  @relation(fields: [userId], references: [id])
  group Group @relation(fields: [groupId], references: [id])

  @@unique([userId, groupId])
}

model Season {
  id         String       @id @default(cuid())
  groupId    String
  number     Int
  status     SeasonStatus @default(VOTING)
  startsAt   DateTime
  revealAt   DateTime
  endsAt     DateTime
  createdAt  DateTime     @default(now())

  group     Group          @relation(fields: [groupId], references: [id])
  questions SeasonQuestion[]
  votes     Vote[]
  results   SeasonResult[]
}

enum SeasonStatus {
  VOTING
  REVEALED
  CLOSED
}

model Question {
  id       String         @id @default(cuid())
  text     String
  category QuestionCategory
  source   QuestionSource @default(SYSTEM)
  groupId  String?        // null = системный вопрос
  authorId String?
  status   QuestionStatus @default(ACTIVE)
  createdAt DateTime      @default(now())

  seasonQuestions SeasonQuestion[]
}

enum QuestionCategory {
  HOT
  FUNNY
  SECRETS
  SKILLS
  ROMANCE
  STUDY
}

enum QuestionSource {
  SYSTEM
  USER
}

enum QuestionStatus {
  ACTIVE
  PENDING
  REJECTED
}

model SeasonQuestion {
  id         String  @id @default(cuid())
  seasonId   String
  questionId String
  order      Int

  season   Season   @relation(fields: [seasonId], references: [id])
  question Question @relation(fields: [questionId], references: [id])

  @@unique([seasonId, questionId])
}

model Vote {
  id         String   @id @default(cuid())
  seasonId   String
  voterId    String
  targetId   String
  questionId String
  createdAt  DateTime @default(now())

  season Season @relation(fields: [seasonId], references: [id])
  voter  User   @relation(fields: [voterId], references: [id])

  @@unique([seasonId, voterId, questionId])
}

model SeasonResult {
  id          String @id @default(cuid())
  seasonId    String
  targetId    String
  questionId  String
  voteCount   Int
  totalVoters Int
  percentage  Float

  season Season @relation(fields: [seasonId], references: [id])

  @@unique([seasonId, targetId, questionId])
}

model Achievement {
  id              String          @id @default(cuid())
  userId          String
  groupId         String
  seasonId        String?
  achievementType AchievementType
  earnedAt        DateTime        @default(now())
  metadata        Json?

  user User @relation(fields: [userId], references: [id])
}

enum AchievementType {
  SNIPER ORACLE TELEPATH BLIND RANDOM
  EXPERT_OF BEST_FRIEND DETECTIVE STRANGER
  LEGEND CHANGEABLE MONOPOLIST ENIGMA RISING PIONEER
  STREAK_VOTER FIRST_VOTER LAST_VOTER NIGHT_OWL ANALYST
  MEDIA CONSPIRATOR RECRUITER
}

model UserGroupStat {
  id                String @id @default(cuid())
  userId            String
  groupId           String
  seasonsPlayed     Int    @default(0)
  votingStreak      Int    @default(0)
  maxVotingStreak   Int    @default(0)
  guessAccuracy     Float  @default(0)
  totalVotesCast    Int    @default(0)
  totalVotesReceived Int   @default(0)
  updatedAt         DateTime @updatedAt

  user User @relation(fields: [userId], references: [id])

  @@unique([userId, groupId])
}

model Detector {
  id        String   @id @default(cuid())
  userId    String
  seasonId  String
  groupId   String
  createdAt DateTime @default(now())

  user User @relation(fields: [userId], references: [id])

  @@unique([userId, seasonId])
}

model CrystalLog {
  id          String          @id @default(cuid())
  userId      String
  delta       Int
  balance     Int
  type        CrystalLogType
  description String?
  externalId  String?         @unique
  createdAt   DateTime        @default(now())

  user User @relation(fields: [userId], references: [id])
}

enum CrystalLogType {
  PURCHASE SPEND_DETECTOR SPEND_ATTRIBUTES SPEND_QUESTION BONUS
}

model FcmToken {
  id        String   @id @default(cuid())
  userId    String
  token     String   @unique
  platform  String
  createdAt DateTime @default(now())

  user User @relation(fields: [userId], references: [id])
}
```

---

## 5. API соглашения

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
  T03 — Fastify приложение, плагины, базовая структура

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
- Не хранить баланс кристаллов как отдельное поле — только через `CrystalLog`
