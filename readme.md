# Uptime API

Backend for a website and API uptime monitoring service, with downtime alerts and public status pages.

Frontend: _(link to Angular repository, coming soon)_

## Tech stack

- Go (standard library `net/http`)

## Getting started

Requirements: Go 1.22+

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