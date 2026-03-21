# T02 — SQL схема, миграции, sqlc, seed

## Цель
Создать SQL-миграции, настроить sqlc для кодогенерации, засидировать 200+ вопросов.

## Миграции (internal/db/migrations/)

### 001_init.up.sql
Полная схема из master context раздела 4 — все CREATE TABLE, CREATE TYPE, индексы.
Скопировать SQL из master context дословно.

### 001_init.down.sql
```sql
DROP TABLE IF EXISTS next_season_votes, push_preferences, reports, reactions,
  card_cache, fcm_tokens, crystal_logs, detectors, user_group_stats,
  achievements, season_results, votes, season_questions, questions,
  seasons, group_members, groups, users CASCADE;
DROP TYPE IF EXISTS push_category, achievement_type, crystal_log_type,
  question_status, question_source, question_category, season_status;
```

## sqlc queries (internal/db/queries/)

### users.sql
```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1;

-- name: GetUserByAppleID :one
SELECT * FROM users WHERE apple_id = $1;

-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (id, phone, apple_id, google_id, username, avatar_emoji, birth_year)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users SET username = $2, avatar_emoji = $3, avatar_url = $4,
  birth_year = $5, updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

### groups.sql
```sql
-- name: CreateGroup :one
INSERT INTO groups (id, name, invite_code, admin_id, telegram_chat_username)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: GetGroupByInviteCode :one
SELECT * FROM groups WHERE invite_code = $1;

-- name: GetUserGroups :many
SELECT g.* FROM groups g
JOIN group_members gm ON gm.group_id = g.id
WHERE gm.user_id = $1
ORDER BY gm.joined_at DESC;

-- name: CountUserGroups :one
SELECT COUNT(*) FROM group_members WHERE user_id = $1;

-- name: CountGroupMembers :one
SELECT COUNT(*) FROM group_members WHERE group_id = $1;

-- name: AddGroupMember :one
INSERT INTO group_members (id, user_id, group_id)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetGroupMembers :many
SELECT u.id, u.username, u.avatar_emoji, u.avatar_url
FROM users u
JOIN group_members gm ON gm.user_id = u.id
WHERE gm.group_id = $1
ORDER BY gm.joined_at ASC;

-- name: UpdateGroupTelegram :exec
UPDATE groups SET telegram_chat_id = $2, telegram_chat_username = $3,
  telegram_connect_code = NULL, telegram_connect_expiry = NULL WHERE id = $1;

-- name: SetGroupConnectCode :exec
UPDATE groups SET telegram_connect_code = $2, telegram_connect_expiry = $3 WHERE id = $1;

