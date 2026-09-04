APP_NAME=gadget-store-api

.PHONY: run build test fmt tidy clean
APP_NAME=gadget-store-api

.PHONY: run build test fmt tidy clean

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP_NAME) ./cmd/api

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

clean:
	rm -rf bin