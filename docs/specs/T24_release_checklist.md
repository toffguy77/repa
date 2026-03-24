# T24 — E2E чеклист и подготовка к релизу

## Цель
Финальная проверка всех сценариев, конфигурация для production, подготовка к сабмиту.

## E2E сценарии (проверить вручную или написать тесты)

### Критические пути

**Регистрация и группа:**
- [ ] Новый пользователь: OTP → profile setup → создать группу → получить invite link
- [ ] Второй пользователь: вступить по ссылке → появиться в списке участников
- [ ] Deeplink `/join/{code}` открывает приложение и предлагает вступить

**Голосование:**
- [ ] Проголосовать за всех участников → прогресс 100%
- [ ] Прервать голосование → вернуться → продолжить с того же места
- [ ] Попытка проголосовать за себя → ошибка (или участник не отображается)
- [ ] Повторный голос по тому же вопросу → ошибка

**Reveal:**
- [ ] Принудительный Reveal (через Redis: установить revealAt = now)
- [ ] Карточка отображается с атрибутами
- [ ] Скрытые атрибуты заблюрены
- [ ] Шеринг PNG открывает нативный share sheet
- [ ] Детектор за 10 кристаллов показывает voters

**Монетизация:**
- [ ] Открыть магазин → выбрать пакет → открыть в браузере
- [ ] Вернуться в приложение → баланс обновился
- [ ] Недостаточно кристаллов → сообщение + переход в магазин

**Telegram:**
- [ ] Привязать чат через connect-код
- [ ] `/repa` команда возвращает статус
- [ ] Reveal-пост публикуется без индивидуальных карточек
- [ ] Удаление бота → автоотвязка

**Ачивки:**
- [ ] После Reveal ачивки отображаются в профиле
- [ ] Стрик увеличивается при каждом голосовании
- [ ] Легенда генерируется корректно

### Тест анонимности
- [x] GET /seasons/:id/reveal не содержит voterId в контексте конкретного голоса
- [x] GET /seasons/:id/detector содержит только список userId (без questionId)
- [x] GET /seasons/:id/progress не раскрывает кто именно проголосовал
- [x] Source scan: handler + service files checked for voter_id in JSON tags

## Backend production конфигурация

### Dockerfile (backend)
```dockerfile
# See backend/Dockerfile — multi-stage Go build with Chromium for chromedp
FROM golang:1.26-alpine AS builder
# ... (full file in backend/Dockerfile)
```

### Telegram Webhook регистрация
```bash
curl -X POST https://api.telegram.org/bot{TOKEN}/setWebhook \
  -d url=https://api.repa.app/api/v1/telegram/webhook \
  -d secret_token={TELEGRAM_WEBHOOK_SECRET}
```

### Asynq — cron jobs
All cron jobs are registered in `backend/cmd/server/main.go` → `startScheduler()` using asynq scheduler.
Already implemented: reveal-checker, push schedule (Mon–Sun), telegram season-start.

## Flutter production конфигурация

### iOS
- [x] Bundle ID: `app.repa.repa`
- [x] Universal Links: `apple-app-site-association` served from backend `/.well-known/` route
- [x] APNs сертификат загружен в Firebase
- [x] Privacy strings в Info.plist (camera, photo library, save to gallery)
- [x] `flutter build ipa --release` (verified --no-codesign)

### Android
- [x] Package: `app.repa.repa`
- [x] App Links: `assetlinks.json` served from backend `/.well-known/` route
- [ ] `flutter build appbundle --release`
- [x] ProGuard правила для Flutter/Gson/Firebase/Crashlytics

### Обоих
- [x] Firebase `google-services.json` / `GoogleService-Info.plist` для prod окружения
- [x] API base URL → configurable via `--dart-define=API_BASE_URL=...`
- [x] OTP код защищён флагом `DEV_MODE` (не возвращается в production)
- [x] Crashlytics включён

## App Store / Google Play

### Описание для сторов
```
«Репа» — узнай, что о тебе думают друзья.

Создай группу, пригласи класс или компанию.
Каждую неделю — анонимное голосование.
В пятницу в 20:00 — Reveal: твоя карточка репутации.

🍆 Репа — это зеркало. Честное.
```

### Возрастной рейтинг
- iOS: 12+ (юмористический контент)
- Android: PEGI 12

### Категория
- iOS: Social Networking
- Android: Social

## Критерии готовности к релизу
- [ ] Все E2E сценарии пройдены (manual)
- [x] Тест анонимности пройден (automated, 5 tests)
- [ ] Backend задеплоен, webhook зарегистрирован
- [x] Asynq cron jobs зарегистрированы в startScheduler()
- [x] Flutter builds без ошибок для iOS (verified)
- [x] Universal Links / App Links routes served from backend
- [ ] Push уведомления доходят на реальных устройствах (manual)
