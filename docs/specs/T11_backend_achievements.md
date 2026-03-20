# T11 — Backend: движок расчёта ачивок

## Цель
BullMQ worker для расчёта ачивок после каждого Reveal.

## Job: `calculate-achievements` (параметр: seasonId)

Для каждого участника группы проверить условия всех ачивок и записать новые.

### Реализовать расчёт следующих ачивок:

**Угадывание (требуют анализа votes текущего сезона):**

```typescript
// SNIPER: угадал ≥ 80% вопросов (проголосовал за того, кто получил большинство)
async function checkSniper(voterId, seasonId): Promise<boolean>

// TELEPATH: угадал всех правильно в сессии (100%)
async function checkTelepath(voterId, seasonId): Promise<boolean>

// BLIND: точность < 20%
async function checkBlind(voterId, seasonId): Promise<boolean>

// ORACLE: 3 сезона подряд точность > 70% (нужна история)
async function checkOracle(userId, groupId): Promise<boolean>
```

**Знание конкретных людей:**
```typescript
// EXPERT_OF: угадывал правильно за конкретного участника 5+ раз подряд
// metadata: { targetId, targetUsername }
async function checkExpertOf(voterId, groupId): Promise<AchievementMeta | null>
```

**Репутация (анализ SeasonResult):**
```typescript
// LEGEND: один и тот же топ-атрибут 5 сезонов подряд
async function checkLegend(userId, groupId): Promise<boolean>

// MONOPOLIST: получает ≥ 70% голосов по одному вопросу
async function checkMonopolist(userId, seasonId): Promise<boolean>

// PIONEER: первым получил данный атрибут в группе
async function checkPioneer(userId, seasonId, groupId): Promise<AchievementMeta | null>
// metadata: { questionId, questionText }
```

**Активность:**
```typescript
// STREAK_VOTER: голосовал N недель подряд → обновить votingStreak в UserGroupStat
// Ачивка выдаётся при 5, 10, 20 сезонах подряд
async function updateVotingStreak(userId, groupId, seasonId): Promise<void>

// FIRST_VOTER: первым завершил голосование в сезоне
async function checkFirstVoter(userId, seasonId): Promise<boolean>

// NIGHT_OWL: голосовал в 23:00–03:00 МСК (по timestamp первого vote)
async function checkNightOwl(userId, seasonId): Promise<boolean>
```

**Шеринг:**
```typescript
// RECRUITER: привёл 3+ участников (по joinedAt после его joinedAt в группе)
async function checkRecruiter(userId, groupId): Promise<boolean>
```

### Основной воркер:

```typescript
async function calculateAchievements(seasonId: string) {
  const season = /* fetch with group and members */
  
  for (const member of season.group.members) {
    const newAchievements: AchievementType[] = []
    
    // Проверить каждую ачивку
    // Записать только те, которых ещё нет у пользователя в этой группе
    // (кроме STREAK_VOTER — она выдаётся несколько раз)
    
    await prisma.achievement.createMany({
      data: newAchievements.map(type => ({
        userId: member.userId, groupId: season.groupId,
        seasonId, achievementType: type
      })),
      skipDuplicates: true
    })
  }
  
  // Обновить UserGroupStat для всех участников
  await updateGroupStats(seasonId)
}
```

### Обновление UserGroupStat:
```typescript
async function updateGroupStats(seasonId: string) {
  // Обновить для каждого участника:
  // seasonsPlayed++
  // votingStreak (если голосовал) или → 0
  // guessAccuracy (скользящее среднее за последние 5 сезонов)
  // totalVotesCast, totalVotesReceived
}
```

## Структура файлов
```
src/jobs/
└── achievements.job.ts
src/modules/achievements/
├── achievements.service.ts   # расчёт ачивок
└── achievements.test.ts
```

## Тесты
- SNIPER: проголосовал за победителя в 8 из 10 вопросов → получил ачивку
- LEGEND: один топ-атрибут 5 сезонов подряд → получил ачивку
- Дублирующаяся ачивка не записывается дважды
- votingStreak корректно обновляется (увеличивается при голосовании, сбрасывается при пропуске)

## Критерии готовности
- [ ] Воркер запускается после Reveal
- [ ] Ачивки рассчитываются для всех участников
- [ ] UserGroupStat обновляется
- [ ] Нет дублирования ачивок
- [ ] Тесты проходят
