# T20 — Flutter: Telegram UI

## Цель
Экраны привязки Telegram, кнопки перехода в чат, шеринг карточки.

## Что реализовать

### TelegramSetupScreen (в настройках группы, только для admin)

**Состояние: не привязан**
- Иллюстрация + объяснение зачем нужна интеграция
- Кнопка «Подключить Telegram»
- Tap → POST /groups/:id/telegram/generate-code → показать ConnectInstructionSheet

**ConnectInstructionSheet (BottomSheet)**
```
Как подключить:
1. Добавьте @repaapp_bot в ваш Telegram-чат
2. Сделайте бота администратором
3. Напишите в чат:
   [REPA-X7K2]  [Скопировать]

После выполнения нажмите «Проверить»
[Скопировать код]  [Открыть Telegram]  [Проверить]
```

- «Открыть Telegram» → `url_launcher` с `tg://`
- «Проверить» → polling GET /groups/:id (проверить telegramChatId != null)
- Таймер: код действителен 24 часа, показать countdown

**Состояние: привязан**
- Зелёный чекмарк + «Подключено»
- Название или username чата
- Настройки: что публиковать автоматически (будущее)
- Кнопка «Отвязать»

### Кнопка «Открыть в Telegram» в GroupScreen
```dart
// Показывается только если group.telegramUsername != null
// Или если telegramChatId привязан
IconButton(
  icon: Icon(Icons.telegram),  // или SVG иконка Telegram
  onPressed: () => launchUrl(Uri.parse('https://t.me/${group.telegramUsername}')),
)
```

### Кнопка «Поделиться в Telegram-чат» на RevealScreen
```dart
// Показывается только если группа имеет привязанный Telegram
// Tap → POST /seasons/:id/share-to-telegram
// Feedback: SnackBar «Карточка опубликована в чате»
```

### Нативный Share + Telegram

Кнопка «Поделиться» → Share.shareXFiles():
- Сначала скачать PNG карточки с S3 URL во временный файл
- `Share.shareXFiles([XFile(path)], text: 'Моя репа 🍆 repa.app')`
- Пользователь сам выбирает куда поделиться (Telegram, ВКонтакте, и т.д.)

## Критерии готовности
- [ ] Connect flow понятен без инструкции
- [ ] Код копируется в буфер обмена
- [ ] Polling подтверждает привязку
- [ ] Кнопка Telegram видна только при привязанном чате
- [ ] Share sheet открывается с PNG карточкой
