# Story 2.4: EFS (Elastic File System) Service Enumeration

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate EFS file systems**,
so that **I can discover network-attached storage accessible with the credentials, which may contain sensitive data and mount targets**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/efs/` directory with `calls.go`
2. **AC2:** Implement `DescribeFileSystems()` API call using AWS SDK v1.44.266 EFS client (`github.com/aws/aws-sdk-go/service/efs`)
3. **AC3:** Implement AWSService interface: `Name="efs:DescribeFileSystems"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates EFS client per region using `sess.Copy()`, calls `DescribeFileSystems`, aggregates results
5. **AC5:** `Process()` displays each file system: FileSystemId, Name, LifeCycleState, SizeInBytes, NumberOfMountTargets, Encrypted
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results — if no file systems found after all regions, call `utils.PrintAccessGranted(debug, "efs:DescribeFileSystems", "file systems")` and return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `ecs`, before `elasticbeanstalk`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: valid file systems with all fields, encrypted vs unencrypted, empty results, access denied, nil field handling
10. **AC10:** Package naming: `efs` (lowercase, single word)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/efs/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR28 requirement partially fulfilled: System enumerates EFS file systems

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/efs/`
  - [x] Create `calls.go` with package `efs`
  - [x] Define `var EfsCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `efs.New(regionSess)` per region using `sess.Copy()`, call `svc.DescribeFileSystems(&efs.DescribeFileSystemsInput{})`, aggregate `FileSystemDescriptions` across regions
  - [x] Return aggregated `[]*efs.FileSystemDescription` from Call(), or first error encountered

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "efs:DescribeFileSystems", err)`, return error ScanResult
  - [x] Type-assert output to `[]*efs.FileSystemDescription`
  - [x] If empty slice and no error: call `utils.PrintAccessGranted(debug, "efs:DescribeFileSystems", "file systems")`, return empty results
  - [x] For each file system, extract: `FileSystemId` (`*string`), `Name` (`*string`), `LifeCycleState` (`*string`), `SizeInBytes.Value` (`*int64`), `NumberOfMountTargets` (`*int64`), `Encrypted` (`*bool`) — all with nil checks
  - [x] Build `types.ScanResult` entries with: ServiceName="EFS", MethodName="efs:DescribeFileSystems", ResourceType="file-system"
  - [x] Call `utils.PrintResult()` with formatted output
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/efs"` to `services/services.go`
  - [x] Add `allServices = append(allServices, efs.EfsCalls...)` after `ecs.ECSCalls...` and before `elasticbeanstalk.ElasticBeanstalkCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/efs/calls_test.go`
  - [x] Follow Story 2.2/2.3 test pattern: Process()-only tests with pre-built mock data
  - [x] Test cases: valid file systems (all fields populated), encrypted vs unencrypted, empty results, access denied error, nil field handling

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/efs/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow Established Epic 2 Pattern

Stories 2.1 (ACM), 2.2 (Cognito User Pools), and 2.3 (Config) established the pattern for Epic 2 service additions. This story is simpler than 2.3 — it uses a single API call (`DescribeFileSystems`) returning a flat slice, matching the Cognito pattern.

### CRITICAL: Use sess.Copy() for Region Iteration

Story 2.3 code review identified that mutating `sess.Config.Region` directly is unsafe. The fix was to use `sess.Copy()`. **YOU MUST USE `sess.Copy()`** for safe session handling:

```go
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := efs.New(regionSess)
    // ...
}
```

**DO NOT** use the older pattern from Story 2.2 (Cognito) which mutates `sess.Config.Region` directly.

### AWS EFS SDK Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/efs`

**Key API: DescribeFileSystems**
- `efs.New(sess)` — creates EFS client
- `svc.DescribeFileSystems(&efs.DescribeFileSystemsInput{})` — returns `*efs.DescribeFileSystemsOutput`
- `Output.FileSystems` — `[]*efs.FileSystemDescription`
- May need pagination via `Marker`/`NextMarker` for accounts with many file systems (initial implementation without pagination is acceptable — most accounts have < 100 EFS file systems)

**FileSystemDescription fields:**
- `FileSystemId` — `*string` — e.g., "fs-12345678"
- `Name` — `*string` — user-assigned name (can be nil if no Name tag set)
- `LifeCycleState` — `*string` — "available", "creating", "deleting", "deleted", "updating"
- `SizeInBytes` — `*FileSystemSize` — nested struct
  - `SizeInBytes.Value` — `*int64` — size in bytes (always present when SizeInBytes is non-nil)
- `NumberOfMountTargets` — `*int64` — count of mount targets
- `Encrypted` — `*bool` — whether file system is encrypted
- `CreationTime` — `*time.Time` — when created (not in AC5 but available)
- `PerformanceMode` — `*string` — "generalPurpose" or "maxIO" (not in AC5 but available)

