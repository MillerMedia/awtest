---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
lastStep: 8
status: 'complete'
completedAt: '2026-03-07'
inputDocuments:
  - prd-phase2.md
  - architecture.md
  - prd.md
  - product-brief-awtest-2026-02-27.md
  - epics.md
workflowType: 'architecture'
project_name: 'awtest'
user_name: 'Kn0ck0ut'
date: '2026-03-07'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**

Phase 2 adds 41 new functional requirements (FR67-FR107) organized into 7 categories:

**Concurrent Enumeration (FR67-73):** Core architectural change — replace sequential service loop with configurable worker pool. Speed presets (safe=1, fast=5, insane=20 workers) map to concurrency levels. Results must be identical regardless of concurrency. Thread-safe result collection with deterministic output ordering by service name.

**Rate Limit Resilience (FR74-78):** Per-service exponential backoff with jitter. Throttling on one service must not block others. System distinguishes throttling (retry) from access denied (skip) and service errors (report). Complete results even under throttling at any concurrency level.

**Concurrent Progress Reporting (FR79-84):** In-place updating progress display for concurrent mode. Sequential mode preserves existing per-service output. Progress writes to stderr, suppressed in quiet mode or non-TTY environments.

**New Service Enumeration (FR85-95):** 11 new AWS services following established AWSService template pattern. No new architectural patterns required — pure service additions.

**Concurrent Safety (FR96-99):** Read-only guarantee maintained under concurrency. Credential safety under concurrent error paths. Timeout propagation and graceful drain of in-progress scans.

**Flag Interaction (FR100-103):** Speed preset and concurrency flag validation, conflict resolution, header display.

**Documentation (FR104-107):** README, CONTRIBUTING.md, and service template updates for concurrent-aware development.

**Non-Functional Requirements:**

Phase 2 adds 25 NFRs (NFR35-NFR59) across 5 categories:

**Performance (NFR35-41):** Sub-30s at insane, sub-60s at fast, no regression at safe. 80%+ linear scaling efficiency. <100MB memory. Progress updates at 2+ Hz. <100ms formatting overhead.

**Security (NFR42-45):** Read-only guarantee under concurrency. No credential leakage through crash paths. Thread-safe session sharing. No external data transmission.

**Reliability (NFR46-51):** Fault isolation per goroutine. Sequential/concurrent result parity. Panic recovery. Clean goroutine termination within 1s. Bounded retry convergence. Max 15s backoff per service.

**Integration (NFR52-55):** Session concurrency verification. New services follow existing interface. Alphabetical registration. Formatter compatibility.

**Maintainability (NFR56-59):** Concurrency encapsulated in worker pool — services remain concurrency-unaware. 70%+ coverage including race detection. Concurrent comparison test suite in standard make test.

### Technical Constraints & Dependencies

**Brownfield Foundation (Phase 1 Architecture):**
- Go 1.19, AWS SDK v1 (`github.com/aws/aws-sdk-go v1.44.266`)
- AWSService struct with Call/Process pattern — 46 services implemented
- Result collection via `[]ScanResult` with OutputFormatter interface
- Flag-based configuration, GoReleaser distribution, testify testing
- Context-aware service calls with timeout and cancellation already in place
- `--concurrency` flag already defined (default 1, range 1-20)

**Concurrency Constraints:**
- AWS SDK v1 `session.Session` must be verified thread-safe for concurrent goroutine access
- Result slice requires mutex or channel-based protection
- Progress counter requires atomic operations
- Panic recovery must be per-goroutine to prevent process crash
- Context cancellation must propagate from main timeout to all workers

**Backward Compatibility:**
- `--speed=safe` (default) must produce identical behavior to Phase 1
- All existing flags unchanged
- Output format/structure unchanged regardless of speed preset
- Exit codes unchanged

### Cross-Cutting Concerns Identified

**1. Thread-Safe Result Collection:**
Spans all concurrent service executions. Results arrive out of order, must be collected safely, sorted deterministically by service name before formatting.

**2. Panic Recovery & Fault Isolation:**
Each goroutine must recover from panics independently. One service crash cannot affect other concurrent services or lose their results.

