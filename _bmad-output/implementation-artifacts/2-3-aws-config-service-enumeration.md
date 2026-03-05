# Story 2.3: AWS Config Service Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate AWS Config recorders and rules**,
so that **I can discover configuration compliance tracking accessible with the credentials, revealing monitored resources and compliance rules**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/config/` directory with `calls.go`
2. **AC2:** Implement `DescribeConfigurationRecorders()` and `DescribeConfigRules()` API calls using AWS SDK v1.44.266 ConfigService client (`github.com/aws/aws-sdk-go/service/configservice`)
3. **AC3:** Implement AWSService interface: `Name="config:DescribeConfigurationRecorders"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates ConfigService client per region, calls both `DescribeConfigurationRecorders` and `DescribeConfigRules`, aggregates results using a combined struct
5. **AC5:** `Process()` displays configuration recorders (Name, RoleARN, Recording status) and config rules (ConfigRuleName, ConfigRuleState, Source)
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results — if both recorders and rules are empty after all regions, return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `cognitouserpools`, before `dynamodb`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: valid config data with recorders and rules, empty recorders, empty rules, access denied, nil field handling
10. **AC10:** Package naming: `config` (lowercase, single word)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/config/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR30 requirement partially fulfilled: System enumerates Config recorders and rules

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/config/`
  - [x] Create `calls.go` with package `config`
  - [x] Define combined results struct for holding both recorders and rules
  - [x] Define `var ConfigCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `configservice.New(sess)` per region, call both `svc.DescribeConfigurationRecorders` and `svc.DescribeConfigRules`, aggregate into combined struct
  - [x] Return combined struct from Call(), or first error encountered

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "config:DescribeConfigurationRecorders", err)`, return error ScanResult
  - [x] Type-assert output to combined results struct
  - [x] For each recorder, extract: `Name` (`*string`), `RoleARN` (`*string`) with nil checks
  - [x] For each rule, extract: `ConfigRuleName` (`*string`), `ConfigRuleState` (`*string`), `Source.Owner` (`*string`) with nil checks
  - [x] Build `types.ScanResult` entries with: ServiceName="Config", MethodName="config:DescribeConfigurationRecorders" (for recorders) and "config:DescribeConfigRules" (for rules)
  - [x] Call `utils.PrintResult()` with formatted output
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/config"` to `services/services.go`
  - [x] Add `allServices = append(allServices, config.ConfigCalls...)` after `cognitouserpools.CognitoUserPoolsCalls...` and before `dynamodb.DynamoDBCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/config/calls_test.go`
  - [x] Follow Story 2.2 test pattern: Process()-only tests with pre-built mock data
  - [x] Test cases: valid recorders + rules, empty recorders, empty rules, access denied error, nil field handling

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/config/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow Story 2.1/2.2 Pattern with Combined Results Extension

Stories 2.1 (ACM) and 2.2 (Cognito User Pools) established the pattern for Epic 2 service additions. This story extends the pattern by calling TWO APIs in a single AWSService entry. The key difference: you need a combined results struct since `Call()` returns a single `interface{}`.

### AWS Config SDK Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/configservice`

**IMPORTANT:** The AWS SDK package is `configservice` but the project package directory is `config`. These are different names intentionally — the SDK package name is the AWS SDK convention, the project package name follows the epics specification.

**Key APIs:**

**1. DescribeConfigurationRecorders:**
- `configservice.New(sess)` — creates Config client
- `svc.DescribeConfigurationRecorders(&configservice.DescribeConfigurationRecordersInput{})` — returns `*configservice.DescribeConfigurationRecordersOutput`
- `Output.ConfigurationRecorders` — `[]*configservice.ConfigurationRecorder`
- No pagination needed — typically 1 recorder per account per region

**ConfigurationRecorder fields:**
- `Name` — `*string` — recorder name (usually "default")
- `RoleARN` — `*string` — IAM role ARN used by recorder
- `RecordingGroup` — `*RecordingGroup` — what resources to record
  - `RecordingGroup.AllSupported` — `*bool` — records all resource types
  - `RecordingGroup.IncludeGlobalResourceTypes` — `*bool` — includes global resources

**2. DescribeConfigRules:**
- `svc.DescribeConfigRules(&configservice.DescribeConfigRulesInput{})` — returns `*configservice.DescribeConfigRulesOutput`
- `Output.ConfigRules` — `[]*configservice.ConfigRule`
- May need pagination via `NextToken` for accounts with many rules

**ConfigRule fields:**
- `ConfigRuleName` — `*string` — rule name
- `ConfigRuleState` — `*string` — "ACTIVE", "DELETING", "DELETING_RESULTS", "EVALUATING"
- `Source` — `*Source` — rule source
  - `Source.Owner` — `*string` — "CUSTOM_LAMBDA" or "AWS"
- `ConfigRuleArn` — `*string` — rule ARN

**No new dependencies needed** — `configservice` is part of `aws-sdk-go v1.44.266` already in go.mod.

### Combined Results Pattern

