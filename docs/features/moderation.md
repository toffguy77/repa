# AI Moderation & User Questions

## Overview

Users can submit custom questions for their group. Each question is moderated by an AI (Anthropic Claude) before becoming available. Groups also have a report system for flagging inappropriate content.

## API Endpoints

All endpoints require Bearer JWT authentication.

### `POST /api/v1/groups/:groupId/questions`
Create a custom question for the group.
- **Body:** `{ "text": "string", "category": "HOT|FUNNY|SECRETS|SKILLS|ROMANCE|STUDY" }`
- **Success 201:** `{ "data": { "question": QuestionDto } }`
- **Behavior:** Submits the question to AI moderation. The question starts as PENDING.

### `GET /api/v1/groups/:groupId/questions`
List questions available to the group.
- **Success 200:** `{ "data": { "questions": [QuestionDto] } }`
- **Behavior:** Returns system questions filtered by the group's enabled categories, merged with user-created ACTIVE questions for this group.

### `DELETE /api/v1/groups/:groupId/questions/:questionId`
Soft-delete a user-created question.
- **Success 200:** `{ "data": {} }`
- **Error 403:** Only the question author or a group admin can delete.

### `POST /api/v1/groups/:groupId/questions/:questionId/report`
Report a question as inappropriate.
- **Body:** `{ "reason": "string?" }` (optional reason text)
- **Success 200:** `{ "data": {} }`
- **Constraint:** One report per user per question.

## AI Moderation Flow

1. User submits a question via `POST /groups/:groupId/questions`
2. Backend calls Anthropic API to evaluate the question
3. Three possible outcomes:
   - **Approved** -> question status set to `ACTIVE`, immediately available
   - **Rejected** -> question status set to `REJECTED`, not shown to users
   - **Timeout/error** -> question stays `PENDING`, available for manual review

### Tone is assigned independently of the verdict

Every stored question also gets a tone (`questions.tone` — `WARM`, `NEUTRAL`, `EDGY`), assigned by
`ClassifyTone(text, category)` in `internal/service/questions/tone.go`. This is **not** part of the
allow/reject decision and does not depend on it:

- The moderator decides whether the question may exist at all. Tone decides which groups may be asked
  it. A question the moderator allowed can still be a jab, and a kind-only group must not receive it.
- A question is never rejected for being edgy, and a rejected question still gets a tone — the two
  answers are stored side by side.
- When the tone cannot be determined from the text, the category decides (SKILLS → WARM, HOT/SECRETS →
  EDGY, everything else → NEUTRAL). The result is always one of the three; nothing is stored untoned,
  and an undeterminable question stays usable.

See `docs/features/groups.md` → **Question tone** for the classification rules and the group setting
that consumes them.

## Anthropic Client

- Implementation: `internal/lib/anthropic.go` — plain HTTP client (no SDK)
- Timeout: 5 seconds
- System prompt: Russian-language instructions for evaluating question appropriateness
- Nil-safe: if the client is not configured (no API key), moderation is skipped and the question goes to PENDING

## Error Codes

| Code | Meaning |
|------|---------|
| `NOT_MEMBER` | User is not a member of the group |
| `NOT_FOUND` | Question not found |
| `FORBIDDEN` | Not the author or admin |
| `ALREADY_REPORTED` | User already reported this question |

## Architecture

```text
backend/
├── internal/
│   ├── handler/questions/handler.go    # HTTP handlers for question CRUD + report
│   ├── service/questions/service.go    # Business logic, AI moderation call
│   ├── lib/anthropic.go               # Anthropic API client (plain HTTP)
│   └── db/
│       └── queries/questions.sql       # CreateUserQuestion, ListGroupQuestions, SoftDeleteQuestion, CreateReport
```

## Person reports

Separate from question reports, because the two carry different columns and a merged feed would be
sparse in both directions.

### `GET /api/v1/admin/user-reports`
- **Query:** `page`, `limit` (max 100, default 20)
- **Success 200:** `{ "data": { "reports": [{ id, reported_id, reported_username, reporter_id,
  reporter_username, group_id, reason, created_at }], "total": number } }`
- Nullable columns are flattened to plain strings. (The older question-report endpoint still returns raw
  sqlc rows, where `sql.NullString` marshals as `{String, Valid}`; changing that is a separate, breaking
  change.)

Triage stays manual, as it already is for questions. Filing a report never notifies the reported member.

See `docs/features/groups.md` → Member safety for blocking, removal and bans.
