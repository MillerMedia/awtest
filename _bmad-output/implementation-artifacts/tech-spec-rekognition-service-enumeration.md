---
title: 'Rekognition Service Enumeration'
slug: 'rekognition-service-enumeration'
created: '2026-03-03'
status: 'complete'
stepsCompleted: [1, 2, 3, 4]
tech_stack: ['Go 1.19', 'aws-sdk-go v1.44.266', 'service/rekognition']
files_to_modify: ['cmd/awtest/services/rekognition/calls.go (new)', 'cmd/awtest/services/services.go (modify)']
code_patterns: ['AWSService slice pattern', 'Call/Process function pair', 'ScanResult return', 'HandleAWSError for errors', 'PrintResult for console output', 'ModuleName: types.DefaultModuleName']
test_patterns: ['No unit tests for service calls — all require live AWS', 'Manual verification via go build + binary execution']
---

# Tech-Spec: Rekognition Service Enumeration

**Created:** 2026-03-03

## Overview

### Problem Statement

awtest doesn't detect AWS Rekognition resources. Rekognition has enumerable top-level resources (face collections, stream processors, custom label projects) that should be visible during security scans.

### Solution

Add a `rekognition` service package following the established service pattern with 3 top-level enumeration calls: `ListCollections`, `ListStreamProcessors`, `DescribeProjects`. Register in `services.go`.

**Note:** `ListMediaAnalysisJobs` was dropped — it doesn't exist in `aws-sdk-go` v1.44.266.

### Scope

**In Scope:**
- `rekognition:ListCollections` — enumerate face collections (returns `CollectionIds []*string`)
- `rekognition:ListStreamProcessors` — enumerate video stream processors (returns `StreamProcessors []*StreamProcessor` with Name, Status)
- `rekognition:DescribeProjects` — enumerate custom label projects (returns `ProjectDescriptions []*ProjectDescription` with ProjectArn, Status)
- Registration in `services.go` `AllServices()` function
- Backward-compatible console output via `utils.PrintResult()`
- `[]types.ScanResult` return from all `Process` functions

**Out of Scope:**
- Second-level enumeration (faces in collections, datasets in projects, project versions, etc.)
- `ListMediaAnalysisJobs` (not available in SDK v1.44.266)
- Epic 2 integration or sprint tracking
- New external dependencies
- Unit tests for service calls (no existing services have unit tests — they all require live AWS)

## Context for Development

### Codebase Patterns

- Each service is a package under `cmd/awtest/services/<name>/`
- Service calls defined as `var <Name>Calls = []types.AWSService{...}` slice
- Each `AWSService` entry has: `Name` (string like `"glacier:ListVaults"`), `Call` (func creating SDK client + API call), `Process` (func handling errors + building results), `ModuleName` (use `types.DefaultModuleName`)
- `Call` pattern: `svc := <service>.New(sess)` → call API → return output (or specific field), err
- `Process` pattern: check err → `utils.HandleAWSError()` + return error ScanResult → iterate results → build `types.ScanResult` per resource → `utils.PrintResult()` for console → return slice
- Registered in `services.go` via `allServices = append(allServices, <pkg>.<Name>Calls...)`
- **Reference implementation:** `glacier/calls.go` is the cleanest single-call example

### Files to Reference

| File | Purpose |
| ---- | ------- |
| `cmd/awtest/services/glacier/calls.go` | Cleanest reference — single List call with simple output |
| `cmd/awtest/services/services.go` | Service registration — add import + append line |
| `cmd/awtest/types/types.go` | `AWSService` struct (Name, Call, Process, ModuleName) and `ScanResult` struct |
| `cmd/awtest/utils/output.go` | `HandleAWSError(debug, callName, err)` and `PrintResult(debug, "", callName, msg, nil)` |

### Technical Decisions

- Uses `aws-sdk-go` v1 (already in go.mod) — `service/rekognition` package included in SDK
- No new go.mod dependencies required
- Follows exact same Call/Process pattern as all 34 existing services
- Console output uses `utils.PrintResult()` for backward compatibility
- `ListCollections` returns bare string IDs (no struct) — use collection ID as ResourceName, "collection" as ResourceType
- `DescribeProjects` used instead of a separate `ListProjects` — it returns project details directly
- `ListStreamProcessors` returns `StreamProcessor` struct with Name and Status

### SDK Output Shapes (Verified)

```go
// ListCollections
CollectionIds []*string  // Just string IDs

// ListStreamProcessors
StreamProcessors []*StreamProcessor {
    Name   *string
    Status *string  // enum: StreamProcessorStatus
}

// DescribeProjects
ProjectDescriptions []*ProjectDescription {
    ProjectArn        *string
    Status            *string  // enum: ProjectStatus
    CreationTimestamp *time.Time
}
```

