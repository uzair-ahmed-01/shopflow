# Learning Notes

This document acts as a log for backend engineering takeaways from building ShopFlow.

## Clean Architecture

- Delivery layer (Handlers) should not leak any database details or models.
- Service layer contains all business logic and controls transactions.
- Repository layer abstracts database interaction.

### Handlers Security & Generic Request Decoding

In production APIs, decoding user JSON input directly is unsafe and prone to boilerplate duplication.
1. **Generic JSON Decoders (`DecodeJSON[T]`)**: Utilizing Go generics allows writing a single, reusable request decoder, reducing boilerplate error checking by 70% in HTTP handlers.
2. **Payload Size Limitation (`http.MaxBytesReader`)**: Restricting the maximum request body size (e.g., 1MB) prevents memory exhaustion Denial of Service (DoS) attacks where clients upload huge files to crash the server.
3. **Strict Validation (`DisallowUnknownFields()`)**: Instructing the JSON decoder to reject unknown properties stops silent data loss and typing mistakes (e.g., client sends `pricee` instead of `price`, causing the app to silently process the value as 0 without warning).

## Concurrency and Worker Pools

### 1. Why a Worker Pool?
In backend services, processing background tasks (like updating statuses, sending emails) directly inside incoming HTTP requests slows down API responses and overloads server resources.
A **Worker Pool** decouples job ingestion from task execution:
- **Dispatcher (Producer)**: Runs on a tick (using `time.Ticker`), retrieves tasks (`PENDING` orders), and sends task IDs into a buffered channel.
- **Worker Pool (Consumers)**: A fixed number of background goroutines (e.g., 3 workers) constantly read task IDs from the channel and process them concurrently.
- **Benefits**: Limits resource utilization (controls how many database queries run simultaneously), improves API latency.

### 2. Core Go Concurrency Concepts Used
- **Goroutines**: Lightweight execution threads. Spawning workers using `go worker()` takes minimal memory (~2KB starting size) compared to OS threads.
- **Channels**: Safe communication pipes.
  - `jobChan chan int`: Buffered job queue. Workers read from it safely without explicit locks (`sync.Mutex`), because Go handles channel concurrency internally.
  - `stopChan chan struct{}`: Broadcasts shutdown signal. Closing this channel instantly unblocks all listeners reading from it.
- **sync.WaitGroup**: Synchronization tool. Counts active worker goroutines. Main thread calls `wg.Wait()` during shutdown to prevent exiting before background workers finish active tasks.
- **context.Context**: Passes deadlines and cancellation signals across boundaries (middleware -> handlers -> services -> repositories).

### 3. Graceful Shutdown Flow
Graceful shutdown prevents data corruption by ensuring active tasks finish before the process terminates:
1. Server receives termination signal (`SIGINT` or `SIGTERM`).
2. Main thread calls `processor.Stop()`.
3. Stop channel (`stopChan`) is closed, causing the dispatcher loop to exit.
4. Job channel (`jobChan`) is closed. Workers finish processing any remaining items in the buffer.
5. Once the job channel is drained, workers exit, decrementing the `WaitGroup`.
6. Main thread finishes `wg.Wait()` and closes DB connections cleanly.

## JWT Refresh Token Lifecycle & Revocation

### 1. The Need for Refresh Tokens
JWT access tokens are stateless and self-contained. Once signed, they cannot be easily revoked before expiration. 
- **Security Strategy**: We shorten the access token lifetime (e.g., 15 minutes) and issue a long-lived **Refresh Token** (e.g., 7 days) persisted in the database.
- **Revocation**: If a user logs out or a token is compromised, the refresh token is marked as `revoked_at` in the database, locking out further access token renewals.

### 2. Token Rotation (Replay Protection)
To prevent stolen refresh tokens from being reused indefinitely:
- Every time a client requests a new access token via `POST /api/v1/auth/refresh`, the old refresh token is **invalidated (revoked)** and a **new refresh token** is generated and returned (rotated).
- **Concurrency Protection**: We use atomic database updates (`SET revoked_at = NOW() WHERE token = ? AND revoked_at IS NULL`) so if multiple simultaneous requests attempt to reuse the same refresh token, only the first one succeeds.

### 3. Session Revocation (Logout)
- `POST /api/v1/auth/logout` invalidates the active refresh token session by writing the `revoked_at` timestamp. Future refresh calls using that token will fail.

## Role-Based Access Control (RBAC)

### 1. Authentication vs. Authorization
- **Authentication**: Establishes *who* the user is (JWT validation in `AuthMiddleware`).
- **Authorization**: Establishes *what* the user is allowed to do (RBAC verification in `RequireRole` middleware).

### 2. JWT Role Claims
- The user's role (e.g. `customer` or `admin`) is added to the database `users` table and encoded directly inside the JWT access token claims. This avoids querying the database on every HTTP request to resolve permissions, maintaining high performance.

### 3. Middleware Chaining
To secure administrative operations without bloating route handlers, we wrap handlers using a functional middleware chain:
```go
router.Handle("POST /api/v1/products", authMiddleware(adminMiddleware(http.HandlerFunc(handler))))
```
- **Execution Flow**:
  1. `AuthMiddleware` verifies JWT and injects `AuthUser` struct into context.
  2. `RequireRole("admin")` retrieves `AuthUser` from context, reads the role, and returns `403 Forbidden` if role constraints are not met.
  3. Target handler executes safely.

## Structured Logging with Zerolog

### 1. Plain Text vs. Structured Logging
- **Plain Text Logging (`log.Printf`)**: Hard to parse or query. Good for human reading in development, bad for machines/aggregators in production.
- **Structured Logging (`zerolog`)**: Prints key-value logs formatted as JSON. Allows log aggregators (ELK, CloudWatch, Datadog) to instantly search, filter, and alert based on attributes (e.g. status code, execution duration).

### 2. Capturing Response Status Codes in Go Middleware
Go's standard `http.ResponseWriter` interface does not expose a method to read the HTTP status code after it has been written. To log response status codes:
- We create a custom `responseWriter` wrapper struct that implements `http.ResponseWriter`.
- We override the `WriteHeader(code int)` method to save the status code to a field before delegating to the original response writer.
- This allows our logging middleware to inspect the status code *after* downstream handlers have executed.

### 3. Zerolog Contextual Logging
- Logging context fields:
  ```go
  log.Info().Str("method", r.Method).Int("status", statusCode).Msg("HTTP Request")
  ```
- This guarantees structured fields are indexed separately from the human-readable text message, providing powerful log search capabilities.

## Redis Caching

- TBD

## PostgreSQL & Event Consistency

- TBD
