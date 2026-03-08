# Code Review: Story 6.4 — Rate Limit Resilience with Exponential Backoff

**Story:** 6-4-rate-limit-resilience-exponential-backoff  
**Git vs Story Discrepancies:** 1  
**Issues Found:** 0 Critical, 2 High, 2 Medium, 2 Low  

---

## Git vs Story Cross-Check

- **Story File List:** `cmd/awtest/backoff.go`, `cmd/awtest/backoff_test.go`, `cmd/awtest/worker_pool.go`, `cmd/awtest/main.go`
- **Git reality:** `backoff.go`, `backoff_test.go` untracked (new); `worker_pool.go`, `main.go` modified; `_bmad-output/implementation-artifacts/sprint-status.yaml` modified.
- **Discrepancy:** `sprint-status.yaml` was changed (story 6.4 set to `review`) but is not listed in the story File List → MEDIUM (incomplete documentation). All claimed application source files are present and match git changes.

---

## 🔴 CRITICAL ISSUES

_None._

---

## 🟠 HIGH ISSUES

### 1. Timer leak when context is cancelled during backoff sleep

**Location:** `cmd/awtest/backoff.go:59-65`

**Finding:** The code uses `time.After(delay)` in the select. When the context is cancelled, the select returns on `<-ctx.Done()` and the function exits without receiving from `time.After(delay)`. The timer created by `time.After` is not stopped and remains in the heap until it fires (after `delay`). Per Go docs, this can leak: "The underlying Timer is not recovered by the garbage collector until the timer fires."

**Evidence:**
```go
select {
case <-ctx.Done():
    return results, category
case <-time.After(delay):
    // Continue to next retry
}
```

**Recommendation:** Use `time.NewTimer(delay)`, defer `t.Stop()`, and use `t.C` in the select so the timer is stopped when returning on context cancellation.

---

### 2. No test that total backoff delay per service stays within NFR51 (15s)

**Location:** Story NFR51 / `backoff_test.go`

**Finding:** NFR51 states "maximum total delay per service is 15 seconds". The implementation caps each individual delay at 15s and uses at most 3 sleeps (attempts 0–2), so the sum is bounded in practice (~10.5s worst case). There is no test that asserts the sum of delays for one service never exceeds 15 seconds. A future change (e.g. cap bug or extra retry) could violate NFR51 without being caught.

**Recommendation:** Add a test that runs `scanWithBackoff` with a mock that records sleep durations (or test `calculateBackoff` for attempts 0,1,2 and assert sum ≤ 15s) to lock in NFR51.

---

## 🟡 MEDIUM ISSUES

### 3. File List does not include `sprint-status.yaml`

**Location:** Story Dev Agent Record → File List

**Finding:** When the story status was set to `review`, `sprint-status.yaml` was updated but is not listed in the File List. This makes it harder for reviewers and tooling to see all changed artifacts.

**Recommendation:** Add `_bmad-output/implementation-artifacts/sprint-status.yaml` to the File List when it is intentionally updated for the story.

---

### 4. Exhausted-retries test can take ~4–10 seconds and may slow CI

**Location:** `cmd/awtest/backoff_test.go:189-208` (`TestScanWithBackoffExhaustedRetries`)

**Finding:** The test uses production backoff constants (1s base, 2x, jitter). With 4 calls (initial + 3 retries), each with real sleeps, worst-case total is ~10.5s. The story notes "accept that tests with 3 retries will take ~3-6 seconds" but does not document this for CI or provide a short-delay variant.

**Recommendation:** Either document the expected test duration in the story/README, or add a build tag / test flag for a "short backoff" test variant that uses smaller delays for fast CI.

---

## 🟢 LOW ISSUES

### 5. Package-level documentation missing for `backoff.go`

**Location:** `cmd/awtest/backoff.go:1-20`

**Finding:** The file has no package comment describing the retry policy (base delay, multiplier, max retries, max delay, jitter). New maintainers must read the constants and code to understand the contract.

**Recommendation:** Add a short package or file comment, e.g. "Package main implements exponential backoff for throttled AWS calls: base 1s, 2x multiplier, ±50% jitter, max 3 retries, 15s cap per service (NFR51)."

---

### 6. No integration test that one throttled service does not block others (AC2)

**Location:** Story AC2 / test suite

**Finding:** AC2 requires "other concurrent services continue executing unblocked" when one service is throttling. Unit tests cover `scanWithBackoff` in isolation and per-service independence is inherent in the design (no shared state). There is no integration test that runs the worker pool with one service that throttles and others that succeed, and asserts both completion and that the non-throttled services' results are present.

**Recommendation:** Add an integration test (e.g. in `worker_pool_test.go` or `backoff_test.go`) that runs multiple services with one mock throttling and verifies all results and ordering.

---

## AC / Task Verification Summary

| AC / Task | Status | Notes |
|-----------|--------|--------|
| AC1 Exponential backoff (base 1s, 2x, jitter, 3 retries, 15s cap) | ✓ | calculateBackoff + scanWithBackoff; tests cover range, jitter, max cap |
| AC2 Per-service independent backoff | ✓ | No shared state; worker pool calls scanWithBackoff per service (no integration test) |
| AC3 Jitter prevents retry storms | ✓ | TestCalculateBackoffJitterVariation; formula [0.5, 1.5) |
| AC4 Transparent success after retry | ✓ | TestScanWithBackoffTransparentSuccess |
| AC5 Exhausted retries → rate-limited error | ✓ | TestScanWithBackoffExhaustedRetries + RateLimitedErrorMessage |
| AC6 Context cancellation abandons retries | ✓ | select ctx.Done(); TestScanWithBackoffContextCancellation (timer leak: High #1) |
| AC7 Complete results for non-throttled services | ✓ | Design + per-service backoff |
| Task 1 backoff.go | ✓ | Constants, calculateBackoff, scanWithBackoff, context-aware select |
| Task 2 Integrate into worker pool and sequential | ✓ | worker_pool.go:58, main.go:306 use scanWithBackoff |
| Task 3 backoff_test.go | ✓ | 11 tests; NFR51 total-delay test missing (High #2) |
| Task 4 make test, backward compatibility | ✓ | safe mode unchanged; no sync in services/ |

---

## Next Steps

What should I do with these issues?

1. **Fix them automatically** — Fix the timer leak (use NewTimer + Stop), add NFR51 test and package comment, update story File List with sprint-status.yaml.
2. **Create action items** — Add a "Review Follow-ups (AI)" subsection to the story Tasks/Subtasks with `[ ] [AI-Review][Severity] Description [file:line]`.
3. **Show me details** — Deep dive into specific issues.

Reply with **1**, **2**, or the issue number(s) to examine.
