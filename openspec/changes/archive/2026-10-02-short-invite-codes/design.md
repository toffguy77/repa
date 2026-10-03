# Design

## Context

See proposal.md — Why. Today `CreateGroup` sets `InviteCode: uuid.New().String()` and
`GetGroupByInviteCode` does an exact-match lookup; the mobile join screen debounces a preview
call against whatever the user typed. The deep link route `/join/:code` and the
`pending_invite_code` handoff through `flutter_secure_storage` already work and are not part
of this change.

## Goals / Non-Goals

**Goals:**

- A code a 15-year-old can read across a classroom and another can type without errors.
- Normalisation in exactly one place, so the API, the deep link, and the paste path cannot
  disagree about what a valid code is.
- Not breaking links already in people's chats.

**Non-Goals:**

- Changing the invite *URL host* or the deep-link mechanism.
- Code expiry, usage limits, or per-invite attribution (the next change adds attribution at
  the share layer, not the code layer).

## Decisions

### Alphabet and length: 6 characters from a 31-symbol set

Alphabet: `23456789ABCDEFGHJKMNPQRSTUVWXYZ` — digits `0`/`1` and letters `O`/`I`/`L` removed,
because those are the pairs people actually confuse when reading a code aloud or off a
screen. That leaves 31 symbols, so 6 characters give 31⁶ ≈ 887 million codes.

Length 6 is the trade-off point: 5 characters (28.6M) starts to feel guessable for a product
whose groups are private-by-obscurity, and 7 is one more thing to hold in working memory
while walking to the next desk. At 6, a brute-force attempt needs ~10⁸ requests to find one
real group, which the existing rate limiting makes uninteresting; the invite is a convenience
boundary, not a security boundary, and the group-size cap plus the admin's ability to
regenerate are the real controls.

Codes are generated with `crypto/rand`, not `math/rand`: a predictable sequence would let
someone enumerate newly created groups, which is exactly the property the alphabet choice is
protecting.

### Normalisation is one function, used by every entry point

`NormalizeInviteCode(raw string) string` upper-cases, trims, strips a leading invite URL
prefix, and drops separator characters a user might add (spaces, hyphens). It is applied in
the service, so the handler, the deep link, and the preview path cannot disagree.

Alternatives considered: normalising in the handler (rejected — the deep-link path and any
future bot command bypass the handler) and making the *client* the authority on validity
(rejected — then the server's idea of a valid code depends on the client version).

The client still does one narrow thing: `extractInviteCode` reduces a pasted link to the bare
code. That is not a second copy of the validity rule — the code travels as a URL **path
segment**, so a pasted link's slashes break routing before the request reaches the service.
Case and separators remain the server's business.

Separator stripping is conditional on the result still being short-code sized. Stripping
unconditionally mangles a legacy UUID, whose hyphens *are* part of the stored value — which
would break exactly the links the compatibility requirement exists to protect.

### Case-insensitive uniqueness via a functional unique index

A `CREATE UNIQUE INDEX ... ON groups (upper(invite_code))` enforces the guarantee in the
database rather than in application logic, so a race between two simultaneous group creations
cannot produce two groups sharing a code. The generator retries on conflict, and after a
bounded number of attempts creation fails loudly rather than silently issuing a duplicate —
with 887M codes, exhausting the retries means something is wrong, not that the space is full.

The existing exact-match `GetGroupByInviteCode` query is replaced by an `upper()`-based
lookup so it can use that same index.

### Legacy UUID codes keep working by doing nothing special

Because the lookup is `upper(invite_code) = upper($1)` and old codes are stored as-is,
existing UUIDs resolve without a migration or a compatibility branch. The length/charset rule
applies only to *generation*; no constraint is added that would reject the rows already
there. This is deliberately asymmetric — validating stored codes would mean rewriting them,
and rewriting them would break the links this requirement exists to protect.

### Presentation: grouped as `AB2 CD3`

The app shows the code split into two triples with a wide gap. This is the one place the
format is cosmetic rather than semantic — the stored and transmitted code has no space in it,
and normalisation strips any the user types back in. Splitting it is what makes a 6-character
string read as two chunks instead of one blur.

## Risks / Trade-offs

- **A shorter code is easier to guess than a UUID** → accepted: a found group still caps at
  50 members, an admin can regenerate, and the join endpoint is rate-limited. Treating the
  invite as a security boundary would require invite approval, which is a different product.
- **Two code formats coexist indefinitely** → the lookup handles both with no branch, and
  regeneration migrates a group opportunistically. No scheduled backfill, because rewriting a
  code is exactly the breakage this avoids.
- **Stripping hyphens makes `AB2-CD3` valid input** → intended; a user who sees the grouped
  display and types what they see should succeed.

## Migration Plan

1. Migration `005_invite_code_short`: add the `upper(invite_code)` unique index, drop the
   plain unique constraint it supersedes.
2. `make sqlc`.
3. Deploy. New groups get short codes; existing groups keep theirs until regenerated.
4. Rollback: restore the plain unique index; the code path tolerates both formats either way.
