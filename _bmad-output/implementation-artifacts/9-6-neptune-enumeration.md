# Story 9.6: Neptune Enumeration

Status: done

<!-- Generated: 2026-03-13 by BMAD Create Story Workflow -->
<!-- Epic: 9 - Infrastructure & Data Service Expansion (Phase 2 Epic 4) -->
<!-- FR: FR113 | Source: epics-phase2.md#Story 4.6 -->
<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate Neptune DB clusters, instances, and parameter groups,
So that I can discover graph database infrastructure and access configurations.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/neptune/` directory with `calls.go` implementing Neptune service enumeration with 3 AWSService entries.

2. **AC2:** Implement `neptune:DescribeDBClusters` API call — iterates all regions in `types.Regions`, creates Neptune client per region using config override pattern (`neptune.New(sess, &aws.Config{Region: aws.String(region)})`), calls `DescribeDBClustersWithContext` with `Marker` pagination. Each cluster listed with DBClusterIdentifier, DBClusterArn, Status, Engine, EngineVersion, Endpoint, ReaderEndpoint, Port, MultiAZ, StorageEncrypted, KmsKeyId, DeletionProtection, IAMDatabaseAuthenticationEnabled, DBClusterParameterGroup, ClusterCreateTime, and Region.

3. **AC3:** Implement `neptune:DescribeDBInstances` API call — iterates all regions in `types.Regions`, creates Neptune client per region using config override pattern, calls `DescribeDBInstancesWithContext` with `Marker` pagination. Each instance listed with DBInstanceIdentifier, DBInstanceArn, DBInstanceClass, Engine, EngineVersion, DBInstanceStatus, EndpointAddress, EndpointPort, DBClusterIdentifier, AvailabilityZone, PubliclyAccessible, StorageEncrypted, AutoMinorVersionUpgrade, and Region.

4. **AC4:** Implement `neptune:DescribeDBClusterParameterGroups` API call — iterates all regions in `types.Regions`, creates Neptune client per region using config override pattern, calls `DescribeDBClusterParameterGroupsWithContext` with `Marker` pagination. Each parameter group listed with DBClusterParameterGroupName, DBClusterParameterGroupArn, Description, DBParameterGroupFamily, and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without Neptune access, Neptune is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register Neptune service in `services/services.go` `AllServices()` function in alphabetical order (after `mediaconvert`, before `opensearch`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 3 API calls.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/neptune/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `neptune:DescribeDBClusters` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/neptune/`
  - [x] Create `calls.go` with `package neptune`
  - [x] Define `var NeptuneCalls = []types.AWSService{...}` with 3 entries
  - [x] Define local struct `npCluster` with fields: DBClusterIdentifier, DBClusterArn, Status, Engine, EngineVersion, Endpoint, ReaderEndpoint, Port, MultiAZ, StorageEncrypted, KmsKeyId, DeletionProtection, IAMDatabaseAuthenticationEnabled, DBClusterParameterGroup, ClusterCreateTime, Region
  - [x] Implement `extractCluster` helper — takes `*neptune.DBCluster` and `region` string, returns `npCluster` with nil-safe pointer dereferencing. `ClusterCreateTime` is `*time.Time` — format with `time.RFC3339`. `Port` is `*int64` — format with `fmt.Sprintf("%d", *field)`. Boolean fields (`MultiAZ`, `StorageEncrypted`, `DeletionProtection`, `IAMDatabaseAuthenticationEnabled`) are `*bool` — format with `fmt.Sprintf("%t", *field)`.
  - [x] Implement first entry: Name `"neptune:DescribeDBClusters"`
  - [x] Call: iterate `types.Regions`, create `neptune.New(sess, &aws.Config{Region: aws.String(region)})`, call `DescribeDBClustersWithContext` with `Marker` pagination. Use `output.DBClusters` to get cluster list. Per-region errors: `break` pagination loop, don't abort scan.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]npCluster`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"Neptune"`, ResourceType=`"db-cluster"`, ResourceName=clusterIdentifier
  - [x] `utils.PrintResult` format: `"Neptune DB Cluster: %s (Status: %s, Engine: %s %s, Region: %s)"` with `utils.ColorizeItem(clusterIdentifier)`

