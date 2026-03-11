# Story 7.2: AWS Organizations Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate AWS Organizations accounts, OUs, and service control policies,
So that I can map the organizational account structure and identify cross-account access opportunities.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/organizations/` directory with `calls.go` implementing Organizations service enumeration.

2. **AC2:** Implement `organizations:ListAccounts` API call — Organizations is a **global service** (us-east-1 only, NO region iteration). Create Organizations client, call `ListAccountsWithContext` with pagination (NextToken loop), aggregate `[]*organizations.Account` results. Each account listed with account ID, name, email, status, and ARN.

3. **AC3:** Implement `organizations:ListOrganizationalUnits` API call — call `ListRootsWithContext` to discover root(s), then iteratively traverse the OU tree using `ListOrganizationalUnitsForParentWithContext` (BFS queue pattern — NOT recursive goroutines). Each OU listed with OU ID, name, ARN, and parent ID for hierarchy context.

4. **AC4:** Implement `organizations:ListPolicies` API call — call `ListPoliciesWithContext` with Filter `SERVICE_CONTROL_POLICY`, paginate with NextToken. Each SCP listed with policy ID, name, and ARN.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without Organizations access (non-management account), Organizations is skipped silently (access denied handling via existing error classification in safeScan). Handle `AWSOrganizationsNotInUseException` gracefully — skip, not an error.

7. **AC7:** Register Organizations service in `services/services.go` `AllServices()` function in alphabetical order (after `lambda`, before `rds`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid accounts/OUs/policies, empty results, access denied errors, nil field handling, `AWSOrganizationsNotInUseException` handling.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/organizations/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `organizations:ListAccounts` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/organizations/`
  - [x] Create `calls.go` with `package organizations`
  - [x] Define `var OrganizationsCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"organizations:ListAccounts"`
  - [x] Call: set region to `us-east-1` (global service — NO region loop), create `organizations.New(sess)`, paginate `ListAccountsWithContext` via NextToken loop, aggregate `[]*organizations.Account`
  - [x] Process: handle error → `utils.HandleAWSError`, type-assert `[]*organizations.Account`, extract `AccountId`, `Name`, `Email`, `Status`, `Arn`, `JoinedTimestamp` with nil checks, build `ScanResult` with ServiceName=`"Organizations"`, ResourceType=`"account"`
  - [x] `utils.PrintResult` format: `"Organizations Account: %s (ID: %s, Status: %s)"` with `utils.ColorizeItem(accountName)`

