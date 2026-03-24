# E2E Test Suite Design

## Goal
Comprehensive end-to-end tests for all 40+ backend API endpoints against real PostgreSQL and Redis via testcontainers-go. Cover happy paths, error cases, business rule validation, and anonymity guarantees.

## Approach: Testcontainers

- `testcontainers-go` spins up PostgreSQL 16 + Redis 7 per test run
- Full Echo server wired identically to `cmd/server/main.go` (minus external services: S3, FCM, Telegram, Anthropic, YuKassa)
- JWT tokens minted in-test using the same `golang-jwt/v5` signing
- Migrations applied automatically at suite startup
- System questions seeded for voting tests

## Structure

```
backend/tests/e2e/
├── setup_test.go          # TestMain, containers, server builder, helpers
├── health_test.go         # GET /health
├── auth_test.go           # OTP send/verify, get me, update profile, username check, push prefs, delete account
├── groups_test.go         # CRUD, join/leave, invite preview/regenerate, limits (10 groups, 50 members), admin-only ops
├── voting_test.go         # Get session, cast vote, progress, self-vote block, already-voted, season-not-voting
├── reveal_test.go         # Get reveal, members cards, open hidden (5 crystals), detector (10 crystals), not-revealed guard
├── crystals_test.go       # Balance, packages, webhook processing
├── reactions_test.go      # Create/get reactions, invalid emoji, self-reaction, not-revealed guard
├── questions_test.go      # Create, list, delete (author-only), report
├── profile_test.go        # Member stats endpoint
├── admin_test.go          # Basic auth, list/resolve reports, stats
├── push_test.go           # Register token, question candidates, vote question
├── anonymity_test.go      # voter_id never appears in any reveal/voting API response
```

## Test Infrastructure (`setup_test.go`)

### TestSuite struct
- Postgres + Redis testcontainers
- `*pgxpool.Pool`, `*redis.Client`
- `*echo.Echo` fully wired server
- Helper methods: `createUser()`, `createGroup()`, `joinGroup()`, `createSeason()`, `seedQuestions()`, `mintToken()`, `request()`

### Lifecycle
- `TestMain` → start containers → apply migrations → seed base data
- Each `Test*` function creates its own users/groups to avoid cross-test pollution
- Teardown: containers destroyed automatically

### Auth Helpers
- `mintToken(userID, username)` → valid JWT using test `JWT_SECRET`
- `request(method, path, body, token)` → returns `*http.Response` + parsed JSON

## Coverage Per Domain

### Auth (8 tests)
- OTP send → returns code in dev mode
- OTP verify → returns token + user
- OTP verify wrong code → 401
- Get me → user data
- Update profile (username, emoji, birth_year)
- Username check (available / taken)
- Update push preferences
- Delete account → user gone

### Groups (12 tests)
- Create group → 201, group + invite_url
- List groups → user's groups with active season
- Get group → members, season
- Join via invite code → member added
- Join preview → name, count, admin
- Leave group
- Update group (admin only, non-admin → 403)
- Regenerate invite link (admin only)
- Max 10 groups per user → CONFLICT
- Max 50 members per group → CONFLICT
- Already member → CONFLICT
- ROMANCE category blocked for under-18

### Voting (8 tests)
- Get voting session → questions + targets (self excluded)
- Cast vote → 201, progress updated
- Self-vote → 400 SELF_VOTE
- Already voted same question → CONFLICT
- Target not member → 400
- Invalid question → 400
- Season not voting → 400
- Progress endpoint → counts + quorum info

### Reveal (7 tests)
- Get reveal (after manual status flip to REVEALED) → results
- Members cards
- Open hidden (with crystals) → attributes revealed
- Open hidden (no crystals) → 402
- Buy detector (with crystals) → voter list (no question binding)
- Buy detector (no crystals) → 402
- Season not revealed → 400

### Crystals (4 tests)
- Get balance (initial = 0)
- Get packages → list
- Webhook processing (payment.succeeded) → balance updated
- Webhook idempotent (same external_id twice)

### Reactions (5 tests)
- Create reaction → success
- Get reactions → list with emojis
- Invalid emoji → 400
- Self-reaction → 400
- Season not revealed → 400

### Questions (5 tests)
- Create custom question → success
- List group questions
- Delete own question → success
- Delete other's question → 403
- Report question

### Profile (2 tests)
- Get member profile → stats
- Not a member → 403

### Admin (4 tests)
- No auth → 401
- Wrong credentials → 401
- List reports + resolve
- Get stats

### Push (3 tests)
- Register FCM token
- Get question candidates
- Vote for next-season question

### Health (2 tests)
- All healthy → 200
- Returns db + redis status

### Anonymity (3 tests)
- Reveal response never contains voter_id linked to votes
- Detector returns only voter ID list, no question binding
- Voting session response excludes voter info

**Total: ~63 test cases**

## Dependencies to Add
- `github.com/testcontainers/testcontainers-go` (+ postgres/redis modules)
- `github.com/stretchr/testify` (assertions — optional, can use stdlib)

## Make Target
```makefile
test-e2e:  ## Run e2e tests (requires Docker)
	go test -v -count=1 -timeout 300s ./tests/e2e/...
```
