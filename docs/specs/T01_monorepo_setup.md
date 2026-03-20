# T01 — Монорепозиторий, Docker, конфиги

## Цель
Создать рабочую основу проекта: монорепо со структурой папок, Docker Compose для локальной разработки, TypeScript-конфиги для backend.

## Что реализовать

### Структура монорепо
```
repa/
├── backend/
│   ├── src/
│   │   ├── modules/
│   │   ├── lib/
│   │   ├── jobs/
│   │   ├── plugins/
│   │   └── app.ts
│   ├── prisma/
│   │   └── schema.prisma      # пустая схема, будет заполнена в T02
│   ├── .env.example           # все переменные из master context
│   ├── package.json
│   └── tsconfig.json
├── mobile/
│   └── .gitkeep               # Flutter проект создаётся в T05
├── docker-compose.yml
├── .gitignore
└── README.md
```

### docker-compose.yml
Сервисы:
- `postgres`: образ `postgres:16-alpine`, порт 5432, volume для данных, healthcheck
- `redis`: образ `redis:7-alpine`, порт 6379, healthcheck

### backend/package.json
Зависимости:
```json
{
  "dependencies": {
    "fastify": "^4",
    "@fastify/jwt": "^8",
    "@fastify/cors": "^9",
    "@fastify/helmet": "^11",
    "@fastify/rate-limit": "^9",
    "@fastify/multipart": "^8",
    "@prisma/client": "^5",
    "bullmq": "^5",
    "ioredis": "^5",
    "zod": "^3",
    "firebase-admin": "^12",
    "@aws-sdk/client-s3": "^3",
    "@anthropic-ai/sdk": "^0.20",
    "node-telegram-bot-api": "^0.64",
    "axios": "^1",
    "date-fns": "^3",
    "date-fns-tz": "^3"
  },
  "devDependencies": {
    "prisma": "^5",
    "typescript": "^5",
    "tsx": "^4",
    "vitest": "^1",
    "@types/node": "^20"
  },
  "scripts": {
    "dev": "tsx watch src/server.ts",
    "build": "tsc",
    "test": "vitest",
    "db:migrate": "prisma migrate dev",
    "db:seed": "tsx prisma/seed.ts",
    "db:studio": "prisma studio"
  }
}
```

### backend/tsconfig.json
- `strict: true`
- `target: ES2022`
- `module: NodeNext`
- `moduleResolution: NodeNext`
- `outDir: dist`
- `rootDir: src`

### backend/src/server.ts
- Создаёт Fastify инстанс
- Загружает `.env` через `process.env`
- Слушает на `PORT` из env
- Логирует старт

### backend/src/app.ts
- Экспортирует функцию `buildApp()` которая:
  - Регистрирует `@fastify/helmet`, `@fastify/cors`, `@fastify/jwt`
  - Добавляет глобальный error handler в формате `{ error: { code, message } }`
  - Регистрирует health check route `GET /health → { status: "ok" }`
  - Возвращает Fastify инстанс

### backend/src/lib/prisma.ts
Singleton PrismaClient:
```typescript
import { PrismaClient } from '@prisma/client'
export const prisma = new PrismaClient()
```

### backend/src/lib/redis.ts
Singleton ioredis клиент с обработкой `connect` и `error` событий.

### .gitignore
node_modules, dist, .env, *.log, .DS_Store, build/, android/, ios/ (для flutter)

### README.md
Инструкция: `docker compose up -d` → `npm install` → `npm run db:migrate` → `npm run dev`

## Критерии готовности
- [ ] `docker compose up -d` поднимает postgres и redis без ошибок
- [ ] `npm run dev` стартует сервер, нет TypeScript ошибок
- [ ] `GET /health` возвращает 200 `{ status: "ok" }`
- [ ] Prisma может подключиться к БД
- [ ] Redis клиент подключается без ошибок

## Не делать
- Не реализовывать бизнес-логику
- Не создавать Flutter проект (T05)
- Не заполнять Prisma схему (T02)
