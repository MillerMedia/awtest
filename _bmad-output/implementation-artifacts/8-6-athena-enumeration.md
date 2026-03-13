# Story 8.6: Athena Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate Athena workgroups, saved queries, and query execution history,
So that I can discover analytics queries that may reveal data access patterns and S3 data locations.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/athena/` directory with `calls.go` implementing Athena service enumeration with 3 AWSService entries.

2. **AC2:** Implement `athena:ListWorkGroups` API call — iterates all regions in `types.Regions`, creates Athena client per region using config override pattern (`athena.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListWorkGroupsWithContext` with NextToken pagination (max 50 per page). Each workgroup listed with Name, State, Description, CreationTime, EngineVersion, and Region.

3. **AC3:** Implement `athena:ListNamedQueries` API call — iterates all regions, creates Athena client per region using config override, first lists all workgroups (via `ListWorkGroupsWithContext`, paginated), then for each workgroup calls `ListNamedQueriesWithContext` with WorkGroup filter and NextToken pagination (max 50 per page) to collect NamedQueryIds, then calls `BatchGetNamedQueryWithContext` (max 50 IDs per batch) to retrieve full details. Each named query listed with Name, NamedQueryId, Database, QueryString, WorkGroup, Description, and Region.

4. **AC4:** Implement `athena:ListQueryExecutions` API call — iterates all regions, creates Athena client per region using config override, first lists all workgroups (via `ListWorkGroupsWithContext`, paginated), then for each workgroup calls `ListQueryExecutionsWithContext` with WorkGroup filter and NextToken pagination (max 50 per page) to collect QueryExecutionIds, then calls `BatchGetQueryExecutionWithContext` (max 50 IDs per batch) to retrieve full details. Each execution listed with QueryExecutionId, Query (truncated to 200 chars), StatementType, Status, Database, OutputLocation, WorkGroup, SubmissionDateTime, and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without Athena access, Athena is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register Athena service in `services/services.go` `AllServices()` function in alphabetical order (after `appsync`, before `backup`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 3 API calls.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/athena/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `athena:ListWorkGroups` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/athena/`
  - [x] Create `calls.go` with `package athena`
  - [x] Define `var AthenaCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"athena:ListWorkGroups"`
  - [x] Call: iterate `types.Regions`, create `athena.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `ListWorkGroupsWithContext` with NextToken pagination loop (max 50 per page). Define local struct `atWorkGroup` with fields: Name, State, Description, CreationTime, EngineVersion, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]atWorkGroup`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"Athena"`, ResourceType=`"workgroup"`, ResourceName=workgroupName
  - [x] `utils.PrintResult` format: `"Athena Workgroup: %s (State: %s, Engine: %s, Region: %s)"` with `utils.ColorizeItem(workgroupName)`

- [x] Task 2: Implement `athena:ListNamedQueries` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"athena:ListNamedQueries"`
  - [x] Call: iterate regions -> create Athena client with config override -> first list all workgroups via `ListWorkGroupsWithContext` (paginated, max 50) -> for each workgroup, call `ListNamedQueriesWithContext` with `WorkGroup` filter and NextToken pagination (max 50 per page), collecting all NamedQueryIds -> batch IDs into groups of 50 -> call `BatchGetNamedQueryWithContext` for each batch -> extract NamedQuery details. Define local struct `atNamedQuery` with fields: Name, NamedQueryId, Database, QueryString, WorkGroup, Description, Region. Per-workgroup errors: log with `utils.HandleAWSError` and `continue` to next workgroup. Per-region errors: `break` to next region.
  - [x] Process: type-assert `[]atNamedQuery`, build `ScanResult` with ServiceName=`"Athena"`, ResourceType=`"named-query"`, ResourceName=queryName
  - [x] `utils.PrintResult` format: `"Athena Named Query: %s (Database: %s, WorkGroup: %s, Region: %s)"` with `utils.ColorizeItem(queryName)`

