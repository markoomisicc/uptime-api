-include .env
export

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

tools:
	@go install github.com/air-verse/air@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

setup:
	@test -f .env || cp .env.example .env
	@docker compose up -d --wait
	@$(MAKE) migrate-up

dev:
	@air

migrate-up:
	@migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	@migrate -path migrations -database "$(DB_URL)" down 1

migrate-create:
	@migrate create -ext sql -dir migrations -seq $(name)

generate:
	@sqlc generate