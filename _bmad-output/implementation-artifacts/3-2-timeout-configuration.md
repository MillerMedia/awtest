# Story 3.2: Timeout Configuration

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional with time-constrained engagements**,
I want **to set a maximum scan timeout**,
so that **I can ensure scans complete within my engagement window and don't hang indefinitely on slow API calls**.

## Acceptance Criteria

1. **AC1:** Add `-timeout` flag accepting duration value (e.g., `-timeout=5m`, `-timeout=300s`) with default 5 minutes (FR35)
2. **AC2:** Implement timeout context in `main.go` using `context.WithTimeout()`
3. **AC3:** Pass timeout context to each service `Call()` method
4. **AC4:** Modify `AWSService.Call` signature to accept `context.Context` parameter: `func(context.Context, *session.Session) (interface{}, error)`
5. **AC5:** Update ALL existing service `Call()` implementations to accept context and use `WithContext` AWS SDK v1 API variants
6. **AC6:** When timeout is reached, cancel remaining service enumerations gracefully
7. **AC7:** Display timeout warning: "Scan timeout reached after [duration]. N services not scanned."
8. **AC8:** List services that were not scanned due to timeout
9. **AC9:** Ensure partial results from completed services are still output
10. **AC10:** Exit code 0 if timeout occurs (not an error, just incomplete scan)
11. **AC11:** Handle context cancellation in AWS SDK calls -- terminate API calls cleanly
12. **AC12:** Write unit tests for timeout logic covering: timeout before first service, timeout mid-scan, timeout after all services, no timeout (default behavior)
13. **AC13:** `go build ./cmd/awtest` compiles successfully
14. **AC14:** `go test ./cmd/awtest/...` passes (all existing + new tests)
15. **AC15:** `go vet ./cmd/awtest/...` passes clean
16. **AC16:** FR35 requirement fulfilled: Users can set maximum scan timeout duration

## Tasks / Subtasks

- [x] Task 1: Modify AWSService type to accept context (AC: 4)
  - [x] Edit `cmd/awtest/types/types.go` -- change `Call` field from `func(*session.Session) (interface{}, error)` to `func(context.Context, *session.Session) (interface{}, error)`
  - [x] Add `"context"` import to types.go

- [x] Task 2: Update ALL 46 service Call() implementations (AC: 5, 11)
  - [x] Update each `calls.go` across all 46 service packages to accept `context.Context` as first parameter
  - [x] Replace AWS SDK API calls with `WithContext` variants (e.g., `svc.ListBuckets(input)` → `svc.ListBucketsWithContext(ctx, input)`)
  - [x] Add `"context"` import to each updated file
  - [x] Service packages to update (all under `cmd/awtest/services/`):
    - acm, amplify, apigateway, appsync, batch, cloudformation, cloudfront, cloudtrail, cloudwatch, cloudwatchlogs
    - codepipeline, cognito, cognitoidentity, config, dynamodb, ec2, ecs, efs, eks, elasticache
    - elasticbeanstalk, eventbridge, fargate, glacier, glue, iam, iot, ivs, ivschat, ivsrealtime
    - kms, lambda, rds, redshift, rekognition, route53, s3, secretsmanager, ses, sns
    - sqs, ssm, states, sts, transcribe, vpc, waf (plus any others discovered)

- [x] Task 3: Add timeout flag and context to main.go (AC: 1, 2, 3, 6, 7, 8, 9, 10)
  - [x] Add `timeout` flag: `flag.Duration("timeout", 5*time.Minute, "Maximum scan timeout")`
  - [x] Create timeout context: `ctx, cancel := context.WithTimeout(context.Background(), *timeout)`
  - [x] Modify scan loop to check context before each service call
  - [x] Pass `ctx` to each `service.Call(ctx, sess)` invocation
  - [x] Track which services were not scanned due to timeout
  - [x] When timeout reached: print warning message with duration and count of unscanned services
  - [x] List unscanned service names to stderr
  - [x] Ensure all completed results are still passed to the formatter/output
  - [x] Exit code 0 on timeout (partial results are valid)

