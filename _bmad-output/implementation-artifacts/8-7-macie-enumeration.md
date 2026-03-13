# Story 8.7: Macie Enumeration

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate Macie findings, classification jobs, and monitored S3 buckets,
So that I can identify where sensitive data has been detected across S3 buckets and understand data classification coverage.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/macie2/` directory with `calls.go` implementing Macie service enumeration with 3 AWSService entries.

2. **AC2:** Implement `macie2:ListClassificationJobs` API call — iterates all regions in `types.Regions`, creates Macie2 client per region using config override pattern (`macie2.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListClassificationJobsWithContext` with NextToken pagination (max 25 per page). Each job listed with JobId, Name, JobType (ONE_TIME/SCHEDULED), JobStatus (RUNNING/PAUSED/CANCELLED/COMPLETE/IDLE/USER_PAUSED), CreatedAt, and Region.

3. **AC3:** Implement `macie2:ListFindings` API call — iterates all regions, creates Macie2 client per region using config override, calls `ListFindingsWithContext` with NextToken pagination (max 50 per page) to collect FindingIds, then calls `GetFindingsWithContext` (max 50 IDs per batch) to retrieve full details. Each finding listed with Id, Type, Title, Severity (description string), Category (CLASSIFICATION/POLICY), Count, CreatedAt, and Region.

4. **AC4:** Implement `macie2:DescribeBuckets` API call — iterates all regions, creates Macie2 client per region using config override, calls `DescribeBucketsWithContext` with NextToken pagination (max 50 per page). Each monitored bucket listed with BucketName, AccountId, BucketArn, ObjectCount, SizeInBytes, ClassifiableObjectCount, SensitivityScore, and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without Macie access or Macie not enabled in a region, Macie is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register Macie service in `services/services.go` `AllServices()` function in alphabetical order (after `lambda`, before `opensearch`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 3 API calls.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/macie2/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `macie2:ListClassificationJobs` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/macie2/`
  - [x] Create `calls.go` with `package macie2`
  - [x] Define `var Macie2Calls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"macie2:ListClassificationJobs"`
  - [x] Call: iterate `types.Regions`, create `macie2.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `ListClassificationJobsWithContext` with NextToken pagination loop (max 25 per page). Define local struct `mcClassificationJob` with fields: JobId, Name, JobType, JobStatus, CreatedAt, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]mcClassificationJob`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"Macie"`, ResourceType=`"classification-job"`, ResourceName=jobName (or JobId if Name is empty)
  - [x] `utils.PrintResult` format: `"Macie Classification Job: %s (Type: %s, Status: %s, Region: %s)"` with `utils.ColorizeItem(jobName)`

- [x] Task 2: Implement `macie2:ListFindings` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"macie2:ListFindings"`
  - [x] Call: iterate regions -> create Macie2 client with config override -> call `ListFindingsWithContext` with NextToken pagination (max 50 per page), collecting all FindingIds -> batch IDs into groups of 50 -> call `GetFindingsWithContext` for each batch -> extract Finding details. Define local struct `mcFinding` with fields: Id, Type, Title, Severity, Category, Count, CreatedAt, Region. Per-region errors: `break` to next region.
  - [x] Implement `batchGetFindings` helper function (following athena's batch-get pattern with single retry for failed IDs)
  - [x] Implement `extractFinding` helper function for nil-safe field extraction from `*macie2.Finding`
  - [x] Process: type-assert `[]mcFinding`, build `ScanResult` with ServiceName=`"Macie"`, ResourceType=`"finding"`, ResourceName=findingId
  - [x] `utils.PrintResult` format: `"Macie Finding: %s (Type: %s, Severity: %s, Category: %s, Region: %s)"` with `utils.ColorizeItem(findingId)`

- [x] Task 3: Implement `macie2:DescribeBuckets` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"macie2:DescribeBuckets"`
  - [x] Call: iterate regions -> create Macie2 client with config override -> call `DescribeBucketsWithContext` with NextToken pagination (max 50 per page). Define local struct `mcMonitoredBucket` with fields: BucketName, AccountId, BucketArn, ObjectCount, SizeInBytes, ClassifiableObjectCount, SensitivityScore, Region. Per-region errors: `break` to next region.
  - [x] Process: type-assert `[]mcMonitoredBucket`, build `ScanResult` with ServiceName=`"Macie"`, ResourceType=`"monitored-bucket"`, ResourceName=bucketName
  - [x] `utils.PrintResult` format: `"Macie Monitored Bucket: %s (Objects: %s, Classifiable: %s, Sensitivity: %s, Region: %s)"` with `utils.ColorizeItem(bucketName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/macie2"` to `services/services.go` (alphabetical in imports: after `lambda`, before `opensearch`)
  - [x] Add `allServices = append(allServices, macie2.Macie2Calls...)` after `lambda.LambdaCalls...` and before `opensearch.OpenSearchCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/macie2/calls_test.go`
  - [x] Test `ListClassificationJobs` Process: valid jobs with details (job ID, name, type, status, creation time), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListFindings` Process: valid findings with details (ID, type, title, severity, category, count, creation time), empty results, error handling, nil fields, type assertion failure
  - [x] Test `DescribeBuckets` Process: valid buckets with details (bucket name, account ID, ARN, object count, size, classifiable count, sensitivity score), empty results, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following Athena/Backup test pattern
  - [x] Access Process via `Macie2Calls[0].Process`, `Macie2Calls[1].Process`, `Macie2Calls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/macie2/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := macie2.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := macie2.New(sess)
}
```

### Macie is a REGIONAL Service

AWS Macie is **regional** — classification jobs, findings, and monitored buckets exist per-region. Iterate `types.Regions` for all three API calls, following the same pattern as Athena, Backup, SageMaker, etc.

**IMPORTANT: Macie may not be enabled in many regions.** Unlike most AWS services, Macie must be explicitly enabled per-region before it can be used. If Macie is not enabled in a region, API calls will return errors (typically `AccessDeniedException` or `Macie is not enabled` error). Handle as a non-fatal error: log with `utils.HandleAWSError` and `break` to next region. This is the same pattern used for GuardDuty, Security Hub, and other optional services.

### Macie2 SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/macie2`

**IMPORTANT:** The Go package name is `macie2` (NOT `macie` — the original Amazon Macie Classic was deprecated). The local package name is also `macie2`, same pattern as athena/backup/sagemaker/codebuild where the local package name matches the AWS SDK package name. Within `calls.go`, `macie2.New()` and `macie2.ListClassificationJobsInput{}` refer to the **AWS SDK package**, while local types (structs, variables) are referenced directly without package prefix.

**API Methods:**

1. **ListClassificationJobs (Paginated, returns full objects):**
   - `svc.ListClassificationJobsWithContext(ctx, &macie2.ListClassificationJobsInput{MaxResults: aws.Int64(25), NextToken: nextToken})` -> `*macie2.ListClassificationJobsOutput`
   - `.Items` -> `[]*macie2.JobSummary`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: max 25
   - Each `JobSummary` has:
     - `JobId *string`
     - `Name *string`
     - `JobType *string` ("ONE_TIME" or "SCHEDULED")
     - `JobStatus *string` ("RUNNING", "PAUSED", "CANCELLED", "COMPLETE", "IDLE", "USER_PAUSED")
     - `CreatedAt *time.Time`
     - `BucketCriteria *macie2.S3BucketCriteriaForJob` (optional, can be nil)
     - `LastRunErrorStatus *macie2.LastRunErrorStatus` (optional)

2. **ListFindings (Paginated, returns IDs only):**
   - `svc.ListFindingsWithContext(ctx, &macie2.ListFindingsInput{MaxResults: aws.Int64(50), NextToken: nextToken})` -> `*macie2.ListFindingsOutput`
   - `.FindingIds` -> `[]*string` (IDs only — must batch-get details)
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: max 50

3. **GetFindings (Non-paginated, batch by IDs):**
   - `svc.GetFindingsWithContext(ctx, &macie2.GetFindingsInput{FindingIds: idBatch})` -> `*macie2.GetFindingsOutput`
   - `.Findings` -> `[]*macie2.Finding`
   - **Max 50 IDs per call** — batch input IDs into groups of 50
   - Each `Finding` has:
     - `Id *string`
     - `Type *string` (e.g., "SensitiveData:S3Object/Multiple", "Policy:IAMUser/S3BucketPublic")
     - `Title *string`
     - `Description *string` (optional)
     - `Severity *macie2.Severity`:
       - `Description *string` ("High", "Medium", "Low")
       - `Score *int64`
     - `Category *string` ("CLASSIFICATION" or "POLICY")
     - `Count *int64`
     - `CreatedAt *time.Time`
     - `Region *string`
     - `ResourcesAffected *macie2.ResourcesAffected` (optional, contains S3Bucket/S3Object info)

4. **DescribeBuckets (Paginated, returns full objects):**
   - `svc.DescribeBucketsWithContext(ctx, &macie2.DescribeBucketsInput{MaxResults: aws.Int64(50), NextToken: nextToken})` -> `*macie2.DescribeBucketsOutput`
   - `.Buckets` -> `[]*macie2.BucketMetadata`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: max 50
   - Each `BucketMetadata` has:
     - `BucketName *string`
     - `AccountId *string`
     - `Region *string`
     - `BucketArn *string`
     - `ObjectCount *int64`
     - `SizeInBytes *int64`
     - `ClassifiableObjectCount *int64`
     - `SensitivityScore *int64`

**No new dependencies needed** — Macie2 is part of `aws-sdk-go v1.44.266` already in go.mod.

### Pagination Pattern (Calls 1, 3)

Calls 1 and 3 use NextToken pagination with full objects returned. **MaxResults is 25 for ListClassificationJobs and 50 for DescribeBuckets**. Follow this exact pattern:

```go
var allJobs []mcClassificationJob
for _, region := range types.Regions {
    svc := macie2.New(sess, &aws.Config{Region: aws.String(region)})
    var nextToken *string
    for {
        input := &macie2.ListClassificationJobsInput{
            MaxResults: aws.Int64(25),
        }
        if nextToken != nil {
            input.NextToken = nextToken
        }
        output, err := svc.ListClassificationJobsWithContext(ctx, input)
        if err != nil {
            utils.HandleAWSError(false, "macie2:ListClassificationJobs", err)
            break
        }
        for _, job := range output.Items {
            // nil-safe extraction, append to allJobs
        }
        if output.NextToken == nil {
            break
        }
        nextToken = output.NextToken
    }
}
```

### ListFindings + GetFindings Batch Pattern (Call 2)

Call 2 uses the same List-IDs-then-Batch-Get pattern as Athena's ListNamedQueries. ListFindings returns only finding IDs. You must call GetFindings in batches of 50 to get full details. Use streaming batch processing — batch-get per pagination page rather than accumulating all IDs first:

```go
// Per region:
var nextToken *string
for {
    input := &macie2.ListFindingsInput{
        MaxResults: aws.Int64(50),
    }
    if nextToken != nil {
        input.NextToken = nextToken
    }
    output, err := svc.ListFindingsWithContext(ctx, input)
    if err != nil {
        utils.HandleAWSError(false, "macie2:ListFindings", err)
        break
    }

    // Batch-get this page's IDs immediately (max 50 per page = max 50 per batch)
    if len(output.FindingIds) > 0 {
        allFindings = append(allFindings,
            batchGetFindings(ctx, svc, output.FindingIds, region)...)
    }

    if output.NextToken == nil {
        break
    }
    nextToken = output.NextToken
}
```

### Batch-Get Findings Helper (with single retry)

Follow Athena's `batchGetNamedQueries` pattern — note that `GetFindings` does NOT have `UnprocessedFindingIds` like Athena's BatchGet APIs. It either succeeds or fails for the whole batch:

```go
func batchGetFindings(ctx context.Context, svc *macie2.Macie2, ids []*string, region string) []mcFinding {
    var results []mcFinding

    batchOutput, err := svc.GetFindingsWithContext(ctx, &macie2.GetFindingsInput{
        FindingIds: ids,
    })
    if err != nil {
        utils.HandleAWSError(false, "macie2:GetFindings", err)
        return results
    }

    for _, f := range batchOutput.Findings {
        results = append(results, extractFinding(f, region))
    }

    return results
}
```

### Nested Struct Field Extraction

Macie2 has nested structs that require careful nil-safe extraction:

```go
// Severity (nested in Finding)
severity := ""
if f.Severity != nil && f.Severity.Description != nil {
    severity = *f.Severity.Description
}

// Count (int64 pointer)
count := ""
if f.Count != nil {
    count = fmt.Sprintf("%d", *f.Count)
}

// CreatedAt (time.Time pointer)
createdAt := ""
if f.CreatedAt != nil {
    createdAt = f.CreatedAt.Format(time.RFC3339)
}

// ObjectCount/SizeInBytes (int64 pointers in BucketMetadata)
objectCount := ""
if b.ObjectCount != nil {
    objectCount = fmt.Sprintf("%d", *b.ObjectCount)
}
```

### Local Struct Definitions

```go
type mcClassificationJob struct {
    JobId     string
    Name      string
    JobType   string
    JobStatus string
    CreatedAt string
    Region    string
}

type mcFinding struct {
    Id        string
    Type      string
    Title     string
    Severity  string
    Category  string
    Count     string
    CreatedAt string
    Region    string
}

type mcMonitoredBucket struct {
    BucketName              string
    AccountId               string
    BucketArn               string
    ObjectCount             string
    SizeInBytes             string
    ClassifiableObjectCount string
    SensitivityScore        string
    Region                  string
}
```

### Variable & Naming Conventions

- **Package:** `macie2` (directory: `cmd/awtest/services/macie2/`)
- **Exported variable:** `Macie2Calls` (`[]types.AWSService`)
- **AWSService.Name values:** `"macie2:ListClassificationJobs"`, `"macie2:ListFindings"`, `"macie2:DescribeBuckets"`
- **ScanResult.ServiceName:** `"Macie"` (PascalCase, human-readable — NOT "Macie2")
- **ScanResult.ResourceType:** `"classification-job"`, `"finding"`, `"monitored-bucket"` (lowercase hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **Local struct prefix:** `mc` (for Macie, following `at` for Athena, `bk` for Backup, `sm` for SageMaker pattern)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/macie2"` (same name as local package — handled same as athena/backup/sagemaker/codebuild pattern)

### Registration Order in services.go

Insert alphabetically — `macie2` comes after `lambda`, before `opensearch`:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/lambda"
"github.com/MillerMedia/awtest/cmd/awtest/services/macie2"      // NEW — after lambda, before opensearch
"github.com/MillerMedia/awtest/cmd/awtest/services/opensearch"

// In AllServices():
allServices = append(allServices, lambda.LambdaCalls...)
allServices = append(allServices, macie2.Macie2Calls...)     // NEW — after lambda, before opensearch
allServices = append(allServices, opensearch.OpenSearchCalls...)
```

### Testing Pattern

Follow the Athena/Backup test pattern — test Process() functions only with pre-built mock data:

```go
func TestListClassificationJobsProcess(t *testing.T) {
    process := Macie2Calls[0].Process
    // Table-driven tests: valid jobs (ID, name, type, status, created time), empty, errors, nil fields, type assertion failure
}

func TestListFindingsProcess(t *testing.T) {
    process := Macie2Calls[1].Process
    // Table-driven tests: valid findings (ID, type, title, severity, category, count, created time), empty, errors, nil fields, type assertion failure
}

func TestDescribeBucketsProcess(t *testing.T) {
    process := Macie2Calls[2].Process
    // Table-driven tests: valid buckets (bucket name, account ID, ARN, object count, size, classifiable count, sensitivity score), empty, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `macie2/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `macie2.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** pass more than 50 IDs to GetFindings — batch into groups of 50
- **DO NOT** use the deprecated `macie` package — must use `macie2` (Amazon Macie Classic was deprecated in favor of Macie2)
- **DO NOT** confuse `macie2.ListClassificationJobsInput` (AWS SDK type) with local package types — AWS SDK `macie2` is the imported package, local types are referenced without prefix
- **DO NOT** use MaxResults > 25 for ListClassificationJobs or > 50 for other Macie2 APIs

### Previous Story Intelligence

**From Story 8.6 (Athena — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- NextToken pagination: exact pattern with `if nextToken != nil { input.NextToken = nextToken }` before call
- Batch-get pattern: `batchGetNamedQueries` and `batchGetQueryExecutions` helpers with single retry for unprocessed IDs — **directly applicable** to `batchGetFindings` (though GetFindings may not have unprocessed ID tracking)
- Streaming batch processing: IDs batch-fetched per pagination page (max 50) instead of accumulating all IDs first — same pattern for ListFindings
- `extractNamedQuery` / `extractQueryExecution` helper functions for nil-safe extraction — apply same pattern as `extractFinding`
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `AthenaCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "Macie", MethodName: "macie2:ListClassificationJobs", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests
- `truncateRuneSafe()` helper for string truncation — not needed for Macie (no long query strings)
- 22 tests across 4 test functions

**From Story 8.5 (AWS Backup):**
- 4 AWSService entries per file — Macie has 3 entries (simpler)
- Dependent call pattern: list parent resources first (vaults), then query per resource — Macie doesn't need this (ListFindings is flat, not per-resource)
- Error handling with `isResourceNotFound` — not needed for Macie

**From Story 7.2 Code Review Findings:**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [LOW] Tests should cover nil fields comprehensively

**From Story 7.1 Code Review Findings:**
- [HIGH] Always add pagination from the start (NextToken loops on paginated APIs)
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `60147ae` — Add Athena enumeration with 3 API calls (Story 8.6)
- `2bdd8e2` — Add SageMaker enumeration with 4 API calls (Story 8.4)
- `d7271c8` — Add OpenSearch enumeration with 3 API calls (Story 8.3)
- `0dd5f6a` — Add CodeCommit enumeration with 2 API calls (Story 8.2)
- `d6dd093` — Add CodeBuild enumeration with 3 API calls (Story 8.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add Macie enumeration with 3 API calls (Story 8.7)"`

### FRs Covered

- **FR95:** System enumerates Macie findings, classification jobs, and sensitive data discovery results

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/macie2/
+-- calls.go            # Macie service implementation (3 AWSService entries)
+-- calls_test.go       # Process() tests for all 3 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                    # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                   # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/athena/calls.go          # Reference implementation (regional + batch-get + 3 APIs, most recent)
cmd/awtest/services/athena/calls_test.go     # Reference test pattern (most recent)
cmd/awtest/services/backup/calls.go          # Reference implementation (regional + 4 APIs)
cmd/awtest/services/guardduty/calls.go       # Reference implementation (regional, optional service like Macie)
go.mod                                       # AWS SDK already includes macie2 package
```

### References

- [Source: epics-phase2.md#Story 3.7: Macie Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR95] — Macie enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/athena/calls.go] — Most recent reference implementation (regional + batch-get, 3 APIs)
- [Source: cmd/awtest/services/athena/calls_test.go] — Most recent reference test pattern
- [Source: cmd/awtest/services/backup/calls.go] — Reference implementation (regional + 4 APIs)
- [Source: cmd/awtest/services/guardduty/calls.go] — Reference for optional service handling (GuardDuty must be enabled)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (macie2 goes after lambda, before opensearch)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes macie2 package)
- [Source: 8-6-athena-enumeration.md] — Most recent story (batch-get pattern, 3 APIs)
- [Source: 8-5-aws-backup-enumeration.md] — Previous story (regional iteration, 4 APIs)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no debugging required.

### Completion Notes List

- Implemented 3 Macie2 AWSService entries: ListClassificationJobs, ListFindings, DescribeBuckets
- All 3 Call functions iterate types.Regions with config override pattern (safe for concurrent execution)
- ListClassificationJobs uses NextToken pagination (max 25 per page), extracts JobId, Name, JobType, JobStatus, CreatedAt
- ListFindings uses List+BatchGet pattern: ListFindingsWithContext (max 50 IDs per page) → GetFindingsWithContext (max 50 per batch) with batchGetFindings helper and extractFinding for nil-safe field extraction
- DescribeBuckets uses NextToken pagination (max 50 per page), extracts BucketName, AccountId, BucketArn, ObjectCount, SizeInBytes, ClassifiableObjectCount, SensitivityScore
- All Process functions handle errors, type assertion failures, and nil-safe fields
- Classification job ResourceName falls back to JobId when Name is empty
- Registered in services.go alphabetically after lambda, before opensearch
- 16 tests across 3 test functions: all pass with race detector enabled
- No sync primitives used — concurrency-unaware per NFR57
- go build, go test, go vet, go test -race all pass clean
- ✅ Resolved review finding [Medium]: batchGetFindings now returns error and propagates to lastErr in ListFindings Call, breaking pagination on failure
- ✅ Resolved review finding [Medium]: batchGetFindings now includes single retry for transient errors (matching Athena batch-get pattern)
- ✅ Resolved review finding [Low]: Implementation now matches text description (single retry present)

### Change Log

- 2026-03-13: Implemented Macie2 service enumeration with 3 API calls (ListClassificationJobs, ListFindings, DescribeBuckets), 16 unit tests, registered in AllServices()
- 2026-03-13: Addressed code review findings — 3 items resolved (2 Medium: error propagation in ListFindings + single retry in batchGetFindings, 1 Low: doc/impl mismatch)

### File List

**Created:**
- cmd/awtest/services/macie2/calls.go
- cmd/awtest/services/macie2/calls_test.go

**Modified:**
- cmd/awtest/services/services.go
