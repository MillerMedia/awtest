# Story 9.2: Direct Connect Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate Direct Connect connections, virtual interfaces, and gateways,
So that I can discover dedicated network links between AWS and on-premises infrastructure.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/directconnect/` directory with `calls.go` implementing Direct Connect service enumeration with 3 AWSService entries.

2. **AC2:** Implement `directconnect:DescribeConnections` API call — iterates all regions in `types.Regions`, creates Direct Connect client per region using config override pattern (`directconnect.New(sess, &aws.Config{Region: aws.String(region)})`), calls `DescribeConnectionsWithContext` (non-paginated, returns all connections in the region). Each connection listed with ConnectionId, ConnectionName, ConnectionState, Bandwidth, Location, OwnerAccount, PartnerName, and Region.

3. **AC3:** Implement `directconnect:DescribeVirtualInterfaces` API call — iterates all regions in `types.Regions`, creates Direct Connect client per region using config override, calls `DescribeVirtualInterfacesWithContext` (non-paginated, returns all virtual interfaces in the region). Each virtual interface listed with VirtualInterfaceId, VirtualInterfaceName, VirtualInterfaceState, VirtualInterfaceType, ConnectionId, Vlan, Asn, AmazonAddress, CustomerAddress, and Region.

4. **AC4:** Implement `directconnect:DescribeDirectConnectGateways` API call — creates Direct Connect client for a single region (gateways are global resources), calls `DescribeDirectConnectGatewaysWithContext` with NextToken pagination. Each gateway listed with DirectConnectGatewayId, DirectConnectGatewayName, DirectConnectGatewayState, AmazonSideAsn, and OwnerAccount.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without Direct Connect access, Direct Connect is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register Direct Connect service in `services/services.go` `AllServices()` function in alphabetical order (after `config`, before `dynamodb`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 3 API calls.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/directconnect/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `directconnect:DescribeConnections` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/directconnect/`
  - [x] Create `calls.go` with `package directconnect`
  - [x] Define `var DirectConnectCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"directconnect:DescribeConnections"`
  - [x] Call: iterate `types.Regions`, create `directconnect.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `DescribeConnectionsWithContext` (non-paginated — returns all connections in one call). Define local struct `dcConnection` with fields: ConnectionId, ConnectionName, ConnectionState, Bandwidth, Location, OwnerAccount, PartnerName, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Implement `extractConnection` helper function — takes `*directconnect.Connection` and `region` string, returns `dcConnection` with nil-safe pointer dereferencing.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]dcConnection`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"DirectConnect"`, ResourceType=`"connection"`, ResourceName=connectionName
  - [x] `utils.PrintResult` format: `"Direct Connect Connection: %s (State: %s, Bandwidth: %s, Location: %s, Region: %s)"` with `utils.ColorizeItem(connectionName)`

- [x] Task 2: Implement `directconnect:DescribeVirtualInterfaces` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"directconnect:DescribeVirtualInterfaces"`
  - [x] Call: iterate regions -> create Direct Connect client with config override -> call `DescribeVirtualInterfacesWithContext` (non-paginated — returns all VIs in one call). Define local struct `dcVirtualInterface` with fields: VirtualInterfaceId, VirtualInterfaceName, VirtualInterfaceState, VirtualInterfaceType, ConnectionId, Vlan, Asn, AmazonAddress, CustomerAddress, Region. Per-region errors: `break` to next region.
  - [x] Implement `extractVirtualInterface` helper function — takes `*directconnect.VirtualInterface` and `region` string, returns `dcVirtualInterface` with nil-safe pointer dereferencing. Note: Vlan and Asn are `*int64` — convert to string with `fmt.Sprintf("%d", *ptr)`.
  - [x] Process: type-assert `[]dcVirtualInterface`, build `ScanResult` with ServiceName=`"DirectConnect"`, ResourceType=`"virtual-interface"`, ResourceName=viName
  - [x] `utils.PrintResult` format: `"Direct Connect Virtual Interface: %s (Type: %s, State: %s, VLAN: %s, Connection: %s, Region: %s)"` with `utils.ColorizeItem(viName)`

