.PHONY: run build test test-unit test-integration sqlc tidy fmt lint migrate-up migrate-down docker-up docker-down

run:
	go run ./server

build:
	go build -o bin/server ./server

test:
	go test ./...

test-unit:
	go test ./tests/unit/... ./src/...

test-integration:
	go test -tags=integration ./tests/integration/...

sqlc:
	sqlc generate

tidy:
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run

migrate-up:
	goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	goose -dir migrations postgres "$$DATABASE_URL" down

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
