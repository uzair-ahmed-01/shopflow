# Learning Notes

This document acts as a log for backend engineering takeaways from building ShopFlow.

## Clean Architecture

- Delivery layer (Handlers) should not leak any database details or models.
- Service layer contains all business logic and controls transactions.
- Repository layer abstracts database interaction.

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

## Redis Caching

- TBD

## PostgreSQL & Event Consistency

- TBD
