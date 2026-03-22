# T19 — Backend: Telegram bot и интеграция

## Цель
Telegram бот для публикации постов в привязанные чаты. Connect-flow. Автопосты.

## src/lib/telegram.ts
```typescript
import TelegramBot from 'node-telegram-bot-api'

export const bot = new TelegramBot(process.env.TELEGRAM_BOT_TOKEN!, {
  webHook: { port: 0 }  // webhook настраивается отдельно
})

export async function sendMessage(chatId: string, text: string, options?: object): Promise<void>
export async function sendPhoto(chatId: string, imageUrl: string, caption: string): Promise<void>
export async function sendMediaGroup(chatId: string, media: object[]): Promise<void>
```

## Webhook endpoint

### POST /api/v1/telegram/webhook
```typescript
// Проверить X-Telegram-Bot-Api-Secret-Token header
// Обработать updates:

// Команда /connect CODE
if (text.startsWith('/connect ')) {
  const code = text.split(' ')[1]
  const group = await findGroupByConnectCode(code)
  if (!group) return bot.sendMessage(chatId, '❌ Код не найден или истёк')
  
  // Проверить что бот является администратором чата
  const chatMember = await bot.getChatMember(chatId, bot.options.botId)
  if (!['administrator', 'creator'].includes(chatMember.status)) {
    return bot.sendMessage(chatId, '❌ Сделайте бота администратором чата')
  }
  
  await prisma.group.update({
    where: { id: group.id },
    data: {
      telegramChatId: String(chatId),
      telegramConnectCode: null,
      telegramConnectExpiry: null
    }
  })
  await bot.sendMessage(chatId, `✅ Группа «${group.name}» подключена к Репе!`)
}

// Команда /repa
if (text === '/repa') {
  const group = await findGroupByChatId(chatId)
  const season = /* активный сезон */
  const progress = /* прогресс голосования */
  await bot.sendMessage(chatId, formatSeasonStatus(group, season, progress))
}

// Команда /disconnect
if (text === '/disconnect') {
  // Только если отправитель — администратор группы в Репе
  await prisma.group.update({ data: { telegramChatId: null } })
  await bot.sendMessage(chatId, '✅ Telegram-чат отвязан от Репы')
}

// Событие: бот удалён из чата
if (update.my_chat_member?.new_chat_member?.status === 'kicked') {
  await prisma.group.update({
    where: { telegramChatId: String(chatId) },
    data: { telegramChatId: null }
  })
}
```

## Groups API — connect endpoints

### POST /api/v1/groups/:groupId/telegram/generate-code
```typescript
// Только admin
// Генерировать connect-код REPA-XXXX (8 символов)
// Сохранить в group.telegramConnectCode с TTL 24 часа
// Вернуть инструкцию

{
  data: {
    connectCode: string,           // REPA-X7K2
    instruction: string,           // «Добавьте @repaapp_bot в чат и напишите /connect REPA-X7K2»
    expiresAt: string
  }
}
```

### DELETE /api/v1/groups/:groupId/telegram
```typescript
// Отвязать Telegram (admin only)
{ data: { disconnected: true } }
```

## BullMQ Jobs — Telegram посты

### Job: `telegram-season-start` (запускается из weekly-scheduler)
```typescript
async function postSeasonStart(groupId: string) {
  const group = await prisma.group.findUnique({ where: { id: groupId }})
  if (!group.telegramChatId) return
  
  await bot.sendMessage(group.telegramChatId,
    `🍆 Новый сезон в группе «${group.name}»!\n\nГолосуй в приложении 👇`,
    { reply_markup: { inline_keyboard: [[
      { text: '🗳 Проголосовать', url: `https://repa.app/group/${groupId}` }
    ]]}}
  )
}
```

### Job: `telegram-reveal-post`
```typescript
async function postReveal(seasonId: string) {
  const group = /* fetch */
  if (!group.telegramChatId) return
  
  const results = /* топ атрибуты по группе, агрегированные */
  const topAchievement = /* ачивка недели если есть */
  
  const text = formatRevealPost(group.name, results, topAchievement)
  // Формат:
  // 🍆 Репа подвела итоги — [Название]
  // 
  // Топ атрибуты этой недели:
  // • «Убежит при пожаре» — 3 человека
  // • «Знает секреты» — 5 человек
  //
  // 👑 Ачивка недели: Серёга — «Легенда группы»
  //
  // [Открыть свою репу →]  [Посмотреть всех →]
  
  await bot.sendMessage(group.telegramChatId, text, {
    reply_markup: { inline_keyboard: [[
      { text: '🍆 Открыть свою репу', url: `https://repa.app/group/${groupId}/reveal` },
      { text: '👥 Посмотреть всех', url: `https://repa.app/group/${groupId}/reveal/members` }
    ]]}
  })
}
```

### Job: `telegram-share-card` (ручной шеринг с клиента)
```typescript
async function shareCard(userId: string, seasonId: string) {
  const group = /* по seasonId */
  if (!group.telegramChatId) return
  
  const card = /* CardCache для userId + seasonId */
  const user = /* username */
  
  await bot.sendPhoto(
    group.telegramChatId,
    card.imageUrl,
    `🍆 @${user} поделился своей репой`
  )
}
```

### POST /api/v1/seasons/:seasonId/share-to-telegram
```typescript
// Ручной шеринг карточки в Telegram-чат группы
// Проверить: группа имеет telegramChatId
// Запустить job telegram-share-card

{ data: { shared: true } }
```

## Критерии готовности
- [ ] `/connect CODE` привязывает чат
- [ ] Бот проверяет права администратора перед привязкой
- [ ] Удаление бота из чата → автоотвязка
- [ ] Reveal-пост публикуется без индивидуальных карточек
- [ ] Ручной шеринг работает
- [ ] `/repa` возвращает статус сезона
