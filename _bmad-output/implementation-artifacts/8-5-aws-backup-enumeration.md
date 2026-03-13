# Story 8.5: AWS Backup Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate Backup vaults, plans, recovery points, and vault access policies,
So that I can discover backup data stores and identify cross-account backup access.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/backup/` directory with `calls.go` implementing Backup service enumeration with 4 AWSService entries.

2. **AC2:** Implement `backup:ListBackupVaults` API call — iterates all regions in `types.Regions`, creates Backup client per region using config override pattern (`backup.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListBackupVaultsWithContext` with NextToken pagination (max 1000 per page). Each vault listed with BackupVaultName, BackupVaultArn, NumberOfRecoveryPoints, EncryptionKeyArn, CreationDate, Locked status, and Region.

3. **AC3:** Implement `backup:ListBackupPlans` API call — iterates all regions, creates Backup client per region using config override, calls `ListBackupPlansWithContext` with NextToken pagination (max 1000 per page). Each plan listed with BackupPlanName, BackupPlanId, BackupPlanArn, CreationDate, LastExecutionDate, VersionId, and Region.

4. **AC4:** Implement `backup:ListRecoveryPointsByBackupVault` API call — iterates all regions, first lists all vaults (via `ListBackupVaultsWithContext`), then for each vault calls `ListRecoveryPointsByBackupVaultWithContext` with NextToken pagination (max 1000 per page). Each recovery point listed with RecoveryPointArn, BackupVaultName, ResourceArn, ResourceType, Status, CreationDate, BackupSizeInBytes, and Region.

5. **AC5:** Implement `backup:GetBackupVaultAccessPolicy` API call — iterates all regions, first lists all vaults (via `ListBackupVaultsWithContext`), then for each vault calls `GetBackupVaultAccessPolicyWithContext`. Handle `ResourceNotFoundException` gracefully (vault has no access policy — skip silently, do not report as error). Each policy listed with BackupVaultName, BackupVaultArn, Policy (JSON string), and Region.

6. **AC6:** All four Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

7. **AC7:** Given credentials without Backup access, Backup is skipped silently (access denied handling via existing error classification in safeScan).

8. **AC8:** Register Backup service in `services/services.go` `AllServices()` function in alphabetical order (after `batch`, before `certificatemanager`).

9. **AC9:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 4 API calls.

10. **AC10:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

11. **AC11:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/backup/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `backup:ListBackupVaults` (AC: 1, 2, 6, 7, 10)
  - [x] Create directory `cmd/awtest/services/backup/`
  - [x] Create `calls.go` with `package backup`
  - [x] Define `var BackupCalls = []types.AWSService{...}` with 4 entries
  - [x] Implement first entry: Name `"backup:ListBackupVaults"`
  - [x] Call: iterate `types.Regions`, create `backup.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `ListBackupVaultsWithContext` with NextToken pagination loop (max 1000 per page). Define local struct `bkVault` with fields: Name, Arn, RecoveryPointCount, EncryptionKeyArn, CreationDate, Locked, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]bkVault`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"Backup"`, ResourceType=`"backup-vault"`, ResourceName=vaultName
  - [x] `utils.PrintResult` format: `"Backup Vault: %s (Recovery Points: %d, Locked: %v, Region: %s)"` with `utils.ColorizeItem(vaultName)`

- [x] Task 2: Implement `backup:ListBackupPlans` (AC: 3, 6, 7, 10)
  - [x] Implement second entry: Name `"backup:ListBackupPlans"`
  - [x] Call: iterate regions -> create Backup client with config override -> `ListBackupPlansWithContext` with NextToken pagination (max 1000 per page). Define local struct `bkPlan` with fields: Name, PlanId, Arn, CreationDate, LastExecutionDate, VersionId, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: type-assert `[]bkPlan`, build `ScanResult` with ServiceName=`"Backup"`, ResourceType=`"backup-plan"`, ResourceName=planName
  - [x] `utils.PrintResult` format: `"Backup Plan: %s (Last Execution: %s, Region: %s)"` with `utils.ColorizeItem(planName)`

