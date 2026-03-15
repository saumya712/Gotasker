# GoTasker 🚀

A concurrent task queue backend built in Go — designed to learn and practice real-world concurrency patterns from scratch.

---

## What It Does

Users POST an array of tasks → 3 workers process them concurrently in the background → task status can be checked in real time via GET.

---

## Project Structure

```
practice/
├── cmd/server/
│   └── main.go          # Entry point, server setup, graceful shutdown
├── handler/
│   └── handler.go       # HTTP handlers — orchestrates logic layer
├── domain/
│   └── task.go          # Core entities: Task struct, Status type, constants
├── logic/
│   └── logic.go         # Business logic: TaskStore, worker pool, processing
└── go.mod
```

---

## Architecture

### Clean Layered Design
- **Domain** → Core entities (`Task` struct, `Status` type). Zero external dependencies.
- **Logic** → Business logic: `TaskStore`, worker pool, task processing.
- **Handler** → HTTP layer. Binds JSON, calls logic, returns responses.
- **Main** → Wires everything together. Owns the server lifecycle.

Handlers never contain business logic. Logic never touches HTTP.

---

## Concurrency Model

```
handler
  └── go Processtasks()          ← goroutine (Parent, tracked by Store.Wg)
        ├── go workerpool()      ← worker 1  (tracked by localwg)
        ├── go workerpool()      ← worker 2  (tracked by localwg)
        └── go workerpool()      ← worker 3  (tracked by localwg)
```

### Two WaitGroups for two levels of concurrency

| WaitGroup | Tracks | Purpose |
|-----------|--------|---------|
| `localwg` | 3 worker goroutines | Waits for all workers to drain the job queue |
| `Store.Wg` | `Processtasks` goroutine | Tells `main` when all processing is fully done |

`localwg.Wait()` blocks `Processtasks` until all workers finish → only then `defer Store.Wg.Done()` fires → only then `main` unblocks from `Store.Wg.Wait()` and exits cleanly.

### Buffered Job Queue
```go
jobqueue := make(chan *domain.Task, len(tasks))
```
Capacity equals number of tasks — pushing all tasks never blocks. Workers drain it concurrently at their own pace.

---

## Thread-Safe Task Store

```go
type taskstore struct {
    tasks  map[int]*domain.Task
    mu     sync.Mutex
    nextid int
    Wg     sync.WaitGroup
}
```

All reads and writes to the task map are protected by `sync.Mutex`. Multiple goroutines can safely update task statuses without data races.

---

## Task Lifecycle

```
POST /posttasks
  → Addtasks() — tasks added to store immediately (Pending)
  → go Processtasks() — workers pick up tasks in background
  → 200 response — HTTP doesn't wait for processing

Worker picks up task:
  Pending → Processing → Done
```

---

## Graceful Shutdown

On `Ctrl+C` (SIGINT/SIGTERM):

```
1. server.Shutdown(ctx)   — stop accepting new HTTP requests
2. Store.Wg.Wait()        — wait for all workers to finish current tasks
3. exit cleanly ✅         — zero task loss guaranteed
```

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit // blocks until signal

server.Shutdown(ctx)   // HTTP layer down
logic.Store.Wg.Wait() // worker layer down
```

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/posttasks` | Submit an array of tasks |
| GET | `/gettask/:id` | Get a task by ID and check its status |

### POST `/posttasks`
```json
{
    "tasks": [
        {"Taskmsg": "task1"},
        {"Taskmsg": "task2"},
        {"Taskmsg": "task3"},
        {"Taskmsg": "task4"},
        {"Taskmsg": "task5"}
    ]
}
```

### GET `/gettask/1`
```json
{
    "Id": 1,
    "Taskmsg": "task1",
    "Status": "processing",
    "Createdat": "2026-03-15T10:00:00Z",
    "Completedat": "0001-01-01T00:00:00Z"
}
```

---

## Key Concepts Learned

- **Unbuffered channels deadlock** if sender and receiver aren't both ready simultaneously — fix by running them in separate goroutines
- **`WaitGroup.Add()` must happen before the goroutine launches** — never inside it, or the wait can return before the goroutine even starts
- **`sync.Mutex` over channels** for simple shared state — right tool for read/write protection on a map
- **Two-level WaitGroups** for nested goroutine trees — one per level of concurrency
- **Graceful shutdown is ordered** — HTTP layer first, then worker layer, never the other way around
- **Buffered channels as job queues** — decouples producers from consumers, absorbs bursts without blocking

---

## Tech Stack

- **Go** — language
- **Gin** — HTTP framework
- **sync.Mutex** — thread-safe store
- **sync.WaitGroup** — goroutine lifecycle management
- **Buffered channels** — job queue
- **os/signal** — graceful shutdown

---

## Run Locally

```bash
git clone https://github.com/saumya712/Gotasker.git
cd Gotasker
go mod tidy
go run cmd/server/main.go
```

Server starts on `:8080`.