**3. Context Propagation & Cancellation:**
Main timeout context must propagate to all workers. On cancellation, in-progress services drain before output formatting begins.

**4. Rate Limit Backoff Isolation:**
Per-service backoff state — no global coordination. Throttled services retry independently without blocking other concurrent services.

**5. Progress Tracking:**
Atomic counter incremented as each service completes. In-place terminal update at 2+ Hz. Suppressed in non-TTY/quiet mode.

**6. Credential Safety Under Concurrency:**
Shared session read-only access. No credential leakage through concurrent error paths, panic stack traces, or goroutine lifecycle.

## Starter Template Evaluation

### Primary Technology Domain

**CLI Tool** — Go-based AWS security enumeration CLI, established in Phase 1 with 46 services implemented.

### Brownfield Foundation Assessment

This is a brownfield project extending an existing, production-ready codebase. No starter template evaluation is needed — the existing codebase serves as the architectural foundation.

### Existing Technical Stack (Phase 1 Established)

**Language & Runtime:**
- Go 1.19
- Single binary compilation, cross-platform (darwin/linux/windows, amd64/arm64)

**Core Dependencies:**
- `github.com/aws/aws-sdk-go v1.44.266` — AWS SDK v1 for all service calls
- `github.com/spf13/cobra v1.7.0` — CLI framework and flag parsing
- `github.com/stretchr/testify v1.8.2` — Test assertions

**Build Tooling:**
- GoReleaser for cross-platform builds and GitHub releases
- Makefile for development workflow (`make build`, `make test`, `make lint`)
- GitHub Actions for CI/CD release automation

**Testing Framework:**
- Go standard `testing` package with testify assertions
- `go test ./...` with race detection support (`-race` flag)
- Table-driven test patterns established

**Code Organization:**
- `cmd/awtest/main.go` — Entry point
- `services/` — AWSService implementations (one per file, alphabetically registered)
- `types/` — Shared types (`ScanResult`, `AWSService`, `OutputFormatter`)
- `utils/` — Shared utilities (session management, formatting)

**Development Experience:**
- `go build` / `go run` for local development
- `make test` for test execution
- `golangci-lint` for linting

### Phase 2 Technical Additions Required

The existing foundation supports Phase 2 with these additions:

- **Concurrency primitives**: `sync.WaitGroup`, `sync.Mutex`, `sync/atomic` — all stdlib, no new dependencies
- **Context propagation**: `context.Context` already used for timeouts — extends to worker cancellation
- **Progress display**: Terminal control via ANSI escape codes or minimal dependency — stdlib capable

**No new external dependencies anticipated** for core concurrency features. The Go standard library provides all required concurrency primitives.

### Rationale

The Phase 1 codebase provides a proven, production-released foundation. All Phase 2 features (concurrent execution, rate limiting, progress reporting, new services) layer onto existing patterns using Go's stdlib concurrency primitives. Introducing a new framework or starter would be counterproductive — the architectural decisions are already made and validated through Phase 1 delivery.

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
1. Concurrency model — worker pool design
2. Result collection mechanism
3. Panic recovery strategy

**Important Decisions (Shape Architecture):**
4. Rate limit backoff strategy
5. Progress reporting architecture

**Deferred Decisions (Not Applicable):**
- Data Architecture — N/A (no database, read-only AWS enumeration)
- Authentication — N/A (delegates to AWS credential chain)
- Frontend Architecture — N/A (CLI tool)

### Concurrency Model

**Decision:** Buffered channel + fixed goroutine pool

Pre-spawn N worker goroutines (where N = concurrency level from speed preset or `--concurrency` flag). Feed services to workers via a buffered channel. Workers pull services from the channel until it's closed.

**Rationale:** Zero external dependencies — uses Go stdlib only (`sync`, `sync/atomic`). Provides explicit control over worker lifecycle, natural backpressure via channel buffering, and clean drain semantics on context cancellation. Bounded resource usage by design.

**Affects:** Core scan loop, service execution, all cross-cutting concerns

### Result Collection

**Decision:** Mutex-protected shared slice

Workers append `ScanResult` entries to a shared `[]ScanResult` under `sync.Mutex`. After all workers complete, results are sorted by service name for deterministic output.

