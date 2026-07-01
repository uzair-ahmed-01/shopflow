# Project Requirements

This document defines the detailed functional and non-functional requirements for ShopFlow.

## Functional Requirements

### 1. Authentication & User Management
- **User Registration**:
  - Inputs: `name` (required, 2-100 characters), `email` (required, valid format, unique), `password` (required, minimum 8 characters).
  - Security: Passwords must be hashed using bcrypt before database insertion.
- **User Login**:
  - Inputs: `email`, `password`.
  - Output: Returns a JWT containing user ID and email on successful authentication.
  - Session Security: JWT signature verified using HS256 and a secret key loaded from env. Token expires after 24 hours.
- **Protected Routes**:
  - Middleware intercepts requests, extracts Bearer token, validates it, and injects user identity context.

### 2. Category Module
- **Create Category**:
  - Inputs: `name` (required, unique, 2-50 characters), `description` (optional).
  - Access Control: Requires authentication.
- **List Categories**:
  - Returns a list of all categories. Public access.

### 3. Product Module
- **Create Product**:
  - Inputs: `name` (required, 2-100 characters), `description` (optional), `price` (required, positive integer representing cents), `stock` (required, non-negative integer), `category_id` (required, must point to valid Category).
  - Access Control: Requires authentication.
- **Update Product**:
  - Inputs: `name`, `description`, `price`, `stock`, `category_id` (all optional, updates fields provided).
  - Access Control: Requires authentication.
- **Delete Product**:
  - Access Control: Requires authentication. Hard delete.
- **List Products**:
  - Inputs: Query params `page` (default 1) and `limit` (default 10, max 100).
  - Performance: Fetch from Redis cache first. If cache miss, fetch from DB and write to cache.
  - Invalidation: Cache is cleared on product create, update, or delete.

### 4. Cart Module
- **Add Product to Cart**:
  - Inputs: `product_id` (must exist), `quantity` (positive integer).
  - Rules: Adds item to user's cart. If item already in cart, adds quantity. Returns updated cart.
- **Remove Product from Cart**:
  - Inputs: `product_id`.
  - Output: Removes the product from the user's cart entirely.
- **View Cart**:
  - Output: List of cart items with product details (id, name, price, quantity) and cart `total` sum in cents.

### 5. Order Module
- **Place Order**:
  - Rules:
    - Converts user's current cart items to an order.
    - Runs in a single database transaction:
      1. Check stock for each product. If stock < quantity, rollback transaction and fail order.
      2. Decrement product stock.
      3. Create order record with state `PENDING`.
      4. Create order item records capturing the price at purchase.
      5. Empty user's cart.
    - Emits an `OrderCreated` event to the internal channel for background processing.
- **View Orders**:
  - Returns historical orders placed by the authenticated user.
- **View Order Details**:
  - Returns full details of a specific order including items, purchase prices, status, and total.

### 6. Event-Driven Background Processing
- **Asynchronous Worker Pool**:
  - Listens to internal channel for events.
  - When an `OrderCreated` event is received:
    1. Simulates payment capture (logs progress).
    2. Updates order status to `PAID` (or `COMPLETED` on success).
    3. Triggers mock notification/email job (simulating SMTP latency via sleep).
  - Pool capacity: Set to 3 worker goroutines running concurrently.

---

## Non-Functional Requirements

### 1. Caching & State Management
- **Redis Cache-Aside Pattern**:
  - Caches paginated product lists.
  - Cache TTL is set to 10 minutes.
  - Any product creation, update, or deletion must invalidate the cached lists to prevent stale reads.

### 2. Robust Transactions
- **ACID Order Processing**:
  - Strict transaction isolation to prevent race conditions during concurrent stock deduction.

### 3. Graceful Shutdown
- **OS Signal Handling**:
  - Captures `SIGINT` (Ctrl+C) and `SIGTERM`.
  - Allows HTTP server to drain active requests (5 seconds timeout).
  - Stops background workers gracefully by closing event channels and letting active jobs finish.