Since `Call()` returns a single `interface{}` but this service calls two APIs, use a combined struct:

```go
type configResults struct {
    Recorders []*configservice.ConfigurationRecorder
    Rules     []*configservice.ConfigRule
}
```

`Call()` returns `&configResults{...}` and `Process()` type-asserts to `*configResults`.

### Reference Implementation Pattern

```go
package config

import (
    "fmt"
    "github.com/MillerMedia/awtest/cmd/awtest/types"
    "github.com/MillerMedia/awtest/cmd/awtest/utils"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/configservice"
    "time"
)

type configResults struct {
    Recorders []*configservice.ConfigurationRecorder
    Rules     []*configservice.ConfigRule
}

var ConfigCalls = []types.AWSService{
    {
        Name: "config:DescribeConfigurationRecorders",
        Call: func(sess *session.Session) (interface{}, error) {
            var allRecorders []*configservice.ConfigurationRecorder
            var allRules []*configservice.ConfigRule
            for _, region := range types.Regions {
                sess.Config.Region = aws.String(region)
                svc := configservice.New(sess)

                recOutput, err := svc.DescribeConfigurationRecorders(&configservice.DescribeConfigurationRecordersInput{})
                if err != nil {
                    return nil, err
                }
                allRecorders = append(allRecorders, recOutput.ConfigurationRecorders...)

                rulesOutput, err := svc.DescribeConfigRules(&configservice.DescribeConfigRulesInput{})
                if err != nil {
                    return nil, err
                }
                allRules = append(allRules, rulesOutput.ConfigRules...)
            }
            return &configResults{Recorders: allRecorders, Rules: allRules}, nil
        },
        Process: func(output interface{}, err error, debug bool) []types.ScanResult {
            // Handle error, type-assert to *configResults
            // Process recorders and rules separately
            // Follow ACM Process() pattern for each
        },
        ModuleName: types.DefaultModuleName,
    },
}
```

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `config` |
| Package variable | `ConfigCalls` |
| AWSService.Name | `"config:DescribeConfigurationRecorders"` |
| ScanResult.ServiceName (recorders) | `"Config"` |
| ScanResult.MethodName (recorders) | `"config:DescribeConfigurationRecorders"` |
| ScanResult.ResourceType (recorders) | `"configuration-recorder"` |
| ScanResult.ServiceName (rules) | `"Config"` |
| ScanResult.MethodName (rules) | `"config:DescribeConfigRules"` |
| ScanResult.ResourceType (rules) | `"config-rule"` |

### Registration Order in services.go

**CORRECTION from epics:** Epics say "after codepipeline, before cognitoidentity" but alphabetically `config` comes AFTER `cognitouserpools` (co-g < co-n). Insert after `cognitouserpools` and before `dynamodb`:

```go
allServices = append(allServices, cognitouserpools.CognitoUserPoolsCalls...)
allServices = append(allServices, config.ConfigCalls...)  // NEW
allServices = append(allServices, dynamodb.DynamoDBCalls...)
```

Import alphabetically after `cognitouserpools`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/cognitouserpools"
"github.com/MillerMedia/awtest/cmd/awtest/services/config"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/dynamodb"
```

### Process() Output Format

**For recorders:**
```go
utils.PrintResult(debug, "", "config:DescribeConfigurationRecorders",
    fmt.Sprintf("Found Config Recorder: %s (Role: %s)", utils.ColorizeItem(recorderName), roleArn), nil)
```

**For rules:**
```go
utils.PrintResult(debug, "", "config:DescribeConfigRules",
    fmt.Sprintf("Found Config Rule: %s (State: %s, Owner: %s)", utils.ColorizeItem(ruleName), ruleState, sourceOwner), nil)
