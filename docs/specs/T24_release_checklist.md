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
- [ ] GET /seasons/:id/reveal не содержит voterId в контексте конкретного голоса
- [ ] GET /seasons/:id/detector содержит только список userId (без questionId)
- [ ] GET /seasons/:id/progress не раскрывает кто именно проголосовал

## Backend production конфигурация

### Dockerfile (backend)
```dockerfile
FROM node:22-alpine
RUN apk add --no-cache chromium
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium
WORKDIR /app
COPY package*.json .
RUN npm ci --only=production
COPY dist .
CMD ["node", "server.js"]
```

### Telegram Webhook регистрация
```bash
curl -X POST https://api.telegram.org/bot{TOKEN}/setWebhook \
  -d url=https://api.repa.app/api/v1/telegram/webhook \
  -d secret_token={TELEGRAM_WEBHOOK_SECRET}
```

### BullMQ — регистрация всех cron jobs
```typescript
// src/jobs/index.ts — зарегистрировать все cron jobs при старте
export async function startAllJobs() {
  await queues.reveal.add('reveal-checker', {}, { repeat: { every: 60000 }})
  await queues.push.add('weekly-scheduler', {}, { repeat: { cron: '0 14 * * 1' }})
  // ... все остальные cron jobs
}
```

## Flutter production конфигурация

### iOS
- [ ] Bundle ID: `app.repa`
- [ ] Universal Links: `apple-app-site-association` на `repa.app`
- [ ] APNs сертификат загружен в Firebase
- [ ] Privacy strings в Info.plist (camera, photo library)
- [ ] `flutter build ipa --release`

### Android
- [ ] Package: `app.repa`
- [ ] App Links: `assetlinks.json` на `repa.app/.well-known/`
- [ ] `flutter build appbundle --release`
- [ ] ProGuard правила для Retrofit/Gson

### Обоих
- [ ] Firebase `google-services.json` / `GoogleService-Info.plist` для prod окружения
- [ ] API base URL → production endpoint
- [ ] Убрать OTP код из dev-ответа
- [ ] Crashlytics включён

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
- [ ] Все E2E сценарии пройдены
- [ ] Тест анонимности пройден
- [ ] Backend задеплоен, webhook зарегистрирован
- [ ] BullMQ cron jobs активны
- [ ] Flutter builds без ошибок для iOS и Android
- [ ] Universal Links / App Links работают
- [ ] Push уведомления доходят на реальных устройствах
