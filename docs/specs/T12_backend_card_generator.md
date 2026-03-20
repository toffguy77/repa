# T12 — Backend: генерация PNG карточек

## Цель
Сервис рендеринга карточки репутации в PNG через Puppeteer. Загрузка в S3.

## Архитектура
Puppeteer рендерит HTML-шаблон → скриншот → загрузка в Yandex Object Storage → URL в БД.

## Реализовать

### src/lib/s3.ts
```typescript
import { S3Client, PutObjectCommand, GetObjectCommand } from '@aws-sdk/client-s3'
import { getSignedUrl } from '@aws-sdk/s3-request-presigner'

export const s3 = new S3Client({
  endpoint: process.env.S3_ENDPOINT,
  region: process.env.S3_REGION,
  credentials: {
    accessKeyId: process.env.S3_ACCESS_KEY!,
    secretAccessKey: process.env.S3_SECRET_KEY!,
  }
})

export async function uploadBuffer(
  key: string,
  buffer: Buffer,
  contentType: string
): Promise<string>  // возвращает публичный URL

export async function getPublicUrl(key: string): Promise<string>
```

### src/modules/cards/card.template.ts
HTML-шаблон карточки (1080×1920px):

```typescript
export function buildCardHtml(data: CardData): string {
  // CardData:
  // { username, avatarEmoji, topAttributes, reputationTitle, groupName, seasonNumber }
  
  // Дизайн карточки:
  // - Фон: тёмно-фиолетовый градиент (статичный, без CSS-градиентов — используй SVG фон)
  // - Логотип 🍆 и «РЕПА» сверху
  // - Аватар-эмодзи в круге
  // - Никнейм пользователя
  // - Репутационный титул
  // - Топ-3 атрибута с процентами (прогресс-бары)
  // - Название группы и номер сезона снизу
  // Шрифт: системный sans-serif, встроить через base64 если нужно кириллица
}
```

### src/modules/cards/card.service.ts
```typescript
export async function generateCardImage(seasonResultData: CardData): Promise<string> {
  // 1. Запустить Puppeteer (headless)
  // 2. Установить viewport 1080x1920
  // 3. setContent(buildCardHtml(data))
  // 4. Подождать networkidle
  // 5. screenshot({ type: 'png', fullPage: false })
  // 6. Закрыть браузер
  // 7. Загрузить buffer в S3: key = `cards/{seasonId}/{userId}.png`
  // 8. Вернуть публичный URL
}

// Puppeteer инстанс — singleton с переиспользованием браузера
// Не создавать новый браузер на каждый запрос
```

### Интеграция в Reveal

В `reveal-process.job.ts` после записи SeasonResult:
```typescript
// Для каждого участника запустить генерацию карточки параллельно
await Promise.all(
  members.map(m => generateAndSaveCard(m.userId, seasonId))
)
```

Сохранять URL в `SeasonResult` или отдельной таблице `CardCache`:
```prisma
model CardCache {
  id       String @id @default(cuid())
  userId   String
  seasonId String
  imageUrl String
  createdAt DateTime @default(now())
  
  @@unique([userId, seasonId])
}
```

### Эндпоинт

#### GET /api/v1/seasons/:seasonId/my-card-url
```typescript
// Вернуть URL карточки текущего пользователя
// Если ещё не сгенерирована — запустить генерацию и вернуть статус
{ data: { imageUrl: string | null, status: 'ready' | 'generating' } }
```

## Критерии готовности
- [ ] Puppeteer запускается в Docker (добавить chromium зависимость)
- [ ] HTML-шаблон рендерится с кириллицей
- [ ] PNG загружается в S3
- [ ] URL возвращается в API
- [ ] Браузер переиспользуется между запросами

## Docker
Добавить в backend Dockerfile:
```dockerfile
RUN apt-get install -y chromium
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium
```
