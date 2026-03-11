# Story 8.2: CodeCommit Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate CodeCommit repositories and branches,
So that I can discover source code repositories accessible with these credentials.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/codecommit/` directory with `calls.go` implementing CodeCommit service enumeration.

2. **AC2:** Implement `codecommit:ListRepositories` API call — iterates all regions in `types.Regions`, creates CodeCommit client per region using config override pattern (`codecommit.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListRepositoriesWithContext` (paginated via NextToken) to collect all repository names, then calls `BatchGetRepositoriesWithContext` for each batch of repository names (max 25 per call) to retrieve full repository metadata. Each repository listed with RepositoryName, Arn, CloneUrlHttp, CloneUrlSsh, DefaultBranch, Description, and Region.

3. **AC3:** Implement `codecommit:ListBranches` API call — iterates all regions, creates CodeCommit client per region using config override, calls `ListRepositoriesWithContext` (paginated) to get repository names, then for each repository calls `ListBranchesWithContext` (paginated via NextToken) to enumerate all branch names. Each result listed with RepositoryName, BranchName, and Region.

4. **AC4:** Both Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

5. **AC5:** Given credentials without CodeCommit access, CodeCommit is skipped silently (access denied handling via existing error classification in safeScan).

6. **AC6:** Register CodeCommit service in `services/services.go` `AllServices()` function in alphabetical order (after `codebuild`, before `codepipeline`).

7. **AC7:** Write table-driven tests in `calls_test.go` covering: valid repos/branches, empty results, access denied errors, nil field handling, type assertion failure handling.

8. **AC8:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

9. **AC9:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/codecommit/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `codecommit:ListRepositories` (AC: 1, 2, 4, 5, 8)
  - [x] Create directory `cmd/awtest/services/codecommit/`
  - [x] Create `calls.go` with `package codecommit`
  - [x] Define `var CodeCommitCalls = []types.AWSService{...}` with 2 entries
  - [x] Implement first entry: Name `"codecommit:ListRepositories"`
  - [x] Call: iterate `types.Regions`, create `codecommit.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), paginate `ListRepositoriesWithContext` via NextToken to collect all repository names. Then batch-get repository details via `BatchGetRepositoriesWithContext` (max 25 names per call). Define local struct `ccRepository` with fields: Name, Arn, CloneUrlHttp, CloneUrlSsh, DefaultBranch, Description, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]ccRepository`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"CodeCommit"`, ResourceType=`"repository"`, ResourceName=repositoryName
  - [x] `utils.PrintResult` format: `"CodeCommit Repository: %s (HTTP: %s, Default Branch: %s, Region: %s)"` with `utils.ColorizeItem(repositoryName)`

- [x] Task 2: Implement `codecommit:ListBranches` (AC: 3, 4, 5, 8)
  - [x] Implement second entry: Name `"codecommit:ListBranches"`
  - [x] Call: iterate regions -> create CodeCommit client with config override -> `ListRepositoriesWithContext` (paginated) -> for each repository, `ListBranchesWithContext` (paginated via NextToken) -> collect all branch names. Define local struct `ccBranch` with fields: RepositoryName, BranchName, Region. Per-region errors: `break` to next region, don't abort scan. Per-repository errors in ListBranches: `continue` to next repository, don't abort region.
  - [x] Process: type-assert `[]ccBranch`, build `ScanResult` with ServiceName=`"CodeCommit"`, ResourceType=`"branch"`, ResourceName=`repositoryName + "/" + branchName`
  - [x] `utils.PrintResult` format: `"CodeCommit Branch: %s/%s (Region: %s)"` with `utils.ColorizeItem(repositoryName)`

- [x] Task 3: Register service in AllServices() (AC: 6)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/codecommit"` to `services/services.go` (alphabetical in imports: after `codebuild`, before `codepipeline`)
  - [x] Add `allServices = append(allServices, codecommit.CodeCommitCalls...)` after `codebuild.CodeBuildCalls...` and before `codepipeline.CodePipelineCalls...`

- [x] Task 4: Write unit tests (AC: 7, 9)
  - [x] Create `cmd/awtest/services/codecommit/calls_test.go`
  - [x] Test `ListRepositories` Process: valid repos with details (name, ARN, clone URLs, default branch, description), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListBranches` Process: valid branches with repo names, empty results, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following CodeBuild/GuardDuty test pattern
  - [x] Access Process via `CodeCommitCalls[0].Process`, `CodeCommitCalls[1].Process`

- [x] Task 5: Build and verify (AC: 9)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/codecommit/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := codecommit.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := codecommit.New(sess)
}
```

