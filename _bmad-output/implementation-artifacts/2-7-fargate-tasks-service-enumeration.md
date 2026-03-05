# Story 2.7: Fargate Tasks Service Enumeration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Fargate tasks**,
so that **I can discover serverless container workloads accessible with the credentials, revealing running containers, task definitions, and potentially sensitive environment configurations**.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/fargate/` directory with `calls.go`
2. **AC2:** Implement Fargate task enumeration using AWS SDK v1.44.266 ECS client (`github.com/aws/aws-sdk-go/service/ecs`) -- Fargate uses the ECS API, NOT a separate SDK package
3. **AC3:** Implement AWSService interface: `Name="ecs:ListFargateTasks"`, `Call()`, `Process()`, `ModuleName=types.DefaultModuleName`
4. **AC4:** `Call()` iterates all regions in `types.Regions`, creates ECS client per region using `sess.Copy()`, calls `ListClusters` to get cluster ARNs, then for each cluster calls `ListTasks` with `LaunchType: aws.String("FARGATE")` filter, then calls `DescribeTasks` to get task details -- aggregates all `*ecs.Task` results
5. **AC5:** `Process()` displays each Fargate task: TaskArn, ClusterArn, LastStatus, DesiredStatus, TaskDefinitionArn, Cpu, Memory, LaunchType
6. **AC6:** Handle access-denied errors using `utils.HandleAWSError`
7. **AC7:** Handle empty results -- if no Fargate tasks found after all regions/clusters, call `utils.PrintAccessGranted(debug, "ecs:ListFargateTasks", "Fargate tasks")` and return empty results slice
8. **AC8:** Register service in `services/services.go` `AllServices()` function alphabetically after `eventbridge`, before `glacier`
9. **AC9:** Write table-driven tests in `calls_test.go` covering: single Fargate task, multiple tasks across clusters, empty results, access denied, nil field handling
10. **AC10:** Package naming: `fargate` (lowercase, single word)
11. **AC11:** `go build ./cmd/awtest` compiles successfully
12. **AC12:** `go test ./cmd/awtest/services/fargate/...` passes
13. **AC13:** `go vet ./cmd/awtest/...` passes clean
14. **AC14:** FR15 requirement fulfilled: System enumerates Fargate container services

## Tasks / Subtasks

- [x] Task 1: Create service package and implement Call() (AC: 1, 2, 3, 4, 10)
  - [x] Create directory `cmd/awtest/services/fargate/`
  - [x] Create `calls.go` with package `fargate`
  - [x] Define `var FargateCalls = []types.AWSService{...}`
  - [x] Implement `Call()`: iterate `types.Regions`, create `ecs.New(regionSess)` per region using `sess.Copy(&aws.Config{Region: aws.String(region)})`, call `svc.ListClusters(&ecs.ListClustersInput{})` to get cluster ARNs, then for each cluster call `svc.ListTasks(&ecs.ListTasksInput{Cluster: clusterArn, LaunchType: aws.String("FARGATE")})`, then call `svc.DescribeTasks(&ecs.DescribeTasksInput{Cluster: clusterArn, Tasks: taskArns})` to get full task details
  - [x] Handle pagination: `ListClusters` uses `NextToken`, `ListTasks` uses `NextToken`
  - [x] Use resilient per-region error handling (continue to next region on error, like Story 2.6 pattern)
  - [x] Return aggregated `[]*ecs.Task` from Call(), or nil on complete failure

- [x] Task 2: Implement Process() method (AC: 3, 5, 6, 7)
  - [x] Handle error case: call `utils.HandleAWSError(debug, "ecs:ListFargateTasks", err)`, return error ScanResult
  - [x] Type-assert output to `[]*ecs.Task`
  - [x] Handle type assertion failure (like ElastiCache pattern)
  - [x] If empty slice and no error: call `utils.PrintAccessGranted(debug, "ecs:ListFargateTasks", "Fargate tasks")`, return empty results
  - [x] For each task, extract: `TaskArn` (`*string`), `ClusterArn` (`*string`), `LastStatus` (`*string`), `DesiredStatus` (`*string`), `TaskDefinitionArn` (`*string`), `Cpu` (`*string`), `Memory` (`*string`), `LaunchType` (`*string`) -- all with nil checks
  - [x] Build `types.ScanResult` entries with: ServiceName="Fargate", MethodName="ecs:ListFargateTasks", ResourceType="task"
  - [x] Call `utils.PrintResult()` with formatted output
  - [x] Return results slice

- [x] Task 3: Register service in AllServices() (AC: 8)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/fargate"` to `services/services.go`
  - [x] Add `allServices = append(allServices, fargate.FargateCalls...)` after `eventbridge.EventbridgeCalls...` and before `glacier.GlacierCalls...`