- [x] Task 3: Implement `athena:ListQueryExecutions` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"athena:ListQueryExecutions"`
  - [x] Call: iterate regions -> create Athena client with config override -> first list all workgroups via `ListWorkGroupsWithContext` (paginated, max 50) -> for each workgroup, call `ListQueryExecutionsWithContext` with `WorkGroup` filter and NextToken pagination (max 50 per page), collecting all QueryExecutionIds -> batch IDs into groups of 50 -> call `BatchGetQueryExecutionWithContext` for each batch -> extract QueryExecution details. Define local struct `atQueryExecution` with fields: QueryExecutionId, Query, StatementType, Status, Database, OutputLocation, WorkGroup, SubmissionDateTime, Region. Per-workgroup errors: log with `utils.HandleAWSError` and `continue` to next workgroup. Per-region errors: `break` to next region.
  - [x] Process: type-assert `[]atQueryExecution`, build `ScanResult` with ServiceName=`"Athena"`, ResourceType=`"query-execution"`, ResourceName=queryExecutionId
  - [x] `utils.PrintResult` format: `"Athena Query Execution: %s (Status: %s, Database: %s, WorkGroup: %s, Region: %s)"` with `utils.ColorizeItem(queryExecutionId)`
  - [x] **IMPORTANT:** Truncate `Query` field to 200 characters in the struct to avoid excessive output. Store truncated version with "..." suffix if truncated.

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/athena"` to `services/services.go` (alphabetical in imports: after `appsync`, before `backup`)
  - [x] Add `allServices = append(allServices, athena.AthenaCalls...)` after `appsync.AppSyncCalls...` and before `backup.BackupCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/athena/calls_test.go`
  - [x] Test `ListWorkGroups` Process: valid workgroups with details (name, state, description, creation time, engine version), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListNamedQueries` Process: valid named queries with details (name, query ID, database, query string, workgroup, description), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListQueryExecutions` Process: valid executions with details (execution ID, query snippet, statement type, status, database, output location, workgroup, submission time), empty results, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following CodeBuild/CodeCommit/OpenSearch/SageMaker/Backup test pattern
  - [x] Access Process via `AthenaCalls[0].Process`, `AthenaCalls[1].Process`, `AthenaCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/athena/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := athena.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := athena.New(sess)
}
```

### Athena is a REGIONAL Service

AWS Athena is **regional** — workgroups, named queries, and query executions exist per-region. Iterate `types.Regions` for all three API calls, following the same pattern as CodeBuild, CodeCommit, OpenSearch, SageMaker, Backup.

If Athena returns `AccessDeniedException` in a region, handle as a non-fatal error: log with `utils.HandleAWSError` and `break` to next region.

### Athena SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/athena`

**IMPORTANT:** The Go package name is `athena` which is the SAME as our local package name. This is the same pattern as backup/sagemaker/codebuild — Go resolves this correctly. Within `calls.go`, `athena.New()` and `athena.ListWorkGroupsInput{}` refer to the **AWS SDK package**, while local types (structs, variables) are referenced directly without package prefix.

**API Methods:**

1. **ListWorkGroups (Paginated):**
   - `svc.ListWorkGroupsWithContext(ctx, &athena.ListWorkGroupsInput{MaxResults: aws.Int64(50), NextToken: nextToken})` -> `*athena.ListWorkGroupsOutput`
   - `.WorkGroups` -> `[]*athena.WorkGroupSummary`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: max 50
   - Each `WorkGroupSummary` has:
     - `Name *string`
     - `State *string` ("ENABLED" or "DISABLED")
     - `Description *string`
     - `CreationTime *time.Time`
     - `EngineVersion *athena.EngineVersion` (nested struct with `SelectedEngineVersion *string`, `EffectiveEngineVersion *string`)

2. **ListNamedQueries (Paginated, returns IDs only):**
   - `svc.ListNamedQueriesWithContext(ctx, &athena.ListNamedQueriesInput{MaxResults: aws.Int64(50), WorkGroup: aws.String(wgName), NextToken: nextToken})` -> `*athena.ListNamedQueriesOutput`
   - `.NamedQueryIds` -> `[]*string` (IDs only — must batch-get details)
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: max 50
   - **IMPORTANT:** `WorkGroup` parameter is REQUIRED to enumerate per-workgroup. Without it, only returns queries from the "primary" workgroup. Must iterate all workgroups discovered in step 1.

