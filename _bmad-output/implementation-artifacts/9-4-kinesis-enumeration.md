# Story 9.4: Kinesis Enumeration

Status: done

<!-- Generated: 2026-03-13 by BMAD Create Story Workflow -->
<!-- Epic: 9 - Infrastructure & Data Service Expansion (Phase 2 Epic 4) -->
<!-- FR: FR111 | Source: epics-phase2.md#Story 4.4 -->
<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a pentester,
I want to enumerate Kinesis streams, shard details, and consumer applications,
So that I can discover real-time data streams and understand data flow paths.

## Acceptance Criteria

1. **AC1:** Create `cmd/awtest/services/kinesis/` directory with `calls.go` implementing Kinesis service enumeration with 3 AWSService entries.

2. **AC2:** Implement `kinesis:ListStreams` API call — iterates all regions in `types.Regions`, creates Kinesis client per region using config override pattern (`kinesis.New(sess, &aws.Config{Region: aws.String(region)})`), calls `ListStreamsWithContext` with `NextToken` pagination to list all stream summaries via `StreamSummaries`. Each stream listed with StreamName, StreamARN, StreamStatus, StreamMode, CreationTimestamp, and Region.

3. **AC3:** Implement `kinesis:ListShards` API call — iterates all regions in `types.Regions`, creates Kinesis client per region using config override pattern. First lists all streams via `ListStreamsWithContext` with `NextToken` pagination (collecting stream names), then for each stream calls `ListShardsWithContext` with `NextToken` pagination. Each shard listed with ShardId, StreamName, ParentShardId, StartingHashKey, EndingHashKey, StartingSequenceNumber, EndingSequenceNumber, and Region.

4. **AC4:** Implement `kinesis:ListStreamConsumers` API call — iterates all regions in `types.Regions`, creates Kinesis client per region using config override pattern. First lists all streams via `ListStreamsWithContext` with `NextToken` pagination (collecting stream names and ARNs), then for each stream calls `ListStreamConsumersWithContext` with the stream's ARN and `NextToken` pagination. Each consumer listed with ConsumerName, ConsumerARN, ConsumerStatus, StreamName, CreationTimestamp, and Region.

5. **AC5:** All three Process() functions handle errors via `utils.HandleAWSError`, type-assert output, and perform nil-safe pointer dereferencing on all AWS SDK pointer fields.

6. **AC6:** Given credentials without Kinesis access, Kinesis is skipped silently (access denied handling via existing error classification in safeScan).

7. **AC7:** Register Kinesis service in `services/services.go` `AllServices()` function in alphabetical order (after `ivsrealtime`, before `kms`).

8. **AC8:** Write table-driven tests in `calls_test.go` covering: valid results, empty results, access denied errors, nil field handling, type assertion failure handling for all 3 API calls.

9. **AC9:** Service contains no sync primitives (`sync`, `sync/atomic`) — concurrency-unaware per NFR57.

10. **AC10:** `go build ./cmd/awtest`, `go test ./cmd/awtest/services/kinesis/...`, and `go vet ./cmd/awtest/...` all pass clean.

## Tasks / Subtasks

- [x] Task 1: Create service package and implement `kinesis:ListStreams` (AC: 1, 2, 5, 6, 9)
  - [x] Create directory `cmd/awtest/services/kinesis/`
  - [x] Create `calls.go` with `package kinesis`
  - [x] Define `var KinesisCalls = []types.AWSService{...}` with 3 entries
  - [x] Implement first entry: Name `"kinesis:ListStreams"`
  - [x] Call: iterate `types.Regions`, create `kinesis.New(sess, &aws.Config{Region: aws.String(region)})` (**DO NOT** mutate `sess.Config.Region` — use config override per 7.2 code review fix), call `ListStreamsWithContext` with `NextToken` pagination. Use `output.StreamSummaries` (available in SDK v1.44.266) to get StreamName, StreamARN, StreamStatus, StreamModeDetails. Define local struct `kinesisStream` with fields: StreamName, StreamARN, StreamStatus, StreamMode, CreationTimestamp, Region. Per-region errors: `break` pagination loop, don't abort scan.
  - [x] Implement `extractStream` helper function — takes `*kinesis.StreamSummary` and `region` string, returns `kinesisStream` with nil-safe pointer dereferencing. Note: `StreamModeDetails` is a nested struct — extract via `summary.StreamModeDetails.StreamMode` with nil checks at each level. `StreamCreationTimestamp` is `*time.Time` — format with `time.RFC3339`.
  - [x] Process: handle error -> `utils.HandleAWSError`, type-assert `[]kinesisStream`, extract fields with nil-safe checks, build `ScanResult` with ServiceName=`"Kinesis"`, ResourceType=`"stream"`, ResourceName=streamName
  - [x] `utils.PrintResult` format: `"Kinesis Stream: %s (Status: %s, Mode: %s, Region: %s)"` with `utils.ColorizeItem(streamName)`

