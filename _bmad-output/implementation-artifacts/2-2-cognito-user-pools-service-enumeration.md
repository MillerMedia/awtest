# Story 2.2: Cognito User Pools Service Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Cognito user pools**,
so that **I can discover user authentication databases accessible with the credentials, which may contain user PII and authentication configurations**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/cognitouserpools/` directory with `calls.go`
2. **AC2:** Implement `ListUserPools()` API call using AWS SDK v1.44.266 Cognito Identity Provider client (`github.com/aws/aws-sdk-go/service/cognitoidentityprovider`)
3. **AC3:** Implement AWSService interface: `Name="cognito-idp:ListUserPools"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates Cognito Identity Provider client per region, calls `ListUserPools` with `MaxResults=60`, aggregates results
5. **AC5:** `Process()` displays each user pool: Name, Id, Status, CreationDate
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results — if user pools list is empty after all regions, return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `cognitoidentity`, before `dynamodb`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: valid user pools, empty results, access denied, nil field handling
10. **AC10:** Package naming: `cognitouserpools` (lowercase, no underscores)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/cognitouserpools/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR27 requirement fulfilled: System enumerates Cognito user pools

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/cognitouserpools/`
  - [x] Create `calls.go` with package `cognitouserpools`
  - [x] Define `var CognitoUserPoolsCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `cognitoidentityprovider.New(sess)` per region, call `svc.ListUserPools(&cognitoidentityprovider.ListUserPoolsInput{MaxResults: aws.Int64(60)})`, aggregate `[]*cognitoidentityprovider.UserPoolDescriptionType`
  - [x] Return aggregated slice from Call(), or first error encountered

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "cognito-idp:ListUserPools", err)`, return error ScanResult
  - [x] Type-assert output to `[]*cognitoidentityprovider.UserPoolDescriptionType`
  - [x] For each user pool, extract: `Name` (`*string`), `Id` (`*string`) with nil checks
  - [x] Build `types.ScanResult` with: ServiceName="CognitoUserPools", MethodName="cognito-idp:ListUserPools", ResourceType="user-pool", ResourceName=poolName, Details=empty map
  - [x] Call `utils.PrintResult()` with formatted output showing pool name and ID
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/cognitouserpools"` to `services/services.go`
  - [x] Add `allServices = append(allServices, cognitouserpools.CognitoUserPoolsCalls...)` after `cognitoidentity.CognitoIdentityCalls...` and before `dynamodb.DynamoDBCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/cognitouserpools/calls_test.go`
  - [x] Follow Story 2.1 test pattern: Process()-only tests with pre-built mock data
  - [x] Test cases: valid user pools, empty results, access denied error, nil field handling

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/cognitouserpools/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow Story 2.1 (ACM) Pattern Exactly

Story 2.1 (Certificate Manager) established the pattern for Epic 2 service additions. Follow it precisely. The key difference from older services: region iteration AND proper nil checks on all pointer fields.

### Cognito User Pools SDK Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/cognitoidentityprovider`

**IMPORTANT:** The SDK package is `cognitoidentityprovider` but the project package directory is `cognitouserpools`. These are different names intentionally — the SDK package name is the AWS SDK convention, the project package name follows the project's service naming convention.

**Key API:**
- `cognitoidentityprovider.New(sess)` — creates Cognito User Pools client
- `svc.ListUserPools(&cognitoidentityprovider.ListUserPoolsInput{MaxResults: aws.Int64(60)})` — returns `*cognitoidentityprovider.ListUserPoolsOutput`
- `ListUserPoolsOutput.UserPools` — `[]*cognitoidentityprovider.UserPoolDescriptionType`
- **MaxResults is REQUIRED** — unlike most List APIs, ListUserPools requires MaxResults (max value: 60)

