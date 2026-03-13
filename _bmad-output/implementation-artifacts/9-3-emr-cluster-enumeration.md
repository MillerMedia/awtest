# Story 9.3: EMR Cluster Enumeration

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate EMR clusters, instance groups, and security configurations,
So that I can discover big data processing infrastructure and potential data access paths.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/emr/` directory with `calls.go` implementing EMR service enumeration with 3 AWSService entries.

2. **AC2:** Implement `emr:ListClusters` API call — iterates all regions in `types.Regions`, creates EMR client per region using config override pattern (`emr.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListClustersWithContext` with `Marker` pagination to list all cluster summaries. For each cluster, calls `DescribeClusterWithContext` to retrieve full details including ReleaseLabel. Each cluster listed with ClusterId, Name, State, ReleaseLabel, and Region.

3. **AC3:** Implement `emr:ListInstanceGroups` API call — iterates all regions in `types.Regions`, creates EMR client per region using config override pattern. First lists all clusters via `ListClustersWithContext` with `Marker` pagination, then for each cluster calls `ListInstanceGroupsWithContext` with `Marker` pagination. Each instance group listed with Id, Name, ClusterId, ClusterName, InstanceGroupType, InstanceType, RequestedInstanceCount, RunningInstanceCount, Market, State, and Region.

4. **AC4:** Implement `emr:ListSecurityConfigurations` API call — iterates all regions in `types.Regions`, creates EMR client per region using config override pattern, calls `ListSecurityConfigurationsWithContext` with `Marker` pagination. Each security configuration listed with Name, CreationDateTime, and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without EMR access, EMR is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register EMR service in `services/services.go` `AllServices()` function in alphabetical order (after `eks`, before `elasticache`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 3 API calls.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/emr/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `emr:ListClusters` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/emr/`
  - [x] Create `calls.go` with `package emr`
  - [x] Define `var EMRCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"emr:ListClusters"`
  - [x] Call: iterate `types.Regions`, create `emr.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `ListClustersWithContext` with `Marker` pagination (NOT NextToken — EMR uses `Marker`). For each cluster summary, call `DescribeClusterWithContext` to get full details including ReleaseLabel. Define local struct `emrCluster` with fields: ClusterId, Name, State, ReleaseLabel, Region. Per-region errors: `break` pagination loop, don't abort scan.
  - [x] Implement `extractCluster` helper function — takes `*emr.Cluster` and `region` string, returns `emrCluster` with nil-safe pointer dereferencing. Note: `Status.State` is nested — extract via `cluster.Status.State` with nil checks at each level.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]emrCluster`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"EMR"`, ResourceType=`"cluster"`, ResourceName=clusterName
  - [x] `utils.PrintResult` format: `"EMR Cluster: %s (State: %s, Release: %s, Region: %s)"` with `utils.ColorizeItem(clusterName)`

- [x] Task 2: Implement `emr:ListInstanceGroups` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"emr:ListInstanceGroups"`
  - [x] Call: iterate regions -> create EMR client with config override -> Step 1: list all clusters via `ListClustersWithContext` with `Marker` pagination (collect cluster IDs and names) -> Step 2: for each cluster, call `ListInstanceGroupsWithContext` with `Marker` pagination. Define local struct `emrInstanceGroup` with fields: Id, Name, ClusterId, ClusterName, InstanceGroupType, InstanceType, RequestedInstanceCount, RunningInstanceCount, Market, State, Region. Per-region errors: `break` to next region for cluster listing; per-cluster errors: `continue` to next cluster.
  - [x] Implement `extractInstanceGroup` helper function — takes `*emr.InstanceGroup`, `clusterId`, `clusterName`, and `region` string, returns `emrInstanceGroup` with nil-safe pointer dereferencing. Note: RequestedInstanceCount and RunningInstanceCount are `*int64` — convert to string with `fmt.Sprintf("%d", *ptr)`. State is nested at `ig.Status.State` — nil-check each level.
  - [x] Process: type-assert `[]emrInstanceGroup`, build `ScanResult` with ServiceName=`"EMR"`, ResourceType=`"instance-group"`, ResourceName=igName
  - [x] `utils.PrintResult` format: `"EMR Instance Group: %s (Cluster: %s, Type: %s, Instance: %s, Count: %s, Region: %s)"` with `utils.ColorizeItem(igName)`

