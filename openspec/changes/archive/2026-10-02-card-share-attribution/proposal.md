# Proposal

## Why

The shared reputation card is the product's only organic acquisition channel, and it contains
no way back into the product. `BuildCardHTML`
(`backend/internal/service/cards/template.go`) renders logo, avatar, username, title, top-3
attributes and a `"{group} · Сезон {n}"` footer — no link, no code, no QR, no call to action.
The share text is a hardcoded constant: `'Моя репа repa.app'`
(`mobile/lib/features/reveal/presentation/reveal_screen.dart`), with no deep link and no
invite code.

So a card seen by 200 people in Stories produces zero attributable installs and zero joins,
and there is no measurement to tell anyone that. The previous change made a code that fits on
a card; this change puts it there and closes the loop.

## What Changes

- The card gains a call-to-action block: the group's short invite code, a QR code encoding the
  invite link, and one line telling a viewer what to do with it.
- **BREAKING** (behavioural): sharing a card now shares a personal invite link carrying the
  group's code and a channel marker, instead of a bare `repa.app` string.
- Joining records **how** the member arrived — link, typed code, card, Telegram, or unknown —
  so the funnel from share to join is measurable rather than assumed.
- Share actions are recorded per season and channel, so "cards shared" and "joins from cards"
  can be compared.
- Admin stats expose both counts.

## Capabilities

### New Capabilities

- `share-attribution`: what a shared artifact carries, and what the system records about how a
  member arrived.

### Modified Capabilities

- `group-invites`: gains a requirement that an invite can be delivered by scanning as well as
  by link and typed code.

## Impact

- **Schema:** `group_members.join_source`, a `share_events` table, and a `join_source` enum.
- **Backend:** `internal/service/cards/` (QR generation and the CTA block),
  `internal/service/groups/` (join source), new share-event recording, admin stats.
- **Dependency:** `github.com/skip2/go-qrcode` for QR rendering.
- **API:** `POST /groups/join/:code` accepts an optional source; a new endpoint records a share;
  admin stats gain the two counts.
- **Mobile:** reveal share flow builds the link and reports the channel; the deep-link handler
  records the source it arrived through.
- **Docs:** `docs/features/cards.md`, `docs/features/groups.md`, a new section in
  `docs/features/moderation.md`'s admin area.

## Non-goals

- Deferred deep linking (install → first-launch attribution through a third-party service like
  Branch). That needs an external SDK and a privacy review; this change makes the link and the
  measurement, not the install-time matching.
- Per-user referral rewards — `crystal-free-economy` owns that, and it depends on this
  change's attribution data.
- Redesigning the card's layout or palette beyond adding the CTA block.
- Server-side rendering of a user's Stories post, or posting on their behalf.
