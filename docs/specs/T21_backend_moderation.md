# T21 — Backend: AI-модерация пользовательских вопросов

## Цель
Модерация пользовательских вопросов через Anthropic API перед добавлением в банк группы.

## src/lib/anthropic.ts
```typescript
import Anthropic from '@anthropic-ai/sdk'
export const anthropic = new Anthropic({ apiKey: process.env.ANTHROPIC_API_KEY })
```

## src/modules/questions/moderation.service.ts

```typescript
const MODERATION_SYSTEM = `Ты модератор вопросов для мобильного приложения «Репа».
Приложение — социальная игра для школьников и студентов 14-22 лет.
Правила допустимых вопросов:
- Вопрос должен быть юмористическим или наблюдательным
- Нельзя: оскорбления, мат, расизм, сексизм, буллинг конкретных людей
- Нельзя: сексуальный контент (даже намёки для аудитории до 18)
- Нельзя: призывы к насилию или самоповреждению
- Можно: безобидный юмор, наблюдения о поведении, лёгкая провокация без агрессии
Отвечай ТОЛЬКО валидным JSON: { "approved": boolean, "reason": string | null }
Reason — короткое объяснение только при отклонении (на русском).`

export async function moderateQuestion(text: string): Promise<ModerationResult> {
  const response = await anthropic.messages.create({
    model: 'claude-haiku-4-5-20251001',   // быстрая и дешёвая модель
    max_tokens: 100,
    system: MODERATION_SYSTEM,
    messages: [{ role: 'user', content: `Вопрос: «${text}»` }]
  })
  
  const content = response.content[0].type === 'text' ? response.content[0].text : ''
  return JSON.parse(content) as ModerationResult
  // { approved: boolean, reason: string | null }
}
```

## Эндпоинты пользовательских вопросов

### POST /api/v1/groups/:groupId/questions
```typescript
// Добавить свой вопрос в банк группы
// Request: { text: string, category: QuestionCategory }
// Валидация: text 10-120 символов

// 1. Проверить лимит: не более 5 кастомных вопросов на пользователя на группу
// 2. Запустить AI-модерацию (synchronously, timeout 5 сек)
// 3. Если approved: создать Question (status: ACTIVE, source: USER, groupId)
// 4. Если rejected: создать с status: REJECTED, вернуть reason

{
  data: {
    question: QuestionDto,
    moderation: { approved: boolean, reason: string | null }
  }
}
```

### GET /api/v1/groups/:groupId/questions
```typescript
// Список всех вопросов группы (системные + пользовательские ACTIVE)
{ data: { questions: QuestionDto[] } }
```

### DELETE /api/v1/groups/:groupId/questions/:questionId
```typescript
// Только admin или автор
// Мягкое удаление (status: REJECTED)
{ data: { deleted: true } }
```

### POST /api/v1/groups/:groupId/questions/:questionId/report
```typescript
// Пожаловаться на вопрос
// Создать запись в отдельной таблице Report для ручной модерации
{ data: { reported: true } }
```

## Модель Report
```prisma
model Report {
  id         String   @id @default(cuid())
  questionId String
  reporterId String
  reason     String?
  createdAt  DateTime @default(now())

  @@unique([questionId, reporterId])
}
```

## Критерии готовности
- [ ] Вопрос проходит AI-модерацию за < 5 сек
- [ ] Оскорбительные вопросы отклоняются с объяснением
- [ ] Нейтральные вопросы проходят
- [ ] Timeout 5 сек → fallback: статус PENDING, ручная проверка
- [ ] Лимит 5 вопросов на пользователя на группу
