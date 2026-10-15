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

### "All time" and season history include CLOSED seasons

`seasons.status` goes VOTING → REVEALED → CLOSED, and the last step happens when the *next* season opens
(`groups.createSeasonForGroup`). A group therefore has at most one REVEALED season at any moment, and
every older season that legitimately revealed is CLOSED.

`GetUserSeasonHistory` and `GetTopAttributeAllTime` filtered on `status = 'REVEALED'`, so the "season
history" and the "all-time top attribute" covered a single season — the current week — while appearing
to work, because they returned rows. Both now match `status IN ('REVEALED', 'CLOSED')`. Any new query
over a group's past must do the same; see `docs/features/groups.md` → **Chronicle**.

### Who may read `guess_accuracy`

**Only its owner.** The field is present on one's own profile and **absent** (not zero) on anyone
else's. Two routes make another member's figure a way to work out how they voted:

1. **Winners are published.** The reveal summary (`group_summary.top_per_question`) and the group
   chronicle both name which member led each question. A member at or near 100% voted for those
   winners; at or near 0% they voted for someone else every time. Over 3–10 questions that pins or
   badly narrows each individual vote.
2. **The figure is a rolling average, so it can be differenced.** It is recomputed at every reveal as
   `new = (old*w + seasonAccuracy) / (w + 1)` with `w = min(seasons_played, 4)`, and `seasons_played`
   is returned next to it. Reading the same profile on two consecutive weeks solves for that week's
   accuracy, and therefore for how many of that week's questions the member matched.

Absent rather than zeroed, because `0` is indistinguishable from a member who genuinely matched
nothing — a client would render a false claim instead of nothing. The decision is made in
`profile.GetProfile`, where the viewer and the viewed member are already both parameters, rather than
in the handler: a privacy decision should not be something the next caller of the service opts into.

`seasons_played`, `voting_streak`, `max_voting_streak`, `total_votes_cast` and `total_votes_received`
stay visible to every member. None of them counts *matches*, so none can be combined with the published
winners to recover a vote. Note that `seasons_played` is the weight `w` above: harmless on its own, and
only harmless because the figure it weighted is no longer published. If the figure is ever reintroduced
in any form, that pairing is the first thing to re-examine.

The general rule, in `CLAUDE.md`: **prefer publishing an order over a measurement.** The chronicle's
знатоки standing is the same comparison done safely — it gives a rank, which cannot be arithmetically
reduced to a vote.

### guess_accuracy and the знатоки standing

`stats.guess_accuracy` is the share of this member's votes that matched the result the group arrived at,
kept in `user_group_stats` and recomputed for every member at every reveal
(`internal/service/achievements`). It is a **rolling average over a 5-season window**: the previous value
is weighted by `min(seasons_played, 4)` against the new season's accuracy, so a single week moves it
without erasing the history.

The group chronicle ranks members by this same stored value rather than recomputing it — recomputing
would produce a number that disagrees with the one shown here. Two consequences follow from the rolling
window:

- The standing is **withheld below two revealed seasons** (`chronicle.MinSeasonsForStanding`): after one
  season the average has averaged nothing, and a ranking built on it reports luck as knowledge.
- Every standing entry carries `seasons_played`, because an accuracy over two seasons and one over ten
  are not comparable figures.

A member with no `user_group_stats` row has not been through a reveal in that group and is **unranked**,
not ranked at zero. See `docs/features/groups.md` → **Знатоки standing**.

### Mobile

`UserStats.guessAccuracy` is **nullable**, and the "Точность угадывания" tile is omitted when it is
absent. The client renders no fallback: a "0%" would read as "never guessed right", which is a
different and false claim about another member.

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