- [x] Task 3: Implement `emr:ListSecurityConfigurations` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"emr:ListSecurityConfigurations"`
  - [x] Call: iterate regions -> create EMR client with config override -> call `ListSecurityConfigurationsWithContext` with `Marker` pagination. Define local struct `emrSecurityConfig` with fields: Name, CreationDateTime, Region. Per-region errors: `break` pagination loop.
  - [x] Implement `extractSecurityConfig` helper function — takes `*emr.SecurityConfigurationSummary` and `region` string, returns `emrSecurityConfig` with nil-safe pointer dereferencing. Note: CreationDateTime is `*time.Time` — format with `time.RFC3339`.
  - [x] Process: type-assert `[]emrSecurityConfig`, build `ScanResult` with ServiceName=`"EMR"`, ResourceType=`"security-configuration"`, ResourceName=configName
  - [x] `utils.PrintResult` format: `"EMR Security Configuration: %s (Created: %s, Region: %s)"` with `utils.ColorizeItem(configName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/emr"` to `services/services.go` (alphabetical in imports: after `eks`, before `elasticache`)
  - [x] Add `allServices = append(allServices, emr.EMRCalls...)` after `eks.EKSCalls...` and before `elasticache.ElastiCacheCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/emr/calls_test.go`
  - [x] Test `ListClusters` Process: valid clusters with details (ID, name, state, release label), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListInstanceGroups` Process: valid instance groups with details (ID, name, cluster ID, cluster name, type, instance type, counts, market, state), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListSecurityConfigurations` Process: valid configs with details (name, creation date), empty results, error handling, nil fields, type assertion failure
  - [x] Test extract helpers: `TestExtractCluster`, `TestExtractInstanceGroup`, `TestExtractSecurityConfig` with AWS SDK types (both populated and nil fields)
  - [x] Use table-driven tests with `t.Run` subtests following CodeDeploy/DirectConnect test pattern
  - [x] Access Process via `EMRCalls[0].Process`, `EMRCalls[1].Process`, `EMRCalls[2].Process`

- [x] Task 6: Vendor EMR SDK package (AC: 10)
  - [x] Run `go mod vendor` or manually ensure `vendor/github.com/aws/aws-sdk-go/service/emr/` is populated
  - [x] EMR package is part of `aws-sdk-go v1.44.266` — already in go.mod, just needs vendoring

- [x] Task 7: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/emr/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := emr.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := emr.New(sess)
}
```

### CRITICAL: EMR Uses `Marker` Pagination, NOT `NextToken`

Unlike CodeDeploy and DirectConnect which use `NextToken`, EMR APIs use `Marker` for pagination. This is a key difference:

```go
var marker *string
for {
    input := &emr.ListClustersInput{}
    if marker != nil {
        input.Marker = marker
    }
    output, err := svc.ListClustersWithContext(ctx, input)
    if err != nil {
        lastErr = err
        utils.HandleAWSError(false, "emr:ListClusters", err)
        break
    }
    for _, cluster := range output.Clusters {
        if cluster != nil {
            // Process cluster...
        }
    }
    if output.Marker == nil {
        break
    }
    marker = output.Marker
}
```

### EMR SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/emr`

**IMPORTANT:** The Go package name is `emr`. The local package name is also `emr`, same pattern as `codedeploy`/`directconnect` where the local package name matches the AWS SDK package name. Within `calls.go`, `emr.New()` and `emr.ListClustersInput{}` refer to the **AWS SDK package**, while local types (structs, variables) are referenced directly without package prefix.

