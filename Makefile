.PHONY: run build test lint docker-up docker-down

run:
	go run ./cmd/bot

build:
	go build -o bin/bot ./cmd/bot

test:
	go test -cover ./...

lint:
	golangci-lint run ./...

docker-up:
	docker compose up --build -d
	docker compose logs -f app

docker-down:
	docker compose down