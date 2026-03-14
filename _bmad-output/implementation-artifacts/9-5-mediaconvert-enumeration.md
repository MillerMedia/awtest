# Story 9.5: MediaConvert Enumeration

Status: done

<!-- Generated: 2026-03-13 by BMAD Create Story Workflow -->
<!-- Epic: 9 - Infrastructure & Data Service Expansion (Phase 2 Epic 4) -->
<!-- FR: FR112 | Source: epics-phase2.md#Story 4.5 -->
<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate MediaConvert jobs, queues, and presets,
So that I can discover media processing infrastructure and associated S3 input/output locations.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/mediaconvert/` directory with `calls.go` implementing MediaConvert service enumeration with 3 AWSService entries.

2. **AC2:** Implement `mediaconvert:ListQueues` API call — iterates all regions in `types.Regions`, creates MediaConvert client per region using config override pattern (`mediaconvert.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListQueuesWithContext` with `NextToken` pagination and `MaxResults: aws.Int64(20)`. Each queue listed with Name, Arn, Status, Type, PricingPlan, Description, SubmittedJobsCount, ProgressingJobsCount, CreatedAt, and Region.

3. **AC3:** Implement `mediaconvert:ListJobs` API call — iterates all regions in `types.Regions`, creates MediaConvert client per region using config override pattern, calls `ListJobsWithContext` with `NextToken` pagination and `MaxResults: aws.Int64(20)`. Each job listed with Id, Arn, Status, Queue, Role (IAM role ARN), JobTemplate, CreatedAt, CurrentPhase, ErrorMessage, and Region.

4. **AC4:** Implement `mediaconvert:ListPresets` API call — iterates all regions in `types.Regions`, creates MediaConvert client per region using config override pattern, calls `ListPresetsWithContext` with `NextToken` pagination and `MaxResults: aws.Int64(20)`. Each preset listed with Name, Arn, Description, Type, Category, CreatedAt, and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without MediaConvert access, MediaConvert is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register MediaConvert service in `services/services.go` `AllServices()` function in alphabetical order (after `macie2`, before `opensearch`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 3 API calls.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/mediaconvert/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `mediaconvert:ListQueues` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/mediaconvert/`
  - [x] Create `calls.go` with `package mediaconvert`
  - [x] Define `var MediaConvertCalls = []types.AWSService{...}` with 3 entries
  - [x] Define local struct `mcQueue` with fields: Name, Arn, Status, Type, PricingPlan, Description, SubmittedJobsCount, ProgressingJobsCount, CreatedAt, Region
  - [x] Implement `extractQueue` helper — takes `*mediaconvert.Queue` and `region` string, returns `mcQueue` with nil-safe pointer dereferencing. `CreatedAt` is `*time.Time` — format with `time.RFC3339`. `SubmittedJobsCount` and `ProgressingJobsCount` are `*int64` — format with `fmt.Sprintf("%d", *field)`.
  - [x] Implement first entry: Name `"mediaconvert:ListQueues"`
  - [x] Call: iterate `types.Regions`, create `mediaconvert.New(sess, &aws.Config{Region: aws.String(region)})`, call `ListQueuesWithContext` with `NextToken` pagination and `MaxResults: aws.Int64(20)`. Use `output.Queues` to get Queue list. Per-region errors: `break` pagination loop, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]mcQueue`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"MediaConvert"`, ResourceType=`"queue"`, ResourceName=queueName
  - [x] `utils.PrintResult` format: `"MediaConvert Queue: %s (Status: %s, Type: %s, Region: %s)"` with `utils.ColorizeItem(queueName)`

- [x] Task 2: Implement `mediaconvert:ListJobs` (AC: 3, 5, 6, 9)
  - [x] Define local struct `mcJob` with fields: Id, Arn, Status, Queue, Role, JobTemplate, CreatedAt, CurrentPhase, ErrorMessage, Region
  - [x] Implement `extractJob` helper — takes `*mediaconvert.Job` and `region` string, returns `mcJob` with nil-safe pointer dereferencing. `CreatedAt` is `*time.Time` — format with `time.RFC3339`.
  - [x] Implement second entry: Name `"mediaconvert:ListJobs"`
  - [x] Call: iterate regions -> create MediaConvert client with config override -> call `ListJobsWithContext` with `NextToken` pagination and `MaxResults: aws.Int64(20)`. Use `output.Jobs` to get Job list. Per-region errors: `break` pagination loop.
  - [x] Process: type-assert `[]mcJob`, build `ScanResult` with ServiceName=`"MediaConvert"`, ResourceType=`"job"`, ResourceName=jobId
  - [x] `utils.PrintResult` format: `"MediaConvert Job: %s (Status: %s, Queue: %s, Region: %s)"` with `utils.ColorizeItem(jobId)`

