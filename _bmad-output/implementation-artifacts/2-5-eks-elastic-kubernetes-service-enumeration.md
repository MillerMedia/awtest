# Story 2.5: EKS (Elastic Kubernetes Service) Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate EKS clusters**,
so that **I can discover Kubernetes clusters accessible with the credentials, revealing container orchestration infrastructure and potential lateral movement targets**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/eks/` directory with `calls.go`
2. **AC2:** Implement `ListClusters()` and `DescribeCluster()` API calls using AWS SDK v1.44.266 EKS client (`github.com/aws/aws-sdk-go/service/eks`)
3. **AC3:** Implement AWSService interface: `Name="eks:ListClusters"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates EKS client per region using `sess.Copy()`, calls `ListClusters` to get cluster names, then `DescribeCluster` for each cluster name, aggregates results
5. **AC5:** `Process()` displays each cluster: Name, Arn, Status, Version, Endpoint, RoleArn, ResourcesVpcConfig (SubnetIds, SecurityGroupIds, VpcId)
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results -- if no clusters found after all regions, call `utils.PrintAccessGranted(debug, "eks:ListClusters", "clusters")` and return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `efs`, before `elasticbeanstalk`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: valid clusters with all fields, multiple clusters, empty results, access denied, nil field handling
10. **AC10:** Package naming: `eks` (lowercase, single word)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/eks/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR15 requirement partially fulfilled: System enumerates EKS container services (ECS already exists, Fargate in Story 2.7)

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/eks/`
  - [x] Create `calls.go` with package `eks`
  - [x] Define `var EKSCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `eks.New(regionSess)` per region using `sess.Copy()`, call `svc.ListClusters(&eks.ListClustersInput{})` to get `[]*string` cluster names, then for each cluster name call `svc.DescribeCluster(&eks.DescribeClusterInput{Name: clusterName})` to get `*eks.Cluster`, aggregate all `*eks.Cluster` across regions
  - [x] Return aggregated `[]*eks.Cluster` from Call(), or first error encountered

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "eks:ListClusters", err)`, return error ScanResult
  - [x] Type-assert output to `[]*eks.Cluster`
  - [x] Handle type assertion failure (like EFS pattern)
  - [x] If empty slice and no error: call `utils.PrintAccessGranted(debug, "eks:ListClusters", "clusters")`, return empty results
  - [x] For each cluster, extract: `Name` (`*string`), `Arn` (`*string`), `Status` (`*string`), `Version` (`*string`), `Endpoint` (`*string`), `RoleArn` (`*string`), `ResourcesVpcConfig.VpcId` (`*string`), `ResourcesVpcConfig.SubnetIds` (`[]*string`), `ResourcesVpcConfig.SecurityGroupIds` (`[]*string`) -- all with nil checks
  - [x] Build `types.ScanResult` entries with: ServiceName="EKS", MethodName="eks:ListClusters", ResourceType="cluster"
  - [x] Call `utils.PrintResult()` with formatted output
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/eks"` to `services/services.go`
  - [x] Add `allServices = append(allServices, eks.EKSCalls...)` after `efs.EfsCalls...` and before `elasticbeanstalk.ElasticBeanstalkCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/eks/calls_test.go`
  - [x] Follow Story 2.4 test pattern: table-driven Process()-only tests with pre-built mock data
  - [x] Test cases: valid cluster (all fields populated), multiple clusters, empty results, access denied error, nil field handling (nil ResourcesVpcConfig, nil Name, etc.)

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/eks/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Follow Established Epic 2 Pattern

Stories 2.1 (ACM), 2.2 (Cognito), 2.3 (Config), and 2.4 (EFS) established the pattern. This story is MORE complex than EFS -- it uses TWO API calls (ListClusters + DescribeCluster), matching the ECS pattern but with the corrected `sess.Copy()` approach.

### CRITICAL: Use sess.Copy() for Region Iteration

