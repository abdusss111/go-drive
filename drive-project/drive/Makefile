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

.PHONY: docker-up
docker-up:
	docker-compose up -d

.PHONY: docker-down
docker-down:
	docker-compose down

.PHONY: docker-logs
docker-logs:
	docker-compose logs -f

.PHONY: setup
setup: docker-up
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	@GODRIVE_DATABASE_URL="postgres://godrive_app:change-me@localhost:5432/godrive?sslmode=disable" make migrate-up
	@echo "Setup complete! Services are running."
	@echo "PostgreSQL: localhost:5432"
	@echo "MinIO API: localhost:9000"
	@echo "MinIO Console: http://localhost:9001"
	@echo "Run 'make run' to start the API server"