- [x] Task 3: Implement `mediaconvert:ListPresets` (AC: 4, 5, 6, 9)
  - [x] Define local struct `mcPreset` with fields: Name, Arn, Description, Type, Category, CreatedAt, Region
  - [x] Implement `extractPreset` helper — takes `*mediaconvert.Preset` and `region` string, returns `mcPreset` with nil-safe pointer dereferencing. `CreatedAt` is `*time.Time` — format with `time.RFC3339`.
  - [x] Implement third entry: Name `"mediaconvert:ListPresets"`
  - [x] Call: iterate regions -> create MediaConvert client with config override -> call `ListPresetsWithContext` with `NextToken` pagination and `MaxResults: aws.Int64(20)`. Use `output.Presets` to get Preset list. Per-region errors: `break` pagination loop.
  - [x] Process: type-assert `[]mcPreset`, build `ScanResult` with ServiceName=`"MediaConvert"`, ResourceType=`"preset"`, ResourceName=presetName
  - [x] `utils.PrintResult` format: `"MediaConvert Preset: %s (Type: %s, Category: %s, Region: %s)"` with `utils.ColorizeItem(presetName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/mediaconvert"` to `services/services.go` (alphabetical in imports: after `macie2`, before `opensearch`)
  - [x] Add `allServices = append(allServices, mediaconvert.MediaConvertCalls...)` after `macie2.Macie2Calls...` and before `opensearch.OpenSearchCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/mediaconvert/calls_test.go`
  - [x] Test `ListQueues` Process: valid queues with details (name, ARN, status, type, pricing plan, description, job counts, created at), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListJobs` Process: valid jobs with details (ID, ARN, status, queue, role, job template, created at, current phase, error message), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListPresets` Process: valid presets with details (name, ARN, description, type, category, created at), empty results, error handling, nil fields, type assertion failure
  - [x] Test extract helpers: `TestExtractQueue`, `TestExtractJob`, `TestExtractPreset` with AWS SDK types (both populated and nil fields)
  - [x] Use table-driven tests with `t.Run` subtests following Kinesis/EMR/CodeDeploy test pattern
  - [x] Access Process via `MediaConvertCalls[0].Process`, `MediaConvertCalls[1].Process`, `MediaConvertCalls[2].Process`

- [x] Task 6: Vendor MediaConvert SDK package (AC: 10)
  - [x] Run `go mod vendor` or manually ensure `vendor/github.com/aws/aws-sdk-go/service/mediaconvert/` is populated
  - [x] MediaConvert package is part of `aws-sdk-go v1.44.266` — already in go.mod, just needs vendoring

- [x] Task 7: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/mediaconvert/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := mediaconvert.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := mediaconvert.New(sess)
}
```

### CRITICAL: DescribeEndpoints is NOT Required

MediaConvert historically required calling `DescribeEndpoints` first to get an account-specific API endpoint URL. **This is no longer required.** AWS deprecated this pattern and now supports standard regional endpoints directly.

Use the standard regional endpoint pattern:
```go
svc := mediaconvert.New(sess, &aws.Config{Region: aws.String(region)})
// Use svc directly — no need to call DescribeEndpoints first
```

**DO NOT** implement the old DescribeEndpoints pattern:
```go
// OLD PATTERN — DO NOT USE:
svc := mediaconvert.New(sess, &aws.Config{Region: aws.String(region)})
epOutput, _ := svc.DescribeEndpoints(...)
svc2 := mediaconvert.New(sess, &aws.Config{Endpoint: aws.String(*epOutput.Endpoints[0].Url)})
```

### MediaConvert Uses `NextToken` Pagination with `MaxResults`

All 3 MediaConvert List APIs use `NextToken` pagination with `MaxResults` (max 20 per page):

```go
var nextToken *string
for {
    input := &mediaconvert.ListQueuesInput{
        MaxResults: aws.Int64(20),
    }
    if nextToken != nil {
        input.NextToken = nextToken
    }
    output, err := svc.ListQueuesWithContext(ctx, input)
    if err != nil {
        lastErr = err
        utils.HandleAWSError(false, "mediaconvert:ListQueues", err)
        break
    }
    for _, queue := range output.Queues {
        if queue != nil {
            allQueues = append(allQueues, extractQueue(queue, region))
        }
    }
    if output.NextToken == nil {
        break
    }
    nextToken = output.NextToken
}
```

### MediaConvert SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/mediaconvert`

