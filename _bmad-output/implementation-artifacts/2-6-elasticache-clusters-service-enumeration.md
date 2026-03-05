# Story 2.6: ElastiCache Clusters Service Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate ElastiCache clusters**,
so that **I can discover Redis/Memcached caching databases accessible with the credentials, which often contain session data, cached credentials, and PII**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/elasticache/` directory with `calls.go`
2. **AC2:** Implement `DescribeCacheClusters()` API call using AWS SDK v1.44.266 ElastiCache client (`github.com/aws/aws-sdk-go/service/elasticache`) with `ShowCacheNodeInfo=true`
3. **AC3:** Implement AWSService interface: `Name="elasticache:DescribeCacheClusters"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates ElastiCache client per region using `sess.Copy()`, calls `DescribeCacheClusters` with `ShowCacheNodeInfo: aws.Bool(true)`, aggregates all `*elasticache.CacheCluster` results
5. **AC5:** `Process()` displays each cluster: CacheClusterId, Engine (redis/memcached), EngineVersion, CacheNodeType, CacheClusterStatus, NumCacheNodes, PreferredAvailabilityZone
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results -- if no clusters found after all regions, call `utils.PrintAccessGranted(debug, "elasticache:DescribeCacheClusters", "clusters")` and return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `eks`, before `elasticbeanstalk`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: Redis cluster, Memcached cluster, multiple clusters, empty results, access denied, nil field handling
10. **AC10:** Package naming: `elasticache` (lowercase, single word)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/elasticache/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR16 requirement fulfilled: System enumerates ElastiCache clusters

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/elasticache/`
  - [x] Create `calls.go` with package `elasticache`
  - [x] Define `var ElastiCacheCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `elasticache.New(regionSess)` per region using `sess.Copy()`, call `svc.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{ShowCacheNodeInfo: aws.Bool(true)})`, aggregate all `Output.CacheClusters` ([]*elasticache.CacheCluster) across regions
  - [x] Return aggregated `[]*elasticache.CacheCluster` from Call(), or first error encountered

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "elasticache:DescribeCacheClusters", err)`, return error ScanResult
  - [x] Type-assert output to `[]*elasticache.CacheCluster`
  - [x] Handle type assertion failure (like EFS pattern)
  - [x] If empty slice and no error: call `utils.PrintAccessGranted(debug, "elasticache:DescribeCacheClusters", "clusters")`, return empty results
  - [x] For each cluster, extract: `CacheClusterId` (`*string`), `Engine` (`*string`), `EngineVersion` (`*string`), `CacheNodeType` (`*string`), `CacheClusterStatus` (`*string`), `NumCacheNodes` (`*int64`), `PreferredAvailabilityZone` (`*string`) -- all with nil checks
  - [x] Build `types.ScanResult` entries with: ServiceName="ElastiCache", MethodName="elasticache:DescribeCacheClusters", ResourceType="cluster"
  - [x] Call `utils.PrintResult()` with formatted output
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/elasticache"` to `services/services.go`
  - [x] Add `allServices = append(allServices, elasticache.ElastiCacheCalls...)` after `eks.EKSCalls...` and before `elasticbeanstalk.ElasticBeanstalkCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/elasticache/calls_test.go`
  - [x] Follow Story 2.4 test pattern: table-driven Process()-only tests with pre-built mock data
  - [x] Test cases: Redis cluster (all fields populated), Memcached cluster, multiple clusters, empty results, access denied error, nil field handling

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/elasticache/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow Established Epic 2 Pattern

Stories 2.1-2.5 established the pattern. ElastiCache is a SINGLE API call pattern (like EFS Story 2.4), NOT a two-API-call pattern (like EKS Story 2.5). The key difference is `ShowCacheNodeInfo=true` parameter on the input.

### CRITICAL: Use sess.Copy() for Region Iteration

Story 2.3 code review identified that mutating `sess.Config.Region` directly is unsafe. The fix was to use `sess.Copy()`. **YOU MUST USE `sess.Copy()`** for safe session handling:

```go
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := elasticache.New(regionSess)
    // ...
}
```

**DO NOT** use the older ECS pattern (`session.NewSession(regionConfig)`) or mutate `sess.Config.Region` directly.

### AWS ElastiCache SDK Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/elasticache`

**Key API: DescribeCacheClusters**
- `elasticache.New(sess)` -- creates ElastiCache client
- `svc.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{ShowCacheNodeInfo: aws.Bool(true)})` -- returns `*elasticache.DescribeCacheClustersOutput`
- `Output.CacheClusters` -- `[]*elasticache.CacheCluster` -- list of cache clusters
- `ShowCacheNodeInfo: aws.Bool(true)` -- includes individual cache node details (important for security context)
- May need pagination via `Marker` for accounts with many clusters (initial implementation without pagination is acceptable)

**CacheCluster fields (from `*elasticache.CacheCluster`):**
- `CacheClusterId` -- `*string` -- cluster identifier (e.g., "my-redis-cluster")
- `Engine` -- `*string` -- "redis" or "memcached"
- `EngineVersion` -- `*string` -- engine version (e.g., "7.0.7" for Redis, "1.6.22" for Memcached)
- `CacheNodeType` -- `*string` -- instance type (e.g., "cache.t3.micro")
- `CacheClusterStatus` -- `*string` -- "available", "creating", "deleting", "modifying", etc.
- `NumCacheNodes` -- `*int64` -- number of nodes in the cluster
- `PreferredAvailabilityZone` -- `*string` -- AZ where cluster is deployed (e.g., "us-east-1a")

**No new dependencies needed** -- `elasticache` is part of `aws-sdk-go v1.44.266` already in go.mod.

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `elasticache` |
| Package variable | `ElastiCacheCalls` |
| AWSService.Name | `"elasticache:DescribeCacheClusters"` |
| ScanResult.ServiceName | `"ElastiCache"` |
| ScanResult.MethodName | `"elasticache:DescribeCacheClusters"` |
| ScanResult.ResourceType | `"cluster"` |

### Registration Order in services.go

Insert after `eks` and before `elasticbeanstalk` (alphabetical):

```go
allServices = append(allServices, eks.EKSCalls...)
allServices = append(allServices, elasticache.ElastiCacheCalls...)  // NEW
allServices = append(allServices, elasticbeanstalk.ElasticBeanstalkCalls...)
```

Import alphabetically after `eks`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/eks"
"github.com/MillerMedia/awtest/cmd/awtest/services/elasticache"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/elasticbeanstalk"
```

### Process() Output Format

```go
utils.PrintResult(debug, "", "elasticache:DescribeCacheClusters",
    fmt.Sprintf("Found ElastiCache Cluster: %s (Engine: %s %s, Type: %s, Status: %s, Nodes: %d, AZ: %s)",
        utils.ColorizeItem(clusterId), engine, engineVersion, nodeType, status, numNodes, az), nil)
```

### Empty Results Handling

Following the same pattern as Story 2.4 (EFS) using `utils.PrintAccessGranted`:

```go
if len(clusters) == 0 {
    utils.PrintAccessGranted(debug, "elasticache:DescribeCacheClusters", "clusters")
    return results
}
```

### Reference Implementation Pattern

```go
package elasticache

import (
    "fmt"
    "github.com/MillerMedia/awtest/cmd/awtest/types"
    "github.com/MillerMedia/awtest/cmd/awtest/utils"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/elasticache"
    "time"
)

