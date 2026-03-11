# Story 8.1: CodeBuild Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate CodeBuild projects and build configurations,
So that I can discover build environment variables containing secrets and access build history.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/codebuild/` directory with `calls.go` implementing CodeBuild service enumeration.

2. **AC2:** Implement `codebuild:ListProjects` API call — iterates all regions in `types.Regions`, creates CodeBuild client per region using config override pattern (`codebuild.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListProjectsWithContext` (paginated via NextToken), then calls `BatchGetProjectsWithContext` for each batch of project names (max 100 per call) to retrieve full project details. Each project listed with Name, ARN, Description, Source Type, Environment Type, and Region.

3. **AC3:** Implement `codebuild:ListProjectEnvironmentVariables` API call — iterates all regions, creates CodeBuild client per region using config override, calls `ListProjectsWithContext` (paginated), then `BatchGetProjectsWithContext` to retrieve project details. For each project, extracts environment variable key names and their Type (PLAINTEXT/PARAMETER_STORE/SECRETS_MANAGER). **MUST NOT** include environment variable values in output — only key names and types. Each result listed with ProjectName, VariableName, VariableType, and Region.

4. **AC4:** Implement `codebuild:ListBuilds` API call — iterates all regions, creates CodeBuild client per region using config override, calls `ListBuildsWithContext` (first page only, no pagination — cap at 100 build IDs per region), then `BatchGetBuildsWithContext` to hydrate build details. Each build listed with BuildId, ProjectName, BuildStatus, StartTime, and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without CodeBuild access, CodeBuild is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register CodeBuild service in `services/services.go` `AllServices()` function in alphabetical order (after `cloudwatch`, before `codepipeline`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid projects/env vars/builds, empty results, access denied errors, nil field handling, type assertion failure handling.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/codebuild/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `codebuild:ListProjects` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/codebuild/`
  - [x] Create `calls.go` with `package codebuild`
  - [x] Define `var CodeBuildCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"codebuild:ListProjects"`
  - [x] Call: iterate `types.Regions`, create `codebuild.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), paginate `ListProjectsWithContext` via NextToken to collect all project names. Then batch-get project details via `BatchGetProjectsWithContext` (max 100 names per call). Define local struct `cbProject` with fields: Name, Arn, Description, SourceType, EnvironmentType, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]cbProject`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"CodeBuild"`, ResourceType=`"project"`, ResourceName=projectName
  - [x] `utils.PrintResult` format: `"CodeBuild Project: %s (Source: %s, Env: %s, Region: %s)"` with `utils.ColorizeItem(projectName)`

- [x] Task 2: Implement `codebuild:ListProjectEnvironmentVariables` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"codebuild:ListProjectEnvironmentVariables"`
  - [x] Call: iterate regions -> create CodeBuild client with config override -> `ListProjectsWithContext` (paginated) -> `BatchGetProjectsWithContext` -> extract environment variables from each project. Define local struct `cbEnvVar` with fields: ProjectName, VariableName, VariableType, Region. **CRITICAL:** Do NOT include the Value field — only Name and Type.
  - [x] Per-region errors: `break` to next region, don't abort scan
  - [x] Process: type-assert `[]cbEnvVar`, build `ScanResult` with ServiceName=`"CodeBuild"`, ResourceType=`"environment-variable"`, ResourceName=`projectName + "/" + variableName`
  - [x] `utils.PrintResult` format: `"CodeBuild Env Var: %s/%s (Type: %s)"` with `utils.ColorizeItem(projectName)`

- [x] Task 3: Implement `codebuild:ListBuilds` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"codebuild:ListBuilds"`
  - [x] Call: iterate regions -> create CodeBuild client with config override -> `ListBuildsWithContext` (first page only, no pagination — cap at 100 build IDs per region) -> `BatchGetBuildsWithContext` to hydrate build details (max 100 IDs per call)
  - [x] Define local struct `cbBuild` with fields: BuildId, ProjectName, BuildStatus, StartTime, Region
  - [x] Per-region errors: `break` to next region, don't abort scan
  - [x] Process: type-assert `[]cbBuild`, build `ScanResult` with ServiceName=`"CodeBuild"`, ResourceType=`"build"`, ResourceName=buildId
  - [x] `utils.PrintResult` format: `"CodeBuild Build: %s (Project: %s, Status: %s, Region: %s)"` with `utils.ColorizeItem(buildId)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/codebuild"` to `services/services.go` (alphabetical in imports: after `cloudwatch`, before `codepipeline`)
  - [x] Add `allServices = append(allServices, codebuild.CodeBuildCalls...)` after `cloudwatch.CloudwatchCalls...` and before `codepipeline.CodePipelineCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/codebuild/calls_test.go`
  - [x] Test `ListProjects` Process: valid projects with details, empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListProjectEnvironmentVariables` Process: valid env vars with types (PLAINTEXT/PARAMETER_STORE/SECRETS_MANAGER), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListBuilds` Process: valid builds with status, empty results, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following GuardDuty test pattern
  - [x] Access Process via `CodeBuildCalls[0].Process`, `CodeBuildCalls[1].Process`, `CodeBuildCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/codebuild/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := codebuild.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := codebuild.New(sess)
}
```