- [x] Task 2: Implement `neptune:DescribeDBInstances` (AC: 3, 5, 6, 9)
  - [x] Define local struct `npInstance` with fields: DBInstanceIdentifier, DBInstanceArn, DBInstanceClass, Engine, EngineVersion, DBInstanceStatus, EndpointAddress, EndpointPort, DBClusterIdentifier, AvailabilityZone, PubliclyAccessible, StorageEncrypted, AutoMinorVersionUpgrade, Region
  - [x] Implement `extractInstance` helper — takes `*neptune.DBInstance` and `region` string, returns `npInstance` with nil-safe pointer dereferencing. **IMPORTANT:** `Endpoint` is a nested struct (`*neptune.Endpoint`) — extract `Endpoint.Address` and `Endpoint.Port` with double nil check (`if instance.Endpoint != nil { if instance.Endpoint.Address != nil { ... } }`). Boolean fields (`PubliclyAccessible`, `StorageEncrypted`, `AutoMinorVersionUpgrade`) are `*bool` — format with `fmt.Sprintf("%t", *field)`.
  - [x] Implement second entry: Name `"neptune:DescribeDBInstances"`
  - [x] Call: iterate regions -> create Neptune client with config override -> call `DescribeDBInstancesWithContext` with `Marker` pagination. Use `output.DBInstances` to get instance list. Per-region errors: `break` pagination loop.
  - [x] Process: type-assert `[]npInstance`, build `ScanResult` with ServiceName=`"Neptune"`, ResourceType=`"db-instance"`, ResourceName=instanceIdentifier
  - [x] `utils.PrintResult` format: `"Neptune DB Instance: %s (Status: %s, Class: %s, Cluster: %s, Region: %s)"` with `utils.ColorizeItem(instanceIdentifier)`

- [x] Task 3: Implement `neptune:DescribeDBClusterParameterGroups` (AC: 4, 5, 6, 9)
  - [x] Define local struct `npParameterGroup` with fields: DBClusterParameterGroupName, DBClusterParameterGroupArn, Description, DBParameterGroupFamily, Region
  - [x] Implement `extractParameterGroup` helper — takes `*neptune.DBClusterParameterGroup` and `region` string, returns `npParameterGroup` with nil-safe pointer dereferencing. All fields are `*string` — straightforward nil checks.
  - [x] Implement third entry: Name `"neptune:DescribeDBClusterParameterGroups"`
  - [x] Call: iterate regions -> create Neptune client with config override -> call `DescribeDBClusterParameterGroupsWithContext` with `Marker` pagination. Use `output.DBClusterParameterGroups` to get parameter group list. Per-region errors: `break` pagination loop.
  - [x] Process: type-assert `[]npParameterGroup`, build `ScanResult` with ServiceName=`"Neptune"`, ResourceType=`"db-cluster-parameter-group"`, ResourceName=parameterGroupName
  - [x] `utils.PrintResult` format: `"Neptune Cluster Parameter Group: %s (Family: %s, Region: %s)"` with `utils.ColorizeItem(parameterGroupName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/neptune"` to `services/services.go` (alphabetical in imports: after `mediaconvert`, before `opensearch`)
  - [x] Add `allServices = append(allServices, neptune.NeptuneCalls...)` after `mediaconvert.MediaConvertCalls...` and before `opensearch.OpenSearchCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/neptune/calls_test.go`
  - [x] Test `DescribeDBClusters` Process: valid clusters with details (identifier, ARN, status, engine, version, endpoint, reader endpoint, port, multi-AZ, encrypted, KMS key, deletion protection, IAM auth, parameter group, create time), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `DescribeDBInstances` Process: valid instances with details (identifier, ARN, class, engine, version, status, endpoint address/port, cluster, AZ, publicly accessible, encrypted, auto upgrade), empty results, error handling, nil fields, type assertion failure
  - [x] Test `DescribeDBClusterParameterGroups` Process: valid parameter groups with details (name, ARN, description, family), empty results, error handling, nil fields, type assertion failure
  - [x] Test extract helpers: `TestExtractCluster`, `TestExtractInstance`, `TestExtractParameterGroup` with AWS SDK types (both populated and nil fields)
  - [x] Use table-driven tests with `t.Run` subtests following MediaConvert/Kinesis test pattern
  - [x] Access Process via `NeptuneCalls[0].Process`, `NeptuneCalls[1].Process`, `NeptuneCalls[2].Process`

