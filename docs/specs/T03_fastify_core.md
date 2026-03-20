# T03 — Fastify: плагины, middleware, базовая структура

## Цель
Настроить production-ready Fastify приложение: аутентификация через JWT, плагины, error handling, структура модулей.

## Что реализовать

### src/plugins/authenticate.ts
Fastify decorator `authenticate` — preHandler для защищённых роутов:
```typescript
fastify.decorate('authenticate', async (request, reply) => {
  await request.jwtVerify()
  // Добавить request.user = { id, username }
})
```

### src/plugins/prisma.ts
Fastify plugin: декорирует инстанс `fastify.prisma`, закрывает соединение при shutdown.

### src/plugins/redis.ts
Fastify plugin: декорирует `fastify.redis`, закрывает при shutdown.

### src/lib/bullmq.ts
```typescript
import { Queue, Worker } from 'bullmq'
import { redis } from './redis'

export const queues = {
  reveal: new Queue('reveal', { connection: redis }),
  push:   new Queue('push',   { connection: redis }),
  telegram: new Queue('telegram', { connection: redis }),
}
```

### src/app.ts (обновить из T01)
Зарегистрировать:
- `@fastify/helmet` с CSP
- `@fastify/cors` (origins из env)
- `@fastify/jwt` (secret из env, алгоритм HS256)
- `@fastify/rate-limit` (глобально 100 req/min)
- Плагин `prisma`
- Плагин `redis`
- Декоратор `authenticate`
- Все роутеры под префиксом `/api/v1`

### src/lib/errors.ts
```typescript
export class AppError extends Error {
  constructor(
    public code: string,
    public message: string,
    public statusCode: number = 400
  ) { super(message) }
}

export const Errors = {
  UNAUTHORIZED: () => new AppError('UNAUTHORIZED', 'Не авторизован', 401),
  FORBIDDEN:    () => new AppError('FORBIDDEN', 'Доступ запрещён', 403),
  NOT_FOUND:    (entity: string) => new AppError('NOT_FOUND', `${entity} не найден`, 404),
  CONFLICT:     (msg: string) => new AppError('CONFLICT', msg, 409),
  VALIDATION:   (msg: string) => new AppError('VALIDATION', msg, 400),
}
```

Global error handler в app.ts:
```typescript
fastify.setErrorHandler((error, request, reply) => {
  if (error instanceof AppError) {
    return reply.status(error.statusCode).send({
      error: { code: error.code, message: error.message }
    })
  }
  // ZodError → 400
  // Prisma unique → 409
  // Остальное → 500
})
```

### src/lib/jwt.ts
Хелперы:
```typescript
export const signToken = (payload: JwtPayload): string
export const verifyToken = (token: string): JwtPayload
// JwtPayload = { userId: string, username: string }
```

### src/modules/health/health.router.ts
```
GET /health → { status: "ok", db: "ok", redis: "ok" }
```
Проверяет подключение к БД (`$queryRaw SELECT 1`) и Redis (`ping`).

### Структура типового модуля (пример-шаблон)
```
src/modules/example/
├── example.router.ts   # Fastify routes, только routing + validation
├── example.service.ts  # Бизнес-логика, работа с БД
├── example.schema.ts   # Zod схемы для request/response
└── example.test.ts     # Vitest unit тесты сервиса
```

### vitest.config.ts
```typescript
export default {
  test: {
    environment: 'node',
    setupFiles: ['./src/test/setup.ts'],
  }
}
```

`src/test/setup.ts` — очистка тестовой БД перед каждым тестом (через `prisma.$transaction`).

## Критерии готовности
- [ ] `GET /api/v1/health` возвращает статус всех зависимостей
- [ ] Защищённый роут без токена возвращает 401 в формате `{ error: { code, message } }`
- [ ] Невалидный запрос возвращает 400 с описанием
- [ ] Rate limit работает
- [ ] `npm test` запускает тесты без ошибок конфигурации

## Не делать
- Не реализовывать бизнес-модули (auth, groups и т.д.) — только инфраструктуру
