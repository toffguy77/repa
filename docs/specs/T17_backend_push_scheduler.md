# T17 — Backend: Push-уведомления и недельный scheduler

## Цель
FCM интеграция и полное расписание пушей retention-системы.

## src/lib/firebase.ts
```typescript
import { initializeApp, cert } from 'firebase-admin/app'
import { getMessaging } from 'firebase-admin/messaging'

export const messaging = getMessaging(initializeApp({ credential: cert({...}) }))

export async function sendPush(
  tokens: string[],
  title: string,
  body: string,
  data?: Record<string, string>
): Promise<void> {
  // sendEachForMulticast
  // Удалять невалидные токены из БД (registration-token-not-registered)
}

export async function sendPushToUser(
  userId: string,
  title: string,
  body: string,
  data?: Record<string, string>
): Promise<void> {
  // Получить все FcmToken пользователя → sendPush
}
```

## Эндпоинт регистрации токена

### POST /api/v1/push/register
```typescript
// Request: { token: string, platform: 'ios' | 'android' }
// Upsert FcmToken (token unique)
{ data: { registered: true } }
```

## Rate limiting пушей
```typescript
// Redis key: `push-count:{userId}:{date}` (date = YYYY-MM-DD МСК)
// TTL: до конца дня
async function canSendPush(userId: string): Promise<boolean> {
  // Не отправлять если:
  // 1. Счётчик за день ≥ 3
  // 2. Текущее время МСК в диапазоне 23:00–09:00
}
```

## BullMQ Jobs — недельное расписание

### Job: `weekly-scheduler` (cron: каждый понедельник 14:00 UTC = 17:00 МСК)
```typescript
// Для всех активных сезонов в статусе VOTING:
// 1. Отправить push «Новый сезон начался» всем участникам
// 2. Поставить в очередь остальные weekly jobs
```

### Job: `tuesday-signal` (cron: каждый вторник 16:00 UTC = 19:00 МСК)
```typescript
// Для участников которые уже проголосовали:
// Пуш: «Кто-то уже ответил на вопросы про тебя 👀»
// data: { screen: 'reveal-waiting', groupId }
```

### Job: `wednesday-quorum` (cron: каждую среду 15:00 UTC)
```typescript
// Для каждого активного сезона:
// Если quorum ≥ 80%: пуш не проголосовавшим — «Все уже ответили, остался только ты»
// Если quorum < 50%: пуш всем — «Reveal под угрозой, не хватает N голосов»
```

### Job: `thursday-teaser` (cron: каждый четверг 17:00 UTC = 20:00 МСК)
```typescript
// Для проголосовавших участников:
// Определить ведущую категорию вопросов по текущим голосам
// Пуш: «Один твой атрибут уже почти определился… [эмодзи категории]»
// data: { screen: 'reveal-waiting', groupId }
```

### Job: `friday-pre-reveal` (cron: каждую пятницу 16:00 UTC)
```typescript
// За 1 час до Reveal:
// Пуш всем: «Через час — репа. Готов? 🍆»
```

### Job: `reveal-notification` (запускается из reveal-process)
```typescript
// После успешного Reveal:
// Пуш всем участникам: «Твоя репа готова 🍆»
// data: { screen: 'reveal', groupId, seasonId }
```

### Job: `sunday-preview` (cron: каждое воскресенье 09:00 UTC = 12:00 МСК)
```typescript
// Для всех групп:
// Пуш: «Голосуй за вопросы следующей недели»
// data: { screen: 'question-vote', groupId }
```

### Job: `sunday-streak-reminder` (cron: каждое воскресенье 15:00 UTC = 18:00 МСК)
```typescript
// Участникам со streak ≥ 3:
// Пуш: «Не прерывай серию — скоро новый сезон 🔥»
```

## Эндпоинты выбора вопросов следующего сезона

### GET /api/v1/groups/:groupId/next-season/question-candidates
```typescript
// 3 случайных вопроса для голосования
{ data: { candidates: QuestionDto[] } }
```

### POST /api/v1/groups/:groupId/next-season/vote-question
```typescript
// Request: { questionId: string }
// Одна группа — один голос за вопрос
// Вопрос с наибольшим числом голосов добавляется в следующий сезон
{ data: { voted: true } }
```

## Структура файлов
```
src/jobs/
├── weekly-scheduler.job.ts
├── tuesday-signal.job.ts
├── wednesday-quorum.job.ts
├── thursday-teaser.job.ts
├── friday-pre-reveal.job.ts
├── reveal-notification.job.ts
├── sunday-preview.job.ts
└── sunday-streak.job.ts
src/modules/push/
├── push.router.ts
└── push.service.ts
```

## Критерии готовности
- [ ] FCM отправляет пуш на тестовый токен
- [ ] Rate limit: не более 3 пушей в день
- [ ] Тихие часы 23:00–09:00 МСК соблюдаются
- [ ] Все cron jobs зарегистрированы в BullMQ
- [ ] Невалидные токены удаляются из БД