**API Methods:**

1. **ListClusters (Paginated with Marker, regional):**
   - `svc.ListClustersWithContext(ctx, &emr.ListClustersInput{Marker: marker})` -> `*emr.ListClustersOutput`
   - `.Clusters` -> `[]*emr.ClusterSummary`
   - Pagination: `Marker *string` in both input and output
   - Each `ClusterSummary` has:
     - `Id *string` (e.g., "j-XXXXXXXXXXXXX")
     - `Name *string`
     - `Status *emr.ClusterStatus` → `.State *string` ("STARTING", "BOOTSTRAPPING", "RUNNING", "WAITING", "TERMINATING", "TERMINATED", "TERMINATED_WITH_ERRORS")
     - `NormalizedInstanceHours *int64`
     - `ClusterArn *string`
   - **Note:** `ClusterSummary` does NOT have `ReleaseLabel` — requires `DescribeCluster` per cluster

2. **DescribeCluster (Per-cluster detail, same region):**
   - `svc.DescribeClusterWithContext(ctx, &emr.DescribeClusterInput{ClusterId: aws.String(id)})` -> `*emr.DescribeClusterOutput`
   - `.Cluster` -> `*emr.Cluster`
   - Each `Cluster` has:
     - `Id *string`
     - `Name *string`
     - `Status *emr.ClusterStatus` → `.State *string`
     - `ReleaseLabel *string` (e.g., "emr-6.10.0")
     - `ServiceRole *string`
     - `AutoTerminate *bool`
     - `LogUri *string`
     - `Applications []*emr.Application` (e.g., Spark, Hive, Hadoop)
     - `Ec2InstanceAttributes *emr.Ec2InstanceAttributes`

3. **ListInstanceGroups (Paginated with Marker, per-cluster):**
   - `svc.ListInstanceGroupsWithContext(ctx, &emr.ListInstanceGroupsInput{ClusterId: aws.String(id), Marker: marker})` -> `*emr.ListInstanceGroupsOutput`
   - `.InstanceGroups` -> `[]*emr.InstanceGroup`
   - Pagination: `Marker *string` in both input and output
   - Each `InstanceGroup` has:
     - `Id *string` (e.g., "ig-XXXXXXXXXXXXX")
     - `Name *string` (e.g., "Master", "Core")
     - `InstanceGroupType *string` ("MASTER", "CORE", "TASK")
     - `InstanceType *string` (e.g., "m5.xlarge")
     - `RequestedInstanceCount *int64`
     - `RunningInstanceCount *int64`
     - `Market *string` ("ON_DEMAND", "SPOT")
     - `Status *emr.InstanceGroupStatus` → `.State *string` ("PROVISIONING", "BOOTSTRAPPING", "RUNNING", "RECONFIGURING", "RESIZING", "SUSPENDED", "TERMINATING", "TERMINATED", "ARRESTED", "SHUTTING_DOWN", "ENDED")

4. **ListSecurityConfigurations (Paginated with Marker, regional):**
   - `svc.ListSecurityConfigurationsWithContext(ctx, &emr.ListSecurityConfigurationsInput{Marker: marker})` -> `*emr.ListSecurityConfigurationsOutput`
   - `.SecurityConfigurations` -> `[]*emr.SecurityConfigurationSummary`
   - Pagination: `Marker *string` in both input and output
   - Each `SecurityConfigurationSummary` has:
     - `Name *string`
     - `CreationDateTime *time.Time`

**EMR SDK package must be vendored** — not yet present in `vendor/`. The package is part of `aws-sdk-go v1.44.266` already in `go.mod`. Run `go mod vendor` to populate.

### List + Describe Pattern (Call 1: ListClusters)

Call 1 uses a **list-then-describe** pattern to get full cluster details including ReleaseLabel:

```go
var allClusters []emrCluster
var lastErr error

for _, region := range types.Regions {
    svc := emr.New(sess, &aws.Config{Region: aws.String(region)})
    var marker *string
    for {
        input := &emr.ListClustersInput{}
        if marker != nil {
            input.Marker = marker
        }
        output, err := svc.ListClustersWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "emr:ListClusters", err)
            break
        }
        for _, summary := range output.Clusters {
            if summary == nil || summary.Id == nil {
                continue
            }
            // Describe each cluster to get ReleaseLabel
            descOutput, err := svc.DescribeClusterWithContext(ctx, &emr.DescribeClusterInput{
                ClusterId: summary.Id,
            })
            if err != nil {
                utils.HandleAWSError(false, "emr:ListClusters", err)
                continue
            }
            if descOutput.Cluster != nil {
                allClusters = append(allClusters, extractCluster(descOutput.Cluster, region))
            }
        }
        if output.Marker == nil {
            break
        }
        marker = output.Marker
    }
}
```

### List Clusters Then Instance Groups Pattern (Call 2)

Call 2 first lists clusters, then for each cluster lists instance groups — same pattern as CodeDeploy's ListDeploymentGroups:

```go
var allInstanceGroups []emrInstanceGroup
var lastErr error

for _, region := range types.Regions {
    svc := emr.New(sess, &aws.Config{Region: aws.String(region)})

    // Step 1: List all cluster IDs and names
    var clusterIds []struct{ Id, Name string }
    var clusterMarker *string
    for {
        input := &emr.ListClustersInput{}
        if clusterMarker != nil {
            input.Marker = clusterMarker
        }
        output, err := svc.ListClustersWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "emr:ListInstanceGroups", err)
            break
        }
        for _, c := range output.Clusters {
            if c != nil && c.Id != nil {
                name := ""
                if c.Name != nil {
                    name = *c.Name
                }
                clusterIds = append(clusterIds, struct{ Id, Name string }{*c.Id, name})
            }
        }
        if output.Marker == nil {
            break
        }
        clusterMarker = output.Marker
    }

    // Step 2: For each cluster, list instance groups
    for _, cluster := range clusterIds {
        var igMarker *string
        for {
            igInput := &emr.ListInstanceGroupsInput{
                ClusterId: aws.String(cluster.Id),
            }
            if igMarker != nil {
                igInput.Marker = igMarker
            }
            igOutput, err := svc.ListInstanceGroupsWithContext(ctx, igInput)
            if err != nil {
                utils.HandleAWSError(false, "emr:ListInstanceGroups", err)
                break
            }
            for _, ig := range igOutput.InstanceGroups {
                if ig != nil {
                    allInstanceGroups = append(allInstanceGroups, extractInstanceGroup(ig, cluster.Id, cluster.Name, region))
                }
            }
            if igOutput.Marker == nil {
                break
            }
            igMarker = igOutput.Marker
        }
    }
}
```

### Simple Paginated Pattern (Call 3: ListSecurityConfigurations)

```go
var allConfigs []emrSecurityConfig
var lastErr error

for _, region := range types.Regions {
    svc := emr.New(sess, &aws.Config{Region: aws.String(region)})
    var marker *string
    for {
        input := &emr.ListSecurityConfigurationsInput{}
        if marker != nil {
            input.Marker = marker
        }
        output, err := svc.ListSecurityConfigurationsWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "emr:ListSecurityConfigurations", err)
            break
        }
        for _, cfg := range output.SecurityConfigurations {
            if cfg != nil {
                allConfigs = append(allConfigs, extractSecurityConfig(cfg, region))
            }
        }
        if output.Marker == nil {
            break
        }
        marker = output.Marker
    }
}
```

### Nil-Safe Field Extraction Helpers

