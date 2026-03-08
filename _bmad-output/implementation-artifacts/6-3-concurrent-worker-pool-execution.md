# Story 6.3: Concurrent Worker Pool Execution

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want services to scan concurrently using a worker pool,
So that scans complete in seconds instead of minutes at higher speed presets.

## Acceptance Criteria

1. **Concurrent execution via worker pool:** Given `--speed=fast` (concurrency=5), 5 worker goroutines process services concurrently via buffered channel dispatch, and all service executions go through the `safeScan` wrapper.
2. **Deterministic output ordering:** Given concurrent workers complete services in arbitrary order, results are sorted by service name so output is identical to `--speed=safe` for the same credentials (FR71).
3. **Thread-safe result collection:** Given multiple workers append results simultaneously, results are protected by `sync.Mutex` — no data corruption (FR72).
4. **Context cancellation and graceful drain:** Given the scan timeout is reached during concurrent execution, context cancellation propagates to workers, in-progress services drain within 1 second (FR98), partial results collected before cancellation are preserved (FR99), and output is formatted with available results.
5. **Safe mode backward compatibility:** Given `--speed=safe` (concurrency=1), behavior is identical to Phase 1 sequential scanning — no regression.
6. **Memory discipline:** Given `--speed=insane` (20 workers) with all services executing concurrently, memory consumption remains under 100MB (NFR39).
7. **Read-only safety:** Given any concurrency level, all operations remain strictly read-only (FR96, NFR42).

## Tasks / Subtasks