- [x] Task 4: Write unit tests (AC: 9, 12)
  - [x] Create `cmd/awtest/services/fargate/calls_test.go`
  - [x] Follow Story 2.6 test pattern: table-driven Process()-only tests with pre-built mock data
  - [x] Test cases: single Fargate task (all fields populated), multiple tasks, empty results, access denied error, nil field handling

- [x] Task 5: Build and verify (AC: 11, 12, 13)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/fargate/...`
  - [x] `go vet ./cmd/awtest/...`

## Dev Notes

### CRITICAL: Fargate Uses ECS API -- NOT a Separate SDK Package

Fargate is NOT a separate AWS service -- it's a **launch type** within ECS. The AWS SDK package to use is `github.com/aws/aws-sdk-go/service/ecs`. Do NOT look for a `fargate` SDK package.

The enumeration requires a **three-step API chain**:
1. `ListClusters()` → get cluster ARNs
2. `ListTasks(cluster, launchType=FARGATE)` → get task ARNs per cluster
3. `DescribeTasks(cluster, tasks)` → get full task details

This is a **multi-API-call pattern** (similar to EKS Story 2.5), NOT a single-API pattern (like EFS/ElastiCache).

### CRITICAL: Use sess.Copy() for Region Iteration

Story 2.3 code review identified that mutating `sess.Config.Region` directly is unsafe. **YOU MUST USE `sess.Copy()`** for safe session handling:

```go
for _, region := range types.Regions {
    regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
    svc := ecs.New(regionSess)
    // ...
}
```

**DO NOT** use the older ECS pattern (`session.NewSession(regionConfig)`) seen in `cmd/awtest/services/ecs/calls.go`. That file uses the OLD unsafe pattern. Follow the NEWER pattern from Stories 2.3+.

### CRITICAL: Separate Package from Existing ECS Service

The existing `cmd/awtest/services/ecs/` package enumerates ECS **clusters**. This story creates a NEW `cmd/awtest/services/fargate/` package that enumerates Fargate **tasks** specifically. They are separate packages -- do NOT modify the existing ECS service.

### AWS ECS SDK Specifics for Fargate (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/ecs`

**API Chain:**

**Step 1: ListClusters**
- `svc.ListClusters(&ecs.ListClustersInput{})` → `*ecs.ListClustersOutput`
- `Output.ClusterArns` → `[]*string` -- list of cluster ARN strings
- Pagination via `NextToken`

**Step 2: ListTasks (per cluster, FARGATE only)**
- `svc.ListTasks(&ecs.ListTasksInput{Cluster: clusterArn, LaunchType: aws.String("FARGATE")})` → `*ecs.ListTasksOutput`
- `Output.TaskArns` → `[]*string` -- list of task ARN strings
- Pagination via `NextToken`
- `LaunchType` filter value must be `"FARGATE"` (uppercase)

**Step 3: DescribeTasks (per cluster, batch of task ARNs)**
- `svc.DescribeTasks(&ecs.DescribeTasksInput{Cluster: clusterArn, Tasks: taskArns})` → `*ecs.DescribeTasksOutput`
- `Output.Tasks` → `[]*ecs.Task` -- full task details
- **Max 100 tasks per DescribeTasks call** -- batch if needed

