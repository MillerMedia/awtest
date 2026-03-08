# Story 6.1: Speed Preset & Concurrency Flag Resolution

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to select `--speed=safe/fast/insane` or `--concurrency=N` to control scan parallelism,
So that I can choose the right speed-vs-OPSEC tradeoff for my engagement.

## Acceptance Criteria

1. **Default behavior preserved:** Running awtest without `--speed` or `--concurrency` defaults to `--speed=safe` (concurrency=1), identical to Phase 1 behavior.
2. **Speed preset mapping:** `--speed=safe` → concurrency 1, `--speed=fast` → concurrency 5, `--speed=insane` → concurrency 20.
3. **Concurrency override:** `--concurrency=N` overrides the speed preset's concurrency value when both are specified (e.g., `--concurrency=10 --speed=fast` → effective concurrency 10).
4. **Invalid speed validation:** `--speed=invalid` exits with error listing valid presets (safe, fast, insane).
5. **Output header display:** Scan output header displays "Speed: {preset} (concurrency: {N})" showing the effective speed preset and concurrency level.
6. **Backward compatibility:** All Phase 1 flags (`--format`, `--output-file`, `--services`, `--exclude-services`, `--timeout`, `--quiet`, `--debug`, credential flags) work unchanged alongside `--speed`.
7. **Concurrency validation preserved:** Existing `--concurrency` validation (1-20 range) still works, including when combined with `--speed`.

## Tasks / Subtasks