- [x] Task 2: Implement `organizations:ListOrganizationalUnits` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"organizations:ListOrganizationalUnits"`
  - [x] Call: set region to `us-east-1`, call `ListRootsWithContext` (paginated) to discover root(s), then BFS traversal — queue root IDs, for each parent call `ListOrganizationalUnitsForParentWithContext` (paginated), enqueue discovered OU IDs for deeper traversal
  - [x] Define local struct `orgOU` with fields: `OUId`, `OUName`, `OUArn`, `ParentId` to capture hierarchy context
  - [x] Process: type-assert `[]orgOU`, build `ScanResult` with ServiceName=`"Organizations"`, ResourceType=`"organizational-unit"`, ResourceName=ouName
  - [x] `utils.PrintResult` format: `"Organizations OU: %s (ID: %s, Parent: %s)"` with `utils.ColorizeItem(ouName)`

- [x] Task 3: Implement `organizations:ListPolicies` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"organizations:ListPolicies"`
  - [x] Call: set region to `us-east-1`, call `ListPoliciesWithContext` with `Filter: aws.String("SERVICE_CONTROL_POLICY")`, paginate with NextToken
  - [x] Process: type-assert `[]*organizations.PolicySummary`, build `ScanResult` with ServiceName=`"Organizations"`, ResourceType=`"service-control-policy"`, ResourceName=policyName
  - [x] `utils.PrintResult` format: `"Organizations SCP: %s (ID: %s)"` with `utils.ColorizeItem(policyName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/organizations"` to `services/services.go` (alphabetical in imports: after `lambda`, before `rds`)
  - [x] Add `allServices = append(allServices, organizations.OrganizationsCalls...)` after `lambda.LambdaCalls...` and before `rds.RDSCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/organizations/calls_test.go`
  - [x] Test `ListAccounts` Process: valid accounts, empty results, access denied error, nil fields
  - [x] Test `ListOrganizationalUnits` Process: valid OUs with hierarchy, empty OUs, error handling
  - [x] Test `ListPolicies` Process: valid SCPs, empty policies, error handling
  - [x] Use table-driven tests with `t.Run` subtests following ECR test pattern
  - [x] Access Process via `OrganizationsCalls[0].Process`, `OrganizationsCalls[1].Process`, `OrganizationsCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/organizations/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: AWS Organizations is a GLOBAL Service — NO Region Iteration

Unlike most services (ECR, ACM, etc.) that iterate over `types.Regions`, **AWS Organizations is a global service** that only operates in `us-east-1`. The Call functions must:

```go
// CORRECT: Global service — set region once, no loop
sess.Config.Region = aws.String("us-east-1")
svc := organizations.New(sess)
```

**DO NOT** iterate over `types.Regions` — Organizations APIs will return the same data regardless of region, and iterating would produce duplicate results.

**Other global services in the codebase for reference:** `iam`, `sts`, `route53`, `cloudfront`, `s3` (ListBuckets) — check these for the global service pattern if needed.

### CRITICAL: Management Account Only

Organizations APIs only work from the **management account** (or delegated administrator). Non-management accounts will receive:
- `AWSOrganizationsNotInUseException` — account is not part of an organization or not the management account
- `AccessDeniedException` — insufficient permissions

Both should be handled gracefully by the existing safeScan error classification. No special handling needed in the service code — just let errors propagate and safeScan will skip silently.

### Organizations SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/organizations`

**API Methods:**

1. **ListAccounts:**
   - `svc.ListAccountsWithContext(ctx, &organizations.ListAccountsInput{})` → `*organizations.ListAccountsOutput`
   - `.Accounts` → `[]*organizations.Account`
   - Account fields: `Id *string`, `Name *string`, `Email *string`, `Status *string` (ACTIVE|SUSPENDED|PENDING_CLOSURE), `Arn *string`, `JoinedTimestamp *time.Time`, `JoinedMethod *string`
   - Paginated via `NextToken *string`

2. **ListRoots:**
   - `svc.ListRootsWithContext(ctx, &organizations.ListRootsInput{})` → `*organizations.ListRootsOutput`
   - `.Roots` → `[]*organizations.Root`
   - Root fields: `Id *string`, `Name *string`, `Arn *string`
   - Paginated via `NextToken *string`
   - Typically returns a single root, but the API supports pagination

3. **ListOrganizationalUnitsForParent:**
   - `svc.ListOrganizationalUnitsForParentWithContext(ctx, &organizations.ListOrganizationalUnitsForParentInput{ParentId: aws.String(parentId)})` → `*organizations.ListOrganizationalUnitsForParentOutput`
   - `.OrganizationalUnits` → `[]*organizations.OrganizationalUnit`
   - OU fields: `Id *string`, `Name *string`, `Arn *string`
   - Paginated via `NextToken *string`
   - Must be called with each discovered OU ID to traverse deeper levels

4. **ListPolicies:**
   - `svc.ListPoliciesWithContext(ctx, &organizations.ListPoliciesInput{Filter: aws.String("SERVICE_CONTROL_POLICY")})` → `*organizations.ListPoliciesOutput`
   - `.Policies` → `[]*organizations.PolicySummary`
   - PolicySummary fields: `Id *string`, `Name *string`, `Arn *string`, `Description *string`, `Type *string`, `AwsManaged *bool`
   - Paginated via `NextToken *string`

**No new dependencies needed** — Organizations is part of `aws-sdk-go v1.44.266` already in go.mod.

### OU Tree Traversal Pattern (BFS Queue)

