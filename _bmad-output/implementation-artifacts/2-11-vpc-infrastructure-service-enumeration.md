# Story 2.11: VPC Infrastructure Service Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate VPC infrastructure**,
so that **I can discover network infrastructure accessible with the credentials, revealing VPCs, subnets, security groups, and network topology**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/vpc/` directory with `calls.go`
2. **AC2:** Implement VPC infrastructure enumeration using AWS SDK v1.44.266 EC2 client (`github.com/aws/aws-sdk-go/service/ec2`) -- VPC APIs are part of the EC2 service namespace
3. **AC3:** Implement AWSService interface: `Name="ec2:DescribeVpcs"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates EC2 client per region using `sess.Copy()`, calls `DescribeVpcs`, `DescribeSubnets`, and `DescribeSecurityGroups` -- aggregates results into a composite struct containing all three resource types with their region
5. **AC5:** `Process()` displays each VPC: VpcId, CidrBlock, IsDefault, State, and counts of associated Subnets and Security Groups per VPC -- also displays summary of subnets (SubnetId, VpcId, CidrBlock, AvailabilityZone) and security groups (GroupId, GroupName, VpcId, Description)
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError` -- resilient per-region error handling (continue to next region on error)
7. **AC7:** Handle empty results -- if no VPCs found after all regions, call `utils.PrintAccessGranted(debug, "ec2:DescribeVpcs", "VPC infrastructure")` and return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically between `transcribe` and `waf` (NOT after waf -- `vpc` < `waf` alphabetically)
9. **AC9:** Write table-driven tests in `calls_test.go` covering: VPC with subnets and security groups, default VPC, multiple VPCs, empty results, access denied, nil field handling
10. **AC10:** Package naming: `vpc` (lowercase, single word, matches directory)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/vpc/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** VPC infrastructure enumeration requirement fulfilled (architectural requirement from Phase 1 service additions)

## Tasks / Subtasks

- [x] Task 1: Create service package and define composite result type (AC: 1, 2, 3, 10)
  - [x] Create directory `cmd/awtest/services/vpc/`
  - [x] Create `calls.go` with package `vpc`
  - [x] Define `VPCInfrastructure` struct to hold aggregated results: `VPCs []*ec2.Vpc`, `Subnets []*ec2.Subnet`, `SecurityGroups []*ec2.SecurityGroup`
  - [x] Define `var VpcCalls = []types.AWSService{...}`

- [x] Task 2: Implement Call() method (AC: 2, 3, 4, 6)
  - [x] Iterate `types.Regions`, create `ec2.New(regionSess)` per region using `sess.Copy(&aws.Config{Region: aws.String(region)})`
  - [x] Call `svc.DescribeVpcs(&ec2.DescribeVpcsInput{})` -- paginate with `NextToken`
  - [x] Call `svc.DescribeSubnets(&ec2.DescribeSubnetsInput{})` -- paginate with `NextToken`
  - [x] Call `svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})` -- paginate with `NextToken`
  - [x] Use resilient per-region error handling (continue to next region on error, `anyRegionSucceeded` + `lastErr` pattern)
  - [x] Aggregate all results into `VPCInfrastructure` struct
  - [x] Return `VPCInfrastructure` from Call(), or nil on complete failure

- [x] Task 3: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "ec2:DescribeVpcs", err)`, return error ScanResult
  - [x] Type-assert output to `VPCInfrastructure`
  - [x] Handle type assertion failure
  - [x] If all slices empty and no error: call `utils.PrintAccessGranted(debug, "ec2:DescribeVpcs", "VPC infrastructure")`, return empty results
  - [x] Build subnet-count-per-VPC map and SG-count-per-VPC map by iterating subnets/SGs and grouping by VpcId
  - [x] For each VPC, extract: `VpcId` (`*string`), `CidrBlock` (`*string`), `IsDefault` (`*bool`), `State` (`*string`) -- all with nil checks
  - [x] Include SubnetCount and SecurityGroupCount in Details map
  - [x] Build `types.ScanResult` entries with: ServiceName="VPC", MethodName="ec2:DescribeVpcs", ResourceType="vpc"
  - [x] Also create ScanResult entries for each subnet (ResourceType="subnet") and each security group (ResourceType="security-group")
  - [x] Call `utils.PrintResult()` with formatted output for each resource
  - [x] Return results slice with `Timestamp = time.Now()` on every result

- [x] Task 4: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/vpc"` to `services/services.go`
  - [x] Add `allServices = append(allServices, vpc.VpcCalls...)` between `transcribe.TranscribeCalls...` and `waf.WafCalls...`

