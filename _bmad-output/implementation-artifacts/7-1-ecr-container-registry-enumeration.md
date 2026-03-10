# Story 7.1: ECR Container Registry Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate ECR container repositories, images, and repository policies,
So that I can discover container images with embedded secrets and overly permissive registry access.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/ecr/` directory with `calls.go` implementing ECR service enumeration.

2. **AC2:** Implement `ecr:DescribeRepositories` API call — iterates all regions in `types.Regions`, creates ECR client per region, calls `DescribeRepositoriesWithContext`, aggregates `[]*ecr.Repository` results. Each repository listed with repository name and URI.

3. **AC3:** Implement `ecr:ListImages` API call — iterates all regions, calls `DescribeRepositoriesWithContext` to discover repos, then for each repository calls `ListImagesWithContext` to enumerate images with tags and digests.

4. **AC4:** Implement `ecr:GetRepositoryPolicy` API call — iterates all regions, calls `DescribeRepositoriesWithContext` to discover repos, then for each repository calls `GetRepositoryPolicyWithContext` to retrieve access permissions. Handle `RepositoryPolicyNotFoundException` gracefully (skip repos without custom policies — not an error).

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without ECR access, ECR is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register ECR service in `services/services.go` `AllServices()` function in alphabetical order (after `dynamodb`, before `ecs`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid repositories/images/policies, empty results, access denied errors, nil field handling.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/ecr/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `ecr:DescribeRepositories` (AC: 1, 2, 5, 9)
  - [x] Create directory `cmd/awtest/services/ecr/`
  - [x] Create `calls.go` with `package ecr`
  - [x] Define `var ECRCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"ecr:DescribeRepositories"`
  - [x] Call: iterate `types.Regions`, set `sess.Config.Region`, create `ecr.New(sess)`, call `svc.DescribeRepositoriesWithContext(ctx, &ecr.DescribeRepositoriesInput{})`, aggregate `output.Repositories`
  - [x] Process: handle error → `utils.HandleAWSError`, type-assert `[]*ecr.Repository`, extract `RepositoryName`, `RepositoryUri`, `RepositoryArn` with nil checks, build `ScanResult` with ServiceName=`"ECR"`, ResourceType=`"repository"`
  - [x] `utils.PrintResult` format: `"ECR Repository: %s (URI: %s)"` with `utils.ColorizeItem(repoName)`

- [x] Task 2: Implement `ecr:ListImages` (AC: 3, 5, 9)
  - [x] Implement second entry: Name `"ecr:ListImages"`
  - [x] Call: iterate regions → `DescribeRepositoriesWithContext` to get repo names → for each repo, call `svc.ListImagesWithContext(ctx, &ecr.ListImagesInput{RepositoryName: aws.String(repoName)})` → build result structs pairing repo name with image identifiers
  - [x] Handle per-repo errors gracefully: if ListImages fails for one repo (access denied on specific repo), continue to next repo
  - [x] Process: type-assert output, extract `ImageTag` and `ImageDigest` per image, build `ScanResult` with ServiceName=`"ECR"`, ResourceType=`"image"`, ResourceName=imageTag (or digest if no tag)
  - [x] `utils.PrintResult` format: `"ECR Image: %s:%s (Digest: %s)"` showing repo:tag

- [x] Task 3: Implement `ecr:GetRepositoryPolicy` (AC: 4, 5, 9)
  - [x] Implement third entry: Name `"ecr:GetRepositoryPolicy"`
  - [x] Call: iterate regions → `DescribeRepositoriesWithContext` to get repo names → for each repo, call `svc.GetRepositoryPolicyWithContext(ctx, &ecr.GetRepositoryPolicyInput{RepositoryName: aws.String(repoName)})`
  - [x] Handle `RepositoryPolicyNotFoundException`: this is NOT an error — it means no custom policy is set. Skip silently and continue to next repo.
  - [x] Handle per-repo access denied: continue to next repo
  - [x] Process: type-assert output, extract `PolicyText` and associated `RepositoryName`, build `ScanResult` with ServiceName=`"ECR"`, ResourceType=`"repository-policy"`, ResourceName=repoName
  - [x] `utils.PrintResult` format: `"ECR Repository Policy: %s"` with `utils.ColorizeItem(repoName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/ecr"` to `services/services.go` (alphabetical in imports: after `ec2`, before `ecs`)
  - [x] Add `allServices = append(allServices, ecr.ECRCalls...)` after `ec2.EC2Calls...` and before `ecs.ECSCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/ecr/calls_test.go`
  - [x] Test `DescribeRepositories` Process: valid repos, empty results, access denied error, nil fields
  - [x] Test `ListImages` Process: valid images with tags, images without tags (digest only), empty images, error handling
  - [x] Test `GetRepositoryPolicy` Process: valid policy text, no policy (empty), error handling
  - [x] Use table-driven tests following ACM/certificatemanager pattern
  - [x] Access Process via `ECRCalls[0].Process`, `ECRCalls[1].Process`, `ECRCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/ecr/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Follow Established Service Patterns Exactly

This is the **first story in Epic 7** (Critical Security Service Expansion / Phase 2 Epic 2). Follow the existing service implementation patterns precisely — use `certificatemanager/calls.go` as the reference implementation.

### Service Implementation Pattern

Every service follows this exact structure (from 46 existing services):

```go
package ecr

import (
	"context"
	"fmt"
	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"github.com/MillerMedia/awtest/cmd/awtest/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"time"
)

var ECRCalls = []types.AWSService{
	{
		Name: "ecr:DescribeRepositories",
		Call: func(ctx context.Context, sess *session.Session) (interface{}, error) {
			// Regional iteration pattern
			var allRepos []*ecr.Repository
			for _, region := range types.Regions {
				sess.Config.Region = aws.String(region)
				svc := ecr.New(sess)
				output, err := svc.DescribeRepositoriesWithContext(ctx, &ecr.DescribeRepositoriesInput{})
				if err != nil {
					return nil, err
				}
				allRepos = append(allRepos, output.Repositories...)
			}
			return allRepos, nil
		},
		Process: func(output interface{}, err error, debug bool) []types.ScanResult {
			// Error handling, type assertion, nil-safe field extraction
			// ... (see ACM pattern)
		},
		ModuleName: types.DefaultModuleName,
	},
	// ... additional entries for ListImages, GetRepositoryPolicy
}
```

### ECR SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/ecr`

**API Methods:**

1. **DescribeRepositories:**
   - `svc.DescribeRepositoriesWithContext(ctx, &ecr.DescribeRepositoriesInput{})` → `*ecr.DescribeRepositoriesOutput`
   - `.Repositories` → `[]*ecr.Repository`
   - Repository fields: `RepositoryName *string`, `RepositoryUri *string`, `RepositoryArn *string`, `RegistryId *string`, `CreatedAt *time.Time`

2. **ListImages:**
   - `svc.ListImagesWithContext(ctx, &ecr.ListImagesInput{RepositoryName: aws.String(name)})` → `*ecr.ListImagesOutput`
   - `.ImageIds` → `[]*ecr.ImageIdentifier`
   - ImageIdentifier fields: `ImageTag *string`, `ImageDigest *string`
   - Requires `RepositoryName` input — must call DescribeRepositories first

3. **GetRepositoryPolicy:**
   - `svc.GetRepositoryPolicyWithContext(ctx, &ecr.GetRepositoryPolicyInput{RepositoryName: aws.String(name)})` → `*ecr.GetRepositoryPolicyOutput`
   - Output fields: `PolicyText *string`, `RegistryId *string`, `RepositoryName *string`
   - Returns `RepositoryPolicyNotFoundException` if repo has no custom policy — this is expected, not an error

**No new dependencies needed** — ECR is part of `aws-sdk-go v1.44.266` already in go.mod.

### Variable & Naming Conventions

- **Package:** `ecr` (directory: `cmd/awtest/services/ecr/`)
- **Exported variable:** `ECRCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"ecr:DescribeRepositories"`, `"ecr:ListImages"`, `"ecr:GetRepositoryPolicy"`
- **ScanResult.ServiceName:** `"ECR"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"repository"`, `"image"`, `"repository-policy"` (lowercase singular)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)

