# Story 1.7: Progress Tracking & Summary Reporting

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security professional running comprehensive scans,
I want real-time progress indicators and a findings summary,
so that I can see scan progress and quickly understand the overall results without reading every line.

## Acceptance Criteria

**Given** the formatter system integrated from Story 1.6
**When** implementing progress and summary features
**Then:**

1. **AC1:** Add `-quiet` flag to suppress informational messages, showing only findings (FR46)
2. **AC2:** Display real-time progress during scan showing: "Scanning [service_name]..." for each service (FR47)
3. **AC3:** Progress messages write to stderr (not stdout) so they don't interfere with formatted output
4. **AC4:** Suppress progress messages when `-quiet` flag is set
5. **AC5:** After scan completes, display summary report showing: total services scanned, services with accessible resources, services with access denied, total resources discovered, scan duration, timestamp (FR48)
6. **AC6:** Summary report respects `-quiet` flag (suppress when quiet mode enabled)
7. **AC7:** Summary formatting adapts to selected output format (plain text for text/table, structured data for JSON/YAML/CSV)
8. **AC8:** Write unit tests for progress tracking and summary generation
9. **AC9:** Verify: `go run ./cmd/awtest -quiet -format=json` shows only JSON output, no progress messages
10. **AC10:** Verify: `go run ./cmd/awtest` shows progress messages and summary report
11. **AC11:** Scan metadata includes timestamp and duration (FR40)
12. **AC12:** Verify `go test ./cmd/awtest/...` passes (all existing + new tests)
13. **AC13:** Verify `go build ./cmd/awtest` succeeds

## Tasks / Subtasks

- [x] Task 1: Add `-quiet` flag and suppress Process() output (AC: 1, 4)
  - [x] Add `quiet` flag: `flag.Bool("quiet", false, "Suppress informational messages, show only findings")` in `main.go` alongside existing flags
  - [x] Add exported `Quiet` variable to `utils` package: `var Quiet bool` in `utils/output.go`
  - [x] Set `utils.Quiet = *quiet` in `main.go` after `flag.Parse()`
  - [x] In `utils.PrintResult()`: if `Quiet` is true, return immediately (no-op) — this suppresses all Process() output from all 34 services without modifying any service code
  - [x] In `utils.PrintAccessGranted()`: same quiet check (it calls PrintResult, so already covered, but verify)
  - [x] In `utils.HandleAWSError()`: same quiet check — suppress error output in quiet mode since errors are captured in ScanResult.Error

- [x] Task 2: Add real-time progress messages (AC: 2, 3, 4)
  - [x] In `main.go` scan loop, before each `service.Call(sess)`, print progress to stderr: `fmt.Fprintf(os.Stderr, "Scanning %s...\n", service.Name)`
  - [x] Guard progress output with `if !*quiet { ... }` to suppress in quiet mode
  - [x] Progress goes to stderr so it never contaminates stdout formatted output

- [x] Task 3: Create `ScanSummary` type and generation (AC: 5, 11)
  - [x] Create `cmd/awtest/types/summary.go` with `ScanSummary` struct:
    ```go
    type ScanSummary struct {
        TotalServices       int
        AccessibleServices  int
        AccessDeniedServices int
        TotalResources      int
        ScanDuration        time.Duration
        Timestamp           time.Time
    }
    ```
  - [x] Create `func GenerateSummary(results []ScanResult, startTime time.Time) ScanSummary` in `types/summary.go`
  - [x] Logic: iterate results, count unique services, categorize by error vs success, count non-error results as resources, compute duration from startTime

- [x] Task 4: Display text summary after scan (AC: 5, 6, 7, 10)
  - [x] In `main.go`, record `startTime := time.Now()` before the scan loop
  - [x] After scan loop, call `summary := types.GenerateSummary(results, startTime)`
  - [x] For text format to stdout (default path): if NOT quiet, print summary to stderr
  - [x] Summary text format:
    ```
    ========================================
    Scan Summary
    ========================================
    Timestamp:          2026-03-03T10:30:00Z
    Duration:           45.2s
    Total Services:     34
    Accessible:         12
    Access Denied:      22
    Resources Found:    87
    ========================================
    ```
  - [x] Print summary to stderr (not stdout) so it doesn't interfere with formatted output piped to files/tools
  - [x] Guard with `if !*quiet { ... }`

