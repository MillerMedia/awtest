# Story 1.6: Format Selection & File Output

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security professional running awtest,
I want to select output format via command-line flag and optionally save to file,
so that I can choose the format that best fits my workflow and save results for later analysis.

## Acceptance Criteria

**Given** all formatters implemented (Stories 1.1-1.5)
**When** integrating formatters into main.go
**Then:**

1. **AC1:** Add `-format` flag accepting values: `text`, `json`, `yaml`, `csv`, `table` (default: `text`)
2. **AC2:** Add `-output-file` flag accepting file path for writing formatted output
3. **AC3:** Implement `getFormatter(format string) (formatters.OutputFormatter, error)` factory function in `main.go`
4. **AC4:** Default format is `text` when `-format` not specified, preserving current colorized terminal output
5. **AC5:** After service enumeration loop, pass collected `results` to selected formatter's `Format()` method
6. **AC6:** If `-output-file` specified, write formatted output to file
7. **AC7:** If `-output-file` NOT specified, write formatted output to stdout
8. **AC8:** Preserve existing colorized output for `text` format when outputting to terminal (current behavior)
9. **AC9:** Validate `-format` flag value; error with supported format list if invalid
10. **AC10:** Handle file write errors gracefully with clear error messages to stderr
11. **AC11:** Write unit tests for `getFormatter()` factory covering all formats + invalid input
12. **AC12:** Verify `go build ./cmd/awtest` succeeds
13. **AC13:** Verify `go test ./cmd/awtest/...` passes (all existing + new tests)

## Tasks / Subtasks

- [x] Task 1: Create TextFormatter (AC: 4, 8)
  - [x] Create `cmd/awtest/formatters/text_formatter.go`
  - [x] Implement `TextFormatter` struct following stateless pattern from JSON/YAML/CSV/Table formatters
  - [x] `Format()` method replicates current `utils.PrintResult()` colorized output by iterating results and building colorized string
  - [x] `FileExtension()` returns `"txt"`
  - [x] Add interface compliance: `var _ OutputFormatter = (*TextFormatter)(nil)`
  - [x] Create `cmd/awtest/formatters/text_formatter_test.go` with tests for: valid results, empty results, error results, interface compliance, constructor, file extension

- [x] Task 2: Add CLI flags to main.go (AC: 1, 2)
  - [x] Add `format` flag: `flag.String("format", "text", "Output format: text, json, yaml, csv, table")`
  - [x] Add `output-file` flag: `flag.String("output-file", "", "Write output to file instead of stdout")`
  - [x] Place flag definitions alongside existing flags (lines 28-37)

- [x] Task 3: Implement getFormatter factory (AC: 3, 9)
  - [x] Create `getFormatter(format string) (formatters.OutputFormatter, error)` in `main.go`
  - [x] Use `strings.ToLower(format)` for case-insensitive matching
  - [x] Switch cases: `text` -> `NewTextFormatter()`, `json` -> `NewJSONFormatter()`, `yaml` -> `NewYAMLFormatter()`, `csv` -> `NewCSVFormatter()`, `table` -> `NewTableFormatter()`
  - [x] Default case returns error: `fmt.Errorf("unsupported format: %s (supported: text, json, yaml, csv, table)", format)`
  - [x] Add import for `formatters` package: `"github.com/MillerMedia/awtest/cmd/awtest/formatters"`

- [x] Task 4: Integrate formatter output in main.go (AC: 5, 6, 7, 8, 10)
  - [x] After `flag.Parse()`, call `getFormatter(*format)` and handle error (print to stderr, exit 1)
  - [x] Replace lines 137-139 comment block with formatter integration
  - [x] For `text` format with NO `-output-file`: keep current behavior (Process() already prints colorized output to stdout) -- skip formatter call
  - [x] For `text` format WITH `-output-file`: call `formatter.Format(results)` and write to file
  - [x] For non-text formats: call `formatter.Format(results)` and output to stdout or file
  - [x] File output: use `os.WriteFile(outputFile, []byte(formatted), 0644)` with error handling
  - [x] File write error: `fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)` then `os.Exit(1)`
  - [x] Formatting error: `fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)` then `os.Exit(1)`

