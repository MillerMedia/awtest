# Story 2.10: Systems Manager (SSM) Parameters Enumeration

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Systems Manager parameters**,
so that **I can discover configuration parameters and secrets accessible with the credentials, which often contain database credentials, API keys, and infrastructure configs**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/systemsmanager/` directory with `calls.go`
2. **AC2:** Implement SSM parameter enumeration using AWS SDK v1.44.266 SSM client (`github.com/aws/aws-sdk-go/service/ssm`)
3. **AC3:** Implement AWSService interface: `Name="ssm:DescribeParameters"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates SSM client per region using `sess.Copy()`, calls `DescribeParameters` with `NextToken`-based pagination -- aggregates all `*ssm.ParameterMetadata` results
5. **AC5:** `Process()` displays each parameter: Name, Type (String/StringList/SecureString), Description, LastModifiedDate, Version -- DO NOT retrieve parameter values (read-only enumeration per NFR7)
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results -- if no parameters found after all regions, call `utils.PrintAccessGranted(debug, "ssm:DescribeParameters", "SSM parameters")` and return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `stepfunctions`, before `transcribe`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: String parameter, SecureString parameter, multiple parameters, empty results, access denied, nil field handling
10. **AC10:** Package naming: `systemsmanager` (lowercase, single word, matches directory)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/systemsmanager/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR31 requirement fulfilled: System enumerates Systems Manager parameters

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/systemsmanager/`
  - [x] Create `calls.go` with package `systemsmanager`
  - [x] Define `var SystemsManagerCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `ssm.New(regionSess)` per region using `sess.Copy(&aws.Config{Region: aws.String(region)})`, call `svc.DescribeParameters(&ssm.DescribeParametersInput{})` with NextToken-based pagination
  - [x] Use resilient per-region error handling (continue to next region on error, `anyRegionSucceeded` + `lastErr` pattern)
  - [x] Return aggregated `[]*ssm.ParameterMetadata` from Call(), or nil on complete failure

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "ssm:DescribeParameters", err)`, return error ScanResult
  - [x] Type-assert output to `[]*ssm.ParameterMetadata`
  - [x] Handle type assertion failure (like Redshift pattern)
  - [x] If empty slice and no error: call `utils.PrintAccessGranted(debug, "ssm:DescribeParameters", "SSM parameters")`, return empty results
  - [x] For each parameter, extract: `Name` (`*string`), `Type` (`*string`), `Description` (`*string`), `LastModifiedDate` (`*time.Time`), `Version` (`*int64`) -- all with nil checks
  - [x] Build `types.ScanResult` entries with: ServiceName="Systems Manager", MethodName="ssm:DescribeParameters", ResourceType="parameter"
  - [x] Call `utils.PrintResult()` with formatted output
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/systemsmanager"` to `services/services.go`
  - [x] Add `allServices = append(allServices, systemsmanager.SystemsManagerCalls...)` after `stepfunctions.StepFunctionsCalls...` and before `transcribe.TranscribeCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/systemsmanager/calls_test.go`
  - [x] Follow Redshift test pattern: table-driven Process()-only tests with pre-built mock data
  - [x] Test cases: String parameter (all fields), SecureString parameter, multiple parameters, empty results, access denied error, nil field handling

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/systemsmanager/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### SSM Uses DescribeParameters API -- Simple Pattern Like Redshift/Step Functions

`DescribeParameters` is a single API call that returns parameter metadata without values. This follows the **same simple pattern as Redshift** (Story 2.8) and **Step Functions** (Story 2.9) -- single API, multi-region iteration, pagination aggregation.

**Package:** `github.com/aws/aws-sdk-go/service/ssm` -- available in `aws-sdk-go v1.44.266` in go.mod. No new dependencies needed.

### CRITICAL: Do NOT Retrieve Parameter Values

NFR7 compliance: This is read-only enumeration. `DescribeParameters` returns metadata only (name, type, description, etc.) -- it does NOT return parameter values. Do NOT call `GetParameter`, `GetParameters`, or `GetParametersByPath`. Only enumerate metadata.

**SecureString parameters** will show `Type: "SecureString"` in the metadata -- this tells the security professional that secrets exist without exposing them.

### CRITICAL: Use sess.Copy() for Region Iteration

Story 2.3 code review identified that mutating `sess.Config.Region` directly is unsafe. **YOU MUST USE `sess.Copy()`** for safe session handling:

```go
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := ssm.New(regionSess)
    // ...
}
```

### AWS SSM SDK Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/ssm`

**API Call:**
- `svc.DescribeParameters(&ssm.DescribeParametersInput{})` -> `*ssm.DescribeParametersOutput`
- `Output.Parameters` -> `[]*ssm.ParameterMetadata`
- Pagination via `NextToken` (string, NOT Marker)
- Optional: `MaxResults` (default 50, max 50) -- do NOT set explicitly, use SDK default

**ParameterMetadata fields (from `*ssm.ParameterMetadata`):**
- `Name` -- `*string` -- parameter name/path (e.g., "/app/config/db-password")
- `Type` -- `*string` -- "String", "StringList", or "SecureString"
- `Description` -- `*string` -- parameter description (may be nil/empty)
- `LastModifiedDate` -- `*time.Time` -- when the parameter was last modified
- `Version` -- `*int64` -- parameter version number

