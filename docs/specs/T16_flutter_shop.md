# T16 — Flutter: магазин кристаллов и платёжный flow

## Цель
Экран магазина, пакеты кристаллов, оплата через внешний браузер, обновление баланса.

## Экраны

### CrystalsShopScreen (`/shop`)
- Текущий баланс 💎 в шапке
- 4 карточки пакетов с ценами и бонусами
- «Популярный» пакет выделен (border акцент)
- Tap на пакет → `_initPurchase(packageId)`

### Платёжный flow
```dart
Future<void> _initPurchase(String packageId) async {
  // 1. POST /crystals/purchase/init → получить paymentUrl + paymentId
  // 2. Сохранить paymentId в state
  // 3. url_launcher.launchUrl(paymentUrl, mode: LaunchMode.externalBrowser)
  // 4. Показать BottomSheet «Ожидание оплаты...» с кнопкой «Я оплатил»
}

Future<void> _verifyPayment(String paymentId) async {
  // Polling: каждые 3 сек, max 10 попыток
  // GET /crystals/purchase/verify/{paymentId}
  // succeeded → показать успех + обновить баланс
  // canceled  → показать ошибку
  // pending   → продолжать polling
}
```

### Deeplink возврата
Обработать `repa.app/payment/return` через `app_links`:
- Извлечь `paymentId` из параметров
- Запустить `_verifyPayment`

### PurchaseSuccessSheet
- Анимация 💎 × N (flutter_animate)
- «+45 кристаллов добавлено»
- Новый баланс
- Кнопка «Отлично»

### CrystalBalanceWidget
Переиспользуемый виджет для показа баланса в AppBar:
```dart
// 💎 42  — тап открывает ShopScreen
```

## Интеграция детектора в RevealScreen
```dart
// DetectorPurchaseFlow:
// 1. Показать цену (10 💎)
// 2. Проверить баланс
// 3. Если хватает → POST /seasons/:id/detector → показать список
// 4. Если не хватает → открыть ShopScreen с deeplink-return
```

## Критерии готовности
- [ ] 4 пакета отображаются корректно
- [ ] Оплата открывается в браузере
- [ ] После возврата deeplink → polling verification
- [ ] Баланс обновляется после успешной оплаты
- [ ] Плашка «Внешний способ оплаты» видна пользователю