- [x] Task 5: Write tests (AC: 11, 12, 13)
  - [x] Create `cmd/awtest/main_test.go` (or add to existing test file)
  - [x] Test `getFormatter("json")` returns non-nil formatter
  - [x] Test `getFormatter("yaml")` returns non-nil formatter
  - [x] Test `getFormatter("csv")` returns non-nil formatter
  - [x] Test `getFormatter("table")` returns non-nil formatter
  - [x] Test `getFormatter("text")` returns non-nil formatter
  - [x] Test `getFormatter("JSON")` returns non-nil (case-insensitive)
  - [x] Test `getFormatter("invalid")` returns error
  - [x] Test `getFormatter("")` returns error

- [x] Task 6: Verify build and tests (AC: 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/formatters/...` (all 56 formatter tests pass)
  - [x] `go test ./cmd/awtest/...` (full regression)
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow Established Patterns

This story is the **integration point** for the entire output format system. It connects the formatters (Stories 1.2-1.5) to the CLI.

**Key Design Decision: Text Format Handling**

For `-format=text` with stdout output (the default case), the existing `Process()` methods in each service already print colorized output via `utils.PrintResult()`. Do NOT duplicate this output. The `TextFormatter` is only needed for `-output-file` scenarios where colorized text needs to be written to a file, or for consistency in the factory pattern.

**Approach for text format to stdout:**
```go
// In main.go, after service enumeration:
if *outputFormat == "text" && *outputFile == "" {
    // Results already printed by Process() methods -- nothing to do
    return
}

// For all other cases, use formatter
formatter, err := getFormatter(*outputFormat)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}

formatted, err := formatter.Format(results)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
    os.Exit(1)
}

if *outputFile != "" {
    if err := os.WriteFile(*outputFile, []byte(formatted), 0644); err != nil {
        fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
        os.Exit(1)
    }
    fmt.Fprintf(os.Stderr, "Output written to %s\n", *outputFile)
} else {
    fmt.Print(formatted)
}
```

**Important:** When `-format=json` (or yaml/csv/table) is specified WITHOUT `-output-file`, the Process() methods still print colorized output AND the formatter prints structured output. This is expected for Story 1.6. Story 1.7 will add `-quiet` flag to suppress Process() output.

### Formatter Pattern Reference

All formatters follow this pattern (from Stories 1.1-1.5):

```go
// Stateless struct
type XxxFormatter struct{}

// Constructor
func NewXxxFormatter() *XxxFormatter { return &XxxFormatter{} }

// Interface compliance
var _ OutputFormatter = (*XxxFormatter)(nil)

// Format method
func (f *XxxFormatter) Format(results []types.ScanResult) (string, error) { ... }

