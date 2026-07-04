# Go Concurrency Guide (Interview-Ready Study Notes)

Yeh guide Go Concurrency ko basic concepts se lekar advanced patterns tak cover karti hai.

---

## 1. Concurrency vs Parallelism

- **Concurrency**: Ek se zyada tasks ko structure karna jo overlapping time periods par chal sakein (multiplexing on single CPU core). E.g., Web server client requests listen bhi kar raha hai aur background database cleanup bhi chala raha hai.
- **Parallelism**: Ek se zyada tasks ko actual same physical time par multiple CPU cores par parallel run karna.
- **Go Philosophy**: *"Concurrently design karein taake parallel execution automatically handle ho sake."*

---

## 2. Goroutines

Goroutines lightweight execution threads hain jo Go runtime manage karta hai (OS threads nahi).

### Key Characteristics:
1. **Memory size**: OS thread ~1MB se start hota hai, jabke Goroutine sirf **~2KB** size se start hoti hai.
2. **Context switching**: OS thread switching costly hoti hai. Goroutines scheduling internal Go Runtime Scheduler (GMP model) handle karta hai jo bohot fast hai.
3. **Syntax**: Kisi bhi function ke aage `go` keyword lagane se new goroutine spawn ho jati hai:

```go
package main

import (
	"fmt"
	"time"
)

func sayHello() {
	fmt.Println("Hello from Goroutine!")
}

func main() {
	go sayHello() // asynchronous start
	time.Sleep(100 * time.Millisecond) // main exit hone se rokne ke liye
}
```

---

## 3. Channels (Communication Pipes)

Goroutines ke beech safely data share karne ke liye Channels use hote hain.
Go Rule: *"Do not communicate by sharing memory; instead, share memory by communicating."*

### Types of Channels:
1. **Unbuffered Channel** (Default):
   - Capacity = 0.
   - Sending thread block rehta hai jab tak receiving thread ready na ho.
   ```go
   ch := make(chan int) // blocks on send/receive
   ```
2. **Buffered Channel**:
   - Capacity > 0.
   - Sending thread tab tak block nahi hota jab tak buffer full na ho.
   ```go
   ch := make(chan int, 100) // holds 100 items before blocking
   ```

### Channel Directionality:
- **`chan T`**: Bidirectional channel (Read/Write both).
- **`<-chan T`**: Receive-only channel.
- **`chan<- T`**: Send-only channel.

---

## 4. Synchronization (sync.WaitGroup & sync.Mutex)

Multiple goroutines ki coordination ke liye.

### A. WaitGroup
Jab aapko background tasks ke complete hone ka wait karna ho tab `sync.WaitGroup` use hota hai:

- `wg.Add(n)`: Counters increase karta hai (total tasks).
- `wg.Done()`: Counter decrement karta hai (task complete).
- `wg.Wait()`: execution block karta hai jab tak counter 0 na ho.

```go
var wg sync.WaitGroup
for i := 1; i <= 3; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        fmt.Printf("Worker %d done\n", id)
    }(i)
}
wg.Wait() // wait until all 3 workers call Done
```

### B. Mutex (Mutual Exclusion)
Shared memory read/write safe karne ke liye variable ko lock karna:

```go
var mu sync.Mutex
var count int

func increment() {
    mu.Lock()
    count++ // only one goroutine can modify count at a time
    mu.Unlock()
}
```

---

## 5. Select Statement

Select statement code block multi-channel coordination ke liye use hota hai. Jo channel pehle ready hoga, select wahi case run karega:

```go
select {
case msg1 := <-ch1:
    fmt.Println("Received msg1:", msg1)
case ch2 <- msg2:
    fmt.Println("Sent msg2")
case <-time.After(1 * time.Second):
    fmt.Println("Timeout reached!")
default:
    fmt.Println("No channels ready (non-blocking case)")
}
```

---

## 6. Channel Close and Range Loop

- Channel ko close karne se downstream receivers ko signal mil jata hai ke transmission end ho gaya hai.
- Closed channel se read karna non-blocking zero value return karta hai.
- **Range Loop**: Loop automatically terminate ho jata hai jab channel close hota hai:

```go
ch := make(chan int, 3)
ch <- 10
ch <- 20
close(ch) // signals range loop to end

for val := range ch {
    fmt.Println(val) // prints 10, then 20, then exits loop safely
}
```

---

## 7. Graceful Shutdown Design Pattern

Production backend code shutdown signal receive karte hi database safe status update aur tasks cleanup complete karta hai:

1. **`stopChan := make(chan struct{})`**: Broadcast shutdown signal.
2. **`close(stopChan)`**: Jo goroutines `<-stopChan` receive kar rahi hain, wo automatically return ho jati hain.
3. **`wg.Wait()`**: Loop exits safely, resources cleanly released before `main.go` terminates.
