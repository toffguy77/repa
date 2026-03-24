# Member Profile

## Overview

The member profile screen shows a user's stats, achievements, legend text, and season history within a group. Accessible by tapping on any member in the group screen or members reveal screen.

## API Endpoint

### `GET /api/v1/groups/:id/members/:userId/profile`

Returns the full profile of a group member.

- **Success 200:** `{ "data": ProfileResponse }`
- **Error 404:** `NOT_FOUND` — user not found
- **Error 403:** `NOT_MEMBER` — target user is not a member of the group
- **Auth:** requester must be a member of the group (JWT user is checked, not just target)

### Response DTO

```json
{
  "user": {
    "username": "string",
    "avatar_emoji": "string?",
    "avatar_url": "string?"
  },
  "stats": {
    "seasons_played": 5,
    "voting_streak": 3,
    "max_voting_streak": 5,
    "guess_accuracy": 72.5,
    "total_votes_cast": 25,
    "total_votes_received": 18,
    "top_attribute_all_time": {
      "question_text": "string",
      "percentage": 45.5
    }
  },
  "achievements": [
    {
      "type": "SNIPER",
      "metadata": {},
      "earned_at": "2026-03-15"
    }
  ],
  "legend": "username — настоящий Снайпер и Легенда группы",
  "season_history": [
    {
      "season_id": "uuid",
      "season_number": 5,
      "top_attribute": "Кто первым убежит при пожаре?",
      "category": "FUNNY",
      "percentage": 45.5
    }
  ]
}
```

### Null-safety: top_attribute_all_time

`GetTopAttributeAllTime` query may return `sql.ErrNoRows` if the user has no votes — this is treated as empty data (nil). Other DB errors propagate normally. Fixed in commit `a3e8377`.

## Legend Generation

Backend generates a short (max 150 chars) text description based on the user's achievements and stats. Priority order:

1. LEGEND + SNIPER -> "настоящий Снайпер и Легенда группы"
2. LEGEND -> "Легенда группы, неизменный лидер"
3. TELEPATH -> "читает мысли участников"
4. ORACLE -> "Оракул, который всегда прав"
5. SNIPER -> "Снайпер, попадающий в цель"
6. MONOPOLIST -> "Монополист, которого не перепутаешь"
7. STREAK_VOTER -> "не пропускает ни одного сезона"
8. NIGHT_OWL -> "Ночная сова, голосует под звёздами"
9. RECRUITER -> "душа компании, привёл друзей"
10. seasons >= 5 -> "опытный участник, N сезонов за плечами"
11. seasons > 0 -> "начинает свой путь в группе"
12. Fallback -> "загадочная личность"

## Mobile (Flutter)

### Screen: MemberProfileScreen (`/groups/:id/members/:userId`)

**Sections:**
1. **Header** — avatar (64px) + username + top attribute as subtitle
2. **Legend** — italic text in purple-light container
3. **Stats** — 2-column grid of StatCards with count-up animation:
   - Seasons played, Voting streak, Guess accuracy, Votes received, Top attribute %, Max streak
4. **Achievements** — horizontal scroll of AchievementBadge widgets (emoji + name + date)
5. **Season history** — last 5 seasons as mini cards (season number, top attribute, percentage)

### Navigation

- Group screen: tap on member row -> push `/groups/:id/members/:userId`
- Members reveal screen: tap on member card -> push `/groups/:id/members/:userId`

### Architecture

```
mobile/lib/features/profile/
├── data/profile_repository.dart
├── domain/profile.dart                    # Freezed models (MemberProfile, UserStats, etc.)
└── presentation/
    ├── profile_notifier.dart              # StateNotifier + Riverpod provider
    ├── member_profile_screen.dart         # Main screen
    └── widgets/
        ├── stat_card.dart                 # Animated stat display
        └── achievement_badge.dart         # Achievement icon with locked/unlocked state
```

### Key Dependencies

- Profile handler -> Profile service -> sqlc Queries (GetUserProfileInfo, GetUserGroupStats, GetUserAchievements, GetTopAttributeAllTime, GetUserSeasonHistory)
- Reuses MemberAvatar widget from groups feature
- Reuses achievement emoji/name mappings (duplicated from achievement_popup.dart for widget independence)
