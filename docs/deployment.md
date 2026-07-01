# Deployment & Execution Guide

This document describes how to deploy and run ShopFlow locally or in containers.

## Prerequisites

- Go 1.22+
- PostgreSQL (running locally on your host machine for local dev)
- Redis (optional, disabled for now)
- Docker and Docker Compose (recommended)

## Run Locally (Without Docker)

1. **Set Environment Variables**:
   Create a `.env` file in the root directory:

   ```env
   PORT=8080
   DB_HOST=127.0.0.1
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=shopflow
   DB_SSLMODE=disable
   # REDIS_ADDR=127.0.0.1:6379 # Redis disabled for now
   JWT_SECRET=supersecretchangeinprod
   ```

2. **Run Postgres**:
   Ensure a local installation of Postgres is running.

3. **Start the API**:

   ```bash
   go run cmd/api/main.go
   ```

## Database Migrations

Migration files are stored in the `/migrations` directory.

### Running Migrations Manually
Use the `migrate` CLI tool (golang-migrate) to apply migrations:

*   **Apply Up Migrations**:
    ```bash
    migrate -path migrations -database "postgres://postgres:postgres@127.0.0.1:5432/shopflow?sslmode=disable" up
    ```

*   **Rollback Down Migrations**:
    ```bash
    migrate -path migrations -database "postgres://postgres:postgres@127.0.0.1:5432/shopflow?sslmode=disable" down
    ```


## Run with Docker Compose

1. **Start all services**:

   ```bash
   docker-compose up --build
   ```

   This compiles the Go application, pulls Postgres image, spins it up, and sets up network linkage automatically.
