# Story 1.2: JSON Output Formatter

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security professional using awtest in automated workflows,
I want JSON-formatted scan results,
so that I can programmatically parse output for SIEM integration, reporting tools, and custom analysis scripts.

## Acceptance Criteria

**Given** the OutputFormatter interface from Story 1.1
**When** implementing JSON output format
**Then:**

1. **AC1:** Create JsonFormatter struct in `cmd/awtest/formatters/json_formatter.go` implementing OutputFormatter interface
2. **AC2:** `Format()` method produces valid JSON conforming to standard JSON schema (NFR20)
3. **AC3:** JSON includes all ScanResult fields (service, method, resource type, resource name, details, timestamp)
4. **AC4:** JSON structure uses camelCase field names following Go JSON conventions
5. **AC5:** `FileExtension()` returns `"json"`
6. **AC6:** Handle empty results gracefully (return valid empty JSON array `[]`)
7. **AC7:** Handle errors in `ScanResult.Error` field by including error message string in JSON output
8. **AC8:** Write table-driven unit tests in `json_formatter_test.go` covering: valid results, empty results, results with errors, timestamp formatting
9. **AC9:** Verify JSON output can be parsed by standard JSON tools (`jq`, Python `json.loads`)
10. **AC10:** Test passes: `go test ./cmd/awtest/formatters/...`

## Tasks / Subtasks

- [x] Validate existing JSON formatter implementation against all ACs (AC: 1-10)
  - [x] Run `go test ./cmd/awtest/formatters/...` and confirm all tests pass
  - [x] Run `go test -cover ./cmd/awtest/formatters/...` and confirm coverage meets 80%+ target
  - [x] Review json_formatter.go for completeness against AC1-AC7
  - [x] Review json_formatter_test.go for coverage against AC8
- [x] Verify JSON output parseability (AC: 9)
  - [x] Build binary: `go build ./cmd/awtest`
  - [x] Confirm JSON output is valid via `echo '[{"serviceName":"S3"}]' | jq .` pattern
  - [x] Confirm empty results produce `[]`
- [x] Fill any gaps identified during validation (AC: 1-10)
  - [x] Add missing test cases if any AC is not covered
  - [x] Fix any implementation issues found
- [x] Run full test suite and verify no regressions (AC: 10)
  - [x] `go test ./cmd/awtest/...`
  - [x] `go build ./cmd/awtest`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL CONTEXT: Implementation Already Exists

The JSON formatter was implemented ahead of schedule during Story 1.1. The following files already exist and are functional:

- `cmd/awtest/formatters/json_formatter.go` - Complete JSONFormatter implementation
- `cmd/awtest/formatters/json_formatter_test.go` - 9 table-driven test cases

**Story 1.1 Completion Notes stated:**
> "JSON formatter (json_formatter.go + json_formatter_test.go) was implemented ahead of schedule and is already tested and working. Nothing calls it yet, so zero risk."

**This story's primary task is VALIDATION, not creation.** The dev agent must:
1. Verify the existing implementation meets ALL acceptance criteria
2. Run tests and confirm they pass
3. Fill any gaps found
4. Mark the story as done once validated

### Existing Implementation Analysis

**json_formatter.go** (already implemented):
```go
package formatters

// jsonScanResult is the JSON-serializable representation of a ScanResult.
// Uses camelCase field names per Go JSON conventions and NFR20.
type jsonScanResult struct {
    ServiceName  string                 `json:"serviceName"`
    MethodName   string                 `json:"methodName"`
    ResourceType string                 `json:"resourceType"`
    ResourceName string                 `json:"resourceName"`
    Details      map[string]interface{} `json:"details"`
    Error        string                 `json:"error,omitempty"`
    Timestamp    string                 `json:"timestamp"`
}

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter { return &JSONFormatter{} }

func (f *JSONFormatter) Format(results []types.ScanResult) (string, error) {
    // Converts []ScanResult to []jsonScanResult with camelCase JSON tags
    // Uses json.MarshalIndent with 2-space indentation
    // Handles errors by converting to string via Error() method
    // Returns "[]" for empty results
}

func (f *JSONFormatter) FileExtension() string { return "json" }
```

**AC Compliance Checklist (Pre-Validated):**

| AC | Requirement | Status | Evidence |
|----|------------|--------|----------|
| AC1 | JsonFormatter in json_formatter.go | DONE | File exists, implements OutputFormatter |
| AC2 | Valid JSON (NFR20) | DONE | Uses encoding/json, MarshalIndent |
| AC3 | All ScanResult fields included | DONE | jsonScanResult maps all 7 fields |
| AC4 | camelCase field names | DONE | JSON tags: serviceName, methodName, etc. |
| AC5 | FileExtension() = "json" | DONE | Returns "json" |
| AC6 | Empty results = [] | DONE | Test "empty results returns empty array" passes |
| AC7 | Error field as string | DONE | omitempty tag, Error() conversion |
| AC8 | Table-driven tests | DONE | 9 test cases in table-driven format |
| AC9 | Parseable by jq/Python | NEEDS VERIFY | Tests use json.Unmarshal, but manual jq test needed |
| AC10 | Tests pass | NEEDS VERIFY | Must run `go test` to confirm |