Story 2.3 code review identified that mutating `sess.Config.Region` directly is unsafe. The fix was to use `sess.Copy()`. **YOU MUST USE `sess.Copy()`** for safe session handling:

```go
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := eks.New(regionSess)
    // ...
}
```

**DO NOT** use the older ECS pattern (`session.NewSession(regionConfig)`) or the Story 2.2 pattern that mutates `sess.Config.Region` directly.

### CRITICAL: Two-API-Call Pattern (ListClusters + DescribeCluster)

EKS requires two API calls unlike EFS (single call). The pattern is similar to ECS:
1. `ListClusters()` returns cluster **names** (strings), not full cluster objects
2. `DescribeCluster(name)` returns the full `*eks.Cluster` for each name

```go
Call: func(sess *session.Session) (interface{}, error) {
    var allClusters []*eks.Cluster
    for _, region := range types.Regions {
        regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
        svc := eks.New(regionSess)
        listOutput, err := svc.ListClusters(&eks.ListClustersInput{})
        if err != nil {
            return nil, err
        }
        for _, clusterName := range listOutput.Clusters {
            descOutput, err := svc.DescribeCluster(&eks.DescribeClusterInput{
                Name: clusterName,
            })
            if err != nil {
                return nil, err
            }
            allClusters = append(allClusters, descOutput.Cluster)
        }
    }
    return allClusters, nil
},
```

### AWS EKS SDK Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/eks`

**Key API 1: ListClusters**
- `eks.New(sess)` -- creates EKS client
- `svc.ListClusters(&eks.ListClustersInput{})` -- returns `*eks.ListClustersOutput`
- `Output.Clusters` -- `[]*string` -- list of cluster names
- May need pagination via `NextToken` for accounts with many clusters (initial implementation without pagination is acceptable)

**Key API 2: DescribeCluster**
- `svc.DescribeCluster(&eks.DescribeClusterInput{Name: clusterName})` -- returns `*eks.DescribeClusterOutput`
- `Output.Cluster` -- `*eks.Cluster` -- full cluster details

**Cluster fields (from `*eks.Cluster`):**
- `Name` -- `*string` -- cluster name (e.g., "my-cluster")
- `Arn` -- `*string` -- full ARN
- `Status` -- `*string` -- "CREATING", "ACTIVE", "DELETING", "FAILED", "UPDATING"
- `Version` -- `*string` -- Kubernetes version (e.g., "1.28")
- `Endpoint` -- `*string` -- API server endpoint URL
- `RoleArn` -- `*string` -- IAM role ARN used by cluster
- `ResourcesVpcConfig` -- `*VpcConfigResponse` -- nested struct (CAN BE NIL)
  - `VpcId` -- `*string`
  - `SubnetIds` -- `[]*string`
  - `SecurityGroupIds` -- `[]*string`
  - `ClusterSecurityGroupId` -- `*string`
  - `EndpointPublicAccess` -- `*bool`
  - `EndpointPrivateAccess` -- `*bool`
- `CreatedAt` -- `*time.Time` (not in AC5 but available)
- `PlatformVersion` -- `*string` (not in AC5 but available)

**No new dependencies needed** -- `eks` is part of `aws-sdk-go v1.44.266` already in go.mod.

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `eks` |
| Package variable | `EKSCalls` |
| AWSService.Name | `"eks:ListClusters"` |
| ScanResult.ServiceName | `"EKS"` |
| ScanResult.MethodName | `"eks:ListClusters"` |
| ScanResult.ResourceType | `"cluster"` |

### Registration Order in services.go

Insert after `efs` and before `elasticbeanstalk` (alphabetical):

```go
allServices = append(allServices, efs.EfsCalls...)
allServices = append(allServices, eks.EKSCalls...)  // NEW
allServices = append(allServices, elasticbeanstalk.ElasticBeanstalkCalls...)
```

