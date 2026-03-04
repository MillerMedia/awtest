# Story 1.3: YAML Output Formatter

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security professional creating readable reports,
I want YAML-formatted scan results,
so that I can produce human-readable structured output for documentation and reports.

## Acceptance Criteria

**Given** the OutputFormatter interface from Story 1.1
**When** implementing YAML output format
**Then:**

1. **AC1:** Create YamlFormatter struct in `cmd/awtest/formatters/yaml_formatter.go` implementing OutputFormatter interface
2. **AC2:** Add `gopkg.in/yaml.v3` dependency to go.mod
3. **AC3:** `Format()` method produces valid YAML with proper indentation and structure
4. **AC4:** YAML includes all ScanResult fields with human-readable formatting
5. **AC5:** `FileExtension()` returns `"yaml"`
6. **AC6:** Handle empty results gracefully (return valid empty YAML sequence `[]`)
7. **AC7:** Handle errors in `ScanResult.Error` field by including error message string in YAML output
8. **AC8:** Write table-driven unit tests in `yaml_formatter_test.go` covering: valid results, empty results, results with errors, special characters in resource names
9. **AC9:** Verify YAML output can be parsed by standard YAML tools (Go yaml.Unmarshal round-trip)
10. **AC10:** Test passes: `go test ./cmd/awtest/formatters/...`

## Tasks / Subtasks

- [x] Add gopkg.in/yaml.v3 dependency (AC: 2)
  - [x] Run `go get gopkg.in/yaml.v3`
  - [x] Verify go.mod updated with yaml.v3 dependency
- [x] Create YAMLFormatter implementation (AC: 1, 3, 4, 5, 6, 7)
  - [x] Create `cmd/awtest/formatters/yaml_formatter.go`
  - [x] Define `yamlScanResult` serialization struct with YAML tags
  - [x] Implement `NewYAMLFormatter()` constructor
  - [x] Implement `Format()` with yaml.Marshal
  - [x] Implement `FileExtension()` returning "yaml"
  - [x] Handle nil Details as empty map `{}`
  - [x] Handle Error field conversion to string with omitempty
  - [x] Handle empty results returning `[]`
  - [x] Add resilient Details serialization (mirror JSON pattern)
- [x] Write comprehensive tests (AC: 8, 9)
  - [x] Create `cmd/awtest/formatters/yaml_formatter_test.go`
  - [x] Test: single valid result with all fields
  - [x] Test: empty results returns `[]`
  - [x] Test: result with error field
  - [x] Test: special characters in resource names
  - [x] Test: multiple results
  - [x] Test: nil details serialized as empty map
  - [x] Test: timestamp in RFC3339 format
  - [x] Test: successful result has no error field
  - [x] Test: YAML round-trip (marshal then unmarshal)
  - [x] Test: interface compliance check
  - [x] Test: constructor not nil
  - [x] Test: FileExtension returns "yaml"
  - [x] Test: resilient serialization with bad Details