```

### Testing Pattern (from Story 2.2)

Create Process()-only tests with pre-built mock data using the combined struct. No AWS SDK mocking needed.

```go
func TestProcess_ValidConfig(t *testing.T) {
    process := ConfigCalls[0].Process
    results := &configResults{
        Recorders: []*configservice.ConfigurationRecorder{
            {Name: aws.String("default"), RoleARN: aws.String("arn:aws:iam::123:role/config-role")},
        },
        Rules: []*configservice.ConfigRule{
            {ConfigRuleName: aws.String("s3-bucket-versioning"), ConfigRuleState: aws.String("ACTIVE"),
             Source: &configservice.Source{Owner: aws.String("AWS")}},
        },
    }
    scanResults := process(results, nil, false)
    // assert recorder + rule results
}
```

Test cases:
1. **Valid config** — recorders + rules present, verify all ScanResult fields and Details
2. **Empty recorders** — only rules present, verify only rule results returned
3. **Empty rules** — only recorders present, verify only recorder results returned
4. **Access denied** — verify error ScanResult returned with correct fields
5. **Nil fields** — verify nil Name/RoleARN/ConfigRuleName handled gracefully

### Edge Cases

1. **No recorders or rules in any region** — Call() returns empty combined struct, Process() returns empty results
2. **Access denied in first region** — Call() returns error immediately (fail fast, matches ACM pattern)
3. **Recorder with nil Name** — defensive nil check, use empty string
4. **Rule with nil Source** — defensive nil check on Source before accessing Source.Owner
5. **Many rules requiring pagination** — for initial implementation, single call without NextToken is acceptable (most accounts have < 25 rules per region); pagination can be added as enhancement

### Architecture Compliance

- **Package:** `config` in `cmd/awtest/services/config/` — MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) — MUST FOLLOW
- **Variable:** `ConfigCalls` exported slice — MUST FOLLOW
- **Type:** `[]types.AWSService` — MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` — MUST FOLLOW (epics erroneously say "Config")
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` — MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` — MUST FOLLOW
- **Nil checks:** Always check `*string` fields before dereferencing — MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) — MUST FOLLOW

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/config/
+-- calls.go            # NEW: AWS Config service implementation
+-- calls_test.go       # NEW: Process() tests
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                      # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                     # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/certificatemanager/calls.go # Story 2.1 reference implementation
cmd/awtest/services/cognitouserpools/calls.go   # Story 2.2 reference implementation
go.mod                                         # AWS SDK already included
```

### Previous Story Intelligence (Story 2.2)

**Key learnings from Story 2.2 (Cognito User Pools):**
- Follow the exact Call/Process pattern — consistency is critical
- `utils.PrintResult()` handles quiet mode automatically via `utils.Quiet` flag
- `utils.HandleAWSError()` detects InvalidKeyError for abort handling
- Region iteration pattern: mutate `sess.Config.Region` in loop, create new client per region
- ScanResult must include `Timestamp: time.Now()`
- Details map should include relevant fields (Id, Status, etc.)
- Process()-only tests are sufficient — no AWS SDK mocking needed
- Tests should cover: valid data, empty results, error cases, nil field handling
- Code review caught missing AC5 fields (Status, CreationDate) — ensure ALL specified display fields are in both PrintResult AND Details map
- Small, focused commits referencing story numbers

### Git Intelligence

**Recent commits (Epic 2 context):**
- `c71ad1b` Add ACM and Cognito User Pools service enumeration (Stories 2.1, 2.2)
- `94ba7a6` Mark Story 1.7 and Epic 1 (Output Format System) as done
- `ba40be3` Add progress tracking, summary reporting, and quiet mode (Story 1.7)

**Patterns established:**
- Commits reference story numbers in message
- Multiple stories can be in a single commit if related
- Build verification (go build, go test, go vet) done before commit

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `config` follows convention (lowercase, single word, matches directory)
- Single `calls.go` file per service — matches all 38+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/config`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.3: AWS Config Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR7-31 Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#Service Enumeration pattern]
- [Source: _bmad-output/implementation-artifacts/2-2-cognito-user-pools-service-enumeration.md — previous story learnings]
- [Source: cmd/awtest/services/certificatemanager/calls.go — Story 2.1 reference implementation]
- [Source: cmd/awtest/services/cognitouserpools/calls.go — Story 2.2 reference implementation]
- [Source: cmd/awtest/services/services.go — AllServices() registration point]
- [Source: cmd/awtest/types/types.go — AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go — PrintResult, HandleAWSError, ColorizeItem]
- [Source: go.mod — aws-sdk-go v1.44.266 (includes configservice package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Implemented AWS Config service enumeration with combined results pattern (three API calls in one AWSService entry)
- `configResults` struct holds `[]*ConfigurationRecorder`, `[]*ConfigurationRecorderStatus`, and `[]*ConfigRule`
- Call() iterates all regions using `sess.Copy()` for safe session handling, calling DescribeConfigurationRecorders, DescribeConfigurationRecorderStatus, and DescribeConfigRules per region
- Process() handles recorders and rules separately with distinct ServiceName/MethodName/ResourceType
- Recorder details include RoleARN and actual RecordingStatus (Recording/Stopped) from recorder status API
- Rule details include State and Source Owner
- All nil pointer checks implemented for *string fields and nested Source struct
- 6 test cases covering: valid data, recorder stopped, empty recorders, empty rules, access denied, nil fields
- All builds, tests, and vet checks pass with no regressions

### Code Review Follow-ups

- [x] HIGH: Fixed unsafe session mutation - now uses `sess.Copy()` instead of mutating shared `sess.Config.Region`
- [x] LOW: Added `DescribeConfigurationRecorderStatus` API call to get actual recording status (Recording/Stopped) per AC5
- [ ] MEDIUM (deferred): Fail-fast on region error is the established codebase pattern (ACM, Cognito all do the same). Recommend addressing codebase-wide in a future story rather than creating inconsistency here.

### Change Log

- 2026-03-05: Implemented Story 2.3 - AWS Config service enumeration (recorders + rules)
- 2026-03-05: Addressed code review - fixed session mutation (sess.Copy), added recording status API

### File List

- cmd/awtest/services/config/calls.go (NEW)
- cmd/awtest/services/config/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED)
