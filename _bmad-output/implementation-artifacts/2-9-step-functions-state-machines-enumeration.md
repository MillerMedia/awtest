# Story 2.9: Step Functions State Machines Enumeration

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Step Functions state machines**,
so that **I can discover workflow orchestration accessible with the credentials, revealing automated processes and integration patterns**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/stepfunctions/` directory with `calls.go`
2. **AC2:** Implement Step Functions state machine enumeration using AWS SDK v1.44.266 SFN client (`github.com/aws/aws-sdk-go/service/sfn`)
3. **AC3:** Implement AWSService interface: `Name="states:ListStateMachines"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates SFN client per region using `sess.Copy()`, calls `ListStateMachines` with `NextToken`-based pagination -- aggregates all `*sfn.StateMachineListItem` results
5. **AC5:** `Process()` displays each state machine: Name, StateMachineArn, Type (STANDARD/EXPRESS), CreationDate
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results -- if no state machines found after all regions, call `utils.PrintAccessGranted(debug, "states:ListStateMachines", "Step Functions state machines")` and return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `sqs`, before `transcribe`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: STANDARD state machine, EXPRESS state machine, multiple state machines, empty results, access denied, nil field handling
10. **AC10:** Package naming: `stepfunctions` (lowercase, single word)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/stepfunctions/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR24 requirement fulfilled: System enumerates Step Functions state machines

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/stepfunctions/`
  - [x] Create `calls.go` with package `stepfunctions`
  - [x] Define `var StepFunctionsCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `sfn.New(regionSess)` per region using `sess.Copy(&aws.Config{Region: aws.String(region)})`, call `svc.ListStateMachines(&sfn.ListStateMachinesInput{})` with NextToken-based pagination
  - [x] Use resilient per-region error handling (continue to next region on error, `anyRegionSucceeded` + `lastErr` pattern)
  - [x] Return aggregated `[]*sfn.StateMachineListItem` from Call(), or nil on complete failure

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "states:ListStateMachines", err)`, return error ScanResult
  - [x] Type-assert output to `[]*sfn.StateMachineListItem`
  - [x] Handle type assertion failure (like Redshift/ElastiCache pattern)
  - [x] If empty slice and no error: call `utils.PrintAccessGranted(debug, "states:ListStateMachines", "Step Functions state machines")`, return empty results
  - [x] For each state machine, extract: `Name` (`*string`), `StateMachineArn` (`*string`), `Type` (`*string`), `CreationDate` (`*time.Time`) -- all with nil checks
  - [x] Build `types.ScanResult` entries with: ServiceName="Step Functions", MethodName="states:ListStateMachines", ResourceType="state-machine"
  - [x] Call `utils.PrintResult()` with formatted output
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/stepfunctions"` to `services/services.go`
  - [x] Add `allServices = append(allServices, stepfunctions.StepFunctionsCalls...)` after `sqs.SQSCalls...` and before `transcribe.TranscribeCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/stepfunctions/calls_test.go`
  - [x] Follow Redshift/ElastiCache test pattern: table-driven Process()-only tests with pre-built mock data
  - [x] Test cases: STANDARD state machine (all fields), EXPRESS state machine, multiple state machines, empty results, access denied error, nil field handling

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/stepfunctions/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### Step Functions Uses Single ListStateMachines API -- Simple Pattern Like ElastiCache/Redshift

`ListStateMachines` is a single API call that returns all state machines. This follows the **same simple pattern as Redshift** (Story 2.8) and **ElastiCache** (Story 2.6), NOT the multi-API-call pattern used by Fargate (Story 2.7) or EKS (Story 2.5).

**Package:** `github.com/aws/aws-sdk-go/service/sfn` -- available in `aws-sdk-go v1.44.266` in go.mod. No new dependencies needed.

### CRITICAL: Epics File Says "Status" Field -- ListStateMachines Does NOT Return Status

The epics file mentions displaying `Status` for state machines, but `ListStateMachines` returns `StateMachineListItem` which does NOT include a `Status` field. Status is only available from `DescribeStateMachine` (per-machine call). **Do NOT add DescribeStateMachine calls** -- follow the single-API pattern. Only display the fields available from `ListStateMachines`: Name, StateMachineArn, Type, CreationDate.

### CRITICAL: Use sess.Copy() for Region Iteration

Story 2.3 code review identified that mutating `sess.Config.Region` directly is unsafe. **YOU MUST USE `sess.Copy()`** for safe session handling:

```go
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := sfn.New(regionSess)
    // ...
}
```

### AWS SFN SDK Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/sfn`

**API Call:**
- `svc.ListStateMachines(&sfn.ListStateMachinesInput{})` -> `*sfn.ListStateMachinesOutput`
- `Output.StateMachines` -> `[]*sfn.StateMachineListItem`
- Pagination via `NextToken` (string, NOT Marker)

