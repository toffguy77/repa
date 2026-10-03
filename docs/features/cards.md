# Card Generation

## Overview

After each Reveal, the system generates a PNG reputation card for every group member. Cards are rendered server-side using chromedp (headless Chrome in Go), uploaded to S3, and cached in the `card_cache` table. Cards can be shared on social media or via Telegram.

## API Endpoints

All endpoints require Bearer JWT authentication.

### `GET /api/v1/seasons/:seasonId/my-card-url`
Get the current user's card image URL for a season.
- **Success 200:** `{ "data": { "image_url": "https://...", "status": "ready" } }`
- **Generating:** `{ "data": { "image_url": null, "status": "generating" } }` — card not yet ready
- **Behavior:** Returns the cached card URL if available, otherwise indicates the card is still being generated.

### Card URL in Reveal response
The `GET /api/v1/seasons/:seasonId/reveal` response includes `card_image_url` in the `my_card` object. This is populated from `card_cache` and may be empty if card generation hasn't completed yet.

## Data Model

### Tables

- **card_cache** — id, user_id (FK users), season_id (FK seasons), image_url, created_at. UNIQUE(user_id, season_id). Stores the S3 URL of the generated card image.

### Key Queries (sqlc)

- `UpsertCardCache` — insert or update card URL for a (user, season) pair
- `GetCardCache` — get cached card URL by user_id + season_id

## Card Design

The card mirrors the mobile app's **dark** palette so the artifact a user shares and the app
they came from read as one product. Values live in one `cardPalette` block in
`internal/service/cards/template.go`; see `docs/features/design-system.md` for the Dart↔Go
pairing and its known drift limitation.

| Card value | Hex | Mirrors (AppColorTokens.dark) |
|---|---|---|
| gradient 0% | `#0b0712` | `canvas` |
| gradient 50% | `#1f1733` | `surfaceRaised` |
| gradient 100% | `#161022` | `surface` |
| attribute bar start, decorative circles | `#9b6dff` / `#7c3aed` | `accent` / `accentFill` |
| body text | `#f6f3ff` | `textPrimary` |
| secondary text | `#afa3cc` | `textSecondary` |

- **Dimensions:** 1080x1920px (9:16 story format)
- **Background:** Dark violet gradient from the palette above, with subtle decorative circles
- **Content (top to bottom):**
  - Logo: eggplant emoji + "РЕПА" text
  - Avatar emoji in a circle
  - Username (bold, 64px)
  - Reputation title (40px, slightly transparent)
  - Top 3 attributes with percentage bars (purple gradient fill)
  - Footer: group name + season number
  - **Call to action:** the group's invite code (grouped `AB2 CD3`), a QR of the invite link,
    and the line "Скачай Репу и введи код"

### Call to action

The card doubles as the group's invitation — a viewer can act on it without any other message,
which is the point: the card is the product's only organic acquisition channel.

- **QR** is rendered in-process by `github.com/skip2/go-qrcode` into a base64 PNG data URI
  (`internal/service/cards/qr.go`), not fetched from a chart service: an outbound dependency in
  the render path would be a new failure mode and would hand every group's invite link to a
  third party.
- **Medium error correction**, because a card gets viewed on screens and re-photographed.
- **The QR sits on a white plate** with a quiet zone even on the dark card. Scanners are far
  more reliable with normal polarity than inverted, so this is the one place the dark palette
  gives way to function.
- **A QR failure degrades, never fails.** `InviteQRDataURI` returns an empty string (not an
  error) and the template omits the `<img>`; the printed code and the instruction remain, so the
  card is still a usable invitation. A decoration must not cost a member their Reveal.
- The QR encodes `https://repa.app/join/{code}?s=card&ref={memberID}`, so joins arriving through
  it are attributed to the card **and** to the member who shared it. The card is generated per
  member, which is what makes the second part possible: a group's invite code identifies the
  group, not the inviter, so without this a referral reward would have to go to the admin.
- **Trade-off:** a publicly shared card exposes the sharer's user id. It is a random UUID that
  grants nothing without authentication, and it is already visible to fellow group members
  through the members API. The alternative — a per-user-per-group referral token — adds a table
  and a lookup to protect an identifier that is not secret.

### `CardData` fields

| Field | Meaning |
|---|---|
| `InviteCode` | the group's live invite code; an empty value omits the whole CTA block |
| `InviteQRDataURI` | base64 PNG data URI, or empty when rendering failed |

## Sharing and attribution

### `POST /api/v1/seasons/:seasonId/shares`
Records that the member shared their card.
- **Body:** `{ "channel": "card" | "telegram" | "link" | "code" }`
- **Success 200:** `{ "data": { "recorded": true } }`
- **Error 400:** `SEASON_NOT_REVEALED` — there is no card to share yet
- **Error 403:** `NOT_MEMBER`
- **Behavior:** an unrecognised channel is recorded as `other` rather than refused — the share
  has already happened in the OS share sheet by the time the server hears about it. A storage
  failure is logged and swallowed for the same reason; the endpoint still validates membership
  and season state, which are real conditions rather than telemetry.