var ElastiCacheCalls = []types.AWSService{
    {
        Name: "elasticache:DescribeCacheClusters",
        Call: func(sess *session.Session) (interface{}, error) {
            var allClusters []*elasticache.CacheCluster
            for _, region := range types.Regions {
                regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
                svc := elasticache.New(regionSess)
                output, err := svc.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
                    ShowCacheNodeInfo: aws.Bool(true),
                })
                if err != nil {
                    return nil, err
                }
                allClusters = append(allClusters, output.CacheClusters...)
            }
            return allClusters, nil
        },
        Process: func(output interface{}, err error, debug bool) []types.ScanResult {
            var results []types.ScanResult

            if err != nil {
                utils.HandleAWSError(debug, "elasticache:DescribeCacheClusters", err)
                return []types.ScanResult{
                    {
                        ServiceName: "ElastiCache",
                        MethodName:  "elasticache:DescribeCacheClusters",
                        Error:       err,
                        Timestamp:   time.Now(),
                    },
                }
            }

            clusters, ok := output.([]*elasticache.CacheCluster)
            if !ok {
                utils.HandleAWSError(debug, "elasticache:DescribeCacheClusters", fmt.Errorf("unexpected output type %T", output))
                return results
            }

            if len(clusters) == 0 {
                utils.PrintAccessGranted(debug, "elasticache:DescribeCacheClusters", "clusters")
                return results
            }

            for _, cluster := range clusters {
                clusterId := ""
                if cluster.CacheClusterId != nil {
                    clusterId = *cluster.CacheClusterId
                }

                engine := ""
                if cluster.Engine != nil {
                    engine = *cluster.Engine
                }

                engineVersion := ""
                if cluster.EngineVersion != nil {
                    engineVersion = *cluster.EngineVersion
                }

                nodeType := ""
                if cluster.CacheNodeType != nil {
                    nodeType = *cluster.CacheNodeType
                }

                status := ""
                if cluster.CacheClusterStatus != nil {
                    status = *cluster.CacheClusterStatus
                }

                var numNodes int64
                if cluster.NumCacheNodes != nil {
                    numNodes = *cluster.NumCacheNodes
                }

                az := ""
                if cluster.PreferredAvailabilityZone != nil {
                    az = *cluster.PreferredAvailabilityZone
                }

                results = append(results, types.ScanResult{
                    ServiceName:  "ElastiCache",
                    MethodName:   "elasticache:DescribeCacheClusters",
                    ResourceType: "cluster",
                    ResourceName: clusterId,
                    Details: map[string]interface{}{
                        "Engine":                  engine,
                        "EngineVersion":           engineVersion,
                        "CacheNodeType":           nodeType,
                        "CacheClusterStatus":      status,
                        "NumCacheNodes":           numNodes,
                        "PreferredAvailabilityZone": az,
                    },
                    Timestamp: time.Now(),
                })

                utils.PrintResult(debug, "", "elasticache:DescribeCacheClusters",
                    fmt.Sprintf("Found ElastiCache Cluster: %s (Engine: %s %s, Type: %s, Status: %s, Nodes: %d, AZ: %s)",
                        utils.ColorizeItem(clusterId), engine, engineVersion, nodeType, status, numNodes, az), nil)
            }
            return results
        },
        ModuleName: types.DefaultModuleName,
    },
}
```

### Testing Pattern (from Story 2.4)

Create table-driven Process()-only tests with pre-built mock data. No AWS SDK mocking needed.

```go
func TestProcess(t *testing.T) {
    process := ElastiCacheCalls[0].Process

    tests := []struct {
        name          string
        input         interface{}
        err           error
        expectedCount int
        expectError   bool
        checkResults  func(t *testing.T, results []types.ScanResult)
    }{
        {
            name: "valid Redis cluster with all fields",
            input: []*elasticache.CacheCluster{
                {
                    CacheClusterId:          aws.String("my-redis-cluster"),
                    Engine:                  aws.String("redis"),
                    EngineVersion:           aws.String("7.0.7"),
                    CacheNodeType:           aws.String("cache.t3.micro"),
                    CacheClusterStatus:      aws.String("available"),
                    NumCacheNodes:           aws.Int64(1),
                    PreferredAvailabilityZone: aws.String("us-east-1a"),
                },
            },
            expectedCount: 1,
            checkResults: func(t *testing.T, results []types.ScanResult) {
                // Verify ServiceName, MethodName, ResourceType, ResourceName, Details
            },
        },
        {
            name: "Memcached cluster",
            input: []*elasticache.CacheCluster{
                {
                    CacheClusterId:     aws.String("my-memcached"),
                    Engine:             aws.String("memcached"),
                    EngineVersion:      aws.String("1.6.22"),
                    CacheNodeType:      aws.String("cache.m5.large"),
                    CacheClusterStatus: aws.String("available"),
                    NumCacheNodes:      aws.Int64(3),
                },
            },
            expectedCount: 1,
        },
        {
            name: "multiple clusters",
            input: []*elasticache.CacheCluster{
                {CacheClusterId: aws.String("redis-1"), Engine: aws.String("redis")},
                {CacheClusterId: aws.String("memcached-1"), Engine: aws.String("memcached")},
            },
            expectedCount: 2,
        },
        {
            name:          "empty results",
            input:         []*elasticache.CacheCluster{},
            expectedCount: 0,
        },
        {
            name:          "access denied error",
            input:         nil,
            err:           fmt.Errorf("AccessDeniedException: User is not authorized"),
            expectedCount: 1,
            expectError:   true,
        },
        {
            name: "nil field handling",
            input: []*elasticache.CacheCluster{
                {
                    CacheClusterId:          nil,
                    Engine:                  nil,
                    EngineVersion:           nil,
                    CacheNodeType:           nil,
                    CacheClusterStatus:      nil,
                    NumCacheNodes:           nil,
                    PreferredAvailabilityZone: nil,
                },
            },
            expectedCount: 1,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            results := process(tc.input, tc.err, false)
            if len(results) != tc.expectedCount {
                t.Fatalf("expected %d results, got %d", tc.expectedCount, len(results))
            }
            if tc.checkResults != nil {
                tc.checkResults(t, results)
            }
        })
    }
}
```

Test cases:
1. **Valid Redis cluster with all fields** -- all fields populated, verify ScanResult fields and Details map
2. **Memcached cluster** -- verify different engine type is handled correctly
3. **Multiple clusters** -- verify correct count and each cluster captured
4. **Empty results** -- verify `PrintAccessGranted` behavior and empty results returned
5. **Access denied** -- verify error ScanResult returned with correct ServiceName/MethodName
6. **Nil field handling** -- verify nil CacheClusterId, nil Engine, nil NumCacheNodes etc. handled gracefully (empty string/zero defaults)

### Edge Cases

1. **No clusters in any region** -- Call() returns empty slice, Process() calls PrintAccessGranted, returns empty results
2. **Access denied on DescribeCacheClusters** -- Call() returns error immediately (fail fast)
3. **Cluster with nil fields** -- defensive nil checks on all pointer fields before dereferencing
4. **Redis vs Memcached** -- both engine types handled identically, Engine field captures which type
5. **Cluster in creating/deleting state** -- still enumerate it, CacheClusterStatus field reflects the state
6. **Many clusters in one region** -- DescribeCacheClusters may need pagination via Marker (initial implementation without pagination is acceptable)

### Architecture Compliance

- **Package:** `elasticache` in `cmd/awtest/services/elasticache/` -- MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) -- MUST FOLLOW
- **Variable:** `ElastiCacheCalls` exported slice -- MUST FOLLOW
- **Type:** `[]types.AWSService` -- MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` -- MUST FOLLOW (all Epic 2 stories use DefaultModuleName)
- **Session handling:** `sess.Copy(&aws.Config{Region: aws.String(region)})` -- MUST FOLLOW (Story 2.3 code review fix)
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` -- MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` -- MUST FOLLOW
- **Nil checks:** Always check `*string`, `*int64`, `*bool` before dereferencing -- MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) -- MUST FOLLOW

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/elasticache/
+-- calls.go            # NEW: AWS ElastiCache service implementation
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
cmd/awtest/services/efs/calls.go               # Story 2.4 reference (sess.Copy + single API pattern)
cmd/awtest/services/eks/calls.go               # Story 2.5 reference (sess.Copy + two API pattern)
go.mod                                         # AWS SDK already included
```

### Previous Story Intelligence (Story 2.5)

**Key learnings from Story 2.5 (EKS):**
- **sess.Copy() is mandatory** -- continued from Story 2.3 fix. All subsequent stories MUST use this pattern.
- Table-driven Process()-only tests are the standard -- ElastiCache tests should follow this from the start.
- Type assertion failure handling included from the start -- ElastiCache should include this.
- All display fields must appear in BOTH PrintResult AND Details map.
- `ScanResult.Timestamp = time.Now()` is required on every result.
- Empty results handled with `utils.PrintAccessGranted` -- same for ElastiCache.
- Code review fix from 2.5: DescribeCluster error handling changed from fail-fast to resilient (continue). ElastiCache uses a SINGLE API call so this doesn't apply, but note the pattern if future enhancements add per-cluster describe calls.

**Key learnings from Story 2.4 (EFS):**
- EFS is the closest pattern match for ElastiCache (single API call, regional iteration, direct result aggregation).
- Copy the EFS structure almost exactly, changing only the service-specific parts.

### Git Intelligence

**Recent commits (Epic 2 context):**
- `312e412` Add EKS service enumeration (Story 2.5)
- `e6ac03e` Mark Stories 2.3 and 2.4 as done
- `a5314de` Add EFS service enumeration (Story 2.4)
- `a416792` Add AWS Config service enumeration (Story 2.3)
- `c71ad1b` Add ACM and Cognito User Pools service enumeration (Stories 2.1, 2.2)

**Patterns established:**
- Commits reference story numbers in message
- Build verification (go build, go test, go vet) done before commit
- Small, focused commits per story

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `elasticache` follows convention (lowercase, single word, matches directory)
- Single `calls.go` file per service -- matches all 38+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/elasticache`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.6: ElastiCache Clusters Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR7-31 Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#Phase 1 Additions Needed]
- [Source: _bmad-output/implementation-artifacts/2-5-eks-elastic-kubernetes-service-enumeration.md -- previous story learnings]
- [Source: cmd/awtest/services/efs/calls.go -- Story 2.4 reference (closest pattern match: sess.Copy + single API)]
- [Source: cmd/awtest/services/efs/calls_test.go -- Story 2.4 test reference (table-driven Process()-only)]
- [Source: cmd/awtest/services/services.go -- AllServices() registration point]
- [Source: cmd/awtest/types/types.go -- AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go -- PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem]
- [Source: go.mod -- aws-sdk-go v1.44.266 (includes elasticache package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None - clean implementation with no issues encountered.

### Completion Notes List

- Implemented ElastiCache DescribeCacheClusters service enumeration following the EFS (Story 2.4) single-API-call pattern
- Used `sess.Copy()` for safe regional session handling per Story 2.3 code review fix
- `ShowCacheNodeInfo: aws.Bool(true)` included in API call for security-relevant cache node details
- All 7 cluster fields extracted with nil-safe dereferencing: CacheClusterId, Engine, EngineVersion, CacheNodeType, CacheClusterStatus, NumCacheNodes, PreferredAvailabilityZone
- Type assertion failure handling included from the start (Story 2.5 learning)
- Registered in services.go alphabetically between eks and elasticbeanstalk
- 6 table-driven Process() tests: Redis cluster, Memcached cluster, multiple clusters, empty results, access denied, nil fields
- All tests pass, go build and go vet clean, no regressions in full test suite
- Code review follow-up: Call() now resilient to per-region errors (continues to next region instead of failing entirely)
- Code review follow-up: Added Marker-based pagination for large accounts with many clusters
- Code review low finding (WithContext) deferred - project-wide pattern, not specific to this story

### File List

- `cmd/awtest/services/elasticache/calls.go` (NEW) - ElastiCache service implementation
- `cmd/awtest/services/elasticache/calls_test.go` (NEW) - Process() unit tests
- `cmd/awtest/services/services.go` (MODIFIED) - Added elasticache import and registration

## Change Log

- 2026-03-05: Implemented Story 2.6 - ElastiCache Clusters Service Enumeration (all 5 tasks, all 14 ACs satisfied)
- 2026-03-05: Addressed code review findings - 2 medium items resolved (regional resilience, pagination), 1 low deferred (WithContext is project-wide)