- [x] Task 5: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/vpc/calls_test.go`
  - [x] Follow established test pattern: table-driven Process()-only tests with pre-built mock data
  - [x] Test cases: VPC with subnets and SGs, default VPC (IsDefault=true), multiple VPCs with mixed subnets/SGs, empty results, access denied error, nil field handling

- [x] Task 6: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/vpc/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### VPC Uses EC2 SDK Package -- Most Complex Epic 2 Story

VPC APIs (`DescribeVpcs`, `DescribeSubnets`, `DescribeSecurityGroups`) are part of the **EC2 service namespace** in AWS. Unlike previous Epic 2 stories (single API per service), this story requires **THREE separate API calls** aggregated into a single result.

**Package:** `github.com/aws/aws-sdk-go/service/ec2` -- this is the SAME package used by the existing `cmd/awtest/services/ec2/` service. This is fine -- different Go packages can import the same SDK package.

### CRITICAL: Use sess.Copy() for Region Iteration

Story 2.3 code review identified that mutating `sess.Config.Region` directly is unsafe (the existing EC2 service in `cmd/awtest/services/ec2/calls.go` still uses the OLD unsafe pattern -- do NOT copy that). **YOU MUST USE `sess.Copy()`** for safe session handling:

```go
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := ec2.New(regionSess)
    // ...
}
```

### CRITICAL: Do NOT Confuse with Existing EC2 Service

The existing `cmd/awtest/services/ec2/` package handles EC2 instances (DescribeInstances). This new `cmd/awtest/services/vpc/` package handles VPC infrastructure. They are SEPARATE services with SEPARATE packages, even though both use the `ec2` SDK package.

### Composite Result Type

Since Call() must return a single `(interface{}, error)`, define a composite struct:

```go
type VPCInfrastructure struct {
    VPCs           []*ec2.Vpc
    Subnets        []*ec2.Subnet
    SecurityGroups []*ec2.SecurityGroup
}
```

### AWS EC2 SDK Specifics for VPC APIs (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/ec2`

**API Call 1 - DescribeVpcs:**
- `svc.DescribeVpcs(&ec2.DescribeVpcsInput{})` -> `*ec2.DescribeVpcsOutput`
- `Output.Vpcs` -> `[]*ec2.Vpc`
- Pagination via `NextToken` (string)
- Fields from `*ec2.Vpc`:
  - `VpcId` -- `*string` -- VPC identifier (e.g., "vpc-0123456789abcdef0")
  - `CidrBlock` -- `*string` -- primary CIDR block (e.g., "10.0.0.0/16")
  - `State` -- `*string` -- "available" or "pending"
  - `IsDefault` -- `*bool` -- whether this is the default VPC
  - `OwnerId` -- `*string` -- AWS account ID
  - `Tags` -- `[]*ec2.Tag` -- resource tags

**API Call 2 - DescribeSubnets:**
- `svc.DescribeSubnets(&ec2.DescribeSubnetsInput{})` -> `*ec2.DescribeSubnetsOutput`
- `Output.Subnets` -> `[]*ec2.Subnet`
- Pagination via `NextToken` (string)
- Fields from `*ec2.Subnet`:
  - `SubnetId` -- `*string` -- subnet identifier
  - `VpcId` -- `*string` -- parent VPC ID (used for correlation)
  - `CidrBlock` -- `*string` -- subnet CIDR block
  - `AvailabilityZone` -- `*string` -- AZ (e.g., "us-east-1a")
  - `State` -- `*string` -- "available" or "pending"
  - `AvailableIpAddressCount` -- `*int64` -- remaining IPs
  - `DefaultForAz` -- `*bool` -- whether default for this AZ

**API Call 3 - DescribeSecurityGroups:**
- `svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})` -> `*ec2.DescribeSecurityGroupsOutput`
- `Output.SecurityGroups` -> `[]*ec2.SecurityGroup`
- Pagination via `NextToken` (string)
- Fields from `*ec2.SecurityGroup`:
  - `GroupId` -- `*string` -- security group identifier
  - `GroupName` -- `*string` -- security group name
  - `VpcId` -- `*string` -- parent VPC ID (used for correlation)
  - `Description` -- `*string` -- group description
  - `IpPermissions` -- `[]*ec2.IpPermission` -- inbound rules (count for display)
  - `IpPermissionsEgress` -- `[]*ec2.IpPermission` -- outbound rules (count for display)

### Pagination Pattern for All Three APIs

All three APIs use the same NextToken pagination pattern:

