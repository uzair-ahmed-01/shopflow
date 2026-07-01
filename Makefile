.PHONY: build run test clean docker-up docker-down migrate-up migrate-down

# Variables
BINARY_NAME=shopflow
MIGRATE_CMD=migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/shopflow?sslmode=disable"

build:
	@echo "Building binary..."
	go build -o bin/$(BINARY_NAME) cmd/api/main.go

run:
	@echo "Running application..."
	go run cmd/api/main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f $(BINARY_NAME)

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