- [x] Task 1: Create `speed.go` with speed preset resolution logic (AC: #1, #2, #3, #4)
  - [x] Define speed preset constants (`SpeedSafe`, `SpeedFast`, `SpeedInsane`)
  - [x] Define preset-to-concurrency mapping (`safe=1`, `fast=5`, `insane=20`)
  - [x] Implement `resolveSpeedAndConcurrency()` function that takes speed flag string and concurrency flag int, returns effective concurrency + preset name + error
  - [x] Handle override logic: if `--concurrency` explicitly set AND `--speed` set, concurrency wins
  - [x] Handle validation: reject invalid speed preset values with descriptive error
  - [x] Handle default: no flags → safe preset, concurrency 1
- [x] Task 2: Create `speed_test.go` with comprehensive table-driven tests (AC: #1-#4, #7)
  - [x] Test default (no flags) → safe, concurrency=1
  - [x] Test each speed preset → correct concurrency mapping
  - [x] Test concurrency override scenarios
  - [x] Test invalid speed preset → error with valid options listed
  - [x] Test concurrency out-of-range with speed preset → error
  - [x] Test edge cases: `--concurrency=1 --speed=insane`, `--concurrency=20 --speed=safe`
- [x] Task 3: Update `main.go` to add `--speed` flag and integrate resolution (AC: #1, #5, #6)
  - [x] Add `--speed` flag definition (type: string, default: "safe")
  - [x] Call `resolveSpeedAndConcurrency()` after `flag.Parse()`
  - [x] Remove or update the Phase 2 placeholder message for `--concurrency > 1`
  - [x] Display "Speed: {preset} (concurrency: {N})" in scan output header (stderr, before scan begins)
  - [x] Wire resolved concurrency value into existing scan flow (still sequential for now — actual worker pool is Story 6.3)
- [x] Task 4: Verify backward compatibility (AC: #6, #7)
  - [x] Run full existing test suite (`make test`) — all pass
  - [x] Verify `--concurrency` validation still works via existing `concurrency_test.go`
  - [x] Manual smoke test: awtest with Phase 1 flags only (no `--speed`) behaves identically

## Dev Notes

### Architecture & Design Decisions

- **New file `cmd/awtest/speed.go`:** Contains all speed preset resolution logic. Follows architecture decision to place concurrency-related files in `cmd/awtest/` alongside `main.go`. [Source: architecture-phase2.md#Project Structure]
- **No new dependencies:** Uses Go stdlib only. No external packages needed for flag resolution. [Source: architecture-phase2.md#Phase 2 Technical Additions]
- **Go naming conventions:** Constants use `PascalCase` (`SpeedSafe`, `SpeedFast`, `SpeedInsane`), NOT `SCREAMING_SNAKE`. File names use `snake_case.go`. [Source: architecture-phase2.md#Naming Patterns]
- **This story does NOT implement concurrent execution.** It only resolves flags and displays the speed/concurrency in the output header. The worker pool (Story 6.3) will consume the resolved concurrency value later.

### Existing Code Context

- **`--concurrency` flag already defined** in `main.go:53` with `MinConcurrency=1`, `MaxConcurrency=20` constants and `validateConcurrency()` function (from Story 3.3).
- **Phase 2 placeholder message** in `main.go:71-73` prints a note when `--concurrency > 1` saying concurrent enumeration is a Phase 2 feature. This should be removed or replaced with actual speed resolution logic.
- **`concurrency_test.go` already exists** with table-driven tests for `validateConcurrency()`. Do NOT break these tests.
- **Scan output header** currently prints identity info (STS caller identity). Speed/concurrency display should be added after this existing output, writing to stderr.

### Flag Interaction Design

The `--speed` and `--concurrency` flags interact as follows:

| Scenario | Effective Preset | Effective Concurrency |
|----------|-----------------|----------------------|
| No flags | safe | 1 |
| `--speed=safe` | safe | 1 |
| `--speed=fast` | fast | 5 |
| `--speed=insane` | insane | 20 |
| `--concurrency=10` (no --speed) | safe (custom) | 10 |
| `--concurrency=10 --speed=fast` | fast (overridden) | 10 |
| `--speed=invalid` | ERROR | ERROR |

**Key detection challenge:** Go's `flag` package doesn't distinguish between "flag not set" and "flag set to default value." To detect whether `--concurrency` was explicitly provided (and should override `--speed`), use `flag.Visit()` to check if the flag was actually set on the command line. This is the standard Go pattern for detecting explicit flag usage.

### FRs Covered

- **FR68:** `--speed` flag with safe/fast/insane presets
- **FR69:** `--concurrency=N` numeric override (1-20)
- **FR70:** Speed preset to concurrency level mapping
- **FR100:** Speed flag validation (only valid presets accepted)
- **FR101:** `--concurrency` overrides `--speed` deterministically
- **FR102:** Output header displays speed preset and effective concurrency
- **FR103:** Phase 1 flag backward compatibility

### Project Structure Notes

- New files go in `cmd/awtest/`: `speed.go`, `speed_test.go`
- Aligns with architecture-phase2.md project structure showing `speed.go` in `cmd/awtest/`
- No changes to `services/`, `types/`, `formatters/`, or `utils/` packages
- No changes to `go.mod` (no new dependencies)

### References

- [Source: architecture-phase2.md#Naming Patterns] — Go naming conventions for Phase 2
- [Source: architecture-phase2.md#Project Structure & Boundaries] — File placement for speed.go
- [Source: architecture-phase2.md#Core Architectural Decisions] — Speed preset mapping rationale
- [Source: epics-phase2.md#Story 1.1] — Acceptance criteria and BDD scenarios
- [Source: prd-phase2.md#Command Structure] — Flag interaction rules and preset-to-concurrency table
- [Source: prd-phase2.md#Flag Interaction & Validation] — FR100-103 specifications
- [Source: cmd/awtest/main.go:25-28] — MinConcurrency/MaxConcurrency constants
- [Source: cmd/awtest/main.go:53] — Existing --concurrency flag definition
- [Source: cmd/awtest/main.go:67-73] — Existing concurrency validation and Phase 2 placeholder
- [Source: cmd/awtest/main.go:271-280] — validateConcurrency() function
- [Source: cmd/awtest/concurrency_test.go] — Existing concurrency validation tests (must not break)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered.

### Completion Notes List

- Created `speed.go` with `SpeedSafe`/`SpeedFast`/`SpeedInsane` constants, `speedPresets` map, `SpeedResult` struct, and `resolveSpeedAndConcurrency()` function
- Created `speed_test.go` with 14 table-driven test cases covering all AC scenarios (defaults, preset mappings, concurrency overrides, invalid presets, out-of-range concurrency)
- Updated `main.go`: added `--speed` flag (default: "safe"), `flag.Visit()` detection for explicit `--concurrency`, replaced Phase 2 placeholder with speed resolution call, added "Speed: {preset} (concurrency: {N})" to output header on stderr
- Full test suite passes (22 tests in cmd/awtest, all packages green, zero regressions)
- Code review fixes: wired `speedResult.Concurrency` into `scanServices` signature, moved banner to stderr for consistent stream usage, strengthened AC4 test assertions, replaced custom `contains` with `strings.Contains`, updated `timeout_test.go` for new `scanServices` signature

### File List

- `cmd/awtest/speed.go` (new) — Speed preset constants, mapping, and resolution logic
- `cmd/awtest/speed_test.go` (new) — 14 table-driven tests for speed resolution
- `cmd/awtest/main.go` (modified) — Added --speed flag, integrated resolution, removed Phase 2 placeholder, added header display, wired concurrency into scanServices, moved banner to stderr
- `cmd/awtest/timeout_test.go` (modified) — Updated scanServices calls to include concurrency parameter

### Change Log

- 2026-03-07: Implemented speed preset and concurrency flag resolution (Story 6.1)
- 2026-03-07: Applied code review fixes — concurrency wired into scan flow, consistent stderr output, stricter test assertions