**StateMachineListItem fields (from `*sfn.StateMachineListItem`):**
- `Name` -- `*string` -- state machine name (e.g., "my-workflow")
- `StateMachineArn` -- `*string` -- full ARN (e.g., "arn:aws:states:us-east-1:123456789:stateMachine:my-workflow")
- `Type` -- `*string` -- "STANDARD" or "EXPRESS"
- `CreationDate` -- `*time.Time` -- when the state machine was created

**IMPORTANT:** The `Type` field uses string constants, not an enum. Expected values are `"STANDARD"` and `"EXPRESS"`.

**Pagination Pattern:**
```go
input := &sfn.ListStateMachinesInput{}
for {
    output, err := svc.ListStateMachines(input)
    if err != nil {
        lastErr = err
        regionFailed = true
        break
    }
    allStateMachines = append(allStateMachines, output.StateMachines...)
    if output.NextToken == nil {
        break
    }
    input.NextToken = output.NextToken
}
```

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `stepfunctions` |
| Package variable | `StepFunctionsCalls` |
| AWSService.Name | `"states:ListStateMachines"` |
| ScanResult.ServiceName | `"Step Functions"` |
| ScanResult.MethodName | `"states:ListStateMachines"` |
| ScanResult.ResourceType | `"state-machine"` |

**Note on AWSService.Name:** The AWS IAM service prefix for Step Functions is `states` (not `sfn` or `stepfunctions`). This matches the IAM policy action format: `states:ListStateMachines`.

### Registration Order in services.go

Insert after `sqs` and before `transcribe` (alphabetical):

```go
allServices = append(allServices, sqs.SQSCalls...)
allServices = append(allServices, stepfunctions.StepFunctionsCalls...)  // NEW
allServices = append(allServices, transcribe.TranscribeCalls...)
```

Import alphabetically after `sqs`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/sqs"
"github.com/MillerMedia/awtest/cmd/awtest/services/stepfunctions"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/transcribe"
```

### Process() Output Format

```go
utils.PrintResult(debug, "", "states:ListStateMachines",
    fmt.Sprintf("Found Step Functions State Machine: %s (ARN: %s, Type: %s, Created: %s)",
        utils.ColorizeItem(name), arn, smType, creationDate), nil)
