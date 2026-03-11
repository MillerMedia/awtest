# Story 7.3: GuardDuty Enumeration

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate GuardDuty detectors, findings, and suppression filters,
So that I can identify detection coverage gaps and understand what security monitoring is active.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/guardduty/` directory with `calls.go` implementing GuardDuty service enumeration.

2. **AC2:** Implement `guardduty:ListDetectors` API call — iterates all regions in `types.Regions`, creates GuardDuty client per region using config override pattern (`guardduty.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListDetectorsWithContext` (paginated via NextToken), then for each detector ID calls `GetDetectorWithContext` to retrieve status and configuration. Each detector listed with detector ID, status (ENABLED/DISABLED), service role ARN, created timestamp, and enabled feature names.

3. **AC3:** Implement `guardduty:GetFindings` API call — iterates all regions, discovers detectors via `ListDetectorsWithContext`, then for each detector calls `ListFindingsWithContext` (limited to first 50 findings via MaxResults, sorted by severity DESC) to get finding IDs, then calls `GetFindingsWithContext` to batch-hydrate finding details. Each finding listed with finding ID, type, title, severity (numeric), and region.

4. **AC4:** Implement `guardduty:ListFilters` API call — iterates all regions, discovers detectors via `ListDetectorsWithContext`, then for each detector calls `ListFiltersWithContext` (paginated via NextToken) to get filter names, then calls `GetFilterWithContext` for each filter to retrieve action (ARCHIVE=suppression) and description. Each filter listed with filter name, action, description, and detector ID.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without GuardDuty access, GuardDuty is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register GuardDuty service in `services/services.go` `AllServices()` function in alphabetical order (after `glue`, before `iam`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid detectors/findings/filters, empty results, access denied errors, nil field handling, type assertion failure handling.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/guardduty/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `guardduty:ListDetectors` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/guardduty/`
  - [x] Create `calls.go` with `package guardduty`
  - [x] Define `var GuardDutyCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"guardduty:ListDetectors"`
  - [x] Call: iterate `types.Regions`, create `guardduty.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), paginate `ListDetectorsWithContext` via NextToken, then for each detector ID call `GetDetectorWithContext` to hydrate details. Define local struct `guarddutyDetector` with fields: DetectorId, Status, ServiceRole, CreatedAt, Region, Features (comma-separated feature names string). Per-detector GetDetector errors: `continue` to next detector, don't abort scan.
  - [x] Process: handle error → `utils.HandleAWSError`, type-assert `[]guarddutyDetector`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"GuardDuty"`, ResourceType=`"detector"`, ResourceName=detectorId
  - [x] `utils.PrintResult` format: `"GuardDuty Detector: %s (Status: %s, Region: %s)"` with `utils.ColorizeItem(detectorId)`

- [x] Task 2: Implement `guardduty:GetFindings` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"guardduty:GetFindings"`
  - [x] Call: iterate regions → create GuardDuty client with config override → `ListDetectorsWithContext` to discover detectors → for each detector: `ListFindingsWithContext` with `MaxResults: aws.Int64(50)` and `SortCriteria: &guardduty.SortCriteria{AttributeName: aws.String("severity"), OrderBy: aws.String("DESC")}` (first page only, no pagination — cap at 50 findings per detector per region) → `GetFindingsWithContext` with the returned finding IDs to hydrate details
  - [x] Define local struct `guarddutyFinding` with fields: FindingId, Type, Title, Severity (float64), Description, Region, DetectorId
  - [x] Per-detector errors in ListFindings/GetFindings: `continue` to next detector, don't abort scan
  - [x] Process: type-assert `[]guarddutyFinding`, build `ScanResult` with ServiceName=`"GuardDuty"`, ResourceType=`"finding"`, ResourceName=findingTitle (or findingType if title empty)
  - [x] `utils.PrintResult` format: `"GuardDuty Finding: %s (Severity: %.1f, Type: %s)"` with `utils.ColorizeItem(title)`

- [x] Task 3: Implement `guardduty:ListFilters` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"guardduty:ListFilters"`
  - [x] Call: iterate regions → create GuardDuty client with config override → `ListDetectorsWithContext` to discover detectors → for each detector: `ListFiltersWithContext` (paginated via NextToken) to get filter names → `GetFilterWithContext` for each filter name to hydrate action and description
  - [x] Define local struct `guarddutyFilter` with fields: FilterName, Action, Description, DetectorId, Region
  - [x] Per-filter GetFilter errors: `continue` to next filter, don't abort scan
  - [x] Per-detector errors in ListFilters: `continue` to next detector, don't abort scan
  - [x] Process: type-assert `[]guarddutyFilter`, build `ScanResult` with ServiceName=`"GuardDuty"`, ResourceType=`"filter"`, ResourceName=filterName
  - [x] `utils.PrintResult` format: `"GuardDuty Filter: %s (Action: %s, Detector: %s)"` with `utils.ColorizeItem(filterName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/guardduty"` to `services/services.go` (alphabetical in imports: after `glue`, before `iam`)
  - [x] Add `allServices = append(allServices, guardduty.GuardDutyCalls...)` after `glue.GlueCalls...` and before `iam.IAMCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/guardduty/calls_test.go`
  - [x] Test `ListDetectors` Process: valid detectors with features, empty results, access denied error, nil fields, type assertion failure
  - [x] Test `GetFindings` Process: valid findings with severity/type, empty findings, error handling, nil fields, type assertion failure
  - [x] Test `ListFilters` Process: valid filters with action/description, empty filters, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following Organizations test pattern
  - [x] Access Process via `GuardDutyCalls[0].Process`, `GuardDutyCalls[1].Process`, `GuardDutyCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/guardduty/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := guardduty.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := guardduty.New(sess)
}
```