**IMPORTANT:** The Go package name is `mediaconvert`. The local package name is also `mediaconvert` — same pattern as `kinesis`/`emr`/`codedeploy` where the local package name matches the AWS SDK package name. Within `calls.go`, `mediaconvert.New()` and `mediaconvert.ListQueuesInput{}` refer to the **AWS SDK package**, while local types (`mcQueue`, `mcJob`, `mcPreset`) are referenced directly without package prefix.

**API Methods:**

1. **ListQueues (Paginated with NextToken, regional):**
   - `svc.ListQueuesWithContext(ctx, &mediaconvert.ListQueuesInput{MaxResults: aws.Int64(20), NextToken: nextToken})` -> `*mediaconvert.ListQueuesOutput`
   - `.Queues` -> `[]*mediaconvert.Queue`
   - Pagination: `NextToken *string` in both input and output
   - Each `Queue` has:
     - `Name *string` — unique per account
     - `Arn *string`
     - `Status *string` — `"ACTIVE"` or `"PAUSED"`
     - `Type_ *string` — `"SYSTEM"` or `"CUSTOM"` (**NOTE: field name is `Type_` not `Type` in Go SDK to avoid keyword conflict**)
     - `PricingPlan *string` — `"ON_DEMAND"` or `"RESERVED"`
     - `Description *string`
     - `SubmittedJobsCount *int64`
     - `ProgressingJobsCount *int64`
     - `CreatedAt *time.Time`

2. **ListJobs (Paginated with NextToken, regional):**
   - `svc.ListJobsWithContext(ctx, &mediaconvert.ListJobsInput{MaxResults: aws.Int64(20), NextToken: nextToken})` -> `*mediaconvert.ListJobsOutput`
   - `.Jobs` -> `[]*mediaconvert.Job`
   - Pagination: `NextToken *string` in both input and output
   - Each `Job` has:
     - `Id *string` — unique job identifier
     - `Arn *string`
     - `Status *string` — `"SUBMITTED"`, `"PROGRESSING"`, `"COMPLETE"`, `"CANCELED"`, `"ERROR"`
     - `Queue *string` — queue ARN or name
     - `Role *string` — **IAM role ARN (security-relevant: reveals roles used for media processing)**
     - `JobTemplate *string` — template used (if any)
     - `CreatedAt *time.Time`
     - `CurrentPhase *string` — `"PROBING"`, `"TRANSCODING"`, `"UPLOADING"`
     - `ErrorMessage *string`
     - `ErrorCode *int64`
     - `Settings *mediaconvert.JobSettings` — full transcode config (**DO NOT extract — too large, not security-relevant**)

3. **ListPresets (Paginated with NextToken, regional):**
   - `svc.ListPresetsWithContext(ctx, &mediaconvert.ListPresetsInput{MaxResults: aws.Int64(20), NextToken: nextToken})` -> `*mediaconvert.ListPresetsOutput`
   - `.Presets` -> `[]*mediaconvert.Preset`
   - Pagination: `NextToken *string` in both input and output
   - Each `Preset` has:
     - `Name *string` — unique per account
     - `Arn *string`
     - `Description *string`
     - `Type_ *string` — `"SYSTEM"` or `"CUSTOM"` (**NOTE: `Type_` not `Type`**)
     - `Category *string`
     - `CreatedAt *time.Time`
     - `LastUpdated *time.Time`
     - `Settings *mediaconvert.PresetSettings` — encoding config (**DO NOT extract**)

### IMPORTANT: `Type_` Field Name

The Go SDK uses `Type_` (with trailing underscore) for the Type field on Queue and Preset structs because `type` is a reserved keyword in Go. Extract with:

```go
queueType := ""
if queue.Type_ != nil {
    queueType = *queue.Type_
}
```

### Simple Paginated Pattern (All 3 Calls)

All 3 MediaConvert API calls follow the same simple paginated pattern — no nested calls required (unlike Kinesis ListShards/ListStreamConsumers which require listing parent resources first).