- [x] Task 6: Vendor Neptune SDK package (AC: 10)
  - [x] Run `go mod vendor` or manually ensure `vendor/github.com/aws/aws-sdk-go/service/neptune/` is populated
  - [x] Neptune package is part of `aws-sdk-go v1.44.266` — already in go.mod, just needs vendoring

- [x] Task 7: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/neptune/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := neptune.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := neptune.New(sess)
}
```

### CRITICAL: Neptune Uses Marker-Based Pagination (NOT NextToken)

Neptune uses **Marker-based pagination** (same pattern as RDS), NOT the NextToken pattern used by MediaConvert/Kinesis. The pagination token field is called `Marker` in both input and output structs.

```go
var allClusters []npCluster
var lastErr error

for _, region := range types.Regions {
    svc := neptune.New(sess, &aws.Config{Region: aws.String(region)})
    var marker *string
    for {
        input := &neptune.DescribeDBClustersInput{}
        if marker != nil {
            input.Marker = marker
        }
        output, err := svc.DescribeDBClustersWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "neptune:DescribeDBClusters", err)
            break
        }
        for _, cluster := range output.DBClusters {
            if cluster != nil {
                allClusters = append(allClusters, extractCluster(cluster, region))
            }
        }
        if output.Marker == nil {
            break
        }
        marker = output.Marker
    }
}

if len(allClusters) == 0 && lastErr != nil {
    return nil, lastErr
}
return allClusters, nil
```

### CRITICAL: DBInstance.Endpoint is a Nested Struct

Unlike DBCluster where `Endpoint` is a `*string`, DBInstance's `Endpoint` is a **nested struct** (`*neptune.Endpoint`) containing `Address *string` and `Port *int64`. Requires double nil-checking:

```go
endpointAddress := ""
endpointPort := ""
if instance.Endpoint != nil {
    if instance.Endpoint.Address != nil {
        endpointAddress = *instance.Endpoint.Address
    }
    if instance.Endpoint.Port != nil {
        endpointPort = fmt.Sprintf("%d", *instance.Endpoint.Port)
    }
}
```

### Neptune SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/neptune`

**IMPORTANT:** The Neptune SDK is part of aws-sdk-go v1.44.266 but is **NOT currently vendored**. Run `go mod vendor` before building to populate `vendor/github.com/aws/aws-sdk-go/service/neptune/`.

**API Methods:**

1. **DescribeDBClusters (Paginated with Marker, regional):**
   - `svc.DescribeDBClustersWithContext(ctx, &neptune.DescribeDBClustersInput{Marker: marker})` -> `*neptune.DescribeDBClustersOutput`
   - `.DBClusters` -> `[]*neptune.DBCluster`
   - Pagination: `Marker *string` in both input and output
   - Each `DBCluster` has:
     - `DBClusterIdentifier *string` — cluster name/identifier
     - `DBClusterArn *string`
     - `Status *string` — "available", "creating", "deleting", etc.
     - `Engine *string` — "neptune"
     - `EngineVersion *string` — e.g., "1.2.0.2"
     - `Endpoint *string` — writer endpoint (connection string)
     - `ReaderEndpoint *string` — reader endpoint
     - `Port *int64` — database port (default 8182)
     - `MultiAZ *bool` — multi-AZ deployment
     - `StorageEncrypted *bool` — encryption at rest
     - `KmsKeyId *string` — KMS key ARN (security-relevant: reveals key used)
     - `DeletionProtection *bool` — deletion protection enabled
     - `IAMDatabaseAuthenticationEnabled *bool` — IAM auth (security-relevant: password vs IAM)
     - `DBClusterParameterGroup *string` — parameter group name
     - `ClusterCreateTime *time.Time` — creation timestamp

