# Story 8.4: SageMaker Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate SageMaker notebook instances, endpoints, models, and training jobs,
So that I can discover ML infrastructure and potential data access through notebooks.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/sagemaker/` directory with `calls.go` implementing SageMaker service enumeration.

2. **AC2:** Implement `sagemaker:ListNotebookInstances` API call — iterates all regions in `types.Regions`, creates SageMaker client per region using config override pattern (`sagemaker.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListNotebookInstancesWithContext` with NextToken pagination (max 100 per page). Each notebook instance listed with NotebookInstanceName, ARN, Status, InstanceType, and Region.

3. **AC3:** Implement `sagemaker:ListEndpoints` API call — iterates all regions, creates SageMaker client per region using config override, calls `ListEndpointsWithContext` with NextToken pagination (max 100 per page). Each endpoint listed with EndpointName, ARN, Status, CreationTime, and Region.

4. **AC4:** Implement `sagemaker:ListModels` API call — iterates all regions, creates SageMaker client per region using config override, calls `ListModelsWithContext` with NextToken pagination (max 100 per page). Each model listed with ModelName, ARN, CreationTime, and Region.

5. **AC5:** Implement `sagemaker:ListTrainingJobs` API call — iterates all regions, creates SageMaker client per region using config override, calls `ListTrainingJobsWithContext` with NextToken pagination (max 100 per page). Each training job listed with TrainingJobName, ARN, Status, CreationTime, and Region.

6. **AC6:** All four Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

7. **AC7:** Given credentials without SageMaker access, SageMaker is skipped silently (access denied handling via existing error classification in safeScan).

8. **AC8:** Register SageMaker service in `services/services.go` `AllServices()` function in alphabetical order (after `rekognition`, before `secretsmanager`).

9. **AC9:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 4 API calls.

10. **AC10:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

11. **AC11:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/sagemaker/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `sagemaker:ListNotebookInstances` (AC: 1, 2, 6, 7, 10)
  - [x] Create directory `cmd/awtest/services/sagemaker/`
  - [x] Create `calls.go` with `package sagemaker`
  - [x] Define `var SageMakerCalls = []types.AWSService{...}` with 4 entries
  - [x] Implement first entry: Name `"sagemaker:ListNotebookInstances"`
  - [x] Call: iterate `types.Regions`, create `sagemaker.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `ListNotebookInstancesWithContext` with NextToken pagination loop (max 100 per page). Define local struct `smNotebook` with fields: Name, Arn, Status, InstanceType, URL, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]smNotebook`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"SageMaker"`, ResourceType=`"notebook-instance"`, ResourceName=notebookName
  - [x] `utils.PrintResult` format: `"SageMaker Notebook: %s (Status: %s, Type: %s, Region: %s)"` with `utils.ColorizeItem(notebookName)`

- [x] Task 2: Implement `sagemaker:ListEndpoints` (AC: 3, 6, 7, 10)
  - [x] Implement second entry: Name `"sagemaker:ListEndpoints"`
  - [x] Call: iterate regions -> create SageMaker client with config override -> `ListEndpointsWithContext` with NextToken pagination (max 100 per page). Define local struct `smEndpoint` with fields: Name, Arn, Status, CreationTime, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: type-assert `[]smEndpoint`, build `ScanResult` with ServiceName=`"SageMaker"`, ResourceType=`"endpoint"`, ResourceName=endpointName
  - [x] `utils.PrintResult` format: `"SageMaker Endpoint: %s (Status: %s, Region: %s)"` with `utils.ColorizeItem(endpointName)`

- [x] Task 3: Implement `sagemaker:ListModels` (AC: 4, 6, 7, 10)
  - [x] Implement third entry: Name `"sagemaker:ListModels"`
  - [x] Call: iterate regions -> create SageMaker client with config override -> `ListModelsWithContext` with NextToken pagination (max 100 per page). Define local struct `smModel` with fields: Name, Arn, CreationTime, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: type-assert `[]smModel`, build `ScanResult` with ServiceName=`"SageMaker"`, ResourceType=`"model"`, ResourceName=modelName
  - [x] `utils.PrintResult` format: `"SageMaker Model: %s (Created: %s, Region: %s)"` with `utils.ColorizeItem(modelName)`

- [x] Task 4: Implement `sagemaker:ListTrainingJobs` (AC: 5, 6, 7, 10)
  - [x] Implement fourth entry: Name `"sagemaker:ListTrainingJobs"`
  - [x] Call: iterate regions -> create SageMaker client with config override -> `ListTrainingJobsWithContext` with NextToken pagination (max 100 per page). Define local struct `smTrainingJob` with fields: Name, Arn, Status, CreationTime, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: type-assert `[]smTrainingJob`, build `ScanResult` with ServiceName=`"SageMaker"`, ResourceType=`"training-job"`, ResourceName=jobName
  - [x] `utils.PrintResult` format: `"SageMaker Training Job: %s (Status: %s, Region: %s)"` with `utils.ColorizeItem(jobName)`

