# Load .env file (optional)
-include .env

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: run build clean migrate-up migrate-down migrate-create migrate-install seed

run:
	./bin/server

build:
	go build -o bin/server cmd/server/main.go

clean:
	if exist bin\server del bin\server

migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "❌ Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations $(name)

migrate-version:
	migrate -path migrations -database "$(DB_URL)" version

migrate-reset:
	migrate -path migrations -database "$(DB_URL)" drop -f
	migrate -path migrations -database "$(DB_URL)" up

seed:
	@echo "Seeding is handled via migrations. Run 'make migrate-up'."

swagger:
	swag init -g cmd/server/main.go -o docs --parseDependency --parseDepth 2

test:
	go test ./...

test-cover:
	go test -cover ./...

test-verbose:
	go test -v ./...