**Task fields (from `*ecs.Task`):**
- `TaskArn` -- `*string` -- full task ARN (e.g., "arn:aws:ecs:us-east-1:123456:task/cluster/task-id")
- `ClusterArn` -- `*string` -- cluster ARN the task runs in
- `TaskDefinitionArn` -- `*string` -- task definition used
- `LastStatus` -- `*string` -- "RUNNING", "STOPPED", "PENDING", etc.
- `DesiredStatus` -- `*string` -- "RUNNING" or "STOPPED"
- `LaunchType` -- `*string` -- should be "FARGATE" (filtered)
- `Cpu` -- `*string` -- CPU units (e.g., "256", "512", "1024")
- `Memory` -- `*string` -- memory in MiB (e.g., "512", "1024", "2048")

**No new dependencies needed** -- `ecs` is part of `aws-sdk-go v1.44.266` already in go.mod.

### Naming Conventions (from established patterns)

| Component | Value |
|-----------|-------|
| Package directory | `fargate` |
| Package variable | `FargateCalls` |
| AWSService.Name | `"ecs:ListFargateTasks"` |
| ScanResult.ServiceName | `"Fargate"` |
| ScanResult.MethodName | `"ecs:ListFargateTasks"` |
| ScanResult.ResourceType | `"task"` |

### Registration Order in services.go

Insert after `eventbridge` and before `glacier` (alphabetical):

```go
allServices = append(allServices, eventbridge.EventbridgeCalls...)
allServices = append(allServices, fargate.FargateCalls...)  // NEW
allServices = append(allServices, glacier.GlacierCalls...)
```

Import alphabetically after `eventbridge`:

```go
"github.com/MillerMedia/awtest/cmd/awtest/services/eventbridge"
"github.com/MillerMedia/awtest/cmd/awtest/services/fargate"  // NEW
"github.com/MillerMedia/awtest/cmd/awtest/services/glacier"
```

### Process() Output Format

```go
utils.PrintResult(debug, "", "ecs:ListFargateTasks",
    fmt.Sprintf("Found Fargate Task: %s (Status: %s, Desired: %s, CPU: %s, Memory: %s, TaskDef: %s)",
        utils.ColorizeItem(taskArn), lastStatus, desiredStatus, cpu, memory, taskDefArn), nil)
```

### Empty Results Handling

```go
if len(tasks) == 0 {
    utils.PrintAccessGranted(debug, "ecs:ListFargateTasks", "Fargate tasks")
    return results
}
```

### Reference Implementation Pattern