**UserPoolDescriptionType fields:**
- `Name` — `*string` — user pool name
- `Id` — `*string` — user pool ID (e.g., "us-east-1_abc123xyz")
- `Status` — `*string` — pool status
- `CreationDate` — `*time.Time` — creation timestamp
- `LastModifiedDate` — `*time.Time` — last modification timestamp
- `LambdaConfig` — `*LambdaConfigType` — Lambda triggers (not needed for basic enumeration)

**No new dependencies needed** — `cognitoidentityprovider` is part of `aws-sdk-go v1.44.266` already in go.mod.

### Reference Implementation Pattern

Follow the ACM implementation from Story 2.1 (`cmd/awtest/services/certificatemanager/calls.go`):

```go
package cognitouserpools

import (
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"time"
)

var CognitoUserPoolsCalls = []types.AWSService{
	{
		Name: "cognito-idp:ListUserPools",
		Call: func(sess *session.Session) (interface{}, error) {
			var allUserPools []*cognitoidentityprovider.UserPoolDescriptionType
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := cognitoidentityprovider.New(sess)
				output, err := svc.ListUserPools(&cognitoidentityprovider.ListUserPoolsInput{
					MaxResults: aws.Int64(60),
				})
				if err != nil {
					return nil, err
				}
				allUserPools = append(allUserPools, output.UserPools...)
			}
			return allUserPools, nil
		},
		Process: func(output interface{}, err error, debug bool) []types.ScanResult {
			// ... follow ACM Process() pattern exactly
		},
		ModuleName: types.DefaultModuleName,
	},
}
```

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `cognitouserpools` |
| Package variable | `CognitoUserPoolsCalls` |
| AWSService.Name | `"cognito-idp:ListUserPools"` |
| ScanResult.ServiceName | `"CognitoUserPools"` |
| ScanResult.MethodName | `"cognito-idp:ListUserPools"` |
| ScanResult.ResourceType | `"user-pool"` |

### Registration Order in services.go

Insert alphabetically between `cognitoidentity` and `dynamodb`:

```go
allServices = append(allServices, cognitoidentity.CognitoIdentityCalls...)
allServices = append(allServices, cognitouserpools.CognitoUserPoolsCalls...)  // NEW
allServices = append(allServices, dynamodb.DynamoDBCalls...)
```

### Process() Output Format

```go
utils.PrintResult(debug, "", "cognito-idp:ListUserPools",
    fmt.Sprintf("Found User Pool: %s (ID: %s)", utils.ColorizeItem(poolName), poolId), nil)
```

### Differences from Sibling Service (cognitoidentity)

The existing `cognitoidentity` package enumerates Identity Pools (federated identities) — a different Cognito service. Key differences:
- **cognitoidentity** does NOT iterate regions (single-region call) — **cognitouserpools MUST iterate regions** (Cognito User Pools is regional, following the ACM pattern from Story 2.1)
- **cognitoidentity** does NOT have nil checks on pointer fields — **cognitouserpools MUST have nil checks** (learned from Story 2.1)
- Different SDK package: `cognitoidentity` vs `cognitoidentityprovider`
- Different API: `ListIdentityPools` vs `ListUserPools`

### Testing Pattern (from Story 2.1)

Create Process()-only tests with pre-built mock data. No AWS SDK mocking needed.

```go
func TestProcess_ValidPools(t *testing.T) {
    process := CognitoUserPoolsCalls[0].Process
    pools := []*cognitoidentityprovider.UserPoolDescriptionType{
        {Name: aws.String("my-pool"), Id: aws.String("us-east-1_abc123")},
    }
    results := process(pools, nil, false)
    // assert results
}
```

Test cases:
1. **Valid user pools** — verify ScanResult fields populated correctly
2. **Empty results** — verify empty slice returned, no panic
3. **Access denied** — verify error ScanResult returned with correct fields
4. **Nil fields** — verify nil Name/Id handled gracefully (empty string, no panic)

### Edge Cases