```

### Empty Results Handling

```go
if len(stateMachines) == 0 {
    utils.PrintAccessGranted(debug, "states:ListStateMachines", "Step Functions state machines")
    return results
}
```

### CreationDate Formatting

`CreationDate` is `*time.Time`. Format it for display:

```go
creationDate := ""
if sm.CreationDate != nil {
    creationDate = sm.CreationDate.Format("2006-01-02 15:04:05")
}
```

Store the raw `*time.Time` (or formatted string) in the Details map for output formatters.

### Reference Implementation Pattern

Follow Redshift (`cmd/awtest/services/redshift/calls.go`) as the primary reference -- same single-API pattern with resilient per-region error handling.

Key differences from Redshift:
1. Uses `sfn` package instead of `redshift`
2. Uses `NextToken` for pagination (not `Marker`)
3. Fewer fields to extract (4 fields vs 8 for Redshift)
4. Has `*time.Time` field (`CreationDate`) -- requires time formatting for display
5. No nested struct fields (simpler than Redshift's `Endpoint`)
6. Simpler overall -- no `*bool` or `*int64` fields, mostly `*string`

### Testing Pattern (from Redshift Story 2.8)

Create table-driven Process()-only tests with pre-built mock data. No AWS SDK mocking needed.

Test cases:
1. **STANDARD state machine with all fields** -- all fields populated, verify all ScanResult fields and Details map
2. **EXPRESS state machine** -- verify Type="EXPRESS" is captured correctly
3. **Multiple state machines** -- verify correct count and resource names
4. **Empty results** -- verify PrintAccessGranted behavior and empty results returned
5. **Access denied** -- verify error ScanResult returned with correct ServiceName/MethodName
6. **Nil field handling** -- verify nil Name, nil StateMachineArn, nil Type, nil CreationDate handled gracefully

**IMPORTANT for tests:** When creating mock `sfn.StateMachineListItem`:
```go
&sfn.StateMachineListItem{
    Name:            aws.String("my-workflow"),
    StateMachineArn: aws.String("arn:aws:states:us-east-1:123456789:stateMachine:my-workflow"),
    Type:            aws.String("STANDARD"),
    CreationDate:    &time.Time{}, // or a specific time
}
```

**Note:** You'll need to import `time` in the test file for `CreationDate`.

### Edge Cases

1. **No state machines in any region** -- ListStateMachines returns empty, Process() calls PrintAccessGranted
2. **Access denied in all regions** -- Call() returns nil + error, Process() handles error
3. **Access denied in some regions** -- Call() continues to next region (resilient pattern), returns partial results
4. **Nil CreationDate** -- defensive nil check, default to empty string
5. **Nil Type** -- defensive nil check, default to empty string
6. **Pagination across many state machines** -- handle NextToken for accounts with many state machines

### Architecture Compliance

- **Package:** `stepfunctions` in `cmd/awtest/services/stepfunctions/` -- MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) -- MUST FOLLOW
- **Variable:** `StepFunctionsCalls` exported slice -- MUST FOLLOW
- **Type:** `[]types.AWSService` -- MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` -- MUST FOLLOW (all Epic 2 stories use DefaultModuleName)
- **Session handling:** `sess.Copy(&aws.Config{Region: aws.String(region)})` -- MUST FOLLOW (Story 2.3 code review fix)
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` -- MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` -- MUST FOLLOW
- **Nil checks:** Always check `*string`, `*time.Time` before dereferencing -- MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) -- MUST FOLLOW
- **SDK version:** AWS SDK Go v1.44.266 -- MUST FOLLOW (do NOT use SDK v2)

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/stepfunctions/
+-- calls.go            # NEW: Step Functions state machines service implementation
+-- calls_test.go       # NEW: Process() tests
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                      # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                     # PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem
cmd/awtest/services/redshift/calls.go          # Story 2.8 reference (sess.Copy + resilient + NextToken pagination) -- PRIMARY REFERENCE
cmd/awtest/services/redshift/calls_test.go     # Story 2.8 test reference (table-driven Process()-only)
cmd/awtest/services/elasticache/calls.go       # Story 2.6 reference (sess.Copy + resilient + Marker pagination)
cmd/awtest/services/services.go                # AllServices() registration point
go.mod                                         # AWS SDK already included (sfn package available)
```

### Previous Story Intelligence (Story 2.8 - Redshift)

**Key learnings from Story 2.8 (Redshift):**
- **sess.Copy() is mandatory** -- continued from Story 2.3 fix
- **Resilient per-region errors** with `anyRegionSucceeded` + `lastErr` tracking pattern
- **Pagination included from the start** -- avoid code review rework
- Table-driven Process()-only tests are the standard
- Type assertion failure handling included from the start
- All display fields must appear in BOTH PrintResult AND Details map
- `ScanResult.Timestamp = time.Now()` is required on every result
- Empty results handled with `utils.PrintAccessGranted`
- Endpoint formatting fix: handle nested structs carefully (not applicable here -- no nested structs)
- Cross-cutting review findings (MaxRecords, silent region failures) deferred -- do NOT add these

### Git Intelligence

**Recent commits (Epic 2 context):**
- `ebf7392` Mark Story 2.8 as done
- `40c71c3` Add Redshift clusters service enumeration (Story 2.8)
- `f814fab` Fix false "Access granted" when all regions return access denied
- `f46641a` Add Fargate tasks service enumeration (Story 2.7)
- `12e6771` Add ElastiCache service enumeration (Story 2.6)

**Key insight from f814fab:** The `anyRegionSucceeded` + `lastErr` pattern in Call() is critical. If all regions return access denied, Call() must return the error (not nil) so Process() can properly report it. This was a bug fix applied after Story 2.7.

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `stepfunctions` follows convention (lowercase, single word, matches directory)
- Single `calls.go` file per service -- matches all 38+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/stepfunctions`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.9: Step Functions State Machines Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR7-31 Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#stepfunctions/calls.go in file structure]
- [Source: _bmad-output/implementation-artifacts/2-8-redshift-clusters-service-enumeration.md -- previous story learnings]
- [Source: cmd/awtest/services/redshift/calls.go -- PRIMARY reference (sess.Copy + resilient + pagination)]
- [Source: cmd/awtest/services/redshift/calls_test.go -- test reference (table-driven Process()-only)]
- [Source: cmd/awtest/services/services.go -- AllServices() registration point]
- [Source: cmd/awtest/types/types.go -- AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go -- PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem]
- [Source: go.mod -- aws-sdk-go v1.44.266 (includes sfn package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Implemented Step Functions state machine enumeration following the Redshift (Story 2.8) pattern
- Call() uses sess.Copy() for safe per-region session handling with NextToken-based pagination
- Process() extracts Name, StateMachineArn, Type, CreationDate with nil checks on all fields
- CreationDate formatted as "2006-01-02 15:04:05" for display, stored as string in Details map
- Registered in AllServices() alphabetically between sqs and sts
- 6 table-driven Process()-only tests: STANDARD, EXPRESS, multiple, empty, access denied, nil fields
- All tests pass, go build succeeds, go vet clean, full regression suite green

### Change Log

- 2026-03-05: Implemented Story 2.9 - Step Functions state machine enumeration

### File List

- cmd/awtest/services/stepfunctions/calls.go (NEW)
- cmd/awtest/services/stepfunctions/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED)