The OU hierarchy requires traversal from root(s) down. Use an iterative BFS approach (queue) — **DO NOT** use recursion or goroutines:

```go
type orgOU struct {
    OUId     string
    OUName   string
    OUArn    string
    ParentId string
}

// BFS traversal
var allOUs []orgOU
queue := []string{} // parent IDs to visit

// Step 1: Get roots
rootsInput := &organizations.ListRootsInput{}
for {
    rootsOutput, err := svc.ListRootsWithContext(ctx, rootsInput)
    if err != nil {
        return nil, err
    }
    for _, root := range rootsOutput.Roots {
        if root.Id != nil {
            queue = append(queue, *root.Id)
        }
    }
    if rootsOutput.NextToken == nil {
        break
    }
    rootsInput.NextToken = rootsOutput.NextToken
}

// Step 2: BFS — process queue
for len(queue) > 0 {
    parentId := queue[0]
    queue = queue[1:]

    ouInput := &organizations.ListOrganizationalUnitsForParentInput{
        ParentId: aws.String(parentId),
    }
    for {
        ouOutput, err := svc.ListOrganizationalUnitsForParentWithContext(ctx, ouInput)
        if err != nil {
            break // Skip this parent on error, continue with remaining queue
        }
        for _, ou := range ouOutput.OrganizationalUnits {
            ouId := ""
            if ou.Id != nil {
                ouId = *ou.Id
                queue = append(queue, ouId) // Enqueue for deeper traversal
            }
            // ... build orgOU struct
            allOUs = append(allOUs, orgOU{...})
        }
        if ouOutput.NextToken == nil {
            break
        }
        ouInput.NextToken = ouOutput.NextToken
    }
}
```

**Key:** Per-parent errors should be handled gracefully — `break` from the inner pagination loop and continue with the next queued parent. Don't abort the entire scan.

### Variable & Naming Conventions

- **Package:** `organizations` (directory: `cmd/awtest/services/organizations/`)
- **Exported variable:** `OrganizationsCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"organizations:ListAccounts"`, `"organizations:ListOrganizationalUnits"`, `"organizations:ListPolicies"`
- **ScanResult.ServiceName:** `"Organizations"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"account"`, `"organizational-unit"`, `"service-control-policy"` (lowercase singular)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)

### Registration Order in services.go

Insert alphabetically — `organizations` comes after `lambda` (via kms → lambda), before `rds`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/organizations"

// In AllServices():
allServices = append(allServices, lambda.LambdaCalls...)
allServices = append(allServices, organizations.OrganizationsCalls...)  // NEW
allServices = append(allServices, rds.RDSCalls...)
```

### Testing Pattern

Follow the ECR test pattern — test Process() functions only with pre-built mock data:

```go
func TestListAccountsProcess(t *testing.T) {
    process := OrganizationsCalls[0].Process
    // Table-driven tests with valid accounts, empty, errors, nil fields
}

func TestListOrganizationalUnitsProcess(t *testing.T) {
    process := OrganizationsCalls[1].Process
    // Test with OUs with hierarchy info, empty results, errors
}