```go
package fargate

import (
    "fmt"
    "github.com/MillerMedia/awtest/cmd/awtest/types"
    "github.com/MillerMedia/awtest/cmd/awtest/utils"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ecs"
    "time"
)

var FargateCalls = []types.AWSService{
    {
        Name: "ecs:ListFargateTasks",
        Call: func(sess *session.Session) (interface{}, error) {
            var allTasks []*ecs.Task
            for _, region := range types.Regions {
                regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
                svc := ecs.New(regionSess)

                // Step 1: List all clusters in this region
                var clusterArns []*string
                listClustersInput := &ecs.ListClustersInput{}
                for {
                    clustersOutput, err := svc.ListClusters(listClustersInput)
                    if err != nil {
                        break // Resilient: skip region on error
                    }
                    clusterArns = append(clusterArns, clustersOutput.ClusterArns...)
                    if clustersOutput.NextToken == nil {
                        break
                    }
                    listClustersInput.NextToken = clustersOutput.NextToken
                }

                // Step 2: For each cluster, list Fargate tasks
                for _, clusterArn := range clusterArns {
                    var taskArns []*string
                    listTasksInput := &ecs.ListTasksInput{
                        Cluster:    clusterArn,
                        LaunchType: aws.String("FARGATE"),
                    }
                    for {
                        tasksOutput, err := svc.ListTasks(listTasksInput)
                        if err != nil {
                            break // Resilient: skip cluster on error
                        }
                        taskArns = append(taskArns, tasksOutput.TaskArns...)
                        if tasksOutput.NextToken == nil {
                            break
                        }
                        listTasksInput.NextToken = tasksOutput.NextToken
                    }

                    // Step 3: Describe tasks to get full details
                    if len(taskArns) > 0 {
                        describeOutput, err := svc.DescribeTasks(&ecs.DescribeTasksInput{
                            Cluster: clusterArn,
                            Tasks:   taskArns,
                        })
                        if err != nil {
                            continue // Resilient: skip cluster on error
                        }
                        allTasks = append(allTasks, describeOutput.Tasks...)
                    }
                }
            }
            return allTasks, nil
        },
        Process: func(output interface{}, err error, debug bool) []types.ScanResult {
            var results []types.ScanResult

            if err != nil {
                utils.HandleAWSError(debug, "ecs:ListFargateTasks", err)
                return []types.ScanResult{
                    {
                        ServiceName: "Fargate",
                        MethodName:  "ecs:ListFargateTasks",
                        Error:       err,
                        Timestamp:   time.Now(),
                    },
                }
            }

            tasks, ok := output.([]*ecs.Task)
            if !ok {
                utils.HandleAWSError(debug, "ecs:ListFargateTasks", fmt.Errorf("unexpected output type %T", output))
                return results
            }

            if len(tasks) == 0 {
                utils.PrintAccessGranted(debug, "ecs:ListFargateTasks", "Fargate tasks")
                return results
            }

            for _, task := range tasks {
                taskArn := ""
                if task.TaskArn != nil {
                    taskArn = *task.TaskArn
                }

                clusterArn := ""
                if task.ClusterArn != nil {
                    clusterArn = *task.ClusterArn
                }

                lastStatus := ""
                if task.LastStatus != nil {
                    lastStatus = *task.LastStatus
                }

                desiredStatus := ""
                if task.DesiredStatus != nil {
                    desiredStatus = *task.DesiredStatus
                }

                taskDefArn := ""
                if task.TaskDefinitionArn != nil {
                    taskDefArn = *task.TaskDefinitionArn
                }

                cpu := ""
                if task.Cpu != nil {
                    cpu = *task.Cpu
                }

                memory := ""
                if task.Memory != nil {
                    memory = *task.Memory
                }

                launchType := ""
                if task.LaunchType != nil {
                    launchType = *task.LaunchType
                }

                results = append(results, types.ScanResult{
                    ServiceName:  "Fargate",
                    MethodName:   "ecs:ListFargateTasks",
                    ResourceType: "task",
                    ResourceName: taskArn,
                    Details: map[string]interface{}{
                        "ClusterArn":        clusterArn,
                        "LastStatus":        lastStatus,
                        "DesiredStatus":     desiredStatus,
                        "TaskDefinitionArn": taskDefArn,
                        "Cpu":               cpu,
                        "Memory":            memory,
                        "LaunchType":        launchType,
                    },
                    Timestamp: time.Now(),
                })

                utils.PrintResult(debug, "", "ecs:ListFargateTasks",
                    fmt.Sprintf("Found Fargate Task: %s (Status: %s, Desired: %s, CPU: %s, Memory: %s, TaskDef: %s)",
                        utils.ColorizeItem(taskArn), lastStatus, desiredStatus, cpu, memory, taskDefArn), nil)
            }
            return results
        },
        ModuleName: types.DefaultModuleName,
    },
}
```

### Testing Pattern (from Story 2.6)

Create table-driven Process()-only tests with pre-built mock data. No AWS SDK mocking needed.

