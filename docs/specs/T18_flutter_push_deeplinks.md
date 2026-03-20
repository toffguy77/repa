# T18 — Flutter: Push handling, deeplinks, реакции на карточки

## Цель
Обработка FCM пушей, deeplinks, реакции на карточки участников.

## FCM Push Handling

### Foreground пуши (приложение открыто)
Показывать in-app снекбар/баннер, не системное уведомление.

### Background / Terminated пуши
Tap на уведомление → навигация по `data.screen`:
```dart
// data: { screen: 'reveal', groupId, seasonId }
//   → /groups/{groupId}/reveal
// data: { screen: 'reveal-waiting', groupId }
//   → /groups/{groupId} (с подсветкой прогресс-бара)
// data: { screen: 'vote', groupId }
//   → /groups/{groupId}/vote
// data: { screen: 'question-vote', groupId }
//   → /groups/{groupId}/question-vote
// data: { screen: 'shop' }
//   → /shop
```

### Firebase настройка
```dart
// main.dart
await Firebase.initializeApp()
await FirebaseMessaging.instance.requestPermission()
final token = await FirebaseMessaging.instance.getToken()
// POST /push/register { token, platform }

// Обработчики:
FirebaseMessaging.onMessage.listen(_handleForeground)
FirebaseMessaging.onMessageOpenedApp.listen(_handleTap)
FirebaseMessaging.instance.getInitialMessage().then(_handleTap)
```

## Deeplinks (app_links)

```dart
// В AppRouter инициализировать app_links
// Обработать:
// repa.app/join/{code}        → JoinGroupScreen
// repa.app/payment/return     → верификация платежа
// repa.app/group/{groupId}    → GroupScreen
```

## Реакции на карточки

### Backend (добавить в T10 или здесь)

```typescript
// POST /api/v1/seasons/:seasonId/members/:targetId/reactions
// Request: { emoji: '😂' | '🔥' | '💀' | '👀' | '🫡' }
// Один пользователь — одна реакция на одну карточку за сезон (upsert)
// Реакции анонимны

// GET /api/v1/seasons/:seasonId/members/:targetId/reactions
// Response: { '😂': 5, '🔥': 3, '💀': 1, '👀': 2, '🫡': 0 }
// (без userId — анонимно)
```

Модель:
```prisma
model Reaction {
  id       String @id @default(cuid())
  seasonId String
  reactorId String
  targetId  String
  emoji     String
  createdAt DateTime @default(now())

  @@unique([seasonId, reactorId, targetId])
}
```

### Flutter — ReactionBar
```dart
// Отображается под карточкой участника в MembersRevealScreen
// 5 эмодзи кнопок с счётчиками
// Tap → анимация отправки + optimistic update
// Уже поставленная реакция подсвечена
```

### Push при получении реакции
Backend: после записи реакции → `push.add('reaction', { targetId, seasonId })`
Job отправляет: «Кто-то отреагировал на твою репу»

## Критерии готовности
- [ ] FCM токен регистрируется при запуске
- [ ] Tap на пуш открывает нужный экран
- [ ] Deeplink `/join/{code}` работает
- [ ] Deeplink `/payment/return` запускает верификацию
- [ ] Реакции отображаются анонимно
- [ ] Optimistic update реакции (без задержки UI)