- [x] Verify build and tests pass (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/formatters/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -cover ./cmd/awtest/formatters/...` (target: 80%+)
- [x] Run full regression (AC: 10)
  - [x] `go test ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow the JSONFormatter Pattern Exactly

The YAML formatter must mirror the JSONFormatter implementation pattern established in Stories 1.1 and 1.2. The JSON formatter was refined through code review and represents the canonical pattern.

**Pattern Source:** `cmd/awtest/formatters/json_formatter.go` (post code-review version)

**Key Pattern Elements:**
1. **Separate serialization struct** (`yamlScanResult`) with YAML tags - do NOT marshal `types.ScanResult` directly (the `Error` field is `error` type, not serializable)
2. **Compact output** - Use `yaml.Marshal` (not custom indentation) for clean YAML
3. **Resilient Details serialization** - Validate Details map can be marshaled; if not, use empty map + error marker
4. **Nil Details → empty map** `{}` - Never let nil Details serialize as `null`
5. **Error field as string** with omitempty - Absent when no error
6. **Timestamp as RFC3339 string** - Use `time.RFC3339` constant, not hardcoded format
7. **Constructor function** - `NewYAMLFormatter()` returns `*YAMLFormatter`
8. **Error wrapping** - `fmt.Errorf("yaml formatting failed: %w", err)`

### Existing JSONFormatter Reference (Post Code-Review)

```go
// json_formatter.go - CANONICAL PATTERN TO FOLLOW
type jsonScanResult struct {
    ServiceName  string                 `json:"serviceName"`
    MethodName   string                 `json:"methodName"`
    ResourceType string                 `json:"resourceType"`
    ResourceName string                 `json:"resourceName"`
    Details      map[string]interface{} `json:"details"`
    Error        string                 `json:"error,omitempty"`
    Timestamp    string                 `json:"timestamp"`
}

func (f *JSONFormatter) Format(results []types.ScanResult) (string, error) {
    jsonResults := make([]jsonScanResult, 0, len(results))
    for _, r := range results {
        jr := jsonScanResult{
            ServiceName:  r.ServiceName,
            MethodName:   r.MethodName,
            ResourceType: r.ResourceType,
            ResourceName: r.ResourceName,
            Details:      map[string]interface{}{},
            Timestamp:    r.Timestamp.Format(time.RFC3339),
        }
        if r.Details != nil {
            if _, err := json.Marshal(r.Details); err == nil {
                jr.Details = r.Details
            } else {
                jr.Error = fmt.Sprintf("details serialization error: %v", err)
            }
        }
        if r.Error != nil {
            if jr.Error != "" {
                jr.Error = r.Error.Error() + "; " + jr.Error
            } else {
                jr.Error = r.Error.Error()
            }
        }
        jsonResults = append(jsonResults, jr)
    }
    data, err := json.Marshal(jsonResults)
    if err != nil {
        return "", fmt.Errorf("json formatting failed: %w", err)
    }
    return string(data), nil
}
```

### YAML-Specific Implementation Notes

**Serialization Struct for YAML:**
```go
// yamlScanResult is the YAML-serializable representation of a ScanResult.
type yamlScanResult struct {
    ServiceName  string                 `yaml:"serviceName"`
    MethodName   string                 `yaml:"methodName"`
    ResourceType string                 `yaml:"resourceType"`
    ResourceName string                 `yaml:"resourceName"`
    Details      map[string]interface{} `yaml:"details"`
    Error        string                 `yaml:"error,omitempty"`
    Timestamp    string                 `yaml:"timestamp"`
}
```

**YAML Empty Results:** `yaml.Marshal([]yamlScanResult{})` produces `[]\n`. This is valid YAML. Verify this behavior in tests.

**YAML Special Characters:** YAML auto-quotes strings containing special chars (`:`, `#`, `{`, etc.). The `yaml.v3` library handles this automatically. Add a test case with resource names like `"arn:aws:s3:::my-bucket"` and `"key with: colon"`.

**YAML vs JSON Differences to Note:**
- YAML uses `omitempty` tag similarly to JSON - empty/zero values omitted
- YAML produces multi-line output (not single-line like compact JSON)
- YAML uses indentation-based structure (handled by yaml.v3 library)
- `yaml.Marshal` returns `[]byte` with trailing newline - trim or keep consistently

### Architecture Compliance

**From Architecture Document:**

- **Formatter Interface Pattern:** OutputFormatter with `Format()` and `FileExtension()` - MUST FOLLOW
- **Package:** `formatters` (lowercase, no underscores) - MUST FOLLOW
- **File naming:** `yaml_formatter.go` (underscores for multi-word) - MUST FOLLOW
- **Type naming:** `YAMLFormatter` (PascalCase exported, YAML all-caps per Go conventions) - MUST FOLLOW
- **Constructor:** `NewYAMLFormatter()` (Go convention) - MUST FOLLOW
- **Error handling:** Returns `fmt.Errorf("yaml formatting failed: %w", err)` - MUST FOLLOW
- **New dependency:** `gopkg.in/yaml.v3` - explicitly specified in architecture document
- **Testing:** Testify not required (stdlib testing used in existing formatter tests)

**YAML Library Decision:**
- Architecture explicitly chose `gopkg.in/yaml.v3`
- go.sum has yaml.v2 v2.2.8 (transitive), but go.mod does NOT have it
- Must add yaml.v3 as explicit dependency: `go get gopkg.in/yaml.v3`

### Technical Requirements

**Go Version:** 1.19 (existing project standard)

**New Dependency:**
- `gopkg.in/yaml.v3` (add via `go get gopkg.in/yaml.v3`)

**Existing Dependencies (unchanged):**
- `github.com/aws/aws-sdk-go v1.44.266`
- `github.com/logrusorgru/aurora v2.0.3+incompatible`
- `encoding/json` (stdlib, used by JSON formatter)

**Build Verification:**
```bash
go get gopkg.in/yaml.v3          # Add dependency
go build ./cmd/awtest             # Must succeed
go test ./cmd/awtest/formatters/... # Must pass
go vet ./cmd/awtest/...           # Must pass
```

### File Structure

**Files to CREATE:**
```
cmd/awtest/formatters/
├── yaml_formatter.go          # NEW: YAMLFormatter implementation
└── yaml_formatter_test.go     # NEW: YAML formatter tests
```

**Files to MODIFY:**
```
go.mod                          # Add gopkg.in/yaml.v3 dependency
go.sum                          # Auto-updated by go get
```

**Files that ALREADY EXIST (DO NOT MODIFY):**
```
cmd/awtest/formatters/
├── output_formatter.go        # OutputFormatter interface (from Story 1.1)
├── output_formatter_test.go   # Interface tests (from Story 1.1)
├── json_formatter.go          # JSONFormatter (from Stories 1.1/1.2) - REFERENCE ONLY
└── json_formatter_test.go     # JSON tests (from Stories 1.1/1.2) - REFERENCE ONLY

cmd/awtest/types/types.go      # ScanResult struct (from Story 1.1) - DO NOT MODIFY
cmd/awtest/main.go             # Result collection (from Story 1.1) - DO NOT MODIFY
cmd/awtest/services/           # All service files - DO NOT MODIFY
```

### Previous Story Intelligence

**Story 1.2 (JSON Output Formatter) - Key Learnings:**
- JSON formatter was already implemented during Story 1.1, Story 1.2 was primarily validation
- Code review findings (applied post-review, now canonical pattern):
  - Switched `json.MarshalIndent` to `json.Marshal` for compact SIEM-friendly output
  - Added per-result Details validation for resilient serialization
  - Replaced hardcoded time format string with `time.RFC3339` constant
  - Nil Details map now serializes as `{}` instead of `null`
- 20 tests pass, coverage 88.9%
- Fixed pre-existing `go vet` issue in `cmd/awtest/utils/output.go` (unkeyed struct literal)

**Story 1.1 (Formatter Interface & Result Collection) - Foundation:**
- All 34 services return `[]types.ScanResult` from Process methods
- Backward compatibility maintained: `utils.PrintResult()` calls kept in Process methods
- Test coverage: types 100%, formatters 91.7%
- Patterns established: table-driven tests, interface compliance checks, separate serialization struct
- JSON formatter was implemented ahead of schedule during this story

**Patterns to Reuse:**
- Interface compliance check: `var _ OutputFormatter = (*YAMLFormatter)(nil)`
- Table-driven tests with `t.Run()` sub-tests
- Separate serialization struct (`yamlScanResult`) for YAML tags
- Constructor function: `NewYAMLFormatter()`
- Resilient serialization: validate Details before marshaling
- Nil Details → empty map pattern
- Error as string with omitempty

### Git Intelligence

**Recent Commits:**
```
ed241ba Add Rekognition service enumeration (ListCollections, ListStreamProcessors, DescribeProjects)
f7dcee9 Validate JSON formatter and address code review findings (Story 1.2)
6017a06 Add .gitignore, tests, and update service calls with result collection
c9d801d Add awtest source code to private working repo
```

**Patterns from Recent Work:**
- Small, incremental commits with descriptive messages
- Tests co-located with implementations
- Build verification after each change
- Code review findings addressed in subsequent commits
- Pre-existing issues fixed when encountered (go vet errors)

### Edge Cases to Handle

1. **Empty results** → `yaml.Marshal([]yamlScanResult{})` should produce `[]\n` - verify this
2. **Nil Details map** → Initialize as `map[string]interface{}{}` (serialize as `details: {}`)
3. **Error with no other fields** → Error string present, other fields at zero values
4. **Special characters in ResourceName** → YAML auto-quotes: `"arn:aws:s3:::bucket"`, `"key: with colon"`, `"hash#value"`
5. **Large result sets** → `yaml.Marshal` handles efficiently
6. **Concurrent access** → YAMLFormatter is stateless (no shared state), safe by design
7. **Unserializable Details values** → Resilient serialization: catch marshal error, use empty Details + error marker
8. **Trailing newline** → `yaml.Marshal` adds trailing `\n` - keep for consistency with YAML spec

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/`)
- Follows package-per-directory convention
- Tests co-located with implementation (`*_test.go` in same directory)
- Clear separation: `types` package for data, `formatters` package for presentation
- No conflicts with existing file structure

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.3: YAML Output Formatter]
- [Source: _bmad-output/planning-artifacts/architecture.md#Output Format Architecture]
- [Source: _bmad-output/planning-artifacts/architecture.md#Implementation Patterns]
- [Source: _bmad-output/planning-artifacts/architecture.md#Dependencies - gopkg.in/yaml.v3]
- [Source: _bmad-output/implementation-artifacts/1-2-json-output-formatter.md#Completion Notes]
- [Source: _bmad-output/implementation-artifacts/1-1-formatter-interface-result-collection.md#Completion Notes]
- [Source: cmd/awtest/formatters/json_formatter.go] - Canonical implementation pattern
- [Source: cmd/awtest/formatters/json_formatter_test.go] - Test pattern reference
- [Source: cmd/awtest/formatters/output_formatter.go] - Interface definition
- [Source: cmd/awtest/types/types.go] - ScanResult struct definition

## Dev Agent Record

### Agent Model Used

Story created by: claude-opus-4-6 (Claude Code CLI)
Story implemented by: claude-opus-4-6 (Claude Code CLI)

### Debug Log References

- yaml.v3 panics on unserializable types (func, chan) instead of returning an error like encoding/json. Added `tryMarshalYAML()` helper with `recover()` to handle this safely in resilient Details serialization, preserving the specific panic message for debugging.

### Completion Notes List

- Implemented YAMLFormatter following the canonical JSONFormatter pattern from Story 1.2
- Created `yamlScanResult` serialization struct with YAML tags (camelCase field names)
- `Format()` converts `[]types.ScanResult` to YAML using `yaml.v3` Marshal
- `FileExtension()` returns `"yaml"`
- Nil Details maps serialize as empty `{}`, not null
- Error field uses string with omitempty (absent when no error)
- Timestamps formatted as RFC3339 strings
- Resilient Details serialization uses `tryMarshalYAML()` with panic recovery, preserving specific error/panic messages (yaml.v3 panics on unserializable types unlike encoding/json which returns errors)
- Empty results produce valid YAML sequence `[]`
- 13 YAML-specific tests added (33 total formatter tests pass)
- Test coverage: 90.2% (exceeds 80% target)
- All acceptance criteria (AC1-AC10) satisfied
- No regressions in existing tests
- go build, go vet, go test all pass cleanly

### Change Log

- 2026-03-03: Implemented YAML output formatter (Story 1.3) - YAMLFormatter with full test suite
- 2026-03-03: Addressed code review findings - 1 Medium resolved (preserved specific error detail in resilient serialization via `tryMarshalYAML()`), 1 Low resolved (panic message now returned for debugging)

### File List

- cmd/awtest/formatters/yaml_formatter.go (NEW)
- cmd/awtest/formatters/yaml_formatter_test.go (NEW)
- go.mod (MODIFIED - added gopkg.in/yaml.v3 v3.0.1)
- go.sum (MODIFIED - auto-updated by go get)
