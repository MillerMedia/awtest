# Story 1.4: CSV Output Formatter

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security professional analyzing scan results in spreadsheets,
I want CSV-formatted scan results,
so that I can import findings into Excel/Google Sheets for filtering, sorting, and pivot table analysis.

## Acceptance Criteria

**Given** the OutputFormatter interface from Story 1.1
**When** implementing CSV output format
**Then:**

1. **AC1:** Create CsvFormatter struct in `cmd/awtest/formatters/csv_formatter.go` implementing OutputFormatter interface
2. **AC2:** `Format()` method produces valid CSV with header row: `Service,Method,ResourceType,ResourceName,Details,Timestamp,Error`
3. **AC3:** CSV escapes special characters (commas, quotes, newlines) following RFC 4180
4. **AC4:** Details field flattens `map[string]interface{}` to semicolon-separated `key=value` pairs (semicolons avoid conflict with CSV comma delimiter)
5. **AC5:** `FileExtension()` returns `"csv"`
6. **AC6:** Handle empty results gracefully (return CSV with header row only)
7. **AC7:** Handle errors in `ScanResult.Error` field by populating Error column
8. **AC8:** Write table-driven unit tests in `csv_formatter_test.go` covering: valid results, empty results, results with special characters requiring escaping, results with complex Details maps
9. **AC9:** Verify CSV output can be parsed by Go's `encoding/csv` Reader (round-trip test)
10. **AC10:** Test passes: `go test ./cmd/awtest/formatters/...`

## Tasks / Subtasks

- [x] Create CSVFormatter implementation (AC: 1, 2, 3, 4, 5, 6, 7)
  - [x] Create `cmd/awtest/formatters/csv_formatter.go`
  - [x] Define `CSVFormatter` struct (stateless, like JSON/YAML formatters)
  - [x] Implement `NewCSVFormatter()` constructor
  - [x] Implement `Format()` using `encoding/csv` Writer with `bytes.Buffer`
  - [x] Write header row: Service,Method,ResourceType,ResourceName,Details,Timestamp,Error
  - [x] Convert each ScanResult to CSV record with proper field mapping
  - [x] Implement `flattenDetails()` helper to convert `map[string]interface{}` to `key=value;key=value` string
  - [x] Handle nil Details as empty string `""`
  - [x] Handle Error field: populate column with `error.Error()` string, empty string when nil
  - [x] Format Timestamp as RFC3339 string
  - [x] Implement resilient Details serialization (catch unserializable values)
  - [x] Implement `FileExtension()` returning `"csv"`
- [x] Write comprehensive tests (AC: 8, 9)
  - [x] Create `cmd/awtest/formatters/csv_formatter_test.go`
  - [x] Test: single valid result with all fields
  - [x] Test: empty results returns header row only
  - [x] Test: result with error field populates Error column
  - [x] Test: special characters (commas, quotes, newlines) properly escaped per RFC 4180
  - [x] Test: multiple results produce correct row count
  - [x] Test: nil details produces empty Details column
  - [x] Test: timestamp in RFC3339 format
  - [x] Test: successful result has empty Error column
  - [x] Test: complex Details map with multiple key-value pairs
  - [x] Test: CSV round-trip (write then parse with csv.Reader)
  - [x] Test: interface compliance check
  - [x] Test: constructor not nil
  - [x] Test: FileExtension returns "csv"
  - [x] Test: resilient serialization with unserializable Details values
  - [x] Test: Details value ordering is deterministic (sorted keys)