```go
var allVpcs []*ec2.Vpc
input := &ec2.DescribeVpcsInput{}
for {
    output, err := svc.DescribeVpcs(input)
    if err != nil {
        lastErr = err
        regionFailed = true
        break
    }
    allVpcs = append(allVpcs, output.Vpcs...)
    if output.NextToken == nil {
        break
    }
    input.NextToken = output.NextToken
}
```

Repeat the same pattern for DescribeSubnets and DescribeSecurityGroups.

### VPC-Subnet-SecurityGroup Correlation in Process()

In Process(), build maps to correlate subnets and SGs to their parent VPCs:

```go
subnetsByVPC := make(map[string]int)
for _, subnet := range infra.Subnets {
    if subnet.VpcId != nil {
        subnetsByVPC[*subnet.VpcId]++
    }
}

sgsByVPC := make(map[string]int)
for _, sg := range infra.SecurityGroups {
    if sg.VpcId != nil {
        sgsByVPC[*sg.VpcId]++
    }
}
```

Then include counts in VPC results:
```go
results = append(results, types.ScanResult{
    ServiceName:  "VPC",
    MethodName:   "ec2:DescribeVpcs",
    ResourceType: "vpc",
    ResourceName: vpcId,
    Details: map[string]interface{}{
        "CidrBlock":          cidrBlock,
        "State":              state,
        "IsDefault":          isDefault,
        "SubnetCount":        subnetsByVPC[vpcId],
        "SecurityGroupCount": sgsByVPC[vpcId],
    },
    Timestamp: time.Now(),
})
```

### Process() Output Format

**For VPCs:**
```go
utils.PrintResult(debug, "", "ec2:DescribeVpcs",
    fmt.Sprintf("Found VPC: %s (CIDR: %s, State: %s, Default: %v, Subnets: %d, SecurityGroups: %d)",
        utils.ColorizeItem(vpcId), cidrBlock, state, isDefault, subnetCount, sgCount), nil)
```

**For Subnets:**
```go
utils.PrintResult(debug, "", "ec2:DescribeSubnets",
    fmt.Sprintf("Found Subnet: %s (VPC: %s, CIDR: %s, AZ: %s)",
        utils.ColorizeItem(subnetId), vpcId, cidrBlock, az), nil)
```

**For Security Groups:**
```go
utils.PrintResult(debug, "", "ec2:DescribeSecurityGroups",
    fmt.Sprintf("Found Security Group: %s (%s, VPC: %s, InboundRules: %d, OutboundRules: %d)",
        utils.ColorizeItem(groupId), groupName, vpcId, inboundCount, outboundCount), nil)
```

### Naming Conventions

| Component | Value |
|-----------|-------|
| Package directory | `vpc` |
| Package variable | `VpcCalls` |
| AWSService.Name | `"ec2:DescribeVpcs"` |
| ScanResult.ServiceName | `"VPC"` |
| ScanResult.MethodName (VPCs) | `"ec2:DescribeVpcs"` |
| ScanResult.MethodName (Subnets) | `"ec2:DescribeSubnets"` |
| ScanResult.MethodName (SGs) | `"ec2:DescribeSecurityGroups"` |
| ScanResult.ResourceType (VPCs) | `"vpc"` |
| ScanResult.ResourceType (Subnets) | `"subnet"` |
| ScanResult.ResourceType (SGs) | `"security-group"` |

**Note on AWSService.Name:** The AWS IAM service prefix for VPC APIs is `ec2`. There is no separate `vpc:` namespace in IAM. The primary API call is `ec2:DescribeVpcs`, so use that as the AWSService.Name.

### Registration Order in services.go

Insert between `transcribe` and `waf` (alphabetical by package name: `v` comes before `w`):

```go
allServices = append(allServices, transcribe.TranscribeCalls...)
allServices = append(allServices, vpc.VpcCalls...)  // NEW
allServices = append(allServices, waf.WafCalls...)
```