- [x] Task 5: Add summary to structured output formats (AC: 7)
  - [x] For JSON/YAML: modify formatter output to wrap results in envelope with metadata:
    ```json
    {
      "metadata": {
        "timestamp": "2026-03-03T10:30:00Z",
        "duration": "45.2s",
        "totalServices": 34,
        "accessibleServices": 12,
        "accessDeniedServices": 22,
        "totalResources": 87
      },
      "results": [...]
    }
    ```
  - [x] Add `FormatWithSummary(results []types.ScanResult, summary types.ScanSummary) (string, error)` method to `OutputFormatter` interface
  - [x] Implement `FormatWithSummary` in all 5 formatters (JSON, YAML, CSV, Table, Text)
  - [x] Keep existing `Format()` method working unchanged for backward compatibility — `FormatWithSummary` is the new primary method used by main.go
  - [x] For CSV: add summary as header comment rows before the data rows
  - [x] For Table: append summary section below the table
  - [x] For Text: append summary section below the text output
  - [x] In `main.go`: call `formatter.FormatWithSummary(results, summary)` instead of `formatter.Format(results)`

- [x] Task 6: Suppress banner in quiet mode (AC: 1, 9)
  - [x] Wrap the ASCII banner print block (lines 21-28 of main.go) with `if !*quiet { ... }`
  - [x] The banner is informational — quiet mode should suppress it

- [x] Task 7: Write unit tests (AC: 8, 12)
  - [x] Create `cmd/awtest/types/summary_test.go`:
    - Test GenerateSummary with mixed results (success + error)
    - Test GenerateSummary with empty results
    - Test GenerateSummary with all errors
    - Test GenerateSummary with all successes
    - Test duration calculation
    - Test unique service counting
  - [x] Update `cmd/awtest/main_test.go`:
    - Test quiet flag exists
  - [x] Update formatter tests to test FormatWithSummary:
    - JSON: verify metadata envelope present
    - YAML: verify metadata section present
    - CSV: verify header comments present
    - Table: verify summary section present
    - Text: verify summary section present
  - [x] Verify all existing 56+ formatter tests still pass

