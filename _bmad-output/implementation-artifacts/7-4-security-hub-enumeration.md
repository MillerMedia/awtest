# Story 7.4: Security Hub Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate Security Hub findings, enabled products, and compliance status,
So that I can understand the aggregated security posture and identify compliance gaps.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/securityhub/` directory with `calls.go` implementing Security Hub service enumeration.

2. **AC2:** Implement `securityhub:GetEnabledStandards` API call — iterates all regions in `types.Regions`, creates Security Hub client per region using config override pattern (`securityhub.New(sess, &aws.Config{Region: aws.String(region)})`), calls `GetEnabledStandardsWithContext` (paginated via NextToken), listing each enabled compliance standard with subscription ARN, standard ARN, and status (PENDING/READY/INCOMPLETE/DELETING/FAILED).

3. **AC3:** Implement `securityhub:GetFindings` API call — iterates all regions, creates Security Hub client per region using config override, calls `GetFindingsWithContext` with filter `RecordState = ACTIVE`, sorted by severity label DESC, limited to first 100 findings per region (MaxResults=100, first page only, no pagination). Each finding listed with ID, title, severity label, compliance status, product name, resource type, and region.

4. **AC4:** Implement `securityhub:ListEnabledProductsForImport` API call — iterates all regions, creates Security Hub client per region using config override, calls `ListEnabledProductsForImportWithContext` (paginated via NextToken), listing each product subscription ARN with the region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without Security Hub access, or Security Hub not enabled in a region (ResourceNotFoundException), Security Hub is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register Security Hub service in `services/services.go` `AllServices()` function in alphabetical order (after `secretsmanager`, before `ses`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid standards/findings/products, empty results, access denied errors, nil field handling, type assertion failure handling.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/securityhub/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `securityhub:GetEnabledStandards` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/securityhub/`
  - [x] Create `calls.go` with `package securityhub`
  - [x] Define `var SecurityHubCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"securityhub:GetEnabledStandards"`
  - [x] Call: iterate `types.Regions`, create `securityhub.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), paginate `GetEnabledStandardsWithContext` via NextToken. Define local struct `shStandard` with fields: StandardsSubscriptionArn, StandardsArn, StandardsStatus, Region. Per-region errors where Security Hub is not enabled (ResourceNotFoundException): `break` to next region, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]shStandard`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"SecurityHub"`, ResourceType=`"standard"`, ResourceName=standardsArn (extracting human-readable name from ARN if possible)
  - [x] `utils.PrintResult` format: `"Security Hub Standard: %s (Status: %s, Region: %s)"` with `utils.ColorizeItem(standardName)`

- [x] Task 2: Implement `securityhub:GetFindings` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"securityhub:GetFindings"`
  - [x] Call: iterate regions -> create Security Hub client with config override -> `GetFindingsWithContext` with `Filters: &securityhub.AwsSecurityFindingFilters{RecordState: []*securityhub.StringFilter{{Comparison: aws.String("EQUALS"), Value: aws.String("ACTIVE")}}}`, `MaxResults: aws.Int64(100)`, `SortCriteria: []*securityhub.SortCriterion{{Field: aws.String("SeverityLabel"), SortOrder: aws.String("desc")}}` (first page only, no pagination — cap at 100 findings per region)
  - [x] Define local struct `shFinding` with fields: Id, Title, SeverityLabel, ComplianceStatus, ProductName, ResourceType, Region, GeneratorId
  - [x] Per-region errors where Security Hub is not enabled: `break` to next region, don't abort scan
  - [x] Process: type-assert `[]shFinding`, build `ScanResult` with ServiceName=`"SecurityHub"`, ResourceType=`"finding"`, ResourceName=title (or generatorId if title empty)
  - [x] `utils.PrintResult` format: `"Security Hub Finding: %s (Severity: %s, Compliance: %s)"` with `utils.ColorizeItem(title)`