Import alphabetically between `transcribe` and `waf`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/transcribe"
"github.com/MillerMedia/awtest/cmd/awtest/services/vpc"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/waf"
```

**Note:** The epics file incorrectly states "after waf (last in alphabetical order)" -- `vpc` < `waf` alphabetically, so it goes BEFORE waf.

### Empty Results Handling

```go
if len(infra.VPCs) == 0 && len(infra.Subnets) == 0 && len(infra.SecurityGroups) == 0 {
    utils.PrintAccessGranted(debug, "ec2:DescribeVpcs", "VPC infrastructure")
    return []types.ScanResult{}
}
```

### Resilient Per-Region Error Handling

Follow the established `anyRegionSucceeded` + `lastErr` pattern from Stories 2.8-2.10. If DescribeVpcs fails in a region, skip that region entirely (don't try subnets/SGs for that region). If all regions fail, return the last error.

```go
var lastErr error
anyRegionSucceeded := false
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := ec2.New(regionSess)
    regionFailed := false

    // DescribeVpcs
    vpcs, err := describeAllVpcs(svc)
    if err != nil {
        lastErr = err
        regionFailed = true
    }

    // Only fetch subnets/SGs if VPCs call succeeded
    if !regionFailed {
        subnets, err := describeAllSubnets(svc)
        if err != nil {
            // Non-fatal: VPCs still valid, just no subnet data
            // Continue but note the error
        }
        sgs, err := describeAllSecurityGroups(svc)
        if err != nil {
            // Non-fatal: VPCs still valid, just no SG data
        }
        // aggregate results...
        anyRegionSucceeded = true
    }
}
```

### Testing Pattern

Create table-driven Process()-only tests with pre-built mock data. No AWS SDK mocking needed.

Test cases:
1. **VPC with subnets and SGs** -- VPC with 2 subnets and 2 SGs, verify correlation counts
2. **Default VPC** -- verify IsDefault=true is captured correctly
3. **Multiple VPCs** -- 2+ VPCs with different subnet/SG distributions
4. **Empty results** -- verify PrintAccessGranted behavior and empty results returned
5. **Access denied** -- verify error ScanResult returned with correct ServiceName/MethodName
6. **Nil field handling** -- verify nil VpcId, nil CidrBlock, nil State, nil IsDefault handled gracefully

**IMPORTANT for tests:** When creating mock EC2 types:
```go
import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/ec2"
)

// Mock VPC
&ec2.Vpc{
    VpcId:     aws.String("vpc-0123456789abcdef0"),
    CidrBlock: aws.String("10.0.0.0/16"),
    State:     aws.String("available"),
    IsDefault: aws.Bool(true),
}

// Mock Subnet
&ec2.Subnet{
    SubnetId:         aws.String("subnet-0123456789abcdef0"),
    VpcId:            aws.String("vpc-0123456789abcdef0"),
    CidrBlock:        aws.String("10.0.1.0/24"),
    AvailabilityZone: aws.String("us-east-1a"),
}