- [x] Task 4: Write unit tests (AC: 12, 14)
  - [x] Create `cmd/awtest/timeout_test.go` with timeout-specific tests
  - [x] Test: context cancellation before first service returns immediately
  - [x] Test: context cancellation mid-scan preserves partial results
  - [x] Test: no timeout (context.Background()) processes all services
  - [x] Test: already-expired context skips all services

- [x] Task 5: Build and verify (AC: 13, 14, 15)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: AWSService.Call Signature Change (Breaking Change to All Services)

The current `Call` field signature in `cmd/awtest/types/types.go:39`:
```go
Call: func(*session.Session) (interface{}, error)
```

Must change to:
```go
Call: func(context.Context, *session.Session) (interface{}, error)
```

This is a **compile-breaking change** that requires updating ALL 46 service `calls.go` files. The compiler will catch every missed file -- the project will not compile until all are updated. This is actually a safety feature: you cannot forget a file.

### AWS SDK v1 WithContext Pattern

AWS SDK v1 (`github.com/aws/aws-sdk-go v1.44.266`) supports context via `WithContext` method variants on every API call. Every SDK call like `svc.ListBuckets(input)` has a corresponding `svc.ListBucketsWithContext(ctx, input)` variant.

**Example transformation (S3):**
```go
// BEFORE
Call: func(sess *session.Session) (interface{}, error) {
    svc := s3.New(sess)
    output, err := svc.ListBuckets(&s3.ListBucketsInput{})
    return output, err
},

// AFTER
Call: func(ctx context.Context, sess *session.Session) (interface{}, error) {
    svc := s3.New(sess)
    output, err := svc.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
    return output, err
},
```

### CRITICAL: Main.go Timeout Loop Pattern

The architecture doc specifies checking context at the loop level AND passing context into service calls:

```go
timeout := flag.Duration("timeout", 5*time.Minute, "Maximum scan timeout")
// ... after flag.Parse()

ctx, cancel := context.WithTimeout(context.Background(), *timeout)
defer cancel()

var skippedServices []string
for _, service := range filteredSvcs {
    select {
    case <-ctx.Done():
        // Collect remaining unscanned services
        skippedServices = append(skippedServices, service.Name)
        continue
    default:
        if !*quiet {
            fmt.Fprintf(os.Stderr, "Scanning %s...\n", service.Name)
        }
        output, err := service.Call(ctx, sess)
        serviceResults := service.Process(output, err, *debug)
        results = append(results, serviceResults...)
    }
}

if len(skippedServices) > 0 {
    fmt.Fprintf(os.Stderr, "\nScan timeout reached after %s. %d services not scanned:\n", *timeout, len(skippedServices))
    for _, name := range skippedServices {
        fmt.Fprintf(os.Stderr, "  - %s\n", name)
    }
}
```

**Key detail:** Once `ctx.Done()` fires, all remaining services in the loop collect into `skippedServices` via `continue`. The `select` with `default` is non-blocking -- it only triggers when context is already cancelled.

**Also key:** If a service call is in-progress when the timeout fires, the `WithContext` call will return with a `context.DeadlineExceeded` error. The `Process()` function will handle this as an error for that service (existing error handling). The next iteration's `select` will then catch `ctx.Done()` and skip remaining services.

### CRITICAL: Exit Code Behavior

The epic specifies **exit code 0** on timeout. This is different from the architecture doc's `os.Exit(1)`. Follow the epic: timeout produces partial results, which is valid output. Do NOT `os.Exit(1)` on timeout.

### CRITICAL: Service Files Are Mechanical Updates

Updating 46 service files is tedious but mechanical. For each `calls.go`:
1. Add `"context"` to imports
2. Change `func(sess *session.Session)` to `func(ctx context.Context, sess *session.Session)`
3. Replace every SDK API call `svc.Method(input)` with `svc.MethodWithContext(ctx, input)`

**Watch for services with multiple API calls in a single Call function** -- some services make several SDK calls (e.g., EC2 describes instances across regions, DynamoDB lists tables then describes each). ALL SDK calls in the function must get the `WithContext` variant.

**Watch for services using `sess.Copy()`** -- several services copy the session for region iteration (e.g., EC2, RDS). The context should be passed to the SDK calls made with the copied session too.

