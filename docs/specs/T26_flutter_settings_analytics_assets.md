# T26 — Flutter: настройки, аналитика, иконка, splash

## Цель
Экран настроек, загрузка аватара, аналитика, иконка приложения, splash screen, staging конфиги.

---

## 1. Backend: недостающие эндпоинты

### GET /api/v1/auth/username-check?username=xxx
```typescript
// Не требует авторизации (нужен при регистрации до получения токена)
// Проверить уникальность username
// Валидация: 3-20 символов, только a-z0-9_ и кириллица

{ data: { available: boolean } }
```

### POST /api/v1/auth/avatar
```typescript
// Загрузить фото аватара
// multipart/form-data, поле: file (image/jpeg, image/png, max 5MB)
// 1. Валидировать MIME-тип и размер
// 2. Ресайз до 256×256 через sharp
// 3. Загрузить в S3: key = `avatars/{userId}.jpg`
// 4. Обновить user.avatarUrl

// Dependencies: добавить sharp в package.json

{ data: { avatarUrl: string } }
```

---

## 2. Flutter: экран настроек

### SettingsScreen (`/settings`)
Открывается из HomeScreen (иконка шестерёнки в AppBar).

**Секции:**

**Профиль**
- Аватар (фото или эмодзи) + кнопка редактировать
- Никнейм + кнопка изменить (если прошло 30 дней)
- Год рождения

**Загрузка фото аватара:**
```dart
// image_picker пакет (добавить в pubspec.yaml)
// Tap на аватар → BottomSheet: «Камера» / «Галерея» / «Выбрать эмодзи»
// После выбора → crop до квадрата → POST /auth/avatar
// Optimistic update аватара пока загружается
```

**Уведомления**
```dart
// Switch для каждой категории (хранить в flutter_secure_storage):
// «Старт голосования» (понедельник)
// «Напоминание» (среда-четверг)
// «Reveal готов» (пятница)
// «Реакции на карточку»
// «Следующий сезон»
//
// При выключении — отправить на backend:
// PATCH /api/v1/push/preferences { category: string, enabled: boolean }
```

Backend endpoint:
```typescript
// PATCH /api/v1/push/preferences
// Request: { category: PushCategory, enabled: boolean }
// Хранить в отдельной таблице PushPreference или в Redis
// Учитывать при отправке в T17 jobs

{ data: { updated: true } }
```

**Аккаунт**
- Кнопка «Выйти» → очистить токен → редирект на auth
- Кнопка «Удалить аккаунт» → confirmation dialog → DELETE /auth/account

**О приложении**
- Версия приложения (`package_info_plus` пакет)
- Ссылка на политику конфиденциальности
- Ссылка на правила использования

---

## 3. Аналитика (Firebase Analytics)

### Добавить в pubspec.yaml
```yaml
firebase_analytics: ^10
```

### core/analytics/analytics_service.dart
```dart
class AnalyticsService {
  final FirebaseAnalytics _analytics;

  // Auth события
  Future<void> logSignUp(String method)   // method: 'phone' | 'apple' | 'google'
  Future<void> logLogin(String method)

  // Группы
  Future<void> logGroupCreated()
  Future<void> logGroupJoined()

  // Голосование
  Future<void> logVotingStarted(String groupId)
  Future<void> logVotingCompleted(String groupId, int questionsAnswered)
  Future<void> logVotingAbandoned(String groupId, int questionsAnswered)

  // Reveal
  Future<void> logRevealOpened(String groupId)
  Future<void> logCardShared(String method)   // method: 'telegram' | 'native'
  Future<void> logDetectorPurchased()
  Future<void> logHiddenAttributesOpened()

  // Монетизация
  Future<void> logPurchaseInitiated(String packageId, int priceKopecks)
  Future<void> logPurchaseCompleted(String packageId, int crystals)

  // Ачивки
  Future<void> logAchievementUnlocked(String achievementType)

  // Retention
  Future<void> logReactionSent()
  Future<void> logQuestionVoted()
}
```

Вызывать в нужных местах Notifier'ов и экранов. Не в UI-виджетах напрямую.

---

## 4. Иконка приложения

### flutter_launcher_icons
```yaml
# pubspec.yaml dev_dependencies:
flutter_launcher_icons: ^0.13

# flutter_launcher_icons.yaml:
flutter_launcher_icons:
  android: true
  ios: true
  image_path: "assets/icon/icon.png"         # 1024×1024 PNG
  adaptive_icon_background: "#1A0533"        # тёмно-фиолетовый
  adaptive_icon_foreground: "assets/icon/icon_foreground.png"
```

