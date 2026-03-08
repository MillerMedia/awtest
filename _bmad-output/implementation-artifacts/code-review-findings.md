# Code Review: Story 6.5 — Concurrent Progress Reporting

**Story:** 6-5-concurrent-progress-reporting  
**Git vs Story Discrepancies:** 1  
**Issues Found:** 0 Critical, 1 High, 3 Medium, 2 Low  

---

## Git vs Story Cross-Check

- **Story File List:** `cmd/awtest/progress.go`, `cmd/awtest/progress_test.go`, `cmd/awtest/worker_pool.go`, `cmd/awtest/worker_pool_test.go`, `cmd/awtest/backoff_test.go`, `cmd/awtest/main.go`, `go.mod`, `go.sum`
- **Git reality:** All listed application files are present (progress.go, progress_test.go untracked; rest modified). `_bmad-output/implementation-artifacts/sprint-status.yaml` is modified but **not** in the story File List.
- **Discrepancy:** `sprint-status.yaml` was changed (story 6.5 set to `review`) but is not listed in the story File List → MEDIUM (incomplete documentation). All claimed source files match git.

---

## 🔴 CRITICAL ISSUES

_None._

---

## 🟠 HIGH ISSUES

### 1. Progress line not padded — NFR40 / story “no flickering” violated

**Location:** `cmd/awtest/progress.go:58`

**Finding:** The story and Dev Notes require: “Pad with trailing spaces to ensure previous longer text is cleared.” The implementation uses:

```go
fmt.Fprintf(p.writer, "\rScanning... %d/%d services complete", count, p.total)
```

There is no padding. When the completed count goes from two digits to one (e.g. 15 → 9), the previous digit(s) remain on screen (e.g. “5” from “15”). Similarly, when total is two digits and count goes from 10 to 9, “1” can remain. This causes visible flicker/leftover characters and violates the explicit padding requirement and NFR40 (“without flickering”).

**Recommendation:** Use a fixed-width progress line, e.g. pad the message to a minimum length (e.g. 50 chars) so each update overwrites the previous one:

```go
msg := fmt.Sprintf("\rScanning... %d/%d services complete", count, p.total)
fmt.Fprintf(p.writer, "\r%-50s\r", msg)
```

(or equivalent) so the line is always fully overwritten.

---

## 🟡 MEDIUM ISSUES

### 2. File List does not include `sprint-status.yaml`

**Location:** Story Dev Agent Record → File List

**Finding:** When the story status was set to `review`, `sprint-status.yaml` was updated but is not listed in the File List. This hides part of the change set from reviewers and tooling.

**Recommendation:** Add `_bmad-output/implementation-artifacts/sprint-status.yaml` to the File List when it is updated for the story.

---

### 3. Non-TTY test does not enforce behavior (AC4)

**Location:** `cmd/awtest/progress_test.go:19-26` (`TestNewProgressReporterNonTTYReturnsNil`)

**Finding:** The test only logs; it does not assert that `p` is nil or non-nil. In environments where stderr is a TTY (e.g. some IDEs or terminals), the reporter would be non-nil and progress would be shown. The test would still pass, so AC4 (“Non-TTY suppresses progress”) is not enforced by this test.

**Recommendation:** Either assert `p == nil` when stderr is not a TTY (and skip or use a pipe when it is), or use a test that forces non-TTY (e.g. redirect stderr to a pipe or use a known non-TTY fd) and then assert `p == nil`.

---

### 4. Progress count can exceed “results” on drain (minor UX)

**Location:** `cmd/awtest/worker_pool.go:59-65`

**Finding:** `progress.Increment()` is called for every service that finishes processing, including when `drained == 1` (results not appended). So after a timeout/drain, the user can see “46/46 services complete” while the summary shows fewer accessible services (some results discarded). This is consistent with “services finished” but can be confusing.

**Recommendation:** Document this in Dev Notes or consider incrementing only when results are appended (so progress reflects “completed and kept”). Either behavior is defensible; the current one should be explicit.

---

## 🟢 LOW ISSUES

### 5. Possible final flicker when stopping progress

**Location:** `cmd/awtest/progress.go:66-73` (`Stop()`)

**Finding:** `Stop()` closes `p.done` then immediately writes the clear line. The ticker goroutine may still perform one more `Fprintf` after the clear (e.g. select chooses `ticker.C` before `p.done`). So the last thing on screen could briefly be the progress line again before the summary is printed.

**Recommendation:** Optional: add a short sleep or synchronize with the ticker goroutine (e.g. wait for a “stopped” signal) before writing the clear line to avoid this rare flicker.

---

### 6. `TestProgressWritesToStderr` only checks a manually built struct

**Location:** `cmd/awtest/progress_test.go:113-124`

**Finding:** The test verifies that a reporter constructed with `writer: os.Stderr` has `p.writer == os.Stderr`. It does not verify that `newProgressReporter()` actually assigns `os.Stderr` when progress is enabled. Behavior is correct in code review, but the test could be stronger by asserting that the reporter returned in a TTY+non-quiet scenario uses stderr (e.g. via pipe or by checking the field when not nil).

**Recommendation:** Low priority; add a test that creates a reporter in a scenario where it is non-nil and asserts the writer is stderr, or document that this test only checks the struct field.

---

## AC / Task Verification Summary

| AC / Task | Status | Notes |
|-----------|--------|--------|
| AC1 In-place progress at 2 Hz, no flickering | ⚠️ Partial | 500ms ticker ✓; missing padding (High #1) |
| AC2 Progress replaced by summary | ✓ | Stop() clears line; summary follows |
| AC3 Quiet suppresses progress | ✓ | newProgressReporter(..., true) → nil |
| AC4 Non-TTY suppresses progress | ✓ | isTerminal(stderr) checked; test weak (Medium #3) |
| AC5 Sequential preserves per-service output | ✓ | concurrency ≤ 1 path unchanged |
| AC6 Progress to stderr only | ✓ | writer = os.Stderr |
| AC7 Atomic progress counter | ✓ | atomic.AddInt64; Increment after append |
| Task 1 golang.org/x/term | ✓ | go.mod has v0.1.0, Go 1.19 |
| Task 2 progress.go | ⚠️ | All subtasks done except padding (High #1) |
| Task 3 Worker pool integration | ✓ | progress param, Increment() after append |
| Task 4 main.go scanServices | ✓ | Progress only in concurrent path |
| Task 5 progress_test.go | ✓ | 11 tests; Non-TTY test weak (Medium #3) |
| Task 6 make test, no sync in services/ | ✓ | All tests pass with -race; no sync in services/ |

---

## Next Steps

What should I do with these issues?

1. **Fix them automatically** — Add progress line padding, update story File List with sprint-status.yaml, strengthen Non-TTY test (e.g. pipe-based assert or skip when TTY).
2. **Create action items** — Add a “Review Follow-ups (AI)” subsection to the story Tasks/Subtasks with `[ ] [AI-Review][Severity] Description [file:line]`.
3. **Show me details** — Deep dive into specific issues.

Reply with **1**, **2**, or the issue number(s) to examine.