3. **BatchGetNamedQuery (Non-paginated, batch by IDs):**
   - `svc.BatchGetNamedQueryWithContext(ctx, &athena.BatchGetNamedQueryInput{NamedQueryIds: idBatch})` -> `*athena.BatchGetNamedQueryOutput`
   - `.NamedQueries` -> `[]*athena.NamedQuery`
   - `.UnprocessedNamedQueryIds` -> `[]*athena.UnprocessedNamedQueryId` (failed lookups — log and continue)
   - **Max 50 IDs per call** — batch input IDs into groups of 50
   - Each `NamedQuery` has:
     - `Name *string`
     - `NamedQueryId *string`
     - `Database *string`
     - `QueryString *string`
     - `WorkGroup *string`
     - `Description *string`

4. **ListQueryExecutions (Paginated, returns IDs only):**
   - `svc.ListQueryExecutionsWithContext(ctx, &athena.ListQueryExecutionsInput{MaxResults: aws.Int64(50), WorkGroup: aws.String(wgName), NextToken: nextToken})` -> `*athena.ListQueryExecutionsOutput`
   - `.QueryExecutionIds` -> `[]*string` (IDs only — must batch-get details)
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: max 50
   - **IMPORTANT:** `WorkGroup` parameter is REQUIRED for thorough enumeration (same as ListNamedQueries). Results are returned in reverse chronological order.

5. **BatchGetQueryExecution (Non-paginated, batch by IDs):**
   - `svc.BatchGetQueryExecutionWithContext(ctx, &athena.BatchGetQueryExecutionInput{QueryExecutionIds: idBatch})` -> `*athena.BatchGetQueryExecutionOutput`
   - `.QueryExecutions` -> `[]*athena.QueryExecution`
   - `.UnprocessedQueryExecutionIds` -> `[]*athena.UnprocessedQueryExecutionId` (failed lookups — log and continue)
   - **Max 50 IDs per call** — batch input IDs into groups of 50
   - Each `QueryExecution` has:
     - `QueryExecutionId *string`
     - `Query *string` (the actual SQL — truncate to 200 chars for output)
     - `StatementType *string` ("DDL", "DML", "UTILITY")
     - `WorkGroup *string`
     - `Status *athena.QueryExecutionStatus`:
       - `State *string` ("QUEUED", "RUNNING", "SUCCEEDED", "FAILED", "CANCELLED")
       - `SubmissionDateTime *time.Time`
       - `CompletionDateTime *time.Time`
     - `QueryExecutionContext *athena.QueryExecutionContext`:
       - `Database *string`
       - `Catalog *string`
     - `ResultConfiguration *athena.ResultConfiguration`:
       - `OutputLocation *string` (S3 path — security-relevant!)

**No new dependencies needed** — Athena is part of `aws-sdk-go v1.44.266` already in go.mod.

### Pagination Pattern (Calls 1, 2 list phase, 3 list phase)

Calls 1, 2, and 3 use NextToken pagination. **MaxResults is 50 for ALL Athena list APIs** (NOT 1000 like Backup). Follow this exact pattern:

```go
var allWorkGroups []atWorkGroup
for _, region := range types.Regions {
    svc := athena.New(sess, &aws.Config{Region: aws.String(region)})
    var nextToken *string
    for {
        input := &athena.ListWorkGroupsInput{
            MaxResults: aws.Int64(50),
        }
        if nextToken != nil {
            input.NextToken = nextToken
        }
        output, err := svc.ListWorkGroupsWithContext(ctx, input)
        if err != nil {
            utils.HandleAWSError(false, "athena:ListWorkGroups", err)
            break
        }
        for _, wg := range output.WorkGroups {
            // nil-safe extraction, append to allWorkGroups
        }
        if output.NextToken == nil {
            break
        }
        nextToken = output.NextToken
    }
}
```

### Dependent Call Pattern (Calls 2 and 3)

Calls 2 and 3 depend on workgroup names from ListWorkGroups. Each Call function must independently list workgroups first, then iterate. This is the same pattern as Backup's recovery points and vault policies:

