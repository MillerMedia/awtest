# Story 10.3: Service Template Concurrent-Safe Update

Status: review

<!-- Generated: 2026-03-13 by BMAD Create Story Workflow -->
<!-- Epic: 10 - Documentation & Contributor Enablement (Phase 2 Epic 5) -->
<!-- FRs: FR107 | Source: epics-phase2.md#Story 5.3 -->
<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a contributor,
I want the service template updated for concurrent-safe patterns,
So that new services I create automatically follow the correct patterns for parallel execution.

## Acceptance Criteria

1. **AC1:** The template `calls.go.tmpl` follows the existing `Call(ctx, sess)`/`Process(output, err, debug)` pattern with **no** `sync`, `sync/atomic`, or concurrency imports.

2. **AC2:** The template `calls.go.tmpl` includes a prominent comment block at the top stating that services must remain concurrency-unaware — the worker pool and `safeScan` wrapper handle all parallelism and panic recovery transparently.

3. **AC3:** A test template file (`calls_test.go.tmpl`) is added to `_template/` that includes a note about race detection testing (`make test` runs with `-race` by default) and demonstrates the table-driven test pattern used by all services.

4. **AC4:** The template README (`_template/README.md`) is updated to include a "Concurrent Safety" section explaining:
   - Services are automatically executed concurrently at `--speed=fast/insane`
   - No concurrency-specific code is needed
   - What NOT to do (import sync, spawn goroutines, use global mutable state)

5. **AC5:** The `example_calls.go.reference` file includes the concurrency-awareness comment consistent with the updated template.

6. **AC6:** New services created from the template register in `GetServices()` with alphabetical ordering maintained and work correctly at all concurrency levels (1-20 workers) without modification.

7. **AC7:** The `calls_test.go.tmpl` test template demonstrates testing a single API call's `Process()` function with at minimum: success case, error case, empty results case, and type assertion failure case.

## Tasks / Subtasks

- [x] Task 1: Update `calls.go.tmpl` with concurrency-awareness header (AC: 1, 2)
  - [x] Add a comment block after the package declaration explaining the concurrent safety contract
  - [x] Verify no sync/sync/atomic imports exist (already the case — just confirm)
  - [x] Add inline comment near `Call()` noting it executes via `safeScan` wrapper automatically

- [x] Task 2: Create `calls_test.go.tmpl` test template (AC: 3, 7)
  - [x] Create `cmd/awtest/services/_template/calls_test.go.tmpl`
  - [x] Include table-driven test pattern matching existing service tests (e.g., neptune, sagemaker)
  - [x] Include test cases: success, error, empty results, type assertion failure
  - [x] Add comment noting `make test` runs with `-race` flag by default — race conditions in tests will be caught
  - [x] Include placeholder patterns consistent with `calls.go.tmpl` (SERVICENAME, DISPLAYNAME, etc.)

- [x] Task 3: Update `_template/README.md` with concurrent safety section (AC: 4)
  - [x] Add "Concurrent Safety" section after the existing Steps section
  - [x] Explain that services run concurrently via worker pool — no code needed
  - [x] List anti-patterns: no `sync` imports, no goroutines, no global mutable state, no stdout writes
  - [x] Add a step to the Steps section for creating tests (referencing the new test template)

- [x] Task 4: Update `example_calls.go.reference` with concurrency note (AC: 5)
  - [x] Add matching concurrency-awareness comment header consistent with updated `calls.go.tmpl`
  - [x] Keep all existing annotations intact

- [x] Task 5: Verify template produces concurrent-safe services (AC: 6)
  - [x] Confirm template output has no sync imports
  - [x] Confirm template follows Call/Process pattern compatible with safeScan wrapper
  - [x] Confirm README documents alphabetical registration in GetServices()

## Dev Notes

### This is a Documentation/Template-Only Story

This story modifies files in `cmd/awtest/services/_template/` only. No Go source code compilation or test execution needed. The template files use `.tmpl` and `.reference` extensions — they are NOT compiled by Go.

### Files to MODIFY

```
cmd/awtest/services/_template/calls.go.tmpl           # Add concurrency-awareness comment
cmd/awtest/services/_template/README.md                # Add concurrent safety section + test step
cmd/awtest/services/_template/example_calls.go.reference  # Add matching concurrency comment
```

### Files to CREATE

```
cmd/awtest/services/_template/calls_test.go.tmpl       # New test template with race detection note
```

### Files to REFERENCE (DO NOT MODIFY)

```
cmd/awtest/services/neptune/calls.go                   # Recent service — pattern reference
cmd/awtest/services/sagemaker/calls_test.go            # Test pattern reference
cmd/awtest/services/services.go                        # GetServices() registry pattern
cmd/awtest/worker_pool.go                              # Worker pool — understand concurrency model
cmd/awtest/safe_scan.go                                # safeScan wrapper — understand service contract
CONTRIBUTING.md                                        # Concurrent testing docs (Story 10.2)
Makefile                                               # make test includes -race flag
```