### CRITICAL: Do NOT Change Process() Signature

Only `Call()` gets context. `Process()` does not make API calls and does not need context. Do NOT modify the `Process` field signature.

### CRITICAL: Do NOT Introduce External Dependencies

Continue using the standard Go `flag` package. Do NOT add cobra, urfave/cli, or any other CLI framework. The `context` package is part of Go's standard library.

### Go Version Constraint

Go 1.19 -- no generics, no `slices` package, no `strings.Cut()`. Use traditional for loops and string operations. `context` package is available in Go 1.19.

### Services That Make Multiple SDK Calls Per Call()

These services require special attention as they have multiple API calls within a single `Call` function that ALL need `WithContext`:
- **dynamodb**: ListTables, then DescribeTable for each, ListBackups, ListExports
- **ec2**: DescribeInstances (may iterate regions)
- **ecs**: ListClusters + DescribeServices
- **iam**: Multiple IAM calls (users, roles, policies)
- **cloudtrail**: DescribeTrails + GetTrailStatus
- **cloudwatchlogs**: DescribeLogGroups + DescribeLogStreams
- **transcribe**: Multiple list calls
- **vpc**: DescribeVpcs + DescribeSubnets + DescribeSecurityGroups etc.

For these, ensure EVERY SDK call gets the `WithContext` variant.

### Handling context.DeadlineExceeded in Process()

When a `WithContext` call is cancelled, it returns an error wrapping `context.DeadlineExceeded`. The existing `Process()` error handling will treat this as a service error (likely printing a debug message or recording an error result). This is acceptable behavior -- no special handling needed in `Process()`.

### Project Structure Notes

