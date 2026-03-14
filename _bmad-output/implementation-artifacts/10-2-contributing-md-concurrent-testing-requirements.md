# Story 10.2: CONTRIBUTING.md Concurrent Testing Requirements

Status: done

<!-- Generated: 2026-03-13 by BMAD Create Story Workflow -->
<!-- Epic: 10 - Documentation & Contributor Enablement (Phase 2 Epic 5) -->
<!-- FRs: FR106 | Source: epics-phase2.md#Story 5.2 -->
<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a contributor,
I want CONTRIBUTING.md to document concurrent testing requirements,
So that I can write services that work correctly under parallel execution.

## Acceptance Criteria

1. **AC1:** The Testing Standards section documents the requirement to run `go test -race ./...` (or equivalently `make test`, which already includes `-race`) and explains that race detection catches data races in concurrent execution paths.

2. **AC2:** The Testing Standards section explains concurrent comparison testing expectations — that service scan results must be identical regardless of concurrency level (`--speed=safe` vs `--speed=insane`), and this is verified by automated comparison tests.

3. **AC3:** The rule that services must **not** import `sync`, `sync/atomic`, or any concurrency primitives is explicitly stated in the Code Standards section, with a brief explanation of why (services are concurrency-unaware by design).

4. **AC4:** The service implementation section explains that services are concurrency-unaware — the `safeScan` wrapper handles panic recovery and error classification, and the worker pool handles parallelism transparently. Contributors do not need to write any concurrency-specific code.

5. **AC5:** The Service Validation Checklist includes a new item verifying that the service does not import `sync` or `sync/atomic` packages.

6. **AC6:** The Service Validation Checklist includes a new item verifying that `make test` passes with the race detector (already the default, but making it explicit reminds contributors this is mandatory).

7. **AC7:** No structural changes to sections that don't relate to concurrent testing (Release Process, PR Title Format, Development Requirements, etc.).

## Tasks / Subtasks

- [x] Task 1: Add Concurrent Architecture section to Code Standards (AC: 3, 4)
  - [x] Add subsection explaining services are concurrency-unaware by design
  - [x] State the rule: services must NOT import `sync`, `sync/atomic`, or spawn goroutines
  - [x] Explain the `safeScan` wrapper as the concurrency boundary
  - [x] Note that the worker pool (`worker_pool.go`) handles all parallelism — contributors don't touch it

- [x] Task 2: Update Testing Standards section (AC: 1, 2)
  - [x] Document that `make test` runs with `-race` flag by default (race detector enabled)
  - [x] Explain what the race detector catches and why it matters for concurrent execution
  - [x] Document concurrent comparison testing: results must be identical at any concurrency level
  - [x] Add example of running race detection manually: `go test -v -race ./...`

- [x] Task 3: Update Service Validation Checklist (AC: 5, 6)
  - [x] Add checklist item: "Service does not import `sync` or `sync/atomic` packages"
  - [x] Add checklist item: "`make test` passes with race detector (default — do not use `-race=false`)"

- [x] Task 4: Update "Adding a New AWS Service" steps (AC: 4)
  - [x] Add note in step 9 (or as a new awareness step) that `make test` runs with race detection and services must pass cleanly
  - [x] Add brief note that services are automatically run concurrently by the worker pool — no concurrency code needed

- [x] Task 5: Verify no unintended changes (AC: 7)
  - [x] Confirm Release Process section unchanged
  - [x] Confirm PR Title Format and PR Description sections unchanged
  - [x] Confirm Development Requirements section unchanged
  - [x] Confirm all existing content preserved (additions only, no removals of existing guidance)

## Dev Notes

### This is a Documentation-Only Story

This story modifies **only CONTRIBUTING.md**. No Go code changes. No tests to write or run. The acceptance criteria focus on content accuracy and completeness of contributor documentation.

### Key Concurrent Architecture Concepts to Document

The Phase 2 concurrency architecture has a critical design principle that contributors must understand:

**Services are concurrency-unaware.** The entire concurrency layer is encapsulated in:
- `cmd/awtest/worker_pool.go` — spawns N worker goroutines, feeds services via buffered channel
- `cmd/awtest/safe_scan.go` — wraps each service call with `defer/recover` panic recovery and error classification (throttle/denied/error)
- `cmd/awtest/backoff.go` — per-service exponential backoff with jitter for throttled API calls
- `cmd/awtest/progress.go` — atomic counter + ticker goroutine for progress reporting
- `cmd/awtest/speed.go` — speed preset resolution (`safe`=1, `fast`=5, `insane`=20 workers)

Services in `services/*/calls.go` implement `Call()` and `Process()` exactly as before. They receive a `context.Context` and `*session.Session`, make AWS API calls, and return `[]ScanResult`. They never import `sync`, never spawn goroutines, never coordinate with other services.

**Why this matters for contributors:** A new service added by a contributor will automatically be executed concurrently when the user runs `--speed=fast` or `--speed=insane`. The contributor doesn't need to do anything special — but they must NOT break this contract by adding sync primitives, global mutable state, or goroutines inside their service.

### Race Detector — Already in `make test`

The Makefile already runs tests with the race detector:
```makefile
test: ## Run all tests with race detector and coverage
	go test -v -race -coverprofile=coverage.out ./...
```

So every `make test` run already catches race conditions. The CONTRIBUTING.md just needs to make contributors aware this is happening and why.

### Error Classification — Relevant Context for Contributors

