# T22 — Backend: Rate limiting и security hardening

## Цель
Защита API: rate limits, валидация анонимности, input sanitization, security headers.

## Что реализовать

### Rate limits по эндпоинтам
```typescript
// Глобально: 100 req/min (уже в T03)
// Дополнительно:

// Auth OTP: 3 запроса на телефон в час
// POST /auth/otp/send → ключ: `rl:otp:{phone}`, limit: 3, window: 3600s

// Voting: 30 голосов в минуту (защита от скриптов)
// POST /seasons/:id/votes → ключ: `rl:vote:{userId}`, limit: 30, window: 60s

// Questions: 10 вопросов в час
// POST /groups/:id/questions → ключ: `rl:question:{userId}`, limit: 10, window: 3600s

// Card generation: 5 запросов в час
// GET /seasons/:id/my-card-url → ключ: `rl:card:{userId}`, limit: 5, window: 3600s
```

### Проверка анонимности (тест)
```typescript
// src/tests/anonymity.test.ts
// Критический тест: убедиться что ни один API endpoint
// не возвращает связку voterId + targetId + questionId

// Сценарий:
// 1. Создать группу из 3 участников
// 2. Проголосовать
// 3. Пройтись по всем GET эндпоинтам, проверить ответы
// 4. Убедиться что voterId нигде не возвращается в контексте конкретного голоса
```

### Input sanitization
```typescript
// Middleware для всех текстовых полей:
// - Trim whitespace
// - Ограничить длину (уже в Zod схемах)
// - Запретить HTML теги (strip)
// - Запретить null bytes

// Добавить в Zod схемы:
const sanitizedString = z.string().trim().transform(s => s.replace(/<[^>]*>/g, ''))
```

### Security headers (уже @fastify/helmet, усилить)
```typescript
// В app.ts:
fastify.register(helmet, {
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'"],
      // ...
    }
  },
  crossOriginEmbedderPolicy: false  // для медиа файлов
})
```

### Telegram webhook защита
```typescript
// Проверять заголовок X-Telegram-Bot-Api-Secret-Token
// Секрет устанавливается при регистрации webhook
fastify.addHook('preHandler', async (req, reply) => {
  if (req.url.startsWith('/api/v1/telegram/webhook')) {
    const secret = req.headers['x-telegram-bot-api-secret-token']
    if (secret !== process.env.TELEGRAM_WEBHOOK_SECRET) {
      return reply.status(401).send({ error: { code: 'UNAUTHORIZED', message: '' }})
    }
  }
})
```

### ЮKassa webhook защита
```typescript
// Проверять IP: только диапазоны ЮKassa
// https://yookassa.ru/developers/using-api/webhooks
const YUKASSA_IPS = ['185.71.76.0/27', '185.71.77.0/27', '77.75.153.0/25', ...]

fastify.addHook('preHandler', async (req, reply) => {
  if (req.url === '/api/v1/crystals/purchase/webhook') {
    const ip = req.ip
    if (!isInAllowedRange(ip, YUKASSA_IPS)) {
      return reply.status(403).send(...)
    }
  }
})
```

### Тест: age check
```typescript
// Пользователь с birthYear < 2007 (< 18 лет):
// POST /groups с category: ROMANCE → должен вернуть 400
```

## Критерии готовности
- [ ] Rate limits работают для OTP (3/час на телефон)
- [ ] Тест анонимности проходит
- [ ] HTML теги strip из текстовых полей
- [ ] Telegram webhook проверяет secret token
- [ ] ЮKassa webhook проверяет IP
- [ ] Несовершеннолетние не могут добавить ROMANCE категорию
