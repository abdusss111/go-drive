GO_FILES := $(shell find . -name '*.go' -not -path "./vendor/*")

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database "$${GODRIVE_DATABASE_URL}" up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database "$${GODRIVE_DATABASE_URL}" down 1

.PHONY: run
run:
	go run ./cmd/api

