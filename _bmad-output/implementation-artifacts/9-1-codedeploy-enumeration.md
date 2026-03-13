# Story 9.1: CodeDeploy Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate CodeDeploy applications, deployment groups, and deployment configurations,
So that I can discover deployment pipelines and understand how code reaches production.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/codedeploy/` directory with `calls.go` implementing CodeDeploy service enumeration with 3 AWSService entries.

2. **AC2:** Implement `codedeploy:ListApplications` API call — iterates all regions in `types.Regions`, creates CodeDeploy client per region using config override pattern (`codedeploy.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListApplicationsWithContext` with NextToken pagination to collect application names, then calls `BatchGetApplicationsWithContext` (max 25 per batch) to retrieve full details. Each application listed with Name, ApplicationId, ComputePlatform, CreateTime, LinkedToGitHub, and Region.

3. **AC3:** Implement `codedeploy:ListDeploymentGroups` API call — iterates all regions, creates CodeDeploy client per region using config override, first calls `ListApplicationsWithContext` to get all application names, then for each application calls `ListDeploymentGroupsWithContext` with NextToken pagination to collect group names, then calls `BatchGetDeploymentGroupsWithContext` (max 25 per batch per application) to retrieve full details. Each deployment group listed with ApplicationName, GroupName, DeploymentGroupId, DeploymentConfigName, ComputePlatform, ServiceRoleArn, and Region.

4. **AC4:** Implement `codedeploy:ListDeploymentConfigs` API call — iterates all regions, creates CodeDeploy client per region using config override, calls `ListDeploymentConfigsWithContext` with NextToken pagination to collect config names, then calls `GetDeploymentConfigWithContext` per config to retrieve full details. Each deployment config listed with Name, DeploymentConfigId, ComputePlatform, CreateTime, and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without CodeDeploy access, CodeDeploy is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register CodeDeploy service in `services/services.go` `AllServices()` function in alphabetical order (after `codecommit`, before `codepipeline`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 3 API calls.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/codedeploy/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `codedeploy:ListApplications` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/codedeploy/`
  - [x] Create `calls.go` with `package codedeploy`
  - [x] Define `var CodeDeployCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"codedeploy:ListApplications"`
  - [x] Call: iterate `types.Regions`, create `codedeploy.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `ListApplicationsWithContext` with NextToken pagination loop (no MaxResults param — API returns default page size). Collect all application names as `[]*string`. Then batch into groups of 25 and call `batchGetApplications` helper. Define local struct `cdApplication` with fields: Name, ApplicationId, ComputePlatform, CreateTime, LinkedToGitHub, Region. Per-region errors: `break` to next region, don't abort scan.
  - [x] Implement `batchGetApplications` helper function — takes `ctx`, `svc`, app name batch `[]*string`, and `region` string. Calls `BatchGetApplicationsWithContext`. Returns `[]cdApplication`. On error: log with `utils.HandleAWSError`, return empty slice.
  - [x] Process: handle error → `utils.HandleAWSError`, type-assert `[]cdApplication`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"CodeDeploy"`, ResourceType=`"application"`, ResourceName=appName
  - [x] `utils.PrintResult` format: `"CodeDeploy Application: %s (Platform: %s, Created: %s, Region: %s)"` with `utils.ColorizeItem(appName)`

