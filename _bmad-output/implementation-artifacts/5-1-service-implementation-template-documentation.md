# Story 5.1: Service Implementation Template & Documentation

Status: done

## Story

As a **community contributor wanting to add an AWS service**,
I want **a complete service implementation template with examples**,
so that **I can quickly add new services following established patterns without reading the entire codebase**.

## Acceptance Criteria

1. Create `cmd/awtest/services/_template/` directory with service template file
2. Template includes complete boilerplate: package declaration, AWS SDK imports, types/utils imports, AWSService slice with Name, Call(), Process(), ModuleName
3. Template Call() function includes AWS client creation, context-aware API call (`WithContext`), error handling pattern
4. Template Process() function includes error handling with `utils.HandleAWSError`, result iteration, `utils.PrintResult()`, ScanResult creation with all fields
5. Template includes service registration export: `var SERVICECalls = []types.AWSService{...}`
6. Template includes comprehensive TODO markers for customization (e.g., `// TODO: Replace SERVICENAME`)
7. Template follows naming conventions: PascalCase exports, camelCase unexported, lowercase package name
8. Create `cmd/awtest/services/_template/README.md` explaining template usage step-by-step
9. Include real-world reference: copy a clean existing service (e.g., CertificateManager) alongside template as reference implementation
10. Template compiles after placeholder replacement (verify by creating a sample service from the template)
11. Document template structure in a top-level CONTRIBUTING.md (basic version -- Story 5.2 expands it)
12. FR65 requirement fulfilled: Clear service implementation template provided

## Tasks / Subtasks

