# Story 6.5: Concurrent Progress Reporting

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want real-time progress during concurrent scans,
So that I know the scan is active and how many services have completed.

## Acceptance Criteria

1. **In-place progress display in concurrent TTY mode:** Given a user runs awtest with `--speed=fast` or `--speed=insane` in a TTY terminal, when the scan is in progress, then an in-place updating progress display shows "Scanning... 15/46 services complete" on stderr, and the display updates at minimum 2 Hz without flickering (NFR40).

2. **Progress replaced by summary on completion:** Given a concurrent scan completes, when all services finish, then the progress display is replaced by summary statistics (FR80).

3. **Quiet mode suppresses progress:** Given a user runs awtest with `--quiet` flag, when the scan executes, then no progress display is shown (FR82).

4. **Non-TTY suppresses progress:** Given awtest output is piped to a file or another command (non-TTY), when the scan executes, then no progress display is shown (FR83).

5. **Sequential mode preserves existing output:** Given a user runs awtest with `--speed=safe` (sequential mode), when the scan executes, then the existing per-service "Scanning [service_name]..." output is preserved (FR84), and no in-place progress counter is used.

6. **Progress writes to stderr only:** Given progress is displayed during concurrent scans, when output is written, then progress writes to stderr only — stdout contains only formatted scan results (FR81).

7. **Atomic progress counter:** Given a concurrent scan with all services at `--speed=insane`, when services complete at different times, then the progress counter increments atomically as each service finishes, regardless of completion order.

## Tasks / Subtasks