```go
func extractCluster(cluster *emr.Cluster, region string) emrCluster {
    id := ""
    if cluster.Id != nil {
        id = *cluster.Id
    }
    name := ""
    if cluster.Name != nil {
        name = *cluster.Name
    }
    state := ""
    if cluster.Status != nil && cluster.Status.State != nil {
        state = *cluster.Status.State
    }
    releaseLabel := ""
    if cluster.ReleaseLabel != nil {
        releaseLabel = *cluster.ReleaseLabel
    }
    return emrCluster{
        ClusterId:    id,
        Name:         name,
        State:        state,
        ReleaseLabel: releaseLabel,
        Region:       region,
    }
}

func extractInstanceGroup(ig *emr.InstanceGroup, clusterId, clusterName, region string) emrInstanceGroup {
    id := ""
    if ig.Id != nil {
        id = *ig.Id
    }
    name := ""
    if ig.Name != nil {
        name = *ig.Name
    }
    igType := ""
    if ig.InstanceGroupType != nil {
        igType = *ig.InstanceGroupType
    }
    instanceType := ""
    if ig.InstanceType != nil {
        instanceType = *ig.InstanceType
    }
    requestedCount := ""
    if ig.RequestedInstanceCount != nil {
        requestedCount = fmt.Sprintf("%d", *ig.RequestedInstanceCount)
    }
    runningCount := ""
    if ig.RunningInstanceCount != nil {
        runningCount = fmt.Sprintf("%d", *ig.RunningInstanceCount)
    }
    market := ""
    if ig.Market != nil {
        market = *ig.Market
    }
    state := ""
    if ig.Status != nil && ig.Status.State != nil {
        state = *ig.Status.State
    }
    return emrInstanceGroup{
        Id:                     id,
        Name:                   name,
        ClusterId:              clusterId,
        ClusterName:            clusterName,
        InstanceGroupType:      igType,
        InstanceType:           instanceType,
        RequestedInstanceCount: requestedCount,
        RunningInstanceCount:   runningCount,
        Market:                 market,
        State:                  state,
        Region:                 region,
    }
}

func extractSecurityConfig(cfg *emr.SecurityConfigurationSummary, region string) emrSecurityConfig {
    name := ""
    if cfg.Name != nil {
        name = *cfg.Name
    }
    creationDateTime := ""
    if cfg.CreationDateTime != nil {
        creationDateTime = cfg.CreationDateTime.Format(time.RFC3339)
    }
    return emrSecurityConfig{
        Name:             name,
        CreationDateTime: creationDateTime,
        Region:           region,
    }
}
```

### Local Struct Definitions

```go
type emrCluster struct {
    ClusterId    string
    Name         string
    State        string
    ReleaseLabel string
    Region       string
}

type emrInstanceGroup struct {
    Id                     string
    Name                   string
    ClusterId              string
    ClusterName            string
    InstanceGroupType      string
    InstanceType           string
    RequestedInstanceCount string
    RunningInstanceCount   string
    Market                 string
    State                  string
    Region                 string
}

type emrSecurityConfig struct {
    Name             string
    CreationDateTime string
    Region           string
}
```

### Variable & Naming Conventions

