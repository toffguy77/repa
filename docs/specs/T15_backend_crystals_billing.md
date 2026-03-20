# T15 — Backend: Кристаллы и ЮKassa

## Цель
Система виртуальной валюты (кристаллы) и интеграция с ЮKassa для пополнения.

## Логика баланса

Баланс = сумма всех `CrystalLog.delta` для пользователя.
Никогда не хранить баланс как отдельное поле — только через лог.

```typescript
async function getBalance(userId: string): Promise<number> {
  const result = await prisma.crystalLog.aggregate({
    where: { userId },
    _sum: { delta: true }
  })
  return result._sum.delta ?? 0
}

async function spendCrystals(
  userId: string,
  amount: number,
  type: CrystalLogType,
  description: string
): Promise<number> {
  // Атомарная транзакция:
  // 1. Проверить баланс
  // 2. Если недостаточно → throw AppError('INSUFFICIENT_CRYSTALS')
  // 3. Записать CrystalLog с delta = -amount
  // 4. Вернуть новый баланс
}
```

## Пакеты кристаллов

```typescript
export const CRYSTAL_PACKAGES = [
  { id: 'starter',  crystals: 10,  bonus: 0,  priceKopecks: 5900  },
  { id: 'popular',  crystals: 30,  bonus: 5,  priceKopecks: 14900 },
  { id: 'advanced', crystals: 70,  bonus: 15, priceKopecks: 29900 },
  { id: 'max',      crystals: 160, bonus: 40, priceKopecks: 59900 },
] as const
```

## Эндпоинты

### GET /api/v1/crystals/balance
```typescript
{ data: { balance: number } }
```

### GET /api/v1/crystals/packages
```typescript
{ data: { packages: CrystalPackageDto[] } }
```

### POST /api/v1/crystals/purchase/init
```typescript
// Request: { packageId: string }
// Логика:
// 1. Найти пакет
// 2. Создать платёж в ЮKassa
// 3. Сохранить payment_id в Redis (TTL 1 час) → связь с userId и packageId
// 4. Вернуть URL для редиректа

{ data: { paymentUrl: string, paymentId: string } }

// ЮKassa API:
// POST https://api.yookassa.ru/v3/payments
// Auth: Basic {shopId}:{secretKey}
// Body: { amount, currency: 'RUB', confirmation: { type: 'redirect', return_url }, description }
```

### POST /api/v1/crystals/purchase/webhook
```typescript
// ЮKassa Webhook (не требует JWT)
// Проверить подпись через IP whitelist ЮKassa или HMAC
// При event = payment.succeeded:
//   1. Найти userId по payment_id из Redis
//   2. Зачислить кристаллы (crystals + bonus)
//   3. Записать CrystalLog (type: PURCHASE, externalId: paymentId)
//   4. Вернуть 200

{ } // пустой ответ, статус 200
```

### GET /api/v1/crystals/purchase/verify/:paymentId
```typescript
// Проверить статус платежа (polling с клиента после возврата из браузера)
{ data: { status: 'pending' | 'succeeded' | 'canceled', newBalance?: number } }
```

### POST /api/v1/seasons/:seasonId/detector
```typescript
// Купить детектор за 10 кристаллов
// Атомарно: spendCrystals + createDetector
// Вернуть список userId проголосовавших за текущего пользователя

{
  data: {
    voters: { userId, username, avatarEmoji }[],
    crystalBalance: number
  }
}
```

## Структура файлов
```
src/modules/crystals/
├── crystals.router.ts
├── crystals.service.ts
├── crystals.schema.ts
└── crystals.test.ts
src/lib/yukassa.ts   # ЮKassa API клиент
```

## Тесты
- Зачислить кристаллы → getBalance → корректный баланс
- spendCrystals при нехватке → ошибка
- Webhook succeeded → зачисление
- Двойной webhook → идемпотентность (externalId unique)

## Критерии готовности
- [ ] Баланс считается через агрегацию лога
- [ ] Платёж создаётся в ЮKassa (или заглушка в dev)
- [ ] Webhook зачисляет кристаллы
- [ ] Детектор атомарно списывает и возвращает данные
- [ ] Анонимность: детектор возвращает только voter list, не привязку к ответам