// Mock Security Group
&ec2.SecurityGroup{
    GroupId:     aws.String("sg-0123456789abcdef0"),
    GroupName:   aws.String("default"),
    VpcId:       aws.String("vpc-0123456789abcdef0"),
    Description: aws.String("default VPC security group"),
    IpPermissions:       []*ec2.IpPermission{},
    IpPermissionsEgress: []*ec2.IpPermission{},
}
```

### Edge Cases

1. **No VPCs in any region** -- DescribeVpcs returns empty, Process() calls PrintAccessGranted
2. **Access denied in all regions** -- Call() returns nil + error, Process() handles error
3. **Access denied in some regions** -- Call() continues to next region (resilient pattern), returns partial results
4. **VPC exists but no subnets/SGs accessible** -- Still display VPC with counts of 0
5. **DescribeVpcs succeeds but DescribeSubnets fails** -- Still return VPC data, show subnet count as 0
6. **Nil IsDefault** -- defensive nil check, default to false
7. **Security groups without VpcId** -- EC2-Classic SGs may have nil VpcId; handle gracefully
8. **Large accounts** -- Pagination on all three APIs; handle NextToken for accounts with many VPCs/subnets/SGs

### Architecture Compliance

- **Package:** `vpc` in `cmd/awtest/services/vpc/` -- MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) -- MUST FOLLOW
- **Variable:** `VpcCalls` exported slice -- MUST FOLLOW
- **Type:** `[]types.AWSService` -- MUST FOLLOW
- **ModuleName:** `types.DefaultModuleName` -- MUST FOLLOW (all Epic 2 stories use DefaultModuleName)
- **Session handling:** `sess.Copy(&aws.Config{Region: aws.String(region)})` -- MUST FOLLOW (Story 2.3 code review fix)
- **Error handling:** `utils.HandleAWSError(debug, methodName, err)` -- MUST FOLLOW
- **Region iteration:** `for _, region := range types.Regions` -- MUST FOLLOW
- **Nil checks:** Always check `*string`, `*bool`, `*int64` before dereferencing -- MUST FOLLOW
- **Go version:** 1.19 (no generics, no new stdlib features) -- MUST FOLLOW
- **SDK version:** AWS SDK Go v1.44.266 -- MUST FOLLOW (do NOT use SDK v2)

### File Structure

**Files to CREATE:**
```
cmd/awtest/services/vpc/
+-- calls.go            # NEW: VPC infrastructure service implementation
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
cmd/awtest/services/systemsmanager/calls.go    # Story 2.10 reference (sess.Copy + resilient + pagination)
cmd/awtest/services/systemsmanager/calls_test.go # Story 2.10 test reference
cmd/awtest/services/ec2/calls.go               # Existing EC2 service (DO NOT confuse with VPC -- different package)
cmd/awtest/services/services.go                # AllServices() registration point
go.mod                                         # AWS SDK already included (ec2 package available)
```

### Previous Story Intelligence (Story 2.10 - Systems Manager)

**Key learnings from Story 2.10 (Systems Manager):**
- **sess.Copy() is mandatory** -- continued pattern from Story 2.3 fix
- **Resilient per-region errors** with `anyRegionSucceeded` + `lastErr` tracking pattern
- **Pagination included from the start** -- avoid code review rework
- Table-driven Process()-only tests are the standard
- Type assertion failure handling included from the start
- All display fields must appear in BOTH PrintResult AND Details map
- `ScanResult.Timestamp = time.Now()` is required on every result
- Empty results handled with `utils.PrintAccessGranted`

### Git Intelligence

**Recent commits (Epic 2 context):**
- `712d5ea` Mark Story 2.10 as done
- `e55e8c7` Add Systems Manager SSM parameters service enumeration (Story 2.10)
- `c94b560` Mark Story 2.9 as done
- `a086ad4` Add Step Functions state machines service enumeration (Story 2.9)
- `ebf7392` Mark Story 2.8 as done
- `40c71c3` Add Redshift clusters service enumeration (Story 2.8)
- `f814fab` Fix false "Access granted" when all regions return access denied

**Key insight from f814fab:** The `anyRegionSucceeded` + `lastErr` pattern in Call() is critical. If all regions return access denied, Call() must return the error (not nil) so Process() can properly report it.

**This is the LAST story in Epic 2.** After completion, Epic 2 can be marked as done (pending retrospective).

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `vpc` follows convention (lowercase, single word, matches directory)
- Single `calls.go` file per service -- matches all 40+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/vpc`
- Uses same `ec2` SDK package as existing `cmd/awtest/services/ec2/` -- this is fine, Go allows it

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.11: VPC Infrastructure Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#Phase 1 Additions -- VPC]
- [Source: _bmad-output/planning-artifacts/architecture.md#Service Enumeration Pattern]
- [Source: _bmad-output/implementation-artifacts/2-10-systems-manager-ssm-parameters-enumeration.md -- previous story learnings]
- [Source: cmd/awtest/services/systemsmanager/calls.go -- recent reference (sess.Copy + resilient + NextToken pagination)]
- [Source: cmd/awtest/services/systemsmanager/calls_test.go -- test reference (table-driven Process()-only)]
- [Source: cmd/awtest/services/ec2/calls.go -- existing EC2 service (DO NOT confuse with VPC)]
- [Source: cmd/awtest/services/services.go -- AllServices() registration point]
- [Source: cmd/awtest/types/types.go -- AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go -- PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem]
- [Source: go.mod -- aws-sdk-go v1.44.266 (includes ec2 package)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None - clean implementation with no issues.

### Completion Notes List

- Implemented VPC infrastructure enumeration with three aggregated API calls (DescribeVpcs, DescribeSubnets, DescribeSecurityGroups)
- Used `VPCInfrastructure` composite struct to aggregate results from all three APIs
- Call() iterates all regions using `sess.Copy()` (safe session pattern from Story 2.3 fix)
- Resilient per-region error handling: if DescribeVpcs fails in a region, skip that region entirely; if subnets/SGs fail, VPC data still preserved
- Pagination with NextToken implemented for all three APIs
- Process() builds VPC-to-subnet and VPC-to-SG correlation maps for count display
- Produces ScanResult entries for VPCs (ResourceType="vpc"), Subnets (ResourceType="subnet"), and Security Groups (ResourceType="security-group")
- All nil pointer checks in place for *string, *bool fields
- Registered in AllServices() alphabetically between transcribe and waf
- 6 table-driven test cases covering: VPC with subnets/SGs, default VPC, multiple VPCs, empty results, access denied, nil field handling
- All tests pass, go build succeeds, go vet clean, no regressions

### Change Log

- 2026-03-05: Implemented VPC infrastructure service enumeration (Story 2.11) - created vpc package with calls.go and calls_test.go, registered in services.go

### File List

- cmd/awtest/services/vpc/calls.go (NEW)
- cmd/awtest/services/vpc/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED)
