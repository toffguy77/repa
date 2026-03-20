# T08 — Backend: Voting API

## Цель
API для сессии голосования: получить вопросы, записать голоса, отслеживать прогресс.

## Эндпоинты

### GET /api/v1/seasons/:seasonId/voting-session
```typescript
// Получить вопросы для голосования текущего пользователя
// Проверки:
// 1. Пользователь — участник группы сезона
// 2. Сезон в статусе VOTING
// 3. Пользователь ещё не завершил голосование (нет записей vote для всех вопросов)

// Логика:
// Получить SeasonQuestions с вопросами
// Для каждого вопроса определить уже ли проголосовал пользователь
// Вернуть список участников (targets) для выбора

{
  data: {
    seasonId: string,
    questions: VotingQuestionDto[],
    targets: TargetDto[],       // все участники кроме себя
    progress: { answered: number, total: number }
  }
}

// VotingQuestionDto: { id, questionId, text, category, answered: boolean }
// TargetDto: { userId, username, avatarEmoji, avatarUrl }
```

### POST /api/v1/seasons/:seasonId/votes
```typescript
// Записать голос
// Request: { questionId: string, targetId: string }
// Проверки:
// 1. Участник группы
// 2. Сезон VOTING
// 3. Не голосовал по этому вопросу ранее (unique constraint)
// 4. targetId — участник группы, не сам пользователь

{ data: { vote: { questionId, targetId }, progress: { answered, total } } }
```

### GET /api/v1/seasons/:seasonId/progress
```typescript
// Общий прогресс голосования в группе (без раскрытия кто голосовал)
{
  data: {
    votedCount: number,          // сколько участников завершили голосование
    totalCount: number,          // всего участников
    quorumReached: boolean,
    quorumThreshold: number,     // 50% или 40% для малых групп
    userVoted: boolean           // текущий пользователь уже голосовал?
  }
}
```

## Бизнес-правила
- Нельзя голосовать за себя (`targetId !== voterId`)
- Нельзя изменить голос после подтверждения
- Quorum: ≥ 50% для групп ≥ 8 человек, ≥ 40% для < 8 человек
- Если пользователь прервал сессию — сохранённые голоса остаются, можно продолжить

## Структура файлов
```
src/modules/voting/
├── voting.router.ts
├── voting.service.ts
├── voting.schema.ts
└── voting.test.ts
```

## Тесты
- Получить сессию → ответить на все вопросы → прогресс 100%
- Повторный голос по тому же вопросу → 409
- Голос за себя → 400
- Участник не группы → 403
- Сезон не в VOTING → 400

## Критерии готовности
- [ ] Все 3 эндпоинта работают
- [ ] Анонимность: `/progress` не раскрывает кто именно голосовал
- [ ] Уникальность голоса соблюдается
- [ ] Тесты проходят
