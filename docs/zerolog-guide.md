# Structured Logging Guide (Zerolog)

Structured logging is a key backend engineering concept. This guide explains how structured logging is designed and implemented in **ShopFlow** using the `rs/zerolog` package.

---

## 1. Plain Text vs. Structured Logging

### Plain Text Logging (`log.Printf`)
Traditional logs print plain text lines:
```text
2026/07/05 14:00:00 [DISPATCHER] Processing order ID 101: PENDING -> IN_PROGRESS...
```
*   **Limitation**: Hard for search query systems (ELK, Datadog) to parse or filter. Searching for all failed operations or filtering logs by `user_id = 45` requires expensive regex searches.

### Structured Logging (JSON)
Structured loggers output key-value context in machine-readable JSON format:
```json
{"level":"info","worker_id":2,"order_id":101,"time":1710000000,"message":"[WORKER] Processing order ID: PENDING -> IN_PROGRESS..."}
```
*   **Advantage**: Log collectors parse the JSON keys instantly. You can query `worker_id:2 AND level:error` with zero overhead.

---

## 2. Why Zerolog?

`rs/zerolog` is a popular logging package in Go due to its **zero-allocation** architecture. 
- It formats JSON logs directly into a byte buffer without creating temporary objects on the heap.
- This minimizes Garbage Collection (GC) pauses, keeping high-performance APIs fast.

### Zerolog API Basics
Logs are constructed using a builder pattern:
```go
log.Info().Str("database", "postgres").Int("max_conns", 25).Msg("Database connection pool configured")
```
- **Levels**: `.Debug()`, `.Info()`, `.Warn()`, `.Error()`, `.Fatal()`, `.Panic()`
- **Context Fields**: `.Str()`, `.Int()`, `.Bool()`, `.Err()` (attaches full stack/error messages)

---

## 3. ShopFlow Logging Architecture

ShopFlow implements structured logging across the entire HTTP lifecycle, background workers, and central error helper boundaries:

```
                  +--------------------------------+
                  |  HTTP Request (Client)         |
                  +---------------+----------------+
                                  |
                                  v
                  +---------------+----------------+
                  |  RequestLoggerMiddleware       |  <-- Logs incoming IP, Method, URI
                  +---------------+----------------+
                                  |
                                  v
                  +---------------+----------------+
                  |  HTTP Request Handlers         |
                  +---------------+----------------+
                                  | (Internal Error)
                                  v
                  +---------------+----------------+
                  |  Central SendError Helper      |  <-- Logs raw DB/System errors to Zerolog
                  +--------------------------------+
```

### Pattern A: Centralized HTTP Error Logging (Variadic `errs ...error`)
To avoid writing duplicate `log.Error().Err(err).Msg(...)` inside every handler, ShopFlow implements a centralized error handler inside [response.go](file:///e:/Practice%20Area/GoBackend/Project/shopflow/internal/handler/response.go):

```go
func SendError(w http.ResponseWriter, status int, message string, code string, errs ...error) {
	// 1. If an internal error is passed, log it to Zerolog
	if len(errs) > 0 && errs[0] != nil {
		log.Error().Err(errs[0]).Str("code", code).Int("status", status).Msg(message)
	} else if status >= 500 {
		log.Error().Str("code", code).Int("status", status).Msg(message)
	}

	// 2. Respond to the client with a safe, generic error message (prevents details leak)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]string{
			"message": message,
			"code":    code,
		},
	})
}
```
*   **Why Variadic?**: Declaring `errs ...error` is backward-compatible. Calls with only 4 arguments (e.g. standard validation errors like `SendError(w, status, msg, code)`) compile without edits. If a database failure occurs, we pass the 5th parameter `err` to log it.

---

### Pattern B: Intercepting Status Codes in Middleware
Go's native `http.ResponseWriter` interface does not allow reading the HTTP status code after it has been sent. To log response metrics:
1. We wrap the response writer in a custom struct that captures the code:
   ```go
   type responseWriter struct {
       http.ResponseWriter
       statusCode int
   }

   func (rw *responseWriter) WriteHeader(code int) {
       rw.statusCode = code
       rw.ResponseWriter.WriteHeader(code)
   }
   ```
2. We log the details in the middleware after calling `next.ServeHTTP()`:
   ```go
   func RequestLoggerMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           start := time.Now()
           rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

           next.ServeHTTP(rw, r)

           duration := time.Since(start)

           log.Info().
               Str("method", r.Method).
               Str("path", r.URL.Path).
               Int("status", rw.statusCode).
               Float64("duration_ms", float64(duration.Microseconds())/1000.0).
               Msg("HTTP Request")
       })
   }
   ```

---

### Pattern C: Background Worker Instrumentation
Inside [order_processor.go](file:///e:/Practice%20Area/GoBackend/Project/shopflow/internal/worker/order_processor.go), the dispatcher and worker pools use structured logging to report processing context:
```go
log.Info().Int("worker_id", workerID).Int("order_id", orderID).Msg("[WORKER] Processing order ID: PENDING -> IN_PROGRESS...")
```
- Lets monitor system health (e.g. dispatcher throughput, worker job queue delays, execution limits).