**Rationale:** Simplest approach that matches the existing `[]ScanResult` pattern. Critical section is minimal (single append), contention negligible at ≤20 workers. No additional goroutines or channels needed.

**Affects:** Result formatting, output ordering, all formatters

### Rate Limit Backoff

**Decision:** Per-service inline exponential backoff with jitter

Each worker handles its own retry loop when encountering throttling (HTTP 429 / RequestLimitExceeded). Exponential backoff with randomized jitter, capped at 15s per retry, max 3 retries. Backoff state is local to each service execution — no shared coordination.

**Rationale:** PRD requires per-service isolation — throttling on one service must not block others. Inline backoff keeps retry logic self-contained within each worker's service execution. No shared state means no contention.

**Affects:** Service execution within workers, error classification (throttle vs denied vs error)

### Progress Reporting

**Decision:** Atomic counter + ticker goroutine

`sync/atomic` counter incremented as each service completes. A dedicated ticker goroutine writes in-place progress updates to stderr at ≥2Hz using ANSI escape codes (`\r` carriage return for in-place update).

**Conditions:**
- Concurrent mode only (sequential preserves existing per-service output)
- Suppressed when `--quiet` flag set or stdout is not a TTY
- Ticker goroutine started before workers, stopped after all workers complete

**Rationale:** Lock-free increment on the hot path (atomic). Rendering decoupled from workers via independent ticker. Clean separation of concerns.

**Affects:** stderr output, TTY detection, quiet mode interaction

### Panic Recovery

**Decision:** Wrapper function (`safeScan`)

A `safeScan(service AWSService, ...)` function wraps each service execution with `defer/recover`. On panic, converts the panic to an error result (service marked as errored, not missing). Provides a single testable location for recovery logic and consistent error reporting.

**Rationale:** More testable than inline defers scattered across workers. Single location for panic-to-error conversion, logging, and credential scrubbing from stack traces. Each worker calls `safeScan` for every service, ensuring uniform fault isolation.

**Affects:** Worker goroutine implementation, error reporting, fault isolation

### Decision Impact Analysis

**Implementation Sequence:**
1. `safeScan` wrapper — foundational, used by everything else
2. Worker pool with channel-based dispatch — core concurrency engine
3. Mutex-protected result collection — connects workers to output
4. Inline backoff within `safeScan` — resilience layer
5. Atomic progress counter + ticker — observability layer

**Cross-Component Dependencies:**
- Worker pool → `safeScan` (workers call safeScan for each service)
- Worker pool → result mutex (workers append results)
- Worker pool → atomic counter (workers increment on completion)
- Ticker goroutine → atomic counter (reads count for display)
- All components → context (cancellation propagation)

## Implementation Patterns & Consistency Rules

### Critical Conflict Points Identified

7 areas where AI agents could make different implementation choices when working on Phase 2 concurrency features and new services.

### Naming Patterns

**Go Code Naming:**
- Functions: `PascalCase` for exported, `camelCase` for unexported — standard Go conventions
- Variables: `camelCase` — e.g., `workerCount`, `scanResult`, `backoffDuration`
- Constants: `PascalCase` for exported, `camelCase` for unexported — NOT `SCREAMING_SNAKE`
- Files: `snake_case.go` — e.g., `worker_pool.go`, `safe_scan.go`, `progress_reporter.go`
- Test files: `snake_case_test.go` co-located with source — e.g., `worker_pool_test.go`

**Service File Naming:**
- One file per service: `services/<service_name>.go` — lowercase, no underscores unless multi-word
- Follow existing pattern: `acm.go`, `cognito.go`, `elasticache.go`, `stepfunctions.go`

**Speed Preset Naming:**
- Constants: `SpeedSafe`, `SpeedFast`, `SpeedInsane`
- Flag value strings: `"safe"`, `"fast"`, `"insane"` (lowercase in CLI)

### Structure Patterns

**New File Placement:**
- Concurrency engine (`worker_pool.go`, `safe_scan.go`): in project root alongside `main.go` scan logic, or in a new `scanner/` package
- Progress reporting: alongside worker pool code
- Backoff logic: alongside `safeScan` (part of the scan execution path)
- New services: `services/` directory, alphabetically registered in `GetServices()`

