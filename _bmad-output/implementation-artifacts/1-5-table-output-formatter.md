# Story 1.5: Table Output Formatter

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security professional viewing scan results in terminal,
I want ASCII table-formatted scan results,
so that I can quickly scan findings in a structured, readable table layout.

## Acceptance Criteria

**Given** the OutputFormatter interface from Story 1.1
**When** implementing ASCII table output format
**Then:**

1. **AC1:** Create TableFormatter struct in `cmd/awtest/formatters/table_formatter.go` implementing OutputFormatter interface
2. **AC2:** Add `github.com/olekukonko/tablewriter` v0.0.5 dependency to go.mod
3. **AC3:** `Format()` method produces ASCII table with columns: Service, Method, Resource Type, Resource Name, Timestamp
4. **AC4:** Table uses borders, proper alignment, and column wrapping for readability
5. **AC5:** `FileExtension()` returns `"txt"`
6. **AC6:** Handle empty results gracefully (return "No results found" message)
7. **AC7:** Handle errors in `ScanResult.Error` field by including error indicator in table row
8. **AC8:** Limit table width to 120 characters for terminal readability
9. **AC9:** Write table-driven unit tests in `table_formatter_test.go` covering: valid results, empty results, long resource names requiring wrapping, results with errors
10. **AC10:** Verify table renders correctly in 80-column and 120-column terminals
11. **AC11:** Test passes: `go test ./cmd/awtest/formatters/...`

## Tasks / Subtasks

- [x] Add tablewriter dependency (AC: 2)
  - [x] Run `go get github.com/olekukonko/tablewriter@v0.0.5`
  - [x] Verify go.mod and go.sum updated correctly
  - [x] Verify `go build ./cmd/awtest` still succeeds
- [x] Create TableFormatter implementation (AC: 1, 3, 4, 5, 6, 7, 8)
  - [x] Create `cmd/awtest/formatters/table_formatter.go`
  - [x] Define `TableFormatter` struct (stateless, like JSON/YAML/CSV formatters)
  - [x] Implement `NewTableFormatter()` constructor
  - [x] Implement `Format()` method using `tablewriter.NewWriter()` with `bytes.Buffer`
  - [x] Set table configuration: borders enabled, auto-wrap enabled, column max width for 120-char limit
  - [x] Set header row: Service, Method, Resource Type, Resource Name, Timestamp
  - [x] Write data rows from ScanResult fields
  - [x] Handle empty results: return "No results found" string instead of empty table
  - [x] Handle ScanResult.Error: append "[ERROR]" indicator to Resource Name column or add error info
  - [x] Implement `FileExtension()` returning `"txt"`
  - [x] Add interface compliance check: `var _ OutputFormatter = (*TableFormatter)(nil)`
- [x] Write comprehensive tests (AC: 9, 10)
  - [x] Create `cmd/awtest/formatters/table_formatter_test.go`
  - [x] Test: single valid result produces table with correct columns
  - [x] Test: empty results returns "No results found" message
  - [x] Test: multiple results produce correct row count
  - [x] Test: long resource names are handled (wrapping or truncation)
  - [x] Test: result with error includes error indicator
  - [x] Test: result with nil error has no error indicator
  - [x] Test: table output contains borders
  - [x] Test: table output line width does not exceed 120 characters
  - [x] Test: interface compliance check
  - [x] Test: constructor not nil
  - [x] Test: FileExtension returns "txt"
  - [x] Test: timestamp formatted correctly in table
- [x] Verify build and tests pass (AC: 11)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/formatters/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -cover ./cmd/awtest/formatters/...` (target: 80%+)
- [x] Run full regression (AC: 11)
  - [x] `go test ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow the Established Formatter Pattern

The Table formatter must mirror the JSON/YAML/CSV formatter implementation patterns established in Stories 1.1-1.4. These were refined through code review and represent the canonical approach.

**Pattern Sources:**
- `cmd/awtest/formatters/json_formatter.go` -- canonical Format() pattern
- `cmd/awtest/formatters/csv_formatter.go` -- most recent formatter, well-tested pattern
- `cmd/awtest/formatters/yaml_formatter.go` -- resilient serialization with panic recovery