- [x] Task 3: Implement `backup:ListRecoveryPointsByBackupVault` (AC: 4, 6, 7, 10)
  - [x] Implement third entry: Name `"backup:ListRecoveryPointsByBackupVault"`
  - [x] Call: iterate regions -> create Backup client with config override -> first list all vaults via `ListBackupVaultsWithContext` (paginated) -> for each vault, call `ListRecoveryPointsByBackupVaultWithContext` with vault name and NextToken pagination (max 1000 per page). Define local struct `bkRecoveryPoint` with fields: RecoveryPointArn, VaultName, ResourceArn, ResourceType, Status, CreationDate, BackupSizeInBytes, Region. Per-vault errors: log with `utils.HandleAWSError` and continue to next vault. Per-region errors: `break` to next region.
  - [x] Process: type-assert `[]bkRecoveryPoint`, build `ScanResult` with ServiceName=`"Backup"`, ResourceType=`"recovery-point"`, ResourceName=recoveryPointArn
  - [x] `utils.PrintResult` format: `"Backup Recovery Point: %s (Vault: %s, Resource: %s, Status: %s, Region: %s)"` with `utils.ColorizeItem(recoveryPointArn)`

- [x] Task 4: Implement `backup:GetBackupVaultAccessPolicy` (AC: 5, 6, 7, 10)
  - [x] Implement fourth entry: Name `"backup:GetBackupVaultAccessPolicy"`
  - [x] Call: iterate regions -> create Backup client with config override -> first list all vaults via `ListBackupVaultsWithContext` (paginated) -> for each vault, call `GetBackupVaultAccessPolicyWithContext` with vault name. Handle `ResourceNotFoundException` gracefully (no policy exists — skip, do not log as error). Define local struct `bkVaultPolicy` with fields: VaultName, VaultArn, Policy, Region. Per-vault errors (non-ResourceNotFound): log with `utils.HandleAWSError` and continue.
  - [x] Process: type-assert `[]bkVaultPolicy`, build `ScanResult` with ServiceName=`"Backup"`, ResourceType=`"vault-access-policy"`, ResourceName=vaultName
  - [x] `utils.PrintResult` format: `"Backup Vault Access Policy: %s (Region: %s)"` with `utils.ColorizeItem(vaultName)`

- [x] Task 5: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/backup"` to `services/services.go` (alphabetical in imports: before `batch`)
  - [x] Add `allServices = append(allServices, backup.BackupCalls...)` before `batch.BatchCalls...`

- [x] Task 6: Write unit tests (AC: 9, 11)
  - [x] Create `cmd/awtest/services/backup/calls_test.go`
  - [x] Test `ListBackupVaults` Process: valid vaults with details (name, ARN, recovery point count, encryption key, locked status), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListBackupPlans` Process: valid plans with details (name, plan ID, ARN, creation date, last execution date), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListRecoveryPointsByBackupVault` Process: valid recovery points with details (ARN, vault name, resource ARN, resource type, status, size), empty results, error handling, nil fields, type assertion failure
  - [x] Test `GetBackupVaultAccessPolicy` Process: valid policies with details (vault name, ARN, policy JSON), empty results, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following CodeBuild/CodeCommit/OpenSearch/SageMaker test pattern
  - [x] Access Process via `BackupCalls[0].Process`, `BackupCalls[1].Process`, `BackupCalls[2].Process`, `BackupCalls[3].Process`

- [x] Task 7: Build and verify (AC: 11)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/backup/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := backup.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := backup.New(sess)
}
```

### Backup is a REGIONAL Service

AWS Backup is **regional** — vaults, plans, and recovery points exist per-region. Iterate `types.Regions` for all four API calls, following the same pattern as CodeBuild, CodeCommit, OpenSearch, SageMaker.

If Backup returns `AccessDeniedException` in a region, handle as a non-fatal error: log with `utils.HandleAWSError` and `break` to next region.

### Backup SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/backup`

