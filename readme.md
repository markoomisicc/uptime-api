# Uptime API

Backend for a website and API uptime monitoring service, with downtime alerts and public status pages.

Frontend: _(link to Angular repository, coming soon)_

## Tech stack

- Go (standard library `net/http`)
- PostgreSQL 18
- Docker Compose

## Getting started

Requirements: Go 1.22+, Docker

### 1. Start the database

```bash
cp .env.example .env
docker compose up -d
```

If port 5432 is already in use, change `DB_PORT` in `.env`.

### 2. Run the API

```bash
go run ./cmd/api
```

The server starts on `http://localhost:8080` (set the `PORT` environment variable to change it).

### Development with live reload

```bash
go install github.com/air-verse/air@latest
air
```

## Endpoints

| Method | Path      | Description         |
|--------|-----------|---------------------|
| GET    | `/health` | Server health check |