- [x] Task 5: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/sagemaker"` to `services/services.go` (alphabetical in imports: after `s3`, before `secretsmanager`)
  - [x] Add `allServices = append(allServices, sagemaker.SageMakerCalls...)` after `s3.S3Calls...` and before `secretsmanager.SecretsManagerCalls...`

- [x] Task 6: Write unit tests (AC: 9, 11)
  - [x] Create `cmd/awtest/services/sagemaker/calls_test.go`
  - [x] Test `ListNotebookInstances` Process: valid notebooks with details (name, ARN, status, instance type, URL), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListEndpoints` Process: valid endpoints with details (name, ARN, status, creation time), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListModels` Process: valid models (name, ARN, creation time), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListTrainingJobs` Process: valid training jobs (name, ARN, status, creation time), empty results, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following CodeBuild/CodeCommit/OpenSearch test pattern
  - [x] Access Process via `SageMakerCalls[0].Process`, `SageMakerCalls[1].Process`, `SageMakerCalls[2].Process`, `SageMakerCalls[3].Process`

- [x] Task 7: Build and verify (AC: 11)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/sagemaker/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

- [x] Task 8: Review Follow-ups (AI)
  - [x] [AI-Review][Medium] Missing `DefaultCodeRepository` in Notebook Enumeration [cmd/awtest/services/sagemaker/calls.go]
  - [x] [AI-Review][Medium] Missing `LastModifiedTime` in Notebook/Endpoint/TrainingJob Enumeration [cmd/awtest/services/sagemaker/calls.go]

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := sagemaker.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := sagemaker.New(sess)
}
```

### SageMaker is a REGIONAL Service

SageMaker is **regional** — notebooks, endpoints, models, and training jobs exist per-region. Iterate `types.Regions` for all four API calls, following the same pattern as CodeBuild, CodeCommit, OpenSearch.

If SageMaker returns `AccessDeniedException` in a region, handle as a non-fatal error: log with `utils.HandleAWSError` and `break` to next region.

### SageMaker SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/sagemaker`

**IMPORTANT:** The Go package name is `sagemaker` which is the SAME as our local package name. This is the same pattern as CodeBuild — Go resolves this correctly. Within `calls.go`, `sagemaker.New()` and `sagemaker.ListNotebookInstancesInput{}` refer to the **AWS SDK package**, while local types (structs, variables) are referenced directly without package prefix.

**API Methods — All 4 are Paginated:**

1. **ListNotebookInstances:**
   - `svc.ListNotebookInstancesWithContext(ctx, &sagemaker.ListNotebookInstancesInput{MaxResults: aws.Int64(100), NextToken: nextToken})` -> `*sagemaker.ListNotebookInstancesOutput`
   - `.NotebookInstances` -> `[]*sagemaker.NotebookInstanceSummary`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: 1-100
   - Each `NotebookInstanceSummary` has:
     - `NotebookInstanceName *string`
     - `NotebookInstanceArn *string`
     - `NotebookInstanceStatus *string` (InService, Pending, Stopped, Stopping, Failed, Deleting, Updating)
     - `InstanceType *string` (e.g., "ml.t2.medium", "ml.m5.xlarge")
     - `Url *string` (Jupyter notebook URL)
     - `CreationTime *time.Time`
     - `LastModifiedTime *time.Time`
     - `DefaultCodeRepository *string`

2. **ListEndpoints:**
   - `svc.ListEndpointsWithContext(ctx, &sagemaker.ListEndpointsInput{MaxResults: aws.Int64(100), NextToken: nextToken})` -> `*sagemaker.ListEndpointsOutput`
   - `.Endpoints` -> `[]*sagemaker.EndpointSummary`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: 1-100
   - Each `EndpointSummary` has:
     - `EndpointName *string`
     - `EndpointArn *string`
     - `EndpointStatus *string` (InService, Creating, Updating, Failed, Deleting, OutOfService)
     - `CreationTime *time.Time`
     - `LastModifiedTime *time.Time`

3. **ListModels:**
   - `svc.ListModelsWithContext(ctx, &sagemaker.ListModelsInput{MaxResults: aws.Int64(100), NextToken: nextToken})` -> `*sagemaker.ListModelsOutput`
   - `.Models` -> `[]*sagemaker.ModelSummary`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: 1-100
   - Each `ModelSummary` has:
     - `ModelName *string`
     - `ModelArn *string`
     - `CreationTime *time.Time`
     - `LastModifiedTime *time.Time`

