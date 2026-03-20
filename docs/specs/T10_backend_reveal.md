# T10 — Backend: Reveal Engine

## Цель
Scheduled job для Reveal: проверка кворума, агрегация результатов, запуск ачивок и пушей.

## BullMQ Jobs

### Job: `reveal-checker` (запускается каждую минуту)
```typescript
// Найти все сезоны где:
// status = VOTING AND revealAt <= now()
// Для каждого такого сезона запустить reveal-process job
```

### Job: `reveal-process` (параметр: seasonId)
```typescript
async function processReveal(seasonId: string) {
  const season = await prisma.season.findUnique({ include: { group: { include: { members: true }}}})

  // 1. Проверить кворум
  const totalMembers = season.group.members.length
  const threshold = totalMembers < 8 ? 0.4 : 0.5
  const uniqueVoters = await prisma.vote.groupBy({
    by: ['voterId'], where: { seasonId }
  })
  if (uniqueVoters.length / totalMembers < threshold) {
    // Кворум не набран — отложить reveal на 2 часа, max 3 попытки
    // Если 3 попытки исчерпаны — reveal принудительно
    return
  }

  // 2. Агрегировать результаты
  // Для каждой пары (questionId, targetId) посчитать voteCount
  // totalVoters = uniqueVoters.length
  // percentage = voteCount / totalVoters * 100

  // 3. Записать SeasonResult (upsert)

  // 4. Обновить статус сезона → REVEALED

  // 5. Запустить ачивки (T11)
  await queues.achievements.add('calculate', { seasonId })

  // 6. Запустить push-уведомления (T17)
  await queues.push.add('reveal-notification', { seasonId })

  // 7. Запустить Telegram пост (T19, если есть telegramChatId)
  if (season.group.telegramChatId) {
    await queues.telegram.add('reveal-post', { seasonId })
  }
}
```

## Эндпоинты Reveal

### GET /api/v1/seasons/:seasonId/reveal
```typescript
// Получить карточку репутации текущего пользователя
// Проверки: участник группы, сезон REVEALED

{
  data: {
    myCard: {
      topAttributes: AttributeDto[],      // топ-3, всегда открыты
      hiddenAttributes: AttributeDto[],   // остальные, заблюрены
      reputationTitle: string,            // автогенерированный титул
      trend: TrendDto,                    // сравнение с прошлым сезоном
      newAchievements: AchievementDto[],  // новые ачивки этого сезона
      cardImageUrl: string,               // URL PNG карточки (S3)
    },
    groupSummary: {
      topAttributesPerQuestion: TopAttributeDto[],  // топ по каждому вопросу
      voterCount: number,
    }
  }
}

// AttributeDto: { questionId, questionText, percentage, rank }
// TrendDto: { attribute: string, change: 'up' | 'down' | 'same', delta: number }
```

### GET /api/v1/seasons/:seasonId/members-cards
```typescript
// Карточки всех участников группы (только после Reveal)
// Показывает только топ-3 атрибута каждого участника

{
  data: {
    members: MemberCardDto[]
  }
}
// MemberCardDto: { userId, username, avatarEmoji, topAttributes: AttributeDto[], reputationTitle }
```

### POST /api/v1/seasons/:seasonId/reveal/open-hidden
```typescript
// Открыть скрытые атрибуты (стоит 5 кристаллов)
// Проверить баланс → списать → вернуть все атрибуты
{ data: { allAttributes: AttributeDto[], crystalBalance: number } }
```

## Логика репутационного титула
```typescript
function generateTitle(topAttributes: string[]): string {
  // Маппинг топ-атрибутов в титулы
  // Примеры:
  // "Убежит при пожаре" → "Олимпийский чемпион по бегу"
  // "Знает чужие секреты" → "Хранитель тайн"
  // "Притворяется что не смотрит аниме" → "Глубоко в шкафу"
  // Fallback: "Загадка века"
}
```

## Структура файлов
```
src/modules/reveal/
├── reveal.router.ts
├── reveal.service.ts
├── reveal.schema.ts
└── reveal.test.ts
src/jobs/
├── reveal-checker.job.ts
└── reveal-process.job.ts
```

## Критерии готовности
- [ ] reveal-checker запускается по расписанию
- [ ] При revealAt <= now() и кворуме → SeasonResult записываются
- [ ] Статус сезона меняется на REVEALED
- [ ] GET /reveal возвращает правильные данные
- [ ] Тест: агрегация при разных распределениях голосов
