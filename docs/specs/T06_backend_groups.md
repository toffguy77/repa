# T06 — Backend: Groups API

## Цель
API для создания групп, инвайт-ссылок, вступления, управления участниками.

## Эндпоинты

### POST /api/v1/groups
```typescript
// Request
{ name: string, categories: QuestionCategory[], telegramUsername?: string }
// Валидация: name 3-40 символов, categories непустой массив
// Логика:
// 1. Проверить лимит групп пользователя (≤ 10)
// 2. Создать Group (adminId = userId, inviteCode = cuid())
// 3. Создать GroupMember для создателя
// 4. Создать первый Season (status: VOTING, startsAt: ближайший понедельник, revealAt: ближайшая пятница 17:00 UTC)
// 5. Выбрать вопросы для сезона из bank (по категориям, случайно, 10 вопросов)

// Response
{ data: { group: GroupDto, inviteUrl: string } }
// inviteUrl = `https://repa.app/join/{inviteCode}`
```

### GET /api/v1/groups
```typescript
// Список групп текущего пользователя с активным сезоном
{ data: { groups: GroupListItemDto[] } }

// GroupListItemDto:
{
  id, name, memberCount, inviteCode,
  activeSeason: { id, status, revealAt, votedCount, totalCount, userVoted: boolean }
  telegramUsername: string | null
}
```

### GET /api/v1/groups/:id
```typescript
// Полная информация о группе
// Проверить что пользователь — участник
{
  data: {
    group: {
      id, name, adminId, telegramUsername, inviteCode,
      members: MemberDto[],
      activeSeason: SeasonDto,
    }
  }
}

// MemberDto: { id, username, avatarEmoji, avatarUrl, isAdmin, votingStreak }
```

### POST /api/v1/groups/join/:inviteCode
```typescript
// Вступить в группу по инвайт-коду
// Проверки:
// 1. Группа существует
// 2. Пользователь не является участником (409 если уже)
// 3. Лимит участников ≤ 50
// 4. Лимит групп пользователя ≤ 10

{ data: { group: GroupDto } }
```

### GET /api/v1/groups/join/:inviteCode/preview
```typescript
// Превью группы перед вступлением (не требует членства)
{ data: { name: string, memberCount: number, adminUsername: string } }
```

### DELETE /api/v1/groups/:id/leave
```typescript
// Выйти из группы
// Если пользователь — admin и он последний → удалить группу
// Если admin и есть другие → передать admin следующему участнику (по joinedAt)
{ data: { left: true } }
```

### PATCH /api/v1/groups/:id
```typescript
// Только admin
// Request: { name?: string, telegramUsername?: string }
{ data: { group: GroupDto } }
```

### GET /api/v1/groups/:id/invite-link
```typescript
// Регенерировать инвайт-ссылку (только admin)
{ data: { inviteUrl: string } }
```

## Логика создания сезона

```typescript
function getNextSeasonDates(): { startsAt: Date, revealAt: Date, endsAt: Date } {
  // startsAt = ближайший понедельник 00:00 МСК
  // revealAt = пятница той же недели 17:00 UTC (= 20:00 МСК)
  // endsAt   = воскресенье 20:59 UTC (= 23:59 МСК)
}
```

Вопросы для сезона: выбрать случайно из банка по указанным категориям, количество = `min(10, participantCount * 2)`, минимум 5.

## Структура файлов
```
src/modules/groups/
├── groups.router.ts
├── groups.service.ts
├── groups.schema.ts
└── groups.test.ts
```

## Тесты
- Создать группу → получить inviteUrl
- Вступить по inviteCode → появиться в members
- Лимит 10 групп → 409
- Лимит 50 участников → 409
- Не участник не может GET /groups/:id
- Admin может PATCH, рядовой участник — нет

## Критерии готовности
- [ ] Все эндпоинты работают
- [ ] inviteUrl содержит правильный inviteCode
- [ ] Первый сезон создаётся автоматически при создании группы
- [ ] Тесты проходят

---

## Дополнение: Job `season-creator`

### src/jobs/season-creator.job.ts

**Cron:** каждое воскресенье в 18:00 UTC (= 21:00 МСК)

```typescript
async function createNextSeasons() {
  // Найти все группы где:
  // 1. Есть активный сезон со статусом REVEALED или CLOSED
  // 2. НЕТ сезона со статусом VOTING (следующий ещё не создан)

  const groups = await prisma.group.findMany({
    where: {
      seasons: {
        some: { status: { in: ['REVEALED', 'CLOSED'] } },
        none: { status: 'VOTING' }
      }
    },
    include: {
      members: true,
      seasons: { orderBy: { number: 'desc' }, take: 1 }
    }
  })

  for (const group of groups) {
    if (group.members.length < 3) continue  // группа не активна

    const lastSeason = group.seasons[0]
    const nextNumber = (lastSeason?.number ?? 0) + 1
    const { startsAt, revealAt, endsAt } = getNextSeasonDates()

    // Выбрать вопросы: системные по категориям группы + победивший вопрос из голосования
    const questions = await selectQuestionsForSeason(group, lastSeason?.id)

    await prisma.$transaction(async (tx) => {
      const season = await tx.season.create({
        data: { groupId: group.id, number: nextNumber, startsAt, revealAt, endsAt }
      })
      await tx.seasonQuestion.createMany({
        data: questions.map((q, i) => ({
          seasonId: season.id, questionId: q.id, order: i
        }))
      })
      // Закрыть прошлый сезон если ещё не закрыт
      if (lastSeason?.status === 'REVEALED') {
        await tx.season.update({
          where: { id: lastSeason.id },
          data: { status: 'CLOSED' }
        })
      }
    })
  }
}

// Зарегистрировать в src/jobs/index.ts:
await queues.reveal.add('season-creator', {}, {
  repeat: { cron: '0 18 * * 0' }  // воскресенье 18:00 UTC
})
```

### Логика выбора вопросов следующего сезона
```typescript
async function selectQuestionsForSeason(group, prevSeasonId?) {
  // 1. Получить победивший вопрос из голосования (если есть)
  //    — вопрос с наибольшим числом голосов в next-season/vote-question
  // 2. Добавить случайные системные вопросы по категориям группы
  //    — исключить вопросы уже использованные в последних 3 сезонах (ротация)
  //    — total: min(10, memberCount * 2), минимум 5
}
```

### Критерии готовности job'а
- [ ] Новый сезон создаётся каждое воскресенье для всех активных групп
- [ ] Победивший вопрос из голосования включается в ротацию
- [ ] Вопросы из последних 3 сезонов не повторяются
- [ ] Если группа < 3 участников — сезон не создаётся
- [ ] Транзакция: создание сезона + закрытие предыдущего атомарно