Import alphabetically after `efs`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/efs"
"github.com/MillerMedia/awtest/cmd/awtest/services/eks"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/elasticbeanstalk"
```

### Process() Output Format

```go
// Format VPC config for readability
vpcInfo := fmt.Sprintf("VPC: %s, Subnets: %d, SGs: %d", vpcId, len(subnetIds), len(sgIds))

utils.PrintResult(debug, "", "eks:ListClusters",
    fmt.Sprintf("Found EKS Cluster: %s (Status: %s, Version: %s, Endpoint: %s, Role: %s, %s)",
        utils.ColorizeItem(name), status, version, endpoint, roleArn, vpcInfo), nil)
```

### Empty Results Handling

Following the same pattern as Story 2.4 (EFS) using `utils.PrintAccessGranted`:

```go
if len(clusters) == 0 {
    utils.PrintAccessGranted(debug, "eks:ListClusters", "clusters")
    return results
}
```

### Reference Implementation Pattern

```go
package eks

import (
    "fmt"
    "github.com/MillerMedia/awtest/cmd/awtest/types"
    "github.com/MillerMedia/awtest/cmd/awtest/utils"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/eks"
    "time"
)

var EKSCalls = []types.AWSService{
    {
        Name: "eks:ListClusters",
        Call: func(sess *session.Session) (interface{}, error) {
            var allClusters []*eks.Cluster
            for _, region := range types.Regions {
                regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
                svc := eks.New(regionSess)
                listOutput, err := svc.ListClusters(&eks.ListClustersInput{})
                if err != nil {
                    return nil, err
                }
                for _, clusterName := range listOutput.Clusters {
                    descOutput, err := svc.DescribeCluster(&eks.DescribeClusterInput{
                        Name: clusterName,
                    })
                    if err != nil {
                        return nil, err
                    }
                    allClusters = append(allClusters, descOutput.Cluster)
                }
            }
            return allClusters, nil
        },
        Process: func(output interface{}, err error, debug bool) []types.ScanResult {
            var results []types.ScanResult

            if err != nil {
                utils.HandleAWSError(debug, "eks:ListClusters", err)
                return []types.ScanResult{
                    {
                        ServiceName: "EKS",
                        MethodName:  "eks:ListClusters",
                        Error:       err,
                        Timestamp:   time.Now(),
                    },
                }
            }

            clusters, ok := output.([]*eks.Cluster)
            if !ok {
                utils.HandleAWSError(debug, "eks:ListClusters", fmt.Errorf("unexpected output type %T", output))
                return results
            }

            if len(clusters) == 0 {
                utils.PrintAccessGranted(debug, "eks:ListClusters", "clusters")
                return results
            }

            for _, cluster := range clusters {
                name := ""
                if cluster.Name != nil {
                    name = *cluster.Name
                }

                arn := ""
                if cluster.Arn != nil {
                    arn = *cluster.Arn
                }

                status := ""
                if cluster.Status != nil {
                    status = *cluster.Status
                }

                version := ""
                if cluster.Version != nil {
                    version = *cluster.Version
                }

                endpoint := ""
                if cluster.Endpoint != nil {
                    endpoint = *cluster.Endpoint
                }

                roleArn := ""
                if cluster.RoleArn != nil {
                    roleArn = *cluster.RoleArn
                }

                vpcId := ""
                var subnetCount int
                var sgCount int
                if cluster.ResourcesVpcConfig != nil {
                    if cluster.ResourcesVpcConfig.VpcId != nil {
                        vpcId = *cluster.ResourcesVpcConfig.VpcId
                    }
                    subnetCount = len(cluster.ResourcesVpcConfig.SubnetIds)
                    sgCount = len(cluster.ResourcesVpcConfig.SecurityGroupIds)
                }

                results = append(results, types.ScanResult{
                    ServiceName:  "EKS",
                    MethodName:   "eks:ListClusters",
                    ResourceType: "cluster",
                    ResourceName: name,
                    Details: map[string]interface{}{
                        "Arn":      arn,
                        "Status":   status,
                        "Version":  version,
                        "Endpoint": endpoint,
                        "RoleArn":  roleArn,
                        "VpcId":    vpcId,
                        "Subnets":  subnetCount,
                        "SecurityGroups": sgCount,
                    },
                    Timestamp: time.Now(),
                })

                vpcInfo := fmt.Sprintf("VPC: %s, Subnets: %d, SGs: %d", vpcId, subnetCount, sgCount)
                utils.PrintResult(debug, "", "eks:ListClusters",
                    fmt.Sprintf("Found EKS Cluster: %s (Status: %s, Version: %s, Endpoint: %s, Role: %s, %s)",
                        utils.ColorizeItem(name), status, version, endpoint, roleArn, vpcInfo), nil)
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
    process := EKSCalls[0].Process

    tests := []struct {
        name          string
        input         interface{}
        err           error
        expectedCount int
        expectError   bool
        checkResults  func(t *testing.T, results []types.ScanResult)
    }{
        {
            name: "valid cluster with all fields",
            input: []*eks.Cluster{
                {
                    Name:     aws.String("my-cluster"),
                    Arn:      aws.String("arn:aws:eks:us-east-1:123456789012:cluster/my-cluster"),
                    Status:   aws.String("ACTIVE"),
                    Version:  aws.String("1.28"),
                    Endpoint: aws.String("https://ABC123.gr7.us-east-1.eks.amazonaws.com"),
                    RoleArn:  aws.String("arn:aws:iam::123456789012:role/eks-role"),
                    ResourcesVpcConfig: &eks.VpcConfigResponse{
                        VpcId:            aws.String("vpc-12345"),
                        SubnetIds:        []*string{aws.String("subnet-1"), aws.String("subnet-2")},
                        SecurityGroupIds: []*string{aws.String("sg-1")},
                    },
                },
            },
            expectedCount: 1,
            checkResults: func(t *testing.T, results []types.ScanResult) {
                // Verify ServiceName, MethodName, ResourceType, ResourceName, Details
            },
        },
        // ... more test cases
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
1. **Valid cluster with all fields** -- all fields populated, verify ScanResult fields and Details map
2. **Multiple clusters** -- verify correct count and each cluster's fields captured
3. **Empty results** -- verify `PrintAccessGranted` behavior and empty results returned
4. **Access denied** -- verify error ScanResult returned with correct ServiceName/MethodName
5. **Nil field handling** -- verify nil Name, nil ResourcesVpcConfig, nil Version etc. handled gracefully (empty string/zero defaults)

### Edge Cases

1. **No clusters in any region** -- Call() returns empty slice, Process() calls PrintAccessGranted, returns empty results
2. **Access denied on ListClusters** -- Call() returns error immediately (fail fast)
3. **Access denied on DescribeCluster** -- Call() returns error on first DescribeCluster failure (fail fast, matches established pattern)
4. **Cluster with nil ResourcesVpcConfig** -- defensive nil check on the entire struct before accessing nested fields
5. **Cluster in CREATING/DELETING state** -- still enumerate it, Status field will reflect the state
6. **Many clusters in one region** -- ListClusters may need pagination via NextToken (initial implementation without pagination is acceptable)

### Architecture Compliance

- **Package:** `eks` in `cmd/awtest/services/eks/` -- MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) -- MUST FOLLOW
- **Variable:** `EKSCalls` exported slice -- MUST FOLLOW
- **Type:** `[]types.AWSService` -- MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` -- MUST FOLLOW (epics say "EKS" but all Epic 2 stories use DefaultModuleName)
- **Session handling:** `sess.Copy(&aws.Config{Region: aws.String(region)})` -- MUST FOLLOW (Story 2.3 code review fix)
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` -- MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` -- MUST FOLLOW
- **Nil checks:** Always check `*string`, `*int64`, `*bool`, nested structs, and slice fields before dereferencing -- MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) -- MUST FOLLOW

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/eks/
+-- calls.go            # NEW: AWS EKS service implementation
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
cmd/awtest/services/ecs/calls.go               # ECS reference (two-API ListClusters+DescribeClusters pattern)
cmd/awtest/services/cognitouserpools/calls.go  # Story 2.2 reference (single API pattern)
go.mod                                         # AWS SDK already included
```

### Previous Story Intelligence (Story 2.4)

**Key learnings from Story 2.4 (EFS):**
- **sess.Copy() is mandatory** -- continued from Story 2.3 fix. All subsequent stories MUST use this pattern.
- Table-driven tests were required per code review (AC9) -- EKS tests should be table-driven from the start.
- Type assertion failure handling added to EFS -- EKS should include this from the start.
- All 6 display fields must appear in BOTH PrintResult AND Details map -- EKS has more fields (8 in Details map).
- `ScanResult.Timestamp = time.Now()` is required on every result.
- Empty results handled with `utils.PrintAccessGranted` -- same for EKS.

**Key learnings from Story 2.3 (Config):**
- Multiple API calls combined into single AWSService is the established pattern (Config called both DescribeConfigurationRecorders and DescribeConfigurationRecorderStatus).
- EKS similarly combines ListClusters + DescribeCluster.

**Key learnings from ECS (existing service):**
- ECS uses the SAME ListClusters + DescribeClusters two-call pattern.
- **BUT** ECS uses the OLD `session.NewSession(regionConfig)` pattern -- DO NOT COPY THIS.
- ECS uses `aws.StringValue()` and `aws.Int64Value()` helpers -- EKS can use either explicit nil checks (EFS pattern) or these helpers. The EFS nil-check pattern is preferred for consistency with recent Epic 2 stories.

### Git Intelligence

**Recent commits (Epic 2 context):**
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
- Package name `eks` follows convention (lowercase, single word, matches directory)
- Single `calls.go` file per service -- matches all 38+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/eks`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.5: EKS (Elastic Kubernetes Service) Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR7-31 Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#Phase 1 Additions Needed]
- [Source: _bmad-output/implementation-artifacts/2-4-efs-elastic-file-system-service-enumeration.md -- previous story learnings, sess.Copy() pattern, table-driven tests]
- [Source: cmd/awtest/services/efs/calls.go -- Story 2.4 reference (sess.Copy + single API pattern)]
- [Source: cmd/awtest/services/ecs/calls.go -- ECS reference (two-API ListClusters+DescribeClusters pattern, BUT uses old session pattern)]
- [Source: cmd/awtest/services/services.go -- AllServices() registration point]
- [Source: cmd/awtest/types/types.go -- AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go -- PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem]
- [Source: go.mod -- aws-sdk-go v1.44.266 (includes eks package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Implemented EKS service enumeration with two-API-call pattern (ListClusters + DescribeCluster) using sess.Copy() for safe region iteration
- Process() extracts all required fields (Name, Arn, Status, Version, Endpoint, RoleArn, VpcId, SubnetIds count, SecurityGroupIds count) with full nil safety including nil ResourcesVpcConfig handling
- Type assertion failure handling included from the start (EFS pattern)
- Registered in AllServices() alphabetically after EFS, before ElasticBeanstalk
- 5 table-driven test cases covering: valid cluster with all fields, multiple clusters, empty results, access denied, nil field handling
- All verification passed: go build, go test (all pass), go vet (clean), full regression suite (no regressions)
- Code review fix: Changed DescribeCluster error handling from fail-fast (return nil, err) to resilient (continue) so a single cluster failure doesn't abort the entire scan

### Change Log

- 2026-03-05: Implemented Story 2.5 - EKS Elastic Kubernetes Service enumeration
- 2026-03-05: Addressed code review - 1 medium issue resolved (resilient DescribeCluster error handling), 2 low issues acknowledged as by-design (pagination out of scope per story, Call() tests follow established Epic 2 Process()-only pattern)

### File List

- cmd/awtest/services/eks/calls.go (NEW)
- cmd/awtest/services/eks/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED)