- [x] Task 2: Implement `kinesis:ListShards` (AC: 3, 5, 6, 9)
  - [x] Implement second entry: Name `"kinesis:ListShards"`
  - [x] Call: iterate regions -> create Kinesis client with config override -> Step 1: list all streams via `ListStreamsWithContext` with `NextToken` pagination (collect stream names from `StreamSummaries`) -> Step 2: for each stream, call `ListShardsWithContext` with `StreamName` parameter and `NextToken` pagination. Define local struct `kinesisShard` with fields: ShardId, StreamName, ParentShardId, StartingHashKey, EndingHashKey, StartingSequenceNumber, EndingSequenceNumber, Region. Per-region errors: `break` to next region for stream listing; per-stream errors: `continue` to next stream.
  - [x] Implement `extractShard` helper function — takes `*kinesis.Shard`, `streamName`, and `region` string, returns `kinesisShard` with nil-safe pointer dereferencing. Note: `HashKeyRange` is nested — extract `StartingHashKey` via `shard.HashKeyRange.StartingHashKey` with nil checks. Similarly `SequenceNumberRange` is nested — extract `StartingSequenceNumber` via `shard.SequenceNumberRange.StartingSequenceNumber` with nil checks.
  - [x] Process: type-assert `[]kinesisShard`, build `ScanResult` with ServiceName=`"Kinesis"`, ResourceType=`"shard"`, ResourceName=shardId
  - [x] `utils.PrintResult` format: `"Kinesis Shard: %s (Stream: %s, Region: %s)"` with `utils.ColorizeItem(shardId)`

- [x] Task 3: Implement `kinesis:ListStreamConsumers` (AC: 4, 5, 6, 9)
  - [x] Implement third entry: Name `"kinesis:ListStreamConsumers"`
  - [x] Call: iterate regions -> create Kinesis client with config override -> Step 1: list all streams via `ListStreamsWithContext` with `NextToken` pagination (collect stream names and ARNs from `StreamSummaries`) -> Step 2: for each stream, call `ListStreamConsumersWithContext` with `StreamARN` parameter and `NextToken` pagination. Define local struct `kinesisConsumer` with fields: ConsumerName, ConsumerARN, ConsumerStatus, StreamName, CreationTimestamp, Region. Per-stream errors: `continue` to next stream.
  - [x] Implement `extractConsumer` helper function — takes `*kinesis.Consumer`, `streamName`, and `region` string, returns `kinesisConsumer` with nil-safe pointer dereferencing. Note: `ConsumerCreationTimestamp` is `*time.Time` — format with `time.RFC3339`.
  - [x] Process: type-assert `[]kinesisConsumer`, build `ScanResult` with ServiceName=`"Kinesis"`, ResourceType=`"consumer"`, ResourceName=consumerName
  - [x] `utils.PrintResult` format: `"Kinesis Consumer: %s (Stream: %s, Status: %s, Region: %s)"` with `utils.ColorizeItem(consumerName)`

- [x] Task 4: Register service in AllServices() (AC: 7)
  - [x] Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/kinesis"` to `services/services.go` (alphabetical in imports: after `ivsrealtime`, before `kms`)
  - [x] Add `allServices = append(allServices, kinesis.KinesisCalls...)` after `ivsrealtime.IvsRealtimeCalls...` and before `kms.KMSCalls...`