4. **ListTrainingJobs:**
   - `svc.ListTrainingJobsWithContext(ctx, &sagemaker.ListTrainingJobsInput{MaxResults: aws.Int64(100), NextToken: nextToken})` -> `*sagemaker.ListTrainingJobsOutput`
   - `.TrainingJobSummaries` -> `[]*sagemaker.TrainingJobSummary`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: 1-100
   - Each `TrainingJobSummary` has:
     - `TrainingJobName *string`
     - `TrainingJobArn *string`
     - `TrainingJobStatus *string` (InProgress, Completed, Failed, Stopping, Stopped)
     - `CreationTime *time.Time`
     - `TrainingEndTime *time.Time` (only for terminal statuses)
     - `LastModifiedTime *time.Time`

**No new dependencies needed** — SageMaker is part of `aws-sdk-go v1.44.266` already in go.mod.

### Pagination Pattern (All 4 Calls)

All 4 SageMaker List APIs use the same NextToken pagination. Follow this exact pattern:

```go
var allNotebooks []smNotebook
for _, region := range types.Regions {
    svc := sagemaker.New(sess, &aws.Config{Region: aws.String(region)})
    var nextToken *string
    for {
        input := &sagemaker.ListNotebookInstancesInput{
            MaxResults: aws.Int64(100),
        }
        if nextToken != nil {
            input.NextToken = nextToken
        }
        output, err := svc.ListNotebookInstancesWithContext(ctx, input)
        if err != nil {
            utils.HandleAWSError(false, "sagemaker:ListNotebookInstances", err)
            break
        }
        for _, nb := range output.NotebookInstances {
            // nil-safe extraction, append to allNotebooks
        }
        if output.NextToken == nil {
            break
        }
        nextToken = output.NextToken
    }
}
```

### Local Struct Definitions

```go
type smNotebook struct {
    Name         string
    Arn          string
    Status       string
    InstanceType string
    URL          string
    Region       string
}

type smEndpoint struct {
    Name         string
    Arn          string
    Status       string
    CreationTime string
    Region       string
}

type smModel struct {
    Name         string
    Arn          string
    CreationTime string
    Region       string
}

type smTrainingJob struct {
    Name         string
    Arn          string
    Status       string
    CreationTime string
    Region       string
}
```

### Variable & Naming Conventions