- **File to MODIFY:** `cmd/awtest/types/types.go` (Call signature change)
- **Files to MODIFY:** All 46 `calls.go` files under `cmd/awtest/services/*/`
- **File to MODIFY:** `cmd/awtest/main.go` (add timeout flag + context loop)
- **File to CREATE:** `cmd/awtest/services/timeout_test.go` (or integrate into existing test structure)
- No new packages or dependencies needed

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 3.2: Timeout Configuration (line 844)]
- [Source: _bmad-output/planning-artifacts/architecture.md#Timeout Implementation (lines 691-708)]
- [Source: _bmad-output/planning-artifacts/architecture.md#Configuration Flag Pattern (lines 650-659)]
- [Source: cmd/awtest/types/types.go:39 -- AWSService.Call current signature]
- [Source: cmd/awtest/main.go:166-173 -- current scan loop]
- [Source: cmd/awtest/services/services.go -- AllServices() registry]

### Previous Story Intelligence (Story 3.1 Learnings)

- Story 3.1 successfully modified `main.go` scan loop to add service filtering between `AllServices()` and the loop
- The `filteredSvcs` variable is already in place -- build timeout on top of the filtered list
- All output messages go to **stderr** (not stdout) to avoid corrupting structured output -- follow same pattern for timeout messages
- Flag variables use pointer pattern: `flag.Duration()` returns `*time.Duration`
- Table-driven tests with `t.Run()` subtests are the established testing pattern
- `go vet` must pass clean on all new code

### Git Intelligence

Recent commits:
- `c491f87` Add service filtering with include/exclude flags (Story 3.1)
- `b1fde48` Mark Story 2.11 as done
- `535cf39` Add VPC infrastructure service enumeration (Story 2.11)
- `712d5ea` Mark Story 2.10 as done
- `e55e8c7` Add Systems Manager SSM parameters service enumeration (Story 2.10)

Story 3.1 commit shows the pattern for modifying main.go's scan infrastructure. This story continues that pattern with deeper changes (touching all service files).

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

### Completion Notes List

- Task 1: Changed `AWSService.Call` signature in `types.go` to accept `context.Context` as first parameter
- Task 2: Updated all 46 service `calls.go` files: added `"context"` import, changed Call signatures, replaced SDK API calls with `WithContext` variants. Updated EC2 helper function `getInstanceUserData` to accept and pass context.
- Code Review Fix (High): Resolved AC11 gap -- S3, IAM, SQS, STS Process() methods now receive context via output map and use WithContext for all SDK calls. Context is passed through the existing `interface{}` output parameter without changing Process() signature.
- Task 3: Added `-timeout` flag (default 5m) to main.go. Implemented `context.WithTimeout` scan loop with graceful cancellation. Extracted `scanServices()` function for testability. Timeout messages go to stderr; exit code 0 on timeout; partial results preserved.
- Task 4: Created 5 timeout unit tests: NoTimeout, AlreadyExpiredContext, TimeoutMidScan, TimeoutBeforeFirstService, PartialResultsPreserved. All pass.
- Task 5: `go build`, `go test ./cmd/awtest/...`, `go vet ./cmd/awtest/...` all pass clean.

### File List

- cmd/awtest/types/types.go (modified - Call signature change)
- cmd/awtest/main.go (modified - timeout flag, context loop, scanServices function)
- cmd/awtest/timeout_test.go (created - timeout unit tests)
- cmd/awtest/services/amplify/calls.go (modified)
- cmd/awtest/services/apigateway/calls.go (modified)
- cmd/awtest/services/appsync/calls.go (modified)
- cmd/awtest/services/batch/calls.go (modified)
- cmd/awtest/services/certificatemanager/calls.go (modified)
- cmd/awtest/services/cloudformation/calls.go (modified)
- cmd/awtest/services/cloudfront/calls.go (modified)
- cmd/awtest/services/cloudtrail/calls.go (modified)
- cmd/awtest/services/cloudwatch/calls.go (modified)
- cmd/awtest/services/codepipeline/calls.go (modified)
- cmd/awtest/services/cognitoidentity/calls.go (modified)
- cmd/awtest/services/cognitouserpools/calls.go (modified)
- cmd/awtest/services/config/calls.go (modified)
- cmd/awtest/services/dynamodb/calls.go (modified)
- cmd/awtest/services/ec2/calls.go (modified)
- cmd/awtest/services/ecs/calls.go (modified)
- cmd/awtest/services/efs/calls.go (modified)
- cmd/awtest/services/eks/calls.go (modified)
- cmd/awtest/services/elasticache/calls.go (modified)
- cmd/awtest/services/elasticbeanstalk/calls.go (modified)
- cmd/awtest/services/eventbridge/calls.go (modified)
- cmd/awtest/services/fargate/calls.go (modified)
- cmd/awtest/services/glacier/calls.go (modified)
- cmd/awtest/services/glue/calls.go (modified)
- cmd/awtest/services/iam/calls.go (modified)
- cmd/awtest/services/iot/calls.go (modified)
- cmd/awtest/services/ivs/calls.go (modified)
- cmd/awtest/services/ivschat/calls.go (modified)
- cmd/awtest/services/ivsrealtime/calls.go (modified)
- cmd/awtest/services/kms/calls.go (modified)
- cmd/awtest/services/lambda/calls.go (modified)
- cmd/awtest/services/rds/calls.go (modified)
- cmd/awtest/services/redshift/calls.go (modified)
- cmd/awtest/services/rekognition/calls.go (modified)
- cmd/awtest/services/route53/calls.go (modified)
- cmd/awtest/services/s3/calls.go (modified)
- cmd/awtest/services/secretsmanager/calls.go (modified)
- cmd/awtest/services/ses/calls.go (modified)
- cmd/awtest/services/sns/calls.go (modified)
- cmd/awtest/services/sqs/calls.go (modified)
- cmd/awtest/services/stepfunctions/calls.go (modified)
- cmd/awtest/services/sts/calls.go (modified)
- cmd/awtest/services/systemsmanager/calls.go (modified)
- cmd/awtest/services/transcribe/calls.go (modified)
- cmd/awtest/services/vpc/calls.go (modified)
- cmd/awtest/services/waf/calls.go (modified)

### Change Log

- 2026-03-05: Implemented timeout configuration (Story 3.2) - Added -timeout flag, context propagation to all 46 services, timeout-aware scan loop, and 5 unit tests
- 2026-03-05: Addressed code review findings - Resolved 1 High severity issue (AC11 Process context leak in S3/IAM/SQS/STS)