Share text is built client-side by `buildShareText`
(`mobile/lib/features/reveal/domain/share_link.dart`) and carries
`https://repa.app/join/{code}?s={channel}` — the code as the path segment the router already
handles, the channel as a query parameter `NormalizeInviteCode` strips. The pre-change text was
the constant `Моя репа repa.app`, which carried no way back into the product.

### Funnel

`GET /api/v1/admin/stats` reports `shares_by_channel` and `joins_by_source` side by side, so
the loop from a shared card to a joined member can be judged rather than guessed. A failure of
either query leaves the key present but empty rather than taking down the stats page.

`share_events` deliberately holds no vote data: a table keyed by user and season must not
acquire anything correlatable with answers.

## Attribute ordering

A card's attributes are ordered by percentage, with tone breaking near-ties: within a band of
**5 percentage points** (`cardorder.BandPoints`) the warmer attribute leads. Tone never reorders past
a clearly stronger result — a person who overwhelmingly got one edgy attribute sees it first, because
the card's value is that it is true. But at 40% and 38% the order was arbitrary anyway, and the card is
what the person is invited to share, so the arbitrary choice favours what they would want to send.

Five points is chosen against group size: in groups of 5–50 a single vote is worth 2–20 points, so the
band is wide enough to break the ties that actually occur and narrow enough that it cannot flip a real
difference.

The rule lives in `internal/cardorder` (`cardorder.ByTone`), not in the reveal service, because four
places build a card from the same results and must agree:

| Path | Where |
| --- | --- |
| A member's own card | `reveal.splitAttributes` |
| The group's members-cards list | `reveal.GetMembersCards` (carries `question_tone` across from `GetAllSeasonResultsWithUsers`) |
| The purchased "show everything" view | `reveal.OpenHidden` |
| The shared PNG | `cards.cardTopAttributes`, used by the `cards:generate` worker |

The trend line and the reputation title are both taken from the attribute the card *leads with*, not
the highest percentage, so they describe what the reader sees first.

The PNG matters most — it is the artifact a person is invited to send — and it is produced furthest
from the API, which is why the rule is not a private helper of the reveal service. The failure mode is
quiet: an unset tone reads as neutral for every row, which is indistinguishable from plain percentage
order and fails nothing. Each path therefore has its own test.

The ordering is **not** a comparison function, because "within 5 points" is not transitive — 50% edgy,
47% neutral and 44% warm would give warm < neutral < edgy < warm, and sorting on an inconsistent
comparator yields an undefined order that can lift the 44% attribute above the 50% one. Instead
`cardorder.ByTone` sorts by percentage and walks the list into *anchor clusters*: a cluster holds
the attributes within the band of the cluster's own strongest attribute, and tone orders only inside a
cluster. The anchor is fixed, so nothing is ever promoted past an attribute more than 5 points
stronger, however long the chain of near-ties.

Ordering happens before the top-three / hidden split, so tone decides who is on the shared card rather
than only the order within it. Nothing is removed: the result is a permutation of its input, and the
hidden-attributes purchase still shows everything.

## Business Rules

- Cards are generated asynchronously after reveal via `cards:generate` asynq task.
- One browser instance is reused for all cards in a season (tab-per-card).
- If card generation fails for a user, it is skipped (logged as error) — other users' cards still generate.
- Cards are uploaded to S3 at path `cards/{seasonId}/{userId}.png`.
- Card URL is upserted into `card_cache` (idempotent).

## Worker Jobs

### `cards:generate` (queued after reveal, payload: seasonID)
1. Fetch season, group, and all members
2. Start headless Chrome browser (single instance for all cards)
3. For each member:
   - Fetch season results (top 3 attributes)
   - Build HTML template
   - Render to PNG via chromedp (SetDocumentContent + FullScreenshot)
   - Upload to S3
   - Upsert card_cache

## Architecture

```
backend/
├── internal/
│   ├── handler/reveal/handler.go          # GetMyCardURL endpoint
│   ├── service/cards/
│   │   ├── service.go                     # GenerateCardsForSeason, GetCardURL, renderHTMLToPNG
│   │   ├── template.go                    # BuildCardHTML, CardData, escapeHTML
│   │   └── template_test.go              # Template rendering + XSS tests
│   ├── worker/tasks/cards.go              # HandleCardsGenerate asynq task handler
│   └── db/
│       └── queries/card_cache.sql         # UpsertCardCache, GetCardCache
```

### Key Dependencies

- Cards service -> sqlc Queries (season results, card_cache, group, members)
- Cards service -> S3 client (upload PNG)
- Cards service -> chromedp (headless Chrome rendering)
- Reveal worker -> enqueues `cards:generate` after successful reveal
- Reveal handler -> Cards service (GetMyCardURL endpoint)
- Reveal service -> card_cache query (populates CardImageURL in GetReveal)