### Registration Order in services.go

Insert alphabetically — `ecr` comes after `ec2`, before `ecs`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/ecr"

// In AllServices():
allServices = append(allServices, ec2.EC2Calls...)
allServices = append(allServices, ecr.ECRCalls...)     // NEW
allServices = append(allServices, ecs.ECSCalls...)
```

### Multi-Step Call Pattern (ListImages and GetRepositoryPolicy)

For calls that depend on DescribeRepositories results, each Call function independently discovers repos first:

```go
// Pattern for dependent calls (ListImages, GetRepositoryPolicy):
Call: func(ctx context.Context, sess *session.Session) (interface{}, error) {
    type repoImage struct {
        RepoName string
        ImageTag string
        ImageDigest string
    }
    var allImages []repoImage

    for _, region := range types.Regions {
        sess.Config.Region = aws.String(region)
        svc := ecr.New(sess)

        // Step 1: Discover repos
        repoOutput, err := svc.DescribeRepositoriesWithContext(ctx, &ecr.DescribeRepositoriesInput{})
        if err != nil {
            return nil, err
        }

        // Step 2: For each repo, call secondary API
        for _, repo := range repoOutput.Repositories {
            if repo.RepositoryName == nil {
                continue
            }
            imagesOutput, err := svc.ListImagesWithContext(ctx, &ecr.ListImagesInput{
                RepositoryName: repo.RepositoryName,
            })
            if err != nil {
                continue // Skip repo on error, continue to next
            }
            for _, img := range imagesOutput.ImageIds {
                // Build result struct...
            }
        }
    }
    return allImages, nil
},
```

**Key:** Per-repo errors in secondary calls should be handled gracefully — `continue` to next repo, don't abort the entire scan.

### GetRepositoryPolicy Error Handling

`GetRepositoryPolicy` returns `RepositoryPolicyNotFoundException` when a repo has no custom policy. This is **expected** and should be handled differently from access denied:

```go
if err != nil {
    if awsErr, ok := err.(awserr.Error); ok {
        if awsErr.Code() == "RepositoryPolicyNotFoundException" {
            continue // No custom policy — skip, not an error
        }
    }
    continue // Other errors — skip repo, continue
}
```

Import `"github.com/aws/aws-sdk-go/aws/awserr"` for error type assertion.

### Return Types for Multi-Step Calls

Since Call must return `(interface{}, error)`, define local structs for complex results:

```go
// For ListImages:
type ecrImage struct {
    RepoName    string
    ImageTag    string
    ImageDigest string
}

