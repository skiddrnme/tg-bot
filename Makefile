.PHONY: run build test lint

run:
	air

build:
	go build -o bot cmd/bot/main.go

test:
	go test -cover ./...

lint:
	golangci-lint run