Создать `assets/icon/icon.png`:
- Фон: тёмно-фиолетовый `#1A0533`
- Эмодзи 🍆 по центру (или SVG-иллюстрация баклажана)
- Размер: 1024×1024

```bash
dart run flutter_launcher_icons
```

---

## 5. Splash Screen

### flutter_native_splash
```yaml
# pubspec.yaml dev_dependencies:
flutter_native_splash: ^2.4

# flutter_native_splash.yaml:
flutter_native_splash:
  color: "#1A0533"
  image: assets/splash/logo.png    # 🍆 + РЕПА логотип, белый на прозрачном
  android_12:
    color: "#1A0533"
    icon_background_color: "#1A0533"
    image: assets/splash/logo.png
  ios: true
```

```bash
dart run flutter_native_splash:create
```

### Логика splash в main.dart
```dart
// Пока инициализируется приложение:
// 1. Firebase.initializeApp()
// 2. Проверить токен в secure storage
// 3. Если токен есть → GET /auth/me (валидация)
// 4. Redirect: авторизован → /home, нет → /auth/phone
// Splash держится пока идёт инициализация (FlutterNativeSplash.remove() в конце)
```

---

## 6. Staging конфигурация

### Flutter Flavors
```dart
// lib/core/config/app_config.dart
enum Flavor { dev, staging, prod }

class AppConfig {
  final Flavor flavor;
  final String apiBaseUrl;
  final String appUrl;

  static const dev = AppConfig(
    flavor: Flavor.dev,
    apiBaseUrl: 'http://localhost:3000/api/v1',
    appUrl: 'http://localhost:3000',
  );
  static const staging = AppConfig(
    flavor: Flavor.staging,
    apiBaseUrl: 'https://staging-api.repa.app/api/v1',
    appUrl: 'https://staging.repa.app',
  );
  static const prod = AppConfig(
    flavor: Flavor.prod,
    apiBaseUrl: 'https://api.repa.app/api/v1',
    appUrl: 'https://repa.app',
  );
}
```

```bash
# Запуск по flavor:
flutter run --dart-define=FLAVOR=dev
flutter run --dart-define=FLAVOR=staging
flutter build ipa --dart-define=FLAVOR=prod
```

### Backend envs
Создать три `.env` файла:
- `.env.development` — localhost
- `.env.staging` — staging сервер, тестовые ключи ЮKassa/Telegram
- `.env.production` — prod

Добавить в `package.json`:
```json
"dev":     "NODE_ENV=development tsx watch src/server.ts",
"staging": "NODE_ENV=staging node dist/server.js",
"start":   "NODE_ENV=production node dist/server.js"
```

### Force Update механизм
```typescript
// Backend: GET /api/v1/app/version
// Response: { minVersion: '1.0.0', latestVersion: '1.2.0', forceUpdate: boolean }

// Flutter: проверять при старте через package_info_plus
// Если текущая версия < minVersion → показать ForceUpdateScreen (нельзя закрыть)
// Если текущая < latestVersion → показать мягкий баннер «Доступно обновление»
```

### Backoffice endpoint (минимальный)
```typescript
// Простая Basic Auth защита (admin:password из env)
// GET /admin/reports → список жалоб на вопросы с пагинацией
// PATCH /admin/reports/:id → { action: 'approve' | 'reject' }
//   approve → Question.status = ACTIVE
//   reject  → Question.status = REJECTED
// GET /admin/stats → DAU, MAU, groups count, revenue last 7 days
```

---

## Новые пакеты для pubspec.yaml
```yaml
dependencies:
  image_picker: ^1.1
  package_info_plus: ^8
  connectivity_plus: ^6
  firebase_analytics: ^10

dev_dependencies:
  flutter_launcher_icons: ^0.13
  flutter_native_splash: ^2.4
```

---

## Критерии готовности
- [ ] username-check endpoint работает без авторизации
- [ ] Фото аватара загружается в S3, ресайзится до 256×256
- [ ] SettingsScreen содержит все секции
- [ ] Настройки пушей сохраняются и учитываются при отправке
- [ ] Analytics события логируются (проверить в Firebase DebugView)
- [ ] Иконка 🍆 отображается на iOS и Android
- [ ] Splash screen с тёмным фоном
- [ ] `flutter run --dart-define=FLAVOR=dev` работает
- [ ] Force update экран показывается при устаревшей версии
- [ ] Backoffice `/admin/reports` доступен
