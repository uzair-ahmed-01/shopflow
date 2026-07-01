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

- `/cmd/api/main.go`: Application entrypoint. Sets up dependency injection, database connection, Redis client, routes, and starts HTTP server.
- `/migrations/`: SQL migration files for database schema versioning.
- `/internal/`: Contains core business logic and database entities:
  - `/internal/models/`: Shared domain model structs (User, Product, Category, etc.) to prevent circular imports.
  - `/internal/auth/`: Service, repository, and handler for users.
  - `/internal/product/`: Service, repository, and handler for products.
  - `/internal/category/`: Service, repository, and handler for categories.
  - `/internal/cart/`: Service, repository, and handler for carts.
  - `/internal/order/`: Service, repository, and handler for orders.
- `/internal/db/`: DB connection pools, migrations, helper functions.
- `/internal/middleware/`: JWT authentication validation, logging, recover middlewares.
- `/internal/cache/`: Redis client and caching wrappers.
- `/internal/events/`: Event models and event bus/channel implementation.
- `/internal/worker/`: Worker pool implementation to process async background tasks.
