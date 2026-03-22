# Achievements

## Overview

The achievements engine calculates and awards achievements to group members after each Reveal. It runs as an asynq worker task (`achievements:calculate`) enqueued by the reveal processor. It also updates `user_group_stats` (voting streaks, accuracy, totals) for each member.

## Achievement Types

### Accuracy-based (require voting)

| Type | Condition | Repeatable |
|---|---|---|
| SNIPER | Voted for the winner on >= 80% of questions | No |
| TELEPATH | Voted for the winner on 100% of questions | No |
| BLIND | Accuracy < 20% | No |
| ORACLE | Accuracy > 70% for 3 consecutive seasons | No |

"Winner" = the target who received the most votes for a given question.

### Reputation-based (from season_results)

| Type | Condition | Metadata | Repeatable |
|---|---|---|---|
| MONOPOLIST | Received >= 70% of votes on a single question | question_id, question_text, percentage | No |
| PIONEER | First person in the group to be #1 on a question that never had a top result before | question_id, question_text | No |
| LEGEND | Same top attribute for 5 consecutive revealed seasons | - | No |

### Activity-based

| Type | Condition | Metadata | Repeatable |
|---|---|---|---|
| STREAK_VOTER | Voted N seasons in a row (milestones: 5, 10, 20) | streak | Yes (at each milestone) |
| FIRST_VOTER | First person to complete all votes in the season | - | No |
| NIGHT_OWL | First vote cast between 23:00-03:00 MSK | - | No |

### Social

| Type | Condition | Repeatable |
|---|---|---|
| RECRUITER | 3+ members joined the group after this user | No |

### Not yet implemented

| Type | Condition | Status |
|---|---|---|
| EXPERT_OF | Correctly predicted a specific member's wins 5+ times in a row across seasons | TODO |

## Worker Job

### `achievements:calculate` (queued, payload: season_id)

1. Load season + group members
2. Compute winner per question (most votes)
3. For each member:
   - Check accuracy-based achievements (SNIPER, TELEPATH, BLIND, ORACLE)
   - Check reputation-based achievements (MONOPOLIST, PIONEER, LEGEND)
   - Check activity-based achievements (FIRST_VOTER, NIGHT_OWL)
   - Check social achievements (RECRUITER)
4. Update `user_group_stats` for all members:
   - `seasons_played++`
   - `voting_streak` incremented if voted, reset to 0 if not
   - `max_voting_streak` updated if new streak exceeds it
   - `guess_accuracy` — rolling weighted average (capped at 5-season window)
   - `total_votes_cast` and `total_votes_received` incremented
   - Grant STREAK_VOTER at milestones (5, 10, 20)

### Deduplication

- Most achievements are granted once per group (checked via `HasAchievementInGroup`)
- STREAK_VOTER is the exception — granted at each milestone

### Integration with Reveal API

`GET /api/v1/seasons/:seasonId/reveal` returns `new_achievements` in the `my_card` response — lists achievements earned in the current season for the requesting user.

## Data Model

### Tables

- **achievements** — id, user_id, group_id, season_id (nullable for non-season achievements like RECRUITER), achievement_type (enum), metadata (JSONB), earned_at
- **user_group_stats** — id, user_id, group_id, seasons_played, voting_streak, max_voting_streak, guess_accuracy, total_votes_cast, total_votes_received, updated_at

### Key Queries (sqlc)

- `HasAchievementInGroup` — dedup check: COUNT by user_id, group_id, achievement_type
- `CreateAchievement` — insert with JSONB metadata
- `GetSeasonAchievements` — all achievements for a season (used by reveal API)
- `GetWinnerPerQuestion` — DISTINCT ON (question_id), highest vote_count per question
- `GetVotesByVoterInSeason` — user's votes (question_id, target_id) for accuracy calc
- `GetFirstCompletedVoter` — voter who finished all questions first (by last_vote_at)
- `GetFirstVoteTimeByUser` — MIN(created_at) for night owl check
- `GetLastNRevealedSeasons` — for ORACLE and LEGEND history
- `GetTopAttributeForUser` — top result by percentage for a user in a season
- `CountMembersJoinedAfterUser` — for RECRUITER
- `UpsertUserGroupStats` — ON CONFLICT upsert for stats
- `CountVotesCastByUser`, `CountVotesReceivedByUser` — for stats update

## Architecture

```
backend/
├── internal/
│   ├── service/achievements/
│   │   ├── service.go           # CalculateAchievements, all check* functions, updateGroupStats
│   │   └── service_test.go      # 12 tests covering all achievement types + stats
│   ├── worker/tasks/
│   │   └── achievements.go      # HandleAchievements asynq task handler
│   └── db/
│       ├── queries/achievements.sql
│       ├── queries/votes.sql        # GetWinnerPerQuestion, GetVotesByVoterInSeason, etc.
│       ├── queries/season_results.sql # GetTopAttributeForUser
│       ├── queries/seasons.sql       # GetLastNRevealedSeasons
│       └── queries/groups.sql        # CountMembersJoinedAfterUser, GetGroupMemberIDs
```

### Key Dependencies

- Reveal worker (T10) enqueues `achievements:calculate` after successful reveal
- Achievement service -> sqlc Queries (no external dependencies)
- Reveal API reads achievements via `GetSeasonAchievements` to populate `new_achievements`