2. **DescribeDBInstances (Paginated with Marker, regional):**
   - `svc.DescribeDBInstancesWithContext(ctx, &neptune.DescribeDBInstancesInput{Marker: marker})` -> `*neptune.DescribeDBInstancesOutput`
   - `.DBInstances` -> `[]*neptune.DBInstance`
   - Pagination: `Marker *string` in both input and output
   - Each `DBInstance` has:
     - `DBInstanceIdentifier *string` — instance name
     - `DBInstanceArn *string`
     - `DBInstanceClass *string` — e.g., "db.r5.large"
     - `Engine *string` — "neptune"
     - `EngineVersion *string`
     - `DBInstanceStatus *string` — "available", "creating", etc.
     - `Endpoint *neptune.Endpoint` — **NESTED STRUCT** (not a string!)
       - `.Address *string` — connection hostname
       - `.Port *int64` — connection port
     - `DBClusterIdentifier *string` — parent cluster
     - `AvailabilityZone *string` — AZ placement
     - `PubliclyAccessible *bool` — **SECURITY-CRITICAL: internet-facing?**
     - `StorageEncrypted *bool` — encryption at rest
     - `AutoMinorVersionUpgrade *bool` — auto-patching

3. **DescribeDBClusterParameterGroups (Paginated with Marker, regional):**
   - `svc.DescribeDBClusterParameterGroupsWithContext(ctx, &neptune.DescribeDBClusterParameterGroupsInput{Marker: marker})` -> `*neptune.DescribeDBClusterParameterGroupsOutput`
   - `.DBClusterParameterGroups` -> `[]*neptune.DBClusterParameterGroup`
   - Pagination: `Marker *string` in both input and output
   - Each `DBClusterParameterGroup` has:
     - `DBClusterParameterGroupName *string` — group name
     - `DBClusterParameterGroupArn *string`
     - `Description *string` — group description
     - `DBParameterGroupFamily *string` — engine family (e.g., "neptune1")

### No MaxRecords Needed

Unlike MediaConvert which requires `MaxResults: aws.Int64(20)`, Neptune's Describe APIs have a default page size of 100 which is sufficient for enumeration. No need to set `MaxRecords` in input.

### Nil-Safe Field Extraction Helpers