### Concurrency-Awareness Comment to Add

The following comment block should be added to `calls.go.tmpl` and `example_calls.go.reference` after the package declaration and before the import block:

```go
// CONCURRENT SAFETY: This service is concurrency-unaware by design.
// The worker pool (worker_pool.go) executes services concurrently at --speed=fast/insane.
// The safeScan wrapper handles panic recovery and error classification automatically.
// DO NOT import "sync", "sync/atomic", or spawn goroutines in this file.
// DO NOT use global mutable state. Each Call() invocation must be self-contained.
```

### Test Template Pattern

Based on existing service tests (sagemaker, neptune, etc.), the test template should follow this structure:

```go
package SERVICENAME

import (
    "errors"
    "testing"
    "time"

    "github.com/aws/aws-sdk-go/service/AWSSDKPACKAGE"
)

// NOTE: `make test` runs with -race flag by default.
// Race conditions in tests will be caught automatically.
// Services must NOT import sync or sync/atomic.

func TestDISPLAYNAME_Process(t *testing.T) {
    tests := []struct {
        name      string
        output    interface{}
        err       error
        wantLen   int
        wantError bool
    }{
        {
            name: "successful response with results",
            output: []*AWSSDKPACKAGE.RESULTTYPE{...},
            wantLen: 1,
        },
        {
            name:      "error response",
            err:       errors.New("access denied"),
            wantLen:   1,
            wantError: true,
        },
        {
            name:    "empty results",
            output:  []*AWSSDKPACKAGE.RESULTTYPE{},
            wantLen: 0,
        },
        {
            name:    "nil output no error",
            output:  nil,
            wantLen: 0,
        },
        {
            name:    "wrong type assertion",
            output:  "unexpected",
            wantLen: 0,
        },
    }
    // ... test execution loop
}
```

### Current Template State — What Needs to Change

**`calls.go.tmpl` (122 lines):**
- Currently has TODO comments for placeholder guidance
- Already follows correct Call/Process pattern with no sync imports
- Missing: concurrency-awareness comment header
- Missing: note that Call() is invoked via safeScan wrapper

**`README.md` (93 lines):**
- Has 9 steps (create dir, copy template, replace placeholders, etc.)
- Missing: test template step (contributors need to create tests)
- Missing: concurrent safety section
- Missing: reference to race detection in test step

**`example_calls.go.reference` (116 lines):**
- Annotated CertificateManager implementation
- Has template-to-implementation mapping annotations
- Missing: concurrency-awareness comment header

**`calls_test.go.tmpl` — DOES NOT EXIST:**
- Must be created from scratch
- Must follow existing test patterns (sagemaker, neptune tests)
- Must include race detection note
- Must use same placeholder naming convention (SERVICENAME, DISPLAYNAME, etc.)

### Existing Test Patterns to Follow

From `cmd/awtest/services/sagemaker/calls_test.go`:
- Tests call `SageMakerCalls[0].Process(output, err, false)` directly
- Table-driven with `name`, `output`, `err`, `wantLen`, `wantError` fields
- Assertions check: `ServiceName`, `MethodName`, `ResourceType`, `ResourceName`, `Details`, `Error`
- Uses standard `testing` package (NOT testify — testify is not in go.mod)
- Includes nil field handling tests

### Placeholder Consistency

All template files use these consistent placeholders (from `_template/README.md`):

| Placeholder | Replace With | Example |
|---|---|---|
| `SERVICENAME` | Package name (lowercase) | `certificatemanager` |
| `DISPLAYNAME` | PascalCase service name | `CertificateManager` |
| `SERVICECalls` | Exported var (`DISPLAYNAMECalls`) | `CertificateManagerCalls` |
| `AWSSDKPACKAGE` | AWS SDK v1 service package | `acm` |
| `AWSPREFIX` | AWS API prefix (lowercase) | `acm` |
| `APIMETHOD` | API method name | `ListCertificates` |
| `APIMETHODINPUT` | SDK input struct | `ListCertificatesInput` |
| `RESULTTYPE` | SDK response item type | `CertificateSummary` |
| `RESPONSEFIELD` | Response field name | `CertificateSummaryList` |
| `NAMEFIELD` | Resource name field | `DomainName` |
| `RESOURCETYPE` | Lowercase singular resource type | `certificate` |

The new `calls_test.go.tmpl` must use these same placeholders.

### Previous Story Intelligence