- [x] Task 1: Add `golang.org/x/term` dependency for TTY detection (AC: #4)
  - [x] Run `go get golang.org/x/term` to add the dependency to go.mod
  - [x] Verify Go 1.19 compatibility (stdlib `os.IsTerminal()` only available from Go 1.21+, `golang.org/x/term.IsTerminal()` works on Go 1.19)

- [x] Task 2: Create `progress.go` with progress reporter (AC: #1, #2, #3, #4, #6, #7)
  - [x] Implement `isTerminal(fd int) bool` wrapper around `term.IsTerminal()`
  - [x] Define progress reporter struct with: atomic completed count, total count, done channel, output writer (stderr)
  - [x] Implement `newProgressReporter(total int, quiet bool) *progressReporter` — returns nil if quiet mode or stderr is not a TTY (nil-safe pattern)
  - [x] Implement `(*progressReporter) Increment()` — atomically increments completed count (nil-safe: no-op on nil receiver)
  - [x] Implement `(*progressReporter) Start()` — launches ticker goroutine at 500ms interval (2 Hz), writes in-place progress to stderr using `\r` carriage return (nil-safe: no-op on nil receiver)
  - [x] Implement `(*progressReporter) Stop()` — stops ticker goroutine, clears progress line from stderr (nil-safe: no-op on nil receiver)
  - [x] Progress format: `"\rScanning... %d/%d services complete"` with padding to overwrite previous line
  - [x] On stop: write `"\r"` + spaces to clear line, then `"\r"` to reset cursor

- [x] Task 3: Integrate progress reporter into worker pool (AC: #1, #7)
  - [x] Add `progress *progressReporter` parameter to `runWorkerPool()` function signature
  - [x] In worker goroutine: call `progress.Increment()` after each service completes (after result append, before next job pull)
  - [x] Start progress ticker before spawning workers, stop after all workers complete (or after drain)

- [x] Task 4: Integrate progress reporter into `scanServices()` (AC: #3, #4, #5, #6)
  - [x] In `main.go` `scanServices()`: create progress reporter with `newProgressReporter(len(svcs), quiet)`
  - [x] Sequential mode (concurrency ≤ 1): do NOT use progress reporter — preserve existing per-service "Scanning..." output (FR84)
  - [x] Concurrent mode (concurrency > 1): pass progress reporter to `runWorkerPool()`
  - [x] Start progress before calling `runWorkerPool()`, stop after it returns

- [x] Task 5: Create `progress_test.go` with comprehensive tests (AC: #1-#7)
  - [x] Test `newProgressReporter` returns nil when quiet=true
  - [x] Test `newProgressReporter` returns nil when stderr is not a TTY (in test environment, stderr is typically not a TTY)
  - [x] Test nil receiver methods are safe (Increment, Start, Stop on nil pointer)
  - [x] Test `Increment()` atomically increments counter (concurrent calls from multiple goroutines with -race)
  - [x] Test progress output format contains expected pattern "Scanning..." and "services complete"
  - [x] Test progress writes to stderr (not stdout)
  - [x] Test Start/Stop lifecycle (start ticker, stop ticker, verify no goroutine leak)
  - [x] All tests run with `-race` flag

- [x] Task 6: Run full test suite and verify backward compatibility (AC: #5)
  - [x] Run `make test` — all existing tests pass including race detection
  - [x] Verify `--speed=safe` sequential behavior unchanged (per-service "Scanning..." output preserved)
  - [x] Verify no sync primitives imported in any `services/` package files
  - [x] Verify progress suppressed in non-TTY (test environment)

## Dev Notes

### Architecture & Design Decisions

- **New file `cmd/awtest/progress.go`:** Contains progress reporter with atomic counter and ticker goroutine. Placed in `cmd/awtest/` alongside `worker_pool.go`, `safe_scan.go`, `backoff.go`, `speed.go`. Package: `main`. [Source: architecture-phase2.md#Progress Reporting]
- **Atomic counter + ticker goroutine:** `sync/atomic` counter incremented as each service completes. A dedicated ticker goroutine writes in-place progress updates to stderr at 2+ Hz using ANSI escape codes (`\r` carriage return). [Source: architecture-phase2.md#Progress Reporting]
- **Nil-safe receiver pattern:** `newProgressReporter()` returns nil when progress should be suppressed (quiet mode or non-TTY). All methods use nil receiver checks so callers never need nil guards — simplifies integration code.
- **TTY detection via `golang.org/x/term`:** Go 1.19 requires external `golang.org/x/term.IsTerminal()` — stdlib `os.IsTerminal()` only added in Go 1.21. This is the only new external dependency. [Source: architecture-phase2.md#Phase 2 Technical Additions]
- **No new dependencies beyond `golang.org/x/term`:** Progress uses Go stdlib only (`sync/atomic`, `time`, `fmt`, `os`). [Source: architecture-phase2.md#Phase 2 Technical Additions]

### Progress Display Format

```
\rScanning... 15/46 services complete
```

- Uses `\r` (carriage return) for in-place update — no newline, overwrites previous line
- Pad with trailing spaces to ensure previous longer text is cleared
- On completion: clear the progress line entirely before summary output begins
- Updates at 500ms interval (2 Hz) per NFR40

### TTY Detection Logic

```go
import "golang.org/x/term"

func isTerminal(fd int) bool {
    return term.IsTerminal(fd)
}
```

Check `isTerminal(int(os.Stderr.Fd()))` — progress only displays when stderr is a TTY. This handles:
- Normal terminal use → TTY → show progress
- Piped output (`awtest ... | jq .`) → stderr may still be TTY → show progress
- Redirected stderr (`awtest ... 2>/dev/null`) → not TTY → suppress progress
- CI/CD environments → typically not TTY → suppress progress

**Important:** FR83 says "suppress progress when stdout is not a TTY (piped output)" but the architecture decision writes progress to stderr. The intent is to suppress progress when the user can't see it interactively. Check **stderr** for TTY status since that's where progress writes.

### Progress Reporter Struct

```go
type progressReporter struct {
    completed int64        // atomic counter
    total     int          // total services to scan
    done      chan struct{} // signal to stop ticker
    writer    *os.File     // stderr
}

func newProgressReporter(total int, quiet bool) *progressReporter {
    if quiet || !isTerminal(int(os.Stderr.Fd())) {
        return nil // suppress progress
    }
    return &progressReporter{
        total:  total,
        done:   make(chan struct{}),
        writer: os.Stderr,
    }
}

func (p *progressReporter) Increment() {
    if p == nil {
        return
    }
    atomic.AddInt64(&p.completed, 1)
}

func (p *progressReporter) Start() {
    if p == nil {
        return
    }
    go func() {
        ticker := time.NewTicker(500 * time.Millisecond)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                count := atomic.LoadInt64(&p.completed)
                fmt.Fprintf(p.writer, "\rScanning... %d/%d services complete", count, p.total)
            case <-p.done:
                return
            }
        }
    }()
}

func (p *progressReporter) Stop() {
    if p == nil {
        return
    }
    close(p.done)
    // Clear progress line
    fmt.Fprintf(p.writer, "\r%-50s\r", "") // overwrite with spaces, reset cursor
}
```

### Integration Points

**Worker Pool (`worker_pool.go`):**
```go
// Function signature change:
func runWorkerPool(ctx context.Context, svcs []types.AWSService, sess *session.Session,
    concurrency int, quiet, debug bool, progress *progressReporter) ([]types.ScanResult, []string)

// Inside worker goroutine, after result append (line ~65):
progress.Increment()
```

**Main Scan Orchestration (`main.go` `scanServices()` lines 292-315):**
```go
func scanServices(ctx context.Context, svcs []types.AWSService, sess *session.Session,
    concurrency int, quiet, debug bool) ([]types.ScanResult, []string) {
    // Sequential mode: preserve Phase 1 behavior exactly (no progress reporter)
    if concurrency <= 1 {
        // ... unchanged sequential loop with "Scanning %s..." output ...
    }

    // Concurrent mode: create progress reporter and delegate to worker pool
    progress := newProgressReporter(len(svcs), quiet)
    progress.Start()
    results, skipped := runWorkerPool(ctx, svcs, sess, concurrency, quiet, debug, progress)
    progress.Stop()
    return results, skipped
}
```

### Suppression Conditions

Progress is suppressed (nil reporter returned) when ANY of these is true:
1. `--quiet` flag is set (FR82)
2. stderr is not a TTY — piped/redirected output (FR83)
3. `--speed=safe` / concurrency=1 — sequential mode uses per-service output instead (FR84)

### Previous Story Intelligence (6.4)

- **Worker pool integration point:** `worker_pool.go` line 58: `scanWithBackoff()` call — progress Increment should occur after this call and after result append
- **Testing pattern:** stdlib `testing` only (no testify in `cmd/awtest/` tests). Use `t.Errorf()` for soft assertions, `t.Fatalf()` for hard failures. Table-driven tests.
- **Race detection:** All tests run with `go test -race`. The atomic counter is the core thread-safety mechanism.
- **Drain handling:** Worker pool has atomic `drained` flag and 1-second drain deadline. Progress ticker should stop after drain completes (in `scanServices`, after `runWorkerPool` returns).
- **ErrorCategory `_` discard pattern:** Workers still discard `ErrorCategory` — not relevant to progress.
- **File naming convention:** `snake_case.go` — so `progress.go` and `progress_test.go`

### Git Intelligence

Recent commits:
- `898c4b9` — Add rate limit resilience with exponential backoff (Story 6.4)
- `2d02ab2` — Add concurrent worker pool execution with graceful drain (Story 6.3)
- `6bb8c12` — Add safeScan wrapper with panic recovery and error classification (Story 6.2)
- `02f3875` — Add --speed preset and concurrency flag resolution (Story 6.1)

**Patterns from recent work:**
- Files created in `cmd/awtest/`: `speed.go`, `speed_test.go`, `safe_scan.go`, `safe_scan_test.go`, `worker_pool.go`, `worker_pool_test.go`, `backoff.go`, `backoff_test.go`
- Test files use stdlib `testing` only (no testify assertions)
- `main.go` modified to integrate new features into `scanServices()`
- Each story follows clean layering: 6.1 (flags) → 6.2 (safeScan) → 6.3 (worker pool) → 6.4 (backoff) → 6.5 (progress)
- `go.mod` only updated when new dependencies added

### FRs Covered

- **FR79:** In-place updating progress during concurrent scans showing services completed vs. total
- **FR80:** Progress replaced by summary statistics upon scan completion
- **FR81:** Progress writes to stderr to avoid interfering with stdout formatted output
- **FR82:** Progress suppressed when `--quiet` flag is set
- **FR83:** Progress suppressed when stdout is not a TTY (piped output)
- **FR84:** Sequential mode preserves existing per-service progress reporting

### NFRs Addressed

- **NFR40:** In-place progress updates render at minimum 2 updates per second without flickering or tearing (500ms ticker interval)
- **NFR46:** Fault isolation — progress counter is atomic, independent of result collection

### Anti-Patterns to Avoid

- **DO NOT** use `time.Sleep()` for ticker — use `time.NewTicker()` for consistent intervals
- **DO NOT** write progress to stdout — progress MUST go to stderr only (FR81)
- **DO NOT** add progress inside service files or `services/` package — progress is managed by the worker pool caller
- **DO NOT** use `sync.Mutex` for the progress counter — use `sync/atomic` for lock-free incrementing on the hot path
- **DO NOT** use `fmt.Println` for progress — use `fmt.Fprintf(os.Stderr, "\r...")` for in-place update
- **DO NOT** add a newline to progress output — use `\r` only for in-place overwrite
- **DO NOT** start progress ticker for sequential mode — sequential mode uses its own per-service output (FR84)
- **DO NOT** import `sync` in progress.go unless needed — `sync/atomic` is sufficient for the counter
- **DO NOT** forget to Stop() the progress ticker — leaked ticker goroutine = goroutine leak
- **DO NOT** add progress Increment inside `safeScan` or `scanWithBackoff` — keep it in the worker goroutine after result collection

### Project Structure Notes

- New files in `cmd/awtest/`: `progress.go`, `progress_test.go`
- Modified files: `cmd/awtest/worker_pool.go` (add progress.Increment() after result append, add progress parameter), `cmd/awtest/main.go` (create/start/stop progress reporter in scanServices concurrent path)
- New dependency in `go.mod`: `golang.org/x/term` (for `IsTerminal()` on Go 1.19)
- Package: `main` (same as all other `cmd/awtest/` files)
- No changes to `services/`, `types/`, `formatters/`, or `utils/` packages

### References

- [Source: architecture-phase2.md#Progress Reporting] — Atomic counter + ticker goroutine decision
- [Source: architecture-phase2.md#Concurrency Patterns] — Worker pool contract, progress tracking
- [Source: architecture-phase2.md#Process Patterns] — Scan execution flow step 3: start progress ticker
- [Source: architecture-phase2.md#Implementation Handoff] — Progress is step 4 in implementation sequence
- [Source: architecture-phase2.md#Phase 2 Technical Additions] — golang.org/x/term dependency
- [Source: epics-phase2.md#Story 1.5] — BDD acceptance criteria (FR79-84)
- [Source: prd-phase2.md FR79-84] — Concurrent progress reporting requirements
- [Source: prd-phase2.md NFR40] — Progress updates at 2+ Hz without flickering
- [Source: cmd/awtest/worker_pool.go:58] — scanWithBackoff call where Increment should follow
- [Source: cmd/awtest/worker_pool.go:61-64] — mutex-protected result append (Increment after this)
- [Source: cmd/awtest/main.go:292-315] — scanServices() function with sequential/concurrent branching
- [Source: cmd/awtest/main.go:303-304] — Sequential mode "Scanning %s..." output to preserve
- [Source: go.mod] — Go 1.19, golang.org/x/term not yet in direct dependencies
- [Source: 6-4-rate-limit-resilience-exponential-backoff.md] — Previous story patterns and learnings

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- Initial `go get golang.org/x/term` pulled v0.40.0 and bumped go directive to 1.24.0. After `go mod tidy` and actual import, resolved to v0.1.0 (compatible with Go 1.19). Restored go directive to `go 1.19`.
- `backoff_test.go` also had a `runWorkerPool` call requiring the new `progress` parameter — fixed alongside `worker_pool_test.go`.

### Completion Notes List

- Created `progress.go` with nil-safe `progressReporter` using `sync/atomic` counter and `time.NewTicker` at 500ms (2 Hz). TTY detection via `golang.org/x/term.IsTerminal()`.
- Integrated progress into `worker_pool.go` — added `progress *progressReporter` parameter, `Increment()` called after each service result append.
- Integrated progress into `main.go` `scanServices()` — creates/starts/stops reporter in concurrent path only. Sequential mode untouched (FR84).
- Created 11 tests in `progress_test.go`: nil-safety, atomic concurrency, output format, stderr targeting, lifecycle, goroutine leak prevention, TTY detection.
- All 56 tests pass with `-race` flag, zero regressions.
- No sync primitives in `services/` package. Progress writes to stderr only (FR81).

### Change Log

- 2026-03-08: Implemented concurrent progress reporting (Story 6.5) — added progress.go with nil-safe reporter, integrated into worker pool and scanServices, 11 new tests, golang.org/x/term dependency added.

### File List

- `cmd/awtest/progress.go` (new) — progressReporter with atomic counter, ticker, nil-safe methods
- `cmd/awtest/progress_test.go` (new) — 11 tests covering all ACs
- `cmd/awtest/worker_pool.go` (modified) — added progress parameter and Increment() call
- `cmd/awtest/worker_pool_test.go` (modified) — updated all runWorkerPool calls with progress parameter
- `cmd/awtest/backoff_test.go` (modified) — updated runWorkerPool call with progress parameter
- `cmd/awtest/main.go` (modified) — create/start/stop progress reporter in scanServices concurrent path
- `go.mod` (modified) — added golang.org/x/term v0.1.0 direct dependency
- `go.sum` (modified) — updated checksums for golang.org/x/term and golang.org/x/sys
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified) — story 6.5 status updated to review
