.PHONY: build run run-worker test clean docker-up docker-down migrate-up migrate-down swagger-gen

# Variables
BINARY_NAME=shopflow
WORKER_BINARY_NAME=shopflow-worker
MIGRATE_CMD=migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/shopflow?sslmode=disable"

build:
	@echo "Building API binary..."
	go build -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "Building Worker binary..."
	go build -o bin/$(WORKER_BINARY_NAME) cmd/worker/main.go

run:
	@echo "Running API application..."
	go run cmd/api/main.go

run-worker:
	@echo "Running background worker..."
	go run cmd/worker/main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f $(BINARY_NAME) $(WORKER_BINARY_NAME)

docker-up:
	@echo "Starting Docker services..."
	docker-compose up --build -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

migrate-up:
	@echo "Running database migrations UP..."
	$(MIGRATE_CMD) up

migrate-down:
	@echo "Running database migrations DOWN..."
	$(MIGRATE_CMD) down

swagger-gen:
	@echo "Generating Swagger documentation..."
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go