```go
// Call 2: ListNamedQueries
// Step 1: List all workgroups in region (paginated, max 50)
// Step 2: For each workgroup, list named query IDs (paginated, max 50)
// Step 3: Batch-get named query details (max 50 IDs per batch)

// Call 3: ListQueryExecutions
// Step 1: List all workgroups in region (paginated, max 50)
// Step 2: For each workgroup, list query execution IDs (paginated, max 50)
// Step 3: Batch-get query execution details (max 50 IDs per batch)
```

### Batch-Get Pattern (NEW for this service)

Athena's ListNamedQueries and ListQueryExecutions return only IDs. You must call BatchGetNamedQuery / BatchGetQueryExecution to get details. Batch IDs into groups of 50:

```go
// Collect all IDs from paginated list
var allIds []*string
// ... pagination loop collecting IDs ...

// Batch into groups of 50
for i := 0; i < len(allIds); i += 50 {
    end := i + 50
    if end > len(allIds) {
        end = len(allIds)
    }
    batch := allIds[i:end]

    batchOutput, err := svc.BatchGetNamedQueryWithContext(ctx, &athena.BatchGetNamedQueryInput{
        NamedQueryIds: batch,
    })
    if err != nil {
        utils.HandleAWSError(false, "athena:BatchGetNamedQuery", err)
        continue
    }
    for _, nq := range batchOutput.NamedQueries {
        // nil-safe extraction, append to results
    }
    // Log unprocessed IDs if any
    if len(batchOutput.UnprocessedNamedQueryIds) > 0 {
        utils.HandleAWSError(false, "athena:BatchGetNamedQuery",
            fmt.Errorf("%d unprocessed query IDs", len(batchOutput.UnprocessedNamedQueryIds)))
    }
}
```

### Query String Truncation (Call 3)

Query execution results include the full SQL query which can be very long. Truncate to 200 characters for the stored struct to avoid excessive memory and output:

```go
query := ""
if qe.Query != nil {
    query = *qe.Query
    if len(query) > 200 {
        query = query[:200] + "..."
    }
}
```

### Local Struct Definitions

```go
type atWorkGroup struct {
    Name          string
    State         string
    Description   string
    CreationTime  string
    EngineVersion string
    Region        string
}

type atNamedQuery struct {
    Name         string
    NamedQueryId string
    Database     string
    QueryString  string
    WorkGroup    string
    Description  string
    Region       string
}

type atQueryExecution struct {
    QueryExecutionId   string
    Query              string
    StatementType      string
    Status             string
    Database           string
    OutputLocation     string
    WorkGroup          string
    SubmissionDateTime string
    Region             string
}
```

### Variable & Naming Conventions