- [x] Task 5: Write unit tests (AC: 8, 10)
  - [x] Create `cmd/awtest/services/kinesis/calls_test.go`
  - [x] Test `ListStreams` Process: valid streams with details (name, ARN, status, mode, creation timestamp), empty results, access denied error, nil fields, type assertion failure
  - [x] Test `ListShards` Process: valid shards with details (shard ID, stream name, parent shard ID, hash key range, sequence number range), empty results, error handling, nil fields, type assertion failure
  - [x] Test `ListStreamConsumers` Process: valid consumers with details (name, ARN, status, stream name, creation timestamp), empty results, error handling, nil fields, type assertion failure
  - [x] Test extract helpers: `TestExtractStream`, `TestExtractShard`, `TestExtractConsumer` with AWS SDK types (both populated and nil fields)
  - [x] Use table-driven tests with `t.Run` subtests following EMR/CodeDeploy/DirectConnect test pattern
  - [x] Access Process via `KinesisCalls[0].Process`, `KinesisCalls[1].Process`, `KinesisCalls[2].Process`

- [x] Task 6: Vendor Kinesis SDK package (AC: 10)
  - [x] Run `go mod vendor` or manually ensure `vendor/github.com/aws/aws-sdk-go/service/kinesis/` is populated
  - [x] Kinesis package is part of `aws-sdk-go v1.44.266` — already in go.mod, just needs vendoring

- [x] Task 7: Build and verify (AC: 10)
  - [x] `go build ./cmd/awtest`
  - [x] `go test ./cmd/awtest/services/kinesis/...`
  - [x] `go vet ./cmd/awtest/...`
  - [x] `go test -race ./cmd/awtest/...` (verify no race conditions in full suite)

## Dev Notes

### CRITICAL: Session Region Pattern — Config Override, NOT Mutation

Per code review findings from Story 7.2, **DO NOT** mutate `sess.Config.Region` in place. Use the config override pattern in the service constructor:

```go
// CORRECT: Config override — safe under concurrent execution
for _, region := range types.Regions {
    svc := kinesis.New(sess, &aws.Config{Region: aws.String(region)})
    // ... use svc
}

// WRONG: Session mutation — race condition under concurrent execution
for _, region := range types.Regions {
    sess.Config.Region = aws.String(region)  // DO NOT DO THIS
    svc := kinesis.New(sess)
}
```

### CRITICAL: Kinesis Uses `NextToken` Pagination

Unlike EMR (which uses `Marker`), Kinesis APIs use `NextToken` for pagination. This applies to all 3 API calls:

```go
var nextToken *string
for {
    input := &kinesis.ListStreamsInput{}
    if nextToken != nil {
        input.NextToken = nextToken
    }
    output, err := svc.ListStreamsWithContext(ctx, input)
    if err != nil {
        lastErr = err
        utils.HandleAWSError(false, "kinesis:ListStreams", err)
        break
    }
    for _, summary := range output.StreamSummaries {
        if summary != nil {
            // Process stream...
        }
    }
    if output.NextToken == nil {
        break
    }
    nextToken = output.NextToken
}
```

### Kinesis SDK v1 Specifics (AWS SDK Go v1.44.266)

**Package:** `github.com/aws/aws-sdk-go/service/kinesis`

**IMPORTANT:** The Go package name is `kinesis`. The local package name is also `kinesis`, same pattern as `emr`/`codedeploy`/`directconnect` where the local package name matches the AWS SDK package name. Within `calls.go`, `kinesis.New()` and `kinesis.ListStreamsInput{}` refer to the **AWS SDK package**, while local types (structs, variables) are referenced directly without package prefix.

**API Methods:**

1. **ListStreams (Paginated with NextToken, regional):**
   - `svc.ListStreamsWithContext(ctx, &kinesis.ListStreamsInput{NextToken: nextToken})` -> `*kinesis.ListStreamsOutput`
   - `.StreamSummaries` -> `[]*kinesis.StreamSummary` (available in SDK v1.44.266)
   - `.StreamNames` -> `[]*string` (legacy field — DO NOT USE, prefer StreamSummaries)
   - Pagination: `NextToken *string` in both input and output
   - Each `StreamSummary` has:
     - `StreamName *string`
     - `StreamARN *string`
     - `StreamStatus *string` ("CREATING", "DELETING", "ACTIVE", "UPDATING")
     - `StreamModeDetails *kinesis.StreamModeDetails` → `.StreamMode *string` ("PROVISIONED", "ON_DEMAND")
     - `StreamCreationTimestamp *time.Time`