**Test Organization:**
- Co-located: `*_test.go` next to source files
- Race detection tests: standard tests run with `-race` flag, no separate test files
- Concurrency comparison tests: dedicated test that runs scan at concurrency=1 and concurrency=5, compares results

**No New Packages Unless Necessary:**
- Prefer adding files to existing package structure over creating new packages
- If concurrency code grows beyond 3 files, consider a `scanner/` package

### Error Handling Patterns

**Error Classification (3 categories):**
1. **Throttling** (retry): HTTP 429, `RequestLimitExceeded`, `Throttling` — triggers backoff retry
2. **Access Denied** (skip): `AccessDeniedException`, `UnauthorizedAccess` — skip service, no error reported
3. **Service Error** (report): all other errors — include in results as error entry

**Error Response in Results:**
- Throttled + eventually succeeded: normal result, no error indication
- Throttled + exhausted retries: error result with "rate limited" message
- Access denied: omitted from results (consistent with Phase 1 behavior)
- Panic recovered: error result with "service scan failed" message (no stack trace in output)

**Logging vs Output:**
- Errors in scan results → stdout (via formatter)
- Progress/status → stderr
- Debug/diagnostic → stderr, only with `--verbose` (if added) or suppressed

### Concurrency Patterns

**Worker Pool Contract:**
- Workers receive services via `chan AWSService`, process until channel closed
- Each service execution wrapped in `safeScan` — no exceptions
- Workers never modify shared state except through mutex-protected result append and atomic progress increment
- Workers respect context cancellation — check `ctx.Done()` between retries

**Service Implementation Contract (Unchanged):**
- Services remain concurrency-unaware — no sync primitives inside service code
- Services receive their own `*session.Session` (or shared read-only session)
- Services use the existing `Call`/`Process` pattern
- New services follow the `SERVICE_TEMPLATE.go` pattern exactly

**Backoff Implementation:**
- Base delay: 1 second
- Multiplier: 2x per retry
- Jitter: ±50% randomization on each delay
- Max delay cap: 15 seconds
- Max retries: 3
- Formula: `min(baseDelay * 2^attempt * (0.5 + rand()), 15s)`

### Process Patterns

**Scan Execution Flow:**
1. Parse flags, resolve speed preset → concurrency level
2. Create AWS session, build service list (applying include/exclude filters)
3. Start progress ticker (if concurrent + TTY + not quiet)
4. Spawn N workers, feed services via channel
5. Wait for all workers to complete (`sync.WaitGroup`)
6. Stop progress ticker
7. Sort results by service name
8. Format and output results

**Graceful Shutdown:**
- On context cancellation (timeout): workers finish current service or abandon at next retry check
- Close service channel to prevent new work dispatch
- `WaitGroup.Wait()` with 1-second deadline — then proceed to output with partial results
- Never `os.Exit()` from a worker — only main goroutine controls exit

### Enforcement Guidelines

**All AI Agents MUST:**
- Run `go test -race ./...` after any concurrency-related changes
- Register new services alphabetically in `GetServices()`
- Never add sync primitives inside service implementation files
- Use `safeScan` wrapper for all service executions — no direct calls
- Follow existing `ScanResult` structure — no new result types

**Anti-Patterns:**
- Adding `sync.Mutex` inside a service file (services must be concurrency-unaware)
- Using `go func()` without `defer recover()` (use `safeScan` instead)
- Writing to stdout from worker goroutines (only the formatter writes to stdout)
- Global backoff state shared across services
- Spawning goroutines inside service `Call`/`Process` methods

## Project Structure & Boundaries

### Complete Project Directory Structure

**Existing files (Phase 1)** shown normally. **New Phase 2 additions** marked with `← NEW`.