// Extension method
func (f *XxxFormatter) FileExtension() string { return "xxx" }
```

**Existing formatters:**
- `json_formatter.go` - Compact JSON, camelCase fields, RFC3339 timestamps
- `yaml_formatter.go` - Human-readable YAML with panic recovery
- `csv_formatter.go` - CSV with header row, flattened details
- `table_formatter.go` - ASCII table, 120-char width, column truncation

### TextFormatter Implementation Notes

The `TextFormatter` should replicate what `utils.PrintResult()` produces -- colorized terminal output per result line:
```
[ServiceName] [MethodName] [severity] ResourceType: ResourceName
```

Use the existing `utils.colorizeMessage()` logic. Since `colorizeMessage` is unexported, the TextFormatter can either:
1. Call `utils.PrintResult()` but capture to buffer (not ideal -- PrintResult prints to stdout directly)
2. Duplicate the colorization logic in the formatter (acceptable for isolation)
3. Export `colorizeMessage` from utils (recommended -- rename to `ColorizeMessage`)

**Recommended approach:** Export `ColorizeMessage` from `utils/output.go` (capitalize first letter) and use it in TextFormatter. This avoids duplication while keeping formatters self-contained.

### Architecture Compliance

- **Package:** `formatters` (lowercase) -- MUST FOLLOW
- **File naming:** `text_formatter.go` (underscores for multi-word) -- MUST FOLLOW
- **Type naming:** `TextFormatter` (PascalCase exported) -- MUST FOLLOW
- **Constructor:** `NewTextFormatter()` -- MUST FOLLOW
- **Error handling:** `fmt.Errorf("text formatting failed: %w", err)` -- MUST FOLLOW
- **Testing:** stdlib `testing` package (no testify) -- MUST FOLLOW
- **Flag package:** stdlib `flag` (already used in main.go) -- MUST FOLLOW

### Technical Requirements

**Go Version:** 1.19 (existing project standard)

**No New Dependencies Required** -- all functionality uses stdlib:
- `os` -- for `WriteFile()`
- `strings` -- for `ToLower()` in format matching
- `fmt` -- for error messages
- `flag` -- already in use for CLI flags

**Existing Dependencies (unchanged):**
- `github.com/aws/aws-sdk-go v1.44.266`
- `github.com/logrusorgru/aurora v2.0.3+incompatible`
- `github.com/olekukonko/tablewriter v0.0.5`
- `gopkg.in/yaml.v3 v3.0.1`

### File Structure

**Files to CREATE:**
```
cmd/awtest/formatters/
+-- text_formatter.go          # NEW: TextFormatter implementation
+-- text_formatter_test.go     # NEW: TextFormatter tests
```

**Files to MODIFY:**
```
cmd/awtest/main.go             # Add -format, -output-file flags + getFormatter() + output integration
cmd/awtest/utils/output.go     # Export colorizeMessage -> ColorizeMessage (if needed by TextFormatter)
```

**Files that ALREADY EXIST (DO NOT MODIFY unless noted):**
```
cmd/awtest/formatters/
+-- output_formatter.go        # OutputFormatter interface (Story 1.1) - DO NOT MODIFY
+-- json_formatter.go          # JSONFormatter (Story 1.2) - DO NOT MODIFY
+-- yaml_formatter.go          # YAMLFormatter (Story 1.3) - DO NOT MODIFY
+-- csv_formatter.go           # CSVFormatter (Story 1.4) - DO NOT MODIFY
+-- table_formatter.go         # TableFormatter (Story 1.5) - DO NOT MODIFY
+-- *_test.go                  # All test files - DO NOT MODIFY