2. **ListShards (Paginated with NextToken, per-stream):**
   - `svc.ListShardsWithContext(ctx, &kinesis.ListShardsInput{StreamName: aws.String(name), NextToken: nextToken})` -> `*kinesis.ListShardsOutput`
   - `.Shards` -> `[]*kinesis.Shard`
   - Pagination: `NextToken *string` in both input and output
   - **IMPORTANT:** When using `NextToken` for subsequent pages, do NOT include `StreamName` — only include `StreamName` on the first request. On subsequent requests, only include `NextToken`.
   - Each `Shard` has:
     - `ShardId *string` (e.g., "shardId-000000000000")
     - `ParentShardId *string`
     - `AdjacentParentShardId *string`
     - `HashKeyRange *kinesis.HashKeyRange` → `.StartingHashKey *string`, `.EndingHashKey *string`
     - `SequenceNumberRange *kinesis.SequenceNumberRange` → `.StartingSequenceNumber *string`, `.EndingSequenceNumber *string`

3. **ListStreamConsumers (Paginated with NextToken, per-stream):**
   - `svc.ListStreamConsumersWithContext(ctx, &kinesis.ListStreamConsumersInput{StreamARN: aws.String(arn), NextToken: nextToken})` -> `*kinesis.ListStreamConsumersOutput`
   - `.Consumers` -> `[]*kinesis.Consumer`
   - Pagination: `NextToken *string` in both input and output
   - **IMPORTANT:** `ListStreamConsumers` requires `StreamARN` (not StreamName). Get ARN from `StreamSummary.StreamARN` in ListStreams.
   - Each `Consumer` has:
     - `ConsumerName *string`
     - `ConsumerARN *string`
     - `ConsumerStatus *string` ("CREATING", "DELETING", "ACTIVE")
     - `ConsumerCreationTimestamp *time.Time`

### Simple Paginated Pattern (Call 1: ListStreams)

```go
var allStreams []kinesisStream
var lastErr error

for _, region := range types.Regions {
    svc := kinesis.New(sess, &aws.Config{Region: aws.String(region)})
    var nextToken *string
    for {
        input := &kinesis.ListStreamsInput{}
        if nextToken != nil {
            input.NextToken = nextToken
        }
        output, err := svc.ListStreamsWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "kinesis:ListStreams", err)
            break
        }
        for _, summary := range output.StreamSummaries {
            if summary != nil {
                allStreams = append(allStreams, extractStream(summary, region))
            }
        }
        if output.NextToken == nil {
            break
        }
        nextToken = output.NextToken
    }
}
```

### List Streams Then Shards Pattern (Call 2: ListShards)

```go
var allShards []kinesisShard
var lastErr error

for _, region := range types.Regions {
    svc := kinesis.New(sess, &aws.Config{Region: aws.String(region)})

    // Step 1: List all stream names
    var streamNames []string
    var streamToken *string
    for {
        input := &kinesis.ListStreamsInput{}
        if streamToken != nil {
            input.NextToken = streamToken
        }
        output, err := svc.ListStreamsWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "kinesis:ListShards", err)
            break
        }
        for _, summary := range output.StreamSummaries {
            if summary != nil && summary.StreamName != nil {
                streamNames = append(streamNames, *summary.StreamName)
            }
        }
        if output.NextToken == nil {
            break
        }
        streamToken = output.NextToken
    }

    // Step 2: For each stream, list shards
    for _, streamName := range streamNames {
        var shardToken *string
        first := true
        for {
            input := &kinesis.ListShardsInput{}
            if first {
                input.StreamName = aws.String(streamName)
                first = false
            } else {
                input.NextToken = shardToken
            }
            output, err := svc.ListShardsWithContext(ctx, input)
            if err != nil {
                utils.HandleAWSError(false, "kinesis:ListShards", err)
                break
            }
            for _, shard := range output.Shards {
                if shard != nil {
                    allShards = append(allShards, extractShard(shard, streamName, region))
                }
            }
            if output.NextToken == nil {
                break
            }
            shardToken = output.NextToken
        }
    }
}
```

### List Streams Then Consumers Pattern (Call 3: ListStreamConsumers)