- [x] Task 8: Verify build and all tests (AC: 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/...` (full regression)
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow Established Patterns

This is the **capstone story** for Epic 1 (Output Format System). It completes the CLI output experience by adding progress visibility, summary reporting, and quiet mode.

### Key Design Decisions

**1. Quiet Mode via Package-Level Variable**

The `utils.PrintResult()` function is called by all 34 service Process() methods. Rather than changing the Process() signature (which would require modifying every service file), add a package-level `utils.Quiet` bool that `PrintResult()` checks. This is the least invasive approach — zero service file changes.

```go
// In utils/output.go
var Quiet bool

func PrintResult(debug bool, moduleName string, method string, result string, err error) {
    if Quiet {
        return
    }
    // ... existing implementation
}
```

Set it in main.go after flag.Parse():
```go
utils.Quiet = *quiet
```

**2. Progress Messages to stderr**

Progress MUST go to stderr, not stdout. This is critical because:
- `awtest -format=json | jq .` — progress would corrupt JSON on stdout
- `awtest -format=csv -output-file=out.csv` — progress to stderr is visible, output goes to file
- Quiet mode suppresses both progress AND Process() output

```go
for _, service := range services.AllServices() {
    if !*quiet {
        fmt.Fprintf(os.Stderr, "Scanning %s...\n", service.Name)
    }
    output, err := service.Call(sess)
    serviceResults := service.Process(output, err, *debug)
    results = append(results, serviceResults...)
}
```

**3. FormatWithSummary Interface Extension**

Add `FormatWithSummary` to the `OutputFormatter` interface. This is a breaking interface change, but all 5 implementors are in our codebase. Every formatter must implement it.

```go
type OutputFormatter interface {
    Format(results []types.ScanResult) (string, error)
    FormatWithSummary(results []types.ScanResult, summary types.ScanSummary) (string, error)
    FileExtension() string
}
```

For JSON and YAML, wrap results in metadata envelope. For CSV/Table/Text, append summary after the data.

**4. Summary Generation Logic**

```go
func GenerateSummary(results []ScanResult, startTime time.Time) ScanSummary {
    serviceMap := make(map[string]bool) // track unique services
    accessibleMap := make(map[string]bool)
    deniedMap := make(map[string]bool)
    resourceCount := 0

    for _, r := range results {
        serviceMap[r.ServiceName] = true
        if r.Error != nil {
            deniedMap[r.ServiceName] = true
        } else {
            accessibleMap[r.ServiceName] = true
            resourceCount++
        }
    }

    return ScanSummary{
        TotalServices:        len(serviceMap),
        AccessibleServices:   len(accessibleMap),
        AccessDeniedServices: len(deniedMap),
        TotalResources:       resourceCount,
        ScanDuration:         time.Since(startTime),
        Timestamp:            startTime,
    }
}
```

Note: A service can appear in BOTH accessible and denied maps (e.g., S3 ListBuckets succeeds but ListObjects fails). Count unique services that had at least one success as "accessible" and those with at least one error as "denied". TotalServices is the union.

**5. Text Summary Output to stderr**

Print the summary to stderr (not stdout) so it doesn't interfere with piped formatted output. Only suppress with `-quiet`.

**6. Banner Suppression**

The ASCII art banner (lines 21-28) should be suppressed in quiet mode since it's informational noise. Wrap with `if !*quiet { ... }`.

### Architecture Compliance

- **Package:** `types` for ScanSummary (alongside ScanResult) -- MUST FOLLOW
- **Package:** `formatters` for FormatWithSummary -- MUST FOLLOW
- **File naming:** `summary.go`, `summary_test.go` (in types/) -- MUST FOLLOW
- **Type naming:** `ScanSummary` (PascalCase exported) -- MUST FOLLOW
- **Error handling:** `fmt.Errorf("... failed: %w", err)` -- MUST FOLLOW
- **Testing:** stdlib `testing` package (no testify) -- MUST FOLLOW
- **Flag package:** stdlib `flag` (already used) -- MUST FOLLOW
- **Output to stderr:** `fmt.Fprintf(os.Stderr, ...)` for progress and summary -- MUST FOLLOW

### Technical Requirements

**Go Version:** 1.19 (existing project standard)

**No New Dependencies Required** -- all functionality uses stdlib:
- `time` -- for Duration, time.Now(), time.Since()
- `fmt` -- for Fprintf to stderr
- `os` -- for os.Stderr

**Existing Dependencies (unchanged):**
- `github.com/aws/aws-sdk-go v1.44.266`
- `github.com/logrusorgru/aurora v2.0.3+incompatible`
- `github.com/olekukonenko/tablewriter v0.0.5`
- `gopkg.in/yaml.v3 v3.0.1`

### File Structure

**Files to CREATE:**
```
cmd/awtest/types/
+-- summary.go                # NEW: ScanSummary struct + GenerateSummary()
+-- summary_test.go           # NEW: Summary generation tests
```

**Files to MODIFY:**
```
cmd/awtest/main.go                         # Add -quiet flag, progress output, summary display, FormatWithSummary call, banner guard
cmd/awtest/utils/output.go                 # Add Quiet variable, check in PrintResult/HandleAWSError
cmd/awtest/formatters/output_formatter.go  # Add FormatWithSummary to interface
cmd/awtest/formatters/json_formatter.go    # Implement FormatWithSummary (metadata envelope)
cmd/awtest/formatters/yaml_formatter.go    # Implement FormatWithSummary (metadata envelope)
cmd/awtest/formatters/csv_formatter.go     # Implement FormatWithSummary (header comments)
cmd/awtest/formatters/table_formatter.go   # Implement FormatWithSummary (appended summary)
cmd/awtest/formatters/text_formatter.go    # Implement FormatWithSummary (appended summary)
```

**Files that ALREADY EXIST (reference only):**
```
cmd/awtest/types/types.go                  # ScanResult struct -- DO NOT MODIFY
cmd/awtest/services/services.go            # AllServices() -- DO NOT MODIFY
cmd/awtest/services/*/                     # All service files -- DO NOT MODIFY
```

### Testing Requirements

**New tests to create:**

1. `cmd/awtest/types/summary_test.go`:
   - Test GenerateSummary with mixed results (success + error) — verify counts
   - Test GenerateSummary with empty results — verify all zeros
   - Test GenerateSummary with all errors — verify accessible=0
   - Test GenerateSummary with all successes — verify denied=0
   - Test duration calculation — verify > 0
   - Test unique service counting — verify deduplication

2. Formatter FormatWithSummary tests (add to existing `*_test.go` files):
   - JSON: verify output is object with "metadata" and "results" keys
   - YAML: verify output has metadata section
   - CSV: verify header comments contain summary info
   - Table: verify summary rows appended
   - Text: verify summary block appended

**Regression verification:**
- All 56+ existing formatter tests must still pass
- `go test ./cmd/awtest/...` -- full test suite
- `go vet ./cmd/awtest/...` -- no lint issues

### Previous Story Intelligence

**Story 1.6 (Format Selection & File Output) -- Key Learnings:**
- Exported `ColorizeMessage` from utils (pattern for exporting utils functions)
- `getFormatter()` factory pattern in main.go — will need update to pass summary
- Format validation happens early (before scan loop) — good pattern to follow for quiet flag
- `text` format + stdout = special path (Process() output) — quiet mode changes this behavior
- `outputFormat` and `outputFile` flags defined at lines 39-40 — add `quiet` flag nearby
- Service loop at lines 143-147 — add progress output here
- Text+stdout early return at line 150-152 — this path needs summary output too (unless quiet)
- **Important note from 1.6:** "When `-format=json` is specified WITHOUT `-output-file`, Process() methods still print colorized output AND formatter prints structured output. This is expected for Story 1.6. **Story 1.7 will add `-quiet` flag to suppress Process() output.**" — This is exactly what we're implementing.

**Story 1.1 (Foundation) -- Key Learnings:**
- OutputFormatter interface: `Format(results []types.ScanResult) (string, error)` + `FileExtension() string`
- 34 services return `[]types.ScanResult` via Process()
- All Process() methods call utils.PrintResult() for output

### Git Intelligence

**Recent commits (newest first):**
```
4a11bbe Mark Story 1.3 YAML output formatter as done
0605d5d Mark Stories 1.1 and 1.2 as done
6f4cf10 Mark Story 1.5 table output formatter as done
9043861 Add table output formatter with 120-char width enforcement (Story 1.5)
fc7ed49 Add CSV output formatter with comprehensive tests (Story 1.4)
```

**Patterns from recent work:**
- Small, focused commits referencing story numbers
- Tests co-located with implementations
- Build verification after each change
- Each story modifies only the files it needs

### Edge Cases to Handle

1. **Default behavior unchanged** — `awtest` with no flags shows: banner + progress + Process() output + summary
2. **Quiet + text stdout** — `awtest -quiet` shows nothing on stderr, Process() output on stdout only... wait, quiet suppresses Process() output too. So `awtest -quiet` with text format actually shows nothing (Process suppressed, summary suppressed, progress suppressed). The formatted results would still go to stdout via the formatter path. Hmm — for text+stdout, the early-return path skips the formatter. Need to handle: if quiet AND text AND stdout, still run formatter to get output, then print it.
3. **Quiet + JSON** — `awtest -quiet -format=json` outputs clean JSON to stdout only — perfect for piping
4. **Quiet + output-file** — `awtest -quiet -format=json -output-file=out.json` writes file, nothing to stderr
5. **Progress with slow services** — Progress messages appear in real-time before each service call, giving user feedback
6. **Duration precision** — Use `time.Since(startTime)` truncated to reasonable precision (e.g., `duration.Truncate(time.Millisecond)`)
7. **Service with mixed results** — S3 ListBuckets succeeds but ListObjects on a specific bucket fails. Service appears in both accessible AND denied. Summary should handle this.

**CRITICAL Edge Case #2 above:** Currently for text+stdout (the default), main.go returns early at line 150-152 because Process() already printed everything. With `-quiet`, Process() output is suppressed, so the user gets NO output. Fix: when quiet is set, DON'T take the early-return path — instead, run the formatter to produce output and print it. This means the text+stdout early return should be: `if *outputFormat == "text" && *outputFile == "" && !*quiet { ... print summary to stderr ... return }`. When quiet, fall through to the formatter path.

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/`)
- New `summary.go` in `types/` package — alongside `types.go`
- No new packages needed
- No new dependencies required
- Interface change to OutputFormatter affects all 5 formatter files

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.7: Progress Tracking & Summary Reporting]
- [Source: _bmad-output/planning-artifacts/prd.md#FR40 — scan metadata timestamp/duration]
- [Source: _bmad-output/planning-artifacts/prd.md#FR46 — quiet mode]
- [Source: _bmad-output/planning-artifacts/prd.md#FR47 — real-time progress]
- [Source: _bmad-output/planning-artifacts/prd.md#FR48 — findings summary]
- [Source: _bmad-output/planning-artifacts/architecture.md#Configuration Management — quiet flag definition]
- [Source: _bmad-output/planning-artifacts/architecture.md#Output Format Architecture — progress/summary mentions]
- [Source: _bmad-output/implementation-artifacts/1-6-format-selection-file-output.md — previous story context]
- [Source: cmd/awtest/main.go — current main with formatter integration]
- [Source: cmd/awtest/utils/output.go — PrintResult/HandleAWSError for quiet suppression]
- [Source: cmd/awtest/types/types.go — ScanResult struct, AWSService struct]
- [Source: cmd/awtest/formatters/output_formatter.go — OutputFormatter interface]
- [Source: cmd/awtest/formatters/json_formatter.go — JSON envelope pattern reference]
- [Source: cmd/awtest/services/services.go — AllServices() with 34 services]
- [Source: cmd/awtest/services/s3/calls.go — example Process() with PrintResult calls]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No blockers or debug issues encountered.

### Completion Notes List

- Implemented `-quiet` flag with package-level `utils.Quiet` variable — zero changes to 34 service files
- Added real-time progress messages to stderr (`Scanning [service]...`) with quiet guard
- Created `types.ScanSummary` struct and `GenerateSummary()` function for aggregate scan stats
- Added `FormatWithSummary()` to `OutputFormatter` interface, implemented in all 5 formatters:
  - JSON/YAML: metadata envelope wrapping results
  - CSV: header comment rows with summary
  - Table/Text: appended summary section
- Summary and progress output goes to stderr to avoid contaminating piped stdout
- Banner suppressed in quiet mode
- Text+stdout default path prints summary to stderr; quiet+text falls through to formatter path
- HandleAWSError still detects InvalidKeyError in quiet mode for proper abort handling
- 8 new summary unit tests, 5 new FormatWithSummary tests across all formatters
- All 113 test cases pass, `go build` and `go vet` clean

### Change Log

- 2026-03-03: Story 1.7 implementation complete — progress tracking, summary reporting, quiet mode (all 8 tasks)
- 2026-03-03: Code review fix — AC6 violation: quiet mode now uses Format() instead of FormatWithSummary() to suppress summary

### File List

**New files:**
- cmd/awtest/types/summary.go
- cmd/awtest/types/summary_test.go

**Modified files:**
- cmd/awtest/main.go
- cmd/awtest/utils/output.go
- cmd/awtest/formatters/output_formatter.go
- cmd/awtest/formatters/json_formatter.go
- cmd/awtest/formatters/yaml_formatter.go
- cmd/awtest/formatters/csv_formatter.go
- cmd/awtest/formatters/table_formatter.go
- cmd/awtest/formatters/text_formatter.go (previously untracked new file from Story 1.6, modified here to add FormatWithSummary)
- cmd/awtest/formatters/output_formatter_test.go
- cmd/awtest/formatters/json_formatter_test.go
- cmd/awtest/formatters/yaml_formatter_test.go
- cmd/awtest/formatters/csv_formatter_test.go
- cmd/awtest/formatters/table_formatter_test.go
- cmd/awtest/formatters/text_formatter_test.go (previously untracked new file from Story 1.6, modified here to add FormatWithSummary test)
