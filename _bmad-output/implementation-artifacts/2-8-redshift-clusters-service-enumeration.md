# Story 2.8: Redshift Clusters Service Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Redshift data warehouse clusters**,
so that **I can discover data warehouses accessible with the credentials, which often contain large volumes of analytics data and business intelligence**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/redshift/` directory with `calls.go`
2. **AC2:** Implement Redshift cluster enumeration using AWS SDK v1.44.266 Redshift client (`github.com/aws/aws-sdk-go/service/redshift`)
3. **AC3:** Implement AWSService interface: `Name="redshift:DescribeClusters"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates Redshift client per region using `sess.Copy()`, calls `DescribeClusters` with `Marker`-based pagination -- aggregates all `*redshift.Cluster` results
5. **AC5:** `Process()` displays each cluster: ClusterIdentifier, NodeType, ClusterStatus, MasterUsername, DBName, Endpoint (Address:Port), Encrypted, NumberOfNodes
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results -- if no Redshift clusters found after all regions, call `utils.PrintAccessGranted(debug, "redshift:DescribeClusters", "Redshift clusters")` and return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `rds`, before `rekognition`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: valid cluster with all fields, encrypted vs unencrypted, multiple clusters, empty results, access denied, nil field handling
10. **AC10:** Package naming: `redshift` (lowercase, single word)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/redshift/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR17 requirement fulfilled: System enumerates Redshift data warehouses

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/redshift/`
  - [x] Create `calls.go` with package `redshift`
  - [x] Define `var RedshiftCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `redshift.New(regionSess)` per region using `sess.Copy(&aws.Config{Region: aws.String(region)})`, call `svc.DescribeClusters(&redshift.DescribeClustersInput{})` with Marker-based pagination
  - [x] Use resilient per-region error handling (continue to next region on error, like ElastiCache pattern)
  - [x] Return aggregated `[]*redshift.Cluster` from Call(), or nil on complete failure

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "redshift:DescribeClusters", err)`, return error ScanResult
  - [x] Type-assert output to `[]*redshift.Cluster`
  - [x] Handle type assertion failure (like ElastiCache pattern)
  - [x] If empty slice and no error: call `utils.PrintAccessGranted(debug, "redshift:DescribeClusters", "Redshift clusters")`, return empty results
  - [x] For each cluster, extract: `ClusterIdentifier` (`*string`), `NodeType` (`*string`), `ClusterStatus` (`*string`), `MasterUsername` (`*string`), `DBName` (`*string`), `Endpoint.Address` (`*string`) + `Endpoint.Port` (`*int64`), `Encrypted` (`*bool`), `NumberOfNodes` (`*int64`) -- all with nil checks
  - [x] Build `types.ScanResult` entries with: ServiceName="Redshift", MethodName="redshift:DescribeClusters", ResourceType="cluster"
  - [x] Call `utils.PrintResult()` with formatted output
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/redshift"` to `services/services.go`
  - [x] Add `allServices = append(allServices, redshift.RedshiftCalls...)` after `rds.RDSCalls...` and before `rekognition.RekognitionCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/redshift/calls_test.go`
  - [x] Follow ElastiCache test pattern: table-driven Process()-only tests with pre-built mock data
  - [x] Test cases: valid cluster (all fields), encrypted cluster, multiple clusters, empty results, access denied error, nil field handling

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/redshift/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### Redshift Uses Single DescribeClusters API -- Simple Pattern Like ElastiCache

Redshift `DescribeClusters` is a single API call that returns all clusters. This follows the **same simple pattern as ElastiCache** (Story 2.6), NOT the multi-API-call pattern used by Fargate (Story 2.7) or EKS (Story 2.5).

**Package:** `github.com/aws/aws-sdk-go/service/redshift` -- already available in `aws-sdk-go v1.44.266` in go.mod. No new dependencies needed.

### CRITICAL: Use sess.Copy() for Region Iteration

Story 2.3 code review identified that mutating `sess.Config.Region` directly is unsafe. **YOU MUST USE `sess.Copy()`** for safe session handling:

```go
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := redshift.New(regionSess)
    // ...
}
```

### AWS Redshift SDK Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/redshift`

**API Call:**
- `svc.DescribeClusters(&redshift.DescribeClustersInput{})` -> `*redshift.DescribeClustersOutput`
- `Output.Clusters` -> `[]*redshift.Cluster`
- Pagination via `Marker` (same pattern as ElastiCache)