**From Story 10.2 (CONTRIBUTING.md — done):**
- Documentation-only story completed cleanly
- Added concurrent architecture section, race detection docs, validation checklist items
- Key learning: testify is NOT used — standard `testing` package only. Do not reference testify.
- Commit message pattern: `"Update CONTRIBUTING.md with concurrent testing requirements (Story 10.2)"`

**From Story 10.1 (README Update — done):**
- Documentation-only story, straightforward additions to existing content
- Pattern: preserve existing structure, add new sections/subsections

### Git Intelligence

Recent commits:
- `7404c3d Update CONTRIBUTING.md with concurrent testing requirements (Story 10.2)`
- `227d5d6 Update README with 63 services and speed presets (Story 10.1)`

Suggested commit message for this story:
- `"Update service template for concurrent-safe patterns (Story 10.3)"`

### Anti-Patterns to Prevent

The dev agent MUST NOT:
- Add testify imports to the test template (testify is not in go.mod)
- Add sync/sync/atomic imports to calls.go.tmpl
- Modify any files outside `cmd/awtest/services/_template/`
- Remove existing TODO comments from calls.go.tmpl (they guide contributors)
- Change the Call/Process function signatures
- Add a goroutine or channel to the template

### Project Structure Notes

- Template files live in `cmd/awtest/services/_template/`
- `.tmpl` extension prevents Go compilation
- `.reference` extension prevents Go compilation
- New test template must use `.tmpl` extension: `calls_test.go.tmpl`
- All changes are additions/modifications to template files only

### References

- [Source: epics-phase2.md#Story 5.3: Service Template Concurrent-Safe Update] — BDD acceptance criteria
- [Source: prd-phase2.md#FR107] — Service template updated for concurrent-safe implementation patterns
- [Source: architecture-phase2.md#Service Implementation Contract] — Services remain concurrency-unaware, follow Call/Process pattern
- [Source: architecture-phase2.md#Anti-Patterns] — Things contributors must NOT do in service files
- [Source: architecture-phase2.md#Enforcement Guidelines] — Never add sync primitives inside service files
- [Source: cmd/awtest/services/_template/calls.go.tmpl] — Current template (needs concurrency comment)
- [Source: cmd/awtest/services/_template/README.md] — Current template guide (needs concurrent safety section)
- [Source: cmd/awtest/services/_template/example_calls.go.reference] — Annotated reference (needs concurrency comment)
- [Source: cmd/awtest/services/sagemaker/calls_test.go] — Test pattern reference
- [Source: CONTRIBUTING.md#Concurrent Architecture] — Concurrent testing requirements (Story 10.2)
- [Source: 10-2-contributing-md-concurrent-testing-requirements.md] — Previous story learnings

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered. Documentation/template-only story — no compilation or test execution required.

### Completion Notes List

- Task 1: Added 5-line concurrency-awareness comment block to `calls.go.tmpl` after package declaration. Added inline comment on `Call` explaining safeScan wrapper invocation. Verified no sync imports exist.
- Task 2: Created `calls_test.go.tmpl` with table-driven test pattern following sagemaker test conventions. Includes 4 required test cases (success, error, empty, type assertion failure). Race detection note included. Uses consistent placeholder naming (SERVICENAME, DISPLAYNAME, AWSSDKPACKAGE, etc.).
- Task 3: Updated `_template/README.md` — added Step 9 (Create tests) with test template copy instructions, renumbered Run tests to Step 10 with race flag note, added "Concurrent Safety" section documenting anti-patterns.
- Task 4: Added matching concurrency-awareness comment header to `example_calls.go.reference`. All existing annotations preserved.
- Task 5: Verified all templates are concurrent-safe: no sync imports, Call/Process pattern matches safeScan contract, README documents alphabetical registration.

### Change Log

- 2026-03-13: Updated service template files for concurrent-safe patterns (Story 10.3)
  - Added concurrency-awareness comment block to `calls.go.tmpl` and `example_calls.go.reference`
  - Created `calls_test.go.tmpl` test template with race detection note
  - Updated `README.md` with test step and concurrent safety section
- 2026-03-13: Addressed code review findings — 3 items resolved
  - Fixed `calls.go.tmpl` stdout violation: replaced `fmt.Printf` with `utils.HandleAWSError` for type assertion warning
  - Fixed `calls_test.go.tmpl`: replaced local `stringPtr` helper with `aws.String()` from SDK
  - Fixed `calls_test.go.tmpl`: added length guard before `results[0]` access in error path

### File List

- `cmd/awtest/services/_template/calls.go.tmpl` (modified) — Added concurrency-awareness header and safeScan inline comment
- `cmd/awtest/services/_template/calls_test.go.tmpl` (created) — New test template with table-driven pattern
- `cmd/awtest/services/_template/README.md` (modified) — Added test step, concurrent safety section
- `cmd/awtest/services/_template/example_calls.go.reference` (modified) — Added concurrency-awareness header