cmd/awtest/types/types.go      # ScanResult struct (Story 1.1) - DO NOT MODIFY
cmd/awtest/services/           # All service files - DO NOT MODIFY
```

### Testing Requirements

**New tests to create:**

1. `cmd/awtest/formatters/text_formatter_test.go`:
   - Test valid results produce colorized output
   - Test empty results return "No results found"
   - Test error results include error info
   - Test interface compliance: `var _ OutputFormatter = (*TextFormatter)(nil)`
   - Test constructor not nil
   - Test `FileExtension()` returns `"txt"`

2. `cmd/awtest/main_test.go` (getFormatter tests):
   - Test each format string returns correct formatter type
   - Test case-insensitive matching
   - Test invalid format returns error with descriptive message
   - Test empty string returns error

**Regression verification:**
- All 46 existing formatter tests must pass
- `go test ./cmd/awtest/...` -- full test suite
- `go vet ./cmd/awtest/...` -- no lint issues

### Previous Story Intelligence

**Story 1.5 (Table Output Formatter) -- Key Learnings:**
- Stateless formatter pattern confirmed through 5 implementations
- Per-column truncation used for width enforcement
- 11 tests, 90.0% coverage -- all 46 formatter tests pass together
- Code review addressed AC8 violation (width guarantee)
- Pattern: each story adds 2 files (implementation + test) in formatters/

**Story 1.4 (CSV Output Formatter) -- Key Learnings:**
- `flattenDetails()` helper was CSV-specific
- 15 tests, 88.5% coverage
- CRLF handling was CSV-specific

**Story 1.1 (Foundation) -- Key Learnings:**
- OutputFormatter interface: `Format(results []types.ScanResult) (string, error)` + `FileExtension() string`
- 34 services return `[]types.ScanResult`
- main.go already collects results in `var results []types.ScanResult` (line 130)
- Comment on lines 137-139 explicitly notes: "In future stories, this will be replaced with formatter-based output"

### Git Intelligence

**Recent commits (newest first):**
```
4a11bbe Mark Story 1.3 YAML output formatter as done
0605d5d Mark Stories 1.1 and 1.2 as done
6f4cf10 Mark Story 1.5 table output formatter as done
9043861 Add table output formatter with 120-char width enforcement (Story 1.5)
fc7ed49 Add CSV output formatter with comprehensive tests (Story 1.4)
5598c70 Add YAML output formatter with comprehensive tests (Story 1.3)
```

**Patterns from recent work:**
- Small, focused commits with descriptive messages referencing story numbers
- Tests co-located with implementations
- Build verification after each change
- Each formatter story: 2 new files + go.mod if new dependency

### Edge Cases to Handle

1. **Default behavior unchanged** -- Running `awtest` with no flags works exactly as before (text to stdout via Process())
2. **Text format to file** -- `-format=text -output-file=results.txt` writes colorized text to file
3. **Non-text format to stdout** -- `-format=json` prints JSON to stdout (Process() output also prints -- acceptable until Story 1.7 quiet mode)
4. **Invalid format** -- `-format=xml` prints error to stderr, exits 1
5. **File write permission denied** -- Clear error message to stderr, exit 1
6. **Empty results** -- Each formatter handles empty results gracefully (already implemented)
7. **Case-insensitive format** -- `-format=JSON` works same as `-format=json`
8. **Output file already exists** -- Overwrite (standard behavior for CLI tools)

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/`)
- Follows package-per-directory convention
- Tests co-located with implementation (`*_test.go` in same directory)
- No new dependencies required
- No conflicts with existing file structure

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.6: Format Selection & File Output]
- [Source: _bmad-output/planning-artifacts/architecture.md#Output Format Architecture]
- [Source: _bmad-output/planning-artifacts/architecture.md#Configuration Management]
- [Source: _bmad-output/implementation-artifacts/1-5-table-output-formatter.md#Completion Notes]
- [Source: cmd/awtest/main.go] -- Current main with result collection (lines 130-139)
- [Source: cmd/awtest/formatters/output_formatter.go] -- Interface definition
- [Source: cmd/awtest/utils/output.go] -- PrintResult/colorizeMessage for TextFormatter reference
- [Source: cmd/awtest/types/types.go] -- ScanResult struct

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No blocking issues encountered during implementation.

### Completion Notes List

- Exported `colorizeMessage` → `ColorizeMessage` in `utils/output.go` (recommended approach from Dev Notes) to enable TextFormatter to reuse colorization logic without duplication
- Created `TextFormatter` following the established stateless formatter pattern (6th formatter in the system)
- TextFormatter produces colorized output matching `utils.PrintResult()` format: `[ServiceName] [MethodName] [severity] ResourceType: ResourceName`
- Added `-format` and `-output-file` CLI flags to `main.go` alongside existing flags
- Implemented `getFormatter()` factory with case-insensitive matching via `strings.ToLower()` and descriptive error for invalid formats
- Integrated formatter output in main.go: text+stdout preserves existing behavior (Process() output), all other combinations use formatter pipeline
- Format validation happens early (before service enumeration) to fail fast on invalid format
- File output uses `os.WriteFile` with `0644` permissions and stderr confirmation message
- All error paths write to stderr and exit with code 1
- 10 new TextFormatter tests (6 Format subtests + FileExtension + constructor + interface compliance)
- 4 new getFormatter tests covering all formats, case-insensitivity, invalid input, and empty string
- All 56 formatter tests pass, full `go test ./cmd/awtest/...` passes, `go vet` clean

### File List

**New files:**
- `cmd/awtest/formatters/text_formatter.go` — TextFormatter implementation
- `cmd/awtest/formatters/text_formatter_test.go` — TextFormatter tests
- `cmd/awtest/main_test.go` — getFormatter factory tests

**Modified files:**
- `cmd/awtest/main.go` — Added `-format`/`-output-file` flags, `getFormatter()` factory, formatter output integration, `formatters` import
- `cmd/awtest/utils/output.go` — Exported `colorizeMessage` → `ColorizeMessage` (capitalized) and updated internal caller

## Change Log

- 2026-03-03: Implemented Story 1.6 — Format Selection & File Output. Added TextFormatter, CLI flags (`-format`, `-output-file`), getFormatter factory, and formatter output integration in main.go. Exported ColorizeMessage from utils. 14 new tests added.
