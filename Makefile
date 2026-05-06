.PHONY: build run test lint clean docker-up docker-down migrate

BINARY=quizarena
CMD=./cmd/server

build:
	go build -o bin/$(BINARY) $(CMD)

run:
	go run $(CMD)/main.go

test:
	go test ./... -v -race -count=1

test-unit:
	go test ./tests/unit/... -v -race

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	@echo "Migrations are applied automatically on server start."

tidy:
	go mod tidy

.DEFAULT_GOAL := build
