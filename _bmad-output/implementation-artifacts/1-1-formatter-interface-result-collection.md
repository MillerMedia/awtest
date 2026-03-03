# Story 1.1: Formatter Interface & Result Collection

Status: in-progress

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a developer implementing awtest enhancements,
I want a clean formatter interface and result collection system,
so that scan results can be formatted in multiple output formats without coupling service logic to presentation.

## Acceptance Criteria

**Given** the existing service enumeration architecture
**When** implementing the formatter system foundation
**Then:**

1. **AC1:** Create the OutputFormatter interface in `cmd/awtest/formatters/output_formatter.go` with:
   - `Format(results []ScanResult) (string, error)` method
   - `FileExtension() string` method

2. **AC2:** Define the ScanResult struct in `cmd/awtest/types/types.go` with fields:
   - ServiceName string
   - MethodName string
   - ResourceType string
   - ResourceName string
   - Details map[string]interface{}
   - Error error
   - Timestamp time.Time

3. **AC3:** Modify the `main.go` service enumeration loop to collect results in a `[]ScanResult` slice instead of immediate printing

4. **AC4:** Update service `Process()` methods to return `ScanResult` instead of printing directly

5. **AC5:** Write unit tests for result collection logic achieving >70% coverage

6. **AC6:** Ensure backward compatibility - existing service behavior is preserved

7. **AC7:** Verify compilation succeeds with `go build ./cmd/awtest`

## Tasks / Subtasks

- [ ] Create formatters directory and OutputFormatter interface (AC: 1)
  - [ ] Create cmd/awtest/formatters directory
  - [ ] Implement output_formatter.go with interface definition
  - [ ] Write interface documentation

- [ ] Define ScanResult struct in types package (AC: 2)
  - [ ] Add ScanResult struct to types/types.go
  - [ ] Import time package for Timestamp field
  - [ ] Add helper methods if needed (e.g., HasError())

- [ ] Refactor main.go for result collection (AC: 3)
  - [ ] Add results []ScanResult slice before service loop
  - [ ] Modify service loop to collect results
  - [ ] Preserve existing behavior temporarily

- [ ] Update Process() signature and implementations (AC: 4)
  - [ ] Update types.AWSService.Process signature
  - [ ] Create helper function to convert current printing to ScanResult
  - [ ] Update all 34 existing services to use new pattern
  - [ ] Test each service update individually

- [ ] Implement backward compatibility layer (AC: 6)
  - [ ] Create temporary compatibility function
  - [ ] Ensure output looks identical to current version
  - [ ] Verify no functional changes to user experience

- [ ] Write comprehensive unit tests (AC: 5)
  - [ ] Test ScanResult struct creation
  - [ ] Test result collection in main loop
  - [ ] Test Process() method return values
  - [ ] Achieve >70% code coverage

- [ ] Verification and validation (AC: 7)
  - [ ] Run go build ./cmd/awtest
  - [ ] Run all existing tests
  - [ ] Manual testing with real AWS credentials
  - [ ] Verify output matches current behavior exactly

## Dev Notes

### 🎯 CRITICAL CONTEXT: This Story is the Foundation for Epic 1

This story creates the infrastructure that ALL other Epic 1 stories depend on. You are NOT implementing output formats yet - you are creating the INTERFACE and COLLECTION SYSTEM that makes multiple output formats possible. Think of this as building the foundation before the house.

**What This Story Achieves:**
- Decouples service enumeration logic from output presentation
- Enables future formatters (JSON, YAML, CSV, Table) to operate on collected results
- Maintains 100% backward compatibility with current text output
- Establishes the pattern for all 34 existing services to follow