- [x] Verify build and tests pass (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/formatters/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -cover ./cmd/awtest/formatters/...` (target: 80%+)
- [x] Run full regression (AC: 10)
  - [x] `go test ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow the Established Formatter Pattern

The CSV formatter must mirror the JSON/YAML formatter implementation patterns established in Stories 1.1-1.3. These were refined through code review and represent the canonical approach.

**Pattern Sources:**
- `cmd/awtest/formatters/json_formatter.go` — canonical Format() pattern
- `cmd/awtest/formatters/yaml_formatter.go` — resilient serialization with panic recovery

**Key Pattern Elements:**
1. **Stateless struct** — `CSVFormatter` has no fields (like JSONFormatter, YAMLFormatter)
2. **Constructor function** — `NewCSVFormatter()` returns `*CSVFormatter`
3. **Nil Details → empty string** — Never let nil Details produce unexpected CSV output
4. **Error field as string** — Empty string when no error (not omitempty since CSV always has the column)
5. **Timestamp as RFC3339 string** — Use `time.RFC3339` constant
6. **Error wrapping** — `fmt.Errorf("csv formatting failed: %w", err)`
7. **Interface compliance** — `var _ OutputFormatter = (*CSVFormatter)(nil)`

### CSV-Specific Implementation Notes

**Use Go's `encoding/csv` package** — Do NOT manually construct CSV strings. The stdlib `csv.Writer` handles RFC 4180 escaping automatically (quoting fields with commas, double-quoting embedded quotes, handling newlines).

**Implementation Pattern:**
```go
package formatters

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "sort"
    "strings"
    "time"

    "github.com/MillerMedia/awtest/cmd/awtest/types"
)

type CSVFormatter struct{}

func NewCSVFormatter() *CSVFormatter {
    return &CSVFormatter{}
}

func (f *CSVFormatter) Format(results []types.ScanResult) (string, error) {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)

    // Write header
    header := []string{"Service", "Method", "ResourceType", "ResourceName", "Details", "Timestamp", "Error"}
    if err := writer.Write(header); err != nil {
        return "", fmt.Errorf("csv formatting failed: %w", err)
    }

    // Write data rows
    for _, r := range results {
        record := []string{
            r.ServiceName,
            r.MethodName,
            r.ResourceType,
            r.ResourceName,
            flattenDetails(r.Details),
            r.Timestamp.Format(time.RFC3339),
            formatError(r),
        }
        if err := writer.Write(record); err != nil {
            return "", fmt.Errorf("csv formatting failed: %w", err)
        }
    }

    writer.Flush()
    if err := writer.Error(); err != nil {
        return "", fmt.Errorf("csv formatting failed: %w", err)
    }
    return buf.String(), nil
}

func (f *CSVFormatter) FileExtension() string {
    return "csv"
}
```

**Details Flattening Strategy:**
- Convert `map[string]interface{}` to `key=value` pairs separated by semicolons: `region=us-east-1;count=5`
- Sort keys alphabetically for deterministic output (critical for testing and diff-ability)
- Use `fmt.Sprintf("%v", value)` for value conversion (handles all Go types)
- Nil Details → empty string `""`
- Empty Details map → empty string `""`
- Resilient: if a value causes issues, skip it and note error

**Error Column Handling:**
- `ScanResult.Error == nil` → empty string `""`
- `ScanResult.Error != nil` → `r.Error.Error()` string
- If Details serialization fails, append serialization error message (same pattern as JSON/YAML)

**RFC 4180 Compliance:**
- Go's `encoding/csv` handles this automatically:
  - Fields containing commas are quoted: `"field,with,commas"`
  - Fields containing double quotes are escaped: `"field ""with"" quotes"`
  - Fields containing newlines are quoted: `"field\nwith\nnewlines"`
  - CRLF line endings in output (per RFC 4180)
- Do NOT override csv.Writer defaults — they are RFC 4180 compliant

**CSV Header Order:**
```
Service,Method,ResourceType,ResourceName,Details,Timestamp,Error
```
This matches the ScanResult struct field order (ServiceName, MethodName, ResourceType, ResourceName, Details, Timestamp, Error) for consistency with JSON/YAML field naming.

### Architecture Compliance

**From Architecture Document:**

- **Formatter Interface Pattern:** OutputFormatter with `Format()` and `FileExtension()` — MUST FOLLOW
- **Package:** `formatters` (lowercase, no underscores) — MUST FOLLOW
- **File naming:** `csv_formatter.go` (underscores for multi-word) — MUST FOLLOW
- **Type naming:** `CSVFormatter` (PascalCase exported, CSV all-caps per Go conventions) — MUST FOLLOW
- **Constructor:** `NewCSVFormatter()` (Go convention) — MUST FOLLOW
- **Error handling:** Returns `fmt.Errorf("csv formatting failed: %w", err)` — MUST FOLLOW
- **No new dependencies** — `encoding/csv` is Go stdlib
- **Testing:** stdlib `testing` package (no testify)

### Technical Requirements

**Go Version:** 1.19 (existing project standard)

**Dependencies (NO NEW DEPENDENCIES):**
- `encoding/csv` (stdlib — already available)
- `bytes` (stdlib — for buffer)
- `fmt`, `sort`, `strings`, `time` (stdlib)

**Existing Dependencies (unchanged):**
- `github.com/aws/aws-sdk-go v1.44.266`
- `github.com/logrusorgru/aurora v2.0.3+incompatible`
- `gopkg.in/yaml.v3 v3.0.1`

**Build Verification:**
```bash
go build ./cmd/awtest             # Must succeed
go test ./cmd/awtest/formatters/... # Must pass
go vet ./cmd/awtest/...           # Must pass
```

### File Structure

**Files to CREATE:**
```
cmd/awtest/formatters/
├── csv_formatter.go          # NEW: CSVFormatter implementation
└── csv_formatter_test.go     # NEW: CSV formatter tests
```

**Files that ALREADY EXIST (DO NOT MODIFY):**
```
cmd/awtest/formatters/
├── output_formatter.go        # OutputFormatter interface (from Story 1.1)
├── output_formatter_test.go   # Interface tests (from Story 1.1)
├── json_formatter.go          # JSONFormatter (from Stories 1.1/1.2) - REFERENCE ONLY
├── json_formatter_test.go     # JSON tests - REFERENCE ONLY
├── yaml_formatter.go          # YAMLFormatter (from Story 1.3) - REFERENCE ONLY
└── yaml_formatter_test.go     # YAML tests - REFERENCE ONLY

cmd/awtest/types/types.go      # ScanResult struct (from Story 1.1) - DO NOT MODIFY
cmd/awtest/main.go             # Result collection (from Story 1.1) - DO NOT MODIFY
cmd/awtest/services/           # All service files - DO NOT MODIFY
go.mod                         # DO NOT MODIFY (no new dependencies)
go.sum                         # DO NOT MODIFY (no new dependencies)
```

### Previous Story Intelligence

**Story 1.3 (YAML Output Formatter) — Key Learnings:**
- YAML formatter followed JSON pattern successfully — CSV should do the same
- `tryMarshalYAML()` helper added for panic recovery (yaml.v3 panics on unserializable types like func/chan)
- For CSV: use `encoding/json.Marshal` to validate Details (same as JSON formatter) since CSV ultimately converts Details to a string anyway
- 13 tests, coverage 90.2%
- All existing tests remained passing (33 total formatter tests)

**Story 1.2 (JSON Output Formatter) — Key Learnings:**
- Code review found: switched to compact output (no MarshalIndent)
- Added per-result Details validation for resilient serialization
- Replaced hardcoded time format with `time.RFC3339` constant
- Nil Details map → serialize as `{}` (for JSON/YAML) or `""` (for CSV)

**Story 1.1 (Formatter Interface & Result Collection) — Foundation:**
- OutputFormatter interface: `Format(results []types.ScanResult) (string, error)` + `FileExtension() string`
- 34 services return `[]types.ScanResult`
- Patterns: table-driven tests, interface compliance checks, constructor functions

**Patterns to Reuse:**
- Interface compliance check: `var _ OutputFormatter = (*CSVFormatter)(nil)`
- Table-driven tests with `t.Run()` sub-tests
- Constructor function: `NewCSVFormatter()`
- Resilient serialization: validate Details before flattening
- Error as string in output column

### Git Intelligence

**Recent Commits:**
```
5598c70 Add YAML output formatter with comprehensive tests (Story 1.3)
ed241ba Add Rekognition service enumeration
f7dcee9 Validate JSON formatter and address code review findings (Story 1.2)
6017a06 Add .gitignore, tests, and update service calls with result collection
```

**Patterns from Recent Work:**
- Small, incremental commits with descriptive messages
- Tests co-located with implementations
- Build verification after each change
- Code review findings addressed in subsequent commits

### Edge Cases to Handle

1. **Empty results** → Header row only, no data rows
2. **Nil Details map** → Empty string in Details column
3. **Empty Details map** → Empty string in Details column
4. **Error with no other fields** → Error string in Error column, other fields empty
5. **Special characters in ResourceName** → csv.Writer auto-quotes: `"arn:aws:s3:::bucket"` (contains colons), `"value,with,commas"` (contains commas)
6. **Double quotes in values** → csv.Writer escapes: `"value ""with"" quotes"`
7. **Newlines in values** → csv.Writer quotes: `"value\nwith\nnewlines"`
8. **Details with complex nested values** → `fmt.Sprintf("%v", value)` handles nested maps/slices
9. **Unserializable Details values** → Resilient: use `encoding/json.Marshal` to validate, empty Details + error marker if failed
10. **Deterministic Details ordering** → Sort map keys alphabetically before flattening
11. **CSV line endings** → Go's csv.Writer uses `\r\n` (CRLF) per RFC 4180 — tests should account for this
12. **Large result sets** → csv.Writer with bytes.Buffer is efficient, no concerns

### CSV-Specific Testing Considerations

- **CRLF line endings:** Go's `csv.Writer` produces `\r\n` line endings per RFC 4180. Tests parsing output should use `csv.NewReader()` which handles this automatically.
- **Round-trip test:** Write CSV with formatter, then parse with `csv.NewReader()` — verify field count and values match
- **Details column parsing:** After CSV parse, verify Details column contains expected `key=value;key=value` format
- **Header verification:** First row should always be `["Service", "Method", "ResourceType", "ResourceName", "Details", "Timestamp", "Error"]`

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/`)
- Follows package-per-directory convention
- Tests co-located with implementation (`*_test.go` in same directory)
- No new dependencies — uses only Go stdlib
- No conflicts with existing file structure

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.4: CSV Output Formatter]
- [Source: _bmad-output/planning-artifacts/architecture.md#Output Format Architecture]
- [Source: _bmad-output/planning-artifacts/architecture.md#Implementation Patterns]
- [Source: _bmad-output/planning-artifacts/architecture.md#File Structure]
- [Source: _bmad-output/implementation-artifacts/1-3-yaml-output-formatter.md#Completion Notes]
- [Source: _bmad-output/implementation-artifacts/1-2-json-output-formatter.md#Completion Notes]
- [Source: cmd/awtest/formatters/json_formatter.go] — Canonical Format() pattern
- [Source: cmd/awtest/formatters/yaml_formatter.go] — Resilient serialization reference
- [Source: cmd/awtest/formatters/json_formatter_test.go] — Test pattern reference
- [Source: cmd/awtest/formatters/yaml_formatter_test.go] — Test pattern reference
- [Source: cmd/awtest/formatters/output_formatter.go] — Interface definition
- [Source: cmd/awtest/types/types.go] — ScanResult struct definition

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Implemented CSVFormatter following established JSON/YAML formatter patterns exactly
- Stateless struct with NewCSVFormatter() constructor, Format(), and FileExtension() methods
- Format() uses encoding/csv Writer with bytes.Buffer for RFC 4180 compliance
- flattenDetails() converts map[string]interface{} to sorted key=value;key=value string
- Resilient serialization using json.Marshal validation (same approach as JSON formatter)
- Error column combines scan errors and serialization errors (same pattern as JSON/YAML)
- 15 tests covering all ACs: valid results, empty results, special characters, round-trip, resilient serialization, deterministic ordering, interface compliance
- Coverage: 88.5% (exceeds 80% target)
- All 35 formatter tests pass (15 CSV + 13 JSON + 7 YAML), zero regressions
- No new dependencies added (all stdlib: encoding/csv, bytes, encoding/json, fmt, sort, strings, time)

### Change Log

- 2026-03-03: Implemented CSV output formatter (Story 1.4) — CSVFormatter with comprehensive tests

### File List

- cmd/awtest/formatters/csv_formatter.go (NEW)
- cmd/awtest/formatters/csv_formatter_test.go (NEW)

## Senior Developer Review (AI)

### Review Findings
- **AC Validation**: All 10 ACs are fully implemented and verified.
- **Task Audit**: All tasks marked [x] are completed.
- **Code Quality**: Excellent implementation following established patterns.
- **Test Quality**: comprehensive test suite with 100% pass rate.
- **Security**: No security issues found. Proper escaping used for CSV.

### Outcome
**Approve**

### Validation Checklist
- [x] Story file loaded from `_bmad-output/implementation-artifacts/1-4-csv-output-formatter.md`
- [x] Story Status verified as reviewable (review)
- [x] Epic and Story IDs resolved (1.4)
- [x] Acceptance Criteria cross-checked against implementation
- [x] File List reviewed and validated for completeness
- [x] Tests identified and mapped to ACs
- [x] Code quality review performed on changed files
- [x] Security review performed on changed files and dependencies
- [x] Outcome decided (Approve)
- [x] Status updated to done
- [x] Sprint status synced

_Reviewer: Kn0ck0ut on 2026-03-03_