```go
var allConsumers []kinesisConsumer
var lastErr error

for _, region := range types.Regions {
    svc := kinesis.New(sess, &aws.Config{Region: aws.String(region)})

    // Step 1: List all stream names and ARNs
    var streams []struct{ Name, ARN string }
    var streamToken *string
    for {
        input := &kinesis.ListStreamsInput{}
        if streamToken != nil {
            input.NextToken = streamToken
        }
        output, err := svc.ListStreamsWithContext(ctx, input)
        if err != nil {
            lastErr = err
            utils.HandleAWSError(false, "kinesis:ListStreamConsumers", err)
            break
        }
        for _, summary := range output.StreamSummaries {
            if summary != nil && summary.StreamARN != nil {
                name := ""
                if summary.StreamName != nil {
                    name = *summary.StreamName
                }
                streams = append(streams, struct{ Name, ARN string }{name, *summary.StreamARN})
            }
        }
        if output.NextToken == nil {
            break
        }
        streamToken = output.NextToken
    }

    // Step 2: For each stream, list consumers
    for _, stream := range streams {
        var consumerToken *string
        for {
            input := &kinesis.ListStreamConsumersInput{
                StreamARN: aws.String(stream.ARN),
            }
            if consumerToken != nil {
                input.NextToken = consumerToken
            }
            output, err := svc.ListStreamConsumersWithContext(ctx, input)
            if err != nil {
                utils.HandleAWSError(false, "kinesis:ListStreamConsumers", err)
                break
            }
            for _, consumer := range output.Consumers {
                if consumer != nil {
                    allConsumers = append(allConsumers, extractConsumer(consumer, stream.Name, region))
                }
            }
            if output.NextToken == nil {
                break
            }
            consumerToken = output.NextToken
        }
    }
}
```

### Nil-Safe Field Extraction Helpers

```go
func extractStream(summary *kinesis.StreamSummary, region string) kinesisStream {
    name := ""
    if summary.StreamName != nil {
        name = *summary.StreamName
    }
    arn := ""
    if summary.StreamARN != nil {
        arn = *summary.StreamARN
    }
    status := ""
    if summary.StreamStatus != nil {
        status = *summary.StreamStatus
    }
    mode := ""
    if summary.StreamModeDetails != nil && summary.StreamModeDetails.StreamMode != nil {
        mode = *summary.StreamModeDetails.StreamMode
    }
    creationTimestamp := ""
    if summary.StreamCreationTimestamp != nil {
        creationTimestamp = summary.StreamCreationTimestamp.Format(time.RFC3339)
    }
    return kinesisStream{
        StreamName:        name,
        StreamARN:         arn,
        StreamStatus:      status,
        StreamMode:        mode,
        CreationTimestamp: creationTimestamp,
        Region:            region,
    }
}

func extractShard(shard *kinesis.Shard, streamName, region string) kinesisShard {
    shardId := ""
    if shard.ShardId != nil {
        shardId = *shard.ShardId
    }
    parentShardId := ""
    if shard.ParentShardId != nil {
        parentShardId = *shard.ParentShardId
    }
    startingHashKey := ""
    endingHashKey := ""
    if shard.HashKeyRange != nil {
        if shard.HashKeyRange.StartingHashKey != nil {
            startingHashKey = *shard.HashKeyRange.StartingHashKey
        }
        if shard.HashKeyRange.EndingHashKey != nil {
            endingHashKey = *shard.HashKeyRange.EndingHashKey
        }
    }
    startingSeqNum := ""
    endingSeqNum := ""
    if shard.SequenceNumberRange != nil {
        if shard.SequenceNumberRange.StartingSequenceNumber != nil {
            startingSeqNum = *shard.SequenceNumberRange.StartingSequenceNumber
        }
        if shard.SequenceNumberRange.EndingSequenceNumber != nil {
            endingSeqNum = *shard.SequenceNumberRange.EndingSequenceNumber
        }
    }
    return kinesisShard{
        ShardId:                shardId,
        StreamName:            streamName,
        ParentShardId:         parentShardId,
        StartingHashKey:       startingHashKey,
        EndingHashKey:         endingHashKey,
        StartingSequenceNumber: startingSeqNum,
        EndingSequenceNumber:  endingSeqNum,
        Region:                region,
    }
}

func extractConsumer(consumer *kinesis.Consumer, streamName, region string) kinesisConsumer {
    name := ""
    if consumer.ConsumerName != nil {
        name = *consumer.ConsumerName
    }
    arn := ""
    if consumer.ConsumerARN != nil {
        arn = *consumer.ConsumerARN
    }
    status := ""
    if consumer.ConsumerStatus != nil {
        status = *consumer.ConsumerStatus
    }
    creationTimestamp := ""
    if consumer.ConsumerCreationTimestamp != nil {
        creationTimestamp = consumer.ConsumerCreationTimestamp.Format(time.RFC3339)
    }
    return kinesisConsumer{
        ConsumerName:      name,
        ConsumerARN:       arn,
        ConsumerStatus:    status,
        StreamName:        streamName,
        CreationTimestamp: creationTimestamp,
        Region:            region,
    }
}
```

