# Repa

## Setup

```bash
docker compose up -d
cd backend
cp .env.example .env
make migrate
make seed
make dev
```