### CodeBuild is a REGIONAL Service

CodeBuild is **regional** — projects and builds exist per-region. Iterate `types.Regions` for all 3 API calls, following the same pattern as GuardDuty and Security Hub.

If CodeBuild returns `AccessDeniedException` in a region, handle as a non-fatal error: log with `utils.HandleAWSError` and `break` to next region.

### CodeBuild SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/codebuild`

**API Methods:**

1. **ListProjects:**
   - `svc.ListProjectsWithContext(ctx, &codebuild.ListProjectsInput{})` -> `*codebuild.ListProjectsOutput`
   - `.Projects` -> `[]*string` (project names only — not full objects)
   - Paginated via `NextToken *string`
   - Sort: `SortBy *string` ("NAME" or "CREATED_TIME"), `SortOrder *string` ("ASCENDING" or "DESCENDING")

2. **BatchGetProjects:**
   - `svc.BatchGetProjectsWithContext(ctx, &codebuild.BatchGetProjectsInput{Names: names})` -> `*codebuild.BatchGetProjectsOutput`
   - `.Projects` -> `[]*codebuild.Project`
   - Max 100 names per call
   - Each `Project` has:
     - `Name *string`
     - `Arn *string`
     - `Description *string`
     - `Source *ProjectSource` with `Type *string` (CODECOMMIT/CODEPIPELINE/GITHUB/GITHUB_ENTERPRISE/BITBUCKET/S3/NO_SOURCE)
     - `Environment *ProjectEnvironment` with:
       - `Type *string` (LINUX_CONTAINER/LINUX_GPU_CONTAINER/ARM_CONTAINER/WINDOWS_SERVER_2019_CONTAINER)
       - `Image *string`
       - `ComputeType *string`
       - `EnvironmentVariables []*EnvironmentVariable`
     - `Created *time.Time`
     - `LastModified *time.Time`

3. **EnvironmentVariable:**
   - `Name *string` — variable key name
   - `Type *string` — PLAINTEXT, PARAMETER_STORE, or SECRETS_MANAGER
   - `Value *string` — **DO NOT include in output** (may contain secrets)

4. **ListBuilds:**
   - `svc.ListBuildsWithContext(ctx, &codebuild.ListBuildsInput{})` -> `*codebuild.ListBuildsOutput`
   - `.Ids` -> `[]*string` (build IDs only)
   - Paginated via `NextToken *string`
   - Returns build IDs in reverse chronological order (newest first)

5. **BatchGetBuilds:**
   - `svc.BatchGetBuildsWithContext(ctx, &codebuild.BatchGetBuildsInput{Ids: ids})` -> `*codebuild.BatchGetBuildsOutput`
   - `.Builds` -> `[]*codebuild.Build`
   - Max 100 IDs per call
   - Each `Build` has:
     - `Id *string` (format: "project:build-id")
     - `Arn *string`
     - `ProjectName *string`
     - `BuildStatus *string` (SUCCEEDED/FAILED/FAULT/TIMED_OUT/IN_PROGRESS/STOPPED)
     - `StartTime *time.Time`
     - `EndTime *time.Time`
     - `CurrentPhase *string`

**No new dependencies needed** — CodeBuild is part of `aws-sdk-go v1.44.266` already in go.mod.

### SECURITY: Environment Variable Values

**CRITICAL:** The `codebuild:ListProjectEnvironmentVariables` call MUST NOT include the `Value` field of environment variables in any output (ScanResult Details, PrintResult display, or any logging). Only enumerate:
- `Name` (the key name)
- `Type` (PLAINTEXT, PARAMETER_STORE, or SECRETS_MANAGER)

Variables of type `PARAMETER_STORE` or `SECRETS_MANAGER` contain reference strings (not actual secrets), but `PLAINTEXT` variables may contain actual secrets. To maintain the security posture described in FR89 ("key names, not secret values"), exclude all values.

### BatchGetProjects Batching Pattern