### Existing Test Coverage

**json_formatter_test.go** contains 9 test cases:
1. `single valid result` - Validates JSON structure, camelCase fields
2. `empty results returns empty array` - Confirms `[]` output
3. `result with error` - Validates error serialization
4. `timestamp in RFC3339 format` - Confirms `2026-03-02T14:30:00Z` format
5. `multiple results` - Tests 3-result array serialization
6. `nil details serialized correctly` - Nil map handling
7. `successful result has no error field in JSON` - omitempty verification
8. `camelCase field naming verified` - Explicit field name checking (both positive and negative)
9. `TestJSONFormatter_FileExtension` - FileExtension() = "json"
10. `TestNewJSONFormatter` - Constructor not nil
11. `TestJSONFormatter_ImplementsInterface` - Compile-time interface check

**Coverage from Story 1.1:** formatters package at 91.7%

### Architecture Compliance

**From Architecture Document:**

- **Formatter Interface Pattern:** OutputFormatter with `Format()` and `FileExtension()` - FOLLOWED
- **Package:** `formatters` (lowercase, no underscores) - FOLLOWED
- **File naming:** `json_formatter.go` (underscores for multi-word) - FOLLOWED
- **Type naming:** `JSONFormatter` (PascalCase exported) - FOLLOWED
- **Constructor:** `NewJSONFormatter()` (Go convention) - FOLLOWED
- **Error handling:** Returns `fmt.Errorf("json formatting failed: %w", err)` - FOLLOWED
- **No new dependencies:** Uses stdlib `encoding/json` only - FOLLOWED

**JSON Output Schema:**
```json
[
  {
    "serviceName": "S3",
    "methodName": "s3:ListBuckets",
    "resourceType": "bucket",
    "resourceName": "my-bucket",
    "details": {"region": "us-east-1"},
    "timestamp": "2026-03-02T14:30:00Z"
  },
  {
    "serviceName": "EC2",
    "methodName": "ec2:DescribeInstances",
    "resourceType": "instance",
    "resourceName": "i-1234567890abcdef0",
    "details": {"state": "running"},
    "error": "AccessDeniedException: not authorized",
    "timestamp": "2026-03-02T14:30:01Z"
  }
]
```

**Key Design Decisions Already Made:**
- `jsonScanResult` struct used as serialization model (not direct ScanResult marshal)
- `error` field uses `omitempty` - absent when no error (cleaner JSON)
- Timestamps formatted as RFC3339 strings (not Unix epoch)
- Pretty-printed with 2-space indentation (human-readable)
- `error` type converted to string via `.Error()` method (errors are not JSON-serializable)

### Technical Requirements

**Go Version:** 1.19 (existing project standard)

**Dependencies:** None new required. Uses only:
- `encoding/json` (stdlib)
- `fmt` (stdlib)
- `github.com/MillerMedia/awtest/cmd/awtest/types` (internal)

**Build Verification:**
```bash
go build ./cmd/awtest          # Must succeed
go test ./cmd/awtest/formatters/...  # Must pass
go vet ./cmd/awtest/...        # Must pass
```

### File Structure

**Files that ALREADY EXIST (from Story 1.1):**
```
cmd/awtest/formatters/
├── output_formatter.go        # OutputFormatter interface
├── output_formatter_test.go   # Interface tests
├── json_formatter.go          # JSONFormatter implementation
└── json_formatter_test.go     # JSON formatter tests (9 cases)
```

**Files to LEAVE UNCHANGED:**
```
cmd/awtest/types/types.go      # ScanResult struct (from Story 1.1)
cmd/awtest/main.go             # Result collection loop (from Story 1.1)
cmd/awtest/services/           # All service files unchanged
go.mod                         # No new dependencies
```

### Previous Story Intelligence (Story 1.1)

**Key Learnings:**
- All 34 services now return `[]types.ScanResult` from their `Process` methods
- Backward compatibility maintained: `utils.PrintResult()` calls kept in Process methods
- Test coverage: types 100%, formatters 91.7%
- Fixed pre-existing vet error in `glue/calls.go` (`fmt.Sprintf("%s", workflowName)` with `*string`)
- Fixed unused variable `countStr` in `s3/calls.go`
- JSON formatter was implemented ahead of schedule with full test coverage

**Patterns Established:**
- Table-driven tests with `t.Run()` sub-tests
- Interface compliance checks via `var _ OutputFormatter = (*JSONFormatter)(nil)`
- Separate serialization struct (`jsonScanResult`) for JSON tags
- Constructor function pattern: `NewJSONFormatter()`

### Git Intelligence

**Recent Commits:**
```
6017a06 Add .gitignore, tests, and update service calls with result collection
c9d801d Add awtest source code to private working repo
c094ccf Sync BMAD artifacts 2026-03-02_20:57:24
```