**Key Pattern Elements:**
1. **Stateless struct** -- `TableFormatter` has no fields (like JSONFormatter, YAMLFormatter, CSVFormatter)
2. **Constructor function** -- `NewTableFormatter()` returns `*TableFormatter`
3. **Error handling** -- `fmt.Errorf("table formatting failed: %w", err)`
4. **Interface compliance** -- `var _ OutputFormatter = (*TableFormatter)(nil)`
5. **Timestamp as RFC3339 string** -- Use `time.RFC3339` constant

### Table-Specific Implementation Notes

**Use `github.com/olekukonko/tablewriter` v0.0.5** -- This is the stable, legacy version with the well-documented API. Do NOT use v1.x (completely different API).

**CRITICAL: v0.0.5 API (NOT v1.x):**
```go
import "github.com/olekukonko/tablewriter"

// v0.0.5 uses this API:
table := tablewriter.NewWriter(&buf)
table.SetHeader([]string{"Service", "Method", "Resource Type", "Resource Name", "Timestamp"})
table.SetBorder(true)
table.SetAutoWrapText(true)
table.SetColWidth(30)        // Max column width for wrapping
table.SetRowLine(false)      // No line between rows
table.Append([]string{...})  // Add rows one at a time
table.Render()               // Write to buffer
```

**Implementation Pattern:**
```go
package formatters

import (
    "bytes"
    "fmt"
    "time"

    "github.com/MillerMedia/awtest/cmd/awtest/types"
    "github.com/olekukonko/tablewriter"
)

type TableFormatter struct{}

func NewTableFormatter() *TableFormatter {
    return &TableFormatter{}
}

func (f *TableFormatter) Format(results []types.ScanResult) (string, error) {
    if len(results) == 0 {
        return "No results found", nil
    }

    var buf bytes.Buffer
    table := tablewriter.NewWriter(&buf)
    table.SetHeader([]string{"Service", "Method", "Resource Type", "Resource Name", "Timestamp"})
    table.SetBorder(true)
    table.SetAutoWrapText(true)
    table.SetColWidth(30)

    for _, r := range results {
        resourceName := r.ResourceName
        if r.Error != nil {
            resourceName = resourceName + " [ERROR: " + r.Error.Error() + "]"
        }
        table.Append([]string{
            r.ServiceName,
            r.MethodName,
            r.ResourceType,
            resourceName,
            r.Timestamp.Format(time.RFC3339),
        })
    }

    table.Render()
    return buf.String(), nil
}

func (f *TableFormatter) FileExtension() string {
    return "txt"
}
```

**Column Design Rationale:**
- **5 columns** (not 7 like CSV): Table format is for terminal readability, so Details and Error are excluded as separate columns
- **Details omitted**: Map data doesn't render well in fixed-width table columns; details are available in JSON/YAML/CSV formats
- **Error indicator**: Appended to Resource Name as `[ERROR: message]` to keep the table compact
- **Timestamp included**: Important for scan context

**Width Management:**
- Total table width target: 120 characters max
- Column widths should distribute roughly: Service(15), Method(25), ResourceType(15), ResourceName(30), Timestamp(25) + borders/padding
- `SetColWidth(30)` sets the max width before wrapping
- `SetAutoWrapText(true)` enables automatic wrapping of long text

**Empty Results Handling:**
- Unlike CSV (returns header-only), table returns a plain string: `"No results found"`
- This is more user-friendly for terminal display

### Architecture Compliance

**From Architecture Document:**

- **Formatter Interface Pattern:** OutputFormatter with `Format()` and `FileExtension()` -- MUST FOLLOW
- **Package:** `formatters` (lowercase, no underscores) -- MUST FOLLOW
- **File naming:** `table_formatter.go` (underscores for multi-word) -- MUST FOLLOW
- **Type naming:** `TableFormatter` (PascalCase exported) -- MUST FOLLOW
- **Constructor:** `NewTableFormatter()` (Go convention) -- MUST FOLLOW
- **Error handling:** Returns `fmt.Errorf("table formatting failed: %w", err)` -- MUST FOLLOW
- **New dependency:** `github.com/olekukonko/tablewriter` v0.0.5 (specified in architecture)
- **Testing:** stdlib `testing` package (no testify)