`BatchGetProjects` accepts max 100 project names per call. If a region has >100 projects, batch them:

```go
// Batch project names into groups of 100
for i := 0; i < len(projectNames); i += 100 {
    end := i + 100
    if end > len(projectNames) {
        end = len(projectNames)
    }
    batch := projectNames[i:end]

    batchInput := &codebuild.BatchGetProjectsInput{
        Names: batch,
    }
    batchOutput, err := svc.BatchGetProjectsWithContext(ctx, batchInput)
    if err != nil {
        utils.HandleAWSError(false, "codebuild:ListProjects", err)
        break
    }
    // Process batchOutput.Projects...
}
```

Same pattern applies to `BatchGetBuilds` with build IDs.

### Local Struct Definitions

```go
type cbProject struct {
    Name            string
    Arn             string
    Description     string
    SourceType      string
    EnvironmentType string
    Region          string
}

type cbEnvVar struct {
    ProjectName  string
    VariableName string
    VariableType string
    Region       string
}

type cbBuild struct {
    BuildId     string
    ProjectName string
    BuildStatus string
    StartTime   string
    Region      string
}
```

### ListBuilds — First Page Only

For `codebuild:ListBuilds`, only retrieve the first page of build IDs (up to 100 per region). This prevents overwhelming output while showing the most recent builds. Do NOT paginate — this is intentional to cap results:

```go
listBuildsOutput, err := svc.ListBuildsWithContext(ctx, &codebuild.ListBuildsInput{})
if err != nil {
    lastErr = err
    utils.HandleAWSError(false, "codebuild:ListBuilds", err)
    break
}
// No NextToken pagination — first page only
```

### Variable & Naming Conventions

- **Package:** `codebuild` (directory: `cmd/awtest/services/codebuild/`)
- **Exported variable:** `CodeBuildCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"codebuild:ListProjects"`, `"codebuild:ListProjectEnvironmentVariables"`, `"codebuild:ListBuilds"`
- **ScanResult.ServiceName:** `"CodeBuild"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"project"`, `"environment-variable"`, `"build"` (lowercase singular or hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)

### Registration Order in services.go

Insert alphabetically — `codebuild` comes after `cloudwatch`, before `codepipeline`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/codebuild"

// In AllServices():
allServices = append(allServices, cloudwatch.CloudwatchCalls...)
allServices = append(allServices, codebuild.CodeBuildCalls...)  // NEW
allServices = append(allServices, codepipeline.CodePipelineCalls...)
```

### Testing Pattern

Follow the GuardDuty/SecurityHub test pattern — test Process() functions only with pre-built mock data:

```go
func TestListProjectsProcess(t *testing.T) {
    process := CodeBuildCalls[0].Process
    // Table-driven tests with valid projects, empty, errors, nil fields, type assertion failure
}

func TestListProjectEnvironmentVariablesProcess(t *testing.T) {
    process := CodeBuildCalls[1].Process
    // Test with valid env vars (PLAINTEXT/PARAMETER_STORE/SECRETS_MANAGER types), empty, errors, nil fields, type assertion failure
}

