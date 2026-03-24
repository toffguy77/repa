# Crystals & Billing (T15)

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
