.PHONY: build run

fmt:
	go fmt ./...

build:
	@go build -o bin/api ./cmd/api

run: build
	@./bin/api

test:
	go test ./...

migrate-up:
	@go run ./cmd/migrate up

migrate-down:
	@go run ./cmd/migrate down

tidy:
	go mod tidy

clean:
	rm -rf bin