func TestListPoliciesProcess(t *testing.T) {
    process := OrganizationsCalls[2].Process
    // Test with valid SCPs, empty policies, errors
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `organizations/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** iterate over `types.Regions` — Organizations is a global service (us-east-1 only)
- **DO NOT** use recursive functions for OU traversal — use iterative BFS queue pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use `sess.Copy()` — use `sess.Config.Region = aws.String("us-east-1")` pattern
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** treat `AWSOrganizationsNotInUseException` as a fatal error — let it propagate to safeScan for silent skip

### Previous Story Intelligence

**From Story 7.1 (ECR — previous story in this epic):**
- Use local structs for multi-step call results (e.g., `orgOU` struct for OU hierarchy data)
- Pagination: always use NextToken loop for all API calls
- Per-resource errors: `continue`/`break` to next item, don't abort entire scan
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `OrganizationsCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "Organizations", MethodName: "organizations:ListAccounts", Error: err, Timestamp: time.Now()}}`
- Tests resolved review findings: ensure table-driven with `t.Run`, add type assertion failure handling, test nil/empty edge cases
- RepositoryArn-style pattern: include `Arn` in Details map for enriched output

**From Story 7.1 Code Review Findings:**
- Always add pagination from the start (was a [High] review finding in 7.1)
- Use table-driven tests with `t.Run` subtests from the start (was a [Medium] review finding)
- Include ARN in Details map (was a [Medium] review finding)
- Handle type assertion failures (was a [Low] review finding)
- These patterns are already incorporated into this story's tasks above

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `2c0b4ab` — Add ECR container registry enumeration with 3 API calls (Story 7.1)
- Files created: `services/ecr/calls.go`, `services/ecr/calls_test.go`
- Files modified: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message

### FRs Covered

- **FR86:** System enumerates AWS Organizations accounts, organizational units, and service control policies

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/organizations/
├── calls.go            # Organizations service implementation (3 AWSService entries)
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
cmd/awtest/services/ecr/calls.go # Reference implementation (same epic)
go.mod                           # AWS SDK already includes Organizations package
```

### References

- [Source: epics-phase2.md#Story 2.2: AWS Organizations Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR86] — Organizations enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/ecr/calls.go] — Reference implementation (same epic, multi-step call pattern)
- [Source: cmd/awtest/services/ecr/calls_test.go] — Reference test pattern (table-driven Process-only tests)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes Organizations package)
- [Source: 7-1-ecr-container-registry-enumeration.md] — Previous story in same epic (patterns, review findings)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

None — clean implementation with no debugging required.

### Completion Notes List

- Implemented 3 Organizations API calls (ListAccounts, ListOrganizationalUnits, ListPolicies) as global service (us-east-1 only, no region iteration)
- Used iterative BFS queue pattern for OU hierarchy traversal per story specification
- All Process() functions include error handling via `utils.HandleAWSError`, type assertion failure handling, and nil-safe pointer dereferencing
- Details maps include ARN fields per code review learnings from Story 7.1
- Registered service in AllServices() alphabetically between lambda and rds
- 15 table-driven tests covering valid data, empty results, access denied errors, nil fields, and type assertion failures
- No sync primitives imported — concurrency-unaware per NFR57
- All WithContext API variants used for timeout/cancellation support
- `go build`, `go test`, `go vet`, and `go test -race` all pass clean with zero regressions
- **Code Review Fixes (2026-03-10):**
  - [HIGH] Added missing `JoinedTimestamp` to ListAccounts Details map (formatted as RFC3339 string)
  - [HIGH] Fixed race condition: replaced `sess.Config.Region` mutation with config override in service constructor (`organizations.New(sess, globalRegionConfig)`) — session is no longer mutated in place
  - [MEDIUM] Added `utils.HandleAWSError` call in OU traversal before breaking on per-parent errors — errors are now logged instead of silently swallowed
  - [LOW] ListPolicies scope limited to SERVICE_CONTROL_POLICY — by design per AC4, no change needed

### File List

- `cmd/awtest/services/organizations/calls.go` (NEW) — Organizations service with 3 AWSService entries
- `cmd/awtest/services/organizations/calls_test.go` (NEW) — 15 table-driven Process() tests
- `cmd/awtest/services/services.go` (MODIFIED) — Added organizations import and registration in AllServices()
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (MODIFIED) — Story status updated
- `_bmad-output/implementation-artifacts/7-2-aws-organizations-enumeration.md` (MODIFIED) — Story file updated

### Change Log

- **2026-03-10:** Implemented AWS Organizations enumeration with 3 API calls (ListAccounts, ListOrganizationalUnits, ListPolicies). Global service pattern (us-east-1 only). BFS queue OU traversal. 15 unit tests. Registered in AllServices(). All validation passes.
- **2026-03-10:** Addressed code review findings — 3 items resolved: added JoinedTimestamp, fixed session race condition, added OU traversal error logging. 1 item by design (SCP-only scope per AC4).