This prevents the race condition where concurrent workers sharing the same session could target wrong regions.

### GuardDuty is a REGIONAL Service

Unlike Organizations (global, us-east-1 only), GuardDuty is **regional** — each region has independent detectors, findings, and filters. Iterate `types.Regions` for all 3 API calls.

### GuardDuty SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/guardduty`

**API Methods:**

1. **ListDetectors:**
   - `svc.ListDetectorsWithContext(ctx, &guardduty.ListDetectorsInput{})` → `*guardduty.ListDetectorsOutput`
   - `.DetectorIds` → `[]*string` (detector ID strings)
   - Paginated via `NextToken *string`
   - Most accounts have 0 or 1 detector per region

2. **GetDetector:**
   - `svc.GetDetectorWithContext(ctx, &guardduty.GetDetectorInput{DetectorId: aws.String(id)})` → `*guardduty.GetDetectorOutput`
   - Output fields: `Status *string` (ENABLED/DISABLED), `ServiceRole *string` (IAM role ARN), `CreatedAt *string` (timestamp), `FindingPublishingFrequency *string`, `Features []*DetectorFeatureConfigurationResult`
   - `DetectorFeatureConfigurationResult` fields: `Name *string` (e.g., FLOW_LOGS, CLOUD_TRAIL, S3_DATA_EVENTS, EKS_AUDIT_LOGS, EBS_MALWARE_PROTECTION), `Status *string` (ENABLED/DISABLED)
   - Not paginated (single-item fetch)

3. **ListFindings:**
   - `svc.ListFindingsWithContext(ctx, &guardduty.ListFindingsInput{DetectorId: aws.String(id), MaxResults: aws.Int64(50), SortCriteria: &guardduty.SortCriteria{AttributeName: aws.String("severity"), OrderBy: aws.String("DESC")}})` → `*guardduty.ListFindingsOutput`
   - `.FindingIds` → `[]*string` (finding ID strings only — must hydrate via GetFindings)
   - Paginated via `NextToken` — but we limit to first 50 per detector (no pagination needed)

4. **GetFindings:**
   - `svc.GetFindingsWithContext(ctx, &guardduty.GetFindingsInput{DetectorId: aws.String(id), FindingIds: findingIds})` → `*guardduty.GetFindingsOutput`
   - `.Findings` → `[]*guardduty.Finding`
   - Finding fields: `Id *string`, `Type *string` (e.g., "Recon:EC2/PortProbeUnprotectedPort"), `Title *string`, `Severity *float64` (0-8.9 scale: Low 1-3.9, Medium 4-6.9, High 7-8.9), `Description *string`, `Region *string`, `CreatedAt *string`, `UpdatedAt *string`, `Arn *string`
   - Not paginated — batch by IDs (max 50 per call, matches our ListFindings limit)