func TestListBuildsProcess(t *testing.T) {
    process := CodeBuildCalls[2].Process
    // Test with valid builds with status, empty builds, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `codebuild/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `codebuild.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** include environment variable Values in output — only key names and types (FR89 security requirement)
- **DO NOT** paginate ListBuilds — limit to first 100 per region (first page only)
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor

### Previous Story Intelligence

**From Story 7.4 (Security Hub — most recent completed story):**
- All 3 Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for multi-step call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- Pagination: NextToken loop for paginated APIs (ListProjects, ListBuilds first page only)
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `CodeBuildCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "CodeBuild", MethodName: "codebuild:ListProjects", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests

**From Story 7.2 Code Review Findings:**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [LOW] Tests should cover nil fields comprehensively

**From Story 7.1 Code Review Findings:**
- [HIGH] Always add pagination from the start (NextToken loops on all paginated APIs)
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [MEDIUM] Include ARN in Details map where available
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `71cdff0` — Mark Epic 7 (Critical Security Service Expansion) as done
- `156742d` — Add Security Hub enumeration with 3 API calls (Story 7.4)
- `0de1823` — Add GuardDuty enumeration with 3 API calls (Story 7.3)
- `2023f44` — Add AWS Organizations enumeration with 3 API calls (Story 7.2)
- `2c0b4ab` — Add ECR container registry enumeration with 3 API calls (Story 7.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message

### FRs Covered

- **FR89:** System enumerates CodeBuild projects and build configurations

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/codebuild/
├── calls.go            # CodeBuild service implementation (3 AWSService entries)
└── calls_test.go       # Process() tests for all 3 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                    # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                   # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/guardduty/calls.go       # Reference implementation (regional multi-API pattern)
cmd/awtest/services/guardduty/calls_test.go  # Reference test pattern (table-driven Process-only tests)
cmd/awtest/services/securityhub/calls.go     # Reference implementation (most recent, same patterns)
go.mod                                       # AWS SDK already includes CodeBuild package
```

### References

- [Source: epics-phase2.md#Story 3.1: CodeBuild Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR89] — CodeBuild enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/guardduty/calls.go] — Reference implementation (regional config override pattern)
- [Source: cmd/awtest/services/guardduty/calls_test.go] — Reference test pattern (table-driven Process-only tests)
- [Source: cmd/awtest/services/securityhub/calls.go] — Most recent reference implementation
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (codebuild goes after cloudwatch, before codepipeline)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes CodeBuild package)
- [Source: 7-4-security-hub-enumeration.md] — Most recent story (patterns, all learnings)
- [Source: 7-2-aws-organizations-enumeration.md] — Code review findings: config override, error logging
- [Source: 7-1-ecr-container-registry-enumeration.md] — Code review findings: pagination, table-driven tests

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no debugging required.

### Completion Notes List

- Implemented all 3 CodeBuild API calls following SecurityHub/GuardDuty reference patterns
- `codebuild:ListProjects` — regional iteration, paginated ListProjects + BatchGetProjects (batched 100), nil-safe field extraction
- `codebuild:ListProjectEnvironmentVariables` — regional iteration, paginated ListProjects + BatchGetProjects, extracts env var Name and Type only (Value excluded per FR89 security requirement)
- `codebuild:ListBuilds` — regional iteration, first page only (no pagination), BatchGetBuilds for hydration, StartTime formatted as RFC3339
- All Call functions use config override pattern (`codebuild.New(sess, &aws.Config{Region: ...})`) per Story 7.2 code review fix
- All Process functions handle errors via `utils.HandleAWSError`, type-assert output, nil-safe field access
- No sync primitives imported — concurrency-unaware per NFR57
- Registered in `services.go` alphabetically (after cloudwatch, before codepipeline)
- Table-driven tests for all 3 Process functions: valid data, empty results, access denied, nil fields, type assertion failure
- All env var types tested: PLAINTEXT, PARAMETER_STORE, SECRETS_MANAGER
- `go build`, `go test`, `go vet`, `go test -race` all pass clean with zero regressions

### Change Log

- 2026-03-10: Implemented CodeBuild enumeration with 3 API calls (Story 8.1)

### File List

- `cmd/awtest/services/codebuild/calls.go` (new) — CodeBuild service implementation with 3 AWSService entries
- `cmd/awtest/services/codebuild/calls_test.go` (new) — Table-driven Process() tests for all 3 entries
- `cmd/awtest/services/services.go` (modified) — Added codebuild import and registration in AllServices()

## Senior Developer Review (AI)

- [x] Story file loaded from `_bmad-output/implementation-artifacts/8-1-codebuild-enumeration.md`
- [x] Story Status verified as reviewable (review)
- [x] Epic and Story IDs resolved (8.1)
- [x] Story Context located or warning recorded
- [x] Epic Tech Spec located or warning recorded
- [x] Architecture/standards docs loaded (as available)
- [x] Tech stack detected and documented
- [x] MCP doc search performed (or web fallback) and references captured
- [x] Acceptance Criteria cross-checked against implementation
- [x] File List reviewed and validated for completeness
- [x] Tests identified and mapped to ACs; gaps noted
- [x] Code quality review performed on changed files
- [x] Security review performed on changed files and dependencies
- [x] Outcome decided (Approve)
- [x] Review notes appended under "Senior Developer Review (AI)"
- [x] Change Log updated with review entry
- [x] Status updated according to settings (if enabled)
- [x] Sprint status synced (if sprint tracking enabled)
- [x] Story saved successfully

_Reviewer: Kn0ck0ut on 2026-03-10_

### Review Findings

**🔥 CODE REVIEW FINDINGS, Kn0ck0ut!**

**Story:** 8-1-codebuild-enumeration.md
**Git vs Story Discrepancies:** 0 found
**Issues Found:** 0 High, 0 Medium, 0 Low

## 🟢 PASSED
- All Acceptance Criteria implemented correctly.
- Security requirement to exclude environment variable values is met.
- Pagination and batching logic is correct.
- Concurrency safety (no shared state/sync primitives) is observed.
- Tests are comprehensive and passing.