```go
func TestProcess(t *testing.T) {
    process := FargateCalls[0].Process

    tests := []struct {
        name          string
        input         interface{}
        err           error
        expectedCount int
        expectError   bool
        checkResults  func(t *testing.T, results []types.ScanResult)
    }{
        {
            name: "valid Fargate task with all fields",
            input: []*ecs.Task{
                {
                    TaskArn:           aws.String("arn:aws:ecs:us-east-1:123456:task/my-cluster/abc123"),
                    ClusterArn:        aws.String("arn:aws:ecs:us-east-1:123456:cluster/my-cluster"),
                    LastStatus:        aws.String("RUNNING"),
                    DesiredStatus:     aws.String("RUNNING"),
                    TaskDefinitionArn: aws.String("arn:aws:ecs:us-east-1:123456:task-definition/my-task:1"),
                    Cpu:               aws.String("256"),
                    Memory:            aws.String("512"),
                    LaunchType:        aws.String("FARGATE"),
                },
            },
            expectedCount: 1,
            checkResults: func(t *testing.T, results []types.ScanResult) {
                r := results[0]
                if r.ServiceName != "Fargate" { t.Errorf("expected ServiceName 'Fargate', got '%s'", r.ServiceName) }
                if r.MethodName != "ecs:ListFargateTasks" { t.Errorf("expected MethodName 'ecs:ListFargateTasks', got '%s'", r.MethodName) }
                if r.ResourceType != "task" { t.Errorf("expected ResourceType 'task', got '%s'", r.ResourceType) }
            },
        },
        {
            name: "multiple Fargate tasks",
            input: []*ecs.Task{
                {TaskArn: aws.String("arn:task-1"), LaunchType: aws.String("FARGATE")},
                {TaskArn: aws.String("arn:task-2"), LaunchType: aws.String("FARGATE")},
            },
            expectedCount: 2,
        },
        {
            name:          "empty results",
            input:         []*ecs.Task{},
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
            input: []*ecs.Task{
                {
                    TaskArn:           nil,
                    ClusterArn:        nil,
                    LastStatus:        nil,
                    DesiredStatus:     nil,
                    TaskDefinitionArn: nil,
                    Cpu:               nil,
                    Memory:            nil,
                    LaunchType:        nil,
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
1. **Valid Fargate task with all fields** -- all fields populated, verify ScanResult fields and Details map
2. **Multiple Fargate tasks** -- verify correct count and each task captured
3. **Empty results** -- verify `PrintAccessGranted` behavior and empty results returned
4. **Access denied** -- verify error ScanResult returned with correct ServiceName/MethodName
5. **Nil field handling** -- verify nil TaskArn, nil Cpu, nil Memory etc. handled gracefully (empty string defaults)

### Edge Cases

1. **No clusters in any region** -- ListClusters returns empty, no tasks to enumerate, Process() calls PrintAccessGranted
2. **Clusters exist but no Fargate tasks** -- ListTasks with FARGATE filter returns empty, Process() calls PrintAccessGranted
3. **Access denied on ListClusters** -- Call() continues to next region (resilient pattern)
4. **Access denied on ListTasks** -- Call() continues to next cluster (resilient pattern)
5. **Access denied on DescribeTasks** -- Call() continues to next cluster (resilient pattern)
6. **Task with nil fields** -- defensive nil checks on all pointer fields before dereferencing
7. **Many tasks per cluster** -- DescribeTasks max 100 per call (batch if taskArns > 100)
8. **Pagination on ListClusters** -- handle NextToken for accounts with many clusters
9. **Pagination on ListTasks** -- handle NextToken for clusters with many tasks

### Architecture Compliance

- **Package:** `fargate` in `cmd/awtest/services/fargate/` -- MUST FOLLOW
- **File:** `calls.go` (single file, matching all other services) -- MUST FOLLOW
- **Variable:** `FargateCalls` exported slice -- MUST FOLLOW
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
cmd/awtest/services/fargate/
+-- calls.go            # NEW: Fargate tasks service implementation
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
cmd/awtest/services/elasticache/calls.go       # Story 2.6 reference (sess.Copy + resilient regional errors + pagination)
cmd/awtest/services/ecs/calls.go               # Existing ECS cluster enum (uses OLD session pattern -- DO NOT copy session handling)
go.mod                                         # AWS SDK already included (ecs package available)
```