```go
func extractCluster(cluster *neptune.DBCluster, region string) npCluster {
    identifier := ""
    if cluster.DBClusterIdentifier != nil {
        identifier = *cluster.DBClusterIdentifier
    }
    arn := ""
    if cluster.DBClusterArn != nil {
        arn = *cluster.DBClusterArn
    }
    status := ""
    if cluster.Status != nil {
        status = *cluster.Status
    }
    engine := ""
    if cluster.Engine != nil {
        engine = *cluster.Engine
    }
    engineVersion := ""
    if cluster.EngineVersion != nil {
        engineVersion = *cluster.EngineVersion
    }
    endpoint := ""
    if cluster.Endpoint != nil {
        endpoint = *cluster.Endpoint
    }
    readerEndpoint := ""
    if cluster.ReaderEndpoint != nil {
        readerEndpoint = *cluster.ReaderEndpoint
    }
    port := ""
    if cluster.Port != nil {
        port = fmt.Sprintf("%d", *cluster.Port)
    }
    multiAZ := ""
    if cluster.MultiAZ != nil {
        multiAZ = fmt.Sprintf("%t", *cluster.MultiAZ)
    }
    storageEncrypted := ""
    if cluster.StorageEncrypted != nil {
        storageEncrypted = fmt.Sprintf("%t", *cluster.StorageEncrypted)
    }
    kmsKeyId := ""
    if cluster.KmsKeyId != nil {
        kmsKeyId = *cluster.KmsKeyId
    }
    deletionProtection := ""
    if cluster.DeletionProtection != nil {
        deletionProtection = fmt.Sprintf("%t", *cluster.DeletionProtection)
    }
    iamAuth := ""
    if cluster.IAMDatabaseAuthenticationEnabled != nil {
        iamAuth = fmt.Sprintf("%t", *cluster.IAMDatabaseAuthenticationEnabled)
    }
    parameterGroup := ""
    if cluster.DBClusterParameterGroup != nil {
        parameterGroup = *cluster.DBClusterParameterGroup
    }
    createTime := ""
    if cluster.ClusterCreateTime != nil {
        createTime = cluster.ClusterCreateTime.Format(time.RFC3339)
    }
    return npCluster{
        DBClusterIdentifier:              identifier,
        DBClusterArn:                     arn,
        Status:                           status,
        Engine:                           engine,
        EngineVersion:                    engineVersion,
        Endpoint:                         endpoint,
        ReaderEndpoint:                   readerEndpoint,
        Port:                             port,
        MultiAZ:                          multiAZ,
        StorageEncrypted:                 storageEncrypted,
        KmsKeyId:                         kmsKeyId,
        DeletionProtection:               deletionProtection,
        IAMDatabaseAuthenticationEnabled: iamAuth,
        DBClusterParameterGroup:          parameterGroup,
        ClusterCreateTime:                createTime,
        Region:                           region,
    }
}

func extractInstance(instance *neptune.DBInstance, region string) npInstance {
    identifier := ""
    if instance.DBInstanceIdentifier != nil {
        identifier = *instance.DBInstanceIdentifier
    }
    arn := ""
    if instance.DBInstanceArn != nil {
        arn = *instance.DBInstanceArn
    }
    instanceClass := ""
    if instance.DBInstanceClass != nil {
        instanceClass = *instance.DBInstanceClass
    }
    engine := ""
    if instance.Engine != nil {
        engine = *instance.Engine
    }
    engineVersion := ""
    if instance.EngineVersion != nil {
        engineVersion = *instance.EngineVersion
    }
    status := ""
    if instance.DBInstanceStatus != nil {
        status = *instance.DBInstanceStatus
    }
    endpointAddress := ""
    endpointPort := ""
    if instance.Endpoint != nil {
        if instance.Endpoint.Address != nil {
            endpointAddress = *instance.Endpoint.Address
        }
        if instance.Endpoint.Port != nil {
            endpointPort = fmt.Sprintf("%d", *instance.Endpoint.Port)
        }
    }
    clusterIdentifier := ""
    if instance.DBClusterIdentifier != nil {
        clusterIdentifier = *instance.DBClusterIdentifier
    }
    az := ""
    if instance.AvailabilityZone != nil {
        az = *instance.AvailabilityZone
    }
    publiclyAccessible := ""
    if instance.PubliclyAccessible != nil {
        publiclyAccessible = fmt.Sprintf("%t", *instance.PubliclyAccessible)
    }
    storageEncrypted := ""
    if instance.StorageEncrypted != nil {
        storageEncrypted = fmt.Sprintf("%t", *instance.StorageEncrypted)
    }
    autoUpgrade := ""
    if instance.AutoMinorVersionUpgrade != nil {
        autoUpgrade = fmt.Sprintf("%t", *instance.AutoMinorVersionUpgrade)
    }
    return npInstance{
        DBInstanceIdentifier:    identifier,
        DBInstanceArn:           arn,
        DBInstanceClass:         instanceClass,
        Engine:                  engine,
        EngineVersion:           engineVersion,
        DBInstanceStatus:        status,
        EndpointAddress:         endpointAddress,
        EndpointPort:            endpointPort,
        DBClusterIdentifier:     clusterIdentifier,
        AvailabilityZone:        az,
        PubliclyAccessible:      publiclyAccessible,
        StorageEncrypted:        storageEncrypted,
        AutoMinorVersionUpgrade: autoUpgrade,
        Region:                  region,
    }
}

func extractParameterGroup(pg *neptune.DBClusterParameterGroup, region string) npParameterGroup {
    name := ""
    if pg.DBClusterParameterGroupName != nil {
        name = *pg.DBClusterParameterGroupName
    }
    arn := ""
    if pg.DBClusterParameterGroupArn != nil {
        arn = *pg.DBClusterParameterGroupArn
    }
    description := ""
    if pg.Description != nil {
        description = *pg.Description
    }
    family := ""
    if pg.DBParameterGroupFamily != nil {
        family = *pg.DBParameterGroupFamily
    }
    return npParameterGroup{
        DBClusterParameterGroupName: name,
        DBClusterParameterGroupArn:  arn,
        Description:                 description,
        DBParameterGroupFamily:      family,
        Region:                      region,
    }
}
```