### CodeCommit is a REGIONAL Service

CodeCommit is **regional** — repositories exist per-region. Iterate `types.Regions` for both API calls, following the same pattern as CodeBuild, GuardDuty, and Security Hub.

If CodeCommit returns `AccessDeniedException` in a region, handle as a non-fatal error: log with `utils.HandleAWSError` and `break` to next region.

### CodeCommit SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/codecommit`

**API Methods:**

1. **ListRepositories:**
   - `svc.ListRepositoriesWithContext(ctx, &codecommit.ListRepositoriesInput{})` -> `*codecommit.ListRepositoriesOutput`
   - `.Repositories` -> `[]*codecommit.RepositoryNameIdPair` (name and ID only — not full metadata)
   - Each `RepositoryNameIdPair` has:
     - `RepositoryName *string`
     - `RepositoryId *string`
   - Paginated via `NextToken *string`
   - Optional: `SortBy *string` ("repositoryName" or "lastModifiedDate"), `Order *string` ("ascending" or "descending")

2. **BatchGetRepositories:**
   - `svc.BatchGetRepositoriesWithContext(ctx, &codecommit.BatchGetRepositoriesInput{RepositoryNames: names})` -> `*codecommit.BatchGetRepositoriesOutput`
   - `.Repositories` -> `[]*codecommit.RepositoryMetadata`
   - `.RepositoriesNotFound` -> `[]*string` (names that weren't found)
   - Max 25 names per call
   - Each `RepositoryMetadata` has:
     - `RepositoryName *string`
     - `RepositoryId *string`
     - `Arn *string`
     - `CloneUrlHttp *string`
     - `CloneUrlSsh *string`
     - `DefaultBranch *string`
     - `RepositoryDescription *string`
     - `AccountId *string`
     - `CreationDate *time.Time`
     - `LastModifiedDate *time.Time`
     - `KmsKeyId *string`

3. **ListBranches:**
   - `svc.ListBranchesWithContext(ctx, &codecommit.ListBranchesInput{RepositoryName: repoName})` -> `*codecommit.ListBranchesOutput`
   - `.Branches` -> `[]*string` (branch name strings only — no branch objects)
   - Paginated via `NextToken *string`
   - `RepositoryName` is **required**

**No new dependencies needed** — CodeCommit is part of `aws-sdk-go v1.44.266` already in go.mod.

### BatchGetRepositories Batching Pattern

`BatchGetRepositories` accepts max 25 repository names per call (NOT 100 like CodeBuild). Batch accordingly:

```go
// Batch repository names into groups of 25
for i := 0; i < len(repoNames); i += 25 {
    end := i + 25
    if end > len(repoNames) {
        end = len(repoNames)
    }
    batch := repoNames[i:end]

    batchInput := &codecommit.BatchGetRepositoriesInput{
        RepositoryNames: batch,
    }
    batchOutput, err := svc.BatchGetRepositoriesWithContext(ctx, batchInput)
    if err != nil {
        utils.HandleAWSError(false, "codecommit:ListRepositories", err)
        break
    }
    // Process batchOutput.Repositories...
}
```

### ListBranches — Per-Repository Pagination

Unlike CodeBuild's ListBuilds (first page only), ListBranches should be **fully paginated** per repository since repositories may have many branches:

```go
for _, repoName := range repoNames {
    input := &codecommit.ListBranchesInput{
        RepositoryName: repoName,
    }
    for {
        output, err := svc.ListBranchesWithContext(ctx, input)
        if err != nil {
            utils.HandleAWSError(false, "codecommit:ListBranches", err)
            break // skip this repo, continue to next
        }
        for _, branch := range output.Branches {
            // ... collect branch
        }
        if output.NextToken == nil {
            break
        }
        input.NextToken = output.NextToken
    }
}
```

### Local Struct Definitions

```go
type ccRepository struct {
    Name          string
    Arn           string
    CloneUrlHttp  string
    CloneUrlSsh   string
    DefaultBranch string
    Description   string
    Region        string
}

type ccBranch struct {
    RepositoryName string
    BranchName     string
    Region         string
}
```

### Variable & Naming Conventions

- **Package:** `codecommit` (directory: `cmd/awtest/services/codecommit/`)
- **Exported variable:** `CodeCommitCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"codecommit:ListRepositories"`, `"codecommit:ListBranches"`
- **ScanResult.ServiceName:** `"CodeCommit"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"repository"`, `"branch"` (lowercase singular)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)