```go
var allQueues []mcQueue
var lastErr error

for _, region := range types.Regions {
    svc := mediaconvert.New(sess, &aws.Config{Region: aws.String(region)})
    var nextToken *string
    for {
        input := &mediaconvert.ListQueuesInput{
            MaxResults: aws.Int64(20),
        }
        if nextToken != nil {
            input.NextToken = nextToken
        }
        output, err := svc.ListQueuesWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "mediaconvert:ListQueues", err)
            break
        }
        for _, queue := range output.Queues {
            if queue != nil {
                allQueues = append(allQueues, extractQueue(queue, region))
            }
        }
        if output.NextToken == nil {
            break
        }
        nextToken = output.NextToken
    }
}

if len(allQueues) == 0 && lastErr != nil {
    return nil, lastErr
}
return allQueues, nil
```

### Nil-Safe Field Extraction Helpers

```go
func extractQueue(queue *mediaconvert.Queue, region string) mcQueue {
    name := ""
    if queue.Name != nil {
        name = *queue.Name
    }
    arn := ""
    if queue.Arn != nil {
        arn = *queue.Arn
    }
    status := ""
    if queue.Status != nil {
        status = *queue.Status
    }
    queueType := ""
    if queue.Type_ != nil {
        queueType = *queue.Type_
    }
    pricingPlan := ""
    if queue.PricingPlan != nil {
        pricingPlan = *queue.PricingPlan
    }
    description := ""
    if queue.Description != nil {
        description = *queue.Description
    }
    submittedJobs := ""
    if queue.SubmittedJobsCount != nil {
        submittedJobs = fmt.Sprintf("%d", *queue.SubmittedJobsCount)
    }
    progressingJobs := ""
    if queue.ProgressingJobsCount != nil {
        progressingJobs = fmt.Sprintf("%d", *queue.ProgressingJobsCount)
    }
    createdAt := ""
    if queue.CreatedAt != nil {
        createdAt = queue.CreatedAt.Format(time.RFC3339)
    }
    return mcQueue{
        Name:                 name,
        Arn:                  arn,
        Status:               status,
        Type:                 queueType,
        PricingPlan:          pricingPlan,
        Description:          description,
        SubmittedJobsCount:   submittedJobs,
        ProgressingJobsCount: progressingJobs,
        CreatedAt:            createdAt,
        Region:               region,
    }
}

func extractJob(job *mediaconvert.Job, region string) mcJob {
    id := ""
    if job.Id != nil {
        id = *job.Id
    }
    arn := ""
    if job.Arn != nil {
        arn = *job.Arn
    }
    status := ""
    if job.Status != nil {
        status = *job.Status
    }
    queue := ""
    if job.Queue != nil {
        queue = *job.Queue
    }
    role := ""
    if job.Role != nil {
        role = *job.Role
    }
    jobTemplate := ""
    if job.JobTemplate != nil {
        jobTemplate = *job.JobTemplate
    }
    createdAt := ""
    if job.CreatedAt != nil {
        createdAt = job.CreatedAt.Format(time.RFC3339)
    }
    currentPhase := ""
    if job.CurrentPhase != nil {
        currentPhase = *job.CurrentPhase
    }
    errorMessage := ""
    if job.ErrorMessage != nil {
        errorMessage = *job.ErrorMessage
    }
    return mcJob{
        Id:           id,
        Arn:          arn,
        Status:       status,
        Queue:        queue,
        Role:         role,
        JobTemplate:  jobTemplate,
        CreatedAt:    createdAt,
        CurrentPhase: currentPhase,
        ErrorMessage: errorMessage,
        Region:       region,
    }
}

func extractPreset(preset *mediaconvert.Preset, region string) mcPreset {
    name := ""
    if preset.Name != nil {
        name = *preset.Name
    }
    arn := ""
    if preset.Arn != nil {
        arn = *preset.Arn
    }
    description := ""
    if preset.Description != nil {
        description = *preset.Description
    }
    presetType := ""
    if preset.Type_ != nil {
        presetType = *preset.Type_
    }
    category := ""
    if preset.Category != nil {
        category = *preset.Category
    }
    createdAt := ""
    if preset.CreatedAt != nil {
        createdAt = preset.CreatedAt.Format(time.RFC3339)
    }
    return mcPreset{
        Name:        name,
        Arn:         arn,
        Description: description,
        Type:        presetType,
        Category:    category,
        CreatedAt:   createdAt,
        Region:      region,
    }
}
```

### Local Struct Definitions

