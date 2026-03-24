# --- Build stage ---
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# --- Runtime stage ---
FROM alpine:3.20

RUN apk add --no-cache \
    ca-certificates \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ttf-freefont \
    tzdata

# chromedp needs to know where Chromium is
ENV CHROME_PATH=/usr/bin/chromium-browser
ENV TZ=UTC

WORKDIR /app

COPY --from=builder /server .
COPY internal/db/migrations ./internal/db/migrations

EXPOSE 3000

CMD ["./server"]