- [x] Task 2: Implement `codedeploy:ListDeploymentGroups` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"codedeploy:ListDeploymentGroups"`
  - [x] Call: iterate regions → create CodeDeploy client with config override → call `ListApplicationsWithContext` with NextToken pagination to get all application names → for each application: call `ListDeploymentGroupsWithContext` with NextToken pagination (no MaxResults param) to collect group names → batch into groups of 25 → call `batchGetDeploymentGroups` helper to get details. Define local struct `cdDeploymentGroup` with fields: ApplicationName, GroupName, DeploymentGroupId, DeploymentConfigName, ComputePlatform, ServiceRoleArn, Region. Per-region errors: `break` to next region. Per-application errors for ListDeploymentGroups: log with `utils.HandleAWSError`, `continue` to next application.
  - [x] Implement `batchGetDeploymentGroups` helper function — takes `ctx`, `svc`, `appName` string, group name batch `[]*string`, and `region` string. Calls `BatchGetDeploymentGroupsWithContext` with ApplicationName and DeploymentGroupNames. Returns `[]cdDeploymentGroup`. On error or if `ErrorMessage` is non-nil: log with `utils.HandleAWSError`, return partial results.
  - [x] Process: type-assert `[]cdDeploymentGroup`, build `ScanResult` with ServiceName=`"CodeDeploy"`, ResourceType=`"deployment-group"`, ResourceName=appName/groupName
  - [x] `utils.PrintResult` format: `"CodeDeploy Deployment Group: %s (App: %s, Config: %s, Role: %s, Region: %s)"` with `utils.ColorizeItem(groupName)`

- [x] Task 3: Implement `codedeploy:ListDeploymentConfigs` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"codedeploy:ListDeploymentConfigs"`
  - [x] Call: iterate regions → create CodeDeploy client with config override → call `ListDeploymentConfigsWithContext` with NextToken pagination (no MaxResults param) to collect all config names → for each config name: call `GetDeploymentConfigWithContext` to get details. Define local struct `cdDeploymentConfig` with fields: Name, DeploymentConfigId, ComputePlatform, CreateTime, Region. Per-region errors: `break` to next region. Per-config errors for GetDeploymentConfig: log with `utils.HandleAWSError`, `continue` to next config.
  - [x] Process: type-assert `[]cdDeploymentConfig`, build `ScanResult` with ServiceName=`"CodeDeploy"`, ResourceType=`"deployment-config"`, ResourceName=configName
  - [x] `utils.PrintResult` format: `"CodeDeploy Deployment Config: %s (Platform: %s, Created: %s, Region: %s)"` with `utils.ColorizeItem(configName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/codedeploy"` to `services/services.go` (alphabetical in imports: after `codecommit`, before `codepipeline`)
  - [x] Add `allServices = append(allServices, codedeploy.CodeDeployCalls...)` after `codecommit.CodeCommitCalls...` and before `codepipeline.CodePipelineCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/codedeploy/calls_test.go`
  - [x] Test `ListApplications` Process: valid applications with details (name, ID, platform, creation time, GitHub link), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListDeploymentGroups` Process: valid deployment groups with details (app name, group name, group ID, config name, platform, role ARN), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListDeploymentConfigs` Process: valid configs with details (name, config ID, platform, creation time), empty results, error handling, nil fields, type assertion failure
  - [x] Use table-driven tests with `t.Run` subtests following Macie/Athena test pattern
  - [x] Access Process via `CodeDeployCalls[0].Process`, `CodeDeployCalls[1].Process`, `CodeDeployCalls[2].Process`

- [x] Task 6: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/codedeploy/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := codedeploy.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := codedeploy.New(sess)
}
```

### CodeDeploy is a REGIONAL Service

AWS CodeDeploy is **regional** — applications, deployment groups, and deployment configurations exist per-region. Iterate `types.Regions` for all three API calls, following the same pattern as CodeBuild, Athena, Macie, etc.

### CodeDeploy SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/codedeploy`

**IMPORTANT:** The Go package name is `codedeploy`. The local package name is also `codedeploy`, same pattern as `codebuild`/`codecommit` where the local package name matches the AWS SDK package name. Within `calls.go`, `codedeploy.New()` and `codedeploy.ListApplicationsInput{}` refer to the **AWS SDK package**, while local types (structs, variables) are referenced directly without package prefix.

**API Methods:**

1. **ListApplications (Paginated, returns names only):**
   - `svc.ListApplicationsWithContext(ctx, &codedeploy.ListApplicationsInput{NextToken: nextToken})` → `*codedeploy.ListApplicationsOutput`
   - `.Applications` → `[]*string` (application names only — must batch-get details)
   - Pagination: `NextToken *string` in both input and output
   - **No MaxResults parameter** — API returns default page size

2. **BatchGetApplications (Non-paginated, batch by names):**
   - `svc.BatchGetApplicationsWithContext(ctx, &codedeploy.BatchGetApplicationsInput{ApplicationNames: nameBatch})` → `*codedeploy.BatchGetApplicationsOutput`
   - `.ApplicationsInfo` → `[]*codedeploy.ApplicationInfo`
   - **Max 25 names per call** — batch input names into groups of 25
   - Each `ApplicationInfo` has:
     - `ApplicationId *string`
     - `ApplicationName *string`
     - `ComputePlatform *string` ("Server", "Lambda", "ECS")
     - `CreateTime *time.Time`
     - `LinkedToGitHub *bool`
     - `GitHubAccountName *string`