- [x] Task 1: Create `worker_pool.go` with concurrent worker pool implementation (AC: #1, #2, #3, #4)
  - [x] Define `workerPool` struct with fields: `concurrency int`, `results []types.ScanResult`, `skipped []string`, `mu sync.Mutex`, `wg sync.WaitGroup`
  - [x] Implement `runWorkerPool(ctx context.Context, svcs []types.AWSService, sess *session.Session, concurrency int, quiet, debug bool) ([]types.ScanResult, []string)` function
  - [x] Create buffered channel `chan types.AWSService` sized to `len(svcs)` for service dispatch
  - [x] Spawn `concurrency` worker goroutines, each pulling services from the channel
  - [x] Each worker calls `safeScan(ctx, service, sess, debug)` — never direct Call/Process
  - [x] Workers append results under `sync.Mutex` protection
  - [x] Workers track skipped services (context cancelled) under mutex
  - [x] Close channel after all services submitted, workers exit when channel drained
  - [x] `sync.WaitGroup` to wait for all workers to complete
  - [x] After all workers done: sort results by `ServiceName` for deterministic output
  - [x] Handle context cancellation: workers check `ctx.Done()` before pulling next service

- [x] Task 2: Implement graceful drain with 1-second deadline (AC: #4)
  - [x] On context cancellation, workers finish current in-progress service (safeScan respects context in service.Call)
  - [x] Use `WaitGroup` with a 1-second deadline channel: if workers don't finish within 1s of context cancel, proceed with partial results
  - [x] Collect whatever results exist at drain completion
  - [x] Remaining unstarted services added to skipped list

- [x] Task 3: Update `scanServices()` in `main.go` to delegate to worker pool (AC: #1, #5)
  - [x] If `concurrency == 1`: keep existing sequential loop (preserves Phase 1 behavior exactly)
  - [x] If `concurrency > 1`: call `runWorkerPool()` instead
  - [x] Preserve stderr "Scanning..." output for sequential mode (no change to safe mode)
  - [x] For concurrent mode: suppress per-service "Scanning..." messages (Story 6.5 adds progress reporting)

- [x] Task 4: Create `worker_pool_test.go` with comprehensive tests (AC: #1-#7)
  - [x] Test concurrent execution: N mock services complete in parallel (verify wall-clock speedup)
  - [x] Test deterministic ordering: services completing in random order produce sorted results
  - [x] Test thread safety: run with `-race` flag, multiple workers appending simultaneously
  - [x] Test context cancellation: cancel mid-scan, verify partial results preserved and unstarted services skipped
  - [x] Test graceful drain deadline: slow service exceeding 1s drain gets cut off
  - [x] Test concurrency=1 fallback: sequential behavior identical to direct loop
  - [x] Test all services go through safeScan: panicking service doesn't crash pool
  - [x] Test empty service list: no workers spawned, no results
  - [x] Test single service: works correctly with any concurrency level

- [x] Task 5: Verify backward compatibility and run full test suite (AC: #5, #7)
  - [x] Run `make test` (includes `-race` flag) — all existing tests pass
  - [x] Verify `--speed=safe` produces identical output to Phase 1
  - [x] Verify no sync primitives imported in any `services/` package files

## Dev Notes

### Architecture & Design Decisions

- **New file `cmd/awtest/worker_pool.go`:** Contains the concurrent worker pool. Placed in `cmd/awtest/` alongside `main.go`, `safe_scan.go`, `speed.go`. Package: `main`. [Source: architecture-phase2.md#Project Structure]
- **Buffered channel + fixed goroutine pool:** Pre-spawn N workers, feed services via buffered channel. Workers pull until channel closed. [Source: architecture-phase2.md#Concurrency Model]
- **Mutex-protected shared slice:** Workers append `ScanResult` entries under `sync.Mutex`. Post-scan sort by service name. [Source: architecture-phase2.md#Result Collection]
- **No new dependencies:** Uses Go stdlib only (`sync`, `sync/atomic`, `sort`). [Source: architecture-phase2.md#Phase 2 Technical Additions]
- **Services remain concurrency-unaware:** No sync primitives in `services/` packages. The worker pool handles all parallelism transparently. [Source: architecture-phase2.md#Concurrency Patterns]

### Function Signatures

```go
// runWorkerPool executes services concurrently using a fixed worker pool.
// Returns collected results (sorted by ServiceName) and skipped service names.
func runWorkerPool(ctx context.Context, svcs []types.AWSService, sess *session.Session, concurrency int, quiet, debug bool) ([]types.ScanResult, []string)
```

### Worker Pool Implementation Pattern

```go
func runWorkerPool(ctx context.Context, svcs []types.AWSService, sess *session.Session, concurrency int, quiet, debug bool) ([]types.ScanResult, []string) {
    var (
        results  []types.ScanResult
        skipped  []string
        mu       sync.Mutex
        wg       sync.WaitGroup
    )

    // Buffered channel sized to total services — non-blocking submit
    jobs := make(chan types.AWSService, len(svcs))

    // Spawn workers
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for service := range jobs {
                // Check context before starting each service
                select {
                case <-ctx.Done():
                    mu.Lock()
                    skipped = append(skipped, service.Name)
                    mu.Unlock()
                    continue
                default:
                }

                serviceResults, _ := safeScan(ctx, service, sess, debug)

                mu.Lock()
                results = append(results, serviceResults...)
                mu.Unlock()
            }
        }()
    }

    // Submit all services
    for _, svc := range svcs {
        jobs <- svc
    }
    close(jobs)

    // Wait for workers with drain deadline
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        // All workers completed normally
    case <-time.After(/* drain deadline logic */):
        // Drain timeout — proceed with partial results
    }

    // Sort results by service name for deterministic output
    sort.Slice(results, func(i, j int) bool {
        return results[i].ServiceName < results[j].ServiceName
    })

    return results, skipped
}
```

**Key details:**
- Channel is buffered to `len(svcs)` so all services can be submitted without blocking
- Workers drain the channel; after channel close, they process remaining items then exit
- Context check happens before each service — workers don't pick up new work after cancellation
- In-progress `safeScan` calls respect context through `service.Call(ctx, sess)`
- Mutex critical section is minimal: single append operation
- Post-completion sort ensures deterministic output regardless of completion order

### Graceful Drain Strategy

The drain deadline applies ONLY when context is already cancelled:

```go
// After submitting all jobs and closing channel:
done := make(chan struct{})
go func() {
    wg.Wait()
    close(done)
}()

// If context is already done, apply 1s drain deadline
// If context is not done, wait indefinitely for workers to finish
select {
case <-done:
    // Normal completion
case <-ctx.Done():
    // Context cancelled — give workers 1s to finish current work
    select {
    case <-done:
        // Workers finished within drain window
    case <-time.After(1 * time.Second):
        // Drain timeout expired — proceed with what we have
        // Any services still in jobs channel are skipped
        mu.Lock()
        // Drain remaining jobs from channel as skipped
        for svc := range jobs {
            skipped = append(skipped, svc.Name)
        }
        mu.Unlock()
    }
}
```

**Note:** The `jobs` channel is already closed before waiting, so draining remaining items from it after timeout is safe. In practice, since `jobs` is buffered to `len(svcs)` and all services were submitted before close, the channel will be empty unless workers stopped pulling (which only happens after context cancel + current service completion).

### scanServices Integration

```go
func scanServices(ctx context.Context, svcs []types.AWSService, sess *session.Session, concurrency int, quiet, debug bool) ([]types.ScanResult, []string) {
    // Sequential mode: preserve Phase 1 behavior exactly
    if concurrency <= 1 {
        var results []types.ScanResult
        var skippedServices []string
        for _, service := range svcs {
            select {
            case <-ctx.Done():
                skippedServices = append(skippedServices, service.Name)
                continue
            default:
                if !quiet {
                    fmt.Fprintf(os.Stderr, "Scanning %s...\n", service.Name)
                }
                serviceResults, _ := safeScan(ctx, service, sess, debug)
                results = append(results, serviceResults...)
            }
        }
        return results, skippedServices
    }

    // Concurrent mode: delegate to worker pool
    return runWorkerPool(ctx, svcs, sess, concurrency, quiet, debug)
}
```

### Error Category Handling in Worker Pool

The `ErrorCategory` returned by `safeScan` is currently ignored (same as Story 6.2's integration). Story 6.4 (backoff) will wrap the safeScan call with retry logic based on `ErrorThrottle`. The worker pool should preserve the `_` discard pattern for now — backoff integration happens in Story 6.4.

### Testing Strategy

Tests should create mock `types.AWSService` structs with controllable timing:

```go
// Mock service with configurable delay
func mockService(name string, delay time.Duration) types.AWSService {
    return types.AWSService{
        Name: name,
        Call: func(ctx context.Context, sess *session.Session) (interface{}, error) {
            select {
            case <-time.After(delay):
                return name + "-output", nil
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        },
        Process: func(output interface{}, err error, debug bool) []types.ScanResult {
            if err != nil {
                return []types.ScanResult{{
                    ServiceName: name,
                    MethodName:  name,
                    Error:       err,
                    Timestamp:   time.Now(),
                }}
            }
            return []types.ScanResult{{
                ServiceName: name,
                MethodName:  name,
                Timestamp:   time.Now(),
            }}
        },
    }
}
```

**Race detection:** All tests run with `go test -race`. The mock services with concurrent access to shared results will trigger race detector if mutex is missing.

**Concurrent speedup test:**
```go
// 10 services each taking 100ms. At concurrency=5, should complete in ~200ms, not 1000ms.
start := time.Now()
results, _ := runWorkerPool(ctx, services, nil, 5, true, false)
elapsed := time.Since(start)
if elapsed > 500*time.Millisecond {
    t.Errorf("concurrent execution too slow: %v (expected <500ms)", elapsed)
}
```

**Session for tests:** Pass `nil` for session in tests — safeScan passes it to `service.Call()`, and mock services don't use it.

### FRs Covered

- **FR67:** Concurrent execution engine with configurable worker pool
- **FR71:** Identical results regardless of speed preset (sort by service name)
- **FR72:** Thread-safe result collection (mutex-protected slice)
- **FR73:** Deterministic ordering by service name
- **FR98:** In-progress services drain within 1 second on timeout
- **FR99:** Partial results preserved on timeout/cancellation

### NFRs Addressed

- **NFR35-38:** Performance scaling with concurrent workers
- **NFR39:** Memory under 100MB at insane (bounded goroutine pool)
- **NFR42:** Read-only guarantee maintained under concurrency
- **NFR46:** Fault isolation per goroutine (via safeScan)
- **NFR47:** Concurrent results identical to sequential (sort ensures determinism)
- **NFR48:** Crash recovery without losing other results (safeScan panic recovery)
- **NFR49:** Context cancellation cleanly terminates within 1 second
- **NFR55:** Concurrent execution compatible with all formatters (sorted results, same type)
- **NFR56:** Concurrency encapsulated in worker pool — services remain unaware
- **NFR57:** No concurrency-specific code in services
- **NFR58:** Code coverage with race detection tests

### Project Structure Notes

- New files in `cmd/awtest/`: `worker_pool.go`, `worker_pool_test.go`
- Modified file: `cmd/awtest/main.go` (scanServices delegates to pool when concurrency > 1)
- Package: `main` (same as `safe_scan.go`, `speed.go`)
- No changes to `services/`, `types/`, `formatters/`, or `utils/` packages
- No changes to `go.mod` (all Go stdlib: `sync`, `sort`, `time`)

### Previous Story Intelligence (6.2)

- **safeScan signature:** `func safeScan(ctx context.Context, service types.AWSService, sess *session.Session, debug bool) ([]types.ScanResult, ErrorCategory)` — worker pool calls this directly
- **ErrorCategory ignored for now:** Story 6.2 set `_` discard pattern for ErrorCategory in scanServices. Worker pool should follow same pattern; Story 6.4 adds retry logic.
- **Testing pattern:** Table-driven tests with stdlib `testing` (no testify in `cmd/awtest/` tests). Use `t.Errorf()` for soft assertions, `t.Fatalf()` for hard failures.
- **Code review lesson from 6.1:** Ensure functions are properly wired — don't leave stub connections. The worker pool must be called from scanServices, not just defined.
- **Panic recovery already handled:** safeScan's defer/recover means worker goroutines won't crash even if a service panics. The worker pool does NOT need its own panic recovery — safeScan handles it.

### Git Intelligence

Recent commits:
- `6bb8c12` — Add safeScan wrapper with panic recovery and error classification (Story 6.2)
- `02f3875` — Add --speed preset and concurrency flag resolution (Story 6.1)

**Patterns from recent work:**
- Files created in `cmd/awtest/`: `speed.go`, `speed_test.go`, `safe_scan.go`, `safe_scan_test.go`
- Test files use stdlib `testing` only (no testify assertions)
- `main.go` modified to integrate new features into `scanServices()`
- Sprint status updated as part of implementation

### Anti-Patterns to Avoid

- **DO NOT** add `sync.Mutex` or `sync/atomic` inside any `services/` file — services must remain concurrency-unaware
- **DO NOT** call `service.Call()` or `service.Process()` directly from workers — always use `safeScan()`
- **DO NOT** write to stdout from worker goroutines — only the formatter writes to stdout after all results collected
- **DO NOT** use global backoff state — per-service backoff is Story 6.4's responsibility
- **DO NOT** spawn goroutines inside service Call/Process methods
- **DO NOT** use `os.Exit()` from worker goroutines — only main goroutine controls exit
- **DO NOT** use unbounded goroutines (one per service) — use fixed worker pool sized by concurrency setting

### References

- [Source: architecture-phase2.md#Core Architectural Decisions] — Buffered channel + fixed goroutine pool decision
- [Source: architecture-phase2.md#Result Collection] — Mutex-protected shared slice decision
- [Source: architecture-phase2.md#Concurrency Patterns] — Worker pool contract, service contract
- [Source: architecture-phase2.md#Process Patterns] — Scan execution flow, graceful shutdown
- [Source: architecture-phase2.md#Enforcement Guidelines] — Anti-patterns and requirements
- [Source: architecture-phase2.md#Project Structure] — worker_pool.go placement in cmd/awtest/
- [Source: epics-phase2.md#Story 1.3] — BDD acceptance criteria
- [Source: prd-phase2.md#Concurrent Enumeration] — FR67-73 requirements
- [Source: prd-phase2.md#Concurrent Safety] — FR96-99 requirements
- [Source: prd-phase2.md#Performance] — NFR35-39 performance targets
- [Source: cmd/awtest/main.go:289-311] — Current scanServices() sequential loop
- [Source: cmd/awtest/safe_scan.go:44-66] — safeScan() function with ErrorCategory
- [Source: cmd/awtest/speed.go:19-23] — SpeedResult struct with Concurrency field
- [Source: cmd/awtest/types/types.go:22-32] — ScanResult struct
- [Source: cmd/awtest/types/types.go:39-44] — AWSService struct with Call/Process functions
- [Source: cmd/awtest/services/services.go:53-104] — AllServices() registry
- [Source: go.mod] — Go 1.19, no new deps needed
- [Source: Makefile] — `make test` includes `-race` flag

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Resolved `mockService` naming collision with existing `timeout_test.go` helper — reused existing function instead of duplicating.

### Completion Notes List

- Created `worker_pool.go` with `runWorkerPool()` implementing buffered channel + fixed goroutine pool pattern
- Workers use `safeScan()` exclusively — no direct `Call`/`Process` invocations
- Mutex-protected shared slices for results and skipped services
- Atomic drain flag (`sync/atomic`) prevents data race on drain timeout — workers stop appending after flag is set, sort/return happens only after no more appends possible
- Graceful drain: 1-second deadline after context cancellation, remaining channel items drained into local slice then appended under lock once
- Results sorted by `ServiceName` post-completion for deterministic output
- Guard against concurrency < 1 (defaults to 1) to prevent deadlock
- Updated `scanServices()` in `main.go`: delegates to `runWorkerPool()` when concurrency > 1, preserves sequential Phase 1 behavior when concurrency <= 1
- Sequential mode retains per-service "Scanning..." stderr output; concurrent mode suppresses it (Story 6.5 adds progress)
- 12 tests covering: concurrent speedup, deterministic ordering, thread safety (race detector), context cancellation, graceful drain deadline (with partial result assertions), concurrency=1 fallback, runWorkerPool(1) vs scanServices(1) equivalence, panic isolation, empty list, single service, scanServices delegation
- All tests pass with `-race` flag, no data races detected
- No sync primitives in `services/` packages — services remain concurrency-unaware
- No new dependencies — uses Go stdlib only (`sync`, `sync/atomic`, `sort`, `time`)

### File List

- `cmd/awtest/worker_pool.go` (new) — Concurrent worker pool implementation
- `cmd/awtest/worker_pool_test.go` (new) — Comprehensive worker pool tests (12 tests)
- `cmd/awtest/main.go` (modified) — Updated `scanServices()` to delegate to worker pool when concurrency > 1
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified) — Story status tracking

### Change Log

- 2026-03-08: Addressed code review findings — 8 items resolved: critical data race on drain timeout (atomic flag), high mutex scope issue (local drain slice), medium test improvements (drain assertions, equivalence test, File List fix), low guards (concurrency validation, documentation)
- 2026-03-07: Implemented concurrent worker pool execution (Story 6.3) — buffered channel dispatch, mutex-protected result collection, graceful drain with 1s deadline, deterministic output ordering