// For GetRepositoryPolicy:
type ecrPolicy struct {
    RepoName   string
    PolicyText string
}
```

Process functions type-assert against slices of these structs (e.g., `[]ecrImage`).

### Testing Pattern

Follow the ACM/certificatemanager test pattern — test Process() functions only with pre-built mock data:

```go
func TestDescribeRepositoriesProcess(t *testing.T) {
    process := ECRCalls[0].Process
    // Table-driven tests with valid repos, empty, errors, nil fields
}

func TestListImagesProcess(t *testing.T) {
    process := ECRCalls[1].Process
    // Test with images with tags, digest-only images, empty results
}

func TestGetRepositoryPolicyProcess(t *testing.T) {
    process := ECRCalls[2].Process
    // Test with valid policy text, empty/nil policy, errors
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `ecr/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** use `sess.Copy()` — existing services use `sess.Config.Region = aws.String(region)` pattern (mutate in loop)
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** fail the entire scan if one repo's ListImages or GetRepositoryPolicy fails — continue to next repo
- **DO NOT** treat `RepositoryPolicyNotFoundException` as an error — it's normal for repos without custom policies
- **DO NOT** use generics (Go 1.19 does not support generics)

### Previous Story Intelligence

**From Story 2.1 (ACM — first service addition in Phase 1 Epic 2):**
- Process() tests cover: valid resources, empty results, access denied errors, nil field handling
- Tests access Process via `ECRCalls[0].Process` pattern
- Single file per service: `calls.go` + `calls_test.go`
- Error result: `return []types.ScanResult{{ServiceName: "ECR", MethodName: "ecr:DescribeRepositories", Error: err, Timestamp: time.Now()}}`
- Empty Details map is fine: `Details: map[string]interface{}{}`

**From Story 6.5 (last completed story — concurrent progress):**
- Tests use stdlib `testing` only (no testify in `cmd/awtest/` tests)
- All tests should pass with `go test -race ./...`
- Output buffering is automatic in concurrent mode via `utils.ConcurrentMode`
- No changes needed in services for concurrent support

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `c9f20f0` — Add [hit] severity with green coloring and accessible methods list in scan summary
- `4cbaf04` — Buffer inline output during concurrent scans to fix progress interleaving
- `3fbde26` — Add concurrent progress reporting with TTY detection (Story 6.5)

### FRs Covered

- **FR85:** System enumerates ECR container repositories, images, and repository policies

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/ecr/
├── calls.go            # ECR service implementation (3 AWSService entries)
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
go.mod                           # AWS SDK already includes ECR package
```

### References

- [Source: epics-phase2.md#Story 2.1: ECR Container Registry Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR85] — ECR enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/certificatemanager/calls.go] — Reference implementation for regional service pattern
- [Source: cmd/awtest/services/certificatemanager/calls_test.go] — Reference test pattern (Process-only tests)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes ECR package)
- [Source: 2-1-certificate-manager-acm-service-enumeration.md] — Previous service story pattern
- [Source: 6-5-concurrent-progress-reporting.md] — Latest story learnings (testing patterns, concurrent integration)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no blockers.

### Completion Notes List

- Implemented ECR service with 3 AWSService entries: DescribeRepositories, ListImages, GetRepositoryPolicy
- Used local structs `ecrImage` and `ecrPolicy` for multi-step call results (ListImages and GetRepositoryPolicy both discover repos first, then query per-repo)
- RepositoryPolicyNotFoundException handled gracefully via awserr type assertion — skips repos without custom policies
- Per-repo errors in ListImages and GetRepositoryPolicy handled with `continue` — doesn't abort entire scan
- All Process functions follow ACM pattern: error handling, type assertion, nil-safe pointer dereferencing
- Registered in AllServices() alphabetically between ec2 and ecs
- 12 table-driven test cases (3 test functions with subtests) covering valid data, empty results, access denied errors, nil fields, digest-only images, empty policy text
- No sync primitives — concurrency-unaware per NFR57
- Full test suite passes with race detector (0 failures, 0 regressions)
- ✅ Resolved review finding [High]: Added pagination (NextToken loop) to DescribeRepositories in all 3 Call functions
- ✅ Resolved review finding [High]: Added pagination (NextToken loop) to ListImages per-repo calls
- ✅ Resolved review finding [Medium]: Refactored tests to table-driven style with t.Run subtests
- ✅ Resolved review finding [Medium]: Added RepositoryArn to ScanResult.Details for DescribeRepositories
- ✅ Resolved review finding [Low]: Fixed digest-only image display format (no empty tag segment)
- ✅ Resolved review finding [Low]: Added type assertion failure handling in all 3 Process functions
- ✅ Resolved review finding [Low]: Added empty PolicyText test case for GetRepositoryPolicy

### File List

**Created:**
- `cmd/awtest/services/ecr/calls.go` — ECR service implementation (3 AWSService entries with pagination)
- `cmd/awtest/services/ecr/calls_test.go` — Table-driven Process tests (3 test functions, 12 subtests)

**Modified:**
- `cmd/awtest/services/services.go` — Added ECR import and registration in AllServices()

### Change Log

- 2026-03-10: Addressed code review findings — 7 items resolved (2 High, 2 Medium, 3 Low): added pagination, table-driven tests, RepositoryArn in Details, digest-only format fix, type assertion handling, empty policy test
- 2026-03-10: Implemented ECR Container Registry enumeration service (Story 7.1) — 3 API calls (DescribeRepositories, ListImages, GetRepositoryPolicy), 12 table-driven test cases, registered in AllServices()