```
awtest/
├── .github/
│   └── workflows/
│       └── release.yml
├── .gitignore
├── .goreleaser.yaml
├── cmd/
│   └── awtest/
│       ├── main.go                          # Entry point, scan orchestration
│       ├── concurrency_test.go              # Concurrency flag validation tests
│       ├── timeout_test.go                  # Timeout configuration tests
│       ├── worker_pool.go                   ← NEW (worker pool, channel dispatch, WaitGroup)
│       ├── worker_pool_test.go              ← NEW (pool tests with race detection)
│       ├── safe_scan.go                     ← NEW (safeScan wrapper, panic recovery)
│       ├── safe_scan_test.go                ← NEW (panic recovery, error classification tests)
│       ├── backoff.go                       ← NEW (exponential backoff with jitter)
│       ├── backoff_test.go                  ← NEW (backoff timing, retry logic tests)
│       ├── progress.go                      ← NEW (atomic counter, ticker, TTY detection)
│       ├── progress_test.go                 ← NEW (progress display tests)
│       ├── speed.go                         ← NEW (speed preset resolution, flag validation)
│       ├── speed_test.go                    ← NEW (preset mapping, conflict tests)
│       ├── formatters/
│       │   ├── output_formatter.go          # OutputFormatter interface
│       │   ├── output_formatter_test.go
│       │   ├── text_formatter.go
│       │   ├── text_formatter_test.go
│       │   ├── json_formatter.go
│       │   ├── json_formatter_test.go
│       │   ├── yaml_formatter.go
│       │   ├── yaml_formatter_test.go
│       │   ├── csv_formatter.go
│       │   ├── csv_formatter_test.go
│       │   ├── table_formatter.go
│       │   └── table_formatter_test.go
│       ├── services/
│       │   ├── _template/                   # SERVICE_TEMPLATE for new services
│       │   ├── services.go                  # GetServices() registry
│       │   ├── service_filter.go            # Include/exclude filtering
│       │   ├── service_filter_test.go
│       │   ├── amplify/calls.go
│       │   ├── apigateway/calls.go
│       │   ├── appsync/calls.go
│       │   ├── batch/calls.go
│       │   ├── certificatemanager/calls.go
│       │   ├── cloudformation/calls.go
│       │   ├── cloudfront/calls.go
│       │   ├── cloudtrail/calls.go
│       │   ├── cloudwatch/calls.go
│       │   ├── codepipeline/calls.go
│       │   ├── cognitoidentity/calls.go
│       │   ├── cognitouserpools/calls.go
│       │   ├── config/calls.go
│       │   ├── dynamodb/calls.go
│       │   ├── ec2/calls.go
│       │   ├── ecs/calls.go
│       │   ├── efs/calls.go
│       │   ├── eks/calls.go
│       │   ├── elasticache/calls.go
│       │   ├── elasticbeanstalk/calls.go
│       │   ├── eventbridge/calls.go
│       │   ├── fargate/calls.go
│       │   ├── glacier/calls.go
│       │   ├── glue/calls.go
│       │   ├── iam/calls.go
│       │   ├── iot/calls.go
│       │   ├── ivs/calls.go
│       │   ├── ivschat/calls.go
│       │   ├── ivsrealtime/calls.go
│       │   ├── kms/calls.go
│       │   ├── lambda/calls.go
│       │   ├── rds/calls.go
│       │   ├── redshift/calls.go
│       │   ├── rekognition/calls.go
│       │   ├── route53/calls.go
│       │   ├── s3/calls.go
│       │   ├── secretsmanager/calls.go
│       │   ├── ses/calls.go
│       │   ├── sns/calls.go
│       │   ├── sqs/calls.go
│       │   ├── stepfunctions/calls.go
│       │   ├── sts/calls.go
│       │   ├── systemsmanager/calls.go
│       │   ├── transcribe/calls.go
│       │   ├── vpc/calls.go
│       │   ├── waf/calls.go
│       │   │   # ── Phase 2 New Services (11) ──
│       │   ├── athena/calls.go              ← NEW
│       │   ├── backup/calls.go              ← NEW
│       │   ├── codecommit/calls.go          ← NEW
│       │   ├── codedeploy/calls.go          ← NEW
│       │   ├── directconnect/calls.go       ← NEW
│       │   ├── emr/calls.go                 ← NEW
│       │   ├── guardduty/calls.go           ← NEW
│       │   ├── kinesis/calls.go             ← NEW
│       │   ├── mediaconvert/calls.go        ← NEW
│       │   ├── neptune/calls.go             ← NEW
│       │   └── opensearch/calls.go          ← NEW
│       ├── types/
│       │   ├── types.go                     # ScanResult, AWSService structs
│       │   ├── types_test.go
│       │   ├── summary.go
│       │   └── summary_test.go
│       └── utils/
│           └── output.go                    # Session creation, output helpers
├── CONTRIBUTING.md
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
└── README.md
```