```go
type mcQueue struct {
    Name                 string
    Arn                  string
    Status               string
    Type                 string
    PricingPlan          string
    Description          string
    SubmittedJobsCount   string
    ProgressingJobsCount string
    CreatedAt            string
    Region               string
}

type mcJob struct {
    Id           string
    Arn          string
    Status       string
    Queue        string
    Role         string
    JobTemplate  string
    CreatedAt    string
    CurrentPhase string
    ErrorMessage string
    Region       string
}

type mcPreset struct {
    Name        string
    Arn         string
    Description string
    Type        string
    Category    string
    CreatedAt   string
    Region      string
}
```

### Variable & Naming Conventions

- **Package:** `mediaconvert` (directory: `cmd/awtest/services/mediaconvert/`)
- **Exported variable:** `MediaConvertCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"mediaconvert:ListQueues"`, `"mediaconvert:ListJobs"`, `"mediaconvert:ListPresets"`
- **ScanResult.ServiceName:** `"MediaConvert"` (title case)
- **ScanResult.ResourceType:** `"queue"`, `"job"`, `"preset"` (lowercase)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **Local struct prefix:** `mc` (abbreviation for mediaconvert, following `cd`/`dc` pattern for multi-word service names)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/mediaconvert"` (same name as local package — handled same as kinesis/emr/codedeploy pattern)

### Registration Order in services.go

Insert alphabetically — `mediaconvert` comes after `macie2`, before `opensearch`:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/macie2"
"github.com/MillerMedia/awtest/cmd/awtest/services/mediaconvert"    // NEW — after macie2, before opensearch
"github.com/MillerMedia/awtest/cmd/awtest/services/opensearch"

// In AllServices():
allServices = append(allServices, macie2.Macie2Calls...)
allServices = append(allServices, mediaconvert.MediaConvertCalls...)  // NEW — after macie2, before opensearch
allServices = append(allServices, opensearch.OpenSearchCalls...)
```

### Testing Pattern

Follow the Kinesis/EMR/CodeDeploy test pattern — test Process() functions only with pre-built mock data:

```go
func TestListQueuesProcess(t *testing.T) {
    process := MediaConvertCalls[0].Process
    // Table-driven tests: valid queues (name, ARN, status, type, pricing plan, description, job counts, created at), empty, errors, nil fields, type assertion failure
}

func TestListJobsProcess(t *testing.T) {
    process := MediaConvertCalls[1].Process
    // Table-driven tests: valid jobs (ID, ARN, status, queue, role, job template, created at, current phase, error message), empty, errors, nil fields, type assertion failure
}

func TestListPresetsProcess(t *testing.T) {
    process := MediaConvertCalls[2].Process
    // Table-driven tests: valid presets (name, ARN, description, type, category, created at), empty, errors, nil fields, type assertion failure
}
```

Include extract helper tests with AWS SDK types:
```go
func TestExtractQueue(t *testing.T) { ... }
func TestExtractJob(t *testing.T) { ... }
func TestExtractPreset(t *testing.T) { ... }
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `mediaconvert/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `mediaconvert.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** call `DescribeEndpoints` — this is deprecated; use the standard regional endpoint directly
- **DO NOT** extract `Settings` from Jobs or Presets — these are massive nested structs with full transcode/encoding config that waste memory and are not security-relevant
- **DO NOT** use `Type` field name in local structs for the SDK's `Type_` field — the SDK uses `Type_` (with underscore) because `type` is a Go keyword. Extract from `queue.Type_` / `preset.Type_`, store in local struct field named `Type`
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor

### Key Differences from Previous Stories (9.4 Kinesis, 9.3 EMR)