3. **ListDeploymentGroups (Paginated, returns names only, per application):**
   - `svc.ListDeploymentGroupsWithContext(ctx, &codedeploy.ListDeploymentGroupsInput{ApplicationName: aws.String(appName), NextToken: nextToken})` → `*codedeploy.ListDeploymentGroupsOutput`
   - `.DeploymentGroups` → `[]*string` (group names only)
   - `.ApplicationName` → `*string`
   - Pagination: `NextToken *string` in both input and output
   - **No MaxResults parameter** — API returns default page size
   - **Requires ApplicationName** — must enumerate applications first

4. **BatchGetDeploymentGroups (Non-paginated, batch by names, per application):**
   - `svc.BatchGetDeploymentGroupsWithContext(ctx, &codedeploy.BatchGetDeploymentGroupsInput{ApplicationName: aws.String(appName), DeploymentGroupNames: nameBatch})` → `*codedeploy.BatchGetDeploymentGroupsOutput`
   - `.DeploymentGroupsInfo` → `[]*codedeploy.DeploymentGroupInfo`
   - `.ErrorMessage` → `*string` (non-nil if some groups failed — log but don't abort)
   - **Max 25 names per call** — batch input names into groups of 25
   - Each `DeploymentGroupInfo` has:
     - `ApplicationName *string`
     - `DeploymentGroupName *string`
     - `DeploymentGroupId *string`
     - `DeploymentConfigName *string` (e.g., "CodeDeployDefault.OneAtATime")
     - `ComputePlatform *string` ("Server", "Lambda", "ECS")
     - `ServiceRoleArn *string` (IAM role for deployments)
     - `Ec2TagFilters` → `[]*codedeploy.EC2TagFilter` (optional)
     - `AutoScalingGroups` → `[]*codedeploy.AutoScalingGroup` (optional)

5. **ListDeploymentConfigs (Paginated, returns names only):**
   - `svc.ListDeploymentConfigsWithContext(ctx, &codedeploy.ListDeploymentConfigsInput{NextToken: nextToken})` → `*codedeploy.ListDeploymentConfigsOutput`
   - `.DeploymentConfigsList` → `[]*string` (config names only)
   - Pagination: `NextToken *string` in both input and output
   - **No MaxResults parameter** — API returns default page size
   - Includes AWS built-in configs (e.g., `CodeDeployDefault.OneAtATime`) plus custom configs

6. **GetDeploymentConfig (Non-paginated, single config):**
   - `svc.GetDeploymentConfigWithContext(ctx, &codedeploy.GetDeploymentConfigInput{DeploymentConfigName: aws.String(configName)})` → `*codedeploy.GetDeploymentConfigOutput`
   - `.DeploymentConfigInfo` → `*codedeploy.DeploymentConfigInfo`
   - Each `DeploymentConfigInfo` has:
     - `DeploymentConfigId *string`
     - `DeploymentConfigName *string`
     - `ComputePlatform *string`
     - `CreateTime *time.Time`
     - `MinimumHealthyHosts *codedeploy.MinimumHealthyHosts` (optional nested struct)
       - `Type *string` ("HOST_COUNT" or "FLEET_PERCENT")
       - `Value *int64`
     - `TrafficRoutingConfig *codedeploy.TrafficRoutingConfig` (optional nested struct)

**No new dependencies needed** — CodeDeploy is part of `aws-sdk-go v1.44.266` already in go.mod.

### Pagination Pattern (ListApplications, ListDeploymentGroups, ListDeploymentConfigs)

All three list APIs use NextToken pagination **without MaxResults**. Follow this exact pattern:

```go
var allAppNames []*string
for _, region := range types.Regions {
    svc := codedeploy.New(sess, &aws.Config{Region: aws.String(region)})
    var nextToken *string
    for {
        input := &codedeploy.ListApplicationsInput{}
        if nextToken != nil {
            input.NextToken = nextToken
        }
        output, err := svc.ListApplicationsWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "codedeploy:ListApplications", err)
            break
        }
        allAppNames = append(allAppNames, output.Applications...)
        if output.NextToken == nil {
            break
        }
        nextToken = output.NextToken
    }
    // After collecting names, batch-get details
    if len(allAppNames) > 0 {
        allApps = append(allApps, batchGetApplications(ctx, svc, allAppNames, region)...)
        allAppNames = nil // reset for next region
    }
}
```

### Nested List Pattern (ListDeploymentGroups — Call 2)

Call 2 requires listing applications first, then enumerating deployment groups per application. This is the same nested pattern used in CodeCommit (list repos → list branches per repo):

```go
// Per region:
// 1. List all application names
var appNames []*string
var nextToken *string
for {
    input := &codedeploy.ListApplicationsInput{}
    if nextToken != nil {
        input.NextToken = nextToken
    }
    output, err := svc.ListApplicationsWithContext(ctx, input)
    if err != nil {
        lastErr = err
        utils.HandleAWSError(false, "codedeploy:ListDeploymentGroups", err)
        break
    }
    appNames = append(appNames, output.Applications...)
    if output.NextToken == nil {
        break
    }
    nextToken = output.NextToken
}

// 2. For each application, list and batch-get deployment groups
for _, appNamePtr := range appNames {
    if appNamePtr == nil {
        continue
    }
    appName := *appNamePtr
    var groupNames []*string
    var dgNextToken *string
    for {
        dgInput := &codedeploy.ListDeploymentGroupsInput{
            ApplicationName: aws.String(appName),
        }
        if dgNextToken != nil {
            dgInput.NextToken = dgNextToken
        }
        dgOutput, err := svc.ListDeploymentGroupsWithContext(ctx, dgInput)
        if err != nil {
            utils.HandleAWSError(false, "codedeploy:ListDeploymentGroups", err)
            break
        }
        groupNames = append(groupNames, dgOutput.DeploymentGroups...)
        if dgOutput.NextToken == nil {
            break
        }
        dgNextToken = dgOutput.NextToken
    }
    if len(groupNames) > 0 {
        allGroups = append(allGroups, batchGetDeploymentGroups(ctx, svc, appName, groupNames, region)...)
    }
}
```

### Individual Get Pattern (ListDeploymentConfigs — Call 3)

Call 3 lists config names then gets individual details (no batch-get API for configs):

```go
// Per region: list config names, then get details per config
var configNames []*string
// ... pagination loop to collect configNames ...

for _, namePtr := range configNames {
    if namePtr == nil {
        continue
    }
    configName := *namePtr
    output, err := svc.GetDeploymentConfigWithContext(ctx, &codedeploy.GetDeploymentConfigInput{
        DeploymentConfigName: aws.String(configName),
    })
    if err != nil {
        utils.HandleAWSError(false, "codedeploy:ListDeploymentConfigs", err)
        continue
    }
    if output.DeploymentConfigInfo != nil {
        allConfigs = append(allConfigs, extractDeploymentConfig(output.DeploymentConfigInfo, region))
    }
}
```

### Batch-Get Helpers

**batchGetApplications:**
```go
func batchGetApplications(ctx context.Context, svc *codedeploy.CodeDeploy, names []*string, region string) []cdApplication {
    var results []cdApplication
    // Batch into groups of 25
    for i := 0; i < len(names); i += 25 {
        end := i + 25
        if end > len(names) {
            end = len(names)
        }
        batch := names[i:end]
        output, err := svc.BatchGetApplicationsWithContext(ctx, &codedeploy.BatchGetApplicationsInput{
            ApplicationNames: batch,
        })
        if err != nil {
            utils.HandleAWSError(false, "codedeploy:BatchGetApplications", err)
            continue
        }
        for _, app := range output.ApplicationsInfo {
            results = append(results, extractApplication(app, region))
        }
    }
    return results
}
```

**batchGetDeploymentGroups:**
```go
func batchGetDeploymentGroups(ctx context.Context, svc *codedeploy.CodeDeploy, appName string, names []*string, region string) []cdDeploymentGroup {
    var results []cdDeploymentGroup
    // Batch into groups of 25
    for i := 0; i < len(names); i += 25 {
        end := i + 25
        if end > len(names) {
            end = len(names)
        }
        batch := names[i:end]
        output, err := svc.BatchGetDeploymentGroupsWithContext(ctx, &codedeploy.BatchGetDeploymentGroupsInput{
            ApplicationName:      aws.String(appName),
            DeploymentGroupNames: batch,
        })
        if err != nil {
            utils.HandleAWSError(false, "codedeploy:BatchGetDeploymentGroups", err)
            continue
        }
        if output.ErrorMessage != nil && *output.ErrorMessage != "" {
            utils.HandleAWSError(false, "codedeploy:BatchGetDeploymentGroups",
                fmt.Errorf("partial error: %s", *output.ErrorMessage))
        }
        for _, dg := range output.DeploymentGroupsInfo {
            results = append(results, extractDeploymentGroup(dg, region))
        }
    }
    return results
}
```

### Nil-Safe Field Extraction Helpers

```go
func extractApplication(app *codedeploy.ApplicationInfo, region string) cdApplication {
    name := ""
    if app.ApplicationName != nil {
        name = *app.ApplicationName
    }
    appId := ""
    if app.ApplicationId != nil {
        appId = *app.ApplicationId
    }
    platform := ""
    if app.ComputePlatform != nil {
        platform = *app.ComputePlatform
    }
    createTime := ""
    if app.CreateTime != nil {
        createTime = app.CreateTime.Format(time.RFC3339)
    }
    linkedToGitHub := ""
    if app.LinkedToGitHub != nil {
        linkedToGitHub = fmt.Sprintf("%t", *app.LinkedToGitHub)
    }
    return cdApplication{
        Name:            name,
        ApplicationId:   appId,
        ComputePlatform: platform,
        CreateTime:      createTime,
        LinkedToGitHub:  linkedToGitHub,
        Region:          region,
    }
}

func extractDeploymentGroup(dg *codedeploy.DeploymentGroupInfo, region string) cdDeploymentGroup {
    appName := ""
    if dg.ApplicationName != nil {
        appName = *dg.ApplicationName
    }
    groupName := ""
    if dg.DeploymentGroupName != nil {
        groupName = *dg.DeploymentGroupName
    }
    groupId := ""
    if dg.DeploymentGroupId != nil {
        groupId = *dg.DeploymentGroupId
    }
    configName := ""
    if dg.DeploymentConfigName != nil {
        configName = *dg.DeploymentConfigName
    }
    platform := ""
    if dg.ComputePlatform != nil {
        platform = *dg.ComputePlatform
    }
    roleArn := ""
    if dg.ServiceRoleArn != nil {
        roleArn = *dg.ServiceRoleArn
    }
    return cdDeploymentGroup{
        ApplicationName:      appName,
        GroupName:            groupName,
        DeploymentGroupId:    groupId,
        DeploymentConfigName: configName,
        ComputePlatform:      platform,
        ServiceRoleArn:       roleArn,
        Region:               region,
    }
}

func extractDeploymentConfig(cfg *codedeploy.DeploymentConfigInfo, region string) cdDeploymentConfig {
    name := ""
    if cfg.DeploymentConfigName != nil {
        name = *cfg.DeploymentConfigName
    }
    configId := ""
    if cfg.DeploymentConfigId != nil {
        configId = *cfg.DeploymentConfigId
    }
    platform := ""
    if cfg.ComputePlatform != nil {
        platform = *cfg.ComputePlatform
    }
    createTime := ""
    if cfg.CreateTime != nil {
        createTime = cfg.CreateTime.Format(time.RFC3339)
    }
    return cdDeploymentConfig{
        Name:               name,
        DeploymentConfigId: configId,
        ComputePlatform:    platform,
        CreateTime:         createTime,
        Region:             region,
    }
}
```

### Local Struct Definitions

```go
type cdApplication struct {
    Name            string
    ApplicationId   string
    ComputePlatform string
    CreateTime      string
    LinkedToGitHub  string
    Region          string
}

type cdDeploymentGroup struct {
    ApplicationName      string
    GroupName            string
    DeploymentGroupId    string
    DeploymentConfigName string
    ComputePlatform      string
    ServiceRoleArn       string
    Region               string
}

type cdDeploymentConfig struct {
    Name               string
    DeploymentConfigId string
    ComputePlatform    string
    CreateTime         string
    Region             string
}
```

### Variable & Naming Conventions

- **Package:** `codedeploy` (directory: `cmd/awtest/services/codedeploy/`)
- **Exported variable:** `CodeDeployCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"codedeploy:ListApplications"`, `"codedeploy:ListDeploymentGroups"`, `"codedeploy:ListDeploymentConfigs"`
- **ScanResult.ServiceName:** `"CodeDeploy"` (PascalCase, human-readable)
- **ScanResult.ResourceType:** `"application"`, `"deployment-group"`, `"deployment-config"` (lowercase hyphenated)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **Local struct prefix:** `cd` (for CodeDeploy, following `cb` for CodeBuild, `mc` for Macie, `at` for Athena pattern)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/codedeploy"` (same name as local package — handled same as codebuild/codecommit/athena pattern)

### Registration Order in services.go

Insert alphabetically — `codedeploy` comes after `codecommit`, before `codepipeline`:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/codecommit"
"github.com/MillerMedia/awtest/cmd/awtest/services/codedeploy"      // NEW — after codecommit, before codepipeline
"github.com/MillerMedia/awtest/cmd/awtest/services/codepipeline"

// In AllServices():
allServices = append(allServices, codecommit.CodeCommitCalls...)
allServices = append(allServices, codedeploy.CodeDeployCalls...)     // NEW — after codecommit, before codepipeline
allServices = append(allServices, codepipeline.CodePipelineCalls...)
```

### Testing Pattern

Follow the Macie/Athena test pattern — test Process() functions only with pre-built mock data:

```go
func TestListApplicationsProcess(t *testing.T) {
    process := CodeDeployCalls[0].Process
    // Table-driven tests: valid apps (name, ID, platform, create time, GitHub link), empty, errors, nil fields, type assertion failure
}

func TestListDeploymentGroupsProcess(t *testing.T) {
    process := CodeDeployCalls[1].Process
    // Table-driven tests: valid groups (app name, group name, group ID, config name, platform, role ARN), empty, errors, nil fields, type assertion failure
}

func TestListDeploymentConfigsProcess(t *testing.T) {
    process := CodeDeployCalls[2].Process
    // Table-driven tests: valid configs (name, config ID, platform, create time), empty, errors, nil fields, type assertion failure
}
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `codedeploy/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `codedeploy.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** pass more than 25 names to BatchGetApplications or BatchGetDeploymentGroups — batch into groups of 25
- **DO NOT** confuse `codedeploy.ListApplicationsInput` (AWS SDK type) with local package types — AWS SDK `codedeploy` is the imported package, local types are referenced without prefix

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/codedeploy/
+-- calls.go            # CodeDeploy service implementation (3 AWSService entries)
+-- calls_test.go       # Process() tests for all 3 entries
```

**Files to MODIFY:**
```
cmd/awtest/services/services.go  # Add import + register in AllServices()
```

**Files that ALREADY EXIST (reference only, DO NOT MODIFY):**
```
cmd/awtest/types/types.go                     # AWSService struct, ScanResult, Regions
cmd/awtest/utils/output.go                    # PrintResult, HandleAWSError, ColorizeItem
cmd/awtest/services/codebuild/calls.go        # Reference implementation (regional + batch-get + 3 APIs, similar pattern)
cmd/awtest/services/codebuild/calls_test.go   # Reference test pattern
cmd/awtest/services/codecommit/calls.go       # Reference implementation (nested list: repos → branches)
cmd/awtest/services/macie2/calls.go           # Most recent implementation (regional + batch-get + 3 APIs)
cmd/awtest/services/macie2/calls_test.go      # Most recent test pattern
go.mod                                        # AWS SDK already includes codedeploy package
```

### Previous Story Intelligence

**From Story 8.7 (Macie — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- NextToken pagination: exact pattern with `if nextToken != nil { input.NextToken = nextToken }` before call
- Batch-get helper functions: `batchGetFindings` with single retry — **directly applicable** to `batchGetApplications` and `batchGetDeploymentGroups`
- `extractFinding` / `extractApplication` helper functions for nil-safe extraction
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `Macie2Calls[N].Process` in tests → apply as `CodeDeployCalls[N].Process`
- Error result pattern: `return []types.ScanResult{{ServiceName: "CodeDeploy", MethodName: "codedeploy:ListApplications", Error: err, Timestamp: time.Now()}}`
- Details map: include all relevant fields
- Tests: table-driven with `t.Run` subtests, include nil field tests and type assertion failure tests
- 16 tests across 3 test functions

**From Code Review Findings (Stories 7.1, 7.2):**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [HIGH] Always add pagination from the start (NextToken loops on paginated APIs)
- [MEDIUM] Log errors in traversal loops with `utils.HandleAWSError` before break/continue — don't silently swallow
- [MEDIUM] Use table-driven tests with `t.Run` subtests
- [LOW] Tests should cover nil fields comprehensively
- [LOW] Handle type assertion failures in all Process functions

### Git Intelligence

Recent commits follow pattern: `"Add [feature description] (Story X.Y)"`
- `79e8f63` — Mark Story 8.7 and Epic 8 as done
- `2b15c42` — Add Macie enumeration with 3 API calls (Story 8.7)
- `60147ae` — Add Athena enumeration with 3 API calls (Story 8.6)
- `2bdd8e2` — Add SageMaker enumeration with 4 API calls (Story 8.4)
- `d7271c8` — Add OpenSearch enumeration with 3 API calls (Story 8.3)
- `0dd5f6a` — Add CodeCommit enumeration with 2 API calls (Story 8.2)
- `d6dd093` — Add CodeBuild enumeration with 3 API calls (Story 8.1)
- Files created per story: `services/<name>/calls.go`, `services/<name>/calls_test.go`
- Files modified per story: `services/services.go` (import + registration)
- Pattern: single commit per story with descriptive message
- Expected commit message: `"Add CodeDeploy enumeration with 3 API calls (Story 9.1)"`

### FRs Covered

- **FR108:** System enumerates CodeDeploy applications, deployment groups, and deployment configurations

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### References

- [Source: epics-phase2.md#Story 4.1: CodeDeploy Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR108] — CodeDeploy enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/codebuild/calls.go] — Reference implementation (regional + batch-get, 3 APIs)
- [Source: cmd/awtest/services/codecommit/calls.go] — Reference implementation (nested list: repos → branches)
- [Source: cmd/awtest/services/macie2/calls.go] — Most recent reference implementation (regional + batch-get, 3 APIs)
- [Source: cmd/awtest/services/macie2/calls_test.go] — Most recent reference test pattern
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (codedeploy goes after codecommit, before codepipeline)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes codedeploy package)
- [Source: 8-7-macie-enumeration.md] — Most recent story (batch-get pattern, 3 APIs)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered.

### Completion Notes List

- Implemented all 3 CodeDeploy API calls: ListApplications (with BatchGetApplications), ListDeploymentGroups (nested list with BatchGetDeploymentGroups), ListDeploymentConfigs (with GetDeploymentConfig per config)
- All API calls use config override pattern for region (safe under concurrency per 7.2 code review fix)
- Batch helpers split names into groups of maxBatchSize (25) per API limits
- All nil-safe pointer dereferencing via extract helper functions
- No concurrency primitives used (NFR57 compliant)
- 23 table-driven tests across 7 test functions covering: valid results, empty results, access denied, nil fields, type assertion failure, extract helpers with SDK types, batch constant validation
- Registered in AllServices() alphabetically after codecommit, before codepipeline
- All validation passes: go build, go test, go vet, go test -race — zero failures, zero regressions
- Code review follow-ups addressed: replaced magic number 25 with `const maxBatchSize = 25`, added extract function tests with real AWS SDK types (TestExtractApplication, TestExtractDeploymentGroup, TestExtractDeploymentConfig), added TestMaxBatchSize

### File List

- `cmd/awtest/services/codedeploy/calls.go` (NEW) — CodeDeploy service implementation with 3 AWSService entries, maxBatchSize constant
- `cmd/awtest/services/codedeploy/calls_test.go` (NEW) — 23 table-driven tests (Process + extract helpers + batch constant)
- `cmd/awtest/services/services.go` (MODIFIED) — Added codedeploy import and registration in AllServices()

### Change Log

- 2026-03-13: Implemented CodeDeploy enumeration service with 3 API calls (ListApplications, ListDeploymentGroups, ListDeploymentConfigs), 23 unit tests, registered in AllServices()
- 2026-03-13: Addressed code review findings — replaced magic number with const maxBatchSize, added extract function tests with AWS SDK types (0 High, 1 Medium, 1 Low resolved; 2 Low accepted as tradeoffs)
