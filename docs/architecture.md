# Architecture Design

This document details the clean architecture layout and patterns used in ShopFlow.

## Component Layout

ShopFlow uses standard Clean Architecture:

```
                       +-------------------+
                       |    HTTP Handlers  | (Delivery Layer)
                       +---------+---------+
                                 |
                                 v
                       +---------+---------+
                       |  Service Interfaces | (Use Cases / Business Logic)
                       +---------+---------+
                                 |
                                 v
        +------------------------+------------------------+
        |                                                 |
        v                                                 v
+-------+-----------+                             +-------+-----------+
| Repository Interfaces | (Data Persistence)     |   Workers / Events  | (Infrastructure/Shared)
+-------+-----------+                             +-------+-----------+
        |                                                 |
        v                                                 v
+-------+-----------+                             +-------+-----------+
| PostgreSQL / SQL  |                             | Redis Cache / Go  |
+-------------------+                             +-------------------+
```

## Directory Mapping

- `/cmd/api/main.go`: Application entrypoint. Bootstrap database connection pool, starts background workers, initializes dependency injection, and starts HTTP server. Delegates API routing mapping to `internal/handler/routes.go`.
- `/migrations/`: SQL migration files for database schema versioning.
- `/internal/`: Contains core layers and database entities:
  - `/internal/models/`: Shared domain model structs (User, Product, Category, etc.).
  - `/internal/repository/`: Shared database repositories (UserRepository, etc.).
  - `/internal/service/`: Business logic services (AuthService, etc.).
  - `/internal/handler/`: HTTP handlers (AuthHandler, etc.) and `routes.go` defining REST API endpoint patterns and middleware chain bindings.
- `/internal/db/`: DB connection pools, migrations, helper functions.
- `/internal/middleware/`: JWT authentication validation, logging, recover middlewares.
- `/internal/cache/`: Redis client and caching wrappers.
- `/internal/events/`: Event models and event bus/channel implementation.
- `/internal/worker/`: Worker pool implementation to process async background tasks.