5. **ListFilters:**
   - `svc.ListFiltersWithContext(ctx, &guardduty.ListFiltersInput{DetectorId: aws.String(id)})` → `*guardduty.ListFiltersOutput`
   - `.FilterNames` → `[]*string` (filter name strings — must hydrate via GetFilter)
   - Paginated via `NextToken *string`

6. **GetFilter:**
   - `svc.GetFilterWithContext(ctx, &guardduty.GetFilterInput{DetectorId: aws.String(id), FilterName: aws.String(name)})` → `*guardduty.GetFilterOutput`
   - Output fields: `Name *string`, `Action *string` (NOOP=notify, ARCHIVE=suppress), `Description *string`, `Rank *int64`
   - Not paginated (single-item fetch)

**No new dependencies needed** — GuardDuty is part of `aws-sdk-go v1.44.266` already in go.mod.

### Multi-Step Call Patterns

All 3 entries require discovering detectors first, then querying per-detector. This follows the ECR multi-step pattern (discover repos → query per-repo).

**Entry 1 (ListDetectors):** ListDetectors → GetDetector per detector
**Entry 2 (GetFindings):** ListDetectors → ListFindings per detector → GetFindings batch per detector
**Entry 3 (ListFilters):** ListDetectors → ListFilters per detector → GetFilter per filter name

**Key:** Each entry independently discovers detectors. Don't share state between entries.

### Local Struct Definitions

```go
type guarddutyDetector struct {
    DetectorId  string
    Status      string
    ServiceRole string
    CreatedAt   string
    Region      string
    Features    string  // comma-separated enabled feature names
}

type guarddutyFinding struct {
    FindingId   string
    Type        string
    Title       string
    Severity    float64
    Description string
    Region      string
    DetectorId  string
}

type guarddutyFilter struct {
    FilterName  string
    Action      string
    Description string
    DetectorId  string
    Region      string
}
```

### Detector Features Extraction Pattern

```go
var enabledFeatures []string
if detector.Features != nil {
    for _, feature := range detector.Features {
        if feature.Name != nil && feature.Status != nil && *feature.Status == "ENABLED" {
            enabledFeatures = append(enabledFeatures, *feature.Name)
        }
    }
}
featuresStr := strings.Join(enabledFeatures, ", ")
```

Import `"strings"` for `strings.Join`.

### Findings Limitation Strategy

Limit to **50 findings per detector per region** (first page only, no pagination on ListFindings). This prevents overwhelming output while still showing the most severe findings (sorted by severity DESC). The `GetFindings` batch call handles up to 50 IDs naturally.

```go
listFindingsInput := &guardduty.ListFindingsInput{
    DetectorId: aws.String(detectorId),
    MaxResults: aws.Int64(50),
    SortCriteria: &guardduty.SortCriteria{
        AttributeName: aws.String("severity"),
        OrderBy:       aws.String("DESC"),
    },
}
```

### Variable & Naming Conventions

- **Package:** `guardduty` (directory: `cmd/awtest/services/guardduty/`)
- **Exported variable:** `GuardDutyCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"guardduty:ListDetectors"`, `"guardduty:GetFindings"`, `"guardduty:ListFilters"`
- **ScanResult.ServiceName:** `"GuardDuty"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"detector"`, `"finding"`, `"filter"` (lowercase singular)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)

### Registration Order in services.go

Insert alphabetically — `guardduty` comes after `glue`, before `iam`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/guardduty"

// In AllServices():
allServices = append(allServices, glue.GlueCalls...)
allServices = append(allServices, guardduty.GuardDutyCalls...)  // NEW
allServices = append(allServices, iam.IAMCalls...)
```

### Testing Pattern

Follow the Organizations test pattern — test Process() functions only with pre-built mock data:

```go
func TestListDetectorsProcess(t *testing.T) {
    process := GuardDutyCalls[0].Process
    // Table-driven tests with valid detectors, empty, errors, nil fields, type assertion failure
}

func TestGetFindingsProcess(t *testing.T) {
    process := GuardDutyCalls[1].Process
    // Test with findings with severity/type, empty findings, errors, nil fields, type assertion failure
}

