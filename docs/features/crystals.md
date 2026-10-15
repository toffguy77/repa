# Crystals & Billing (T15)


## Free crystal grants

Before this, `crystal_log_type` had a `BONUS` value that **no production code ever wrote** —
every crystal came from a purchase, which meant the detector (the product's central hook) was
never *tried* before it was sold. The audience is 14–22; the younger half has no card.

| Grant | Amount | Trigger | Idempotency key |
|---|---|---|---|
| Welcome | 10 💎 | Successful registration | `welcome:{userID}` |
| Referral | 5 💎 | An invited member **completes their first voting session** | `referral:{inviteeID}` |
| `STREAK_VOTER` milestone | 5 💎 | The achievement is awarded (5/10/20 seasons) | `achievement:{achievementID}` |
| `RECRUITER` | 10 💎 | The achievement is awarded (3+ members joined after this user) | `achievement:{achievementID}` |

### Why these amounts

- The welcome grant is **exactly one detector**, not two: the goal is to create the want, not to
  satisfy it.
- Two referrals buy a detector, so inviting is a real path to the hook for someone with no card.
- `RECRUITER` pays a whole detector because bringing three people in is the behaviour the
  product most wants.
- Only `STREAK_VOTER` and `RECRUITER` pay. Paying for every achievement would make the currency
  meaningless — and would reward `BLIND` (accuracy under 20%), which is a joke badge.

### Why the referral pays on the first completed *vote*

Rewarding the join would pay for an account, and accounts are free. A completed session means
the invitee answered every question about real people in a group that can actually reveal (see
the participation floors in `docs/features/reveal.md`) — not worth faking for 5 crystals.

### Idempotency

Every grant writes a `BONUS` row in `crystal_logs` with a **derived** `external_id`. That column
is already `UNIQUE`, so a replayed job hits the constraint and the grants service treats that
specific failure as success: the grant exists, which is what the caller wanted. A random id would
make every retry a new payment.

`referral:{inviteeID}` is keyed on the *invitee*, which is what makes "once per invited member"
true by construction rather than by a count query that races.

### Failure behaviour

A grant never returns an error to its caller. Every trigger is a path whose primary job —
registering a user, recording a vote, awarding an achievement — must not be rolled back by a
payout. Failures are logged; because grants are idempotent, a missed grant can be re-driven.

Existing users receive no welcome grant: backfilling would mean paying out for sessions that
already happened.


## Overview
Virtual currency system (crystals) with YuKassa payment integration for purchasing crystal packages.

## Balance
- Computed as `SUM(delta)` from `crystal_logs` table — no separate balance field.
- Query: `GetUserBalance` returns `COALESCE(SUM(delta), 0)`.

## Packages
| ID | Crystals | Bonus | Price (RUB) |
|----|----------|-------|-------------|
| starter | 10 | 0 | 59.00 |
| popular | 30 | 5 | 149.00 |
| advanced | 70 | 15 | 299.00 |
| max | 160 | 40 | 599.00 |

Packages are hardcoded in `internal/service/crystals/service.go`.

## Endpoints

### Protected (JWT required)
- `GET /api/v1/crystals/balance` — returns `{ data: { balance } }`
- `GET /api/v1/crystals/packages` — returns `{ data: { packages } }`
- `POST /api/v1/crystals/purchase/init` — body: `{ package_id }`, creates YuKassa payment, returns `{ data: { payment_url, payment_id } }`
- `GET /api/v1/crystals/purchase/verify/:paymentId` — polls payment status, returns `{ data: { status, new_balance? } }`

### Public (no JWT — called by YuKassa)
- `POST /api/v1/crystals/purchase/webhook` — processes YuKassa webhook events

## Purchase Flow
1. Client calls `POST /crystals/purchase/init` with `package_id`
2. Backend creates payment in YuKassa API, stores `paymentId -> {userId, packageId}` in Redis (TTL 1h)
3. Client opens `payment_url` in external browser
4. User completes payment in YuKassa
5. YuKassa sends `payment.succeeded` webhook to `/crystals/purchase/webhook`
6. Backend credits crystals (crystals + bonus) via `crystal_logs` with `external_id = paymentId`
7. Client polls `GET /crystals/purchase/verify/:paymentId` after returning from browser

## Idempotency
- `crystal_logs.external_id` has a UNIQUE constraint — duplicate webhooks are safely ignored.

## Spending
Crystal spending (detector, hidden attributes) is handled in the reveal service:
- Detector: 10 crystals (`internal/service/reveal/service.go` — `BuyDetector`)
- Hidden attributes: 5 crystals (`internal/service/reveal/service.go` — `OpenHidden`)

## Architecture
```
internal/lib/yukassa.go          — YuKassa REST API client
internal/service/crystals/       — business logic (balance, packages, purchase, webhook)
internal/handler/crystals/       — HTTP handlers
```

## Ownership Verification Fallback

When the Redis key (paymentId -> userId mapping) has expired by the time `verify` is called, the service falls back to scanning `crystal_logs` by `external_id` to confirm the payment belongs to the requesting user.

## Mobile (Flutter)

### Flutter Architecture

```
mobile/lib/features/crystals/
├── data/crystals_repository.dart
├── domain/crystals.dart              # Freezed: CrystalPackage, InitPurchaseResult, VerifyResult
└── presentation/
    ├── crystals_notifier.dart        # StateNotifier + two providers
    ├── crystals_shop_screen.dart     # Main shop screen
    └── widgets/
        ├── crystal_balance_widget.dart
        ├── package_card.dart
        ├── payment_pending_sheet.dart
        └── purchase_success_sheet.dart
```

### Screen: CrystalsShopScreen (`/shop`)

4 package cards with gradient balance header. "Popular" package highlighted. Entry point: `CrystalBalanceWidget` in home AppBar taps to `/shop`.

### Payment Flow (Client-side)

1. `initPurchase` opens YuKassa URL via `url_launcher` with `LaunchMode.externalApplication`
2. On return, deeplink (`app_links`) listening for `/payment/return` path auto-triggers polling
3. Fallback: `PaymentPendingSheet` lets the user manually trigger polling
4. Polling: 3-second interval, max 10 attempts
5. On `succeeded` status — `PurchaseSuccessSheet` with animated crystal count, updates `crystalBalanceProvider`
6. On `canceled` — shows error. Network errors during polling are silently retried.

### Riverpod Providers

- `crystalsProvider` (`StateNotifierProvider.autoDispose`) — shop screen lifecycle
- `crystalBalanceProvider` (non-autoDispose) — global balance for `CrystalBalanceWidget` in AppBar

### DetectorSheet Integration

When user's crystal balance is below 10, the "buy detector" button is replaced with a "Купить кристаллы" button navigating to the shop.

## Configuration
Env vars in `backend/.env`:
- `YUKASSA_SHOP_ID` — YuKassa shop identifier
- `YUKASSA_SECRET_KEY` — YuKassa API secret
- `YUKASSA_RETURN_URL` — URL to redirect user after payment

## Crystal history

### `GET /api/v1/crystals/history`
Lists the user's crystal movements, newest first.
- **Query:** `limit` (default 50, clamped to 100), `offset`
- **Success 200:** `{ "data": { "entries": [{ "delta", "type", "reason", "created_at", "is_grant" }] } }`
- `is_grant` separates free crystals from purchases, so a balance that grew without a payment is
  not mysterious.

### Mobile

`CrystalHistoryList` (`lib/features/crystals/presentation/widgets/crystal_history_list.dart`)
renders the list under the packages on the shop screen:

- A grant shows a gift icon tinted with the `energy` role and the subtitle "Бесплатно"; a
  purchase shows a bag icon in the secondary text colour.
- Amounts are signed and use tabular figures, so a changing list does not shift horizontally.
- An entry with no reason falls back to "Подарок" / "Покупка" rather than rendering blank.
- The history load is wrapped so a failure degrades to an empty list — it must not take the shop
  down with it.

## What crystals buy

| Action | Price |
|---|---|
| Detector — voter count | **free** |
| Detector — one partial hint | 3 💎 |
| Detector — full voter list | 10 💎 |
| Open hidden attributes | 5 💎 |

The detector is a ladder rather than a single purchase; the reasoning and the anonymity limits are in
`docs/features/reveal.md` → Detector ladder.