### Architectural Boundaries

**Concurrency Boundary:**
- Concurrency logic lives in `cmd/awtest/` (worker_pool.go, safe_scan.go, backoff.go, progress.go, speed.go)
- Services in `services/*/calls.go` are **concurrency-unaware** — they never import `sync` or `sync/atomic`
- The boundary is enforced by the `safeScan` wrapper — services are called through it, never directly from workers

**Formatter Boundary:**
- Formatters receive completed, sorted `[]ScanResult` — they are unaware of concurrency
- No changes to formatter interface or implementations for Phase 2
- Formatters only write to the provided `io.Writer` (stdout)

**Service Boundary:**
- Each service is a self-contained package under `services/<name>/`
- Services implement `Call()` and `Process()` on `types.AWSService`
- Services registered alphabetically in `services/services.go` via `GetServices()`
- New Phase 2 services follow identical pattern — no concurrency awareness

### Requirements to Structure Mapping

**Concurrent Enumeration (FR67-73) →** `worker_pool.go`, `speed.go`
**Rate Limit Resilience (FR74-78) →** `backoff.go`, `safe_scan.go`
**Progress Reporting (FR79-84) →** `progress.go`
**New Services (FR85-95) →** `services/{athena,backup,codecommit,...}/calls.go`
**Concurrent Safety (FR96-99) →** `safe_scan.go`, `worker_pool.go`
**Flag Interaction (FR100-103) →** `speed.go`, `main.go`
**Documentation (FR104-107) →** `README.md`, `CONTRIBUTING.md`, `_template/`

### Cross-Cutting Concern Mapping

| Concern | Files |
|---------|-------|
| Thread-safe results | `worker_pool.go` (mutex), `main.go` (sort before format) |
| Panic recovery | `safe_scan.go` (defer/recover wrapper) |
| Context cancellation | `main.go` (parent ctx), `worker_pool.go` (propagation), `backoff.go` (check between retries) |
| Rate limit isolation | `backoff.go` (per-invocation state), `safe_scan.go` (error classification) |
| Progress tracking | `progress.go` (atomic counter + ticker) |
| Credential safety | `utils/output.go` (session creation), `safe_scan.go` (stack trace scrubbing) |

### Data Flow

```
main.go: parse flags → resolve speed preset → create session → build service list
    ↓
worker_pool.go: spawn N workers ← chan AWSService ← service list
    ↓
safe_scan.go: wrap each service call (panic recovery + error classification)
    ↓
backoff.go: retry on throttling (per-service, inline)
    ↓
services/*/calls.go: Call() → AWS API → Process() → ScanResult
    ↓
worker_pool.go: mutex-append result → atomic-increment progress
    ↓
main.go: sort results → formatter.Format() → stdout
```

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:** All architectural decisions use Go stdlib concurrency primitives (`sync`, `sync/atomic`, channels). No external dependency conflicts. Worker pool → safeScan → backoff chain is internally consistent with no circular dependencies. Atomic progress counter operates independently of mutex-protected result collection — no deadlock risk.

**Pattern Consistency:** Naming follows Go conventions consistently. All new source files use `snake_case.go`. Services remain concurrency-unaware — boundary enforced by safeScan wrapper. Error classification categories (throttle/denied/error) map directly to behavioral responses (retry/skip/report).

**Structure Alignment:** New concurrency files placed in `cmd/awtest/` alongside existing scan orchestration. Service packages unchanged. Formatter packages unchanged. No new package boundaries needed.

### Requirements Coverage Validation ✅

**Functional Requirements (FR67-FR107):** All 41 FRs mapped to specific architectural components. Every FR category has at least one dedicated new file. Cross-cutting FRs (concurrent safety) addressed through safeScan wrapper pattern.

