# Uptime API

Backend for a website and API uptime monitoring service, with downtime alerts and public status pages.

Frontend: _(link to Angular repository, coming soon)_

## Tech stack

- Go (standard library `net/http`)
- PostgreSQL 18
- Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations
- [sqlc](https://sqlc.dev) for type-safe database queries

## Getting started

Requirements: Go 1.22+, Docker, `make`

```bash
make tools   # install development tools (once)
make setup   # create .env, start the database and run migrations
make dev     # run the API with live reload
```

The API runs on `http://localhost:8080`. If port 5432 is already in use, change `DB_PORT` in `.env` and run `make setup` again.

## Commands

| Command | Description |
|---|---|
| `make tools` | Install Air, golang-migrate and sqlc |
| `make setup` | Create `.env`, start the database and apply migrations |
| `make dev` | Run the API with live reload |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back the last migration |
| `make migrate-create name=<name>` | Create a new migration |
| `make generate` | Generate Go code from SQL queries |

Migrations that have been merged into `main` must never be edited. Create a new migration instead.

## Endpoints

| Method | Path      | Description                      |
|--------|-----------|----------------------------------|
| GET    | `/health` | Server and database health check |