**IMPORTANT:** The Go package name is `backup` which is the SAME as our local package name. This is the same pattern as SageMaker/CodeBuild — Go resolves this correctly. Within `calls.go`, `backup.New()` and `backup.ListBackupVaultsInput{}` refer to the **AWS SDK package**, while local types (structs, variables) are referenced directly without package prefix.

**API Methods:**

1. **ListBackupVaults (Paginated):**
   - `svc.ListBackupVaultsWithContext(ctx, &backup.ListBackupVaultsInput{MaxResults: aws.Int64(1000), NextToken: nextToken})` -> `*backup.ListBackupVaultsOutput`
   - `.BackupVaultList` -> `[]*backup.VaultListMember`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: min 1, default/max 1000
   - Each `VaultListMember` has:
     - `BackupVaultName *string`
     - `BackupVaultArn *string`
     - `NumberOfRecoveryPoints *int64`
     - `EncryptionKeyArn *string`
     - `CreationDate *time.Time`
     - `Locked *bool`
     - `LockDate *time.Time`
     - `MaxRetentionDays *int64`
     - `MinRetentionDays *int64`
     - `CreatorRequestId *string`

2. **ListBackupPlans (Paginated):**
   - `svc.ListBackupPlansWithContext(ctx, &backup.ListBackupPlansInput{MaxResults: aws.Int64(1000), NextToken: nextToken})` -> `*backup.ListBackupPlansOutput`
   - `.BackupPlansList` -> `[]*backup.PlansListMember`
   - Pagination: `NextToken *string` in both input and output
   - MaxResults: min 1, default/max 1000
   - Each `PlansListMember` has:
     - `BackupPlanName *string`
     - `BackupPlanId *string`
     - `BackupPlanArn *string`
     - `CreationDate *time.Time`
     - `LastExecutionDate *time.Time`
     - `VersionId *string`
     - `DeletionDate *time.Time`
     - `CreatorRequestId *string`

3. **ListRecoveryPointsByBackupVault (Paginated, per-vault):**
   - `svc.ListRecoveryPointsByBackupVaultWithContext(ctx, &backup.ListRecoveryPointsByBackupVaultInput{BackupVaultName: aws.String(vaultName), MaxResults: aws.Int64(1000), NextToken: nextToken})` -> `*backup.ListRecoveryPointsByBackupVaultOutput`
   - `.RecoveryPoints` -> `[]*backup.RecoveryPointByBackupVault`
   - Pagination: `NextToken *string` in both input and output
   - **Required input:** `BackupVaultName *string` — must list vaults first
   - Each `RecoveryPointByBackupVault` has:
     - `RecoveryPointArn *string`
     - `BackupVaultName *string`
     - `BackupVaultArn *string`
     - `ResourceArn *string`
     - `ResourceType *string` (e.g., "EBS", "RDS", "S3", "DynamoDB")
     - `ResourceName *string`
     - `Status *string` ("COMPLETED", "PARTIAL", "DELETING", "EXPIRED")
     - `CreationDate *time.Time`
     - `CompletionDate *time.Time`
     - `BackupSizeInBytes *int64`
     - `EncryptionKeyArn *string`
     - `IsEncrypted *bool`
     - `IamRoleArn *string`

4. **GetBackupVaultAccessPolicy (Non-paginated, per-vault):**
   - `svc.GetBackupVaultAccessPolicyWithContext(ctx, &backup.GetBackupVaultAccessPolicyInput{BackupVaultName: aws.String(vaultName)})` -> `*backup.GetBackupVaultAccessPolicyOutput`
   - **No pagination** — single-resource GET call
   - **Required input:** `BackupVaultName *string` — must list vaults first
   - **IMPORTANT:** Returns `ResourceNotFoundException` when no policy is set on vault — handle gracefully (skip vault, NOT an error)
   - Output fields:
     - `BackupVaultName *string`
     - `BackupVaultArn *string`
     - `Policy *string` (vault access policy as JSON string)

**No new dependencies needed** — Backup is part of `aws-sdk-go v1.44.266` already in go.mod.

### Pagination Pattern (Calls 1-3)

Calls 1, 2, and 3 use NextToken pagination. Follow this exact pattern:

