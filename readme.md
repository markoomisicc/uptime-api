# Uptime API

Backend for a website and API uptime monitoring service, with downtime alerts and public status pages.

Frontend: _(link to Angular repository, coming soon)_

## Tech stack

- Go (standard library `net/http`)
- PostgreSQL 18
- Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations

## Getting started

Requirements: Go 1.22+, Docker, `make`

### 1. Start the database

```bash
cp .env.example .env
docker compose up -d
```

If port 5432 is already in use, change `DB_PORT` in `.env`.

### 2. Run migrations

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
make migrate-up
```

### 3. Run the API

```bash
go run ./cmd/api
```

The server starts on `http://localhost:8080` (set the `PORT` environment variable to change it).

### Development with live reload

```bash
go install github.com/air-verse/air@latest
air
```

## Database migrations

| Command | Description |
|---|---|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back the last migration |
| `make migrate-create name=<name>` | Create a new migration |

Migrations that have been merged into `main` must never be edited. Create a new migration instead.

## Endpoints

| Method | Path      | Description                           |
|--------|-----------|---------------------------------------|
| GET    | `/health` | Server and database health check      |