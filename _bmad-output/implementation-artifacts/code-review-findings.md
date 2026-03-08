# Code Review: Story 6.1 — Speed Preset & Concurrency Flag Resolution

**Story:** 6-1-speed-preset-concurrency-flag-resolution  
**Git vs Story Discrepancies:** 0  
**Issues Found:** 1 High, 2 Medium, 3 Low  

---

## Git vs Story Cross-Check

- **Story File List:** `cmd/awtest/speed.go`, `cmd/awtest/speed_test.go`, `cmd/awtest/main.go`
- **Git changes:** `main.go` modified; `speed.go`, `speed_test.go` untracked (new). No app source files changed that are missing from the story. No false claims. ✓

---

## 🔴 HIGH ISSUES

### 1. Resolved concurrency not wired into scan flow

**Location:** `cmd/awtest/main.go:79–80`, `scanServices` call at line 206  

**Finding:** Task 3 says “Wire resolved concurrency value into existing scan flow (still sequential for now)”. The code only uses `speedResult` for the header line; it is never passed into the scan path. `scanServices(ctx, filteredSvcs, sess, *quiet, *debug)` has no concurrency parameter, so Story 6.3 will have to change both `main.go` and the signature of `scanServices` instead of only adding parallel execution.

**Recommendation:** Pass `speedResult.Concurrency` into `scanServices` now (e.g. add a `concurrency int` parameter and ignore it in the loop until 6.3). That satisfies “wire into existing scan flow” and keeps 6.3 to a single place (worker pool inside `scanServices`).

---

## 🟡 MEDIUM ISSUES

### 2. Invalid-speed test does not assert AC4 “listing valid presets”

**Location:** `cmd/awtest/speed_test.go:89–95`  

**Finding:** AC4 requires that `--speed=invalid` “exits with error listing valid presets (safe, fast, insane)”. The test only checks `errContains: "invalid speed preset"`. It does not assert that the error message contains the list of valid presets. If the error text were changed to drop “(valid presets: safe, fast, insane)”, the test would still pass.

**Recommendation:** Add a stricter check, e.g. `errContains: "valid presets: safe, fast, insane"` for the invalid/empty/uppercase cases, or a separate assertion that the error lists all three presets.

### 3. Mixed stdout vs stderr in startup header

**Location:** `cmd/awtest/main.go:82–90`  

**Finding:** Banner and “Version” use `fmt.Println` (stdout) while “Speed: …” uses `fmt.Fprintf(os.Stderr, ...)`. Redirecting only stdout (e.g. `awtest > out.txt`) leaves the speed line on the terminal; redirecting only stderr loses the banner. Behavior is consistent with the story (“stderr” for the speed line) but the split may surprise users.

**Recommendation:** Either document that the header is split (stdout vs stderr) or move the whole header to one stream for predictable redirection.

---

## 🟢 LOW ISSUES

### 4. Test reimplements `strings.Contains`

**Location:** `cmd/awtest/speed_test.go:204–216`  

**Finding:** Custom `contains` and `searchString` duplicate `strings.Contains`. Same-package tests can use the standard library.

**Recommendation:** Add `"strings"` to imports and use `strings.Contains(err.Error(), tt.errContains)`.

### 5. No “(custom)” indicator when only `--concurrency` is set

**Location:** `cmd/awtest/main.go:89`  

**Finding:** Dev Notes table says `--concurrency=10` (no `--speed`) should show “safe (custom)” for the preset. The code prints “Speed: safe (concurrency: 10)” with no “(custom)” label, so the preset name is not distinguished from the default safe preset.

**Recommendation:** Optional UX improvement: when `concurrencyExplicit` is true and `speedResult.Concurrency != presetConcurrency`, display the preset as e.g. “safe (custom)” or add a short note in the header.

### 6. `speed.go` has no package doc comment

**Location:** `cmd/awtest/speed.go:1`  

**Finding:** File contains exported symbols (`SpeedSafe`, `SpeedFast`, `SpeedInsane`, `SpeedResult`, `resolveSpeedAndConcurrency`) but no `// Package main` or file-level doc describing the speed preset resolution contract.

**Recommendation:** Add a one-line (or short) package/file comment describing speed preset resolution and its use from `main`.

---

## AC / Task Verification Summary

| AC / Task | Status | Notes |
|-----------|--------|--------|
| AC1 Default behavior | ✓ | No flags → safe, concurrency 1 |
| AC2 Speed preset mapping | ✓ | safe=1, fast=5, insane=20 in `speedPresets` |
| AC3 Concurrency override | ✓ | `flag.Visit` + `concurrencyExplicit`; override works |
| AC4 Invalid speed | ✓ | Error message lists valid presets; test could be stricter (see Medium #2) |
| AC5 Output header | ✓ | “Speed: {preset} (concurrency: {N})” on stderr before scan |
| AC6 Backward compatibility | ✓ | Phase 1 flags unchanged |
| AC7 Concurrency validation | ✓ | `validateConcurrency` used when `concurrencyExplicit` |
| Task 1 speed.go | ✓ | Constants, map, `resolveSpeedAndConcurrency` |
| Task 2 speed_test.go | ✓ | 14 table-driven cases; minor quality issues (Medium #2, Low #4) |
| Task 3 main.go | ✓ | Flag, resolution, placeholder removed, header; wiring incomplete (High #1) |
| Task 4 Verify | ✓ | `make test` and concurrency tests pass |

---

## Next Steps

What should I do with these issues?

1. **Fix them automatically** — Apply code and test changes for the above.
2. **Create action items** — Add “Review Follow-ups (AI)” to the story Tasks/Subtasks with `[ ] [AI-Review][Severity] Description [file:line]`.
3. **Show me details** — Deep dive into specific issues.

Reply with **1**, **2**, or the issue number(s) to examine.