-- name: GetGroupByTelegramChatID :one
SELECT * FROM groups WHERE telegram_chat_id = $1;
```

### seasons.sql
```sql
-- name: CreateSeason :one
INSERT INTO seasons (id, group_id, number, starts_at, reveal_at, ends_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetActiveSeasonByGroup :one
SELECT * FROM seasons WHERE group_id = $1 AND status = 'VOTING' LIMIT 1;

-- name: GetSeasonByID :one
SELECT * FROM seasons WHERE id = $1;

-- name: UpdateSeasonStatus :exec
UPDATE seasons SET status = $2 WHERE id = $1;

-- name: GetSeasonsForReveal :many
SELECT * FROM seasons WHERE status = 'VOTING' AND reveal_at <= NOW();

-- name: GetGroupsNeedingNewSeason :many
SELECT DISTINCT g.* FROM groups g
JOIN group_members gm ON gm.group_id = g.id
WHERE NOT EXISTS (SELECT 1 FROM seasons s WHERE s.group_id = g.id AND s.status = 'VOTING')
GROUP BY g.id HAVING COUNT(gm.id) >= 3;
```

### votes.sql
```sql
-- name: CreateVote :one
INSERT INTO votes (id, season_id, voter_id, target_id, question_id)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetVotesBySeasonAndVoter :many
SELECT * FROM votes WHERE season_id = $1 AND voter_id = $2;

-- name: CountUniqueVoters :one
SELECT COUNT(DISTINCT voter_id) FROM votes WHERE season_id = $1;

-- name: GetVotersBySeason :many
SELECT DISTINCT voter_id FROM votes WHERE season_id = $1;

-- name: AggregateVotesByTarget :many
SELECT target_id, question_id, COUNT(*) as vote_count
FROM votes WHERE season_id = $1
GROUP BY target_id, question_id;
```

Аналогично создать queries для: questions, season_results, achievements,
user_group_stats, detectors, crystal_logs, fcm_tokens, reactions, reports.

## cmd/seed/main.go

```go
package main

import (
    "context"
    "github.com/repa-app/repa/internal/config"
    // db connection, sqlc queries
)

func main() {
    cfg := config.Load()
    // Подключиться к БД
    // Вставить вопросы через INSERT ... ON CONFLICT (text) DO NOTHING
    seedQuestions(ctx, q)
}

var questions = []struct {
    Text     string
    Category string
}{
    // HOT — 40 вопросов
    {"Кто первым побежит при пожаре?", "HOT"},
    {"Кто наябедничает учителю первым?", "HOT"},
    {"Кто скорее всего съест чужую еду из холодильника?", "HOT"},
    {"Кто притворяется, что всё знает, но на самом деле нет?", "HOT"},
    {"Кто тайно читает чужие переписки?", "HOT"},
    {"Кто спишет на контрольной и не скажет?", "HOT"},
    {"Кто первым предаст в зомби-апокалипсис?", "HOT"},
    {"Кто громче всех кричит, когда видит насекомое?", "HOT"},
    // ... итого 40

    // FUNNY — 50 вопросов
    {"Кто заблудится в трёх соснах?", "FUNNY"},
    {"Кто будет разговаривать с едой перед тем как её съесть?", "FUNNY"},
    {"Кто переименует кота в честь любимого персонажа аниме?", "FUNNY"},
    {"Кто будет смеяться на похоронах?", "FUNNY"},
    {"Кто потеряет телефон у себя в кармане?", "FUNNY"},
    // ... итого 50

    // SECRETS — 40 вопросов
    {"Кто тайно слушает попсу?", "SECRETS"},
    {"Кто делает вид что не смотрит сериалы, но смотрит?", "SECRETS"},
    {"У кого дома больше всего беспорядка?", "SECRETS"},
    {"Кто втайне читает гороскопы?", "SECRETS"},
    // ... итого 40

    // SKILLS — 35 вопросов
    {"Кто выживет последним в зомби-апокалипсис?", "SKILLS"},
    {"Кто лучше всех умеет врать?", "SKILLS"},
    {"Кто первым напишет книгу?", "SKILLS"},
    // ... итого 35

    // ROMANCE — 20 вопросов
    {"Кто влюбится первым этим летом?", "ROMANCE"},
    {"Кто дольше всех будет переписываться ночью?", "ROMANCE"},
    // ... итого 20

    // STUDY — 20 вопросов
    {"Кто спишет на контрольной и не попадётся?", "STUDY"},
    {"Кто будет учить всё в последний день?", "STUDY"},
    {"Кто получит пятёрку не зная материала?", "STUDY"},
    // ... итого 20
}
```

**Важно:** Seed должен содержать реальные смешные вопросы, минимум по 10 на категорию,
итого ≥ 200. Не заглушки.

## Команды
```bash
# Генерация Go кода из SQL
make sqlc

# Применить миграции
make migrate

# Засидировать вопросы
make seed
```

## Критерии готовности
- [ ] `make migrate` создаёт все таблицы без ошибок
- [ ] `sqlc generate` создаёт Go файлы в `internal/db/sqlc/`
- [ ] `make seed` загружает ≥ 200 вопросов
- [ ] Повторный seed не дублирует (ON CONFLICT DO NOTHING)
- [ ] Все типы и индексы созданы