### Previous Story Intelligence (Story 2.6)

**Key learnings from Story 2.6 (ElastiCache):**
- **sess.Copy() is mandatory** -- continued from Story 2.3 fix. All subsequent stories MUST use this pattern.
- **Resilient per-region errors** -- Call() uses `break` on error to skip to next region instead of failing entirely. Fargate should similarly be resilient per-region AND per-cluster.
- **Pagination included from the start** -- ElastiCache added Marker-based pagination during code review. Fargate should include NextToken pagination from the start for both ListClusters and ListTasks.
- Table-driven Process()-only tests are the standard.
- Type assertion failure handling included from the start.
- All display fields must appear in BOTH PrintResult AND Details map.
- `ScanResult.Timestamp = time.Now()` is required on every result.
- Empty results handled with `utils.PrintAccessGranted`.

### Git Intelligence

**Recent commits (Epic 2 context):**
- `12e6771` Add ElastiCache service enumeration (Story 2.6)
- `312e412` Add EKS service enumeration (Story 2.5)
- `a5314de` Add EFS service enumeration (Story 2.4)
- `a416792` Add AWS Config service enumeration (Story 2.3)

**Patterns established:**
- Commits reference story numbers in message
- Build verification (go build, go test, go vet) done before commit
- Small, focused commits per story

### Project Structure Notes

- Aligns with Go standard project layout (`cmd/<app>/services/<service>/`)
- Package name `fargate` follows convention (lowercase, single word, matches directory)
- Single `calls.go` file per service -- matches all 38+ existing services
- Import path: `github.com/MillerMedia/awtest/cmd/awtest/services/fargate`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.7: Fargate Tasks Service Enumeration]
- [Source: _bmad-output/planning-artifacts/architecture.md#FR7-31 Service Enumeration]
- [Source: _bmad-output/implementation-artifacts/2-6-elasticache-clusters-service-enumeration.md -- previous story learnings]
- [Source: cmd/awtest/services/elasticache/calls.go -- Story 2.6 reference (sess.Copy + resilient + pagination)]
- [Source: cmd/awtest/services/elasticache/calls_test.go -- Story 2.6 test reference (table-driven Process()-only)]
- [Source: cmd/awtest/services/ecs/calls.go -- Existing ECS cluster enum (reference for API usage, NOT session handling)]
- [Source: cmd/awtest/services/services.go -- AllServices() registration point]
- [Source: cmd/awtest/types/types.go -- AWSService struct, ScanResult, Regions]
- [Source: cmd/awtest/utils/output.go -- PrintResult, HandleAWSError, PrintAccessGranted, ColorizeItem]
- [Source: go.mod -- aws-sdk-go v1.44.266 (includes ecs package for Fargate)]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None - clean implementation with no issues.

### Completion Notes List

- Implemented Fargate tasks enumeration using ECS SDK three-step API chain (ListClusters -> ListTasks with FARGATE filter -> DescribeTasks)
- Used sess.Copy() for safe per-region session handling (Story 2.3 fix)
- Implemented resilient per-region and per-cluster error handling (break/continue on errors)
- Added pagination support for both ListClusters and ListTasks (NextToken)
- Added DescribeTasks batching (max 100 per call)
- Process() handles: errors, type assertion failures, empty results, nil field dereferencing
- Registered service in AllServices() alphabetically between eventbridge and glacier
- All 5 test cases pass: valid task, multiple tasks, empty results, access denied, nil fields
- Full regression suite passes with zero failures
- go build, go test, go vet all pass clean

### Change Log

- 2026-03-05: Implemented Story 2.7 - Fargate Tasks Service Enumeration

### File List

- cmd/awtest/services/fargate/calls.go (NEW)
- cmd/awtest/services/fargate/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED)
- _bmad-output/implementation-artifacts/sprint-status.yaml (MODIFIED)