- **Package:** `athena` (directory: `cmd/awtest/services/athena/`)
- **Exported variable:** `AthenaCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"athena:ListWorkGroups"`, `"athena:ListNamedQueries"`, `"athena:ListQueryExecutions"`
- **ScanResult.ServiceName:** `"Athena"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"workgroup"`, `"named-query"`, `"query-execution"` (lowercase hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/athena"` (same name as local package — handled same as backup/sagemaker/codebuild pattern)

### Registration Order in services.go

Insert alphabetically — `athena` comes after `appsync`, before `backup`:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/appsync"
"github.com/MillerMedia/awtest/cmd/awtest/services/athena"    // NEW — after appsync, before backup
"github.com/MillerMedia/awtest/cmd/awtest/services/backup"

// In AllServices():
allServices = append(allServices, appsync.AppSyncCalls...)
allServices = append(allServices, athena.AthenaCalls...)   // NEW — after appsync, before backup
allServices = append(allServices, backup.BackupCalls...)
```

### Testing Pattern

Follow the CodeBuild/CodeCommit/OpenSearch/SageMaker/Backup test pattern — test Process() functions only with pre-built mock data:

```go
func TestListWorkGroupsProcess(t *testing.T) {
    process := AthenaCalls[0].Process
    // Table-driven tests: valid workgroups (name, state, description, creation time, engine version), empty, errors, nil fields, type assertion failure
}

func TestListNamedQueriesProcess(t *testing.T) {
    process := AthenaCalls[1].Process
    // Table-driven tests: valid named queries (name, query ID, database, query string, workgroup, description), empty, errors, nil fields, type assertion failure
}

func TestListQueryExecutionsProcess(t *testing.T) {
    process := AthenaCalls[2].Process
    // Table-driven tests: valid executions (execution ID, query snippet, statement type, status, database, output location, workgroup, submission time), empty, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `athena/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `athena.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** call ListNamedQueries or ListQueryExecutions without WorkGroup parameter — without it, only primary workgroup queries are returned, missing queries in other workgroups
- **DO NOT** pass more than 50 IDs to BatchGetNamedQuery or BatchGetQueryExecution — batch into groups of 50
- **DO NOT** confuse `athena.ListWorkGroupsInput` (AWS SDK type) with local package types — AWS SDK `athena` is the imported package, local types are referenced without prefix
- **DO NOT** use MaxResults > 50 for any Athena list API — the maximum is 50, not 1000

### Nested Struct Field Extraction (EngineVersion, Status, QueryExecutionContext, ResultConfiguration)

Athena has several nested structs that require careful nil-safe extraction:

```go
// EngineVersion (nested in WorkGroupSummary)
engineVersion := ""
if wg.EngineVersion != nil && wg.EngineVersion.EffectiveEngineVersion != nil {
    engineVersion = *wg.EngineVersion.EffectiveEngineVersion
}

// Status (nested in QueryExecution)
status := ""
submissionDateTime := ""
if qe.Status != nil {
    if qe.Status.State != nil {
        status = *qe.Status.State
    }
    if qe.Status.SubmissionDateTime != nil {
        submissionDateTime = qe.Status.SubmissionDateTime.Format(time.RFC3339)
    }
}

// QueryExecutionContext (nested in QueryExecution)
database := ""
if qe.QueryExecutionContext != nil && qe.QueryExecutionContext.Database != nil {
    database = *qe.QueryExecutionContext.Database
}

// ResultConfiguration (nested in QueryExecution)
outputLocation := ""
if qe.ResultConfiguration != nil && qe.ResultConfiguration.OutputLocation != nil {
    outputLocation = *qe.ResultConfiguration.OutputLocation
}
```

### Previous Story Intelligence

**From Story 8.5 (AWS Backup — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- NextToken pagination: exact pattern with `if nextToken != nil { input.NextToken = nextToken }` before call
- Dependent call pattern: list parent resources first (vaults), then query per resource — **directly applicable** to Athena calls 2 and 3 (list workgroups first, then query per workgroup)
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `BackupCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "Athena", MethodName: "athena:ListWorkGroups", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests
- 4 AWSService entries per file — Athena has 3 entries (simpler)

**From Story 8.4 (SageMaker):**
- 4 API calls pattern with regional iteration — same structure
- Local struct naming convention: `sm*` prefix for SageMaker, `bk*` for Backup, use `at*` prefix for Athena

**From Story 8.3 (OpenSearch):**
- Dependent call pattern: list names first, then describe per name — directly applicable
- Per-resource errors in inner loop: log with `utils.HandleAWSError` and `continue` to next resource

**From Story 7.2 Code Review Findings:**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [LOW] Tests should cover nil fields comprehensively

**From Story 7.1 Code Review Findings:**
- [HIGH] Always add pagination from the start (NextToken loops on paginated APIs)
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [MEDIUM] Include ARN in Details map where available
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `2bdd8e2` — Add SageMaker enumeration with 4 API calls (Story 8.4)
- `d7271c8` — Add OpenSearch enumeration with 3 API calls (Story 8.3)
- `0dd5f6a` — Add CodeCommit enumeration with 2 API calls (Story 8.2)
- `d6dd093` — Add CodeBuild enumeration with 3 API calls (Story 8.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add Athena enumeration with 3 API calls (Story 8.6)"`

### FRs Covered

- **FR94:** System enumerates Athena workgroups, saved queries, and query execution history

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/athena/
+-- calls.go            # Athena service implementation (3 AWSService entries)
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
cmd/awtest/services/backup/calls.go          # Reference implementation (regional + dependent call + 4 APIs, most recent)
cmd/awtest/services/backup/calls_test.go     # Reference test pattern (most recent)
cmd/awtest/services/sagemaker/calls.go       # Reference implementation (regional multi-API, 4 calls)
cmd/awtest/services/opensearch/calls.go      # Reference implementation (regional + dependent call pattern)
cmd/awtest/services/codebuild/calls.go       # Reference implementation (regional multi-API + batch pattern)
cmd/awtest/services/codecommit/calls.go      # Reference implementation (regional + pagination pattern)
go.mod                                       # AWS SDK already includes athena package
```

### References

- [Source: epics-phase2.md#Story 3.6: Athena Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR94] — Athena enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/backup/calls.go] — Most recent reference implementation (regional + dependent call, 4 APIs)
- [Source: cmd/awtest/services/backup/calls_test.go] — Most recent reference test pattern
- [Source: cmd/awtest/services/sagemaker/calls.go] — Reference implementation (regional + multi-API, 4 calls)
- [Source: cmd/awtest/services/opensearch/calls.go] — Reference implementation (regional + dependent call pattern)
- [Source: cmd/awtest/services/codebuild/calls.go] — Reference implementation (regional + batch API, 3 calls)
- [Source: cmd/awtest/services/codecommit/calls.go] — Reference implementation (regional + pagination pattern)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (athena goes after appsync, before backup)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes athena package)
- [Source: 8-5-aws-backup-enumeration.md] — Most recent story (dependent call pattern, 4 APIs)
- [Source: 8-4-sagemaker-enumeration.md] — Previous story (regional iteration, 4 calls)
- [Source: 8-3-opensearch-enumeration.md] — Previous story (dependent call pattern)
- [Source: 8-2-codecommit-enumeration.md] — Previous story (pagination learnings)
- [Source: 8-1-codebuild-enumeration.md] — Previous story (3 API calls pattern)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Implemented 3 AWSService entries in `AthenaCalls`: ListWorkGroups, ListNamedQueries, ListQueryExecutions
- All 3 calls iterate `types.Regions` with config override pattern (`athena.New(sess, &aws.Config{Region: ...})`) per 7.2 code review fix
- ListNamedQueries and ListQueryExecutions use dependent call pattern: list workgroups first, then enumerate per workgroup
- Batch-get pattern implemented for named queries and query executions (max 50 IDs per batch)
- Query string truncation to 200 runes (not bytes) with "..." suffix — rune-safe via `truncateRuneSafe()` helper
- Nested struct extraction for EngineVersion, Status, QueryExecutionContext, ResultConfiguration with nil-safe checks
- No sync primitives used — concurrency-unaware per NFR57

**Code Review Follow-up Fixes (2026-03-13):**
- [HIGH] PrintResult now shows Query (truncated to 120 runes) and OutputLocation for ListQueryExecutions; shows QueryString (truncated to 120 runes) for ListNamedQueries — these are the primary pentesting values
- [MEDIUM] Streaming batch processing: IDs are now batch-fetched per pagination page (max 50) instead of accumulating all IDs first — prevents memory issues with large workgroup histories
- [MEDIUM] Query truncation uses `[]rune` slicing via `truncateRuneSafe()` to avoid corrupting multi-byte UTF-8 characters
- [MEDIUM] `StateChangeReason` captured for query executions — shows why queries failed (e.g., "Access Denied", "Table not found")
- [LOW] Single retry added for unprocessed IDs from `BatchGetNamedQuery` and `BatchGetQueryExecution`
- 22 tests across 4 test functions (added: failed execution with StateChangeReason test, truncateRuneSafe tests)
- All tests pass, no regressions, no race conditions

### Change Log

- 2026-03-13: Addressed code review findings — 5 items resolved (1 High, 3 Medium, 1 Low)
- 2026-03-13: Implemented Athena enumeration with 3 API calls (Story 8.6)

### File List

- `cmd/awtest/services/athena/calls.go` (NEW) — Athena service implementation with 3 AWSService entries
- `cmd/awtest/services/athena/calls_test.go` (NEW) — Process() tests for all 3 entries (22 test cases)
- `cmd/awtest/services/services.go` (MODIFIED) — Added athena import and registration in AllServices()
