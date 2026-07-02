# Database Migrations Guide

Guide for managing and running database migrations in ShopFlow.

## Option 1: Direct SQL Execution (Using pgAdmin 4)

If you do not have migration CLI tools installed locally, execute SQL manually.

1. Open **pgAdmin 4**.
2. Connect to local server and select/create `shopflow` database.
3. Right-click `shopflow` database, click **Query Tool**.
4. Open [000001_init_schema.up.sql](file:///e:/Practice%20Area/GoBackend/Project/shopflow/migrations/000001_init_schema.up.sql) file.
5. Copy entire SQL text.
6. Paste into pgAdmin Query Tool editor.
7. Click **Execute** (or press `F5`).

## Option 2: golang-migrate CLI

Using standard `golang-migrate` utility.

### 1. Installation

- **Windows (Chocolatey)**:
  ```bash
  choco install golang-migrate
  ```
- **Windows (Scoop)**:
  ```bash
  scoop install golang-migrate
  ```
- **macOS (Homebrew)**:
  ```bash
  brew install golang-migrate
  ```

### 2. Running Migrations

Run migrations UP:
```bash
migrate -path migrations -database "postgres://<DB_USER>:<DB_PASSWORD>@<DB_HOST>:<DB_PORT>/<DB_NAME>?sslmode=disable" up
```
For example, for local default setups (using `.env` credentials):
```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/shopflow?sslmode=disable" up
```

Run migrations DOWN (roll back):
```bash
migrate -path migrations -database "postgres://<DB_USER>:<DB_PASSWORD>@<DB_HOST>:<DB_PORT>/<DB_NAME>?sslmode=disable" down
```

### 3. Creating New Migrations
To generate new migration files:
```bash
migrate create -ext sql -dir migrations -seq <migration_name>
```
