APP_NAME := user-service
MAIN := ./main.go

.PHONY: all build run test lint migrate-up migrate-down clean

all: build

build:
	go build -o bin/$(APP_NAME) $(MAIN)

run:
	go run $(MAIN) start

test:
	go test ./... -v

lint:
	golangci-lint run ./...

migrate-up:
	goose -dir database/migrations up

migrate-down:
	goose -dir database/migrations down

clean:
	rm -rf bin/
