# Story 5.2: CONTRIBUTING.md Guide for Service Addition

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **community contributor wanting to contribute**,
I want **comprehensive contribution guidelines**,
so that **I can add AWS services, submit quality pull requests, and understand project standards without maintainer guidance**.

## Acceptance Criteria

1. Expand existing `CONTRIBUTING.md` (created in Story 5.1) with full contribution guidelines
2. Include expanded "Adding a New AWS Service" section with step-by-step guide (10 steps from epics):
   - Copy template to `cmd/awtest/services/SERVICENAME/calls.go`
   - Replace placeholders
   - Implement Call() and Process()
   - Register in `services.go` AllServices() alphabetically
   - Write table-driven tests in `calls_test.go`
   - Run `make test` and `go build ./cmd/awtest`
   - Test manually with `go run ./cmd/awtest -debug`
   - Submit PR with service name in title
3. Include "Code Standards" section documenting:
   - Naming conventions (PascalCase exports, camelCase unexported, lowercase package names, no underscores)
   - Error handling patterns (`utils.HandleAWSError`)
   - Test requirements (table-driven tests with testify)
   - Documentation requirements (inline comments for exported functions)
4. Include "Development Workflow" section:
   - Prerequisites: Go 1.19+, make, golangci-lint, GoReleaser
   - Setup: `go mod download`
   - Build: `make build`
   - Test: `make test`
   - Lint: `make lint`
5. Include "Pull Request Process" section:
   - PR title format: "Add [Service Name] enumeration"
   - PR description template with checklist
   - Review process expectations
   - Testing requirements before merge
6. Expand "Service Validation Checklist" with comprehensive items:
   - AWSService interface pattern compliance
   - Package naming (lowercase, no underscores)
   - Error handling uses `utils.HandleAWSError`
   - Tests pass, compilation succeeds
   - Service registered in AllServices() alphabetically
   - Manual testing completed
7. Include "Release Process" section (for maintainers):
   - Tag creation: `git tag v0.x.y`
   - Push tag: `git push origin v0.x.y`
   - GitHub Actions handles release automation via GoReleaser
8. Verify CONTRIBUTING.md accuracy by following it to add a test service (same verification approach as Story 5.1)
9. FR64 and FR66 requirements fulfilled: Documented patterns, validation checklist for consistency

## Tasks / Subtasks