**Additional fields available but LOWER PRIORITY for display:**
- `ARN` -- `*string` -- full ARN
- `DataType` -- `*string` -- "text" or "aws:ec2:image"
- `KeyId` -- `*string` -- KMS key ID (for SecureString)
- `Tier` -- `*string` -- "Standard", "Advanced", or "Intelligent-Tiering"

**Pagination Pattern:**
```go
input := &ssm.DescribeParametersInput{}
for {
    output, err := svc.DescribeParameters(input)
    if err != nil {
        lastErr = err
        regionFailed = true
        break
    }
    allParameters = append(allParameters, output.Parameters...)
    if output.NextToken == nil {
        break
    }
    input.NextToken = output.NextToken
}
```

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `systemsmanager` |
| Package variable | `SystemsManagerCalls` |
| AWSService.Name | `"ssm:DescribeParameters"` |
| ScanResult.ServiceName | `"Systems Manager"` |
| ScanResult.MethodName | `"ssm:DescribeParameters"` |
| ScanResult.ResourceType | `"parameter"` |

**Note on AWSService.Name:** The AWS IAM service prefix for Systems Manager is `ssm` (not `systemsmanager`). This matches the IAM policy action format: `ssm:DescribeParameters`.

### Registration Order in services.go

Insert after `stepfunctions` and before `transcribe` (alphabetical by package name):

```go
allServices = append(allServices, stepfunctions.StepFunctionsCalls...)
allServices = append(allServices, systemsmanager.SystemsManagerCalls...)  // NEW
allServices = append(allServices, transcribe.TranscribeCalls...)
```

Import alphabetically after `stepfunctions` and before `sts`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/stepfunctions"
"github.com/MillerMedia/awtest/cmd/awtest/services/systemsmanager"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/sts"
"github.com/MillerMedia/awtest/cmd/awtest/services/transcribe"
```

### Process() Output Format

```go
utils.PrintResult(debug, "", "ssm:DescribeParameters",
    fmt.Sprintf("Found SSM Parameter: %s (Type: %s, Description: %s, LastModified: %s, Version: %d)",
        utils.ColorizeItem(name), paramType, description, lastModified, version), nil)