- [x] Task 3: Implement `securityhub:ListEnabledProductsForImport` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"securityhub:ListEnabledProductsForImport"`
  - [x] Call: iterate regions -> create Security Hub client with config override -> `ListEnabledProductsForImportWithContext` (paginated via NextToken) to get product subscription ARNs
  - [x] Define local struct `shProduct` with fields: ProductSubscriptionArn, Region
  - [x] Per-region errors where Security Hub is not enabled: `break` to next region, don't abort scan
  - [x] Process: type-assert `[]shProduct`, build `ScanResult` with ServiceName=`"SecurityHub"`, ResourceType=`"product"`, ResourceName=productSubscriptionArn
  - [x] `utils.PrintResult` format: `"Security Hub Product: %s (Region: %s)"` with `utils.ColorizeItem(productArn)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/securityhub"` to `services/services.go` (alphabetical in imports: after `secretsmanager`, before `ses`)
  - [x] Add `allServices = append(allServices, securityhub.SecurityHubCalls...)` after `secretsmanager.SecretsManagerCalls...` and before `ses.SESCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/securityhub/calls_test.go`
  - [x] Test `GetEnabledStandards` Process: valid standards with status, empty results, access denied error, nil fields, type assertion failure
  - [x] Test `GetFindings` Process: valid findings with severity/compliance, empty findings, error handling, nil fields, type assertion failure, finding with empty title uses generatorId
  - [x] Test `ListEnabledProductsForImport` Process: valid products, empty products, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following GuardDuty test pattern
  - [x] Access Process via `SecurityHubCalls[0].Process`, `SecurityHubCalls[1].Process`, `SecurityHubCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/securityhub/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := securityhub.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := securityhub.New(sess)
}
```

### Security Hub is a REGIONAL Service

Security Hub is **regional** — it must be enabled per-region and queried per-region. Iterate `types.Regions` for all 3 API calls, just like GuardDuty.

If Security Hub is NOT enabled in a region, API calls return `ResourceNotFoundException` or `InvalidAccessException`. Handle this as a non-fatal error: log with `utils.HandleAWSError` and `break` to next region.

### Security Hub SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/securityhub`

**API Methods:**

1. **GetEnabledStandards:**
   - `svc.GetEnabledStandardsWithContext(ctx, &securityhub.GetEnabledStandardsInput{})` -> `*securityhub.GetEnabledStandardsOutput`
   - `.StandardsSubscriptions` -> `[]*securityhub.StandardsSubscription`
   - Each StandardsSubscription has: `StandardsSubscriptionArn *string`, `StandardsArn *string`, `StandardsStatus *string` (PENDING/READY/INCOMPLETE/DELETING/FAILED)
   - Paginated via `NextToken *string`

2. **GetFindings:**
   - `svc.GetFindingsWithContext(ctx, &securityhub.GetFindingsInput{...})` -> `*securityhub.GetFindingsOutput`
   - `.Findings` -> `[]*securityhub.AwsSecurityFinding`
   - Key finding fields: `Id *string`, `Title *string`, `Description *string`, `GeneratorId *string`, `ProductName *string`, `Region *string`
   - `Severity` -> `*securityhub.Severity` with `.Label *string` (CRITICAL/HIGH/MEDIUM/LOW/INFORMATIONAL)
   - `Compliance` -> `*securityhub.Compliance` with `.Status *string` (PASSED/FAILED/WARNING/NOT_AVAILABLE)
   - `Resources` -> `[]*securityhub.Resource` with `.Type *string`, `.Id *string`
   - `Workflow` -> `*securityhub.Workflow` with `.Status *string` (NEW/NOTIFIED/SUPPRESSED/RESOLVED)
   - Paginated via `NextToken *string`, `MaxResults *int64` (max 100)
   - **Filter struct:** `AwsSecurityFindingFilters` with `RecordState []*securityhub.StringFilter`
   - **StringFilter:** `Comparison *string` ("EQUALS"), `Value *string`
   - **SortCriterion:** `Field *string` ("SeverityLabel"), `SortOrder *string` ("desc")
   - **SortCriteria is `[]*SortCriterion`** (slice, not single struct)