**Cluster fields (from `*redshift.Cluster`):**
- `ClusterIdentifier` -- `*string` -- unique cluster name (e.g., "my-redshift-cluster")
- `NodeType` -- `*string` -- instance type (e.g., "dc2.large", "ra3.xlplus")
- `ClusterStatus` -- `*string` -- "available", "creating", "deleting", etc.
- `MasterUsername` -- `*string` -- admin database user
- `DBName` -- `*string` -- default database name
- `Endpoint` -- `*redshift.Endpoint` -- contains `Address` (`*string`) and `Port` (`*int64`)
- `Encrypted` -- `*bool` -- whether cluster is encrypted
- `NumberOfNodes` -- `*int64` -- node count

**IMPORTANT: Endpoint is a nested struct, NOT a simple field.** Must nil-check `Endpoint` before accessing `Endpoint.Address` and `Endpoint.Port`:

```go
endpoint := ""
if cluster.Endpoint != nil {
    addr := ""
    if cluster.Endpoint.Address != nil {
        addr = *cluster.Endpoint.Address
    }
    port := int64(0)
    if cluster.Endpoint.Port != nil {
        port = *cluster.Endpoint.Port
    }
    endpoint = fmt.Sprintf("%s:%d", addr, port)
}
```

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `redshift` |
| Package variable | `RedshiftCalls` |
| AWSService.Name | `"redshift:DescribeClusters"` |
| ScanResult.ServiceName | `"Redshift"` |
| ScanResult.MethodName | `"redshift:DescribeClusters"` |
| ScanResult.ResourceType | `"cluster"` |

### Registration Order in services.go

Insert after `rds` and before `rekognition` (alphabetical):

```go
allServices = append(allServices, rds.RDSCalls...)
allServices = append(allServices, redshift.RedshiftCalls...)  // NEW
allServices = append(allServices, rekognition.RekognitionCalls...)
```

Import alphabetically after `rds`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/rds"
"github.com/MillerMedia/awtest/cmd/awtest/services/redshift"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/rekognition"
```

### Process() Output Format

```go
utils.PrintResult(debug, "", "redshift:DescribeClusters",
    fmt.Sprintf("Found Redshift Cluster: %s (Type: %s, Status: %s, User: %s, DB: %s, Endpoint: %s, Encrypted: %v, Nodes: %d)",
        utils.ColorizeItem(clusterId), nodeType, status, masterUser, dbName, endpoint, encrypted, numNodes), nil)