### Technical Requirements

**Go Version:** 1.19 (existing project standard)

**Dependencies (ONE NEW DEPENDENCY):**
- `github.com/olekukonko/tablewriter` v0.0.5 -- **NEW** (specified in architecture document)
- `bytes` (stdlib -- for buffer)
- `fmt`, `time` (stdlib)

**Existing Dependencies (unchanged):**
- `github.com/aws/aws-sdk-go v1.44.266`
- `github.com/logrusorgru/aurora v2.0.3+incompatible`
- `gopkg.in/yaml.v3 v3.0.1`

**Build Verification:**
```bash
go get github.com/olekukonko/tablewriter@v0.0.5  # Add dependency
go build ./cmd/awtest             # Must succeed
go test ./cmd/awtest/formatters/... # Must pass
go vet ./cmd/awtest/...           # Must pass
```

### File Structure

**Files to CREATE:**
```
cmd/awtest/formatters/
+-- table_formatter.go          # NEW: TableFormatter implementation
+-- table_formatter_test.go     # NEW: Table formatter tests
```

**Files to MODIFY:**
```
go.mod                           # Add tablewriter dependency
go.sum                           # Auto-updated by go get
```

**Files that ALREADY EXIST (DO NOT MODIFY):**
```
cmd/awtest/formatters/
+-- output_formatter.go        # OutputFormatter interface (from Story 1.1)
+-- output_formatter_test.go   # Interface tests (from Story 1.1)
+-- json_formatter.go          # JSONFormatter (from Stories 1.1/1.2) - REFERENCE ONLY
+-- json_formatter_test.go     # JSON tests - REFERENCE ONLY
+-- yaml_formatter.go          # YAMLFormatter (from Story 1.3) - REFERENCE ONLY
+-- yaml_formatter_test.go     # YAML tests - REFERENCE ONLY
+-- csv_formatter.go           # CSVFormatter (from Story 1.4) - REFERENCE ONLY
+-- csv_formatter_test.go      # CSV tests - REFERENCE ONLY

cmd/awtest/types/types.go      # ScanResult struct (from Story 1.1) - DO NOT MODIFY
cmd/awtest/main.go             # Result collection (from Story 1.1) - DO NOT MODIFY
cmd/awtest/services/           # All service files - DO NOT MODIFY
```

### Previous Story Intelligence

**Story 1.4 (CSV Output Formatter) -- Key Learnings:**
- CSV formatter followed JSON/YAML pattern successfully -- Table should do the same
- Stateless struct pattern confirmed as canonical
- `flattenDetails()` helper was CSV-specific; Table formatter does NOT need Details flattening since Details column is omitted
- 15 tests, coverage 88.5%
- All 35 existing formatter tests remained passing (15 CSV + 13 JSON + 7 YAML)
- CRLF handling was CSV-specific concern; not relevant for table output

**Story 1.3 (YAML Output Formatter) -- Key Learnings:**
- `tryMarshalYAML()` helper added for panic recovery (yaml.v3 panics on unserializable types)
- For Table: no serialization concern since we're just rendering string fields
- 13 tests, coverage 90.2%

**Story 1.2 (JSON Output Formatter) -- Key Learnings:**
- Code review found: switched to compact output (no MarshalIndent)
- Added per-result Details validation for resilient serialization
- Nil Details map handling established

**Story 1.1 (Formatter Interface & Result Collection) -- Foundation:**
- OutputFormatter interface: `Format(results []types.ScanResult) (string, error)` + `FileExtension() string`
- 34 services return `[]types.ScanResult`
- Patterns: table-driven tests, interface compliance checks, constructor functions

**Patterns to Reuse:**
- Interface compliance check: `var _ OutputFormatter = (*TableFormatter)(nil)`
- Table-driven tests with `t.Run()` sub-tests
- Constructor function: `NewTableFormatter()`
- Error field handling pattern (check nil, convert to string)

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
- Each formatter story adds 2 files: implementation + tests
- Story 1.3 (YAML) added dependency to go.mod/go.sum -- same pattern needed for tablewriter

