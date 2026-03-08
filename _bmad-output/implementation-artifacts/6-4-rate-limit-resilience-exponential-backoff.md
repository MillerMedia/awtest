# Story 6.4: Rate Limit Resilience with Exponential Backoff

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want automatic retry with exponential backoff when AWS throttles API calls,
So that I get complete results even at high concurrency without manual intervention.

## Acceptance Criteria

1. **Exponential backoff on throttling:** Given a service receives an AWS throttling response (429/RequestLimitExceeded), when the backoff retry logic executes, then the service retries with exponential backoff (base 1s, 2x multiplier, +/-50% jitter), maximum 3 retries before marking as rate-limited error, and maximum total delay per service is 15 seconds (NFR51).

2. **Per-service independent backoff:** Given one service is being throttled, when it enters backoff retry, then other concurrent services continue executing unblocked (FR75), and backoff state is per-service, not global.

3. **Jitter prevents retry storms:** Given multiple services retry simultaneously, when jitter is applied to retry timing, then retry attempts are spread across time to prevent retry storms (FR76).

4. **Transparent success after retry:** Given a throttled service eventually succeeds after retry, when results are collected, then the result appears normal — no indication of throttling to the user.

5. **Exhausted retries produce rate-limited error:** Given a throttled service exhausts all 3 retries, when the final retry fails, then the service is included in results as a rate-limited error, and the error message indicates rate limiting.

6. **Context cancellation abandons retries:** Given backoff is in progress and context cancellation occurs, when the retry loop checks context between retries, then the retry is abandoned and partial results are preserved.

7. **Complete results for non-throttled services:** Given any concurrency level from 1 to 20, when throttling occurs on some services, then the system delivers complete results for non-throttled services (FR78).

## Tasks / Subtasks

