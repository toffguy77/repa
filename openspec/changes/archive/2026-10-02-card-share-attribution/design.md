# Design

## Context

See proposal.md — Why. Relevant existing machinery: cards are rendered by chromedp from the
HTML that `BuildCardHTML` returns, so anything that can be expressed as HTML or a data URI can
go on the card. `cardPalette` (added by `youth-design-system`) already holds the colours.
Invite codes are short and transcribable after `short-invite-codes`, which is what makes a CTA
block possible at all. `group_members` records membership with no provenance.

## Goals / Non-Goals

**Goals:**

- A card that works as a standalone invitation — nothing else in the message is required.
- Know the ratio of shares to joins, per channel, from data rather than inference.
- Never lose a member's card because a decoration failed.

**Non-Goals:**

- Install-time attribution (deferred deep linking); the link is made, the install-side matching
  is not.
- Rewarding referrals — that is the next change, and it consumes this change's data.

## Decisions

### QR is rendered server-side into a data URI, not fetched from a service

`github.com/skip2/go-qrcode` renders a PNG in-process which is embedded as a
`data:image/png;base64,…` URI in the card HTML. The alternatives were a third-party chart URL
(rejected: an outbound dependency in the render path, and it leaks every group's invite link to
that service) and client-side rendering (rejected: the card is a server artifact, and the
mobile app is not involved in producing it).

QR error-correction is set to medium: a card is viewed on a screen or re-photographed, so a
little redundancy is worth the slightly denser code. The QR is drawn on a light plate even on
the dark card, because scanners are far more reliable with a quiet zone and normal polarity
than with an inverted one — this is the one place the dark palette gives way to function.

### A failed QR degrades the card instead of failing it

`GenerateCardsForSeason` already skips a member whose card fails and logs it. A QR failure is
not a reason to lose someone's Reveal: the renderer returns an empty data URI, the template
omits the image, and the code and instruction still appear. The card stays a usable invitation
in text form.

### `join_source` is an enum on the membership, not a separate event table

How someone arrived is a property of that membership, is written once, and is always read
alongside it — so it belongs on `group_members` as a nullable enum defaulting to `UNKNOWN`.
A join-events table would add a join to every funnel query for information that never changes.

Unrecognised values map to `UNKNOWN` rather than rejecting the join. A client sending a source
the server does not know is a version skew, and refusing the join over a telemetry field would
trade an acquisition for a datum.

### Share events are a separate table, because they are events

A share can happen many times per member per season, and is interesting as a count over time,
so `share_events(id, user_id, season_id, channel, created_at)` is append-only with no
uniqueness constraint. It deliberately holds no vote data: the anonymity rule means a table
keyed by user and season must not acquire anything that could be correlated with answers.

Recording is best-effort — the write is enqueued, and a failure is logged. The user's share has
already happened in the OS share sheet by then; failing the request would report an error for
something that succeeded.

### The share link carries the channel in a query parameter

`https://repa.app/join/{code}?s={channel}`. The code stays the path segment the router already
handles, and `NormalizeInviteCode` already strips a query string, so the link resolves whether
or not the client understands the parameter. Channels: `card`, `telegram`, `link`, `code`.

## Risks / Trade-offs

- **A visible invite code on a shared card is public** → that is the point; the group caps at 50
  members and the admin can regenerate. A member who does not want that can keep the card
  instead of sharing it, which the product already frames as their choice.
- **A QR makes the card busier** → it sits in the footer at a size that scans but does not
  compete with the attributes; the card's subject is still the person.
- **Channel markers are self-reported by the client** → fine for a funnel; they are not used for
  anything a user could gain from misreporting.
- **`share_events` grows without bound** → one row per share action is small, and nothing in the
  product reads more than aggregates; retention can be added when there is a volume problem.

## Migration Plan

1. Migration `006_share_attribution`: `join_source` enum, `group_members.join_source` defaulting
   to `UNKNOWN`, `share_events` table with an index on `(season_id, channel)`.
2. `go get github.com/skip2/go-qrcode`, `make sqlc`.
3. Deploy. Existing memberships read as `UNKNOWN`, which is accurate — their provenance was
   never recorded.
4. Rollback: drop the table, column and enum; the card falls back to no QR with the binary.
