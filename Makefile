.PHONY: run build test lint migrate-up migrate-down docker-up docker-down swag

# ── Локальная разработка ──────────────────────────────────────────────────────

run: swag
	go run ./cmd/api/main.go

build: swag
	go build -o ./bin/server ./cmd/api/main.go

test:
	go test ./internal/... -v -race

lint:
	golangci-lint run ./...

# ── Swagger ───────────────────────────────────────────────────────────────────

swag:
	swag init -g cmd/api/main.go -o docs

# ── Docker ────────────────────────────────────────────────────────────────────

docker-up: swag
	docker-compose up --build

docker-down:
	docker-compose down

docker-db:
	docker-compose up db -d

# ── Миграции (локально, требует запущенной БД) ────────────────────────────────

migrate-up:
	goose -dir ./migrations postgres "$(DB_DSN)" up

migrate-down:
	goose -dir ./migrations postgres "$(DB_DSN)" down

# ── Утилиты ───────────────────────────────────────────────────────────────────

tidy:
	go mod tidy

.DEFAULT_GOAL := run