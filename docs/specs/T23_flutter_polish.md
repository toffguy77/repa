# T23 — Flutter: design polish, анимации, edge cases

## Цель
Доработать UX до production-качества: пустые состояния, ошибки, анимации, accessibility.

## Что реализовать

### Пустые состояния (EmptyStateWidget)
```dart
// Переиспользуемый виджет:
// - Иллюстрация (эмодзи крупно)
// - Заголовок
// - Подзаголовок
// - CTA кнопка (опционально)

// Применить для:
// - HomeScreen без групп → «Создай первую группу или вступи по ссылке»
// - GroupScreen без участников → «Поделись ссылкой с друзьями»
// - AchievementsScreen без ачивок → «Голосуй и узнай что о тебе думают»
```

### Skeleton loading
```dart
// Пока данные загружаются — показывать skeleton вместо спиннера
// Применить для:
// - Список групп (GroupCard skeleton)
// - Список участников (MemberAvatar skeleton)
// - Reveal Screen (CardSkeleton)
```

### Error states
```dart
// ErrorWidget: иконка, текст ошибки, кнопка «Повторить»
// Обернуть все AsyncValue.when в единый ErrorWidget
// Сетевая ошибка → «Нет соединения, проверь интернет»
// 401 → автоматический logout + редирект на auth
// 5xx → «Что-то пошло не так, попробуй позже»
```

### Offline индикатор
```dart
// Connectivity check (connectivity_plus пакет)
// При потере сети → баннер вверху «Нет соединения»
// При восстановлении → авторефреш текущего экрана
```

### Pull-to-refresh
Добавить на все списочные экраны:
- HomeScreen (список групп)
- GroupScreen (участники + прогресс)
- MembersRevealScreen

### Haptic feedback
```dart
// Выбор участника при голосовании → HapticFeedback.mediumImpact()
// Получение ачивки → HapticFeedback.heavyImpact()
// Успешный платёж → HapticFeedback.heavyImpact()
// Tap на реакцию → HapticFeedback.lightImpact()
```

### Таймер до Reveal (RevealCountdownWidget)
```dart
// Показывается везде где ожидается Reveal
// Формат: «Пятница, 20:00 · осталось 2д 14ч»
// Обновляется каждую минуту через Timer.periodic
// При < 1 часа → красный цвет + «Скоро!»
```

### Анимации переходов
```dart
// go_router pageBuilder с кастомными переходами:
// Drill-down (GroupScreen): slide from right
// Modal (BottomSheet, Reveal): slide from bottom
// Auth flow: fade
```

### AppBar глобальный
```dart
// Содержит баланс кристаллов 💎 N (tap → ShopScreen)
// Уведомление-точка если есть непрочитанный Reveal
```

### Accessibility
```dart
// Все интерактивные элементы: semanticsLabel
// Минимальный tap target 44×44
// Поддержка Dynamic Type (textScaleFactor)
```

## Критерии готовности
- [ ] Все экраны имеют пустое состояние
- [ ] Skeleton loader на всех списках
- [ ] Ошибки обрабатываются и показываются пользователю
- [ ] Haptic feedback на ключевых действиях
- [ ] Таймер обратного отсчёта работает
- [ ] Pull-to-refresh на всех списках
- [ ] 60fps анимации (проверить DevTools)
