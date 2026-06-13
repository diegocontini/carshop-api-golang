.PHONY: run build test test-unit test-integration sqlc tidy fmt lint migrate-up migrate-down migrate-create migrate-status docker-up docker-down

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

migrate-status:
	goose -dir migrations postgres "$$DATABASE_URL" status

# Usage: make migrate-create name=add_index_orders
migrate-create:
	@test -n "$(name)" || (echo "missing name: make migrate-create name=<slug>" && exit 1)
	goose -dir migrations create $(name) sql

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