**Patterns from Recent Work:**
- Small, incremental changes preferred
- Tests co-located with implementations
- Build verification after each change
- Backward compatibility maintained throughout

### Implementation Strategy

Since the implementation already exists, the dev agent should follow this streamlined approach:

1. **Verify Tests Pass** - Run `go test ./cmd/awtest/formatters/...` and confirm all 9+ tests pass
2. **Verify Coverage** - Run `go test -cover ./cmd/awtest/formatters/...` and confirm 80%+ coverage
3. **Review AC Compliance** - Walk through each AC and verify the implementation satisfies it
4. **Verify Build** - Run `go build ./cmd/awtest` and `go vet ./cmd/awtest/...`
5. **Fill Gaps** - If any AC is not fully covered, add tests or fix implementation
6. **Mark Done** - Update story status to done

**Expected Effort:** Minimal (1-2 hours) - primarily validation, not implementation.

### Edge Cases to Verify

1. **Empty results** → Should produce `[]` (not `null` or `""`)
2. **nil Details map** → Should serialize as `null` in JSON (not crash)
3. **Error with no other fields** → Should include error string, omit empty fields
4. **Special characters in ResourceName** → JSON encoding handles automatically
5. **Large result sets** → `json.MarshalIndent` handles efficiently
6. **Concurrent access** → JSONFormatter is stateless (no shared state), safe by design

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/`)
- Follows package-per-directory convention
- Tests co-located with implementation (`*_test.go` in same directory)
- Clear separation: `types` package for data, `formatters` package for presentation

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.2: JSON Output Formatter]
- [Source: _bmad-output/planning-artifacts/architecture.md#Output Format Architecture]
- [Source: _bmad-output/planning-artifacts/architecture.md#Implementation Patterns]
- [Source: _bmad-output/planning-artifacts/prd.md#FR43: Users can export scan results in structured JSON format]
- [Source: _bmad-output/planning-artifacts/prd.md#NFR20: JSON schema conformance for machine parsing]
- [Source: _bmad-output/implementation-artifacts/1-1-formatter-interface-result-collection.md#Completion Notes]
- [Source: cmd/awtest/formatters/json_formatter.go] - Existing implementation
- [Source: cmd/awtest/formatters/json_formatter_test.go] - Existing tests (9 cases)

## Dev Agent Record

### Agent Model Used

Story created by: claude-opus-4-6 (Claude Code CLI)
Implementation agent: claude-opus-4-6 (Claude Code CLI)

### Debug Log References

No debug issues encountered. All tests passed on first run.

### Completion Notes List

- Validated existing JSON formatter implementation (created during Story 1.1) against all 10 ACs — all passed
- All 18 tests pass across formatters package: 8 table-driven JSON formatter tests + 3 standalone JSON tests + 7 output_formatter interface tests
- Coverage: 91.7% (exceeds 80% target)
- JSON output verified parseable by `jq` — both empty array `[]` and full result arrays
- AC compliance: all 10 ACs satisfied (AC1-AC8 via code/test review, AC9 via jq verification, AC10 via test execution)
- Fixed pre-existing `go vet` issue in `cmd/awtest/utils/output.go:87` — unkeyed struct literal for `types.InvalidKeyError` changed to keyed field `Message:`
- No gaps found in implementation or test coverage — no new code or tests needed for the formatter itself
- Full regression suite passed: `go test ./cmd/awtest/...`, `go build ./cmd/awtest`, `go vet ./cmd/awtest/...`
- Addressed code review findings — 4 items resolved (Date: 2026-03-03):
  - [Medium] Switched `json.MarshalIndent` → `json.Marshal` for compact SIEM-friendly output
  - [Medium] Added per-result Details validation for resilient serialization — bad results get empty details + error marker instead of failing entire export
  - [Low] Replaced hardcoded `"2006-01-02T15:04:05Z07:00"` with `time.RFC3339` constant
  - [Low] Nil Details map now serializes as `{}` instead of `null` for consumer-friendliness
- Post-review: 20 tests pass (2 new: CompactOutput, ResilientSerialization), coverage 88.9% (>80% target), full regression green

### File List

- `cmd/awtest/formatters/json_formatter.go` (modified) — Compact output, resilient serialization, time.RFC3339, nil→{} details
- `cmd/awtest/formatters/json_formatter_test.go` (modified) — Updated nil details test, added compact output and resilient serialization tests
- `cmd/awtest/utils/output.go` (modified) — Fixed pre-existing `go vet` error: unkeyed struct literal → keyed field
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified) — Story status tracking updates
- `_bmad-output/implementation-artifacts/1-2-json-output-formatter.md` (modified) — Story task completion and Dev Agent Record

### Change Log

- 2026-03-03: Validated JSON formatter implementation against all ACs. Fixed pre-existing go vet issue in utils/output.go. All tests pass, 91.7% coverage. Story marked for review.
- 2026-03-03: Addressed 4 code review findings (2 medium, 2 low): compact JSON for SIEM, resilient serialization, time.RFC3339 constant, nil details as {}. 20 tests pass, 88.9% coverage.
