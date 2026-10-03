# Proposal

## Why

A group's invite is a raw UUID: `CreateGroup` returns
`"https://repa.app/join/" + inviteCode` where `inviteCode` is `uuid.New().String()`
(`backend/internal/service/groups/service.go`), and `JoinGroupScreen`
(`mobile/lib/features/groups/presentation/join_group_screen.dart`) offers a text field for
the user to *type* that code.

A 36-character UUID cannot be said out loud, written on a whiteboard, remembered between
two classrooms, or typed without errors. The target audience sits in the same physical room
as the people it needs to invite, so the offline path — "код AB12CD, вбей" — is a real
distribution channel, and today it does not exist. It also makes the code unusable on the
shared reputation card, which is the next change's whole premise.

## What Changes

- Invite codes become short, human-transcribable codes: 6 characters from an alphabet that
  removes the characters people confuse (`0/O`, `1/I/L`), case-insensitive on input.
- **BREAKING** (external): `invite_url` moves from `https://repa.app/join/{uuid}` to a short
  form. Existing UUID codes keep working — they are resolved by the same lookup — so links
  already shared do not break.
- Joining accepts a code typed in any case, with or without surrounding whitespace, and also
  accepts a full pasted link. Entry is normalised before lookup.
- A code is guaranteed unique per group at generation time, and regenerating an invite
  issues a new short code and invalidates the old one.
- The app presents the code as something to read aloud — grouped, selectable, copyable —
  rather than as an opaque string.

## Capabilities

### New Capabilities

- `group-invites`: how a group is joined — the shape of an invite code, what the system
  accepts as input, and the guarantees around uniqueness and regeneration.

### Modified Capabilities

None. `season-lifecycle` and `small-group-anonymity` are unaffected; this change alters how
members arrive, not what happens once they have.

## Impact

- **Schema:** `groups.invite_code` gains a length/charset constraint and a case-insensitive
  unique index; existing UUID values are preserved.
- **Backend:** `internal/service/groups/service.go` (code generation, normalisation, join and
  regenerate paths), `internal/db/queries/groups.sql` (case-insensitive lookup).
- **API:** `POST /groups` and `POST /groups/:id/invite-link` return the short URL;
  `POST /groups/join/:inviteCode` and the preview endpoint accept normalised input.
- **Mobile:** `JoinGroupScreen` input normalisation and hint; `GroupScreen` and
  `CreateGroupScreen` invite-sharing presentation.
- **Docs:** `docs/features/groups.md`.

## Non-goals

- Vanity or user-chosen codes (`repa.app/join/9b-matematika`) — collision handling,
  squatting and moderation make that its own change.
- Expiring or single-use invites.
- QR codes and the shared card's call to action — the next change (`card-share-attribution`)
  owns those; this one only makes a code that can fit on a card.
- Deep-link infrastructure changes: the existing `/join/:code` route and
  `pending_invite_code` handoff keep working unchanged.