1. **Simple flat list pattern:** All 3 MediaConvert APIs are simple flat list operations. No nested calls needed (unlike Kinesis which required listing streams first, then listing shards/consumers per stream). This makes the implementation simpler.
2. **MaxResults required:** MediaConvert APIs cap at 20 results per page — always include `MaxResults: aws.Int64(20)` in the input. Kinesis/EMR don't require MaxResults.
3. **`Type_` field name quirk:** The Go SDK uses `Type_` (with underscore) for the Type field on Queue and Preset structs because `type` is a Go keyword. This is unique to MediaConvert — other services don't have this issue.
4. **No DescribeEndpoints:** MediaConvert historically required a `DescribeEndpoints` call before any other API call. This is deprecated — use regional endpoints directly. No other service had this requirement.
5. **Job IAM Role exposure:** The `Role` field on Job structs reveals IAM role ARNs used for media processing — high security value for pentesters identifying lateral movement opportunities.
6. **Int64 count fields:** Queue has `SubmittedJobsCount` and `ProgressingJobsCount` as `*int64` — format with `fmt.Sprintf("%d", *field)`. Most other services don't have integer count fields.
7. **All regional:** All MediaConvert resources are regional — iterate `types.Regions` for all 3 API calls.

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/mediaconvert/
+-- calls.go            # MediaConvert service implementation (3 AWSService entries)
+-- calls_test.go       # Process() tests + extract helper tests for all 3 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                      # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                     # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/kinesis/calls.go           # Reference implementation (simple pagination, extract helpers)
cmd/awtest/services/kinesis/calls_test.go      # Reference test pattern (extract helper tests)
cmd/awtest/services/codedeploy/calls.go        # Reference implementation (NextToken pagination)
cmd/awtest/services/codedeploy/calls_test.go   # Reference test pattern
go.mod                                         # AWS SDK already includes mediaconvert package (needs vendoring)
```

**Vendor directory to POPULATE:**
```
vendor/github.com/aws/aws-sdk-go/service/mediaconvert/  # Run go mod vendor to populate
```

### Previous Story Intelligence

**From Story 9.4 (Kinesis — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- Extract helper functions for nil-safe extraction — directly applicable
- Time formatting with `time.RFC3339` — same pattern needed for `CreatedAt` fields
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `KinesisCalls[N].Process` in tests -> apply as `MediaConvertCalls[N].Process`
- Error result pattern: `return []types.ScanResult{{ServiceName: "MediaConvert", MethodName: "mediaconvert:ListQueues", Error: err, Timestamp: time.Now()}}`
- 24 tests across 6 test functions in Kinesis

**From Story 9.1 (CodeDeploy — NextToken pagination reference):**
- NextToken pagination pattern — directly applicable to all 3 MediaConvert calls
- 23 tests across 7 test functions in CodeDeploy

**From Code Review Findings (Stories 7.1, 7.2):**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [HIGH] Always add pagination from the start — applies to all 3 MediaConvert calls
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [LOW] Tests should cover nil fields comprehensively
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `f528ebb` — Add EMR and Kinesis enumeration with 3 API calls each (Stories 9.3, 9.4)
- `b7a8967` — Add Direct Connect enumeration with 3 API calls (Story 9.2)
- `7b02834` — Add CodeDeploy enumeration with 3 API calls (Story 9.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add MediaConvert enumeration with 3 API calls (Story 9.5)"`

### FRs Covered

- **FR112:** System enumerates MediaConvert jobs, queues, and presets

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### References

- [Source: epics-phase2.md#Story 4.5: MediaConvert Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR112] — MediaConvert enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/kinesis/calls.go] — Reference implementation (simple pagination, extract helpers, time formatting)
- [Source: cmd/awtest/services/kinesis/calls_test.go] — Reference test pattern (extract helper tests)
- [Source: cmd/awtest/services/codedeploy/calls.go] — Reference implementation (NextToken pagination)
- [Source: cmd/awtest/services/codedeploy/calls_test.go] — Reference test pattern
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (mediaconvert goes after macie2, before opensearch)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes mediaconvert package, needs vendoring)
- [Source: 9-4-kinesis-enumeration.md] — Previous story (simple pagination, extract helpers, time formatting)
- [Source: 9-1-codedeploy-enumeration.md] — Reference story (NextToken pagination)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- SDK field name `Type_` documented in story Dev Notes does not exist in vendored aws-sdk-go v1.44.266. Actual field name is `Type` (no underscore). Fixed during build validation.

### Completion Notes List

- Implemented MediaConvert service enumeration with 3 API calls: ListQueues, ListJobs, ListPresets
- All 3 calls use config override pattern for region iteration, NextToken pagination with MaxResults: 20
- 3 local structs (mcQueue, mcJob, mcPreset) with 3 extract helpers for nil-safe pointer dereferencing
- Registered in AllServices() alphabetically after macie2, before opensearch
- 24 tests across 6 test functions: 3 Process tests + 3 extract helper tests
- No sync primitives used — concurrency-unaware per NFR57
- All WithContext variants used for timeout/cancellation support
- go build, go test, go vet, go test -race all pass clean with zero regressions

### Change Log

- 2026-03-13: Implemented MediaConvert enumeration with 3 API calls (Story 9.5)

### File List

- cmd/awtest/services/mediaconvert/calls.go (NEW)
- cmd/awtest/services/mediaconvert/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED — added mediaconvert import and registration)
- vendor/github.com/aws/aws-sdk-go/service/mediaconvert/ (NEW — vendored SDK package)