- **Package:** `sagemaker` (directory: `cmd/awtest/services/sagemaker/`)
- **Exported variable:** `SageMakerCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"sagemaker:ListNotebookInstances"`, `"sagemaker:ListEndpoints"`, `"sagemaker:ListModels"`, `"sagemaker:ListTrainingJobs"`
- **ScanResult.ServiceName:** `"SageMaker"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"notebook-instance"`, `"endpoint"`, `"model"`, `"training-job"` (lowercase hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/sagemaker"` (same name as local package — handled same as codebuild pattern)

### Registration Order in services.go

Insert alphabetically — `sagemaker` comes after `rekognition`, before `secretsmanager`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/sagemaker"

// In AllServices():
allServices = append(allServices, rekognition.RekognitionCalls...)
allServices = append(allServices, sagemaker.SageMakerCalls...)  // NEW
allServices = append(allServices, secretsmanager.SecretsManagerCalls...)
```

### Testing Pattern

Follow the CodeBuild/CodeCommit/OpenSearch test pattern — test Process() functions only with pre-built mock data:

```go
func TestListNotebookInstancesProcess(t *testing.T) {
    process := SageMakerCalls[0].Process
    // Table-driven tests: valid notebooks (name, ARN, status, instance type, URL), empty, errors, nil fields, type assertion failure
}

func TestListEndpointsProcess(t *testing.T) {
    process := SageMakerCalls[1].Process
    // Table-driven tests: valid endpoints (name, ARN, status, creation time), empty, errors, nil fields, type assertion failure
}

func TestListModelsProcess(t *testing.T) {
    process := SageMakerCalls[2].Process
    // Table-driven tests: valid models (name, ARN, creation time), empty, errors, nil fields, type assertion failure
}

func TestListTrainingJobsProcess(t *testing.T) {
    process := SageMakerCalls[3].Process
    // Table-driven tests: valid jobs (name, ARN, status, creation time), empty, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `sagemaker/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `sagemaker.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** use DescribeNotebookInstance (singular) for each notebook — the ListNotebookInstances summary provides sufficient detail (Name, ARN, Status, InstanceType, URL)
- **DO NOT** confuse `sagemaker.ListNotebookInstancesInput` (AWS SDK type) with local package types — AWS SDK `sagemaker` is the imported package, local types are referenced without prefix

### Previous Story Intelligence

**From Story 8.3 (OpenSearch — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- Batch API pattern: collect names first, then batch-describe (OpenSearch uses 5 per batch) — **NOT applicable to SageMaker** since List APIs return sufficient summary detail
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `SageMakerCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "SageMaker", MethodName: "sagemaker:ListNotebookInstances", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests

**From Story 8.2 (CodeCommit):**
- Pagination: NextToken loop for paginated APIs — **directly applicable** to all 4 SageMaker APIs
- Type assertion failure and error handling patterns consistent across all recent stories

**From Story 8.1 (CodeBuild):**
- 3 API calls in one service file — SageMaker extends to 4 calls, same pattern per entry
- Batch-describe pattern (not needed for SageMaker — list summaries are sufficient)

**From Story 7.2 Code Review Findings:**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [LOW] Tests should cover nil fields comprehensively

**From Story 7.1 Code Review Findings:**
- [HIGH] Always add pagination from the start (NextToken loops on paginated APIs) — **critical for SageMaker**, all 4 APIs are paginated
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [MEDIUM] Include ARN in Details map where available
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `d7271c8` — Add OpenSearch enumeration with 3 API calls (Story 8.3)
- `0dd5f6a` — Add CodeCommit enumeration with 2 API calls (Story 8.2)
- `d6dd093` — Add CodeBuild enumeration with 3 API calls (Story 8.1)
- `71cdff0` — Mark Epic 7 (Critical Security Service Expansion) as done
- `156742d` — Add Security Hub enumeration with 3 API calls (Story 7.4)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add SageMaker enumeration with 4 API calls (Story 8.4)"`

### FRs Covered

- **FR92:** System enumerates SageMaker notebook instances, endpoints, models, and training jobs

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/sagemaker/
+-- calls.go            # SageMaker service implementation (4 AWSService entries)
+-- calls_test.go       # Process() tests for all 4 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                    # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                   # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/opensearch/calls.go      # Reference implementation (regional multi-API + batch pattern)
cmd/awtest/services/opensearch/calls_test.go # Reference test pattern (most recent)
cmd/awtest/services/codebuild/calls.go       # Reference implementation (regional multi-API + batch pattern)
cmd/awtest/services/codebuild/calls_test.go  # Reference test pattern (table-driven Process-only tests)
cmd/awtest/services/codecommit/calls.go      # Reference implementation (regional + pagination pattern)
cmd/awtest/services/codecommit/calls_test.go # Reference test pattern
go.mod                                       # AWS SDK already includes sagemaker package
```

### References

- [Source: epics-phase2.md#Story 3.4: SageMaker Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR92] — SageMaker enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/opensearch/calls.go] — Most recent reference implementation (regional + multi-API, 3 calls)
- [Source: cmd/awtest/services/opensearch/calls_test.go] — Most recent reference test pattern
- [Source: cmd/awtest/services/codebuild/calls.go] — Reference implementation (regional + batch API, 3 calls)
- [Source: cmd/awtest/services/codecommit/calls.go] — Reference implementation (regional + pagination pattern, 2 calls)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (sagemaker goes after rekognition, before secretsmanager)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes sagemaker package)
- [Source: 8-3-opensearch-enumeration.md] — Most recent story (patterns, regional iteration)
- [Source: 8-2-codecommit-enumeration.md] — Previous story (pagination learnings)
- [Source: 8-1-codebuild-enumeration.md] — Previous story (3 API calls pattern, batch describe)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered. All implementations followed reference patterns from Stories 8.1-8.3.

### Completion Notes List

- Implemented 4 SageMaker API calls: ListNotebookInstances, ListEndpoints, ListModels, ListTrainingJobs
- All 4 calls use config override pattern for region iteration (per 7.2 code review finding)
- All 4 calls implement NextToken pagination with MaxResults=100 per page
- All Process functions handle errors via utils.HandleAWSError, type-assert output, nil-safe pointer dereferencing
- No sync primitives used — concurrency-unaware per NFR57
- Registered in services.go alphabetically (after s3, before secretsmanager)
- 20 test cases across 4 test functions covering valid results, empty results, access denied, nil fields, type assertion failure
- go build, go test, go vet, go test -race all pass clean with zero regressions
- Addressed code review findings (2 Medium):
  - Added DefaultCodeRepository field to smNotebook struct, extraction, Details map, and tests
  - Added LastModifiedTime field to smNotebook, smEndpoint, smTrainingJob structs, extraction, Details maps, and tests
- Low issues (redundant error logging, verbose pointer dereferencing) intentionally kept for consistency with all other services (opensearch, codecommit, codebuild, etc.)

### Change Log

- 2026-03-12: Implemented SageMaker enumeration with 4 API calls (ListNotebookInstances, ListEndpoints, ListModels, ListTrainingJobs)
- 2026-03-12: Addressed code review findings — added DefaultCodeRepository and LastModifiedTime fields

### File List

- cmd/awtest/services/sagemaker/calls.go (created)
- cmd/awtest/services/sagemaker/calls_test.go (created)
- cmd/awtest/services/services.go (modified — added sagemaker import and registration)