```

### Empty Results Handling

```go
if len(parameters) == 0 {
    utils.PrintAccessGranted(debug, "ssm:DescribeParameters", "SSM parameters")
    return results
}
```

### LastModifiedDate Formatting

`LastModifiedDate` is `*time.Time`. Format it for display:

```go
lastModified := ""
if param.LastModifiedDate != nil {
    lastModified = param.LastModifiedDate.Format("2006-01-02 15:04:05")
}
```

Store the formatted string in the Details map for output formatters.

### Version Field Handling

`Version` is `*int64`. Handle nil check:

```go
var version int64
if param.Version != nil {
    version = *param.Version
}
```

### Reference Implementation Pattern

Follow Redshift (`cmd/awtest/services/redshift/calls.go`) as the primary reference -- same single-API pattern with resilient per-region error handling.

Key differences from Redshift:
1. Uses `ssm` package instead of `redshift`
2. Uses `NextToken` for pagination (not `Marker`)
3. Different fields: Name, Type, Description, LastModifiedDate, Version
4. Has `*time.Time` field (`LastModifiedDate`) -- requires time formatting (same as Step Functions CreationDate)
5. Has `*int64` field (`Version`) -- requires nil check and dereference
6. No nested struct fields (simpler than Redshift's `Endpoint`)

### Testing Pattern (from Redshift Story 2.8)

Create table-driven Process()-only tests with pre-built mock data. No AWS SDK mocking needed.

Test cases:
1. **String parameter with all fields** -- all fields populated, verify all ScanResult fields and Details map
2. **SecureString parameter** -- verify Type="SecureString" is captured correctly
3. **Multiple parameters** -- verify correct count and resource names
4. **Empty results** -- verify PrintAccessGranted behavior and empty results returned
5. **Access denied** -- verify error ScanResult returned with correct ServiceName/MethodName
6. **Nil field handling** -- verify nil Name, nil Type, nil Description, nil LastModifiedDate, nil Version handled gracefully

**IMPORTANT for tests:** When creating mock `ssm.ParameterMetadata`:
```go
&ssm.ParameterMetadata{
    Name:             aws.String("/app/config/db-host"),
    Type:             aws.String("String"),
    Description:      aws.String("Database hostname"),
    LastModifiedDate: &time.Time{}, // or a specific time
    Version:          aws.Int64(3),
}
```

**Note:** You'll need to import `time` in the test file for `LastModifiedDate`.

### Edge Cases

1. **No parameters in any region** -- DescribeParameters returns empty, Process() calls PrintAccessGranted
2. **Access denied in all regions** -- Call() returns nil + error, Process() handles error
3. **Access denied in some regions** -- Call() continues to next region (resilient pattern), returns partial results
4. **Nil LastModifiedDate** -- defensive nil check, default to empty string
5. **Nil Description** -- defensive nil check, default to empty string (common -- many params have no description)
6. **Nil Version** -- defensive nil check, default to 0
7. **Pagination across many parameters** -- handle NextToken for accounts with many parameters (SSM can have thousands)

### Architecture Compliance

- **Package:** `systemsmanager` in `cmd/awtest/services/systemsmanager/` -- MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) -- MUST FOLLOW
- **Variable:** `SystemsManagerCalls` exported slice -- MUST FOLLOW
- **Type:** `[]types.AWSService` -- MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` -- MUST FOLLOW (all Epic 2 stories use DefaultModuleName)
- **Session handling:** `sess.Copy(&aws.Config{Region: aws.String(region)})` -- MUST FOLLOW (Story 2.3 code review fix)
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` -- MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` -- MUST FOLLOW
- **Nil checks:** Always check `*string`, `*time.Time`, `*int64` before dereferencing -- MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) -- MUST FOLLOW
- **SDK version:** AWS SDK Go v1.44.266 -- MUST FOLLOW (do NOT use SDK v2)

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/systemsmanager/
+-- calls.go            # NEW: SSM parameters service implementation
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
cmd/awtest/services/redshift/calls.go          # Story 2.8 reference (sess.Copy + resilient + Marker pagination) -- PRIMARY REFERENCE
cmd/awtest/services/redshift/calls_test.go     # Story 2.8 test reference (table-driven Process()-only)
cmd/awtest/services/stepfunctions/calls.go     # Story 2.9 reference (sess.Copy + resilient + NextToken pagination)
cmd/awtest/services/services.go                # AllServices() registration point
go.mod                                         # AWS SDK already included (ssm package available)
```

### Previous Story Intelligence (Story 2.9 - Step Functions)

**Key learnings from Story 2.9 (Step Functions):**
- **sess.Copy() is mandatory** -- continued from Story 2.3 fix
- **Resilient per-region errors** with `anyRegionSucceeded` + `lastErr` tracking pattern
- **Pagination included from the start** -- avoid code review rework
- Table-driven Process()-only tests are the standard
- Type assertion failure handling included from the start
- All display fields must appear in BOTH PrintResult AND Details map
- `ScanResult.Timestamp = time.Now()` is required on every result
- Empty results handled with `utils.PrintAccessGranted`
- Story 2.9 registered between sqs and sts -- this story registers between stepfunctions and transcribe
- Cross-cutting review findings deferred -- do NOT add these

### Git Intelligence

**Recent commits (Epic 2 context):**
- `c94b560` Mark Story 2.9 as done
- `a086ad4` Add Step Functions state machines service enumeration (Story 2.9)
- `ebf7392` Mark Story 2.8 as done
- `40c71c3` Add Redshift clusters service enumeration (Story 2.8)
- `f814fab` Fix false "Access granted" when all regions return access denied
- `f46641a` Add Fargate tasks service enumeration (Story 2.7)

**Key insight from f814fab:** The `anyRegionSucceeded` + `lastErr` pattern in Call() is critical. If all regions return access denied, Call() must return the error (not nil) so Process() can properly report it. This was a bug fix applied after Story 2.7.

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `systemsmanager` follows convention (lowercase, single word, matches directory; same pattern as `secretsmanager`, `certificatemanager`)
- Single `calls.go` file per service -- matches all 40+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/systemsmanager`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.10: Systems Manager SSM Parameters Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR31 Systems Manager Parameters]
- [Source: _bmad-output/planning-artifacts/architecture.md#Service Enumeration Pattern]
- [Source: _bmad-output/implementation-artifacts/2-9-step-functions-state-machines-enumeration.md -- previous story learnings]
- [Source: cmd/awtest/services/redshift/calls.go -- PRIMARY reference (sess.Copy + resilient + pagination)]
- [Source: cmd/awtest/services/redshift/calls_test.go -- test reference (table-driven Process()-only)]
- [Source: cmd/awtest/services/services.go -- AllServices() registration point]
- [Source: cmd/awtest/types/types.go -- AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go -- PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem]
- [Source: go.mod -- aws-sdk-go v1.44.266 (includes ssm package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

### Completion Notes List

- Implemented SSM DescribeParameters service enumeration following Redshift/Step Functions pattern
- Call() uses sess.Copy() for safe region iteration, NextToken pagination, resilient per-region error handling
- Process() extracts Name, Type, Description, LastModifiedDate, Version with nil checks on all fields
- Only enumerates parameter metadata (NFR7 compliant) -- does NOT retrieve parameter values
- Registered in AllServices() alphabetically between stepfunctions and transcribe
- 6 table-driven Process()-only tests: String param, SecureString param, multiple params, empty results, access denied, nil fields
- All tests pass, go build succeeds, go vet clean, no regressions
- Date: 2026-03-05

### Change Log

- 2026-03-05: Implemented Story 2.10 - SSM Parameters Enumeration (all tasks complete)

### File List

- cmd/awtest/services/systemsmanager/calls.go (NEW)
- cmd/awtest/services/systemsmanager/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED - added import and registration)