### Local Struct Definitions

```go
type npCluster struct {
    DBClusterIdentifier              string
    DBClusterArn                     string
    Status                           string
    Engine                           string
    EngineVersion                    string
    Endpoint                         string
    ReaderEndpoint                   string
    Port                             string
    MultiAZ                          string
    StorageEncrypted                 string
    KmsKeyId                         string
    DeletionProtection               string
    IAMDatabaseAuthenticationEnabled string
    DBClusterParameterGroup          string
    ClusterCreateTime                string
    Region                           string
}

type npInstance struct {
    DBInstanceIdentifier    string
    DBInstanceArn           string
    DBInstanceClass         string
    Engine                  string
    EngineVersion           string
    DBInstanceStatus        string
    EndpointAddress         string
    EndpointPort            string
    DBClusterIdentifier     string
    AvailabilityZone        string
    PubliclyAccessible      string
    StorageEncrypted        string
    AutoMinorVersionUpgrade string
    Region                  string
}

type npParameterGroup struct {
    DBClusterParameterGroupName string
    DBClusterParameterGroupArn  string
    Description                 string
    DBParameterGroupFamily      string
    Region                      string
}
```

### Variable & Naming Conventions

- **Package:** `neptune` (directory: `cmd/awtest/services/neptune/`)
- **Exported variable:** `NeptuneCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"neptune:DescribeDBClusters"`, `"neptune:DescribeDBInstances"`, `"neptune:DescribeDBClusterParameterGroups"`
- **ScanResult.ServiceName:** `"Neptune"` (title case)
- **ScanResult.ResourceType:** `"db-cluster"`, `"db-instance"`, `"db-cluster-parameter-group"` (lowercase, hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **Local struct prefix:** `np` (abbreviation for neptune, following `mc`/`cd`/`dc` pattern)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/neptune"` (same name as local package — handled same as mediaconvert/kinesis/emr pattern)

### Registration Order in services.go

Insert alphabetically — `neptune` comes after `mediaconvert`, before `opensearch`:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/mediaconvert"
"github.com/MillerMedia/awtest/cmd/awtest/services/neptune"       // NEW — after mediaconvert, before opensearch
"github.com/MillerMedia/awtest/cmd/awtest/services/opensearch"

// In AllServices():
allServices = append(allServices, mediaconvert.MediaConvertCalls...)
allServices = append(allServices, neptune.NeptuneCalls...)           // NEW — after mediaconvert, before opensearch
allServices = append(allServices, opensearch.OpenSearchCalls...)
```

### Testing Pattern

Follow the MediaConvert/Kinesis test pattern — test Process() functions only with pre-built mock data:

```go
func TestDescribeDBClustersProcess(t *testing.T) {
    process := NeptuneCalls[0].Process
    // Table-driven tests: valid clusters (identifier, ARN, status, engine, version, endpoint, reader endpoint, port, multi-AZ, encrypted, KMS key, deletion protection, IAM auth, parameter group, create time), empty, errors, nil fields, type assertion failure
}

func TestDescribeDBInstancesProcess(t *testing.T) {
    process := NeptuneCalls[1].Process
    // Table-driven tests: valid instances (identifier, ARN, class, engine, version, status, endpoint address/port, cluster, AZ, publicly accessible, encrypted, auto upgrade), empty, errors, nil fields, type assertion failure
}

func TestDescribeDBClusterParameterGroupsProcess(t *testing.T) {
    process := NeptuneCalls[2].Process
    // Table-driven tests: valid parameter groups (name, ARN, description, family), empty, errors, nil fields, type assertion failure
}
```

Include extract helper tests with AWS SDK types:
```go
func TestExtractCluster(t *testing.T) { ... }
func TestExtractInstance(t *testing.T) { ... }
func TestExtractParameterGroup(t *testing.T) { ... }
```

**IMPORTANT:** For `TestExtractInstance`, the test must construct a `*neptune.Endpoint` nested struct for the `Endpoint` field, not a string. Test both cases: populated Endpoint with Address/Port, and nil Endpoint.

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `neptune/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `neptune.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** use NextToken — Neptune uses **Marker-based pagination** (not NextToken)
- **DO NOT** set MaxRecords in input — Neptune defaults to 100 per page, sufficient for enumeration
- **DO NOT** treat `DBInstance.Endpoint` as a string — it is a `*neptune.Endpoint` nested struct with `Address *string` and `Port *int64`
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** filter by engine type — the Neptune API endpoint only returns Neptune resources, unlike the RDS API which returns all database engines

### Key Differences from Previous Stories (9.5 MediaConvert, 9.4 Kinesis)

1. **Marker-based pagination:** Neptune uses `Marker` pagination (like RDS), NOT `NextToken` (like MediaConvert/Kinesis). The pagination field in input and output is called `Marker`, not `NextToken`.
2. **No MaxRecords/MaxResults needed:** Neptune defaults to 100 results per page. MediaConvert required `MaxResults: aws.Int64(20)`.
3. **Nested Endpoint struct:** `DBInstance.Endpoint` is a `*neptune.Endpoint` struct (with Address/Port), not a flat string. `DBCluster.Endpoint` IS a flat `*string`. This asymmetry is a common pitfall.
4. **Boolean fields:** Neptune has several boolean fields (`MultiAZ`, `StorageEncrypted`, `DeletionProtection`, `IAMDatabaseAuthenticationEnabled`, `PubliclyAccessible`, `AutoMinorVersionUpgrade`) — format with `fmt.Sprintf("%t", *field)`. MediaConvert had no boolean fields.
5. **Int64 fields:** `Port` on DBCluster is `*int64`. Same formatting as MediaConvert's `SubmittedJobsCount`/`ProgressingJobsCount`.
6. **Security-relevant fields:** Neptune exposes `PubliclyAccessible` (internet-facing?), `IAMDatabaseAuthenticationEnabled` (auth method), `KmsKeyId` (encryption key), and endpoint connection strings — all high-value for pentesters mapping attack surface.
7. **Describe vs List:** Neptune uses `Describe*` operations (like RDS), not `List*` operations (like MediaConvert/Kinesis). Describe operations return full resource details directly, no need for follow-up detail calls.
8. **All regional:** All Neptune resources are regional — iterate `types.Regions` for all 3 API calls.

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/neptune/
+-- calls.go            # Neptune service implementation (3 AWSService entries)
+-- calls_test.go       # Process() tests + extract helper tests for all 3 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                      # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                     # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/mediaconvert/calls.go      # Reference implementation (config override, extract helpers)
cmd/awtest/services/mediaconvert/calls_test.go # Reference test pattern (extract helper tests)
cmd/awtest/services/rds/calls.go               # Reference for Marker pagination (but uses old session mutation pattern — DO NOT copy session pattern)
go.mod                                         # AWS SDK already includes neptune package (needs vendoring)
```

**Vendor directory to POPULATE:**
```
vendor/github.com/aws/aws-sdk-go/service/neptune/  # Run go mod vendor to populate
```

### Previous Story Intelligence

**From Story 9.5 (MediaConvert — most recent completed story):**
- Config override pattern confirmed working and mandatory
- Local structs + extract helpers pattern established (mc prefix -> np prefix)
- Per-region errors: `break` pagination loop, don't abort entire scan
- 24 tests across 6 test functions (3 Process + 3 extract helpers)
- SDK field name quirks possible (MediaConvert had `Type_` which turned out to be `Type` in actual vendored SDK) — verify field names after vendoring
- Error result pattern: `return []types.ScanResult{{ServiceName: "Neptune", MethodName: "neptune:DescribeDBClusters", Error: err, Timestamp: time.Now()}}`
- Process via `NeptuneCalls[N].Process` in tests

**From Story 9.4 (Kinesis):**
- Simple paginated list pattern — similar to Neptune Describe operations
- Extract helper functions for nil-safe extraction — directly applicable
- Time formatting with `time.RFC3339` — same pattern needed for `ClusterCreateTime`

**From Story 9.1 (CodeDeploy):**
- Config override pattern first established
- Table-driven tests with `t.Run` subtests

**From Code Review Findings (Stories 7.1, 7.2):**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [HIGH] Always add pagination from the start
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [LOW] Tests should cover nil fields comprehensively
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `d80d7c1` — Add MediaConvert enumeration with 3 API calls (Story 9.5)
- `f528ebb` — Add EMR and Kinesis enumeration with 3 API calls each (Stories 9.3, 9.4)
- `b7a8967` — Add Direct Connect enumeration with 3 API calls (Story 9.2)
- `7b02834` — Add CodeDeploy enumeration with 3 API calls (Story 9.1)
- Expected commit message: `"Add Neptune enumeration with 3 API calls (Story 9.6)"`

### FRs Covered

- **FR113:** System enumerates Neptune DB clusters, instances, and parameter groups

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### References

- [Source: epics-phase2.md#Story 4.6: Neptune Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR113] — Neptune enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/mediaconvert/calls.go] — Reference implementation (config override, extract helpers, pagination)
- [Source: cmd/awtest/services/mediaconvert/calls_test.go] — Reference test pattern (extract helper tests)
- [Source: cmd/awtest/services/rds/calls.go] — Reference for Marker-based pagination pattern (DO NOT copy session mutation)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (neptune goes after mediaconvert, before opensearch)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes neptune package, needs vendoring)
- [Source: 9-5-mediaconvert-enumeration.md] — Previous story (config override, extract helpers, pagination)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered.

### Completion Notes List

- Implemented Neptune service enumeration with 3 API calls: DescribeDBClusters, DescribeDBInstances, DescribeDBClusterParameterGroups
- Used config override pattern for region iteration (concurrency-safe)
- Used Marker-based pagination (Neptune/RDS pattern, not NextToken)
- Handled nested Endpoint struct on DBInstance with double nil-checking
- All boolean fields formatted with `fmt.Sprintf("%t", *field)`
- ClusterCreateTime formatted with `time.RFC3339`
- Registered in services.go alphabetically (after mediaconvert, before opensearch)
- 24 tests across 6 test functions: 3 Process tests + 3 extract helper tests
- TestExtractInstance includes cases for nil Endpoint struct and Endpoint with nil Address/Port
- No sync primitives used — concurrency-unaware per NFR57
- Full test suite passes with no regressions, race detector clean
- Vendored Neptune SDK package from aws-sdk-go v1.44.266

### File List

- cmd/awtest/services/neptune/calls.go (NEW)
- cmd/awtest/services/neptune/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED)
- vendor/github.com/aws/aws-sdk-go/service/neptune/ (VENDORED)

### Change Log

- 2026-03-13: Implemented Neptune enumeration with 3 API calls (DescribeDBClusters, DescribeDBInstances, DescribeDBClusterParameterGroups). 24 tests added. Service registered in AllServices(). Status: review.
