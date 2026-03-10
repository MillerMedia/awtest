---
title: 'Scan Output Hit Visibility'
slug: 'scan-output-hit-visibility'
created: '2026-03-08'
status: 'completed'
stepsCompleted: [1, 2, 3, 4]
tech_stack: [go, aurora, tablewriter]
files_to_modify: [cmd/awtest/utils/output.go, cmd/awtest/types/summary.go, cmd/awtest/main.go, cmd/awtest/formatters/text_formatter.go, cmd/awtest/formatters/table_formatter.go, cmd/awtest/formatters/json_formatter.go, cmd/awtest/formatters/yaml_formatter.go, cmd/awtest/utils/output_test.go, cmd/awtest/types/summary_test.go, cmd/awtest/formatters/text_formatter_test.go, cmd/awtest/formatters/table_formatter_test.go, cmd/awtest/formatters/json_formatter_test.go, cmd/awtest/formatters/yaml_formatter_test.go]
code_patterns: [aurora coloring with severity string, ScanSummary computed by GenerateSummary with map tracking, text/table FormatWithSummary renders summary block, json/yaml use private metadata structs mapped from ScanSummary]
test_patterns: [stdlib testing only, t.Errorf/t.Fatalf, table-driven tests, formatter tests in formatters/*_test.go]
---

# Tech-Spec: Scan Output Hit Visibility

**Created:** 2026-03-08

## Overview

### Problem Statement

Accessible services blend into a wall of "Access denied" lines during scans, and the scan summary only shows a count (e.g., "Accessible: 1") without listing which services were accessible. Pentesters can't quickly identify what they have access to — the most critical information from a scan.

### Solution

(1) Add a `[hit]` severity with green coloring for accessible service results so they visually pop in inline output. (2) Add an accessible services list with full method names to the scan summary, capped at 20 entries with a "see full report" overflow message.

### Scope

**In Scope:**
- New `"hit"` severity in `DetermineSeverity()` for nil-error results
- Green coloring for `[hit]` in `ColorizeMessage()`
- `AccessibleMethodNames []string` field added to `ScanSummary`
- `GenerateSummary()` collects method names for successful results
- `printTextSummary()` renders hit list (capped at 20)
- Text and table formatters show the hit list
- All formatters get the field in `ScanSummary` for free (JSON/YAML/CSV include it in data)

**Out of Scope:**
- Changing JSON/YAML/CSV summary rendering logic (they get the data, consumers filter)
- Changing error severity levels or other coloring
- Resource detail display in summary

## Context for Development

### Codebase Patterns

- All service output goes through `utils.PrintResult()` / `utils.HandleAWSError()` / `utils.PrintAccessGranted()`
- `DetermineSeverity(err error) string` currently always returns `"info"` regardless of error — single place to change for severity logic
- `ColorizeMessage()` uses aurora package — severity coloring: `"high"` → red, everything else → blue. Adding `"hit"` → green is a simple branch
- `ScanSummary` struct in `types/summary.go` has count fields only. `GenerateSummary()` tracks unique services via `map[string]bool` — already iterates all results, easy to collect method names
- `printTextSummary()` in `main.go:247-258` writes to stderr — used for default text+stdout+non-quiet path
- Text/table `FormatWithSummary()` methods render identical summary blocks — both need hit list
- JSON formatter has private `jsonMetadata` struct (line 68-75) mapped from `ScanSummary` — needs new field
- YAML formatter has private `yamlMetadata` struct (line 82-89) mapped from `ScanSummary` — needs new field
- CSV formatter renders summary as comment rows — would need method name list in comments
- `PrintAccessGranted()` calls `PrintResult()` with `err=nil` — so `DetermineSeverity(nil)` returning `"hit"` would automatically colorize all access-granted lines

### Files to Reference

| File | Purpose | Key Lines |
| ---- | ------- | --------- |
| `cmd/awtest/utils/output.go` | `DetermineSeverity()`, `ColorizeMessage()`, `PrintResult()` | L22-24 (severity), L28-48 (colorize), L50-63 (print) |
| `cmd/awtest/types/summary.go` | `ScanSummary` struct, `GenerateSummary()` | L6-13 (struct), L18-42 (generate) |
| `cmd/awtest/main.go` | `printTextSummary()` | L240-251 (approx — locate by function name) |
| `cmd/awtest/formatters/text_formatter.go` | Text `FormatWithSummary()` | L61-81 |
| `cmd/awtest/formatters/table_formatter.go` | Table `FormatWithSummary()` | L65-85 |
| `cmd/awtest/formatters/json_formatter.go` | `jsonMetadata` struct | L68-75 |
| `cmd/awtest/formatters/yaml_formatter.go` | `yamlMetadata` struct | L82-89 |

### Technical Decisions

- Use `"hit"` as severity string (not `"success"` or `"granted"`) — short, pentesting-relevant, visually distinct
- `DetermineSeverity(nil)` returns `"hit"`, `DetermineSeverity(non-nil)` returns `"info"` — minimal change, all access-granted lines auto-colored
- Cap summary hit list at 20 entries — define `maxAccessibleMethodsInSummary = 20` constant in `types/summary.go` and reuse across `printTextSummary`, text formatter, and table formatter to avoid magic numbers
- Collect full method names (e.g., `rekognition:ListCollections`) not just service names, so pentesters know exactly which API calls succeeded
- Use `AccessibleMethodNames []string` (not `AccessibleServiceNames`) to be accurate — it's method-level granularity
- JSON/YAML metadata structs get the new field added for structured consumers; CSV out of scope for now

## Implementation Plan

### Tasks

- [x] Task 1: Update `DetermineSeverity()` to return `"hit"` for nil errors
  - File: `cmd/awtest/utils/output.go`
  - Action: Change `DetermineSeverity(err error) string` from always returning `"info"` to: return `"hit"` when `err == nil`, return `"info"` when `err != nil`
  - Notes: This single change auto-applies to all `PrintResult()` and `PrintAccessGranted()` callers. The text formatter's `Format()` also calls `DetermineSeverity(r.Error)` so formatted output gets the severity too.

- [x] Task 2: Add green coloring for `"hit"` severity in `ColorizeMessage()`
  - File: `cmd/awtest/utils/output.go`
  - Action: In `ColorizeMessage()`, add a branch for `severity == "hit"`: use `aurora.BrightGreen(severity).String()`. The existing branches are `"high"` → red, else → blue. Insert `"hit"` → bright green before the else.
  - Notes: This makes `[hit]` appear in bright green in all colorized output (inline results and text formatter).

- [x] Task 3: Add `AccessibleMethodNames` field and cap constant to `ScanSummary`
  - File: `cmd/awtest/types/summary.go`
  - Action: Add `AccessibleMethodNames []string` field to the `ScanSummary` struct, after `TotalResources`. Add `const MaxAccessibleMethodsInSummary = 20` (exported so main.go and formatters can use it).
  - Notes: Field contains full method names (e.g., `"rekognition:ListCollections"`) for results where `Error == nil`. Sorted alphabetically for deterministic output. The constant is used by Tasks 5-7 for the cap.

- [x] Task 4: Collect accessible method names in `GenerateSummary()`
  - File: `cmd/awtest/types/summary.go`
  - Action: In the `GenerateSummary()` loop, when `r.Error == nil`, append `r.MethodName` to a `[]string` slice. After the loop, sort the slice with `sort.Strings()` and assign to `ScanSummary.AccessibleMethodNames`. Use a map to deduplicate method names before collecting.
  - Notes: Add `"sort"` to imports. Dedup is important because the same method can appear multiple times if a service returns multiple resources.

- [x] Task 5: Render hit list in `printTextSummary()`
  - File: `cmd/awtest/main.go`
  - Action: In `printTextSummary()`, after the "Resources Found" line and before the closing `========`, add a section that lists accessible method names. Cap at 20 entries. Format:
    ```
    Accessible Methods:
      - rekognition:ListCollections
      - rekognition:DescribeProjects
      ... (N more - use --format=json for full list)
    ```
    Only render this section if `len(summary.AccessibleMethodNames) > 0`. If more than 20, show first 20 and append overflow message.
  - Notes: Write to stderr (consistent with rest of summary). Use `aurora.BrightGreen()` for the method names to maintain visual consistency with `[hit]` inline output.

- [x] Task 6: Render hit list in text formatter `FormatWithSummary()`
  - File: `cmd/awtest/formatters/text_formatter.go`
  - Action: In `FormatWithSummary()`, after "Resources Found" line and before closing `========`, add the same accessible methods section as Task 5 (cap at 20, overflow message). No aurora coloring here — this path is used for file output and piped output where ANSI codes may not render.
  - Notes: Plain text, no coloring. The formatter is used when output goes to file or non-default format paths.

- [x] Task 7: Render hit list in table formatter `FormatWithSummary()`
  - File: `cmd/awtest/formatters/table_formatter.go`
  - Action: Same as Task 6 — add accessible methods section to the summary block in `FormatWithSummary()`. Plain text, cap at 20.

- [x] Task 8: Add `AccessibleMethodNames` to JSON metadata struct
  - File: `cmd/awtest/formatters/json_formatter.go`
  - Action: Add `AccessibleMethodNames []string \`json:"accessibleMethodNames,omitempty"\`` to `jsonMetadata` struct. Map from `summary.AccessibleMethodNames` in `FormatWithSummary()`.
  - Notes: `omitempty` keeps output clean when no methods are accessible.

- [x] Task 9: Add `AccessibleMethodNames` to YAML metadata struct
  - File: `cmd/awtest/formatters/yaml_formatter.go`
  - Action: Add `AccessibleMethodNames []string \`yaml:"accessibleMethodNames,omitempty"\`` to `yamlMetadata` struct. Map from `summary.AccessibleMethodNames` in `FormatWithSummary()`.

- [x] Task 10: Update and add tests
  - Files: `cmd/awtest/utils/output_test.go` (new or existing), `cmd/awtest/types/summary_test.go` (new or existing), `cmd/awtest/formatters/text_formatter_test.go`, `cmd/awtest/formatters/table_formatter_test.go`, `cmd/awtest/formatters/json_formatter_test.go`, `cmd/awtest/formatters/yaml_formatter_test.go`
  - Action:
    - Test `DetermineSeverity(nil)` returns `"hit"`, `DetermineSeverity(errors.New("x"))` returns `"info"`
    - Test `ColorizeMessage()` with severity `"hit"` produces output containing green ANSI codes
    - Test `GenerateSummary()` populates `AccessibleMethodNames` with deduplicated, sorted method names
    - Test `GenerateSummary()` with all-error results produces empty `AccessibleMethodNames`
    - Update existing formatter `FormatWithSummary` tests to verify "Accessible Methods:" section appears when hits exist
    - Update existing formatter tests to verify no "Accessible Methods:" section when all results are errors
    - Test JSON metadata includes `accessibleMethodNames` field
    - Test YAML metadata includes `accessibleMethodNames` field
    - Test hit list cap at 20 entries with overflow message
  - Notes: stdlib `testing` only. Table-driven where appropriate.

### Acceptance Criteria

- [ ] AC 1: Given a scan with accessible services, when results are printed inline (text mode, non-quiet), then accessible service lines show `[hit]` in bright green instead of `[info]` in blue.
- [ ] AC 2: Given a scan with error results (access denied), when results are printed inline, then error lines continue to show `[info]` in blue (unchanged behavior).
- [ ] AC 3: Given a scan with accessible services, when the scan summary is displayed (text mode), then the summary includes an "Accessible Methods:" section listing the full method names (e.g., `rekognition:ListCollections`).
- [ ] AC 4: Given a scan with more than 20 accessible methods, when the summary is displayed, then only the first 20 are listed with a message indicating remaining count (e.g., "... (5 more - use --format=json for full list)").
- [ ] AC 5: Given a scan with zero accessible services, when the summary is displayed, then no "Accessible Methods:" section appears.
- [ ] AC 6: Given a scan with `--format=json`, when results are output, then the JSON metadata includes an `accessibleMethodNames` array with the full method names.
- [ ] AC 7: Given a scan with `--format=yaml`, when results are output, then the YAML metadata includes an `accessibleMethodNames` list with the full method names.
- [ ] AC 8: Given a scan with `--format=table`, when the summary is rendered, then the summary includes the accessible methods list (same as text format).
- [ ] AC 9: Given duplicate method names in results (e.g., multiple resources from same API call), when the summary is generated, then method names are deduplicated (each appears once).
- [ ] AC 10: Given accessible method names in the summary, when displayed, then they are sorted alphabetically for consistent output.

## Additional Context

### Dependencies

None — uses existing `aurora` coloring library and `sort` from stdlib, both already available.

### Testing Strategy

- **Unit tests:** `DetermineSeverity()` severity logic, `ColorizeMessage()` green output, `GenerateSummary()` method name collection/dedup/sort
- **Formatter tests:** Update existing `FormatWithSummary` tests in all formatter test files to verify hit list rendering, cap behavior, and structured field inclusion
- **Integration:** Manual test with real AWS credentials — verify `[hit]` lines visually pop in green, summary lists accessible methods
- **Edge cases:** All-denied scan (no hits section), all-accessible scan (many hits, cap triggers), empty scan, single hit

### Notes

- The `DetermineSeverity()` change is the highest-leverage modification — one line change auto-colors all accessible results across inline output and text formatter
- The text formatter's `Format()` method already calls `DetermineSeverity(r.Error)` and passes to `ColorizeMessage()`, so both inline output (via `PrintResult`) and formatted output (via `TextFormatter.Format`) get green `[hit]` automatically
- CSV formatter is intentionally excluded from hit list rendering — CSV is consumed programmatically, not visually
- Future enhancement: could add resource counts per accessible method (e.g., `rekognition:ListCollections (3 resources)`)

## Review Notes
- Adversarial review completed
- Findings: 12 total, 3 fixed, 9 skipped (noise/pre-existing/out-of-scope)
- Resolution approach: auto-fix
- F1 (fixed): `printTextSummary` now respects terminal detection via `isTerminal()` for aurora coloring
- F4 (fixed): Extracted `FormatAccessibleMethods()` helper in `types/summary.go` to eliminate 3x copy-paste
- F8 (fixed): Test uses `aurora.BrightGreen("hit").String()` instead of raw ANSI code