func TestListFiltersProcess(t *testing.T) {
    process := GuardDutyCalls[2].Process
    // Test with valid filters, empty filters, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `guardduty/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `guardduty.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** fail the entire scan if one detector's GetFindings or GetFilter fails — continue to next
- **DO NOT** paginate ListFindings — limit to first 50 per detector (MaxResults=50, no NextToken loop)
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** share detector discovery results between the 3 AWSService entries — each discovers independently
- **DO NOT** use `sess.Copy()` — use config override in service constructor

### Previous Story Intelligence

**From Story 7.2 (Organizations — previous story in this epic):**
- Use config override pattern for region: `service.New(sess, &aws.Config{Region: ...})` instead of `sess.Config.Region` mutation (race condition fix)
- Use local structs for multi-step call results (e.g., `guarddutyDetector`, `guarddutyFinding`, `guarddutyFilter`)
- Pagination: always use NextToken loop for paginated API calls
- Per-resource errors: `continue`/`break` to next item, don't abort entire scan
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `GuardDutyCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "GuardDuty", MethodName: "guardduty:ListDetectors", Error: err, Timestamp: time.Now()}}`
- Include Arn in Details map where available
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests

**From Story 7.2 Code Review Findings:**
- [HIGH] Always use config override for region (race condition prevention) — already incorporated above
- [HIGH] Include all relevant fields in Details map (e.g., JoinedTimestamp was missed in 7.2)
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [LOW] Tests should cover nil fields comprehensively

**From Story 7.1 Code Review Findings:**
- [HIGH] Always add pagination from the start (NextToken loops on all paginated APIs)
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [MEDIUM] Include ARN/Arn in Details map
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `2023f44` — Add AWS Organizations enumeration with 3 API calls (Story 7.2)
- `2c0b4ab` — Add ECR container registry enumeration with 3 API calls (Story 7.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message

### FRs Covered

- **FR87:** System enumerates GuardDuty detectors, findings, and suppression filters

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/guardduty/
├── calls.go            # GuardDuty service implementation (3 AWSService entries)
└── calls_test.go       # Process() tests for all 3 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go        # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go       # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/organizations/calls.go # Reference implementation (same epic, config override pattern)
go.mod                           # AWS SDK already includes GuardDuty package
```

### References

- [Source: epics-phase2.md#Story 2.3: GuardDuty Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR87] — GuardDuty enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/organizations/calls.go] — Reference implementation (same epic, config override pattern for region)
- [Source: cmd/awtest/services/organizations/calls_test.go] — Reference test pattern (table-driven Process-only tests)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (guardduty goes after glue, before iam)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes GuardDuty package)
- [Source: 7-2-aws-organizations-enumeration.md] — Previous story in same epic (patterns, code review findings: config override, error logging)
- [Source: 7-1-ecr-container-registry-enumeration.md] — First story in epic (patterns, code review findings: pagination, table-driven tests)
- [Source: code-review-findings.md] — Code review findings for Story 7.2 (race condition fix, error swallowing fix)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no debug issues.

### Completion Notes List

- Implemented GuardDuty service enumeration with 3 API calls: ListDetectors, GetFindings, ListFilters
- All 3 Call functions iterate `types.Regions` and use config override pattern `guardduty.New(sess, &aws.Config{Region: ...})` per 7.2 code review findings
- ListDetectors: paginated detector discovery + GetDetector hydration with enabled feature extraction
- GetFindings: detector discovery → ListFindings (max 50, severity DESC) → GetFindings batch hydration
- ListFilters: detector discovery → ListFilters (paginated) → GetFilter hydration per filter
- All Process functions handle errors via `utils.HandleAWSError`, type assertion failures, and nil-safe field access
- Per-resource errors use `continue` (don't abort scan) with error logging
- No sync primitives — concurrency-unaware per NFR57
- Registered in `services.go` alphabetically (after glue, before iam)
- Table-driven tests with `t.Run` subtests covering: valid data, empty results, access denied errors, nil/empty fields, type assertion failures
- All tests pass including race detector on full suite

### File List

- `cmd/awtest/services/guardduty/calls.go` (NEW) — GuardDuty service implementation with 3 AWSService entries
- `cmd/awtest/services/guardduty/calls_test.go` (NEW) — Process() tests for all 3 entries (15 test cases)
- `cmd/awtest/services/services.go` (MODIFIED) — Added guardduty import and registration in AllServices()

### Change Log

- 2026-03-10: Implemented GuardDuty enumeration with 3 API calls (detectors, findings, filters) following Organizations reference pattern with config override for region safety
