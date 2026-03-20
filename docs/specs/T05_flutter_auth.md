# T05 — Flutter: создание проекта и Auth screens

## Цель
Создать Flutter-проект с полной инфраструктурой и экраны авторизации (телефон + OTP).

## Создание проекта
```bash
flutter create --org app.repa --project-name repa mobile
```

## pubspec.yaml зависимости
```yaml
dependencies:
  flutter_riverpod: ^2.5
  riverpod_annotation: ^2.3
  go_router: ^13
  dio: ^5.4
  retrofit: ^4.1
  freezed_annotation: ^2.4
  json_annotation: ^4.8
  flutter_secure_storage: ^9
  firebase_core: ^2
  firebase_messaging: ^14
  app_links: ^6
  share_plus: ^7
  url_launcher: ^6
  flutter_animate: ^4
  pinput: ^3          # OTP input
  cached_network_image: ^3
  intl: ^0.19

dev_dependencies:
  build_runner: ^2.4
  riverpod_generator: ^2.3
  freezed: ^2.4
  json_serializable: ^6.7
  retrofit_generator: ^8.1
```

## Структура проекта
```
mobile/lib/
├── main.dart
├── core/
│   ├── api/
│   │   ├── api_client.dart          # Dio настройка
│   │   ├── api_service.dart         # Retrofit интерфейс
│   │   └── models/                  # Response модели (freezed)
│   ├── router/
│   │   └── app_router.dart          # go_router
│   ├── theme/
│   │   ├── app_colors.dart
│   │   ├── app_text_styles.dart
│   │   └── app_theme.dart
│   └── providers/
│       ├── auth_provider.dart       # текущий пользователь
│       └── api_provider.dart        # Dio/Retrofit singleton
└── features/
    └── auth/
        ├── data/
        │   └── auth_repository.dart
        ├── domain/
        │   └── user.dart            # freezed entity
        └── presentation/
            ├── phone_screen.dart
            ├── otp_screen.dart
            ├── profile_setup_screen.dart
            └── auth_notifier.dart
```

## Что реализовать

### core/theme/app_colors.dart
```dart
class AppColors {
  static const primary = Color(0xFF7C3AED);      // фиолетовый
  static const primaryLight = Color(0xFFEDE9FE);
  static const background = Colors.white;
  static const surface = Color(0xFFF9FAFB);
  static const textPrimary = Color(0xFF111827);
  static const textSecondary = Color(0xFF6B7280);
  static const error = Color(0xFFEF4444);
  static const success = Color(0xFF10B981);
}
```

### core/api/api_client.dart
Dio с:
- BaseUrl из конфига (`const String apiBaseUrl`)
- Interceptor: добавляет `Authorization: Bearer {token}` из flutter_secure_storage
- Interceptor: при 401 → очищает токен → редирект на auth
- Обработка DioException → AppException с понятным сообщением

### features/auth/presentation/phone_screen.dart
- Поле ввода телефона с маской `+7 (___) ___-__-__`
- Кнопка «Получить код»
- Валидация формата
- Loading state во время запроса
- Обработка ошибок (rate limit, сетевая ошибка)

### features/auth/presentation/otp_screen.dart
- 6 ячеек Pinput для ввода кода
- Автоотправка при заполнении всех 6 цифр
- Таймер обратного отсчёта 5 минут
- Кнопка «Отправить код повторно» (активна после истечения таймера)
- Loading + error state

### features/auth/presentation/profile_setup_screen.dart
Показывается только при `isNew: true`:
- Поле username (с проверкой уникальности через API в реальном времени, debounce 500ms)
- Выбор аватара-эмодзи из сетки (20 эмодзи)
- Поле год рождения
- Кнопка «Готово»

### core/router/app_router.dart
```dart
// Маршруты:
// /auth/phone
// /auth/otp
// /auth/setup
// /home        (stub, реализуется в T07)
// /group/:id   (stub)

// Redirect: если нет токена → /auth/phone
// Если есть токен → /home
```

### Хранение токена
`flutter_secure_storage`: ключ `auth_token`. Читается в API interceptor.

## Критерии готовности
- [ ] `flutter run` запускается без ошибок
- [ ] Экран ввода телефона отображается
- [ ] OTP отправляется, код вводится, токен сохраняется
- [ ] После успешной авторизации → редирект на /home (заглушка)
- [ ] Новый пользователь проходит profile setup
- [ ] Нет hardcoded строк кроме русскоязычного UI

## Не делать
- Не реализовывать Sign in with Apple/Google (добавить кнопки-заглушки)
- Не реализовывать экраны групп (T07)
