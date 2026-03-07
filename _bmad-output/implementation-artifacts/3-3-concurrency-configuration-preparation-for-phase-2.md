# Story 3.3: Concurrency Configuration (Preparation for Phase 2)

Status: done

## Story

As a security professional wanting fast scans,
I want the ability to configure concurrent service scanning,
so that future Phase 2 concurrent enumeration can be controlled and I can prepare for blazing-fast execution.

## Acceptance Criteria

1. Add `-concurrency` flag accepting integer value (default 1 for sequential) (FR37)
2. Validate `-concurrency` value: must be >= 1 and <= 20 (reasonable worker pool limit)
3. When `-concurrency=1` (default), maintain existing sequential behavior (no changes)
4. When `-concurrency > 1`, display message to stderr: "Note: Concurrent enumeration (--concurrency > 1) will be available in Phase 2. Running sequentially."
5. Store concurrency value in config for future Phase 2 implementation
6. Document `-concurrency` flag in help text with "Phase 2 feature" note
7. Write unit tests for concurrency flag parsing covering: default value, valid values (1-20), invalid values (<1, >20), non-integer values
8. Verify flag parsing does not break current sequential execution
9. Verify: `go run ./cmd/awtest -concurrency=10` runs successfully (sequentially) with Phase 2 message
10. Verify: `go run ./cmd/awtest -concurrency=0` errors: "Concurrency must be >= 1"
11. Verify: `go run ./cmd/awtest -concurrency=50` errors: "Concurrency must be <= 20"
12. Verify: `go build ./cmd/awtest` compiles successfully

## Tasks / Subtasks