1. **No user pools in any region** — Call() returns empty slice, Process() returns empty results
2. **Access denied in first region** — Call() returns error immediately (fail fast, matches ACM pattern)
3. **User pool with nil Name** — defensive nil check, use empty string
4. **User pool with nil Id** — defensive nil check, use empty string
5. **MaxResults is required** — ListUserPools will fail without it, always pass `aws.Int64(60)`

### Architecture Compliance

- **Package:** `cognitouserpools` in `cmd/awtest/services/cognitouserpools/` — MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) — MUST FOLLOW
- **Variable:** `CognitoUserPoolsCalls` exported slice — MUST FOLLOW
- **Type:** `[]types.AWSService` — MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` — MUST FOLLOW
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` — MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` — MUST FOLLOW
- **Nil checks:** Always check `*string` fields before dereferencing — MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) — MUST FOLLOW

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/cognitouserpools/
+-- calls.go            # NEW: Cognito User Pools service implementation
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
cmd/awtest/services/cognitoidentity/calls.go   # Sibling service (Identity Pools) — reference only
cmd/awtest/services/certificatemanager/calls.go # Story 2.1 reference implementation
go.mod                                         # AWS SDK already included
```

### Previous Story Intelligence (Story 2.1)

**Key learnings from Story 2.1 (ACM):**
- Follow the exact Call/Process pattern — consistency is critical
- `utils.PrintResult()` handles quiet mode automatically via `utils.Quiet` flag
- `utils.HandleAWSError()` detects InvalidKeyError for abort handling
- Region iteration pattern: mutate `sess.Config.Region` in loop, create new client per region
- ScanResult must include `Timestamp: time.Now()`
- Details map can be empty `map[string]interface{}{}`
- Process()-only tests are sufficient (Option 1 — no AWS SDK mocking)
- Tests should cover: valid data, empty results, error cases, nil field handling
- Small, focused commits referencing story numbers

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `cognitouserpools` follows convention (lowercase, no underscores, matches directory)
- Single `calls.go` file per service — matches all 34+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/cognitouserpools`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.2: Cognito User Pools Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR7-31 Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#Service Enumeration pattern]
- [Source: _bmad-output/implementation-artifacts/2-1-certificate-manager-acm-service-enumeration.md — previous story learnings]
- [Source: cmd/awtest/services/certificatemanager/calls.go — Story 2.1 reference implementation]
- [Source: cmd/awtest/services/cognitoidentity/calls.go — sibling Cognito service reference]
- [Source: cmd/awtest/services/services.go — AllServices() registration point]
- [Source: cmd/awtest/types/types.go — AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go — PrintResult, HandleAWSError, ColorizeItem]
- [Source: go.mod — aws-sdk-go v1.44.266 (includes cognitoidentityprovider package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Implemented Cognito User Pools service enumeration following Story 2.1 (ACM) pattern exactly
- Call() iterates all regions, creates cognitoidentityprovider client per region, calls ListUserPools with MaxResults=60
- Process() handles errors via utils.HandleAWSError, type-asserts output, nil-checks Name/Id/Status/CreationDate fields, builds ScanResults
- Process() displays all AC5 fields: Name, Id, Status, CreationDate — populated in both Details map and PrintResult output
- Registered in services.go alphabetically between cognitoidentity and dynamodb
- 4 Process()-only tests: valid pools (with Details assertions), empty results, access denied, nil fields — all pass
- Full regression suite passes (go build, go test, go vet all clean)
- Resolved code review finding [High]: AC5 violation — added Status and CreationDate extraction and display

### Change Log

- 2026-03-05: Implemented Story 2.2 — Cognito User Pools service enumeration
- 2026-03-05: Addressed code review findings — 1 item resolved (AC5: added Status/CreationDate to Process output and Details map)

### File List

- cmd/awtest/services/cognitouserpools/calls.go (NEW)
- cmd/awtest/services/cognitouserpools/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED — added import and registration)