### Registration Order in services.go

Insert alphabetically — `codecommit` comes after `codebuild`, before `codepipeline`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/codecommit"

// In AllServices():
allServices = append(allServices, codebuild.CodeBuildCalls...)
allServices = append(allServices, codecommit.CodeCommitCalls...)  // NEW
allServices = append(allServices, codepipeline.CodePipelineCalls...)
```

### Testing Pattern

Follow the CodeBuild/GuardDuty test pattern — test Process() functions only with pre-built mock data:

```go
func TestListRepositoriesProcess(t *testing.T) {
    process := CodeCommitCalls[0].Process
    // Table-driven tests with valid repos (name, ARN, clone URLs, default branch, description), empty, errors, nil fields, type assertion failure
}

func TestListBranchesProcess(t *testing.T) {
    process := CodeCommitCalls[1].Process
    // Test with valid branches per repo, empty results, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `codecommit/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `codecommit.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** call `GetRepository` individually per repo — use `BatchGetRepositories` for efficiency (25 per batch)

### Previous Story Intelligence

**From Story 8.1 (CodeBuild — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- Pagination: NextToken loop for paginated APIs
- Batch API pattern: collect names first, then batch-get details (CodeBuild uses 100 per batch, CodeCommit uses 25 per batch)
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `CodeCommitCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "CodeCommit", MethodName: "codecommit:ListRepositories", Error: err, Timestamp: time.Now()}}`
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
- `d6dd093` — Add CodeBuild enumeration with 3 API calls (Story 8.1)
- `71cdff0` — Mark Epic 7 (Critical Security Service Expansion) as done
- `156742d` — Add Security Hub enumeration with 3 API calls (Story 7.4)
- `0de1823` — Add GuardDuty enumeration with 3 API calls (Story 7.3)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message

### FRs Covered

- **FR90:** System enumerates CodeCommit repositories and branches

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/codecommit/
├── calls.go            # CodeCommit service implementation (2 AWSService entries)
└── calls_test.go       # Process() tests for both entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                    # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                   # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/codebuild/calls.go       # Reference implementation (most recent, regional multi-API + batch pattern)
cmd/awtest/services/codebuild/calls_test.go  # Reference test pattern (table-driven Process-only tests)
go.mod                                       # AWS SDK already includes CodeCommit package
```

### References

- [Source: epics-phase2.md#Story 3.2: CodeCommit Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR90] — CodeCommit enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/codebuild/calls.go] — Most recent reference implementation (regional + batch API pattern)
- [Source: cmd/awtest/services/codebuild/calls_test.go] — Reference test pattern (table-driven Process-only tests)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (codecommit goes after codebuild, before codepipeline)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes CodeCommit package)
- [Source: 8-1-codebuild-enumeration.md] — Most recent story (patterns, batch API learnings)
- [Source: 7-2-aws-organizations-enumeration.md] — Code review findings: config override, error logging
- [Source: 7-1-ecr-container-registry-enumeration.md] — Code review findings: pagination, table-driven tests

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Implemented `codecommit:ListRepositories` with full pagination via NextToken, BatchGetRepositories with 25-per-batch limit, config override region pattern, and nil-safe pointer dereferencing on all SDK fields (Name, Arn, CloneUrlHttp, CloneUrlSsh, DefaultBranch, RepositoryDescription).
- Implemented `codecommit:ListBranches` with ListRepositories pagination to discover repos, then per-repo ListBranches pagination. Per-repository errors break to next repo (continue), per-region errors break to next region.
- Both Process functions follow CodeBuild pattern: error handling via HandleAWSError, type assertion with graceful failure, ScanResult construction with full Details map.
- Registered in services.go alphabetically after codebuild, before codepipeline.
- Table-driven tests for both Process functions covering: valid data, empty results, access denied errors, nil/empty fields, type assertion failures.
- No sync primitives imported — concurrency-unaware per NFR57.
- All builds and tests pass: `go build`, `go test`, `go vet`, `go test -race` all clean with zero regressions.

### Change Log

- 2026-03-10: Implemented CodeCommit enumeration with 2 API calls (ListRepositories, ListBranches) — Story 8.2

### File List

- cmd/awtest/services/codecommit/calls.go (NEW)
- cmd/awtest/services/codecommit/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED — added codecommit import and registration)