**What This Story Does NOT Do:**
- Does NOT implement any actual formatters (that's Stories 1.2-1.5)
- Does NOT change user-visible output (must look identical to v0.3.0)
- Does NOT add new flags or command-line options (that's Story 1.6)

### 🔥 Critical Implementation Guardrails

**AVOID THESE COMMON MISTAKES:**

1. **DON'T implement formatters yet** - This story ONLY creates the interface and struct. Stories 1.2-1.5 implement the actual formatters.

2. **DON'T break existing output** - The current text output must remain EXACTLY the same. Users should see zero visual difference after this story.

3. **DON'T modify all 34 services at once** - Start with 1-2 services as proof of concept, verify they work, then systematically update the rest.

4. **DON'T forget the transition period** - You may need a hybrid approach where Process() BOTH returns ScanResult AND prints (temporarily) to maintain backward compatibility during refactoring.

5. **DON'T overcomplicate the ScanResult struct** - Keep it simple. The Details map can hold anything service-specific. Don't add unnecessary fields.

### 🏗️ Architecture Patterns to Follow

**Current Architecture (Before This Story):**
```go
// main.go - Current flow
for _, service := range services.AllServices() {
    output, err := service.Call(sess)
    service.Process(output, err, debug)  // Prints immediately
}

// Service Process() - Current pattern
Process: func(output interface{}, err error, debug bool) error {
    if err != nil {
        return utils.HandleAWSError(debug, "s3:ListBuckets", err)
    }
    // ... process output ...
    utils.PrintResult(debug, "", "s3:ListBuckets", "S3 bucket: example", nil)
    return nil
}
```

**Target Architecture (After This Story):**
```go
// main.go - New flow
results := []types.ScanResult{}
for _, service := range services.AllServices() {
    output, err := service.Call(sess)
    result := service.Process(output, err, debug)
    results = append(results, result...)
}
// Print results using current text format (for backward compatibility)
for _, result := range results {
    printResultAsText(result)  // Temporary function that mimics utils.PrintResult
}

// Service Process() - New signature
Process: func(output interface{}, err error, debug bool) []types.ScanResult {
    var results []types.ScanResult
    if err != nil {
        return []types.ScanResult{
            {
                ServiceName: "S3",
                MethodName: "s3:ListBuckets",
                Error: err,
                Timestamp: time.Now(),
            },
        }
    }
    // ... process output ...
    results = append(results, types.ScanResult{
        ServiceName: "S3",
        MethodName: "s3:ListBuckets",
        ResourceType: "bucket",
        ResourceName: "example-bucket",
        Details: map[string]interface{}{"region": "us-west-2"},
        Timestamp: time.Now(),
    })
    return results
}
```

**Interface Design Pattern:**
```go
// cmd/awtest/formatters/output_formatter.go
package formatters

import "github.com/MillerMedia/awtest/cmd/awtest/types"

// OutputFormatter defines the interface for all output formatters
type OutputFormatter interface {
    // Format takes scan results and returns formatted output string
    Format(results []types.ScanResult) (string, error)

    // FileExtension returns the file extension for this format (e.g., "json", "yaml", "txt")
    FileExtension() string
}
```

### 📁 Project Structure & File Organization

**Current Structure:**
```
cmd/awtest/
├── main.go                    # Entry point, credential handling, service loop
├── types/
│   └── types.go              # AWSService struct, error types
├── services/
│   ├── services.go           # AllServices() registry
│   ├── s3/calls.go          # S3 service implementation
│   ├── ec2/calls.go         # EC2 service implementation
│   └── ... (32 more services)
└── utils/
    └── output.go             # PrintResult(), HandleAWSError()
```

**New Structure (After This Story):**
```
cmd/awtest/
├── main.go                    # MODIFIED: Add result collection
├── types/
│   └── types.go              # MODIFIED: Add ScanResult struct, update Process signature
├── formatters/               # NEW DIRECTORY
│   └── output_formatter.go  # NEW: OutputFormatter interface
├── services/
│   ├── services.go           # UNCHANGED
│   ├── s3/calls.go          # MODIFIED: Update Process to return []ScanResult
│   ├── ec2/calls.go         # MODIFIED: Update Process to return []ScanResult
│   └── ... (all 34 services must be updated)
└── utils/
    └── output.go             # UNCHANGED (keep for backward compatibility)
```

### 🔧 Technical Requirements

**Go Version:** 1.19 (existing project standard)

**Dependencies:**
- NO new dependencies required for this story
- Uses existing `github.com/aws/aws-sdk-go v1.44.266`
- Uses existing `github.com/logrusorgru/aurora v2.0.3+incompatible` (for colorized output)

**Package Imports:**
- `time` package needed for ScanResult.Timestamp
- No other new imports required

**Type System Changes:**
```go
// types/types.go - ADD this struct
type ScanResult struct {
    ServiceName  string                 // e.g., "S3", "EC2", "IAM"
    MethodName   string                 // e.g., "s3:ListBuckets", "ec2:DescribeInstances"
    ResourceType string                 // e.g., "bucket", "instance", "user"
    ResourceName string                 // e.g., "my-bucket", "i-1234567890abcdef0"
    Details      map[string]interface{} // Service-specific details (region, count, metadata)
    Error        error                  // nil if successful, error if failed
    Timestamp    time.Time              // When this result was collected
}

// Helper method
func (sr ScanResult) HasError() bool {
    return sr.Error != nil
}

// types/types.go - UPDATE this signature
type AWSService struct {
    Name       string
    Call       func(*session.Session) (interface{}, error)
    Process    func(interface{}, error, bool) []ScanResult  // CHANGED: returns []ScanResult instead of error
    ModuleName string
}
```

### 🧪 Testing Requirements

**Test Coverage Target:** >70% (per AC5)

**Test Files to Create:**
1. `cmd/awtest/types/types_test.go` - Test ScanResult struct
2. `cmd/awtest/formatters/output_formatter_test.go` - Test interface definition (minimal, just verify compilation)
3. Integration test in `cmd/awtest/main_test.go` - Test result collection flow

**Key Test Cases:**
```go
// Test ScanResult creation
func TestScanResult_Creation(t *testing.T) {
    result := types.ScanResult{
        ServiceName:  "S3",
        MethodName:   "s3:ListBuckets",
        ResourceType: "bucket",
        ResourceName: "test-bucket",
        Details:      map[string]interface{}{"region": "us-east-1"},
        Timestamp:    time.Now(),
    }
    if result.HasError() {
        t.Error("Expected no error")
    }
}

// Test ScanResult with error
func TestScanResult_WithError(t *testing.T) {
    result := types.ScanResult{
        ServiceName: "S3",
        MethodName:  "s3:ListBuckets",
        Error:       errors.New("access denied"),
        Timestamp:   time.Now(),
    }
    if !result.HasError() {
        t.Error("Expected error to be present")
    }
}

// Test result collection (integration test)
func TestMainLoop_ResultCollection(t *testing.T) {
    // Mock AWS session
    // Run service loop
    // Verify results are collected in slice
    // Verify no results are lost
}
```

**Test Execution:**
```bash
# Run all tests
go test ./cmd/awtest/...

# Check coverage
go test -cover ./cmd/awtest/...

# Generate coverage report
go test -coverprofile=coverage.out ./cmd/awtest/...
go tool cover -html=coverage.out
```

### 🎨 Code Style & Naming Conventions

**From Architecture Document:**
- PascalCase for exported identifiers (OutputFormatter, ScanResult)
- camelCase for unexported identifiers (printResultAsText)
- Lowercase filenames with underscores (output_formatter.go, types_test.go)
- Co-located tests: *_test.go files next to implementation
- Package names: lowercase, no underscores (formatters, not output_formatters)

**Interface Naming:**
- Follow Go conventions: OutputFormatter (not IOutputFormatter or OutputFormatterInterface)
- Methods are imperative verbs: Format(), FileExtension()

**Struct Field Naming:**
- Exported fields use PascalCase
- Use meaningful names: ServiceName not Service, ResourceType not Type

### 🔍 Git Analysis & Recent Work Patterns

**Recent Commits (Last 10):**
```
2e21972 Add ECS service support for testing permissions
4f79e56 Update README.md
221eaf7 Version 0.3.0
b160ef0 Add user data, tags and elastic IP output for EC2 instances
39caa97 Add describe log streams if log group read is enabled
a0610cb Add function configuration output
8782c66 Code to use AWS_PROFILE if no keys are set in command
e311a59 Merge pull request #3 from MillerMedia/session-token-support
2d40794 Added abbreviated flag for session token
8d90e2d Initial support for AWS Session Token
```

**Key Learnings from Recent Commits:**

1. **Service Addition Pattern (2e21972):** ECS service was added following the existing pattern - new directory under services/, calls.go file, registered in services.go AllServices()

2. **Incremental Output Enhancement (b160ef0, 39caa97, a0610cb):** Multiple commits show gradual enhancement of output for existing services - add more details to what's printed, more granular information

3. **Credential Handling Evolution (8782c66, 2d40794, 8d90e2d):** Three commits focused on credential handling improvements - AWS_PROFILE support, session token support. Shows this is a brownfield project being actively enhanced.

4. **Version Tracking (221eaf7):** v0.3.0 release - your changes will be part of the next version

**Pattern to Follow:**
- Small, incremental changes are preferred
- Each service change is self-contained
- Documentation updates accompany feature changes
- Backward compatibility is maintained (AWS_PROFILE was added, not replaced)

### 🛡️ Backward Compatibility Strategy

**CRITICAL:** This story must maintain 100% backward compatibility. Users upgrading from v0.3.0 should see IDENTICAL output.

**Compatibility Approach:**

1. **Phase 1: Add ScanResult without changing behavior**
   - Add ScanResult struct to types.go
   - Create OutputFormatter interface
   - Don't modify any existing code yet

2. **Phase 2: Dual-mode Process() methods**
   - Update Process() signature to return []ScanResult
   - KEEP the utils.PrintResult() calls inside Process() temporarily
   - Return ScanResult AND print - both happen

3. **Phase 3: Gradual migration**
   - Update main.go to collect results
   - Add temporary printResultAsText() helper that formats ScanResult back to current format
   - Verify output is identical

4. **Phase 4: Clean up (Story 1.6)**
   - Remove duplicate printing from Process() methods
   - Rely solely on formatter system
   - This happens in Story 1.6 when -format flag is added

**Temporary Helper Function Pattern:**
```go
// main.go - Temporary backward compatibility function
func printResultAsText(result types.ScanResult, debug bool) {
    if result.HasError() {
        utils.HandleAWSError(debug, result.MethodName, result.Error)
        return
    }

    // Format message to match current output style
    message := fmt.Sprintf("%s: %s", result.ResourceType, result.ResourceName)

    // Add details if present
    if len(result.Details) > 0 {
        // Format details to match current style
    }

    utils.PrintResult(debug, "", result.MethodName, message, nil)
}
```

### 📚 Architecture Compliance Requirements

**From Architecture Document - Architectural Boundaries:**

1. **AWS SDK Boundary**
   - ✅ Services own AWS API calls - UNCHANGED
   - ✅ main.go owns session creation - UNCHANGED
   - 📝 Process() now returns data instead of printing - NEW BOUNDARY

2. **Output Formatting Boundary**
   - ✅ Services collect results - NEW: Services return ScanResult structs
   - ✅ Formatters handle presentation - NEW: OutputFormatter interface defines this
   - 🎯 Clear separation: services COLLECT, formatters PRESENT

3. **Error Handling Boundary**
   - ✅ Services detect errors - UNCHANGED
   - ✅ utils/output.go classifies and reports - KEEP for backward compatibility
   - 📝 Errors also stored in ScanResult.Error - NEW

4. **Configuration Boundary**
   - ✅ main.go owns flag parsing - UNCHANGED
   - ✅ service_filter.go owns filtering logic - UNCHANGED (future story)

5. **Testing Boundary**
   - ✅ Each package tests its own functionality - FOLLOW THIS
   - ✅ No cross-package test dependencies - ENFORCE THIS

**Pattern Compliance:**
- OutputFormatter interface follows Go interface conventions
- ScanResult is a plain data struct (no methods except helper)
- Clear separation of concerns: data collection vs presentation

### 🗂️ File Structure Requirements

**Files to CREATE:**
```
cmd/awtest/formatters/output_formatter.go  # OutputFormatter interface
cmd/awtest/formatters/output_formatter_test.go  # Interface tests (minimal)
cmd/awtest/types/types_test.go  # ScanResult tests
```

**Files to MODIFY:**
```
cmd/awtest/types/types.go  # Add ScanResult struct, update Process signature
cmd/awtest/main.go  # Add result collection loop
cmd/awtest/services/s3/calls.go  # Update Process (example service)
cmd/awtest/services/ec2/calls.go  # Update Process (example service)
... (all 34 services eventually)
```

**Files to LEAVE UNCHANGED:**
```
cmd/awtest/utils/output.go  # Keep for backward compatibility
cmd/awtest/services/services.go  # AllServices() registration unchanged
go.mod  # No new dependencies
go.sum  # No new dependencies
```

### 🚦 Implementation Strategy

**Recommended Approach (Step-by-Step):**

1. **Step 1: Create Foundation (Low Risk)**
   - Create `formatters/` directory
   - Create `output_formatter.go` with interface
   - Add `ScanResult` struct to `types/types.go`
   - Compile and verify no breakage: `go build ./cmd/awtest`

2. **Step 2: Update Service Signature (Medium Risk)**
   - Change `AWSService.Process` signature in `types/types.go`
   - This will cause compilation errors in all 34 services - EXPECTED
   - Fix compilation by updating Process return type

3. **Step 3: Proof of Concept (2 Services)**
   - Update S3 service Process() to return []ScanResult
   - Update EC2 service Process() to return []ScanResult
   - Keep utils.PrintResult() calls inside Process() for now (dual mode)
   - Test manually to verify output is identical

4. **Step 4: Systematic Service Updates**
   - Create a checklist of all 34 services
   - Update each service's Process() method one at a time
   - Test after each update
   - Maintain dual mode (return ScanResult AND print)

5. **Step 5: Main Loop Update**
   - Modify main.go to collect results in []ScanResult slice
   - Add temporary printResultAsText() function
   - Verify output is still identical

6. **Step 6: Testing & Validation**
   - Write unit tests for ScanResult
   - Write integration tests for result collection
   - Achieve >70% coverage
   - Manual testing with real AWS credentials

7. **Step 7: Final Verification**
   - Compare output before/after side-by-side
   - Test with different credential types (explicit keys, AWS_PROFILE, STS)
   - Test error scenarios (access denied, invalid credentials)
   - Ensure version v0.3.0 behavior is preserved

### 🔗 References & Source Materials

**Epic Requirements:**
- [Source: _bmad-output/planning-artifacts/epics.md#Epic 1: Output Format System]
- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.1: Formatter Interface & Result Collection]

**Architecture Decisions:**
- [Source: _bmad-output/planning-artifacts/architecture.md#Architectural Baseline]
- [Source: _bmad-output/planning-artifacts/architecture.md#Project Structure]
- [Source: _bmad-output/planning-artifacts/architecture.md#Output Formatting Boundary]

**PRD Requirements:**
- [Source: _bmad-output/planning-artifacts/prd.md#FR42: Users receive scan results in human-readable text format by default]
- [Source: _bmad-output/planning-artifacts/prd.md#FR43: Users can export scan results in structured JSON format]
- [Source: _bmad-output/planning-artifacts/prd.md#NFR24: Codebase follows Go standard project layout]

**Existing Codebase:**
- [Source: cmd/awtest/main.go:130-138] - Current service enumeration loop
- [Source: cmd/awtest/types/types.go:17-22] - AWSService struct definition
- [Source: cmd/awtest/utils/output.go:51-61] - PrintResult function
- [Source: cmd/awtest/services/s3/calls.go:12-75] - Example service implementation

**Git History:**
- [Commit: 2e21972] Add ECS service support - shows service addition pattern
- [Commit: 221eaf7] Version 0.3.0 - current baseline version
- [Commit: b160ef0] Add user data, tags and elastic IP output - output enhancement pattern

### 🎓 Knowledge Transfer: Understanding the Current System

**How Services Work Today:**

Each service in `cmd/awtest/services/<service>/calls.go` defines a slice of `AWSService` structs. Example from S3:

```go
var S3Calls = []types.AWSService{
    {
        Name: "s3:ListBuckets",
        Call: func(sess *session.Session) (interface{}, error) {
            // AWS SDK call happens here
            svc := s3.New(sess)
            output, err := svc.ListBuckets(&s3.ListBucketsInput{})
            return map[string]interface{}{
                "output": output,
                "sess":   sess,
            }, err
        },
        Process: func(output interface{}, err error, debug bool) error {
            // Error handling
            if err != nil {
                return utils.HandleAWSError(debug, "s3:ListBuckets", err)
            }
            // Extract output
            outputMap, _ := output.(map[string]interface{})
            s3Output, _ := outputMap["output"].(*s3.ListBucketsOutput)

            // Process each bucket and PRINT immediately
            for _, bucket := range s3Output.Buckets {
                utils.PrintResult(debug, "", "s3:ListBuckets",
                    fmt.Sprintf("S3 bucket: %s", *bucket.Name), nil)
            }
            return nil
        },
        ModuleName: types.DefaultModuleName,
    },
}
```

**The Problem This Story Solves:**

Right now, services PRINT results immediately. This makes it impossible to:
- Export to JSON (results are already printed to stdout)
- Save to a file (no data structure to serialize)
- Format differently (output format is hardcoded)

**The Solution This Story Creates:**

After this story, services RETURN result data:

```go
Process: func(output interface{}, err error, debug bool) []types.ScanResult {
    var results []types.ScanResult

    if err != nil {
        return []types.ScanResult{{
            ServiceName: "S3",
            MethodName:  "s3:ListBuckets",
            Error:       err,
            Timestamp:   time.Now(),
        }}
    }

    outputMap, _ := output.(map[string]interface{})
    s3Output, _ := outputMap["output"].(*s3.ListBucketsOutput)

    for _, bucket := range s3Output.Buckets {
        results = append(results, types.ScanResult{
            ServiceName:  "S3",
            MethodName:   "s3:ListBuckets",
            ResourceType: "bucket",
            ResourceName: *bucket.Name,
            Details:      map[string]interface{}{},
            Timestamp:    time.Now(),
        })
    }

    return results
}
```

Now main.go can collect all results, then pass them to ANY formatter (text, JSON, YAML, CSV).

### Project Structure Notes

**Alignment with Go Standard Project Layout:**
✅ Follows `cmd/<app>/` pattern for main applications
✅ Uses package-per-directory convention
✅ Co-locates tests with implementation (*_test.go)
✅ Clear separation of concerns (services, types, utils, formatters)

**New formatters/ Directory:**
- Aligns with Go convention of feature-based packages
- Contains ONLY the OutputFormatter interface (for now)
- Future stories will add json_formatter.go, yaml_formatter.go, etc.

**No Conflicts Detected:**
- No existing formatters/ directory
- No conflicts with current package structure
- ScanResult name doesn't conflict with existing types

### ⚠️ Critical Warnings & Edge Cases

**Edge Case 1: Services That Return Multiple Results**
- S3 service lists multiple buckets, then for each bucket lists objects
- Each bucket is a separate ScanResult
- Each object count is a separate ScanResult
- Don't try to aggregate - return individual results

**Edge Case 2: Services That Return No Results**
- When a service has no resources, return empty []ScanResult, not nil
- Example: `return []types.ScanResult{}` not `return nil`

**Edge Case 3: Error Handling**
- If a service call fails (access denied), return ONE ScanResult with Error field set
- Don't return empty slice on error - that loses the error information
- Example: `return []types.ScanResult{{ServiceName: "EC2", MethodName: "ec2:DescribeInstances", Error: err}}`

**Edge Case 4: Timestamp Precision**
- Use `time.Now()` when creating each ScanResult
- Don't reuse a single timestamp for all results in a service
- This allows tracking which results came first

**Edge Case 5: Details Map**
- Keep it simple - only add details that are meaningful
- Don't dump entire AWS SDK response into Details
- Examples of good details: {"region": "us-east-1", "count": 42, "state": "running"}

**Edge Case 6: Nil Pointers in AWS SDK Responses**
- AWS SDK often returns pointers (*string, *int64)
- Always check for nil before dereferencing
- Example: `if bucket.Name != nil { resourceName = *bucket.Name } else { resourceName = "unknown" }`

### 🎯 Success Criteria Checklist

**Before Marking Story as Done:**

- [ ] OutputFormatter interface exists in formatters/output_formatter.go
- [ ] ScanResult struct exists in types/types.go with all required fields
- [ ] AWSService.Process signature updated to return []ScanResult
- [ ] All 34 services compile without errors
- [ ] Main.go collects results in []ScanResult slice
- [ ] Output is 100% identical to v0.3.0 (verify manually)
- [ ] Unit tests written for ScanResult creation
- [ ] Unit tests achieve >70% coverage
- [ ] Integration test verifies result collection works
- [ ] `go build ./cmd/awtest` succeeds
- [ ] `go test ./cmd/awtest/...` passes
- [ ] Manual testing with real AWS credentials shows identical output
- [ ] Error scenarios tested (access denied, invalid credentials)
- [ ] No regression in existing functionality
- [ ] Code follows Go naming conventions
- [ ] Documentation updated (if needed)

### 💡 Implementation Tips

**Tip 1: Start with the Interface**
Create the OutputFormatter interface first, even though it won't be used yet. This validates your design early.

**Tip 2: Use Table-Driven Tests**
Go convention for tests:
```go
func TestScanResult_Creation(t *testing.T) {
    tests := []struct {
        name    string
        service string
        method  string
        wantErr bool
    }{
        {"valid result", "S3", "s3:ListBuckets", false},
        {"with error", "EC2", "ec2:DescribeInstances", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic here
        })
    }
}
```

**Tip 3: Use Git to Track Progress**
Commit after each phase:
- `git commit -m "Add OutputFormatter interface and ScanResult struct"`
- `git commit -m "Update S3 and EC2 services to return ScanResult"`
- `git commit -m "Update remaining services to return ScanResult"`
- `git commit -m "Update main loop to collect results"`
- `git commit -m "Add unit tests for ScanResult"`

**Tip 4: Verify Each Service Individually**
After updating each service, run just that service to verify it works:
- Build: `go build ./cmd/awtest`
- Test manually with AWS credentials that have access to that service
- Verify output matches exactly

**Tip 5: Keep utils.PrintResult for Now**
Don't remove utils.PrintResult() calls from services yet. Keep them for backward compatibility. They'll be removed in Story 1.6 when the formatter system is fully integrated.

## Dev Agent Record

### Agent Model Used

Story created by: claude-sonnet-4-5 (Claude Code CLI)
Implementation agent: TBD (will be filled during dev-story execution)

### Implementation Checklist

**Pre-Implementation:**
- [ ] Read this entire story document carefully
- [ ] Understand the current architecture (review main.go, types/types.go, example services)
- [ ] Note the 7-step implementation strategy above
- [ ] Verify go build works before making any changes

**Phase 1 - Foundation:**
- [ ] Create cmd/awtest/formatters/ directory
- [ ] Create formatters/output_formatter.go with OutputFormatter interface
- [ ] Add ScanResult struct to types/types.go
- [ ] Add HasError() helper method to ScanResult
- [ ] Verify go build ./cmd/awtest succeeds

**Phase 2 - Update Signature:**
- [ ] Update AWSService.Process signature in types/types.go to return []ScanResult
- [ ] Note: This WILL break all 34 services - that's expected
- [ ] Commit: "Update AWSService.Process signature to return []ScanResult"

**Phase 3 - Proof of Concept:**
- [ ] Update services/s3/calls.go Process() to return []ScanResult
- [ ] Keep utils.PrintResult() calls (dual mode for now)
- [ ] Test S3 service manually
- [ ] Update services/ec2/calls.go Process() to return []ScanResult
- [ ] Test EC2 service manually
- [ ] Verify output matches v0.3.0 exactly
- [ ] Commit: "Update S3 and EC2 services as proof of concept"

**Phase 4 - Systematic Updates (All 34 Services):**
- [ ] Update amplify service
- [ ] Update apigateway service
- [ ] Update appsync service
- [ ] Update batch service
- [ ] Update cloudformation service
- [ ] Update cloudfront service
- [ ] Update cloudtrail service
- [ ] Update cloudwatch service
- [ ] Update codepipeline service
- [ ] Update cognitoidentity service
- [ ] Update dynamodb service
- [ ] Update ecs service
- [ ] Update elasticbeanstalk service
- [ ] Update eventbridge service
- [ ] Update glacier service
- [ ] Update glue service
- [ ] Update iam service
- [ ] Update iot service
- [ ] Update ivs service
- [ ] Update ivschat service
- [ ] Update ivsrealtime service
- [ ] Update kms service
- [ ] Update lambda service
- [ ] Update rds service
- [ ] Update route53 service
- [ ] Update secretsmanager service
- [ ] Update ses service
- [ ] Update sns service
- [ ] Update sqs service
- [ ] Update sts service
- [ ] Update transcribe service
- [ ] Update waf service
- [ ] Verify go build ./cmd/awtest succeeds
- [ ] Commit: "Update all remaining services to return ScanResult"

**Phase 5 - Main Loop Integration:**
- [ ] Add results := []types.ScanResult{} before service loop in main.go
- [ ] Update loop to collect results: results = append(results, result...)
- [ ] Create temporary printResultAsText() function
- [ ] Add loop after service enumeration to print results
- [ ] Test with real AWS credentials
- [ ] Verify output is IDENTICAL to v0.3.0
- [ ] Commit: "Update main loop to collect and print results"

**Phase 6 - Testing:**
- [ ] Create cmd/awtest/types/types_test.go
- [ ] Write TestScanResult_Creation test
- [ ] Write TestScanResult_WithError test
- [ ] Write TestScanResult_HasError test
- [ ] Create cmd/awtest/formatters/output_formatter_test.go (minimal)
- [ ] Run go test ./cmd/awtest/types
- [ ] Run go test ./cmd/awtest/formatters
- [ ] Check coverage: go test -cover ./cmd/awtest/...
- [ ] Ensure >70% coverage achieved
- [ ] Commit: "Add comprehensive unit tests for ScanResult"

**Phase 7 - Final Verification:**
- [ ] Run go build ./cmd/awtest - must succeed
- [ ] Run go test ./cmd/awtest/... - all tests must pass
- [ ] Test with explicit credentials (--aki, --sak)
- [ ] Test with AWS_PROFILE
- [ ] Test with session token (ASIA* credentials)
- [ ] Test with invalid credentials (verify error handling)
- [ ] Test with credentials that have limited permissions
- [ ] Compare output side-by-side with v0.3.0
- [ ] Verify zero visual differences in output
- [ ] Update story status to "done"
- [ ] Commit: "Story 1.1 complete - formatter interface and result collection"

### Debug Log References

*This section will be filled during implementation with references to any issues encountered and their solutions.*

### Completion Notes List

*This section will be filled during implementation with notes about:*
- Any deviations from the plan
- Edge cases discovered
- Performance observations
- Recommendations for future stories

### Files Created

*To be filled during implementation:*
- [ ] cmd/awtest/formatters/output_formatter.go
- [ ] cmd/awtest/formatters/output_formatter_test.go
- [ ] cmd/awtest/types/types_test.go

### Files Modified

*To be filled during implementation:*
- [ ] cmd/awtest/types/types.go (Add ScanResult, update Process signature)
- [ ] cmd/awtest/main.go (Add result collection loop)
- [ ] cmd/awtest/services/s3/calls.go (Update Process)
- [ ] cmd/awtest/services/ec2/calls.go (Update Process)
- [ ] ... (all 34 service files)

### Test Results

*To be filled during implementation:*
```
go test ./cmd/awtest/...
Coverage: ____%

Manual Testing Results:
- Explicit credentials: ✓/✗
- AWS_PROFILE: ✓/✗
- Session token: ✓/✗
- Invalid credentials: ✓/✗
- Limited permissions: ✓/✗
- Output comparison: ✓/✗
```

---

## 🎯 STORY READY FOR DEVELOPMENT

**Context Engine Analysis:** ✅ Complete
**Architecture Guardrails:** ✅ Defined
**Implementation Strategy:** ✅ Documented
**Testing Requirements:** ✅ Specified
**Success Criteria:** ✅ Clear

**Next Action:** Run `/bmad-bmm-dev-story` with this file to begin implementation.

**Estimated Effort:** 4-6 hours (includes testing and verification)
**Risk Level:** Medium (touches all 34 services, but changes are mechanical)
**Dependencies:** None (this is Epic 1 Story 1)
**Blocks:** Stories 1.2, 1.3, 1.4, 1.5, 1.6, 1.7 (entire Epic 1)
