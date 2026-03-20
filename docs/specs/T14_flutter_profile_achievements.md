# T14 — Flutter: профиль участника и ачивки

## Цель
Экран профиля в группе: статистика, коллекция ачивок, «легенда».

## Эндпоинты (реализовать на backend)

### GET /api/v1/groups/:groupId/members/:userId/profile
```typescript
{
  data: {
    user: { username, avatarEmoji, avatarUrl },
    stats: {
      seasonsPlayed, votingStreak, maxVotingStreak,
      guessAccuracy,         // % правильных угадываний
      totalVotesCast,
      totalVotesReceived,
      topAttributeAllTime: { questionText, percentage }
    },
    achievements: AchievementDto[],    // все ачивки в группе
    legend: string,                    // автогенерированный текст
    seasonHistory: SeasonCardDto[],    // последние 5 сезонов
  }
}
```

### Логика генерации «легенды» (backend)
```typescript
function generateLegend(username: string, stats: Stats, achievements: Achievement[]): string {
  // Шаблоны по комбинациям ачивок:
  // SNIPER + LEGEND → «{name} — настоящий Снайпер и Легенда группы. ...»
  // STREAK_VOTER → «{name} не пропускает ни одного сезона. ...»
  // Fallback: «{name} — загадочная личность. ...»
  // Максимум 150 символов
}
```

## Flutter экраны

### MemberProfileScreen (`/groups/:groupId/members/:userId`)

**Секции:**
1. **Шапка**: аватар + ник + репутационный титул текущего сезона
2. **Легенда**: курсивный текст-описание
3. **Статистика**: сетка 2×3 карточек
   - Сезонов сыграно
   - Стрик голосований 🔥
   - Точность угадывания
   - Голосов получено
   - Лучший атрибут
4. **Ачивки**: горизонтальный scroll с карточками ачивок
5. **История сезонов**: последние 5 карточек репутации (миниатюры)

### AchievementBadge
```dart
// Отображение ачивки:
// - Эмодзи иконка (большая)
// - Название
// - Дата получения
// Locked state: серый, замок
// Unlocked state: цветной, subtle glow
```

### StatCard
```dart
// Карточка метрики:
// - Иконка
// - Большое число
// - Подпись
// - Анимация count-up при появлении
```

## Критерии готовности
- [ ] Профиль открывается из списка участников группы
- [ ] Все секции отображаются с реальными данными
- [ ] Легенда генерируется на backend
- [ ] Ачивки отображаются с locked/unlocked состоянием
- [ ] Анимация count-up для чисел