- **Package:** `emr` (directory: `cmd/awtest/services/emr/`)
- **Exported variable:** `EMRCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"emr:ListClusters"`, `"emr:ListInstanceGroups"`, `"emr:ListSecurityConfigurations"`
- **ScanResult.ServiceName:** `"EMR"` (all-caps acronym, following EC2/ECR/ECS/EFS/EKS convention)
- **ScanResult.ResourceType:** `"cluster"`, `"instance-group"`, `"security-configuration"` (lowercase hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **Local struct prefix:** `emr` (for EMR, following `cd` for CodeDeploy, `dc` for DirectConnect pattern)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/emr"` (same name as local package — handled same as codedeploy/directconnect pattern)

### Registration Order in services.go

Insert alphabetically — `emr` comes after `eks`, before `elasticache`:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/eks"
"github.com/MillerMedia/awtest/cmd/awtest/services/emr"              // NEW — after eks, before elasticache
"github.com/MillerMedia/awtest/cmd/awtest/services/elasticache"

// In AllServices():
allServices = append(allServices, eks.EKSCalls...)
allServices = append(allServices, emr.EMRCalls...)                    // NEW — after eks, before elasticache
allServices = append(allServices, elasticache.ElastiCacheCalls...)
```

### Testing Pattern

Follow the CodeDeploy/DirectConnect test pattern — test Process() functions only with pre-built mock data:

```go
func TestListClustersProcess(t *testing.T) {
    process := EMRCalls[0].Process
    // Table-driven tests: valid clusters (ID, name, state, release label), empty, errors, nil fields, type assertion failure
}

func TestListInstanceGroupsProcess(t *testing.T) {
    process := EMRCalls[1].Process
    // Table-driven tests: valid IGs (ID, name, cluster ID, cluster name, type, instance type, counts, market, state), empty, errors, nil fields, type assertion failure
}

func TestListSecurityConfigurationsProcess(t *testing.T) {
    process := EMRCalls[2].Process
    // Table-driven tests: valid configs (name, creation date), empty, errors, nil fields, type assertion failure
}
```

Include extract helper tests with AWS SDK types:
```go
func TestExtractCluster(t *testing.T) { ... }
func TestExtractInstanceGroup(t *testing.T) { ... }
func TestExtractSecurityConfig(t *testing.T) { ... }
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `emr/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `emr.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** use `NextToken` — EMR uses `Marker` for pagination (not NextToken like CodeDeploy/DirectConnect)
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** skip `DescribeCluster` for Call 1 — `ClusterSummary` from `ListClusters` does NOT have `ReleaseLabel`
- **DO NOT** confuse `emr.ListClustersInput` (AWS SDK type) with local package types — AWS SDK `emr` is the imported package, local types are referenced without prefix

### Key Differences from Previous Stories (9.1 CodeDeploy, 9.2 DirectConnect)

1. **Marker pagination (not NextToken):** EMR uses `Marker` for all paginated APIs, not `NextToken` like CodeDeploy/DirectConnect. The pagination loop variable must be named `marker` and use `input.Marker` / `output.Marker`.
2. **List + Describe pattern (Call 1):** Unlike DirectConnect's non-paginated calls or CodeDeploy's batch-get, EMR requires calling `DescribeCluster` individually per cluster to get `ReleaseLabel`. This is because `ClusterSummary` from `ListClusters` does not include `ReleaseLabel`.
3. **Nested Status fields:** EMR `Status` fields are nested structs — `cluster.Status.State` requires nil checks at both `Status` and `State` levels. Same for `ig.Status.State` on instance groups.
4. **Integer pointer fields:** RequestedInstanceCount (`*int64`) and RunningInstanceCount (`*int64`) on instance groups require `fmt.Sprintf("%d", *ptr)` conversion (same as DirectConnect's Vlan/Asn).
5. **No batch-get API:** EMR does not have a BatchGet API. Use `DescribeCluster` individually for each cluster. ListInstanceGroups is already per-cluster.
6. **No global resources:** All EMR resources are regional — iterate `types.Regions` for all 3 API calls (unlike DirectConnect Call 3 which was global).

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/emr/
+-- calls.go            # EMR service implementation (3 AWSService entries)
+-- calls_test.go       # Process() tests + extract helper tests for all 3 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                     # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                    # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/codedeploy/calls.go       # Reference implementation (list + batch-get pattern)
cmd/awtest/services/codedeploy/calls_test.go  # Reference test pattern (extract helper tests)
cmd/awtest/services/directconnect/calls.go    # Reference implementation (Marker-like pagination)
cmd/awtest/services/directconnect/calls_test.go # Reference test pattern
go.mod                                        # AWS SDK already includes emr package (needs vendoring)
```

**Vendor directory to POPULATE:**
```
vendor/github.com/aws/aws-sdk-go/service/emr/  # Run go mod vendor to populate
```

### Previous Story Intelligence

**From Story 9.2 (DirectConnect — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- Extract helper functions for nil-safe extraction — directly applicable
- Integer pointer fields (Vlan, Asn, AmazonSideAsn) used `fmt.Sprintf("%d", *ptr)` — same pattern needed for RequestedInstanceCount/RunningInstanceCount
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `DirectConnectCalls[N].Process` in tests -> apply as `EMRCalls[N].Process`
- Error result pattern: `return []types.ScanResult{{ServiceName: "EMR", MethodName: "emr:ListClusters", Error: err, Timestamp: time.Now()}}`
- 24 tests across 6 test functions in DirectConnect

**From Story 9.1 (CodeDeploy — list+describe pattern reference):**
- List-then-describe pattern (ListApplications -> BatchGetApplications) — EMR uses similar pattern (ListClusters -> DescribeCluster)
- Nested API calls: ListDeploymentGroups first lists apps, then per app lists groups — EMR uses same pattern for ListInstanceGroups (first lists clusters, then per cluster lists IGs)
- `maxBatchSize` constant — **not applicable** (EMR has no batch APIs)
- 23 tests across 7 test functions in CodeDeploy

**From Code Review Findings (Stories 7.1, 7.2):**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [HIGH] Always add pagination from the start — applies to all 3 EMR calls (all use Marker)
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [LOW] Tests should cover nil fields comprehensively
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `b7a8967` — Add Direct Connect enumeration with 3 API calls (Story 9.2)
- `7b02834` — Add CodeDeploy enumeration with 3 API calls (Story 9.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add EMR enumeration with 3 API calls (Story 9.3)"`

### FRs Covered

- **FR110:** System enumerates EMR clusters, instance groups, and security configurations

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### References

- [Source: epics-phase2.md#Story 4.3: EMR Cluster Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR110] — EMR enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/codedeploy/calls.go] — Reference implementation (list + describe pattern, nested API calls)
- [Source: cmd/awtest/services/codedeploy/calls_test.go] — Reference test pattern (extract helper tests)
- [Source: cmd/awtest/services/directconnect/calls.go] — Reference implementation (integer pointer fields)
- [Source: cmd/awtest/services/directconnect/calls_test.go] — Reference test pattern
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (emr goes after eks, before elasticache)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes emr package, needs vendoring)
- [Source: 9-2-direct-connect-enumeration.md] — Previous story (integer pointer conversion, extract helpers)
- [Source: 9-1-codedeploy-enumeration.md] — Reference story (list + describe pattern)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no debugging needed.

### Completion Notes List

- Implemented EMR service enumeration with 3 AWSService entries: ListClusters, ListInstanceGroups, ListSecurityConfigurations
- ListClusters uses list-then-describe pattern (ListClustersWithContext + DescribeClusterWithContext) to get ReleaseLabel
- ListInstanceGroups uses nested list pattern (list clusters, then per-cluster list instance groups)
- ListSecurityConfigurations uses simple paginated regional pattern
- All 3 calls use Marker pagination (not NextToken), config override for region (not session mutation)
- 3 extract helper functions with nil-safe pointer dereferencing including nested Status.State checks
- 6 test functions: 3 Process tests + 3 extract helper tests, 24 test cases total
- Registered in AllServices() alphabetically after eks, before elasticache
- EMR SDK vendored from existing aws-sdk-go v1.44.266
- All builds, tests, vet, and race detection pass clean

### File List

- cmd/awtest/services/emr/calls.go (NEW)
- cmd/awtest/services/emr/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED — added EMR import and registration)
- vendor/github.com/aws/aws-sdk-go/service/emr/ (NEW — vendored SDK package)

### Change Log

- 2026-03-13: Implemented EMR cluster enumeration with 3 API calls (ListClusters, ListInstanceGroups, ListSecurityConfigurations), registered in AllServices(), added 24 unit tests across 6 test functions