- [x] Task 1: Create template directory and boilerplate file (AC: #1, #2, #3, #4, #5, #6, #7)
  - [x] Create `cmd/awtest/services/_template/` directory
  - [x] Create `cmd/awtest/services/_template/calls.go.tmpl` with full boilerplate
  - [x] Package declaration: `package SERVICENAME` with TODO marker
  - [x] Imports block: context, fmt, time, types, utils, aws, session, aws-service-sdk
  - [x] Exported var: `var SERVICECalls = []types.AWSService{...}`
  - [x] Call() function: create SDK client, call API with `WithContext(ctx, ...)`, return output
  - [x] Process() function: error handling with HandleAWSError, iterate results, build ScanResult slice, call PrintResult
  - [x] ModuleName: `types.DefaultModuleName`
  - [x] Add comprehensive inline comments explaining each section
  - [x] Add TODO markers at every customization point

- [x] Task 2: Create template README with step-by-step guide (AC: #8)
  - [x] Create `cmd/awtest/services/_template/README.md`
  - [x] Step 1: Create service directory `cmd/awtest/services/SERVICENAME/`
  - [x] Step 2: Copy `calls.go.tmpl` to `cmd/awtest/services/SERVICENAME/calls.go`
  - [x] Step 3: Replace all placeholders (SERVICENAME, AWSSERVICE, etc.)
  - [x] Step 4: Implement Call() with actual AWS SDK API call
  - [x] Step 5: Implement Process() to extract resources from API response
  - [x] Step 6: Register in `cmd/awtest/services/services.go` AllServices() (alphabetical order)
  - [x] Step 7: Run `go build ./cmd/awtest` to verify compilation
  - [x] Step 8: Run `make test` to verify no regressions
  - [x] Include placeholder reference table mapping template vars to real values

- [x] Task 3: Include reference implementation (AC: #9)
  - [x] Create `cmd/awtest/services/_template/example_calls.go.reference` (copy of certificatemanager/calls.go with added annotations)
  - [x] Add comments explaining how each section maps to the template
  - [x] This file is NOT compiled (use .reference extension to exclude from build)

- [x] Task 4: Verify template compiles after replacement (AC: #10)
  - [x] Create a temporary test service from the template (e.g., `testservice/`)
  - [x] Replace all placeholders with valid values
  - [x] Run `go build ./cmd/awtest` -- must succeed
  - [x] Delete the temporary test service after verification
  - [x] Document the verification in Dev Agent Record

- [x] Task 5: Create basic CONTRIBUTING.md (AC: #11, #12)
  - [x] Create `CONTRIBUTING.md` in repository root
  - [x] Include "Adding a New AWS Service" section with steps referencing the template
  - [x] Include "Service Validation Checklist" section
  - [x] Keep it minimal -- Story 5.2 will expand with full contribution guidelines

## Dev Notes

### Architecture & Constraints

- **Go version:** 1.19 (must match go.mod, Makefile, and GitHub Actions workflow)
- **Module path:** `github.com/MillerMedia/awtest`
- **AWS SDK:** v1 (`github.com/aws/aws-sdk-go`) -- NOT SDK v2
- **Existing services:** 46 services in `cmd/awtest/services/` -- all follow the same AWSService pattern
- **Service registry:** `cmd/awtest/services/services.go` AllServices() function -- new services MUST be added in alphabetical order

### Template Location Decision

The architecture document has a minor inconsistency on template location:
- Architecture Section "Service Implementation Patterns" (line ~850): `cmd/awtest/services/_template/calls.go`
- Architecture Project Structure (line ~1836): `_template/service_calls_template.go` (root)
- Epics file: `_template/service_calls_template.go` (root)

**Decision: Use `cmd/awtest/services/_template/`** because:
1. Co-located with the services it templates
2. Contributors naturally discover it when browsing the services directory
3. The architecture's "Service Implementation Patterns" section (the most detailed reference) uses this path
4. Go's `_` prefix excludes it from the build by convention

### AWSService Interface Pattern (CRITICAL)

Every service MUST implement this exact pattern from `cmd/awtest/types/types.go`:

```go
type AWSService struct {
    Name       string
    Call       func(context.Context, *session.Session) (interface{}, error)
    Process    func(interface{}, error, bool) []ScanResult
    ModuleName string
}
```

**Call() function requirements:**
- Accepts `context.Context` and `*session.Session`
- Creates AWS service client: `svc := awsservice.New(sess)`
- Uses `WithContext(ctx, ...)` variant of API calls (context-aware for timeout support)
- Returns `(interface{}, error)` -- return API output or nil on error

**Process() function requirements:**
- Accepts `(output interface{}, err error, debug bool)`
- Returns `[]types.ScanResult`
- On error: call `utils.HandleAWSError(debug, "service:Method", err)` and return single ScanResult with Error field
- On success: iterate results, create ScanResult for each resource, call `utils.PrintResult()` for console output

### ScanResult Field Mapping

```go
types.ScanResult{
    ServiceName:  "ServiceDisplayName",     // e.g., "CertificateManager", "S3"
    MethodName:   "service:APIMethod",      // e.g., "acm:ListCertificates"
    ResourceType: "resourcetype",           // e.g., "certificate", "bucket"
    ResourceName: resourceNameVar,          // Extracted from API response
    Details:      map[string]interface{}{}, // Optional key-value metadata
    Error:        nil,                      // nil on success, error on failure
    Timestamp:    time.Now(),
}
```

### Service Registration Pattern

In `cmd/awtest/services/services.go`, services are registered via `append` in AllServices():

```go
allServices = append(allServices, newservice.NewServiceCalls...)
```

**MUST maintain alphabetical order** by import path and append statement.

### Template File Extension

Use `.go.tmpl` extension for the template file (NOT `.go`) so:
1. It won't be compiled by the Go toolchain
2. IDE/editors can still provide Go syntax highlighting with proper configuration
3. `go build` won't fail on placeholder syntax

### Reference Implementation: CertificateManager

`cmd/awtest/services/certificatemanager/calls.go` is the ideal reference because:
- Clean, single-API-call service (not complex like S3 with nested calls)
- Multi-region scanning pattern (iterates `types.Regions`)
- Proper error handling with HandleAWSError
- Proper ScanResult construction
- Proper PrintResult usage with ColorizeItem

### Naming Conventions (MUST FOLLOW)

- **Package name:** lowercase, no underscores (e.g., `certificatemanager`, `stepfunctions`)
- **Exported var:** PascalCase + "Calls" suffix (e.g., `CertificateManagerCalls`, `StepFunctionsCalls`)
- **Service Name field:** Human-readable PascalCase (e.g., `"CertificateManager"`)
- **Method Name field:** `"awsprefix:APIMethod"` format (e.g., `"acm:ListCertificates"`)
- **File name:** `calls.go` (always)

### What NOT To Do

- DO NOT use `.go` extension for the template -- it will fail to compile with placeholders
- DO NOT put the template at the repository root -- keep it co-located with services
- DO NOT create a Go code generator -- the architecture specifies "Documentation + Template" approach
- DO NOT modify any existing service files or tests
- DO NOT use AWS SDK v2 -- this project uses AWS SDK v1
- DO NOT add the template service to AllServices() -- it's a template, not a real service

### Testing Approach

- Verify template compiles after placeholder replacement (manual test)
- Run `make test` to confirm no regressions
- Run `go build ./cmd/awtest` to confirm compilation
- No new automated tests needed for documentation files

### Previous Story Intelligence (Story 4.5)

**Learnings from Story 4.5 (last story in Epic 4):**
- Cross-repo token: Default GITHUB_TOKEN failed for Homebrew cask push -- fixed with GH_PAT secret
- macOS Gatekeeper blocks unsigned binaries -- users need `xattr -d com.apple.quarantine`
- `go install` shows version as "dev" (expected -- ldflags only via GoReleaser)
- Go Report Card badge doesn't support `?style=flat-square` parameter
- README updated with banner, badges, full docs, and 46-service list
- Commit pattern: `"Add/Complete [feature] (Story X.Y)"`

### Git Intelligence

Recent commits follow pattern: `"Add [feature] (Story X.Y)"`
- `ab41b6f Complete first release validation (Story 4.5)`
- `451f6bd Update README with full feature docs, banner, and service list (Story 4.5)`
- `e5a943c Add Homebrew tap setup and distribution (Story 4.4)`

### Files to Create

1. `cmd/awtest/services/_template/calls.go.tmpl` -- Service boilerplate template
2. `cmd/awtest/services/_template/README.md` -- Template usage guide
3. `cmd/awtest/services/_template/example_calls.go.reference` -- Annotated reference (CertificateManager)
4. `CONTRIBUTING.md` -- Basic contribution guide (root)

### Files to Modify

- None expected

### Project Structure Notes

- Template directory uses `_` prefix to be excluded from Go build
- All 46 existing services follow the identical AWSService pattern
- Services directory already has `services.go` (registry) and `service_filter.go` (filtering)
- No `_template/` directory exists yet

### References

- [Source: _bmad-output/planning-artifacts/epics.md - Epic 5, Story 5.1, lines 1028-1053]
- [Source: _bmad-output/planning-artifacts/architecture.md - "Service Implementation Patterns" section, lines 831-894]
- [Source: _bmad-output/planning-artifacts/architecture.md - FR64-66 mapping, lines 1711-1717]
- [Source: _bmad-output/planning-artifacts/architecture.md - Project structure, lines 1164-1167, 1835-1844]
- [Source: cmd/awtest/types/types.go - AWSService struct definition, lines 39-44]
- [Source: cmd/awtest/services/services.go - AllServices() registry, lines 53-104]
- [Source: cmd/awtest/services/certificatemanager/calls.go - Reference implementation, lines 1-74]
- [Source: cmd/awtest/utils/output.go - HandleAWSError, PrintResult, ColorizeItem utilities]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

### Completion Notes List

- Task 1: Created `cmd/awtest/services/_template/calls.go.tmpl` with complete AWSService boilerplate including package declaration, imports, Call() with WithContext, Process() with HandleAWSError/PrintResult, ModuleName, and comprehensive TODO markers at every customization point.
- Task 2: Created `cmd/awtest/services/_template/README.md` with 8-step guide and placeholder reference table mapping all template variables to real values.
- Task 3: Created `cmd/awtest/services/_template/example_calls.go.reference` -- annotated copy of CertificateManager service with inline comments mapping each section back to the template placeholders. Uses `.reference` extension to exclude from Go build.
- Task 4: Verified template compiles after placeholder replacement by creating a temporary `testservice/` package using ACM SDK types, running `go build ./cmd/awtest` (success), then deleting the temporary service.
- Task 5: Created `CONTRIBUTING.md` at repo root with "Adding a New AWS Service" section (8 steps referencing the template) and "Service Validation Checklist" section. Kept minimal per story requirements -- Story 5.2 will expand.

### Change Log

- 2026-03-07: Implemented Story 5.1 -- Service Implementation Template & Documentation. Created template directory with boilerplate, README guide, annotated reference implementation, and basic CONTRIBUTING.md.

### File List

- cmd/awtest/services/_template/calls.go.tmpl (new)
- cmd/awtest/services/_template/README.md (new)
- cmd/awtest/services/_template/example_calls.go.reference (new)
- CONTRIBUTING.md (new)
