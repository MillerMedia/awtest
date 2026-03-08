# Story 6.2: safeScan Wrapper with Panic Recovery & Error Classification

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want each service scan wrapped with panic recovery and error classification,
So that one service crash doesn't lose results from other services, and errors are handled appropriately.

## Acceptance Criteria

1. **Panic recovery:** If a service scan panics during execution, safeScan catches it, converts it to an error result for that service, and other services are unaffected.
2. **No credential leakage in panics:** Panic error messages and stack traces never contain credential data (AWS keys, session tokens).
3. **Throttle classification:** AWS throttling responses (HTTP 429, `RequestLimitExceeded`, `Throttling`) are classified as "throttle" (eligible for retry by Story 6.4's backoff).
4. **Access denied classification:** Access denied responses (`AccessDeniedException`, `UnauthorizedAccess`, `UnauthorizedOperation`, `AccessDenied`, `AuthorizationError`) are classified as "denied" (service skipped, no error reported).
5. **Service error classification:** All other errors are classified as "error" (included in results as error entry).
6. **Normal execution passthrough:** When a service completes successfully, safeScan returns the result normally with no overhead.
7. **Credential safety under concurrency:** No credentials appear in panic stack traces or error logs (NFR43). safeScan scrubs stack trace output.

## Tasks / Subtasks

- [x] Task 1: Create `safe_scan.go` with error classification and safeScan wrapper (AC: #1-#7)
  - [x] Define `ErrorCategory` type and constants (`ErrorThrottle`, `ErrorDenied`, `ErrorService`)
  - [x] Implement `classifyAWSError(err error) ErrorCategory` using `awserr.Error` type assertion and `.Code()` matching
  - [x] Implement `safeScan(ctx context.Context, service types.AWSService, sess *session.Session, debug bool) ([]types.ScanResult, ErrorCategory)` with defer/recover
  - [x] On panic: return single error ScanResult with "service scan failed" message (no stack trace)
  - [x] On success: call `service.Call(ctx, sess)` then `service.Process(output, err, debug)`, classify any error, return results + category
  - [x] On error: classify via `classifyAWSError`, return results from Process + category
- [x] Task 2: Create `safe_scan_test.go` with comprehensive tests (AC: #1-#7)
  - [x] Test panic recovery: service that panics returns error result, not crash
  - [x] Test panic with string value (not error type)
  - [x] Test panic message does NOT contain credential-like strings
  - [x] Test throttle classification: `RequestLimitExceeded`, `Throttling` error codes
  - [x] Test access denied classification: `AccessDeniedException`, `UnauthorizedOperation`, `AccessDenied`, `AuthorizationError`
  - [x] Test service error classification: unknown error codes
  - [x] Test normal execution: successful service returns results with `ErrorCategory` = none/empty
  - [x] Test nil error returns no error category
  - [x] Test non-AWS error (plain Go error) classified as service error
- [x] Task 3: Update `scanServices()` in `main.go` to call `safeScan` instead of direct Call/Process (AC: #4, #6)
  - [x] Replace `service.Call(ctx, sess)` + `service.Process(output, err, debug)` with `safeScan(ctx, service, sess, debug)`
  - [x] For access denied results: classification only; filtering deferred to Stories 6.3/6.4 (preserves Phase 1 backward compatibility per Dev Notes recommendation)
  - [x] Preserve existing context cancellation check in the loop
  - [x] Preserve existing "Scanning [service_name]..." stderr output
- [x] Task 4: Verify backward compatibility and run tests (AC: #6)
  - [x] Run `make test` (includes `-race` flag) — all existing tests pass
  - [x] Verify sequential scan behavior unchanged at `--speed=safe`

## Dev Notes

### Architecture & Design Decisions

- **New file `cmd/awtest/safe_scan.go`:** Contains safeScan wrapper and error classification. Placed in `cmd/awtest/` alongside `main.go`, `speed.go`, and future `worker_pool.go`. Package: `main`. [Source: architecture-phase2.md#Project Structure]
- **No new dependencies:** Uses Go stdlib + existing `awserr` from AWS SDK v1. [Source: architecture-phase2.md#Phase 2 Technical Additions]
- **safeScan is the concurrency boundary:** Services called through safeScan, never directly from workers (Story 6.3). This story establishes the wrapper; Story 6.3's worker pool will consume it.
- **Error classification is separate from retry logic:** safeScan classifies errors but does NOT retry. Story 6.4 (backoff) will use the ErrorCategory to decide whether to retry.

### Function Signatures

```go
// ErrorCategory classifies AWS errors for retry/skip/report decisions
type ErrorCategory int

const (
    ErrorNone     ErrorCategory = iota // No error
    ErrorThrottle                       // Retry with backoff (Story 6.4)
    ErrorDenied                         // Skip service silently
    ErrorService                        // Report in results
)

// classifyAWSError determines how to handle an AWS API error
func classifyAWSError(err error) ErrorCategory

// safeScan wraps service execution with panic recovery and error classification
func safeScan(ctx context.Context, service types.AWSService, sess *session.Session, debug bool) ([]types.ScanResult, ErrorCategory)
```

### Error Classification Rules

| AWS Error Code | Category | Action |
|---|---|---|
| `RequestLimitExceeded` | Throttle | Eligible for retry (Story 6.4) |
| `Throttling` | Throttle | Eligible for retry (Story 6.4) |
| `TooManyRequestsException` | Throttle | Eligible for retry (Story 6.4) |
| `AccessDeniedException` | Denied | Skip, no error |
| `AccessDenied` | Denied | Skip, no error |
| `UnauthorizedOperation` | Denied | Skip, no error |
| `AuthorizationError` | Denied | Skip, no error |
| `UnauthorizedAccess` | Denied | Skip, no error |
| Non-AWS error (plain `error`) | Service | Report in results |
| Any other AWS error | Service | Report in results |
| `nil` error | None | Normal result |

**Note:** `InvalidAccessKeyId` and `InvalidClientTokenId` are NOT classified here — they are already handled as fatal errors in `utils.HandleAWSError()` (main.go aborts scan). safeScan should let these propagate through Process() as they do today.

### Existing Code Context

**Current scan loop (main.go:289-311):**
```go
func scanServices(ctx context.Context, svcs []types.AWSService, sess *session.Session, concurrency int, quiet, debug bool) ([]types.ScanResult, []string) {
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
            output, err := service.Call(ctx, sess)
            serviceResults := service.Process(output, err, debug)
            results = append(results, serviceResults...)
        }
    }
    return results, skippedServices
}
```

**Replace lines 305-306** (`output, err := service.Call(...)` and `serviceResults := service.Process(...)`) with `safeScan()` call.

**AWS error type assertion pattern (from utils/output.go:73-110):**
```go
if awsErr, ok := err.(awserr.Error); ok {
    awsErr.Code() // returns string like "AccessDeniedException"
}
```

**Import needed:** `github.com/aws/aws-sdk-go/aws/awserr`

### Panic Recovery Implementation

```go
func safeScan(ctx context.Context, service types.AWSService, sess *session.Session, debug bool) (results []types.ScanResult, category ErrorCategory) {
    defer func() {
        if r := recover(); r != nil {
            // DO NOT include stack trace or panic value details that might contain credentials
            results = []types.ScanResult{{
                ServiceName: service.Name,
                MethodName:  service.Name,
                Error:       fmt.Errorf("service scan failed: panic recovered"),
                Timestamp:   time.Now(),
            }}
            category = ErrorService
        }
    }()

    output, err := service.Call(ctx, sess)
    serviceResults := service.Process(output, err, debug)

    if err != nil {
        return serviceResults, classifyAWSError(err)
    }
    return serviceResults, ErrorNone
}
```

**Key pattern:** Named return values enable defer/recover to set the return value on panic. The panic value `r` is intentionally NOT included in the error message to prevent credential leakage (NFR43).

### Integration with scanServices

After this story, `scanServices` should look like:
```go
// In the default case of the select:
if !quiet {
    fmt.Fprintf(os.Stderr, "Scanning %s...\n", service.Name)
}
serviceResults, errCategory := safeScan(ctx, service, sess, debug)
if errCategory != ErrorDenied {
    results = append(results, serviceResults...)
}
```

**Important:** Access denied results are NOT added to results — this preserves Phase 1 behavior where denied services are printed to stderr (by `utils.HandleAWSError` inside `Process()`) but don't appear in formatted output.

**Note on ErrorDenied filtering:** Currently in Phase 1, access denied services DO produce ScanResult entries with Error set (Process creates them). The current formatters include these in output. If the intent is to match Phase 1 behavior exactly, do NOT filter on ErrorDenied — let Process handle it as before. The ErrorCategory is primarily for Story 6.4's retry logic and Story 6.3's worker pool. **Recommendation: do NOT change the filtering behavior in this story — just add classification. Let scanServices continue to append all results as before.** This preserves backward compatibility and defers behavioral changes to the stories that need them.

### Testing Strategy

Tests should create mock `types.AWSService` structs with custom Call/Process functions:

```go
// Mock service that panics
panicService := types.AWSService{
    Name: "TestPanic",
    Call: func(ctx context.Context, sess *session.Session) (interface{}, error) {
        panic("something went wrong")
    },
    Process: func(output interface{}, err error, debug bool) []types.ScanResult {
        return nil // Never reached
    },
}
```

For AWS error classification tests, create `awserr.New()` errors:
```go
import "github.com/aws/aws-sdk-go/aws/awserr"

throttleErr := awserr.New("Throttling", "Rate exceeded", nil)
deniedErr := awserr.New("AccessDeniedException", "Access denied", nil)
```

**Session for tests:** Use `session.Must(session.NewSession())` with mock config, or pass `nil` if safeScan doesn't use session directly (it passes it to service.Call).

### FRs Covered

- **FR77:** Error classification (throttle vs denied vs error)
- **FR96:** Read-only guarantee maintained (safeScan adds no write operations)
- **FR97:** No credential data in concurrent error paths or crash traces

### NFRs Addressed

- **NFR43:** Credential values never appear in crash stack traces or error logs
- **NFR46:** Individual service failures don't affect other services (fault isolation per goroutine foundation)
- **NFR48:** Concurrent process crashes recovered without crashing process or losing other results

### Project Structure Notes

- New files in `cmd/awtest/`: `safe_scan.go`, `safe_scan_test.go`
- Package: `main` (same as `speed.go`, `main.go`)
- No changes to `services/`, `types/`, `formatters/`, or `utils/` packages
- No changes to `go.mod` (awserr already available via existing AWS SDK dependency)

### Previous Story Intelligence (6.1)

- **Testing pattern:** Table-driven tests with struct fields `name`, `wantErr`, `errContains`. Use `t.Errorf()` for soft assertions, `t.Fatalf()` for hard failures.
- **No testify used in cmd/awtest tests:** Story 6.1 tests use stdlib `testing` only (no `assert` or `require`). Follow this pattern.
- **Flag detection pattern:** `flag.Visit()` used to detect explicit flag usage — not relevant to this story but shows Go idiom preferences.
- **Code review applied:** Story 6.1 required fixes for wiring values into function signatures and strengthening test assertions. Ensure safeScan is properly wired into scanServices from the start.

### References

- [Source: architecture-phase2.md#Core Architectural Decisions] — safeScan wrapper decision and rationale
- [Source: architecture-phase2.md#Error Handling Patterns] — 3-category error classification
- [Source: architecture-phase2.md#Concurrency Patterns] — Worker pool contract requiring safeScan
- [Source: architecture-phase2.md#Implementation Patterns] — Anti-patterns (no sync in services)
- [Source: epics-phase2.md#Story 1.2] — BDD acceptance criteria
- [Source: cmd/awtest/main.go:289-311] — Current scanServices() sequential loop
- [Source: cmd/awtest/types/types.go:22-32] — ScanResult struct definition
- [Source: cmd/awtest/types/types.go:39-44] — AWSService struct with Call/Process functions
- [Source: cmd/awtest/utils/output.go:73-110] — HandleAWSError with awserr.Error type assertion pattern
- [Source: cmd/awtest/formatters/yaml_formatter.go:69-79] — Existing defer/recover panic recovery pattern
- [Source: cmd/awtest/speed.go] — Story 6.1 code patterns and conventions
- [Source: cmd/awtest/speed_test.go] — Story 6.1 test patterns (table-driven, stdlib testing)
- [Source: go.mod] — Go 1.19, AWS SDK v1.44.266 (awserr available)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Created `safe_scan.go` with `ErrorCategory` type (ErrorNone, ErrorThrottle, ErrorDenied, ErrorService), `classifyAWSError()` function using awserr.Error type assertion and Code() matching, and `safeScan()` wrapper with defer/recover panic recovery
- Panic recovery uses named return values to set results on panic; panic value intentionally excluded from error message to prevent credential leakage (NFR43)
- Error classification covers 3 throttle codes, 5 access denied codes, and falls back to ErrorService for unknown AWS errors and plain Go errors
- Created comprehensive test suite covering all classification categories, panic recovery (including string panics and credential leakage verification), normal execution passthrough, and scanServices integration with panic isolation
- Updated `scanServices()` in main.go to use `safeScan()` instead of direct Call/Process — preserves backward compatibility by continuing to append all results regardless of ErrorCategory (per Dev Notes recommendation)
- All tests pass with `-race` flag, zero regressions

### Change Log

- 2026-03-07: Implemented safeScan wrapper with panic recovery and 3-category error classification (Story 6.2)

### File List

- cmd/awtest/safe_scan.go (new) — ErrorCategory type, classifyAWSError(), safeScan() wrapper
- cmd/awtest/safe_scan_test.go (new) — 20 test cases for classification and panic recovery
- cmd/awtest/main.go (modified) — scanServices() updated to use safeScan()
- _bmad-output/implementation-artifacts/sprint-status.yaml (modified) — status updated to review