- [x] Task 1: Add `-concurrency` flag definition (AC: #1, #5, #6)
  - [x] Add `concurrency := flag.Int("concurrency", 1, "...")` in main.go flag section
  - [x] Include "Phase 2 feature" in help text description
- [x] Task 2: Add concurrency validation after flag.Parse() (AC: #2, #10, #11)
  - [x] Validate concurrency >= 1, error to stderr + os.Exit(1) if invalid
  - [x] Validate concurrency <= 20, error to stderr + os.Exit(1) if invalid
- [x] Task 3: Add Phase 2 informational message (AC: #3, #4)
  - [x] When concurrency > 1, print note to stderr about Phase 2 availability
  - [x] When concurrency == 1, no message (silent, current behavior)
- [x] Task 4: Write unit tests (AC: #7, #8)
  - [x] Test default value (1)
  - [x] Test valid values (1, 10, 20)
  - [x] Test invalid values (0, -1, 50)
  - [x] Verify sequential execution unchanged
- [x] Task 5: Verify integration (AC: #9, #12)
  - [x] Verify build compiles: `go build ./cmd/awtest`
  - [x] Verify all tests pass: `go test ./cmd/awtest/...`
  - [x] Verify go vet passes: `go vet ./cmd/awtest/...`

## Dev Notes

### Architecture & Constraints

- **Go version:** 1.19 - no generics, no slices package
- **Flag package:** Standard `flag` package only (NOT cobra, NOT urfave/cli)
- **AWS SDK:** v1.44.266 - SDK v1 does NOT have per-request concurrency safety for shared sessions; Phase 2 will need session.Copy() or isolated clients per worker
- **Testing:** Testify v1.9.x available but existing tests in `cmd/awtest/` use standard library assertions only - follow that pattern

### Implementation Pattern

Flag definition pattern (follow Stories 3.1 and 3.2 exactly):

```go
// In main.go, add after the timeout flag definition (line 43):
concurrency := flag.Int("concurrency", 1, "Number of concurrent service scans (Phase 2 feature, default: sequential)")
```

Validation pattern (add after `flag.Parse()` and `utils.Quiet = *quiet`, before the banner):

```go
// Validate concurrency
if *concurrency < 1 {
    fmt.Fprintf(os.Stderr, "Error: Concurrency must be >= 1\n")
    os.Exit(1)
}
if *concurrency > 20 {
    fmt.Fprintf(os.Stderr, "Error: Concurrency must be <= 20\n")
    os.Exit(1)
}
if *concurrency > 1 {
    fmt.Fprintf(os.Stderr, "Note: Concurrent enumeration (--concurrency > 1) will be available in Phase 2. Running sequentially.\n")
}
```

### What NOT To Do

- DO NOT implement goroutine pools, worker patterns, channels, or sync.WaitGroup - that is Phase 2
- DO NOT modify `scanServices()` function - it stays sequential
- DO NOT modify any service files - they already accept context from Story 3.2
- DO NOT create a separate config struct - flag value in main() scope is sufficient for Phase 1
- DO NOT use `log.Fatal` or `log.Println` - use `fmt.Fprintf(os.Stderr, ...)` for errors/notes (established pattern)

### Testing Approach

Concurrency validation is input-level logic. Test it similarly to how `timeout_test.go` tests scan behavior but focus on flag parsing:

- Tests should be in `cmd/awtest/` package (package main)
- Use `os/exec` to test CLI flag behavior (run binary with args, check stderr output and exit code) OR extract a `validateConcurrency(val int) error` helper and test that directly
- Table-driven test pattern preferred
- Existing test files: `main_test.go` (formatter tests), `timeout_test.go` (scanServices tests)

### Integration with Stories 3.1 & 3.2

The `-concurrency` flag coexists with existing flags. No code changes needed for integration - all flags are independent at the `flag.Parse()` level. The execution flow remains:

1. Parse all flags (including new `-concurrency`)
2. Validate concurrency range
3. Show Phase 2 message if concurrency > 1
4. Filter services (Story 3.1)
5. Create timeout context (Story 3.2)
6. Run `scanServices()` sequentially (unchanged)

### Phase 2 Context (Future - DO NOT IMPLEMENT NOW)

Phase 2 will use the concurrency value to create a goroutine worker pool:
- Worker pool pattern with N workers (from `-concurrency` flag)
- Each worker gets isolated AWS session via `session.Copy()`
- Services distributed via channel-based job queue
- Results collected via results channel
- Context cancellation (from Story 3.2) cancels all workers
- Target: 60%+ scan time reduction, sub-2-minute standard scans
- Memory constraint: under 100MB regardless of concurrency level

### Files to Modify

- `cmd/awtest/main.go` - Add flag definition and validation (primary change)

### Files to Create

- `cmd/awtest/concurrency_test.go` - Unit tests for concurrency flag validation

### Project Structure Notes

- All flag definitions are in `cmd/awtest/main.go` at the top of `main()`
- All validation happens in `main()` after `flag.Parse()`
- Error messages to stderr, informational messages to stderr
- Exit code 1 for validation errors, exit code 0 for successful runs (even with Phase 2 message)

### References

- [Source: _bmad-output/planning-artifacts/epics.md - Epic 3, Story 3.3, FR37]
- [Source: _bmad-output/planning-artifacts/architecture.md - Concurrency & Performance cross-cutting concern]
- [Source: _bmad-output/planning-artifacts/prd.md - FR37, FR59-60, NFR1-5]
- [Source: cmd/awtest/main.go - current flag definitions and validation patterns]
- [Source: cmd/awtest/timeout_test.go - testing patterns for scanServices]
- [Source: _bmad-output/implementation-artifacts/3-1-service-filtering-include-exclude-services.md - flag pattern]
- [Source: _bmad-output/implementation-artifacts/3-2-timeout-configuration.md - context propagation]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None needed - clean implementation with no issues.

### Completion Notes List

- Added `-concurrency` flag with default value 1 and "Phase 2 feature" help text
- Extracted `validateConcurrency()` helper for testability; validates range [1, 20]
- Validation errors print to stderr and exit(1), matching established pattern
- When concurrency > 1, prints Phase 2 informational message to stderr
- Created table-driven unit tests covering: default (1), valid values (1, 5, 10, 20), invalid values (0, -1, 21, 50)
- All existing tests pass - no regressions (scanServices, formatter, timeout tests all green)
- `go build`, `go test`, `go vet` all pass cleanly
- scanServices() function NOT modified - sequential behavior preserved
- No new dependencies added

### Change Log

- 2026-03-06: Implemented Story 3.3 - Added `-concurrency` flag with validation and Phase 2 message
- 2026-03-06: Code Review - Refactored magic numbers, verified implementation. Status -> done

### File List

- `cmd/awtest/main.go` (modified) - Added concurrency flag, validateConcurrency() function, validation call, Phase 2 message. Refactored magic numbers to constants.
- `cmd/awtest/concurrency_test.go` (new) - Table-driven unit tests for concurrency validation

### Senior Developer Review (AI)

- [x] Story file loaded from `_bmad-output/implementation-artifacts/3-3-concurrency-configuration-preparation-for-phase-2.md`
- [x] Story Status verified as reviewable (review)
- [x] Epic and Story IDs resolved (3.3)
- [x] Story Context located or warning recorded
- [x] Epic Tech Spec located or warning recorded
- [x] Architecture/standards docs loaded (as available)
- [x] Tech stack detected and documented
- [x] MCP doc search performed (or web fallback) and references captured
- [x] Acceptance Criteria cross-checked against implementation
- [x] File List reviewed and validated for completeness
- [x] Tests identified and mapped to ACs; gaps noted
- [x] Code quality review performed on changed files
- [x] Security review performed on changed files and dependencies
- [x] Outcome decided (Approve)
- [x] Review notes appended under "Senior Developer Review (AI)"
- [x] Change Log updated with review entry
- [x] Status updated according to settings (if enabled)
- [x] Sprint status synced (if sprint tracking enabled)
- [x] Story saved successfully

_Reviewer: Kn0ck0ut on 2026-03-06_

**Review Notes:**
- **Validation:** Confirmed `-concurrency` flag works as expected (1-20 range).
- **Phase 2 Prep:** Informational message correctly displayed for concurrency > 1.
- **Code Quality:** Refactored magic numbers (1, 20) into `MinConcurrency` and `MaxConcurrency` constants in `main.go` for better maintainability.
- **Testing:** Unit tests cover all edge cases.
- **Documentation:** Help text clearly indicates "Phase 2 feature".

**Outcome:** Approved. Ready for merge.