- [x] Task 1: Expand "Adding a New AWS Service" section (AC: #2)
  - [x] Expand the existing 9-step guide to 10 steps per epics spec
  - [x] Add "Write table-driven tests" step (step 6 in expanded guide)
  - [x] Add "Test manually with debug flag" step
  - [x] Add "Submit PR with service name in title" step
  - [x] Reference template at `cmd/awtest/services/_template/`
  - [x] Reference `example_calls.go.reference` for annotated real-world example

- [x] Task 2: Add "Code Standards" section (AC: #3)
  - [x] Document naming conventions:
    - Package names: lowercase, no underscores (e.g., `certificatemanager`)
    - Exported vars: PascalCase + "Calls" suffix (e.g., `CertificateManagerCalls`)
    - File names: lowercase with underscores for multi-word (e.g., `calls_test.go`)
    - Type names: PascalCase exported, camelCase unexported
    - No `SCREAMING_SNAKE_CASE` constants
  - [x] Document error handling pattern:
    - `utils.HandleAWSError(debug, "prefix:Method", err)` for AWS errors
    - Return single `ScanResult` with Error field on failure
    - No `panic()` for expected errors
  - [x] Document testing standards:
    - Table-driven tests with `t.Run(tt.name, ...)`
    - Testify assertions and mocking (`github.com/stretchr/testify v1.9.0`)
    - Co-locate tests (`*_test.go` in same package)
    - Same package name for test files (not `package X_test`)
  - [x] Document inline comment requirements for exported functions

- [x] Task 3: Add "Development Workflow" section (AC: #4)
  - [x] Prerequisites: Go 1.19+, make
  - [x] Setup: `go mod download`
  - [x] Build: `make build` or `go build ./cmd/awtest`
  - [x] Test: `make test`
  - [x] Lint: `make lint` (if available in Makefile)
  - [x] Debug run: `go run ./cmd/awtest --debug`

- [x] Task 4: Add "Pull Request Process" section (AC: #5)
  - [x] PR title format: "Add [Service Name] enumeration"
  - [x] PR description template/checklist
  - [x] Testing requirements before merge
  - [x] Link back to Service Validation Checklist

- [x] Task 5: Expand "Service Validation Checklist" (AC: #6)
  - [x] Merge existing checklist items with comprehensive architecture checklist
  - [x] Add: AWSService interface pattern compliance
  - [x] Add: Call() uses WithContext variant
  - [x] Add: Process() returns proper ScanResult entries
  - [x] Add: Service Name field uses correct format
  - [x] Add: Documentation (doc comments on exported symbols)

- [x] Task 6: Add "Release Process" section (AC: #7)
  - [x] Tag creation: `git tag v0.x.y`
  - [x] Push tag: `git push origin v0.x.y`
  - [x] GitHub Actions + GoReleaser handles automated release
  - [x] Homebrew cask auto-updates via `homebrew-tap` repo
  - [x] Note: Requires `GH_PAT` secret for cross-repo push (learned from Story 4.4)

- [x] Task 7: Verify accuracy (AC: #8)
  - [x] Follow the expanded CONTRIBUTING.md to create a temporary test service
  - [x] Verify all steps work as documented
  - [x] Delete temporary test service after verification
  - [x] Document verification in Dev Agent Record

## Dev Notes

### Architecture & Constraints

- **Go version:** 1.19 (must match go.mod, Makefile, and GitHub Actions workflow)
- **Module path:** `github.com/MillerMedia/awtest`
- **AWS SDK:** v1 (`github.com/aws/aws-sdk-go`) -- NOT SDK v2
- **Test framework:** testify v1.9.0
- **Existing services:** 46 services in `cmd/awtest/services/`
- **Service registry:** `cmd/awtest/services/services.go` AllServices() function

### What Already Exists (from Story 5.1)

The current `CONTRIBUTING.md` has:
- Basic "Adding a New AWS Service" (9 steps referencing template)
- Basic "Service Validation Checklist" (9 items)
- Basic "Development Requirements" (Go 1.19+, AWS SDK v1)

**This story EXPANDS the existing file -- do NOT rewrite from scratch.** Preserve the existing content structure and add the new sections.

### Template Files (Already Created in Story 5.1)

- `cmd/awtest/services/_template/calls.go.tmpl` -- Service boilerplate template
- `cmd/awtest/services/_template/README.md` -- Detailed template usage guide with placeholder reference table
- `cmd/awtest/services/_template/example_calls.go.reference` -- Annotated CertificateManager as reference

### AWSService Interface Pattern (CRITICAL -- must be documented accurately)

```go
type AWSService struct {
    Name       string
    Call       func(context.Context, *session.Session) (interface{}, error)
    Process    func(interface{}, error, bool) []ScanResult
    ModuleName string
}
```

**Call():** Accepts `context.Context` and `*session.Session`, creates AWS client, uses `WithContext` API variant, returns `(interface{}, error)`

**Process():** Accepts `(output interface{}, err error, debug bool)`, returns `[]types.ScanResult`. On error: `utils.HandleAWSError` + single error ScanResult. On success: iterate, build ScanResults, call `utils.PrintResult()`.

### ScanResult Fields (must document for contributors)

```go
types.ScanResult{
    ServiceName:  "CertificateManager",       // PascalCase display name
    MethodName:   "acm:ListCertificates",      // awsprefix:APIMethod
    ResourceType: "certificate",               // lowercase singular
    ResourceName: resourceNameVar,             // from API response
    Details:      map[string]interface{}{},     // optional metadata
    Error:        nil,                          // nil on success
    Timestamp:    time.Now(),
}
```

### Service Registration Pattern (must document accurately)

In `cmd/awtest/services/services.go`:
- Import: `"github.com/MillerMedia/awtest/cmd/awtest/services/SERVICENAME"`
- Append: `allServices = append(allServices, SERVICENAME.DISPLAYNAMECalls...)`
- **MUST maintain alphabetical order** (except STS first for credential validation)
- Always append full slice (`...`), never individual items

### Naming Conventions Summary (from architecture doc)

| Element | Convention | Example |
|---|---|---|
| Package name | lowercase, no underscores | `certificatemanager` |
| Exported var | PascalCase + "Calls" | `CertificateManagerCalls` |
| Service Name field | PascalCase | `"CertificateManager"` |
| Method Name field | `prefix:APIMethod` | `"acm:ListCertificates"` |
| File name | `calls.go` | always `calls.go` |
| Test file | `calls_test.go` | co-located in same package |
| Type names | PascalCase exported | `ScanResult` |
| Variables | camelCase | `allResults` |

### Error Handling Pattern (from architecture doc)

```go
if err != nil {
    utils.HandleAWSError(debug, "service:Method", err)
    return []types.ScanResult{{
        ServiceName: "ServiceName",
        MethodName:  "service:Method",
        Error:       err,
        Timestamp:   time.Now(),
    }}
}
```

HandleAWSError classifies:
- Invalid credentials -> InvalidKeyError (abort scan)
- Access denied -> prints message, returns nil (continue)
- Other errors -> pretty-prints error details

### Testing Pattern (from architecture doc)

```go
func TestServiceCall(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected []types.ScanResult
    }{
        {name: "success case", ...},
        {name: "error case", ...},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

- Use testify for assertions: `assert.Equal`, `assert.NoError`
- Use testify mocks: `mock.Mock`, `mock.Anything`, `mockClient.AssertExpectations(t)`

### Makefile Targets (verify these exist before documenting)

Check `Makefile` for available targets. Known from Story 4.2:
- `make build` -- build binary
- `make test` -- run tests
- `make lint` -- run linter
- `make clean` -- cleanup

### Release Workflow (from Story 4.3-4.4)

- GitHub Actions workflow at `.github/workflows/release.yml`
- Triggered by tag push (`v*`)
- Uses GoReleaser for cross-platform builds
- Homebrew cask auto-updates via push to `MillerMedia/homebrew-tap`
- Requires `GH_PAT` secret (not default `GITHUB_TOKEN`) for cross-repo push

### What NOT To Do

- DO NOT rewrite CONTRIBUTING.md from scratch -- expand the existing file
- DO NOT modify any source code files (only documentation)
- DO NOT add features or change behavior
- DO NOT use AWS SDK v2 references -- this project uses SDK v1
- DO NOT document features that don't exist yet (e.g., Phase 2 concurrent scanning)
- DO NOT add a CHANGELOG.md unless explicitly requested

### Previous Story Intelligence (Story 5.1)

**Key learnings:**
- Template location: `cmd/awtest/services/_template/` (not repo root)
- Template uses `.go.tmpl` extension (not `.go`) to avoid compilation
- Reference implementation uses `.reference` extension
- Story 5.1 created the basic CONTRIBUTING.md that this story expands
- Verification approach: create temp service, build, test, delete
- Commit pattern: `"Add/Complete [feature] (Story X.Y)"`

### Git Intelligence

Recent commits:
- `2ffd5f8 Update README with complete API call list (46 services, 77 calls)`
- `0bfdbab Add service implementation template and CONTRIBUTING.md (Story 5.1)`
- `ab41b6f Complete first release validation (Story 4.5)`

### Files to Modify

1. `CONTRIBUTING.md` -- Expand with full contribution guidelines

### Files to Create

- None expected (expanding existing file only)

### Project Structure Notes

- All 46 services follow identical AWSService pattern in `cmd/awtest/services/`
- Template directory at `cmd/awtest/services/_template/` with 3 files
- Makefile at repo root with build/test/lint targets
- GitHub Actions workflows at `.github/workflows/`
- GoReleaser config at `.goreleaser.yaml`

### References

- [Source: _bmad-output/planning-artifacts/epics.md - Epic 5, Story 5.2]
- [Source: _bmad-output/planning-artifacts/architecture.md - "Service Implementation Patterns" section, lines 831-894]
- [Source: _bmad-output/planning-artifacts/architecture.md - Naming Conventions, lines 976-1139]
- [Source: _bmad-output/planning-artifacts/architecture.md - Testing Strategy, lines 717-829]
- [Source: _bmad-output/planning-artifacts/architecture.md - Error Handling, lines 1244-1292]
- [Source: _bmad-output/planning-artifacts/architecture.md - FR64-66 mapping, lines 1711-1717]
- [Source: _bmad-output/planning-artifacts/architecture.md - Enforcement Guidelines, lines 1589-1615]
- [Source: _bmad-output/implementation-artifacts/5-1-service-implementation-template-documentation.md - Previous story context]
- [Source: CONTRIBUTING.md - Current basic version from Story 5.1]
- [Source: cmd/awtest/services/_template/README.md - Template usage guide]
- [Source: cmd/awtest/types/types.go - AWSService struct definition]
- [Source: cmd/awtest/services/services.go - AllServices() registry]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None

### Completion Notes List

- Expanded CONTRIBUTING.md from basic 39-line guide to comprehensive contribution documentation
- Task 1: Expanded "Adding a New AWS Service" from 9 steps to 10 steps, adding write tests, manual debug testing, and PR submission steps
- Task 2: Added "Code Standards" section with naming conventions table, error handling patterns, testing standards, and documentation requirements
- Task 3: Added "Development Workflow" section with prerequisites, setup, and common make commands table
- Task 4: Added "Pull Request Process" section with title format, PR description template with checklist, and review process expectations
- Task 5: Expanded "Service Validation Checklist" from 9 items to 16 items, adding AWSService interface compliance, Call/Process signature checks, ScanResult field requirements, and manual testing
- Task 6: Added "Release Process" section for maintainers with tag creation, push, and GitHub Actions automation notes including GH_PAT requirement
- Task 7: Verification completed successfully - created temporary SNS-based test service (tmpverify), registered in services.go, built and tested successfully, then cleaned up. All documented steps work as described.

### Change Log

- 2026-03-07: Expanded CONTRIBUTING.md with comprehensive contribution guidelines (Story 5.2)
- 2026-03-07: Addressed code review findings - 4 items resolved (1 High, 1 Medium, 2 Low)

### File List

- `CONTRIBUTING.md` (modified) -- Expanded with full contribution guidelines
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified) -- Status updated: ready-for-dev -> in-progress -> review
- `_bmad-output/implementation-artifacts/5-2-contributing-md-guide-for-service-addition.md` (modified) -- Story file updated with task completion and Dev Agent Record