### Edge Cases to Handle

1. **Empty results** -- Return "No results found" string (not empty table)
2. **Single result** -- Table with header + 1 data row
3. **Error with no other fields** -- Error indicator appended to ResourceName: `" [ERROR: message]"`
4. **Very long resource names** -- tablewriter auto-wraps at ColWidth(30)
5. **Very long method names** -- Same wrapping behavior
6. **Special characters in values** -- tablewriter handles escaping internally
7. **Nil error** -- No error indicator appended
8. **Multiple errors** -- Each row independently shows its error
9. **Line width check** -- Verify no output line exceeds 120 characters
10. **Consistent timestamp format** -- RFC3339 matching all other formatters

### Table Output Width Testing Notes

- **120-char limit:** Set `SetColWidth()` appropriately and verify with test that splits output by newline and checks `len(line) <= 120`
- **80-char terminal:** Table will wrap naturally since columns have max widths; verify table is still readable
- **tablewriter handles alignment** -- Left-align text columns by default

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/`)
- Follows package-per-directory convention
- Tests co-located with implementation (`*_test.go` in same directory)
- ONE new dependency: `github.com/olekukonko/tablewriter` v0.0.5 (specified in architecture)
- No conflicts with existing file structure

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.5: Table Output Formatter]
- [Source: _bmad-output/planning-artifacts/architecture.md#Output Format Architecture]
- [Source: _bmad-output/planning-artifacts/architecture.md#Implementation Patterns]
- [Source: _bmad-output/planning-artifacts/architecture.md#File Structure]
- [Source: _bmad-output/implementation-artifacts/1-4-csv-output-formatter.md#Completion Notes]
- [Source: cmd/awtest/formatters/json_formatter.go] -- Canonical Format() pattern
- [Source: cmd/awtest/formatters/csv_formatter.go] -- Most recent formatter reference
- [Source: cmd/awtest/formatters/output_formatter.go] -- Interface definition
- [Source: cmd/awtest/types/types.go] -- ScanResult struct definition
- [Source: https://pkg.go.dev/github.com/olekukonko/tablewriter] -- tablewriter v0.0.5 API reference

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered. Clean implementation.

### Completion Notes List

- Added `github.com/olekukonko/tablewriter` v0.0.5 dependency (plus transitive `github.com/mattn/go-runewidth` v0.0.9)
- Implemented `TableFormatter` struct following established stateless formatter pattern from Stories 1.1-1.4
- `Format()` produces ASCII table with 5 columns: Service, Method, Resource Type, Resource Name, Timestamp
- Table enforces 120-char width limit via per-column truncation with ellipsis (Service:16, Method:24, ResourceType:16, ResourceName:26, Timestamp:20)
- Empty results return "No results found" string (user-friendly terminal display)
- Error indicator appended to Resource Name as `[ERROR: message]` for compact table layout
- Compile-time interface compliance check included
- 11 tests written covering all acceptance criteria: valid results, empty results, multiple results, long names (truncation), error indicators, nil errors, borders, line width limit (worst-case data), constructor, file extension, timestamp format
- All 46 formatter tests pass (11 Table + 15 CSV + 13 JSON + 7 YAML) — zero regressions
- Test coverage: 90.0% (exceeds 80% target)
- `go build`, `go test`, `go vet` all pass cleanly
- Addressed code review findings: fixed AC8 violation (column truncation for 120-char guarantee), removed redundant interface test

### Change Log

- 2026-03-03: Implemented Table Output Formatter (Story 1.5) — added tablewriter dependency, TableFormatter implementation, and 12 comprehensive tests
- 2026-03-03: Addressed code review findings — 3 items resolved: fixed AC8 120-char width violation with per-column truncation, removed redundant TestTableFormatter_ImplementsInterface, noted untracked CSV files from Story 1.4

### File List

**New files:**
- cmd/awtest/formatters/table_formatter.go
- cmd/awtest/formatters/table_formatter_test.go

**Modified files:**
- go.mod (added tablewriter v0.0.5 and go-runewidth v0.0.9 dependencies)
- go.sum (auto-updated)