```go
var allVaults []bkVault
for _, region := range types.Regions {
    svc := backup.New(sess, &aws.Config{Region: aws.String(region)})
    var nextToken *string
    for {
        input := &backup.ListBackupVaultsInput{
            MaxResults: aws.Int64(1000),
        }
        if nextToken != nil {
            input.NextToken = nextToken
        }
        output, err := svc.ListBackupVaultsWithContext(ctx, input)
        if err != nil {
            utils.HandleAWSError(false, "backup:ListBackupVaults", err)
            break
        }
        for _, v := range output.BackupVaultList {
            // nil-safe extraction, append to allVaults
        }
        if output.NextToken == nil {
            break
        }
        nextToken = output.NextToken
    }
}
```

### Dependent Call Pattern (Calls 3 and 4)

Calls 3 and 4 depend on vault names from ListBackupVaults. Each Call function must independently list vaults first, then iterate:

```go
// Call 3: ListRecoveryPointsByBackupVault
// Step 1: List all vaults in region (paginated)
// Step 2: For each vault, list recovery points (paginated)

// Call 4: GetBackupVaultAccessPolicy
// Step 1: List all vaults in region (paginated)
// Step 2: For each vault, get access policy (non-paginated)
//         Handle ResourceNotFoundException -> skip (no policy set)
```

### ResourceNotFoundException Handling (Call 4)

`GetBackupVaultAccessPolicy` returns `ResourceNotFoundException` when no access policy exists on a vault. This is **expected behavior**, not an error. Check for this specific error type and skip silently:

```go
output, err := svc.GetBackupVaultAccessPolicyWithContext(ctx, &backup.GetBackupVaultAccessPolicyInput{
    BackupVaultName: aws.String(vaultName),
})
if err != nil {
    // Check if ResourceNotFoundException — vault has no policy, skip silently
    if isResourceNotFound(err) {
        continue
    }
    utils.HandleAWSError(false, "backup:GetBackupVaultAccessPolicy", err)
    continue
}
```

Use `awserr.Error` to check for the specific error code:
```go
import "github.com/aws/aws-sdk-go/aws/awserr"

func isResourceNotFound(err error) bool {
    if aerr, ok := err.(awserr.Error); ok {
        return aerr.Code() == "ResourceNotFoundException"
    }
    return false
}
```

### Local Struct Definitions

```go
type bkVault struct {
    Name              string
    Arn               string
    RecoveryPointCount int64
    EncryptionKeyArn  string
    CreationDate      string
    Locked            bool
    Region            string
}

type bkPlan struct {
    Name              string
    PlanId            string
    Arn               string
    CreationDate      string
    LastExecutionDate string
    VersionId         string
    Region            string
}

type bkRecoveryPoint struct {
    RecoveryPointArn string
    VaultName        string
    ResourceArn      string
    ResourceType     string
    Status           string
    CreationDate     string
    BackupSizeInBytes int64
    Region           string
}

type bkVaultPolicy struct {
    VaultName string
    VaultArn  string
    Policy    string
    Region    string
}
```

### Variable & Naming Conventions