**No new dependencies needed** — `efs` is part of `aws-sdk-go v1.44.266` already in go.mod.

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `efs` |
| Package variable | `EfsCalls` |
| AWSService.Name | `"efs:DescribeFileSystems"` |
| ScanResult.ServiceName | `"EFS"` |
| ScanResult.MethodName | `"efs:DescribeFileSystems"` |
| ScanResult.ResourceType | `"file-system"` |

### Registration Order in services.go

Insert after `ecs` and before `elasticbeanstalk` (alphabetical):

```go
allServices = append(allServices, ecs.ECSCalls...)
allServices = append(allServices, efs.EfsCalls...)  // NEW
allServices = append(allServices, elasticbeanstalk.ElasticBeanstalkCalls...)
```

Import alphabetically after `ecs`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/ecs"
"github.com/MillerMedia/awtest/cmd/awtest/services/efs"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/elasticbeanstalk"
```

### Process() Output Format

```go
// Format size for readability
sizeStr := fmt.Sprintf("%d bytes", sizeValue)

utils.PrintResult(debug, "", "efs:DescribeFileSystems",
    fmt.Sprintf("Found EFS: %s (Name: %s, State: %s, Size: %s, MountTargets: %d, Encrypted: %v)",
        utils.ColorizeItem(fsId), name, lifecycleState, sizeStr, mountTargets, encrypted), nil)
```

### Empty Results Handling

Unlike Stories 2.1-2.3 which silently return empty results, the epics specify using `utils.PrintAccessGranted` for EFS empty results. This function exists in `utils/output.go:65` and is used by older services (apigateway, dynamodb, rekognition):

```go
if len(fileSystems) == 0 {
    utils.PrintAccessGranted(debug, "efs:DescribeFileSystems", "file systems")
    return results
}
```

### Reference Implementation Pattern

```go
package efs

import (
    "fmt"
    "github.com/MillerMedia/awtest/cmd/awtest/types"
    "github.com/MillerMedia/awtest/cmd/awtest/utils"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/efs"
    "time"
)