## Implementation Plan

### Tasks

- [x] Task 1: Create `cmd/awtest/services/rekognition/calls.go`
  - File: `cmd/awtest/services/rekognition/calls.go` (new)
  - Action: Create package `rekognition` with `var RekognitionCalls = []types.AWSService{...}` containing 3 entries:
    1. `rekognition:ListCollections` — Call: `svc.ListCollections(&rekognition.ListCollectionsInput{})`, return `output.CollectionIds, err`. Process: iterate `[]*string`, ResourceType `"collection"`, ResourceName `*id`, print `"Rekognition collection: <id>"`
    2. `rekognition:ListStreamProcessors` — Call: `svc.ListStreamProcessors(&rekognition.ListStreamProcessorsInput{})`, return `output.StreamProcessors, err`. Process: iterate `[]*StreamProcessor`, ResourceType `"stream-processor"`, ResourceName `*sp.Name`, Details `{"status": *sp.Status}`, print `"Rekognition stream processor: <name>"`
    3. `rekognition:DescribeProjects` — Call: `svc.DescribeProjects(&rekognition.DescribeProjectsInput{})`, return `output.ProjectDescriptions, err`. Process: iterate `[]*ProjectDescription`, ResourceType `"project"`, ResourceName `*proj.ProjectArn`, Details `{"status": *proj.Status}`, print `"Rekognition project: <arn>"`
  - Notes: Follow `glacier/calls.go` pattern exactly. Use `types.DefaultModuleName` for ModuleName. Import `"time"` for `time.Now()` in ScanResult.Timestamp. Nil-check all pointer dereferences.

- [x] Task 2: Register Rekognition in `services.go`
  - File: `cmd/awtest/services/services.go` (modify)
  - Action: Add import `"github.com/MillerMedia/awtest/cmd/awtest/services/rekognition"` and add `allServices = append(allServices, rekognition.RekognitionCalls...)` in `AllServices()`. Place alphabetically among existing entries.

- [x] Task 3: Build and vet verification
  - Action: Run `go build ./cmd/awtest` and `go vet ./cmd/awtest/...` to confirm clean compilation
  - Notes: No new go.mod entries should be needed — `service/rekognition` is part of the existing `aws-sdk-go` dependency

- [x] Task 4: Live verification with AWS credentials
  - Action: Run `./awtest` and confirm 3 new Rekognition lines appear in output (likely "Access denied" or resource listings depending on permissions)
  - Notes: The `rekognitionML` IAM user may have actual Rekognition access — check for real resource output

### Acceptance Criteria

- [ ] AC1: Given `cmd/awtest/services/rekognition/calls.go` exists, when inspected, then it contains a `RekognitionCalls` slice with exactly 3 `types.AWSService` entries for ListCollections, ListStreamProcessors, and DescribeProjects
- [ ] AC2: Given the binary is built, when `./awtest` is run with valid AWS credentials, then output includes lines for `rekognition:ListCollections`, `rekognition:ListStreamProcessors`, and `rekognition:DescribeProjects`
- [ ] AC3: Given an AWS account with no Rekognition resources, when scanned, then each call returns an error ScanResult with the error and does not crash
- [ ] AC4: Given an AWS account with Rekognition collections, when `ListCollections` succeeds, then each collection ID appears as a ScanResult with ServiceName "Rekognition", ResourceType "collection", and the collection ID as ResourceName
- [ ] AC5: Given `services.go`, when inspected, then `rekognition.RekognitionCalls` is registered in `AllServices()` and placed alphabetically
- [ ] AC6: Given the codebase, when `go build ./cmd/awtest` and `go vet ./cmd/awtest/...` are run, then both pass with zero errors

## Additional Context

### Dependencies

None new. Uses existing `github.com/aws/aws-sdk-go/service/rekognition` from the v1 SDK already in go.mod.

### Testing Strategy

- No unit tests (consistent with all 34 existing services — they require live AWS)
- Manual verification: `go build` + `go vet` + run binary with AWS credentials
- Confirm 3 new Rekognition calls appear in console output
- Verify error handling with access-denied responses

### Notes

- Standalone quick feature, not part of Epic 2 roadmap
- Motivated by test credentials being on a `rekognitionML` IAM user
- `DescribeProjects` is used instead of a separate `ListProjects` because it returns project details directly and serves as both list and describe
- `ListMediaAnalysisJobs` dropped because it's not in aws-sdk-go v1.44.266
- Risk: Minimal — follows an established, well-tested pattern with no new dependencies or architectural changes