### Local Struct Definitions

```go
type kinesisStream struct {
    StreamName        string
    StreamARN         string
    StreamStatus      string
    StreamMode        string
    CreationTimestamp string
    Region            string
}

type kinesisShard struct {
    ShardId                string
    StreamName            string
    ParentShardId         string
    StartingHashKey       string
    EndingHashKey         string
    StartingSequenceNumber string
    EndingSequenceNumber  string
    Region                string
}

type kinesisConsumer struct {
    ConsumerName      string
    ConsumerARN       string
    ConsumerStatus    string
    StreamName        string
    CreationTimestamp string
    Region            string
}
```

### Variable & Naming Conventions

- **Package:** `kinesis` (directory: `cmd/awtest/services/kinesis/`)
- **Exported variable:** `KinesisCalls` (`[]types.AWSService`)
- **AWSService.Name values:** `"kinesis:ListStreams"`, `"kinesis:ListShards"`, `"kinesis:ListStreamConsumers"`
- **ScanResult.ServiceName:** `"Kinesis"` (title case, not an acronym)
- **ScanResult.ResourceType:** `"stream"`, `"shard"`, `"consumer"` (lowercase)
- **ModuleName:** `types.DefaultModuleName` (`"AWTest"`)
- **Local struct prefix:** `kinesis` (matching package name, following `emr`/`cd`/`dc` pattern)
- **SDK import:** `"github.com/aws/aws-sdk-go/service/kinesis"` (same name as local package — handled same as emr/codedeploy/directconnect pattern)

### Registration Order in services.go

Insert alphabetically — `kinesis` comes after `ivsrealtime`, before `kms`:

```go
// In imports (alphabetical):
"github.com/MillerMedia/awtest/cmd/awtest/services/ivsrealtime"
"github.com/MillerMedia/awtest/cmd/awtest/services/kinesis"        // NEW — after ivsrealtime, before kms
"github.com/MillerMedia/awtest/cmd/awtest/services/kms"

// In AllServices():
allServices = append(allServices, ivsrealtime.IvsRealtimeCalls...)
allServices = append(allServices, kinesis.KinesisCalls...)           // NEW — after ivsrealtime, before kms
allServices = append(allServices, kms.KMSCalls...)
```

### Testing Pattern

Follow the EMR/CodeDeploy/DirectConnect test pattern — test Process() functions only with pre-built mock data:

```go
func TestListStreamsProcess(t *testing.T) {
    process := KinesisCalls[0].Process
    // Table-driven tests: valid streams (name, ARN, status, mode, creation timestamp), empty, errors, nil fields, type assertion failure
}

func TestListShardsProcess(t *testing.T) {
    process := KinesisCalls[1].Process
    // Table-driven tests: valid shards (shard ID, stream name, parent shard ID, hash key range, sequence number range), empty, errors, nil fields, type assertion failure
}

func TestListStreamConsumersProcess(t *testing.T) {
    process := KinesisCalls[2].Process
    // Table-driven tests: valid consumers (name, ARN, status, stream name, creation timestamp), empty, errors, nil fields, type assertion failure
}
```

Include extract helper tests with AWS SDK types:
```go
func TestExtractStream(t *testing.T) { ... }
func TestExtractShard(t *testing.T) { ... }
func TestExtractConsumer(t *testing.T) { ... }
```

Use stdlib `testing` only — `t.Fatalf` for hard failures, `t.Errorf` for soft assertions. No testify in service tests.

### Concurrency Compliance

- **DO NOT** import `sync`, `sync/atomic`, or any concurrency primitives in `kinesis/calls.go`
- Service is concurrency-unaware — the worker pool and safeScan wrapper handle all concurrency concerns
- All API calls use `WithContext` variants for timeout/cancellation support
- Error classification (throttle/denied/error) is handled by safeScan, not the service