3. **ListEnabledProductsForImport:**
   - `svc.ListEnabledProductsForImportWithContext(ctx, &securityhub.ListEnabledProductsForImportInput{})` -> `*securityhub.ListEnabledProductsForImportOutput`
   - `.ProductSubscriptions` -> `[]*string` (ARN strings only — not rich objects)
   - Paginated via `NextToken *string`

**No new dependencies needed** — Security Hub is part of `aws-sdk-go v1.44.266` already in go.mod.

### Findings Filter and Sort Strategy

Filter to `RecordState = ACTIVE` to exclude archived findings. Sort by `SeverityLabel` DESC to show most critical findings first. Limit to **100 findings per region** (first page only, no pagination on GetFindings). This prevents overwhelming output while showing the most important findings.

```go
getFindingsInput := &securityhub.GetFindingsInput{
    Filters: &securityhub.AwsSecurityFindingFilters{
        RecordState: []*securityhub.StringFilter{
            {
                Comparison: aws.String("EQUALS"),
                Value:      aws.String("ACTIVE"),
            },
        },
    },
    MaxResults: aws.Int64(100),
    SortCriteria: []*securityhub.SortCriterion{
        {
            Field:     aws.String("SeverityLabel"),
            SortOrder: aws.String("desc"),
        },
    },
}
```

### Local Struct Definitions

```go
type shStandard struct {
    StandardsSubscriptionArn string
    StandardsArn             string
    StandardsStatus          string
    Region                   string
}

type shFinding struct {
    Id               string
    Title            string
    SeverityLabel    string
    ComplianceStatus string
    ProductName      string
    ResourceType     string
    Region           string
    GeneratorId      string
}

type shProduct struct {
    ProductSubscriptionArn string
    Region                 string
}
```

### Findings Field Extraction Pattern

```go
severityLabel := ""
if finding.Severity != nil && finding.Severity.Label != nil {
    severityLabel = *finding.Severity.Label
}

complianceStatus := ""
if finding.Compliance != nil && finding.Compliance.Status != nil {
    complianceStatus = *finding.Compliance.Status
}

resourceType := ""
if len(finding.Resources) > 0 && finding.Resources[0].Type != nil {
    resourceType = *finding.Resources[0].Type
}
```

### Handling "Not Enabled" Errors

Security Hub returns `ResourceNotFoundException` (or `InvalidAccessException`) when not enabled in a region. This maps to the same access denied handling as other services. In the Call function, when iterating regions:

```go
output, err := svc.GetEnabledStandardsWithContext(ctx, input)
if err != nil {
    lastErr = err
    utils.HandleAWSError(false, "securityhub:GetEnabledStandards", err)
    break // Skip to next region — Security Hub likely not enabled
}
```

Use `break` (not `continue`) for the region-level pagination loop to move on to the next region entirely.

### Variable & Naming Conventions

- **Package:** `securityhub` (directory: `cmd/awtest/services/securityhub/`)
- **Exported variable:** `SecurityHubCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"securityhub:GetEnabledStandards"`, `"securityhub:GetFindings"`, `"securityhub:ListEnabledProductsForImport"`
- **ScanResult.ServiceName:** `"SecurityHub"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"standard"`, `"finding"`, `"product"` (lowercase singular)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)

### Registration Order in services.go

Insert alphabetically — `securityhub` comes after `secretsmanager`, before `ses`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/securityhub"

// In AllServices():
allServices = append(allServices, secretsmanager.SecretsManagerCalls...)
allServices = append(allServices, securityhub.SecurityHubCalls...)  // NEW
allServices = append(allServices, ses.SESCalls...)
```

### Testing Pattern

Follow the GuardDuty test pattern — test Process() functions only with pre-built mock data:

```go
func TestGetEnabledStandardsProcess(t *testing.T) {
    process := SecurityHubCalls[0].Process
    // Table-driven tests with valid standards, empty, errors, nil fields, type assertion failure
}

func TestGetFindingsProcess(t *testing.T) {
    process := SecurityHubCalls[1].Process
    // Test with findings with severity/compliance, empty, errors, nil fields, type assertion failure
    // Test finding with empty title uses generatorId as ResourceName
}

