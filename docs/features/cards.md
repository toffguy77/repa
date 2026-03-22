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

- **Dimensions:** 1080x1920px (9:16 story format)
- **Background:** Dark purple gradient (#1e1033 to #1a0d2e) with SVG, subtle decorative circles
- **Content (top to bottom):**
  - Logo: eggplant emoji + "PEPA" text
  - Avatar emoji in a circle
  - Username (bold, 64px)
  - Reputation title (40px, slightly transparent)
  - Top 3 attributes with percentage bars (purple gradient fill)
  - Footer: group name + season number

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