var EfsCalls = []types.AWSService{
    {
        Name: "efs:DescribeFileSystems",
        Call: func(sess *session.Session) (interface{}, error) {
            var allFileSystems []*efs.FileSystemDescription
            for _, region := range types.Regions {
                regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
                svc := efs.New(regionSess)
                output, err := svc.DescribeFileSystems(&efs.DescribeFileSystemsInput{})
                if err != nil {
                    return nil, err
                }
                allFileSystems = append(allFileSystems, output.FileSystems...)
            }
            return allFileSystems, nil
        },
        Process: func(output interface{}, err error, debug bool) []types.ScanResult {
            var results []types.ScanResult

            if err != nil {
                utils.HandleAWSError(debug, "efs:DescribeFileSystems", err)
                return []types.ScanResult{
                    {
                        ServiceName: "EFS",
                        MethodName:  "efs:DescribeFileSystems",
                        Error:       err,
                        Timestamp:   time.Now(),
                    },
                }
            }

            if fileSystems, ok := output.([]*efs.FileSystemDescription); ok {
                if len(fileSystems) == 0 {
                    utils.PrintAccessGranted(debug, "efs:DescribeFileSystems", "file systems")
                    return results
                }
                for _, fs := range fileSystems {
                    // Extract fields with nil checks
                    // Build ScanResult with Details map
                    // Call utils.PrintResult
                }
            }
            return results
        },
        ModuleName: types.DefaultModuleName,
    },
}
```

### Testing Pattern (from Story 2.2/2.3)

Create Process()-only tests with pre-built mock data. No AWS SDK mocking needed.

```go
func TestProcess_ValidFileSystems(t *testing.T) {
    process := EfsCalls[0].Process
    fileSystems := []*efs.FileSystemDescription{
        {
            FileSystemId:       aws.String("fs-12345678"),
            Name:               aws.String("my-efs"),
            LifeCycleState:     aws.String("available"),
            SizeInBytes:        &efs.FileSystemSize{Value: aws.Int64(1024000)},
            NumberOfMountTargets: aws.Int64(2),
            Encrypted:          aws.Bool(true),
        },
    }
    results := process(fileSystems, nil, false)
    // assert results
}
```

Test cases:
1. **Valid file systems** — all fields populated, verify ScanResult fields and Details map
2. **Encrypted vs unencrypted** — verify Encrypted field is captured correctly for both true/false
3. **Empty results** — verify `PrintAccessGranted` behavior and empty results returned
4. **Access denied** — verify error ScanResult returned with correct ServiceName/MethodName
5. **Nil fields** — verify nil Name, nil SizeInBytes, nil Encrypted handled gracefully (empty string/zero/false defaults)

### Edge Cases

1. **No file systems in any region** — Call() returns empty slice, Process() calls PrintAccessGranted, returns empty results
2. **Access denied in first region** — Call() returns error immediately (fail fast, matches established pattern)
3. **File system with nil Name** — common for EFS created without tags; defensive nil check, use empty string
4. **File system with nil SizeInBytes** — defensive nil check on SizeInBytes before accessing SizeInBytes.Value
5. **File system with nil Encrypted** — defensive nil check, default to false for display

### Architecture Compliance

- **Package:** `efs` in `cmd/awtest/services/efs/` — MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) — MUST FOLLOW
- **Variable:** `EfsCalls` exported slice — MUST FOLLOW
- **Type:** `[]types.AWSService` — MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` — MUST FOLLOW (epics say "EFS" but all Epic 2 stories use DefaultModuleName)
- **Session handling:** `sess.Copy(&aws.Config{Region: aws.String(region)})` — MUST FOLLOW (Story 2.3 code review fix)
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` — MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` — MUST FOLLOW
- **Nil checks:** Always check `*string`, `*int64`, `*bool`, and nested struct fields before dereferencing — MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) — MUST FOLLOW

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/efs/
+-- calls.go            # NEW: AWS EFS service implementation
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
cmd/awtest/services/config/calls.go            # Story 2.3 reference (sess.Copy pattern)
cmd/awtest/services/cognitouserpools/calls.go  # Story 2.2 reference (single API pattern)
go.mod                                         # AWS SDK already included
```

### Previous Story Intelligence (Story 2.3)

**Key learnings from Story 2.3 (Config):**
- **sess.Copy() is mandatory** — Story 2.3 code review caught unsafe session mutation. `sess.Copy()` was the fix. This MUST be used going forward.
- Combined results struct needed only when calling multiple APIs — EFS uses a single API, so a flat slice return is sufficient (simpler, like Cognito pattern)
- `DescribeConfigurationRecorderStatus` was an additional API added during code review — EFS only needs one API call, keeping it simpler
- All nil pointer checks on `*string` fields remain critical
- Process()-only tests with pre-built mock data pattern continues to work well
- 6 test cases in Story 2.3 — aim for 5 covering the key scenarios

**Key learnings from Story 2.2 (Cognito User Pools):**
- Code review caught missing AC5 display fields — ensure ALL specified fields (FileSystemId, Name, LifeCycleState, SizeInBytes, NumberOfMountTargets, Encrypted) appear in both PrintResult AND Details map
- ScanResult must include `Timestamp: time.Now()`

### Git Intelligence

**Recent commits (Epic 2 context):**
- `c71ad1b` Add ACM and Cognito User Pools service enumeration (Stories 2.1, 2.2)
- Story 2.3 changes are currently uncommitted (in review status)

**Patterns established:**
- Commits reference story numbers in message
- Build verification (go build, go test, go vet) done before commit
- Small, focused commits

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `efs` follows convention (lowercase, single word, matches directory)
- Single `calls.go` file per service — matches all 38+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/efs`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.4: EFS (Elastic File System) Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR7-31 Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#Phase 1 Additions Needed]
- [Source: _bmad-output/implementation-artifacts/2-3-aws-config-service-enumeration.md — previous story learnings, sess.Copy() pattern]
- [Source: cmd/awtest/services/config/calls.go — Story 2.3 reference (sess.Copy pattern)]
- [Source: cmd/awtest/services/cognitouserpools/calls.go — Story 2.2 reference (single API pattern)]
- [Source: cmd/awtest/services/services.go — AllServices() registration point]
- [Source: cmd/awtest/types/types.go — AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go — PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem]
- [Source: go.mod — aws-sdk-go v1.44.266 (includes efs package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no issues.

### Completion Notes List

- Implemented EFS DescribeFileSystems service enumeration following established Epic 2 pattern
- Used `sess.Copy()` for safe region iteration (per Story 2.3 code review finding)
- Followed Cognito (Story 2.2) single-API flat slice pattern
- All 6 AC5 fields displayed in PrintResult AND stored in Details map: FileSystemId, Name, LifeCycleState, SizeInBytes, NumberOfMountTargets, Encrypted
- Empty results handled with `utils.PrintAccessGranted` per AC7
- 5 table-driven test cases covering: valid file systems, encrypted vs unencrypted, empty results, access denied, nil field handling
- All builds, tests, and vet pass clean with no regressions
- Addressed code review findings: converted tests to table-driven format (AC9), added type assertion failure handling

### Change Log

- 2026-03-05: Implemented Story 2.4 EFS service enumeration — created efs package with Call()/Process(), registered in AllServices(), wrote 5 unit tests
- 2026-03-05: Addressed code review findings — 2 items resolved (1 Medium: table-driven tests per AC9, 1 Low: type assertion failure handling)

### File List

- `cmd/awtest/services/efs/calls.go` (NEW)
- `cmd/awtest/services/efs/calls_test.go` (NEW)
- `cmd/awtest/services/services.go` (MODIFIED)
