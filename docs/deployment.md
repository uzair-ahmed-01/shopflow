# Deployment & Execution Guide

This document describes how to deploy and run ShopFlow locally or in containers.

## Prerequisites
- Go 1.22+
- PostgreSQL
- Redis
- Docker and Docker Compose (recommended)

## Run Locally (Without Docker)

1. **Set Environment Variables**:
   Create a `.env` file in the root directory:
   ```env
   PORT=8080
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=shopflow
   REDIS_ADDR=localhost:6379
   JWT_SECRET=supersecretchangeinprod
   ```

2. **Run Postgres & Redis**:
   Ensure local installations of Postgres and Redis are running.

3. **Start the API**:
   ```bash
   go run cmd/api/main.go
   ```

## Run with Docker Compose

1. **Start all services**:
   ```bash
   docker-compose up --build
   ```
   This compiles the Go application, pulls Postgres and Redis images, spins them up, and sets up network linkage automatically.