**Non-Functional Requirements (NFR35-NFR59):** All 25 NFRs architecturally supported. Performance NFRs addressed by worker pool scaling. Security NFRs covered by read-only session sharing and credential scrubbing. Reliability NFRs handled by panic recovery and bounded backoff. Maintainability NFRs ensured by concurrency encapsulation pattern.

### Implementation Readiness Validation ✅

**Decision Completeness:** All 5 critical decisions documented with rationale, affected components, and ordered implementation sequence. No ambiguous choices remain.

**Structure Completeness:** Every new file named, placed, and mapped to requirements. Data flow documented from flag parsing through worker dispatch to formatted output.

**Pattern Completeness:** Naming conventions, error classification, backoff formula (with specific parameters), scan execution flow, graceful shutdown sequence, enforcement guidelines, and anti-patterns all specified.

### Gap Analysis Results

**Critical Gaps:** None.

**Important Gaps (Non-Blocking):**
1. **TTY Detection Dependency:** `golang.org/x/term` required for `IsTerminal()` on Go 1.19 (stdlib `os.IsTerminal()` added in Go 1.21). This is a minor, well-established dependency.
2. **Per-Service API Details:** Specific AWS API calls for 11 new services not detailed in architecture — appropriately deferred to individual story specifications.

**Nice-to-Have:**
- `make test-race` convenience target in Makefile

### Architecture Completeness Checklist

**✅ Requirements Analysis**
- [x] Project context thoroughly analyzed (brownfield, Phase 1 foundation)
- [x] Scale and complexity assessed (≤20 concurrent workers, ≤57 services)
- [x] Technical constraints identified (Go 1.19, AWS SDK v1, backward compatibility)
- [x] Cross-cutting concerns mapped (6 concerns to specific files)

**✅ Architectural Decisions**
- [x] Critical decisions documented (5 decisions with rationale)
- [x] Technology stack fully specified (Go stdlib concurrency, no new major deps)
- [x] Integration patterns defined (safeScan wrapper, channel dispatch)
- [x] Performance considerations addressed (atomic ops, bounded workers)

**✅ Implementation Patterns**
- [x] Naming conventions established (Go standard, speed preset constants)
- [x] Structure patterns defined (file placement, test co-location)
- [x] Error handling patterns specified (3-category classification)
- [x] Process patterns documented (scan flow, graceful shutdown, backoff formula)

**✅ Project Structure**
- [x] Complete directory structure defined (all existing + new files)
- [x] Component boundaries established (concurrency/service/formatter)
- [x] Integration points mapped (safeScan as boundary, channel as dispatch)
- [x] Requirements to structure mapping complete (FR→file table)

### Architecture Readiness Assessment

**Overall Status:** READY FOR IMPLEMENTATION

**Confidence Level:** High — brownfield project with proven Phase 1 patterns, stdlib-only concurrency additions, clear boundaries.

**Key Strengths:**
- Zero new external dependencies for core concurrency (except `golang.org/x/term` for TTY)
- Services completely isolated from concurrency concerns
- Deterministic output regardless of concurrency level
- Clean implementation sequence with no circular dependencies

**Areas for Future Enhancement:**
- Go version upgrade (1.19 → 1.21+) would eliminate `golang.org/x/term` dependency
- AWS SDK v2 migration would provide built-in retry/backoff (separate initiative)
- Structured logging could replace stderr progress for machine-readable output

### Implementation Handoff

**AI Agent Guidelines:**
- Follow all architectural decisions exactly as documented
- Use implementation patterns consistently across all components
- Respect concurrency/service/formatter boundaries
- New services must be concurrency-unaware — never import `sync`
- All service executions must go through `safeScan` — no direct calls
- Run `go test -race ./...` after any concurrency-related changes

**First Implementation Priority:**
1. `safeScan` wrapper (foundation for all concurrent execution)
2. Worker pool with channel dispatch (core engine)
3. Speed preset resolution and flag validation
4. Progress reporting
5. Rate limit backoff
6. New service additions (can parallelize with above)
7. Documentation updates
