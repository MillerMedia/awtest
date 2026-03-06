# Story 3.1: Service Filtering (Include/Exclude Services)

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional running targeted scans**,
I want **to include or exclude specific AWS services from enumeration**,
so that **I can run fast triage scans targeting only critical services or comprehensive audits excluding known-noisy services during time-sensitive engagements**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/service_filter.go` with filtering logic
2. **AC2:** Add `-services` flag accepting comma-separated service names (e.g., `-services=s3,ec2,iam`) -- include-only filter (FR33)
3. **AC3:** Add `-exclude-services` flag accepting comma-separated service names (e.g., `-exclude-services=cloudwatch,cloudtrail`) -- exclusion filter (FR34)
4. **AC4:** Implement `FilterServices()` function that takes the full `[]types.AWSService` slice plus include/exclude strings, returns filtered slice
5. **AC5:** Support case-insensitive service name matching (e.g., "S3", "s3" both match S3 services)
6. **AC6:** Support partial matching (e.g., "cognito" matches both `cognitoidentity` and `cognitouserpools` service packages)
7. **AC7:** Validate service names -- print warning (NOT error) for unrecognized service names so typos are visible but don't block the scan
8. **AC8:** Handle conflicting flags: if both `-services` and `-exclude-services` specified, apply `-services` include filter first, then apply exclusions from the included set
9. **AC9:** Modify `main.go` to filter `services.AllServices()` based on parsed filter before the enumeration loop
10. **AC10:** Display "Scanning N of M services..." message showing filtered vs total service count when filters are active
11. **AC11:** When no services match filter, error clearly: "No services matched filter criteria" and exit with non-zero code
12. **AC12:** Write unit tests in `cmd/awtest/services/service_filter_test.go` covering: exact matching, case-insensitive matching, partial matching, include-only, exclude-only, include+exclude combination, invalid service names, empty input, no matches
13. **AC13:** `go build ./cmd/awtest` compiles successfully
14. **AC14:** `go test ./cmd/awtest/...` passes (all existing + new tests)
15. **AC15:** `go vet ./cmd/awtest/...` passes clean
16. **AC16:** FR33 and FR34 requirements fulfilled

## Tasks / Subtasks

- [x] Task 1: Create service filter module (AC: 1, 4, 5, 6, 7, 8)
  - [x] Create `cmd/awtest/services/service_filter.go` in package `services`
  - [x] Implement `FilterServices(allServices []types.AWSService, include, exclude string) []types.AWSService`
  - [x] Implement `parseServiceList(csv string) map[string]bool` helper -- splits comma-separated input, trims whitespace, lowercases
  - [x] Implement `extractServiceName(callName string) string` helper -- extracts service prefix from AWSService.Name (e.g., "s3" from "s3:ListBuckets")
  - [x] Implement partial matching logic: check if any parsed filter name is a substring of or matches the service name
  - [x] Implement include+exclude combination: apply include filter first, then exclude from result
  - [x] Add warning output for unrecognized service names (names that don't match any service in AllServices)

- [x] Task 2: Add CLI flags to main.go (AC: 2, 3, 9, 10, 11)
  - [x] Add `flag.StringVar(&includeServices, "services", "", "Include only specific services (comma-separated, e.g., s3,ec2,iam)")`
  - [x] Add `flag.StringVar(&excludeServices, "exclude-services", "", "Exclude specific services (comma-separated, e.g., cloudwatch,cloudtrail)")`
  - [x] After `services.AllServices()` call, apply `services.FilterServices()` with flag values
  - [x] Add "Scanning N of M services..." message when filters are active
  - [x] Add "No services matched filter criteria" error + `os.Exit(1)` when filtered list is empty

- [x] Task 3: Write unit tests (AC: 12, 14)
  - [x] Create `cmd/awtest/services/service_filter_test.go`
  - [x] Test: no filter returns all services
  - [x] Test: include-only with exact match (e.g., "s3" includes only S3 services)
  - [x] Test: include-only with case-insensitive match (e.g., "S3", "s3")
  - [x] Test: include-only with partial match (e.g., "cognito" matches cognitoidentity + cognitouserpools)
  - [x] Test: exclude-only removes matching services
  - [x] Test: include+exclude combination (include "s3,ec2,iam", exclude "iam" -> s3+ec2 only)
  - [x] Test: no matches returns empty slice
  - [x] Test: empty include/exclude strings return all services

- [x] Task 4: Build and verify (AC: 13, 14, 15)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### Service Name Matching Strategy

The AWSService.Name field uses the format `"service:APICall"` (e.g., `"s3:ListBuckets"`, `"ec2:DescribeInstances"`, `"cognitouserpools:ListUserPools"`). The service prefix before the colon maps to the AWS IAM service namespace.

**Matching approach:** Extract the service prefix from `AWSService.Name` by splitting on `:` and taking the first part. Compare lowercased filter input against this prefix.

**Partial matching:** A filter value like `"cognito"` should match any service whose prefix contains `"cognito"` as a substring. This enables users to type `"cognito"` instead of remembering both `"cognitoidentity"` and `"cognitouserpools"`.

**Current service prefixes** (extracted from AWSService.Name fields across all 50+ registered services):
`acm`, `amplify`, `apigateway`, `appsync`, `batch`, `cloudformation`, `cloudfront`, `cloudtrail`, `cloudwatch`, `cloudwatchlogs`, `codepipeline`, `cognito-idp`, `cognitoidentity`, `config`, `dynamodb`, `ec2`, `ecs`, `efs`, `eks`, `elasticache`, `elasticbeanstalk`, `eventbridge`, `glacier`, `glue`, `iam`, `iot`, `ivs`, `ivsChat`, `ivsRealtime`, `kms`, `lambda`, `rds`, `redshift`, `rekognition`, `route53`, `s3`, `secretsmanager`, `ses`, `sns`, `sqs`, `ssm`, `states`, `sts`, `transcribe`, `waf`

**IMPORTANT -- Shared prefixes:** Some prefixes map to multiple AWSService entries:
- `ec2` -- used by BOTH EC2 instances (`ec2:DescribeInstances`) and VPC (`ec2:DescribeVpcs`)
- `ecs` -- used by BOTH ECS clusters (`ecs:ListClusters`) and Fargate tasks (`ecs:ListFargateTasks`)
- `apigateway` -- has 3 entries (RestApis, GetApiKeys, GetDomainNames)
- `dynamodb` -- has 3 entries (ListTables, ListBackups, ListExports)
- `cloudtrail` -- has 2 entries (DescribeTrails, ListTrails)
- `cloudwatchlogs` -- has 2 entries (DescribeLogGroupsAndStreams, ListMetrics)
- `ivs`, `ivsChat`, `ivsRealtime` -- separate prefixes for IVS services
- `transcribe` -- has 4 entries

The filter should match ALL AWSService entries that share the matched prefix. When user says `-services=ec2`, all `ec2:*` entries should be included.

**IMPORTANT -- Case sensitivity in prefixes:** Most prefixes are lowercase, but note `ivsChat` and `ivsRealtime` use camelCase. The filter MUST lowercase both the input AND the prefix for comparison.

### CRITICAL: How FilterServices Integrates with main.go

Current main.go scan loop (lines ~150-157):
```go
for _, service := range services.AllServices() {
    if !*quiet {
        fmt.Fprintf(os.Stderr, "Scanning %s...\n", service.Name)
    }
    output, err := service.Call(sess)
    serviceResults := service.Process(output, err, *debug)
    results = append(results, serviceResults...)
}
```

**Required change:** Insert filtering between `services.AllServices()` and the loop:
```go
allSvcs := services.AllServices()
filteredSvcs := services.FilterServices(allSvcs, includeServices, excludeServices)
if len(filteredSvcs) == 0 {
    fmt.Fprintln(os.Stderr, "No services matched filter criteria")
    os.Exit(1)
}
if includeServices != "" || excludeServices != "" {
    fmt.Fprintf(os.Stderr, "Scanning %d of %d services...\n", len(filteredSvcs), len(allSvcs))
}
for _, service := range filteredSvcs {
    // ... existing loop body
}
```

### CRITICAL: Use Standard Go flag Package

The project uses the standard Go `flag` package (NOT cobra, NOT urfave/cli). All flags are defined in `main.go` using `flag.StringVar()`, `flag.BoolVar()`, etc. Do NOT add any external CLI framework dependency.

### CRITICAL: Do NOT Modify AWSService Struct

The filtering operates on the existing `[]types.AWSService` slice returned by `AllServices()`. Do NOT modify the `AWSService` struct, `types.go`, or any service implementation. The filter is purely a post-processing step on the service list.

### File Placement: services/ Package, NOT main Package

The architecture doc specifies `cmd/awtest/services/service_filter.go` (in the `services` package). This keeps filtering logic with the service registry. The function signature is:

```go
// FilterServices filters the given service list based on include/exclude patterns.
// Include filter is applied first, then exclude filter removes from the included set.
// Service names are matched case-insensitively with partial matching support.
func FilterServices(allServices []types.AWSService, include, exclude string) []types.AWSService
```

### Edge Cases to Handle

1. **Empty strings** -- `include=""` and `exclude=""` returns all services (no filtering)
2. **Whitespace in CSV** -- `"s3, ec2 , iam"` should work (trim each value)
3. **Duplicate filter values** -- `"s3,s3"` should not cause duplicate results
4. **Overlapping partial matches** -- `"elastic"` matches `elasticache`, `elasticbeanstalk`, `elasticfilesystem` -- all should be included
5. **Include then exclude same service** -- `include=s3,ec2` + `exclude=s3` results in only ec2
6. **STS service** -- STS is always first in AllServices() for credential validation. Consider: should `-services=ec2` still include STS? The architecture doesn't specify. Implement as strict filter (only include what's specified). If STS is needed for auth, the user should include it explicitly or the credential check happens before the service loop.

### Go Version Constraint

Go 1.19 -- no generics, no `slices` package, no `strings.Cut()`. Use traditional for loops and string operations.

### Project Structure Notes

- **File to CREATE:** `cmd/awtest/services/service_filter.go` (in existing `services` package)
- **File to CREATE:** `cmd/awtest/services/service_filter_test.go` (co-located test)
- **File to MODIFY:** `cmd/awtest/main.go` (add flags + filter integration)
- Aligns with architecture doc's specified file locations
- No new packages needed -- filter lives in existing `services` package

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 3.1: Service Filtering (Include/Exclude Services)]
- [Source: _bmad-output/planning-artifacts/architecture.md#Service Filtering Pattern -- lines 1398-1450]
- [Source: _bmad-output/planning-artifacts/architecture.md#Configuration Flag Pattern -- lines 650-714]
- [Source: _bmad-output/planning-artifacts/architecture.md#Testing Patterns -- lines 1457-1559]
- [Source: _bmad-output/planning-artifacts/architecture.md#Naming Conventions -- lines 1060-1139]
- [Source: cmd/awtest/main.go -- current CLI flag definitions and scan loop]
- [Source: cmd/awtest/services/services.go -- AllServices() registry with 50 services]
- [Source: cmd/awtest/types/types.go -- AWSService struct definition (Name, Call, Process, ModuleName)]

### Previous Epic Intelligence (Epic 2 Learnings)

- **sess.Copy() pattern** -- not directly relevant to this story (no AWS API calls), but establishes code quality expectations
- **Table-driven tests** are the standard -- use `t.Run(tt.name, ...)` subtests
- **All display fields in BOTH print AND data structures** -- ensure filter messages appear in stderr (not stdout) to avoid corrupting structured output
- **go vet must pass clean** -- check all new code

### Git Intelligence

Recent commits show Epic 2 completion pattern: each story adds a new service package + registers in AllServices(). This story is different -- it's the first "infrastructure" story that modifies the scan orchestration rather than adding a new service. Take care not to break the existing 50 service registrations.

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None required - clean implementation with all tests passing on first run.

### Completion Notes List

- Created `FilterServices()` with include/exclude support, case-insensitive and partial matching
- Helper functions: `parseServiceList()`, `extractServiceName()`, `matchesFilter()`, `warnUnrecognized()`
- Added `-services` and `-exclude-services` CLI flags to main.go
- Integrated filtering between `AllServices()` and the scan loop
- Added "Scanning N of M services..." message when filters active
- Added "No services matched filter criteria" error with `os.Exit(1)` for empty results
- Warning output for unrecognized service names (stderr, non-blocking)
- 12 test functions covering: no filter, empty strings, exact match, case-insensitive, partial match, multiple EC2 entries, exclude-only, include+exclude combo, no matches, elastic partial, camelCase prefix, whitespace in CSV, parseServiceList helper, extractServiceName helper

### Change Log

- 2026-03-05: Implemented service filtering with include/exclude support (Story 3.1)

### File List

- cmd/awtest/services/service_filter.go (NEW)
- cmd/awtest/services/service_filter_test.go (NEW)
- cmd/awtest/main.go (MODIFIED)