```

### Empty Results Handling

```go
if len(clusters) == 0 {
    utils.PrintAccessGranted(debug, "redshift:DescribeClusters", "Redshift clusters")
    return results
}
```

### Reference Implementation Pattern

Follow ElastiCache (`cmd/awtest/services/elasticache/calls.go`) as the primary reference -- same single-API pattern with Marker pagination and resilient per-region error handling.

Key differences from ElastiCache:
1. Uses `redshift` package instead of `elasticache`
2. Different fields to extract (ClusterIdentifier, NodeType, etc.)
3. Has nested `Endpoint` struct requiring extra nil check
4. Has `*bool` field (`Encrypted`) -- requires `*bool` nil check pattern
5. No input parameters needed (ElastiCache passes `ShowCacheNodeInfo: true`)

### Testing Pattern (from ElastiCache Story 2.6)

Create table-driven Process()-only tests with pre-built mock data. No AWS SDK mocking needed.

Test cases:
1. **Valid cluster with all fields** -- all fields populated including Endpoint, verify all ScanResult fields and Details map
2. **Encrypted cluster** -- verify Encrypted=true is captured correctly
3. **Multiple clusters** -- verify correct count
4. **Empty results** -- verify PrintAccessGranted behavior and empty results returned
5. **Access denied** -- verify error ScanResult returned with correct ServiceName/MethodName
6. **Nil field handling** -- verify nil ClusterIdentifier, nil Endpoint, nil Encrypted etc. handled gracefully

**IMPORTANT for tests:** When creating mock `redshift.Cluster` with Endpoint, build the nested struct:
```go
&redshift.Cluster{
    ClusterIdentifier: aws.String("my-cluster"),
    Endpoint: &redshift.Endpoint{
        Address: aws.String("my-cluster.abc123.us-east-1.redshift.amazonaws.com"),
        Port:    aws.Int64(5439),
    },
    Encrypted:     aws.Bool(true),
    NumberOfNodes: aws.Int64(2),
    // ...
}
```

### Edge Cases

1. **No clusters in any region** -- DescribeClusters returns empty, Process() calls PrintAccessGranted
2. **Access denied in all regions** -- Call() returns nil + error, Process() handles error
3. **Access denied in some regions** -- Call() continues to next region (resilient pattern), returns partial results
4. **Cluster with nil Endpoint** -- can happen for clusters in "creating" status, must handle gracefully
5. **Cluster with nil Encrypted** -- defensive nil check, default to false
6. **Pagination across many clusters** -- handle Marker token for accounts with many clusters

### Architecture Compliance

- **Package:** `redshift` in `cmd/awtest/services/redshift/` -- MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) -- MUST FOLLOW
- **Variable:** `RedshiftCalls` exported slice -- MUST FOLLOW
- **Type:** `[]types.AWSService` -- MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` -- MUST FOLLOW (all Epic 2 stories use DefaultModuleName)
- **Session handling:** `sess.Copy(&aws.Config{Region: aws.String(region)})` -- MUST FOLLOW (Story 2.3 code review fix)
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` -- MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` -- MUST FOLLOW
- **Nil checks:** Always check `*string`, `*int64`, `*bool` before dereferencing -- MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) -- MUST FOLLOW
- **SDK version:** AWS SDK Go v1.44.266 -- MUST FOLLOW (do NOT use SDK v2)

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/redshift/
+-- calls.go            # NEW: Redshift clusters service implementation
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
cmd/awtest/services/elasticache/calls.go       # Story 2.6 reference (sess.Copy + resilient + Marker pagination) -- PRIMARY REFERENCE
cmd/awtest/services/elasticache/calls_test.go  # Story 2.6 test reference (table-driven Process()-only)
go.mod                                         # AWS SDK already included (redshift package available)
```

### Previous Story Intelligence (Story 2.7)

**Key learnings from Story 2.7 (Fargate):**
- **sess.Copy() is mandatory** -- continued from Story 2.3 fix
- **Resilient per-region errors** with `anyRegionSucceeded` + `lastErr` tracking pattern
- **Pagination included from the start** -- avoid code review rework
- Table-driven Process()-only tests are the standard
- Type assertion failure handling included from the start
- All display fields must appear in BOTH PrintResult AND Details map
- `ScanResult.Timestamp = time.Now()` is required on every result
- Empty results handled with `utils.PrintAccessGranted`

### Git Intelligence

**Recent commits (Epic 2 context):**
- `f814fab` Fix false "Access granted" when all regions return access denied
- `f46641a` Add Fargate tasks service enumeration (Story 2.7)
- `12e6771` Add ElastiCache service enumeration (Story 2.6)
- `312e412` Add EKS service enumeration (Story 2.5)
- `e6ac03e` Mark Stories 2.3 and 2.4 as done

**Key insight from f814fab:** The `anyRegionSucceeded` + `lastErr` pattern in Call() is critical. If all regions return access denied, Call() must return the error (not nil) so Process() can properly report it. This was a bug fix applied after Story 2.7.

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `redshift` follows convention (lowercase, single word, matches directory)
- Single `calls.go` file per service -- matches all 38+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/redshift`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.8: Redshift Clusters Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR7-31 Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#redshift package example]
- [Source: _bmad-output/implementation-artifacts/2-7-fargate-tasks-service-enumeration.md -- previous story learnings]
- [Source: cmd/awtest/services/elasticache/calls.go -- PRIMARY reference (sess.Copy + resilient + Marker pagination)]
- [Source: cmd/awtest/services/elasticache/calls_test.go -- test reference (table-driven Process()-only)]
- [Source: cmd/awtest/services/services.go -- AllServices() registration point]
- [Source: cmd/awtest/types/types.go -- AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go -- PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem]
- [Source: go.mod -- aws-sdk-go v1.44.266 (includes redshift package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Implemented Redshift cluster enumeration following ElastiCache pattern (single-API, Marker pagination, resilient per-region errors)
- Used sess.Copy() for safe region iteration per Story 2.3 learnings
- Implemented anyRegionSucceeded + lastErr pattern per f814fab bug fix
- All 8 cluster fields extracted with proper nil checks including nested Endpoint struct and *bool Encrypted field
- All fields present in both PrintResult output and ScanResult.Details map
- Registered in services.go alphabetically after rds, before rekognition
- 8 table-driven Process()-only tests: valid cluster, encrypted cluster, multiple clusters, empty results, access denied, nil fields, endpoint with nil address, endpoint with nil port
- All tests pass, go build succeeds, go vet clean, no regressions
- Resolved review finding [High]: Fixed endpoint formatting to only include colon+port when port exists; address-only when port is nil
- Review findings [Medium] MaxRecords and silent region failures not addressed - these are cross-cutting patterns consistent across all services (ElastiCache, Fargate, EKS, EFS, etc.)

### Change Log

- 2026-03-05: Implemented Redshift clusters service enumeration (Story 2.8)
- 2026-03-05: Addressed code review findings - 1 HIGH item resolved (endpoint formatting), 2 MEDIUM deferred as cross-cutting concerns

### File List

- cmd/awtest/services/redshift/calls.go (NEW)
- cmd/awtest/services/redshift/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED)
