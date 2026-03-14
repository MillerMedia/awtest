# Story 8.3: OpenSearch Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate OpenSearch domains, access policies, and encryption configurations,
So that I can discover search clusters with sensitive data and identify access control weaknesses.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/opensearch/` directory with `calls.go` implementing OpenSearch service enumeration.

2. **AC2:** Implement `opensearch:ListDomains` API call — iterates all regions in `types.Regions`, creates OpenSearch client per region using config override pattern (`opensearchservice.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListDomainNamesWithContext` to get all domain names, then calls `DescribeDomainsWithContext` in batches of 5 to retrieve full domain metadata. Each domain listed with DomainName, ARN, Endpoint, EngineVersion, and Region.

3. **AC3:** Implement `opensearch:DescribeDomainAccessPolicies` API call — iterates all regions, creates OpenSearch client per region using config override, calls `ListDomainNamesWithContext` to get domain names, then calls `DescribeDomainsWithContext` in batches of 5. Each result extracts and lists the AccessPolicies JSON string with DomainName and Region.

4. **AC4:** Implement `opensearch:DescribeDomainEncryption` API call — iterates all regions, creates OpenSearch client per region using config override, calls `ListDomainNamesWithContext` to get domain names, then calls `DescribeDomainsWithContext` in batches of 5. Each result extracts EncryptionAtRestOptions (Enabled, KmsKeyId) and NodeToNodeEncryptionOptions (Enabled) with DomainName and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without OpenSearch access, OpenSearch is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register OpenSearch service in `services/services.go` `AllServices()` function in alphabetical order (after `lambda`, before `organizations`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid domains/policies/encryption, empty results, access denied errors, nil field handling, type assertion failure handling.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/opensearch/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `opensearch:ListDomains` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/opensearch/`
  - [x] Create `calls.go` with `package opensearch`
  - [x] Define `var OpenSearchCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"opensearch:ListDomains"`
  - [x] Call: iterate `types.Regions`, create `opensearchservice.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `ListDomainNamesWithContext` (NOT paginated — returns all domains at once), collect domain names. Then batch-describe domains via `DescribeDomainsWithContext` (max 5 domain names per call). Define local struct `osDomain` with fields: Name, ARN, Endpoint, EngineVersion, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]osDomain`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"OpenSearch"`, ResourceType=`"domain"`, ResourceName=domainName
  - [x] `utils.PrintResult` format: `"OpenSearch Domain: %s (Endpoint: %s, Engine: %s, Region: %s)"` with `utils.ColorizeItem(domainName)`

- [x] Task 2: Implement `opensearch:DescribeDomainAccessPolicies` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"opensearch:DescribeDomainAccessPolicies"`
  - [x] Call: iterate regions -> create OpenSearch client with config override -> `ListDomainNamesWithContext` -> `DescribeDomainsWithContext` (batches of 5) -> extract AccessPolicies. Define local struct `osDomainPolicy` with fields: DomainName, AccessPolicy, Region. Per-region errors: `break` to next region, don't abort scan. Skip domains with empty/nil AccessPolicies.
  - [x] Process: type-assert `[]osDomainPolicy`, build `ScanResult` with ServiceName=`"OpenSearch"`, ResourceType=`"access-policy"`, ResourceName=domainName
  - [x] `utils.PrintResult` format: `"OpenSearch Access Policy: %s (Region: %s)"` with `utils.ColorizeItem(domainName)`

