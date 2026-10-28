# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Health check endpoint (`GET /health`)
- Live reload for local development using Air
- PostgreSQL database setup with Docker Compose
- Database connection pool with configuration loaded from environment variables and `.env`
- Database status in health check response
- Database migrations with golang-migrate
- Initial schema with users, organizations, roles and organization membership
- Type-safe database queries generated with sqlc
- User registration endpoint (`POST /auth/register`)
- Session-based authentication with login, logout and current user endpoints
- User language preference with support for English and German
- Endpoint for updating the current user (`PATCH /auth/me`)
- Email sending via SMTP, with Mailpit for local development
- Email verification with a link sent after registration
- Endpoint for resending the verification email, limited to once per minute per address

### Changed

- Moved HTTP handlers into the `internal/api` package with shared JSON helpers
- Simplified local development setup with `make tools`, `make setup` and `make dev`
- API errors now include a machine-readable code alongside the message
- Login requires a verified email address