- [x] Task 1: Create `backoff.go` with exponential backoff retry logic (AC: #1, #2, #3, #6)
  - [x] Define backoff constants: `backoffBaseDelay = 1 * time.Second`, `backoffMultiplier = 2.0`, `backoffMaxRetries = 3`, `backoffMaxDelay = 15 * time.Second`, `backoffJitterFactor = 0.5`
  - [x] Implement `calculateBackoff(attempt int) time.Duration` — computes delay: `min(baseDelay * 2^attempt * (0.5 + rand()*1.0), maxDelay)` with ±50% jitter
  - [x] Implement `scanWithBackoff(ctx context.Context, service types.AWSService, sess *session.Session, debug bool) ([]types.ScanResult, ErrorCategory)` function
  - [x] Retry loop: call `safeScan()`, check `ErrorCategory` — if `ErrorThrottle`, compute backoff delay, sleep with context-aware select, retry
  - [x] On `ErrorDenied` or `ErrorService` or `ErrorNone`: return immediately (no retry)
  - [x] On max retries exceeded: return rate-limited error result with clear error message
  - [x] Context check between retries: use `select { case <-ctx.Done(): return; case <-time.After(delay): continue }`
  - [x] All backoff state is local to the function invocation — no shared/global state

- [x] Task 2: Integrate backoff into worker pool and sequential scan path (AC: #2, #4, #7)
  - [x] In `worker_pool.go`: replace `safeScan(ctx, service, sess, debug)` call with `scanWithBackoff(ctx, service, sess, debug)` call
  - [x] In `main.go` `scanServices()`: replace `safeScan(ctx, service, sess, debug)` call with `scanWithBackoff(ctx, service, sess, debug)` in sequential path
  - [x] Verify: backoff is per-service invocation — each worker handles its own retry independently
  - [x] Verify: `ErrorCategory` return value still discarded (`_`) in both paths — classification used internally by backoff

- [x] Task 3: Create `backoff_test.go` with comprehensive tests (AC: #1-#7)
  - [x] Test `calculateBackoff` returns values within expected range for each attempt (0, 1, 2)
  - [x] Test `calculateBackoff` jitter produces varied results (call 100 times, verify not all identical)
  - [x] Test `calculateBackoff` delay never exceeds `backoffMaxDelay`
  - [x] Test `scanWithBackoff` with non-throttle error: returns immediately, no retry
  - [x] Test `scanWithBackoff` with access denied: returns immediately, no retry
  - [x] Test `scanWithBackoff` with successful call: returns immediately, no retry
  - [x] Test `scanWithBackoff` with throttle then success: retries and returns success result
  - [x] Test `scanWithBackoff` with throttle exhausting all retries: returns rate-limited error
  - [x] Test `scanWithBackoff` rate-limited error message contains "rate limited" text
  - [x] Test `scanWithBackoff` with context cancellation during backoff sleep: returns promptly
  - [x] Test `scanWithBackoff` preserves normal results after successful retry (AC #4 — transparent success)
  - [x] All tests run with `-race` flag

- [x] Task 4: Run full test suite and verify backward compatibility (AC: #5, #7)
  - [x] Run `make test` — all existing tests pass including race detection
  - [x] Verify `--speed=safe` behavior unchanged (backoff is transparent layer)
  - [x] Verify no sync primitives imported in any `services/` package files

## Dev Notes

### Architecture & Design Decisions

- **New file `cmd/awtest/backoff.go`:** Contains exponential backoff retry logic. Placed in `cmd/awtest/` alongside `safe_scan.go`, `worker_pool.go`, `speed.go`. Package: `main`. [Source: architecture-phase2.md#Project Structure]
- **Per-service inline backoff:** Each invocation of `scanWithBackoff` has its own retry state — no global or shared backoff coordination. This ensures throttling on one service never blocks another. [Source: architecture-phase2.md#Rate Limit Backoff]
- **Wraps safeScan:** `scanWithBackoff` calls `safeScan` internally and inspects the `ErrorCategory` return. On `ErrorThrottle`, it retries. On all other categories, it returns immediately. This preserves the panic recovery and error classification from Story 6.2.
- **No new dependencies:** Uses Go stdlib only (`math/rand`, `time`, `context`). [Source: architecture-phase2.md#Phase 2 Technical Additions]

### Backoff Formula

From architecture-phase2.md#Concurrency Patterns:
```
Base delay: 1 second
Multiplier: 2x per retry
Jitter: ±50% randomization on each delay
Max delay cap: 15 seconds
Max retries: 3
Formula: min(baseDelay * 2^attempt * (0.5 + rand()), 15s)
```

Retry schedule (approximate ranges with jitter):
- Attempt 0 (first retry): 0.5s – 1.5s
- Attempt 1 (second retry): 1.0s – 3.0s
- Attempt 2 (third retry): 2.0s – 6.0s
- Total worst-case: ~10.5s (well within 15s NFR51 cap)

### Function Signatures

```go
// backoff constants
const (
    backoffBaseDelay   = 1 * time.Second
    backoffMultiplier  = 2.0
    backoffMaxRetries  = 3
    backoffMaxDelay    = 15 * time.Second
    backoffJitterFactor = 0.5
)

// calculateBackoff returns the backoff delay for the given retry attempt.
// Applies exponential growth with ±50% jitter, capped at backoffMaxDelay.
func calculateBackoff(attempt int) time.Duration

// scanWithBackoff wraps safeScan with retry logic for throttled requests.
// On ErrorThrottle, retries up to backoffMaxRetries times with exponential backoff.
// All other error categories return immediately without retry.
// Backoff state is local — no global coordination.
func scanWithBackoff(ctx context.Context, service types.AWSService, sess *session.Session, debug bool) ([]types.ScanResult, ErrorCategory)
```

### Integration Points

**Worker Pool (`worker_pool.go` line 58):**
```go
// BEFORE (Story 6.3):
serviceResults, _ := safeScan(ctx, service, sess, debug)

// AFTER (Story 6.4):
serviceResults, _ := scanWithBackoff(ctx, service, sess, debug)
```

**Sequential Path (`main.go` line 306):**
```go
// BEFORE (Story 6.3):
serviceResults, _ := safeScan(ctx, service, sess, debug)

// AFTER (Story 6.4):
serviceResults, _ := scanWithBackoff(ctx, service, sess, debug)
```

### Context-Aware Backoff Sleep Pattern

```go
func scanWithBackoff(ctx context.Context, service types.AWSService, sess *session.Session, debug bool) ([]types.ScanResult, ErrorCategory) {
    for attempt := 0; attempt <= backoffMaxRetries; attempt++ {
        results, category := safeScan(ctx, service, sess, debug)

        if category != ErrorThrottle {
            return results, category
        }

        // Last attempt exhausted — don't sleep, return rate-limited error
        if attempt == backoffMaxRetries {
            break
        }

        delay := calculateBackoff(attempt)
        select {
        case <-ctx.Done():
            // Context cancelled during backoff — return what we have
            return results, category
        case <-time.After(delay):
            // Continue to next retry
        }
    }

    // All retries exhausted — return rate-limited error
    return []types.ScanResult{{
        ServiceName: service.Name,
        MethodName:  service.Name,
        Error:       fmt.Errorf("service rate limited after %d retries", backoffMaxRetries),
        Timestamp:   time.Now(),
    }}, ErrorThrottle
}
```

### Error Category Flow

```
safeScan returns ErrorThrottle → scanWithBackoff retries (up to 3 times)
safeScan returns ErrorDenied   → scanWithBackoff returns immediately (service skipped)
safeScan returns ErrorService  → scanWithBackoff returns immediately (error reported)
safeScan returns ErrorNone     → scanWithBackoff returns immediately (success)
```

### Testing Strategy

Tests should mock service behavior by controlling the `ErrorCategory` returned by `safeScan`. Since `scanWithBackoff` calls `safeScan` internally, create mock services that:
- Return throttling errors (AWS SDK `awserr.New("Throttling", ...)`) for N calls, then succeed
- Return throttling errors indefinitely (to test retry exhaustion)
- Return immediately with success/denied/error (to test no-retry paths)

**Race detection:** All tests run with `go test -race`. The backoff logic itself has no shared state, but integration through the worker pool must remain race-free.

**Timing tests:** Use short delays in tests (multiply constants by small factors or test `calculateBackoff` directly). Do NOT use production-length 1s base delays in unit tests — tests should complete fast.

For testing `scanWithBackoff` with controlled timing:
```go
// Override backoff constants for testing is not needed — test via calculateBackoff directly.
// For scanWithBackoff tests, use mock services that track call counts and
// return throttle errors for the first N calls.
type throttleMockService struct {
    callCount    int
    throttleFor  int // return throttle for first N calls
    name         string
}
```

**Note:** Since backoff constants are package-level, tests cannot easily override them. Instead, test `calculateBackoff` for correctness directly, and test `scanWithBackoff` behavior with understanding that retry delays will use real backoff timing. Keep test services that throttle to 0-1 retries needed to keep test duration reasonable, or accept that tests with 3 retries will take ~3-6 seconds.

Alternative: Make backoff configurable via a struct, but this adds complexity. The simpler approach is to test the math in `calculateBackoff` and test the retry logic behavior separately with acceptance of timing.

### Previous Story Intelligence (6.3)

- **Worker pool integration point:** `worker_pool.go` line 58: `serviceResults, _ := safeScan(ctx, service, sess, debug)` — change to `scanWithBackoff()`
- **Sequential integration point:** `main.go` line 306: same change
- **ErrorCategory `_` discard pattern:** Story 6.3 explicitly noted that ErrorCategory is discarded and "Story 6.4 adds retry logic based on ErrorThrottle". The `scanWithBackoff` function internally uses ErrorCategory but its callers still discard it.
- **Testing pattern:** Table-driven tests with stdlib `testing` (no testify in `cmd/awtest/` tests). Use `t.Errorf()` for soft assertions, `t.Fatalf()` for hard failures.
- **Atomic drain flag:** Worker pool sets `drained` flag via `sync/atomic` — workers stop appending after drain. Backoff retries should also check context before each retry, which naturally handles this case.
- **safeScan signature:** `func safeScan(ctx context.Context, service types.AWSService, sess *session.Session, debug bool) ([]types.ScanResult, ErrorCategory)`

### Git Intelligence

Recent commits:
- `2d02ab2` — Add concurrent worker pool execution with graceful drain (Story 6.3)
- `6bb8c12` — Add safeScan wrapper with panic recovery and error classification (Story 6.2)
- `02f3875` — Add --speed preset and concurrency flag resolution (Story 6.1)

**Patterns from recent work:**
- Files created in `cmd/awtest/`: `speed.go`, `speed_test.go`, `safe_scan.go`, `safe_scan_test.go`, `worker_pool.go`, `worker_pool_test.go`
- Test files use stdlib `testing` only (no testify assertions)
- `main.go` modified to integrate new features into `scanServices()`
- Each story follows clean layering: 6.1 (flags) → 6.2 (safeScan) → 6.3 (worker pool) → 6.4 (backoff)

### FRs Covered

- **FR74:** Exponential backoff on throttling (base 1s, 2x multiplier, jitter, max 3 retries)
- **FR75:** Per-service independent backoff (no shared state, each worker retries independently)
- **FR76:** Jitter prevents retry storms (±50% randomization on delay)
- **FR77:** Error classification for retry decisions (throttle → retry, denied → skip, error → report)
- **FR78:** Complete results for non-throttled services even when some throttle

### NFRs Addressed

- **NFR50:** Exponential backoff retries converge — throttled services eventually succeed or fail definitively, never retry infinitely (max 3 retries)
- **NFR51:** Rate limit backoff adds maximum 15 seconds total delay per service (worst case ~10.5s with jitter)
- **NFR46:** Fault isolation maintained — backoff per-service, throttling on one doesn't affect others

### Anti-Patterns to Avoid

- **DO NOT** create global or shared backoff state — all state must be local to each `scanWithBackoff` invocation
- **DO NOT** add a sleep without context-aware select — always use `select { case <-ctx.Done(): case <-time.After(): }`
- **DO NOT** retry on non-throttle errors — only `ErrorThrottle` triggers retry
- **DO NOT** add sync primitives inside any `services/` file
- **DO NOT** modify `safeScan` — `scanWithBackoff` wraps it, not replaces it
- **DO NOT** use `time.Sleep()` — use `time.After()` in a select for context awareness
- **DO NOT** make the retry loop unbounded — cap at `backoffMaxRetries`
- **DO NOT** expose backoff internals to callers — `scanWithBackoff` has the same external contract as `safeScan`

### Project Structure Notes

- New files in `cmd/awtest/`: `backoff.go`, `backoff_test.go`
- Modified files: `cmd/awtest/worker_pool.go` (line 58: safeScan → scanWithBackoff), `cmd/awtest/main.go` (line 306: safeScan → scanWithBackoff)
- Package: `main` (same as all other `cmd/awtest/` files)
- No changes to `services/`, `types/`, `formatters/`, or `utils/` packages
- No changes to `go.mod` (all Go stdlib: `math/rand`, `time`, `context`, `fmt`)

### References

- [Source: architecture-phase2.md#Rate Limit Backoff] — Per-service inline exponential backoff decision
- [Source: architecture-phase2.md#Concurrency Patterns] — Backoff formula: base 1s, 2x, ±50% jitter, 15s cap, 3 retries
- [Source: architecture-phase2.md#Error Handling Patterns] — 3-category error classification
- [Source: architecture-phase2.md#Implementation Handoff] — Backoff is step 5 in implementation sequence
- [Source: architecture-phase2.md#Process Patterns] — Graceful shutdown: retry checks context between attempts
- [Source: epics-phase2.md#Story 1.4] — BDD acceptance criteria (FR74-78)
- [Source: prd-phase2.md FR74-78] — Rate limit resilience requirements
- [Source: prd-phase2.md NFR50-51] — Bounded retry convergence, max 15s per service
- [Source: cmd/awtest/safe_scan.go:24-42] — classifyAWSError() with ErrorThrottle category
- [Source: cmd/awtest/safe_scan.go:45-66] — safeScan() function signature and behavior
- [Source: cmd/awtest/worker_pool.go:58] — safeScan call to replace with scanWithBackoff
- [Source: cmd/awtest/main.go:306] — safeScan call in sequential path to replace
- [Source: cmd/awtest/main.go:289-315] — scanServices() function
- [Source: 6-3-concurrent-worker-pool-execution.md#Error Category Handling] — "Story 6.4 (backoff) will wrap the safeScan call with retry logic based on ErrorThrottle"

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- All 13 backoff tests pass with `-race` flag (11 original + 2 review follow-ups)
- Full test suite (`go test -race ./...`) passes with zero regressions
- No sync primitives found in any `services/` package files

### Completion Notes List

- Created `cmd/awtest/backoff.go` with `calculateBackoff()` and `scanWithBackoff()` functions implementing exponential backoff with ±50% jitter, 3 max retries, 15s max delay cap, and context-aware sleep via select
- Created `cmd/awtest/backoff_test.go` with 13 comprehensive tests covering: backoff range validation, jitter variation, max delay cap, no-retry for non-throttle/denied/success, throttle-then-success retry, retry exhaustion, rate-limited error message, context cancellation, transparent success after retry, NFR51 total delay validation, and AC2 worker pool integration
- Integrated `scanWithBackoff` into both worker pool (`worker_pool.go:58`) and sequential scan path (`main.go:306`), replacing direct `safeScan` calls
- All backoff state is local per-invocation — no global/shared coordination
- ErrorCategory still discarded (`_`) at both call sites — classification used internally by backoff only
- Code review follow-ups addressed: fixed timer leak (time.NewTimer+Stop), added NFR51 total delay test, added AC2 integration test, added package doc comment, updated File List

### Change Log

- 2026-03-08: Story 6.4 implemented — rate limit resilience with exponential backoff (FR74-78, NFR50-51)
- 2026-03-08: Code review follow-ups — fixed timer leak, added NFR51 + AC2 tests, added package doc, updated File List

### File List

- `cmd/awtest/backoff.go` (new) — exponential backoff retry logic with package doc
- `cmd/awtest/backoff_test.go` (new) — 13 comprehensive backoff tests
- `cmd/awtest/worker_pool.go` (modified) — line 58: safeScan → scanWithBackoff
- `cmd/awtest/main.go` (modified) — line 306: safeScan → scanWithBackoff
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified) — story status → review