- [x] Task 3: Implement `directconnect:DescribeDirectConnectGateways` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"directconnect:DescribeDirectConnectGateways"`
  - [x] Call: create Direct Connect client for single region `types.Regions[0]` (gateways are **global** resources — do NOT iterate all regions to avoid duplicates). Call `DescribeDirectConnectGatewaysWithContext` with NextToken pagination loop. Define local struct `dcGateway` with fields: DirectConnectGatewayId, DirectConnectGatewayName, DirectConnectGatewayState, AmazonSideAsn, OwnerAccount. Note: AmazonSideAsn is `*int64` — convert to string.
  - [x] Implement `extractGateway` helper function — takes `*directconnect.Gateway` (SDK type), returns `dcGateway` with nil-safe pointer dereferencing.
  - [x] Process: type-assert `[]dcGateway`, build `ScanResult` with ServiceName=`"DirectConnect"`, ResourceType=`"gateway"`, ResourceName=gatewayName
  - [x] `utils.PrintResult` format: `"Direct Connect Gateway: %s (State: %s, ASN: %s, Owner: %s)"` with `utils.ColorizeItem(gatewayName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/directconnect"` to `services/services.go` (alphabetical in imports: after `config`, before `dynamodb`)
  - [x] Add `allServices = append(allServices, directconnect.DirectConnectCalls...)` after `config.ConfigCalls...` and before `dynamodb.DynamoDBCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/directconnect/calls_test.go`
  - [x] Test `DescribeConnections` Process: valid connections with details (ID, name, state, bandwidth, location, owner, partner), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `DescribeVirtualInterfaces` Process: valid VIs with details (ID, name, state, type, connection ID, VLAN, ASN, addresses), empty results, error handling, nil fields, type assertion failure
  - [x] Test `DescribeDirectConnectGateways` Process: valid gateways with details (ID, name, state, ASN, owner), empty results, error handling, nil fields, type assertion failure
  - [x] Test extract helpers: `TestExtractConnection`, `TestExtractVirtualInterface`, `TestExtractGateway` with AWS SDK types (both populated and nil fields)
  - [x] Use table-driven tests with `t.Run` subtests following CodeDeploy/Macie test pattern
  - [x] Access Process via `DirectConnectCalls[0].Process`, `DirectConnectCalls[1].Process`, `DirectConnectCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/directconnect/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := directconnect.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := directconnect.New(sess)
}
```

### Direct Connect is a REGIONAL Service with a GLOBAL Exception

AWS Direct Connect connections and virtual interfaces are **regional** — they exist per-region. Iterate `types.Regions` for Calls 1 and 2.

**EXCEPTION:** Direct Connect Gateways are **global** resources. Call 3 (`DescribeDirectConnectGateways`) must be called from a **single region only** (use `types.Regions[0]`, i.e. `"us-east-1"`) to avoid returning duplicate gateway entries. Do NOT iterate all regions for gateways.

### Direct Connect SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/directconnect`

**IMPORTANT:** The Go package name is `directconnect`. The local package name is also `directconnect`, same pattern as `codedeploy`/`codecommit`/`codebuild` where the local package name matches the AWS SDK package name. Within `calls.go`, `directconnect.New()` and `directconnect.DescribeConnectionsInput{}` refer to the **AWS SDK package**, while local types (structs, variables) are referenced directly without package prefix.

**API Methods:**

1. **DescribeConnections (Non-paginated, returns all):**
   - `svc.DescribeConnectionsWithContext(ctx, &directconnect.DescribeConnectionsInput{})` -> `*directconnect.DescribeConnectionsOutput`
   - `.Connections` -> `[]*directconnect.Connection`
   - **No pagination** — returns all connections in the region in one call
   - Each `Connection` has:
     - `ConnectionId *string` (e.g., "dxcon-abc12345")
     - `ConnectionName *string`
     - `ConnectionState *string` ("ordering", "requested", "pending", "available", "down", "deleting", "deleted", "rejected", "unknown")
     - `Bandwidth *string` (e.g., "1Gbps", "10Gbps")
     - `Location *string` (physical facility code, e.g., "EqDC2")
     - `OwnerAccount *string` (AWS account ID)
     - `PartnerName *string` (e.g., "Equinix")
     - `Region *string` (AWS region)
     - `HasLogicalRedundancy *string` ("yes", "no", "unknown")
     - `JumboFrameCapable *bool`
     - `MacSecCapable *bool`
     - `ProviderName *string`
     - `Vlan *int64`

2. **DescribeVirtualInterfaces (Non-paginated, returns all):**
   - `svc.DescribeVirtualInterfacesWithContext(ctx, &directconnect.DescribeVirtualInterfacesInput{})` -> `*directconnect.DescribeVirtualInterfacesOutput`
   - `.VirtualInterfaces` -> `[]*directconnect.VirtualInterface`
   - **No pagination** — returns all virtual interfaces in the region in one call
   - Each `VirtualInterface` has:
     - `VirtualInterfaceId *string` (e.g., "dxvif-abc12345")
     - `VirtualInterfaceName *string`
     - `VirtualInterfaceState *string` ("confirming", "verifying", "pending", "available", "down", "deleting", "deleted", "rejected", "unknown")
     - `VirtualInterfaceType *string` ("private", "public", "transit")
     - `ConnectionId *string` (associated DX connection)
     - `Vlan *int64` (VLAN ID)
     - `Asn *int64` (customer BGP ASN)
     - `AmazonSideAsn *int64` (Amazon BGP ASN)
     - `AmazonAddress *string` (IP address on Amazon side)
     - `CustomerAddress *string` (IP address on customer side)
     - `OwnerAccount *string`
     - `Region *string`
     - `DirectConnectGatewayId *string`
     - `BgpPeers` -> `[]*directconnect.BGPPeer` (optional, nested)

3. **DescribeDirectConnectGateways (Paginated, GLOBAL):**
   - `svc.DescribeDirectConnectGatewaysWithContext(ctx, &directconnect.DescribeDirectConnectGatewaysInput{NextToken: nextToken})` -> `*directconnect.DescribeDirectConnectGatewaysOutput`
   - `.DirectConnectGateways` -> `[]*directconnect.DirectConnectGateway`
   - Pagination: `NextToken *string` in both input and output
   - Has optional `MaxResults *int64` — do NOT set (use default page size)
   - Each `DirectConnectGateway` has:
     - `DirectConnectGatewayId *string`
     - `DirectConnectGatewayName *string`
     - `DirectConnectGatewayState *string` ("pending", "available", "deleting", "deleted")
     - `AmazonSideAsn *int64` (Amazon-side BGP ASN)
     - `OwnerAccount *string`
     - `StateChangeError *string`

**No new dependencies needed** — Direct Connect is part of `aws-sdk-go v1.44.266` already in go.mod.

### Non-Paginated API Pattern (DescribeConnections, DescribeVirtualInterfaces)

Calls 1 and 2 use **non-paginated** APIs — they return all results in a single call. This is simpler than the paginated pattern:

```go
var allConnections []dcConnection
var lastErr error

for _, region := range types.Regions {
    svc := directconnect.New(sess, &aws.Config{Region: aws.String(region)})
    output, err := svc.DescribeConnectionsWithContext(ctx, &directconnect.DescribeConnectionsInput{})
    if err != nil {
        lastErr = err
        utils.HandleAWSError(false, "directconnect:DescribeConnections", err)
        continue  // continue to next region (not break — non-paginated, single call per region)
    }
    for _, conn := range output.Connections {
        if conn != nil {
            allConnections = append(allConnections, extractConnection(conn, region))
        }
    }
}
```

**Note:** Use `continue` (not `break`) on per-region errors since there's no pagination loop to break from. Each region is a single API call.

### Paginated Global API Pattern (DescribeDirectConnectGateways)

Call 3 uses pagination but is called from a single region:

```go
var allGateways []dcGateway
var lastErr error

svc := directconnect.New(sess, &aws.Config{Region: aws.String(types.Regions[0])})
var nextToken *string
for {
    input := &directconnect.DescribeDirectConnectGatewaysInput{}
    if nextToken != nil {
        input.NextToken = nextToken
    }
    output, err := svc.DescribeDirectConnectGatewaysWithContext(ctx, input)
    if err != nil {
        lastErr = err
        utils.HandleAWSError(false, "directconnect:DescribeDirectConnectGateways", err)
        break
    }
    for _, gw := range output.DirectConnectGateways {
        if gw != nil {
            allGateways = append(allGateways, extractGateway(gw))
        }
    }
    if output.NextToken == nil {
        break
    }
    nextToken = output.NextToken
}
```

### Nil-Safe Field Extraction Helpers

```go
func extractConnection(conn *directconnect.Connection, region string) dcConnection {
    connId := ""
    if conn.ConnectionId != nil {
        connId = *conn.ConnectionId
    }
    name := ""
    if conn.ConnectionName != nil {
        name = *conn.ConnectionName
    }
    state := ""
    if conn.ConnectionState != nil {
        state = *conn.ConnectionState
    }
    bandwidth := ""
    if conn.Bandwidth != nil {
        bandwidth = *conn.Bandwidth
    }
    location := ""
    if conn.Location != nil {
        location = *conn.Location
    }
    ownerAccount := ""
    if conn.OwnerAccount != nil {
        ownerAccount = *conn.OwnerAccount
    }
    partnerName := ""
    if conn.PartnerName != nil {
        partnerName = *conn.PartnerName
    }
    return dcConnection{
        ConnectionId:   connId,
        ConnectionName: name,
        ConnectionState: state,
        Bandwidth:      bandwidth,
        Location:       location,
        OwnerAccount:   ownerAccount,
        PartnerName:    partnerName,
        Region:         region,
    }
}

func extractVirtualInterface(vi *directconnect.VirtualInterface, region string) dcVirtualInterface {
    viId := ""
    if vi.VirtualInterfaceId != nil {
        viId = *vi.VirtualInterfaceId
    }
    name := ""
    if vi.VirtualInterfaceName != nil {
        name = *vi.VirtualInterfaceName
    }
    state := ""
    if vi.VirtualInterfaceState != nil {
        state = *vi.VirtualInterfaceState
    }
    viType := ""
    if vi.VirtualInterfaceType != nil {
        viType = *vi.VirtualInterfaceType
    }
    connId := ""
    if vi.ConnectionId != nil {
        connId = *vi.ConnectionId
    }
    vlan := ""
    if vi.Vlan != nil {
        vlan = fmt.Sprintf("%d", *vi.Vlan)
    }
    asn := ""
    if vi.Asn != nil {
        asn = fmt.Sprintf("%d", *vi.Asn)
    }
    amazonAddr := ""
    if vi.AmazonAddress != nil {
        amazonAddr = *vi.AmazonAddress
    }
    customerAddr := ""
    if vi.CustomerAddress != nil {
        customerAddr = *vi.CustomerAddress
    }
    return dcVirtualInterface{
        VirtualInterfaceId:    viId,
        VirtualInterfaceName:  name,
        VirtualInterfaceState: state,
        VirtualInterfaceType:  viType,
        ConnectionId:          connId,
        Vlan:                  vlan,
        Asn:                   asn,
        AmazonAddress:         amazonAddr,
        CustomerAddress:       customerAddr,
        Region:                region,
    }
}

func extractGateway(gw *directconnect.DirectConnectGateway) dcGateway {
    gwId := ""
    if gw.DirectConnectGatewayId != nil {
        gwId = *gw.DirectConnectGatewayId
    }
    name := ""
    if gw.DirectConnectGatewayName != nil {
        name = *gw.DirectConnectGatewayName
    }
    state := ""
    if gw.DirectConnectGatewayState != nil {
        state = *gw.DirectConnectGatewayState
    }
    asn := ""
    if gw.AmazonSideAsn != nil {
        asn = fmt.Sprintf("%d", *gw.AmazonSideAsn)
    }
    owner := ""
    if gw.OwnerAccount != nil {
        owner = *gw.OwnerAccount
    }
    return dcGateway{
        DirectConnectGatewayId:    gwId,
        DirectConnectGatewayName:  name,
        DirectConnectGatewayState: state,
        AmazonSideAsn:            asn,
        OwnerAccount:             owner,
    }
}
```

### Local Struct Definitions

```go
type dcConnection struct {
    ConnectionId    string
    ConnectionName  string
    ConnectionState string
    Bandwidth       string
    Location        string
    OwnerAccount    string
    PartnerName     string
    Region          string
}

type dcVirtualInterface struct {
    VirtualInterfaceId    string
    VirtualInterfaceName  string
    VirtualInterfaceState string
    VirtualInterfaceType  string
    ConnectionId          string
    Vlan                  string
    Asn                   string
    AmazonAddress         string
    CustomerAddress       string
    Region                string
}

type dcGateway struct {
    DirectConnectGatewayId    string
    DirectConnectGatewayName  string
    DirectConnectGatewayState string
    AmazonSideAsn             string
    OwnerAccount              string
}
```

### Variable & Naming Conventions

- **Package:** `directconnect` (directory: `cmd/awtest/services/directconnect/`)
- **Exported variable:** `DirectConnectCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"directconnect:DescribeConnections"`, `"directconnect:DescribeVirtualInterfaces"`, `"directconnect:DescribeDirectConnectGateways"`
- **ScanResult.ServiceName:** `"DirectConnect"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"connection"`, `"virtual-interface"`, `"gateway"` (lowercase hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **Local struct prefix:** `dc` (for DirectConnect, following `cb` for CodeBuild, `cd` for CodeDeploy pattern)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/directconnect"` (same name as local package — handled same as codedeploy/codecommit/codebuild pattern)

### Registration Order in services.go

Insert alphabetically — `directconnect` comes after `config`, before `dynamodb`:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/config"
"github.com/MillerMedia/awtest/cmd/awtest/services/directconnect"    // NEW — after config, before dynamodb
"github.com/MillerMedia/awtest/cmd/awtest/services/dynamodb"

// In AllServices():
allServices = append(allServices, config.ConfigCalls...)
allServices = append(allServices, directconnect.DirectConnectCalls...)  // NEW — after config, before dynamodb
allServices = append(allServices, dynamodb.DynamoDBCalls...)
```

### Testing Pattern

Follow the CodeDeploy/Macie test pattern — test Process() functions only with pre-built mock data:

```go
func TestDescribeConnectionsProcess(t *testing.T) {
    process := DirectConnectCalls[0].Process
    // Table-driven tests: valid connections (ID, name, state, bandwidth, location, owner, partner), empty, errors, nil fields, type assertion failure
}

func TestDescribeVirtualInterfacesProcess(t *testing.T) {
    process := DirectConnectCalls[1].Process
    // Table-driven tests: valid VIs (ID, name, state, type, connection ID, VLAN, ASN, addresses), empty, errors, nil fields, type assertion failure
}

func TestDescribeDirectConnectGatewaysProcess(t *testing.T) {
    process := DirectConnectCalls[2].Process
    // Table-driven tests: valid gateways (ID, name, state, ASN, owner), empty, errors, nil fields, type assertion failure
}
```

Include extract helper tests with AWS SDK types:
```go
func TestExtractConnection(t *testing.T) { ... }
func TestExtractVirtualInterface(t *testing.T) { ... }
func TestExtractGateway(t *testing.T) { ... }
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `directconnect/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `directconnect.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** iterate all regions for DescribeDirectConnectGateways — gateways are global, call from single region only
- **DO NOT** confuse `directconnect.DescribeConnectionsInput` (AWS SDK type) with local package types — AWS SDK `directconnect` is the imported package, local types are referenced without prefix
- **DO NOT** add pagination to DescribeConnections or DescribeVirtualInterfaces — these APIs are non-paginated and return all results in one call

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/directconnect/
+-- calls.go            # Direct Connect service implementation (3 AWSService entries)
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
cmd/awtest/services/codedeploy/calls.go       # Most recent implementation (regional + batch-get + 3 APIs)
cmd/awtest/services/codedeploy/calls_test.go  # Most recent test pattern (extract helper tests + batch constant test)
cmd/awtest/services/macie2/calls.go           # Reference implementation (regional + 3 APIs)
cmd/awtest/services/macie2/calls_test.go      # Reference test pattern
go.mod                                        # AWS SDK already includes directconnect package
```

### Previous Story Intelligence

**From Story 9.1 (CodeDeploy — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan (for non-paginated: `continue` to next region)
- NextToken pagination: exact pattern with `if nextToken != nil { input.NextToken = nextToken }` before call — **applicable to Call 3 only** (gateways)
- Extract helper functions for nil-safe extraction — directly applicable to `extractConnection`, `extractVirtualInterface`, `extractGateway`
- `maxBatchSize` constant — **not applicable** (Direct Connect has no batch-get APIs)
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `CodeDeployCalls[N].Process` in tests -> apply as `DirectConnectCalls[N].Process`
- Error result pattern: `return []types.ScanResult{{ServiceName: "DirectConnect", MethodName: "directconnect:DescribeConnections", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests
- Extract helper tests use real AWS SDK types (e.g., `&codedeploy.ApplicationInfo{...}`)
- 23 tests across 7 test functions

**From Code Review Findings (Stories 7.1, 7.2):**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [HIGH] Always add pagination from the start (NextToken loops on paginated APIs) — applies to Call 3 only
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [LOW] Tests should cover nil fields comprehensively
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `7b02834` — Add CodeDeploy enumeration with 3 API calls (Story 9.1)
- `79e8f63` — Mark Story 8.7 and Epic 8 as done
- `2b15c42` — Add Macie enumeration with 3 API calls (Story 8.7)
- `60147ae` — Add Athena enumeration with 3 API calls (Story 8.6)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add Direct Connect enumeration with 3 API calls (Story 9.2)"`

### Key Differences from Previous Stories

1. **Non-paginated APIs (Calls 1 & 2):** Unlike CodeDeploy/Macie which use paginated list+batch-get pattern, Direct Connect's DescribeConnections and DescribeVirtualInterfaces return all results in a single call. No pagination loop needed.
2. **Global API (Call 3):** DescribeDirectConnectGateways is a global resource API. Unlike regional calls, this is called from a single region only to avoid duplicates.
3. **Integer pointer fields:** Vlan (`*int64`) and Asn (`*int64`) fields require `fmt.Sprintf("%d", *ptr)` conversion, unlike string pointer fields. This is different from CodeDeploy which only had string/time/bool pointers.
4. **No batch-get APIs:** Direct Connect describes resources directly — no list-names-then-batch-get pattern.
5. **Gateway extract helper has no region parameter:** Since gateways are global, `extractGateway` does not take a region string.

### FRs Covered

- **FR109:** System enumerates Direct Connect connections, virtual interfaces, and gateways

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### References

- [Source: epics-phase2.md#Story 4.2: Direct Connect Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR109] — Direct Connect enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/codedeploy/calls.go] — Most recent reference implementation (regional, 3 APIs)
- [Source: cmd/awtest/services/codedeploy/calls_test.go] — Most recent reference test pattern (extract helper tests)
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (directconnect goes after config, before dynamodb)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes directconnect package)
- [Source: 9-1-codedeploy-enumeration.md] — Most recent story (regional + batch-get pattern, 3 APIs)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- SDK type discovery: `directconnect.DirectConnectGateway` does not exist in AWS SDK Go v1 — actual type is `directconnect.Gateway`. Updated extractGateway parameter and tests accordingly.

### Completion Notes List

- Implemented 3 AWSService entries: DescribeConnections (regional, non-paginated), DescribeVirtualInterfaces (regional, non-paginated), DescribeDirectConnectGateways (global, paginated with NextToken)
- All Call functions use config override pattern for region (`directconnect.New(sess, &aws.Config{Region: ...})`) — no session mutation
- Gateways called from single region `types.Regions[0]` to avoid duplicates (global resource)
- Integer pointer fields (Vlan, Asn, AmazonSideAsn) converted to string via `fmt.Sprintf("%d", *ptr)`
- 24 tests across 6 test functions: 3 Process tests (5 subtests each) + 3 extract helper tests (3 subtests each)
- No sync primitives imported — service is concurrency-unaware per NFR57
- Registered in AllServices() alphabetically after config, before dynamodb
- All quality gates pass: go build, go test, go vet, go test -race

### Change Log

- 2026-03-13: Implemented Direct Connect enumeration with 3 API calls (Story 9.2)

### File List

- cmd/awtest/services/directconnect/calls.go (NEW)
- cmd/awtest/services/directconnect/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED — added directconnect import and registration)
- vendor/github.com/aws/aws-sdk-go/service/directconnect/ (NEW — vendored SDK package)