The `safeScan` wrapper classifies AWS errors into 3 categories:
1. **Throttle** (retry with backoff): `RequestLimitExceeded`, `Throttling`, `TooManyRequestsException`
2. **Denied** (skip silently): `AccessDeniedException`, `AccessDenied`, `UnauthorizedOperation`, `AuthorizationError`
3. **Service Error** (report in results): everything else

Contributors don't need to handle this — `safeScan` does it automatically. But they should know that their service's `Process()` function should still use `utils.HandleAWSError` for error display formatting, as this operates at a different layer.

### Anti-Patterns to Document

These should be mentioned as things contributors must NOT do:
- Import `sync` or `sync/atomic` in a service file
- Use `go func()` or spawn goroutines inside `Call()` or `Process()`
- Write to stdout from within a service (only formatters write to stdout; progress writes to stderr)
- Add global mutable state that could be accessed by multiple goroutines
- Modify the shared `session.Session` in ways that aren't thread-safe (region changes are scoped per Call invocation)

### Existing CONTRIBUTING.md Structure to Preserve

```
1. Development Workflow (Prerequisites, Setup, Common Commands)
2. Adding a New AWS Service (10-step guide)
3. Code Standards (Naming, Error Handling, Testing, Documentation)
4. Service Validation Checklist
5. Pull Request Process (Title Format, Description, Review)
6. Release Process
7. Development Requirements
```

New concurrent content should be integrated into existing sections (Code Standards, Testing Standards, Validation Checklist) rather than added as a standalone section, to maintain the document's flow.

### Previous Story Intelligence

**From Story 10.1 (README Update — done):**
- Documentation-only story completed successfully
- Pattern: update existing content, add new subsections, preserve structure
- Commit message pattern: descriptive of what was changed and why

**From Phase 2 service stories (7.x, 8.x, 9.x):**
- All 17 new services follow the exact same pattern: `Call(ctx, sess)` + `Process(output, err, debug)`
- None of them import `sync` — validating the concurrency-unaware design
- All pass `make test` with race detector

### Git Intelligence

Recent commits follow pattern: `"Add [service] enumeration with N API calls (Story X.Y)"` or `"Mark Story X.Y as done"`. For this documentation story, suggested commit message:
- `"Update CONTRIBUTING.md with concurrent testing requirements (Story 10.2)"`

### Project Structure Notes

**Files to MODIFY:**
```
CONTRIBUTING.md  # The only file modified in this story
```

**Files to REFERENCE (DO NOT MODIFY):**
```
cmd/awtest/worker_pool.go          # Worker pool implementation — understand concurrency model
cmd/awtest/safe_scan.go            # safeScan wrapper — understand error classification and panic recovery
cmd/awtest/backoff.go              # Backoff implementation — understand retry logic
cmd/awtest/speed.go                # Speed preset resolution — understand concurrency levels
cmd/awtest/services/_template/     # Service template — verify it has no sync imports
Makefile                           # Verify make test includes -race flag
```

### References

- [Source: epics-phase2.md#Story 5.2: CONTRIBUTING.md Concurrent Testing Requirements] — BDD acceptance criteria
- [Source: prd-phase2.md#FR106] — CONTRIBUTING.md documents concurrent testing requirements for new service additions
- [Source: architecture-phase2.md#Concurrency Patterns] — Worker pool contract, service implementation contract
- [Source: architecture-phase2.md#Enforcement Guidelines] — Rules AI agents and contributors must follow
- [Source: architecture-phase2.md#Anti-Patterns] — Things contributors must NOT do
- [Source: cmd/awtest/worker_pool.go] — Buffered channel + fixed goroutine pool implementation
- [Source: cmd/awtest/safe_scan.go] — safeScan wrapper with panic recovery and error classification
- [Source: Makefile] — `make test` target already includes `-race` flag
- [Source: CONTRIBUTING.md] — Current contributor guide (needs concurrent testing additions)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered. Documentation-only story, no code or test execution required.

### Completion Notes List

- Task 1: Added "Concurrent Architecture — Service Contract" subsection under Code Standards. Documents concurrency-unaware design, the rule against importing sync/sync/atomic, safeScan wrapper explanation, worker pool role, and anti-patterns.
- Task 2: Added "Race Detection" and "Concurrent Comparison Testing" subsections under Testing Standards. Documents make test -race default, explains what race detector catches, provides manual run example, and documents result determinism requirement.
- Task 3: Added two new items to Service Validation Checklist: no sync imports check, and make test with race detector check.
- Task 4: Updated step 9 of "Adding a New AWS Service" to note race detector and automatic concurrent execution by worker pool.
- Task 5: Verified Release Process, PR Title Format, PR Description, and Development Requirements sections remain unchanged. All changes are additions only.
- Code Review Follow-up (HIGH): Fixed testify references — testify is not in go.mod and no services use it. Updated Testing Standards and Development Requirements to reference standard `testing` package. This was a pre-existing error in CONTRIBUTING.md.
- Code Review Follow-up (MEDIUM): Added clarifying note that all new services must include tests in step 6 of "Adding a New AWS Service".
- Code Review Follow-up (MEDIUM): README.md and sprint-status.yaml git changes are from Story 10.1 and workflow status tracking respectively — not from this story's implementation.

### Change Log

- 2026-03-13: Implemented all 5 tasks — updated CONTRIBUTING.md with concurrent testing requirements, service contract documentation, and validation checklist items (Story 10.2)
- 2026-03-13: Addressed code review findings — fixed testify references (HIGH), clarified test requirement for new services (MEDIUM)

### File List

- CONTRIBUTING.md (modified)