- [x] Task 3: Implement `opensearch:DescribeDomainEncryption` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"opensearch:DescribeDomainEncryption"`
  - [x] Call: iterate regions -> create OpenSearch client with config override -> `ListDomainNamesWithContext` -> `DescribeDomainsWithContext` (batches of 5) -> extract EncryptionAtRestOptions and NodeToNodeEncryptionOptions. Define local struct `osDomainEncryption` with fields: DomainName, EncryptionAtRestEnabled, KmsKeyId, NodeToNodeEncryptionEnabled, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Process: type-assert `[]osDomainEncryption`, build `ScanResult` with ServiceName=`"OpenSearch"`, ResourceType=`"encryption-config"`, ResourceName=domainName
  - [x] `utils.PrintResult` format: `"OpenSearch Encryption: %s (AtRest: %t, NodeToNode: %t, Region: %s)"` with `utils.ColorizeItem(domainName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/opensearch"` to `services/services.go` (alphabetical in imports: after `lambda`, before `organizations`)
  - [x] Add `allServices = append(allServices, opensearch.OpenSearchCalls...)` after `lambda.LambdaCalls...` and before `organizations.OrganizationsCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/opensearch/calls_test.go`
  - [x] Test `ListDomains` Process: valid domains with details (name, ARN, endpoint, engine version), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `DescribeDomainAccessPolicies` Process: valid policies with JSON content, empty results, error handling, nil fields, type assertion failure
  - [x] Test `DescribeDomainEncryption` Process: valid encryption configs (both enabled/disabled states), empty results, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following CodeBuild/CodeCommit test pattern
  - [x] Access Process via `OpenSearchCalls[0].Process`, `OpenSearchCalls[1].Process`, `OpenSearchCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/opensearch/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := opensearchservice.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := opensearchservice.New(sess)
}
```

### OpenSearch is a REGIONAL Service

OpenSearch is **regional** — domains exist per-region. Iterate `types.Regions` for all three API calls, following the same pattern as CodeBuild, CodeCommit, GuardDuty, and Security Hub.

If OpenSearch returns `AccessDeniedException` in a region, handle as a non-fatal error: log with `utils.HandleAWSError` and `break` to next region.

### OpenSearch SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/opensearchservice`

**IMPORTANT:** The Go package name is `opensearchservice` (not `opensearch`). The import path is `github.com/aws/aws-sdk-go/service/opensearchservice` and the client constructor is `opensearchservice.New(sess, cfg)`.

**API Methods:**

1. **ListDomainNames:**
   - `svc.ListDomainNamesWithContext(ctx, &opensearchservice.ListDomainNamesInput{})` -> `*opensearchservice.ListDomainNamesOutput`
   - `.DomainNames` -> `[]*opensearchservice.DomainInfo`
   - Each `DomainInfo` has:
     - `DomainName *string`
     - `EngineType *string` ("OpenSearch" or "Elasticsearch")
   - **NOT paginated** — returns all domains at once (AWS limits ~100 per account per region)
   - Optional: `EngineType *string` filter — do NOT use filter, enumerate all engine types

2. **DescribeDomains (batch):**
   - `svc.DescribeDomainsWithContext(ctx, &opensearchservice.DescribeDomainsInput{DomainNames: names})` -> `*opensearchservice.DescribeDomainsOutput`
   - `.DomainStatusList` -> `[]*opensearchservice.DomainStatus`
   - **Max 5 domain names per call** (NOT 25 or 100)
   - Each `DomainStatus` has:
     - `DomainName *string`
     - `DomainId *string`
     - `ARN *string`
     - `Endpoint *string` (nil for VPC-only domains)
     - `Endpoints map[string]*string` (VPC endpoints — key "vpc")
     - `EngineVersion *string` (e.g., "OpenSearch_2.11" or "Elasticsearch_7.10")
     - `AccessPolicies *string` (IAM policy document as JSON string)
     - `EncryptionAtRestOptions *EncryptionAtRestOptions`
       - `Enabled *bool`
       - `KmsKeyId *string`
     - `NodeToNodeEncryptionOptions *NodeToNodeEncryptionOptions`
       - `Enabled *bool`
     - `DomainEndpointOptions *DomainEndpointOptions`
       - `EnforceHTTPS *bool`
       - `TLSSecurityPolicy *string`
     - `Created *bool`
     - `Deleted *bool`
     - `Processing *bool`

**No new dependencies needed** — OpenSearch is part of `aws-sdk-go v1.44.266` already in go.mod.

### DescribeDomains Batching Pattern

`DescribeDomains` accepts max 5 domain names per call. Batch accordingly:

```go
// Batch domain names into groups of 5
for i := 0; i < len(domainNames); i += 5 {
    end := i + 5
    if end > len(domainNames) {
        end = len(domainNames)
    }
    batch := domainNames[i:end]

    descInput := &opensearchservice.DescribeDomainsInput{
        DomainNames: batch,
    }
    descOutput, err := svc.DescribeDomainsWithContext(ctx, descInput)
    if err != nil {
        utils.HandleAWSError(false, "opensearch:ListDomains", err)
        break
    }
    // Process descOutput.DomainStatusList...
}
```

### Endpoint Handling — Public vs VPC

OpenSearch domains can have:
- **Public endpoint:** `DomainStatus.Endpoint` (*string) — e.g., "search-my-domain-abc123.us-east-1.es.amazonaws.com"
- **VPC endpoint:** `DomainStatus.Endpoints` (map[string]*string) — key "vpc", value is the VPC endpoint URL
- A domain has one or the other, not both

Handle both in the Call function:
```go
endpoint := ""
if ds.Endpoint != nil {
    endpoint = *ds.Endpoint
} else if ds.Endpoints != nil {
    if vpcEndpoint, ok := ds.Endpoints["vpc"]; ok && vpcEndpoint != nil {
        endpoint = *vpcEndpoint
    }
}
```

### Local Struct Definitions

```go
type osDomain struct {
    Name          string
    ARN           string
    Endpoint      string
    EngineVersion string
    Region        string
}

type osDomainPolicy struct {
    DomainName   string
    AccessPolicy string
    Region       string
}

type osDomainEncryption struct {
    DomainName                 string
    EncryptionAtRestEnabled    bool
    KmsKeyId                   string
    NodeToNodeEncryptionEnabled bool
    Region                     string
}
```

### Variable & Naming Conventions

- **Package:** `opensearch` (directory: `cmd/awtest/services/opensearch/`)
- **Exported variable:** `OpenSearchCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"opensearch:ListDomains"`, `"opensearch:DescribeDomainAccessPolicies"`, `"opensearch:DescribeDomainEncryption"`
- **ScanResult.ServiceName:** `"OpenSearch"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"domain"`, `"access-policy"`, `"encryption-config"` (lowercase hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/opensearchservice"` (note: package name is `opensearchservice`, NOT `opensearch`)

### Registration Order in services.go

Insert alphabetically — `opensearch` comes after `lambda`, before `organizations`:

```go
// In imports:
"github.com/MillerMedia/awtest/cmd/awtest/services/opensearch"

// In AllServices():
allServices = append(allServices, lambda.LambdaCalls...)
allServices = append(allServices, opensearch.OpenSearchCalls...)  // NEW
allServices = append(allServices, organizations.OrganizationsCalls...)
```

### Testing Pattern

Follow the CodeBuild/CodeCommit test pattern — test Process() functions only with pre-built mock data:

```go
func TestListDomainsProcess(t *testing.T) {
    process := OpenSearchCalls[0].Process
    // Table-driven tests with valid domains (name, ARN, endpoint, engine version), empty, errors, nil fields, type assertion failure
}

func TestDescribeDomainAccessPoliciesProcess(t *testing.T) {
    process := OpenSearchCalls[1].Process
    // Test with valid policies (JSON content), domains with no policy, empty results, errors, nil fields, type assertion failure
}

func TestDescribeDomainEncryptionProcess(t *testing.T) {
    process := OpenSearchCalls[2].Process
    // Test with encryption enabled/disabled combinations, empty results, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `opensearch/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `opensearchservice.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** call `DescribeDomain` (singular) per domain — use `DescribeDomains` (plural) for efficiency (5 per batch)
- **DO NOT** filter by EngineType in ListDomainNames — enumerate ALL domains (both OpenSearch and Elasticsearch engine types)
- **DO NOT** confuse the Go package name `opensearchservice` with the awtest package name `opensearch` — the SDK import is `opensearchservice`, the local package is `opensearch`

### Previous Story Intelligence

**From Story 8.2 (CodeCommit — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- Pagination: NextToken loop for paginated APIs (but ListDomainNames is NOT paginated)
- Batch API pattern: collect names first, then batch-describe (CodeCommit uses 25 per batch, OpenSearch uses 5 per batch)
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `OpenSearchCalls[N].Process` in tests
- Error result pattern: `return []types.ScanResult{{ServiceName: "OpenSearch", MethodName: "opensearch:ListDomains", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests

**From Story 8.1 (CodeBuild):**
- Batch-describe pattern: collect names via list API, then batch-get details
- 3 API calls in one service file — same Call/Process pattern per entry
- Environment variable extraction as separate call — shows how to split same data into different perspectives

**From Story 7.2 Code Review Findings:**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [LOW] Tests should cover nil fields comprehensively

**From Story 7.1 Code Review Findings:**
- [HIGH] Always add pagination from the start (NextToken loops on paginated APIs — note: ListDomainNames is NOT paginated)
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [MEDIUM] Include ARN in Details map where available
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `0dd5f6a` — Add CodeCommit enumeration with 2 API calls (Story 8.2)
- `d6dd093` — Add CodeBuild enumeration with 3 API calls (Story 8.1)
- `71cdff0` — Mark Epic 7 (Critical Security Service Expansion) as done
- `156742d` — Add Security Hub enumeration with 3 API calls (Story 7.4)
- `0de1823` — Add GuardDuty enumeration with 3 API calls (Story 7.3)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add OpenSearch enumeration with 3 API calls (Story 8.3)"`

### FRs Covered

- **FR91:** System enumerates OpenSearch domains, access policies, and encryption configurations

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/opensearch/
+-- calls.go            # OpenSearch service implementation (3 AWSService entries)
+-- calls_test.go       # Process() tests for all 3 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                    # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                   # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/codebuild/calls.go       # Reference implementation (regional multi-API + batch pattern)
cmd/awtest/services/codebuild/calls_test.go  # Reference test pattern (table-driven Process-only tests)
cmd/awtest/services/codecommit/calls.go      # Reference implementation (regional + batch pattern, most recent)
cmd/awtest/services/codecommit/calls_test.go # Reference test pattern (most recent)
go.mod                                       # AWS SDK already includes opensearchservice package
```

### References

- [Source: epics-phase2.md#Story 3.3: OpenSearch Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR91] — OpenSearch enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/codebuild/calls.go] — Reference implementation (regional + batch API pattern, 3 calls)
- [Source: cmd/awtest/services/codebuild/calls_test.go] — Reference test pattern (table-driven Process-only tests)
- [Source: cmd/awtest/services/codecommit/calls.go] — Most recent reference implementation (2 calls, batch pattern)
- [Source: cmd/awtest/services/codecommit/calls_test.go] — Most recent reference test pattern
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (opensearch goes after lambda, before organizations)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes opensearchservice package)
- [Source: 8-2-codecommit-enumeration.md] — Most recent story (patterns, batch API learnings)
- [Source: 8-1-codebuild-enumeration.md] — Previous story (3 API calls pattern, batch describe)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

No errors encountered during implementation.

### Completion Notes List

- Implemented 3 OpenSearch API calls: ListDomains, DescribeDomainAccessPolicies, DescribeDomainEncryption
- All Call functions use config override pattern (`opensearchservice.New(sess, &aws.Config{Region: ...})`) — no session mutation
- ListDomainNames is not paginated — single call per region returns all domains
- DescribeDomains batched at max 5 domain names per call
- Endpoint handling supports both public (Endpoint) and VPC (Endpoints["vpc"]) domains
- AccessPolicies call skips domains with nil/empty policies
- Encryption call extracts EncryptionAtRestOptions and NodeToNodeEncryptionOptions with nil-safe checks
- All Process functions handle errors, type assertion failures, and nil fields
- No sync primitives imported — concurrency-unaware per NFR57
- Registered in services.go alphabetically after lambda, before organizations
- 15 test cases across 3 test functions covering all scenarios
- `go build`, `go test`, `go vet`, and `go test -race` all pass clean with zero regressions

### Change Log

- 2026-03-11: Implemented OpenSearch enumeration with 3 API calls (ListDomains, DescribeDomainAccessPolicies, DescribeDomainEncryption)

### File List

**Created:**
- `cmd/awtest/services/opensearch/calls.go` — OpenSearch service implementation (3 AWSService entries)
- `cmd/awtest/services/opensearch/calls_test.go` — Process() tests for all 3 entries (15 test cases)

**Modified:**
- `cmd/awtest/services/services.go` — Added opensearch import and registration in AllServices()