- **Package:** `backup` (directory: `cmd/awtest/services/backup/`)
- **Exported variable:** `BackupCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"backup:ListBackupVaults"`, `"backup:ListBackupPlans"`, `"backup:ListRecoveryPointsByBackupVault"`, `"backup:GetBackupVaultAccessPolicy"`
- **ScanResult.ServiceName:** `"Backup"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"backup-vault"`, `"backup-plan"`, `"recovery-point"`, `"vault-access-policy"` (lowercase hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/backup"` (same name as local package — handled same as sagemaker/codebuild pattern)
- **Additional import:** `"github.com/aws/aws-sdk-go/aws/awserr"` (for ResourceNotFoundException handling)

### Registration Order in services.go

Insert alphabetically — `backup` comes after `batch`, before `certificatemanager`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/backup"

// In AllServices():
allServices = append(allServices, batch.BatchCalls...)
allServices = append(allServices, backup.BackupCalls...)  // NEW
allServices = append(allServices, certificatemanager.CertificateManagerCalls...)
```

**WAIT — alphabetical order correction:** `backup` comes BEFORE `batch` alphabetically (b-a-c < b-a-t). The correct insertion point is:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/backup"   // NEW — before batch
"github.com/MillerMedia/awtest/cmd/awtest/services/batch"

// In AllServices():
allServices = append(allServices, appsync.AppSyncCalls...)
allServices = append(allServices, backup.BackupCalls...)  // NEW — before batch
allServices = append(allServices, batch.BatchCalls...)
```

### Testing Pattern

Follow the CodeBuild/CodeCommit/OpenSearch/SageMaker test pattern — test Process() functions only with pre-built mock data:

```go
func TestListBackupVaultsProcess(t *testing.T) {
    process := BackupCalls[0].Process
    // Table-driven tests: valid vaults (name, ARN, recovery point count, encryption key, locked), empty, errors, nil fields, type assertion failure
}

func TestListBackupPlansProcess(t *testing.T) {
    process := BackupCalls[1].Process
    // Table-driven tests: valid plans (name, plan ID, ARN, creation date, last execution date), empty, errors, nil fields, type assertion failure
}

func TestListRecoveryPointsByBackupVaultProcess(t *testing.T) {
    process := BackupCalls[2].Process
    // Table-driven tests: valid recovery points (ARN, vault name, resource ARN, type, status, size), empty, errors, nil fields, type assertion failure
}

func TestGetBackupVaultAccessPolicyProcess(t *testing.T) {
    process := BackupCalls[3].Process
    // Table-driven tests: valid policies (vault name, ARN, policy JSON), empty, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `backup/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `backup.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** treat `ResourceNotFoundException` from `GetBackupVaultAccessPolicy` as an error — it means no policy is set, skip silently
- **DO NOT** confuse `backup.ListBackupVaultsInput` (AWS SDK type) with local package types — AWS SDK `backup` is the imported package, local types are referenced without prefix
- **DO NOT** call `ListRecoveryPointsByBackupVault` or `GetBackupVaultAccessPolicy` without first listing vaults — vault names are required input

### Previous Story Intelligence

**From Story 8.4 (SageMaker — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- NextToken pagination: exact pattern with `if nextToken != nil { input.NextToken = nextToken }` before call
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `BackupCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "Backup", MethodName: "backup:ListBackupVaults", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests
- 4 AWSService entries per file — directly applicable

**From Story 8.3 (OpenSearch):**
- Dependent call pattern: list names first, then describe/query per name — **directly applicable** to Backup calls 3 and 4
- Per-resource errors in inner loop: log with `utils.HandleAWSError` and `continue` to next resource, don't abort

**From Story 8.2 (CodeCommit):**
- Pagination: NextToken loop for paginated APIs — directly applicable to all 3 paginated Backup APIs

**From Story 8.1 (CodeBuild):**
- 3 API calls in one service file — Backup extends to 4 calls, same pattern per entry
- Batch-describe pattern (vault listing before per-vault calls) applicable to calls 3 and 4

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
- `0bf797d` — Add SageMaker enumeration with 4 API calls (Story 8.4)
- `d7271c8` — Add OpenSearch enumeration with 3 API calls (Story 8.3)
- `0dd5f6a` — Add CodeCommit enumeration with 2 API calls (Story 8.2)
- `d6dd093` — Add CodeBuild enumeration with 3 API calls (Story 8.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add AWS Backup enumeration with 4 API calls (Story 8.5)"`

### FRs Covered

- **FR93:** System enumerates AWS Backup vaults, backup plans, recovery points, and vault access policies

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/backup/
+-- calls.go            # Backup service implementation (4 AWSService entries)
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
cmd/awtest/services/sagemaker/calls.go       # Reference implementation (regional multi-API, 4 calls, most recent)
cmd/awtest/services/sagemaker/calls_test.go  # Reference test pattern (most recent)
cmd/awtest/services/opensearch/calls.go      # Reference implementation (regional + dependent call pattern)
cmd/awtest/services/opensearch/calls_test.go # Reference test pattern
cmd/awtest/services/codebuild/calls.go       # Reference implementation (regional multi-API + batch pattern)
cmd/awtest/services/codecommit/calls.go      # Reference implementation (regional + pagination pattern)
go.mod                                       # AWS SDK already includes backup package
```

### References

- [Source: epics-phase2.md#Story 3.5: AWS Backup Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR93] — Backup enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/sagemaker/calls.go] — Most recent reference implementation (regional + multi-API, 4 calls)
- [Source: cmd/awtest/services/sagemaker/calls_test.go] — Most recent reference test pattern
- [Source: cmd/awtest/services/opensearch/calls.go] — Reference implementation (regional + dependent call pattern)
- [Source: cmd/awtest/services/codebuild/calls.go] — Reference implementation (regional + batch API, 3 calls)
- [Source: cmd/awtest/services/codecommit/calls.go] — Reference implementation (regional + pagination pattern)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (backup goes before batch)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes backup package)
- [Source: 8-4-sagemaker-enumeration.md] — Most recent story (patterns, regional iteration, 4 calls)
- [Source: 8-3-opensearch-enumeration.md] — Previous story (dependent call pattern)
- [Source: 8-2-codecommit-enumeration.md] — Previous story (pagination learnings)
- [Source: 8-1-codebuild-enumeration.md] — Previous story (3 API calls pattern)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no debugging required.

### Completion Notes List

- Implemented 4 AWSService entries in `backup/calls.go`: ListBackupVaults, ListBackupPlans, ListRecoveryPointsByBackupVault, GetBackupVaultAccessPolicy
- All Call functions iterate `types.Regions` using config override pattern (`backup.New(sess, &aws.Config{Region: ...})`) — no session mutation
- ListBackupVaults and ListBackupPlans: straightforward paginated calls with NextToken (max 1000 per page)
- ListRecoveryPointsByBackupVault: dependent call — lists vaults first (paginated), then lists recovery points per vault (paginated)
- GetBackupVaultAccessPolicy: dependent call — lists vaults first (paginated), then gets access policy per vault. Handles `ResourceNotFoundException` gracefully (vaults without policies are silently skipped via `isResourceNotFound()` helper using `awserr.Error` type assertion)
- All Process functions: error handling via `utils.HandleAWSError`, type assertion with graceful failure, nil-safe field extraction, comprehensive Details maps
- No sync primitives imported — concurrency-unaware per NFR57
- Registered in `services.go` alphabetically before `batch`
- 20 table-driven tests across 4 test functions covering: valid data, empty results, access denied errors, nil fields, type assertion failures
- All tests pass, `go vet` clean, `go test -race` clean with no regressions

### File List

- `cmd/awtest/services/backup/calls.go` (NEW) — Backup service implementation with 4 AWSService entries
- `cmd/awtest/services/backup/calls_test.go` (NEW) — Table-driven Process() tests for all 4 entries
- `cmd/awtest/services/services.go` (MODIFIED) — Added backup import and registration in AllServices()

### Change Log

- 2026-03-12: Implemented AWS Backup enumeration with 4 API calls (ListBackupVaults, ListBackupPlans, ListRecoveryPointsByBackupVault, GetBackupVaultAccessPolicy) and registered in AllServices(). Added comprehensive table-driven tests for all Process functions.
- 2026-03-12: Code review completed. All ACs verified, tests passed, code quality confirmed. Status updated to done.

## Senior Developer Review (AI)

- [x] Story file loaded from `_bmad-output/implementation-artifacts/8-5-aws-backup-enumeration.md`
- [x] Story Status verified as reviewable (review)
- [x] Epic and Story IDs resolved (8.5)
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

_Reviewer: Kn0ck0ut on 2026-03-12_

### Review Notes

**Verdict:** APPROVED

This is an exceptionally clean implementation.
- All 4 ACs for API calls are implemented correctly with pagination and config overrides.
- Dependent calls (Recovery Points, Vault Policies) correctly list vaults first.
- `ResourceNotFoundException` is handled gracefully for vault policies.
- Tests are comprehensive and use the table-driven pattern.
- Registration is in the correct alphabetical order.
- No concurrency primitives are used.