### Anti-Patterns to Avoid

- **DO NOT** mutate `sess.Config.Region` — use `kinesis.New(sess, &aws.Config{Region: aws.String(region)})` config override pattern
- **DO NOT** use `Marker` — Kinesis uses `NextToken` for pagination (unlike EMR which uses `Marker`)
- **DO NOT** use `StreamNames` field from ListStreams — use `StreamSummaries` which includes status and mode details
- **DO NOT** include `StreamName` in ListShards when paginating with `NextToken` — only set `StreamName` on the FIRST request; subsequent pages use only `NextToken`
- **DO NOT** spawn goroutines inside Call or Process — services must be single-threaded
- **DO NOT** write to stdout — use `utils.PrintResult()` which handles buffering in concurrent mode
- **DO NOT** use generics (Go 1.19 does not support generics)
- **DO NOT** use `sess.Copy()` — use config override in service constructor
- **DO NOT** use `StreamName` for `ListStreamConsumers` — it requires `StreamARN`

### Key Differences from Previous Stories (9.3 EMR, 9.1 CodeDeploy, 9.2 DirectConnect)

1. **NextToken pagination (not Marker):** Kinesis uses `NextToken` for all paginated APIs, unlike EMR which uses `Marker`. Same as CodeDeploy/DirectConnect.
2. **StreamSummaries vs simple names:** `ListStreams` returns `StreamSummaries` with full details (name, ARN, status, mode), so no separate describe call is needed. Unlike EMR which needed `DescribeCluster` per cluster.
3. **ListShards StreamName exclusivity:** When paginating `ListShards`, the `StreamName` parameter must only be set on the first request. When using `NextToken`, `StreamName` must NOT be included. This is a unique quirk of the Kinesis API.
4. **ListStreamConsumers requires ARN:** Unlike most AWS APIs that accept names, `ListStreamConsumers` requires `StreamARN`. The ARN must be collected from `StreamSummary.StreamARN` during stream listing.
5. **Nested field patterns:** `StreamModeDetails.StreamMode` (stream mode) and `HashKeyRange.StartingHashKey`/`SequenceNumberRange.StartingSequenceNumber` (shard details) require multi-level nil checks, similar to EMR's `Status.State` pattern.
6. **Time formatting:** Both `StreamCreationTimestamp` and `ConsumerCreationTimestamp` are `*time.Time` — format with `time.RFC3339` (same as EMR's `CreationDateTime`).
7. **No global resources:** All Kinesis resources are regional — iterate `types.Regions` for all 3 API calls.

### Project Structure Notes

**Files to CREATE:**
```
cmd/awtest/services/kinesis/
+-- calls.go            # Kinesis service implementation (3 AWSService entries)
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
cmd/awtest/services/emr/calls.go              # Reference implementation (nested patterns, extract helpers)
cmd/awtest/services/emr/calls_test.go         # Reference test pattern (extract helper tests)
cmd/awtest/services/codedeploy/calls.go       # Reference implementation (NextToken pagination)
cmd/awtest/services/codedeploy/calls_test.go  # Reference test pattern
go.mod                                        # AWS SDK already includes kinesis package (needs vendoring)
```

**Vendor directory to POPULATE:**
```
vendor/github.com/aws/aws-sdk-go/service/kinesis/  # Run go mod vendor to populate
```

### Previous Story Intelligence

**From Story 9.3 (EMR — most recent completed story):**
- All Call functions iterate `types.Regions` and use config override pattern `service.New(sess, &aws.Config{Region: ...})`
- Use local structs for call results (define local types at package level)
- Per-region errors: `break` pagination loop, don't abort entire scan
- Extract helper functions for nil-safe extraction — directly applicable
- Nested struct extraction (`Status.State`) — same pattern needed for `StreamModeDetails.StreamMode`, `HashKeyRange.StartingHashKey`, `SequenceNumberRange.StartingSequenceNumber`
- Time formatting with `time.RFC3339` — same pattern needed for `StreamCreationTimestamp` and `ConsumerCreationTimestamp`
- Type assertion failure: handle with `utils.HandleAWSError` and return empty results
- Process via `EMRCalls[N].Process` in tests -> apply as `KinesisCalls[N].Process`
- Error result pattern: `return []types.ScanResult{{ServiceName: "Kinesis", MethodName: "kinesis:ListStreams", Error: err, Timestamp: time.Now()}}`
- 24 tests across 6 test functions in EMR

**From Story 9.1 (CodeDeploy — NextToken pagination reference):**
- NextToken pagination pattern — directly applicable to all 3 Kinesis calls
- Nested API calls (list apps, then per-app list groups) — similar to Kinesis ListShards/ListStreamConsumers
- 23 tests across 7 test functions in CodeDeploy

**From Code Review Findings (Stories 7.1, 7.2):**
- [HIGH] Always use config override for region (race condition prevention)
- [HIGH] Include all relevant fields in Details map
- [HIGH] Always add pagination from the start — applies to all 3 Kinesis calls
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
- Expected commit message: `"Add Kinesis enumeration with 3 API calls (Story 9.4)"`

### FRs Covered

- **FR111:** System enumerates Kinesis streams, shard details, and consumer applications

### NFRs Addressed

- **NFR53:** New service follows existing AWSService interface pattern — no changes to core enumeration engine
- **NFR54:** New service registers in AllServices() maintaining alphabetical ordering
- **NFR57:** New service requires no concurrency-specific code — worker pool handles parallelism transparently

### References

- [Source: epics-phase2.md#Story 4.4: Kinesis Enumeration] — BDD acceptance criteria
- [Source: prd-phase2.md#FR111] — Kinesis enumeration requirement
- [Source: architecture-phase2.md#Service Boundary] — Services implement Call/Process, registered alphabetically
- [Source: architecture-phase2.md#Concurrency Patterns] — Services must remain concurrency-unaware
- [Source: architecture-phase2.md#Naming Patterns] — Service file naming: `services/<service_name>/calls.go`
- [Source: cmd/awtest/services/emr/calls.go] — Reference implementation (nested patterns, extract helpers, time formatting)
- [Source: cmd/awtest/services/emr/calls_test.go] — Reference test pattern (extract helper tests)
- [Source: cmd/awtest/services/codedeploy/calls.go] — Reference implementation (NextToken pagination)
- [Source: cmd/awtest/services/codedeploy/calls_test.go] — Reference test pattern
- [Source: cmd/awtest/services/services.go] — AllServices() registration point (kinesis goes after ivsrealtime, before kms)
- [Source: cmd/awtest/types/types.go] — AWSService struct, ScanResult, Regions
- [Source: cmd/awtest/utils/output.go] — PrintResult, HandleAWSError, ColorizeItem
- [Source: go.mod] — aws-sdk-go v1.44.266 (includes kinesis package, needs vendoring)
- [Source: 9-3-emr-cluster-enumeration.md] — Previous story (nested structs, time formatting, extract helpers)
- [Source: 9-1-codedeploy-enumeration.md] — Reference story (NextToken pagination, nested API calls)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered.

### Completion Notes List

- Implemented 3 Kinesis AWSService entries: ListStreams, ListShards, ListStreamConsumers
- All 3 API calls use NextToken pagination and config override pattern for region iteration
- ListShards correctly excludes StreamName on paginated requests (only set on first request)
- ListStreamConsumers correctly uses StreamARN (not StreamName)
- 3 extract helpers (extractStream, extractShard, extractConsumer) with nil-safe pointer dereferencing including nested structs (StreamModeDetails, HashKeyRange, SequenceNumberRange)
- 24 tests across 6 test functions: 3 Process tests + 3 extract helper tests
- All tests cover: valid results, empty results, access denied errors, nil fields, type assertion failures
- No sync primitives used — concurrency-unaware per NFR57
- Registered in services.go alphabetically after ivsrealtime, before kms
- Full test suite passes with no regressions, race detector clean

### File List

- cmd/awtest/services/kinesis/calls.go (NEW)
- cmd/awtest/services/kinesis/calls_test.go (NEW)
- cmd/awtest/services/services.go (MODIFIED)
- vendor/github.com/aws/aws-sdk-go/service/kinesis/ (VENDORED)

### Change Log

- 2026-03-13: Implemented Story 9.4 — Kinesis enumeration with 3 API calls (ListStreams, ListShards, ListStreamConsumers), 24 unit tests, registered in AllServices()