func TestListEnabledProductsProcess(t *testing.T) {
    process := SecurityHubCalls[2].Process
    // Test with valid products, empty products, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `securityhub/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `securityhub.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** paginate GetFindings — limit to first 100 per region (MaxResults=100, no NextToken loop)
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** treat ResourceNotFoundException as a fatal error — it means Security Hub is not enabled in that region, just skip

### Previous Story Intelligence

**From Story 7.3 (GuardDuty — previous story in this epic):**
- All 3 Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for multi-step call results (same pattern: define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- Pagination: NextToken loop for paginated APIs (GetEnabledStandards, ListEnabledProductsForImport)
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `SecurityHubCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "SecurityHub", MethodName: "securityhub:GetEnabledStandards", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests
- GuardDuty `Features` stored as `[]string` in struct (not comma-separated string) — follow same pattern for any slice fields

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
- `0de1823` — Add GuardDuty enumeration with 3 API calls (Story 7.3)
- `2023f44` — Add AWS Organizations enumeration with 3 API calls (Story 7.2)
- `2c0b4ab` — Add ECR container registry enumeration with 3 API calls (Story 7.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message

### FRs Covered

- **FR88:** System enumerates Security Hub findings, enabled products, and compliance status

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/securityhub/
├── calls.go            # Security Hub service implementation (3 AWSService entries)
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
cmd/awtest/services/guardduty/calls.go       # Reference implementation (same epic, same patterns)
cmd/awtest/services/guardduty/calls_test.go  # Reference test pattern (table-driven Process-only tests)
go.mod                                       # AWS SDK already includes Security Hub package
```

### References

- [Source: epics-phase2.md#Story 2.4: Security Hub Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR88] — Security Hub enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/guardduty/calls.go] — Reference implementation (same epic, config override pattern for region)
- [Source: cmd/awtest/services/guardduty/calls_test.go] — Reference test pattern (table-driven Process-only tests)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (securityhub goes after secretsmanager, before ses)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes Security Hub package)
- [Source: 7-3-guardduty-enumeration.md] — Previous story in same epic (patterns, all learnings)
- [Source: 7-2-aws-organizations-enumeration.md] — Earlier story in epic (code review findings: config override, error logging)
- [Source: 7-1-ecr-container-registry-enumeration.md] — First story in epic (code review findings: pagination, table-driven tests)
- [Source: code-review-findings.md] — Code review findings for Story 7.2 (race condition fix, error swallowing fix)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered.

### Completion Notes List

- Implemented Security Hub enumeration with 3 API calls following GuardDuty pattern
- `securityhub:GetEnabledStandards` — paginated, iterates all regions with config override, extracts human-readable standard name from ARN
- `securityhub:GetFindings` — filtered to ACTIVE record state, sorted by severity DESC, capped at 100 findings per region (no pagination), findings with empty title fall back to GeneratorId
- `securityhub:ListEnabledProductsForImport` — paginated, iterates all regions with config override
- All 3 Process functions: error handling via HandleAWSError, type assertion with graceful failure, nil-safe pointer dereferencing on all fields
- Config override pattern used for all region iteration (no session mutation per 7.2 code review finding)
- No sync primitives imported — service is concurrency-unaware per NFR57
- Registered in services.go alphabetically after secretsmanager, before ses
- 16 table-driven test cases across 3 test functions covering valid data, empty results, access denied errors, nil fields, type assertion failures, and empty title fallback
- All builds, tests, vet, and race detection pass clean with zero regressions

### Change Log

- 2026-03-11: Implemented Security Hub enumeration service with 3 API calls (Story 7.4)

### File List

- `cmd/awtest/services/securityhub/calls.go` (new) — Security Hub service implementation with 3 AWSService entries
- `cmd/awtest/services/securityhub/calls_test.go` (new) — Table-driven Process() tests for all 3 API calls
- `cmd/awtest/services/services.go` (modified) — Added securityhub import and registration in AllServices()
