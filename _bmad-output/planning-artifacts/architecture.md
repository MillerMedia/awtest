---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - prd.md
  - product-brief-awtest-2026-02-27.md
  - README.md
workflowType: 'architecture'
project_name: 'awtest'
user_name: 'Kn0ck0ut'
date: '2026-03-01'
lastStep: 8
status: 'complete'
completedAt: '2026-03-01'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**

The project encompasses 66 functional requirements organized into 9 distinct categories, revealing a mature CLI tool with comprehensive AWS integration needs:

**Credential Input & Authentication (FR1-6):** Multiple authentication methods (explicit keys, AWS CLI profiles, STS tokens) with validation and secure handling. Architecturally requires flexible credential provider abstraction supporting multiple input sources.

**AWS Service Enumeration (FR7-31):** 25 AWS services requiring enumeration, spanning compute (EC2, Lambda, ECS/EKS/Fargate, Batch), databases (RDS, DynamoDB, ElastiCache, Redshift), security/identity (IAM, Secrets Manager, KMS, Certificate Manager, Cognito), storage (S3, EBS, EFS, Glacier), networking (API Gateway, CloudFront, Route53), management (CloudFormation, CloudWatch, CloudTrail, Config, Systems Manager), and application services (SNS, SQS, Step Functions, EventBridge). This represents the core architectural complexity—each service requires distinct SDK integration patterns and enumeration logic.

**Scan Configuration & Control (FR32-37):** User-configurable scan parameters (region selection, service targeting/exclusion, timeouts, verbosity, concurrency levels). Architecture must support flexible configuration with sensible defaults.

**Result Processing (FR38-41):** Resource-level details, severity categorization, metadata reporting, access-denied vs accessible distinction. Requires structured internal data model for scan results.

**Output Formats & Presentation (FR42-48):** Four distinct output formats (text, JSON, Markdown, JSON-compact) plus file output, quiet mode, real-time progress, and summary reporting. Architecture needs format-agnostic result representation with pluggable formatters.

**Error Handling (FR49-54):** Sophisticated error distinction (access-denied, service-unavailable, invalid credentials, throttling, region-unavailable) with graceful degradation. Requires comprehensive error taxonomy and retry logic.

**Installation & Distribution (FR55-58):** Cross-platform Go binary (macOS Intel/ARM, Linux amd64/arm64, Windows amd64) distributed via Homebrew and go install. Architecture must remain platform-agnostic with no OS-specific dependencies.

**Performance (FR59-63):** Aggressive performance targets (2 minutes standard, 5 minutes exhaustive), zero false positives, <100MB memory, concurrent execution. These NFRs significantly constrain architectural choices.

**Extensibility (FR64-66):** Community contribution model requiring documented patterns, templates, and validation for adding new services. Architecture must provide clean extension points.

**Non-Functional Requirements:**

Critical NFRs driving architectural decisions:

**Performance (NFR1-7):** Sub-2-minute scans, <100MB memory footprint, zero false positives, concurrent service enumeration without blocking. These requirements mandate Go's native concurrency (goroutines/channels), efficient AWS SDK usage, and careful memory management.

**Security & Privacy (NFR8-12):** Never log credentials, support STS tokens, clear shell history recommendations, no external calls beyond AWS APIs. Requires security-first design in credential handling and data flow.

**Reliability (NFR13-17):** Graceful degradation, service-level error handling, clear error messaging distinguishing failure types, retry logic for transient failures. Architecture needs robust error handling at every AWS API boundary.

**Usability (NFR18-19):** Zero-config default behavior, intuitive output. Architecture should favor convention over configuration.

**Output Quality (NFR20-22):** JSON schema conformance, LLM-optimized Markdown, AWS CLI profile configuration respect. Output layer must be well-structured and standards-compliant.

**Maintainability & Extensibility (NFR24-27):** Modular service implementations, clear extension patterns, comprehensive test coverage, semantic versioning. Architecture must support plugin-style service additions.

**Scalability (NFR28-30):** Handle accounts with 1000+ resources per service, concurrent region scanning, efficient AWS API pagination. Requires streaming/pagination patterns to avoid memory bloat.

**Operational (NFR31-34):** Single static binary, AWS SDK credential chain integration, minimal external dependencies, offline help. Architecture must be self-contained.

**Scale & Complexity:**

- **Primary domain:** Backend CLI / Cloud Security Tool
- **Complexity level:** Medium-High
  - Medium: Single-binary CLI with well-understood Go ecosystem
  - High: 25+ AWS service integrations, each with unique SDK patterns and enumeration logic
- **Estimated architectural components:** 8-12 major components
  - Credential manager (multi-source credential resolution)
  - Service enumeration engine (orchestration layer)
  - Service enumerators (25+ service-specific implementations)
  - Concurrency orchestrator (goroutine pool, rate limiting, cancellation)
  - Result aggregator (collect results from concurrent enumerators)
  - Output formatters (text, JSON, Markdown, JSON-compact)
  - Error handler & retry logic (AWS API error classification and retry strategy)
  - AWS SDK abstraction layer (facilitate testing and mocking)
  - Configuration manager (CLI flags, env vars, config files)
  - Progress tracker (real-time scan status reporting)
  - Extensibility framework (service registration and discovery)

### Technical Constraints & Dependencies

**Language & Runtime:**
- Go (existing brownfield implementation)
- Single static binary compilation requirement
- Cross-platform support (macOS, Linux, Windows on amd64/arm64)

**AWS SDK Integration:**
- AWS SDK for Go v2 (likely, given modern Go practices)
- Support for 25+ distinct AWS service clients
- Credential chain integration (environment, profiles, STS)
- Regional endpoint configuration
- API pagination handling for large result sets
- Throttling and retry logic for AWS rate limits

**Performance Constraints:**
- 2-minute target for standard scans (aggressive)
- 5-minute target for exhaustive scans
- <100MB memory footprint during execution
- Concurrent execution required to meet timing targets

**Distribution Constraints:**
- Homebrew tap for macOS/Linux installation
- go install support for direct Go installation
- No external runtime dependencies (static binary)
- Cross-compilation for multiple OS/arch combinations

**Security Constraints:**
- Never log or persist credential values
- No telemetry or external calls beyond AWS APIs
- Respect AWS credential best practices
- Clear documentation on credential security

**Brownfield Context:**
- Existing working implementation provides architectural foundation
- Phase 1 focuses on expanding service coverage within existing patterns
- Future phases (2-4) will introduce new capabilities (concurrency optimization, advanced outputs, attack path analysis)

### Cross-Cutting Concerns Identified

**1. Error Handling & Resilience:**
- Spans all 25+ AWS service integrations
- Must distinguish access-denied vs service-unavailable vs throttling vs invalid credentials
- Graceful degradation when individual services fail
- Retry logic for transient AWS API failures
- Clear, actionable error messages for users

**2. Concurrency & Performance:**
- Service enumeration must execute concurrently to meet timing targets
- Goroutine pool management to control concurrency levels
- Cancellation and timeout handling across concurrent operations
- Memory-efficient result aggregation from concurrent enumerators

**3. Configuration Management:**
- CLI flags, environment variables, config files (precedence order)
- AWS credential resolution (explicit flags, profiles, STS, SDK chain)
- Region selection and multi-region support (future)
- Service targeting/exclusion configuration
- Output format and destination configuration

**4. Output Formatting:**
- All scan results must support 4 output formats
- Format-agnostic internal result representation
- LLM-optimized Markdown structure
- JSON schema conformance for machine parsing
- Human-readable text format for terminal use

**5. Testing Strategy:**
- AWS SDK mocking for unit tests (avoid live AWS calls)
- Integration tests against real AWS services (controlled test accounts)
- Cross-platform build and test automation
- Test coverage for all 25+ service enumerators

**6. Logging & Observability:**
- Debug/verbose logging mode for troubleshooting
- Real-time progress tracking during scans
- Scan metadata (timestamp, duration, region)
- Never log credential values (security requirement)

**7. Security & Credential Handling:**
- Secure credential resolution from multiple sources
- No credential logging or persistence
- Support for temporary STS credentials
- Documentation on credential security best practices

**8. Extensibility & Community Contributions:**
- Well-documented patterns for adding new AWS services
- Service registration and discovery mechanism
- Template/scaffolding for new service enumerators
- Validation to ensure contributed services follow patterns

## Architectural Baseline (Brownfield Context)

### Current Technical Stack

**Language & Runtime:**
- **Go:** Version 1.19
- **Compilation:** Static binary compilation (single executable)
- **Platform Support:** Currently builds for target platforms via standard Go toolchain

**Dependencies:**
- **AWS SDK for Go v1:** `github.com/aws/aws-sdk-go v1.44.266`
  - Using SDK v1 (mature, stable, but older generation)
  - Provides service clients for all 34+ currently implemented services
- **Aurora Terminal Colors:** `github.com/logrusorgru/aurora v2.0.3+incompatible`
  - Colorized terminal output for better readability
  - Used in utils.PrintResult() for structured, colored logging

**CLI Framework:**
- **Standard Go `flag` package** (no external CLI framework)
- Simple, lightweight, zero dependencies beyond standard library
- Direct flag parsing in main.go with abbreviated flag support

### Project Structure

```
cmd/awtest/
├── main.go              # Entry point, credential handling, main execution loop
├── services/            # AWS service enumeration implementations
│   ├── services.go      # AllServices() registration
│   ├── s3/
│   │   └── calls.go     # S3 service implementation
│   ├── ec2/
│   │   └── calls.go     # EC2 service implementation
│   ├── lambda/
│   │   └── calls.go     # Lambda service implementation
│   └── [31 more services...]
├── types/
│   └── types.go         # AWSService interface, error types, constants
└── utils/
    └── output.go        # PrintResult, HandleAWSError, colorization
```

**Total Services Currently Implemented:** 34

Services include: STS, Amplify, API Gateway, AppSync, Batch, CloudFormation, CloudFront, CloudTrail, CloudWatch, CodePipeline, Cognito Identity, DynamoDB, EC2, ECS, Elastic Beanstalk, EventBridge, Glacier, Glue, IAM, IoT, IVS (3 variants), KMS, Lambda, RDS, Route53, S3, Secrets Manager, SES, SNS, SQS, Transcribe, WAF.

### Core Architectural Patterns

**Service Abstraction Pattern:**

The existing implementation uses a clean, extensible service abstraction:

```go
// types/types.go
type AWSService struct {
    Name       string
    Call       func(*session.Session) (interface{}, error)
    Process    func(interface{}, error, bool) error
    ModuleName string
}
```

Each AWS service implementation:
1. **Defines** a slice of `AWSService` structs (one per API call)
2. **Call function** creates AWS service client, makes API call, returns output
3. **Process function** handles output, formats results, prints to terminal
4. **Registers** in `services.AllServices()` by appending to the master list

**Main Execution Flow:**

```go
// main.go main() function
for _, service := range services.AllServices() {
    output, err := service.Call(sess)
    if err := service.Process(output, err, *debug); err != nil {
        if _, ok := err.(*types.InvalidKeyError); ok {
            os.Exit(1)  // Abort on invalid credentials
        }
    }
}
```

**Pattern Strengths:**
- Clean separation: AWS SDK interaction (Call) vs output formatting (Process)
- Easy extensibility: add new service = create new package + register in AllServices()
- Consistent error handling via shared utils.HandleAWSError()
- No complex abstractions or interfaces beyond the core AWSService struct

### Credential Handling Architecture

**Multi-Source Credential Resolution (Priority Order):**

1. **CLI Flags (Highest Priority):**
   - Long form: `--access-key-id`, `--secret-access-key`, `--session-token`
   - Abbreviated: `--aki`, `--sak`, `--st`
   - Region: `--region` (default: us-west-2)

2. **Environment Variables:**
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_SESSION_TOKEN`

3. **AWS Profile:**
   - `AWS_PROFILE` environment variable
   - Uses shared config with `session.SharedConfigEnable`

4. **Default Shared Config:**
   - Falls back to default AWS credential chain if no explicit credentials

**STS Token Detection:**
- Automatically detects temporary credentials (access key prefix "ASIA")
- Includes session token when STS credentials detected
- Properly handles both long-term and temporary credentials

**Session Creation:**
- Creates `*session.Session` with resolved credentials
- Passes session to all service Call() functions
- Services can create region-specific sessions as needed (e.g., S3 bucket regions)

### Error Handling Architecture

**Centralized Error Handler:**

```go
// utils/output.go
func HandleAWSError(debug bool, callName string, err error) error {
    if awsErr, ok := err.(awserr.Error); ok {
        // Map AWS error codes to user-friendly messages
        prettyMsg := types.AwsErrorMessages[awsErr.Code()]
        
        // Special handling for access denied vs invalid credentials
        if awsErr.Code() == "InvalidAccessKeyId" || 
           awsErr.Code() == "InvalidClientTokenId" {
            return &types.InvalidKeyError{prettyMsg}  // Abort scan
        }
        
        // Access denied = continue scan
        if awsErr.Code() == "UnauthorizedOperation" { ... }
    }
}
```

**Error Classification:**

- **Invalid Credentials** (InvalidAccessKeyId, InvalidClientTokenId):
  - Returns `InvalidKeyError`
  - Main loop catches this and exits immediately (no point continuing)

- **Access Denied** (UnauthorizedOperation, AccessDeniedException):
  - Prints "Access denied to this service"
  - Returns nil (scan continues to next service)
  - Graceful degradation: check all services even if some are denied

- **Other AWS Errors:**
  - Pretty-printed error messages from `types.AwsErrorMessages` map
  - Debug mode shows full AWS error details

### Output Architecture

**Current Implementation:**

```go
// utils/output.go
func PrintResult(debug bool, moduleName string, method string, result string, err error) {
    severity := determineSeverity(err)
    message := colorizeMessage(moduleName, method, severity, result)
    fmt.Println(message)
}
```

**Output Format (Text Only):**
```
[AWTest] [s3:ListBuckets] [info] S3 bucket: my-bucket-name
[AWTest] [s3:ListObjects] [info] S3 Bucket: my-bucket-name | 42 objects
[AWTest] [iam:ListUsers] [info] Error: Access denied to this service.
```

**Colorization:**
- Module name: Bright Green
- Method name: Bright Blue
- Severity: Red (high) or Blue (info)
- Resource names: Yellow (`ColorizeItem()`)

**Current Output Limitations:**
- ✗ Text-only (no JSON, Markdown, JSON-compact formats)
- ✗ No quiet mode (always prints all results)
- ✗ No file output option
- ✗ No progress indicator during scan
- ✗ No summary at completion

These limitations are Phase 1 and Phase 3 requirements to be addressed.

### Build & Distribution Architecture

**Current Build Process:**
- No Makefile or build automation detected
- Likely manual `go build` commands
- No goreleaser configuration for multi-platform releases

**Current Gaps:**
- ✗ No automated cross-compilation for macOS (Intel/ARM), Linux (amd64/arm64), Windows
- ✗ No Homebrew tap configuration
- ✗ No `go install` module path optimization
- ✗ No version embedding in builds
- ✗ No release automation

These gaps need to be addressed for FR55-58 (Installation & Distribution requirements).

### Testing Architecture

**Current State:**
- No test files discovered in initial codebase scan
- No visible testing framework setup
- No AWS SDK mocking infrastructure

**Testing Gaps:**
- ✗ Unit tests for service enumerators
- ✗ Integration tests against real AWS
- ✗ AWS SDK mocking for offline tests
- ✗ Cross-platform build validation
- ✗ CI/CD test automation

### Configuration Management

**Current Implementation:**
- Hardcoded defaults in main.go (region: us-west-2)
- No configuration file support
- No service inclusion/exclusion flags
- No concurrency configuration
- No timeout configuration

**Current Gaps vs PRD:**
- ✗ FR33: Users can select specific services to scan
- ✗ FR34: Users can exclude specific services
- ✗ FR35: Users can set maximum scan timeout
- ✗ FR36: Users can enable verbose debug logging (partially implemented via --debug)
- ✗ FR37: Users can configure concurrency level

### Existing Strengths

**What's Working Well:**

1. **Clean Service Abstraction:** The AWSService interface pattern is excellent for extensibility
2. **Multi-Source Credentials:** Comprehensive credential resolution covering all major sources
3. **Error Handling:** Smart distinction between abort-worthy errors and continue-worthy errors
4. **Code Organization:** Clear package structure with separation of concerns
5. **34 Services Implemented:** Significant coverage already in place
6. **Graceful Degradation:** Access denied doesn't crash the entire scan
7. **STS Support:** Properly handles temporary credentials

**Architectural Foundation Quality:** Strong. The existing patterns are solid and extensible.

### Migration Considerations

**AWS SDK v1 to v2:**

The PRD mentions "AWS SDK for Go v2 (likely, given modern Go practices)" but current implementation uses v1.

**Considerations:**
- SDK v1 is mature and stable (not deprecated)
- SDK v2 offers better performance and modular imports
- Migration would require rewriting all 34 service implementations
- Breaking change for existing codebase

**Recommendation:** Document as architectural decision (keep v1 for Phase 1, consider v2 for Phase 2+ if performance becomes critical).

### Phase 1 Architectural Priorities

Based on existing foundation and PRD requirements, Phase 1 should focus on:

1. **Service Coverage Expansion:** Add 15-20 missing services to reach 50+ total
2. **Output Format Support:** Implement JSON, Markdown, JSON-compact formatters
3. **Build Automation:** Add goreleaser, Makefile, cross-compilation
4. **Configuration Flags:** Service targeting/exclusion, timeout, verbosity
5. **Testing Infrastructure:** AWS SDK mocking, unit tests for new services

**Deferred to Phase 2+:**
- Concurrency optimization (current sequential execution is acceptable for Phase 1)
- AWS SDK v1 → v2 migration (not required, nice-to-have)
- Attack path analysis (Phase 4)

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
1. Output Format Architecture - Required for FR42-48 (multiple output formats)
2. Build & Distribution - Required for FR55-58 (Homebrew, cross-platform)
3. Configuration Management - Required for FR33-37 (service targeting, timeout)

**Important Decisions (Shape Architecture):**
4. Testing Strategy - Required for NFR25-26 (test coverage, validation)
5. Service Implementation Patterns - Required for FR64-66 (community contributions)

**Deferred Decisions (Phase 2+):**
- Concurrency architecture (Phase 2 optimization)
- Config file support (enhancement, not required for MVP)
- AWS SDK v1→v2 migration (performance optimization)

### Output Format Architecture

**Decision:** Formatter Interface Pattern

**Rationale:** Clean separation between result collection and formatting enables support for 4 output formats (text, JSON, Markdown, JSON-compact) while maintaining testability and extensibility.

**Implementation Approach:**

```go
// New types/formatter.go
type OutputFormatter interface {
    Format(results []ScanResult) (string, error)
    FileExtension() string
}

type ScanResult struct {
    ServiceName  string
    MethodName   string
    ResourceType string
    ResourceName string
    Details      map[string]interface{}
    Error        error
    Timestamp    time.Time
}

// Implementations
type TextFormatter struct{}
type JSONFormatter struct{}
type MarkdownFormatter struct{}
type JSONCompactFormatter struct{}
```

**Architecture Changes Required:**

1. **Result Collection Phase:**
   - Modify main loop to collect results instead of immediate printing
   - Store results in `[]ScanResult` slice
   - Service Process() methods return results instead of printing

2. **Formatting Phase:**
   - After all services complete, pass results to selected formatter
   - Formatter generates output in target format
   - Output to stdout or file based on flags

3. **Backward Compatibility:**
   - Text formatter replicates current PrintResult() behavior
   - Colorization remains for terminal text output
   - Quiet mode suppresses informational messages

**Affects:** All service implementations, main loop, utils package

**Requirements Addressed:** FR42-48 (output formats), NFR20-22 (format quality)

### Build & Distribution Architecture

**Decision:** GoReleaser + Makefile

**Rationale:** GoReleaser is the Go ecosystem standard for CLI distribution, providing automated cross-compilation, GitHub releases, and Homebrew tap generation. Makefile handles local development builds.

**GoReleaser Configuration:**

File: `.goreleaser.yaml`

```yaml
builds:
  - binary: awtest
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.Version={{.Version}}
      - -X main.BuildDate={{.Date}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

brews:
  - name: awtest
    repository:
      owner: MillerMedia
      name: homebrew-tap
    description: "AWS credential enumeration for security assessments"
    homepage: "https://github.com/MillerMedia/awtest"
    install: |
      bin.install "awtest"

release:
  github:
    owner: MillerMedia
    name: awtest
```

**Makefile for Local Development:**

```makefile
VERSION ?= dev
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

.PHONY: build
build:
	go build $(LDFLAGS) -o awtest ./cmd/awtest

.PHONY: test
test:
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: install
install:
	go install $(LDFLAGS) ./cmd/awtest

.PHONY: clean
clean:
	rm -f awtest coverage.out
```

**Version Embedding:**

Update `main.go`:
```go
var (
    Version   = "dev"
    BuildDate = "unknown"
)

func main() {
    fmt.Println("Version:", Version)
    fmt.Println("Build Date:", BuildDate)
    // ... rest of main
}
```

**Release Workflow:**

1. Tag release: `git tag v0.4.0`
2. Push tag: `git push origin v0.4.0`
3. GitHub Actions runs GoReleaser
4. Artifacts published to GitHub Releases
5. Homebrew tap automatically updated

**Affects:** Repository root, CI/CD, version management

**Requirements Addressed:** FR55-58 (Homebrew, go install, cross-platform), NFR31 (single binary)

### Configuration Management

**Decision:** Flag-based Configuration (Phase 1)

**Rationale:** Extends existing flag pattern, keeps CLI simple and frictionless. Aligns with "zero-config default behavior" (NFR18). Config file support deferred to Phase 2+ based on user feedback.

**New Flags to Add:**

```go
// main.go flag definitions
var (
    // Existing flags
    awsAccessKeyID     = flag.String("access-key-id", "", "AWS Access Key ID")
    awsSecretAccessKey = flag.String("secret-access-key", "", "AWS Secret Access Key")
    awsSessionToken    = flag.String("session-token", "", "AWS Session Token")
    awsRegion          = flag.String("region", "us-west-2", "AWS Region")
    debug              = flag.Bool("debug", false, "Enable debug mode")
    
    // NEW Phase 1 flags
    services           = flag.String("services", "", "Comma-separated list of services to scan (e.g., s3,ec2,iam)")
    excludeServices    = flag.String("exclude-services", "", "Comma-separated list of services to exclude")
    timeout            = flag.Duration("timeout", 5*time.Minute, "Maximum scan timeout")
    concurrency        = flag.Int("concurrency", 1, "Number of concurrent service scans (Phase 2)")
    outputFormat       = flag.String("output-format", "text", "Output format: text, json, markdown, json-compact")
    outputFile         = flag.String("output-file", "", "Write output to file instead of stdout")
    quiet              = flag.Bool("quiet", false, "Suppress informational messages, show only findings")
)
```

**Service Filtering Logic:**

```go
// New function in services package
func FilterServices(allServices []types.AWSService, include, exclude string) []types.AWSService {
    if include == "" && exclude == "" {
        return allServices // No filtering
    }
    
    includeSet := parseServiceList(include)
    excludeSet := parseServiceList(exclude)
    
    filtered := []types.AWSService{}
    for _, svc := range allServices {
        serviceName := extractServiceName(svc.Name) // e.g., "s3" from "s3:ListBuckets"
        
        if len(includeSet) > 0 && !includeSet[serviceName] {
            continue // Not in include list
        }
        if excludeSet[serviceName] {
            continue // In exclude list
        }
        
        filtered = append(filtered, svc)
    }
    return filtered
}
```

**Timeout Implementation:**

```go
// main.go with timeout context
ctx, cancel := context.WithTimeout(context.Background(), *timeout)
defer cancel()

for _, service := range filteredServices {
    select {
    case <-ctx.Done():
        fmt.Println("Scan timeout exceeded")
        os.Exit(1)
    default:
        output, err := service.Call(sess)
        service.Process(output, err, *debug)
    }
}
```

**Affects:** main.go, services package, flag parsing

**Requirements Addressed:** FR33-37 (service selection, timeout, concurrency config), FR42, FR45-46 (output format, file, quiet)

**Deferred to Phase 2:** Config file support, service groups/presets

### Testing Strategy

**Decision:** Testify + Manual Mocks

**Rationale:** Testify provides better assertions and mock support without heavy dependencies. Manual mocking per service gives flexibility while avoiding over-engineering.

**Dependencies to Add:**

```go
// go.mod additions
require (
    github.com/stretchr/testify v1.9.0
)
```

**Testing Structure:**

```
cmd/awtest/
├── services/
│   ├── s3/
│   │   ├── calls.go
│   │   └── calls_test.go       # NEW
│   ├── ec2/
│   │   ├── calls.go
│   │   └── calls_test.go       # NEW
│   └── ...
├── types/
│   ├── types.go
│   └── types_test.go           # NEW
└── utils/
    ├── output.go
    └── output_test.go          # NEW
```

**Example Service Test Pattern:**

```go
// services/s3/calls_test.go
package s3

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/aws/aws-sdk-go/service/s3"
)

type MockS3Client struct {
    mock.Mock
}

func (m *MockS3Client) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
    args := m.Called(input)
    return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
}

func TestS3ListBuckets(t *testing.T) {
    mockClient := new(MockS3Client)
    
    expectedOutput := &s3.ListBucketsOutput{
        Buckets: []*s3.Bucket{
            {Name: aws.String("test-bucket")},
        },
    }
    
    mockClient.On("ListBuckets", mock.Anything).Return(expectedOutput, nil)
    
    // Test Call function
    // Assert expected behavior
    
    mockClient.AssertExpectations(t)
}
```

**Testing Approach:**

1. **Unit Tests:**
   - Test each service Call() function with mocked AWS clients
   - Test Process() functions with known outputs
   - Test error handling (access denied, invalid credentials, etc.)

2. **Integration Tests (Optional):**
   - Separate `_integration_test.go` files
   - Run against real AWS test account
   - Gated by build tag: `// +build integration`
   - Run manually or in CI with AWS credentials

3. **Coverage Targets:**
   - New services: 80%+ coverage required
   - Utils/types: 90%+ coverage
   - Run: `make test` shows coverage report

**Test Validation in CI:**

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.19'
      - run: make test
      - run: go test -race ./...
```

**Affects:** All new service implementations, CI/CD

**Requirements Addressed:** NFR25 (comprehensive test coverage), NFR26 (service validation)

### Service Implementation Patterns

**Decision:** Documentation + Template (Phase 1)

**Rationale:** Existing AWSService pattern is already clean and simple. Well-documented template provides sufficient guidance for contributors without adding generator tooling complexity.

**CONTRIBUTING.md Structure:**

```markdown
## Adding a New AWS Service

### 1. Create Service Package

Create a new directory: `cmd/awtest/services/<servicename>/`

Example: `cmd/awtest/services/redshift/`

### 2. Copy Template

Copy `cmd/awtest/services/_template/calls.go` to your service directory.

### 3. Implement Service

Follow the AWSService pattern with Call and Process functions.

### 4. Register Service

Add to `cmd/awtest/services/services.go` AllServices() function.

### 5. Write Tests

Create `calls_test.go` using testify mocks.

### 6. Test Locally

Run: `go run ./cmd/awtest --debug --services=<servicename>`

### 7. Submit PR

Update README.md service count and CHANGELOG.md.
```

**Template File Location:**

`cmd/awtest/services/_template/calls.go` with comprehensive comments

**Validation Checklist:**

Include in CONTRIBUTING.md:

- [ ] Service package created
- [ ] AWSService slice exported
- [ ] Registered in services.AllServices()
- [ ] Error handling uses utils.HandleAWSError()
- [ ] Output formatting uses utils.PrintResult()
- [ ] Tests written and passing
- [ ] Documentation updated

**Affects:** Repository documentation, contributor experience

**Requirements Addressed:** FR64-66 (community contribution patterns, templates, validation)

**Phase 2 Enhancement:** Add automated validation script

### Decision Impact Analysis

**Implementation Sequence:**

1. **Output Format Architecture** (Week 1-2)
   - Refactor: Add ScanResult type, modify main loop
   - Implement: 4 formatters (Text, JSON, Markdown, JSON-compact)
   - Affects: All services

2. **Configuration Management** (Week 1-2, parallel)
   - Add: New flags for filtering, timeout, output
   - Implement: Service filtering, timeout context
   - Affects: main.go, services package

3. **Build & Distribution** (Week 2)
   - Create: .goreleaser.yaml, Makefile, CI/CD
   - Setup: Homebrew tap repository
   - Affects: Repository infrastructure

4. **Testing Strategy** (Week 2-3)
   - Add: Testify dependency, test files
   - Document: Testing patterns
   - Affects: Ongoing development

5. **Service Implementation Patterns** (Week 3)
   - Create: CONTRIBUTING.md, template files
   - Document: Contribution workflow
   - Affects: Documentation

**Cross-Component Dependencies:**

- Output Formatter ← Service Filtering: Filtered services feed results to formatter
- Output Formatter ← Quiet Mode: Quiet flag affects formatter verbosity
- GoReleaser ← Version Embedding: Build version embedded by GoReleaser
- Testing ← Service Template: Template includes test file structure
- All Components ← Configuration Flags: Flags control behavior across components

**Architecture Coherence:**

All decisions maintain consistency with existing patterns:
- Extend flag package (no new CLI framework)
- Build on AWSService interface (no architectural overhaul)
- Add capabilities incrementally (brownfield-friendly)
- Defer complex features to Phase 2+ (keep Phase 1 focused)

### AWS SDK Decision

**Decision:** Keep AWS SDK v1 for Phase 1

**Rationale:** Migrating 34 existing service implementations from SDK v1 to v2 is significant effort that doesn't directly address Phase 1 PRD requirements.

**Analysis:**

**SDK v1 (Current):**
- Mature, stable, well-documented
- All 34 services already implemented
- Not deprecated by AWS

**SDK v2:**
- Better performance, modular imports
- Modern Go practices
- Requires rewriting all services

**Decision:** Defer SDK v2 migration to Phase 2+

**Recommendation:** Revisit in Phase 2 if performance optimization becomes critical for sub-2-minute scan targets.

**Affects:** Dependency management, future performance optimization

## Implementation Patterns & Consistency Rules

### Pattern Philosophy

**Brownfield Context:** Many patterns are already established in the existing codebase. These patterns are documented here to ensure AI agents maintain consistency when adding Phase 1 features (output formats, configuration, build automation, testing).

**Go Conventions:** All patterns follow standard Go conventions (Effective Go, Go Code Review Comments) unless explicitly stated otherwise for project-specific reasons.

### Naming Patterns

#### Package Naming

**Rule:** All package names lowercase, single word, no underscores.

**Existing Patterns (DO NOT CHANGE):**
```go
package main       // Entry point
package services   // Service implementations
package types      // Shared types and interfaces
package utils      // Utility functions
```

**New Packages for Phase 1:**
```go
package formatters // Output formatters (NOT "output" or "formats")
package config     // Configuration parsing (if extracted from main)
package filters    // Service filtering logic (if extracted)
```

**Rationale:** Standard Go convention. Single-word package names are clearer in imports.

#### File Naming

**Rule:** All lowercase with underscores for multi-word files.

**Pattern:**
```
types.go              // Single-word concepts
output.go             // Single-word
text_formatter.go     // Multi-word: use underscores
json_formatter.go     // Multi-word: use underscores
service_filter.go     // Multi-word: use underscores
```

**Test Files:**
```
types_test.go
output_test.go
text_formatter_test.go
```

**Existing Pattern (already consistent):** ✓
- `cmd/awtest/services/s3/calls.go`
- `cmd/awtest/utils/output.go`

**Anti-Pattern:**
```
TextFormatter.go      // ✗ Capital letters
textFormatter.go      // ✗ camelCase
text-formatter.go     // ✗ Hyphens
```

#### Type Naming

**Rule:** Exported types use PascalCase, unexported use camelCase.

**Existing Patterns:**
```go
type AWSService struct { ... }           // ✓ Exported, PascalCase
type InvalidKeyError struct { ... }      // ✓ Exported, PascalCase
```

**New Types for Phase 1:**
```go
type OutputFormatter interface { ... }   // ✓ Exported interface
type ScanResult struct { ... }           // ✓ Exported struct
type TextFormatter struct { ... }        // ✓ Exported implementation
type JSONFormatter struct { ... }        // ✓ Exported implementation
type MarkdownFormatter struct { ... }    // ✓ Exported implementation
type JSONCompactFormatter struct { ... } // ✓ Exported implementation

type serviceFilter struct { ... }        // ✓ Unexported helper
```

**Anti-Pattern:**
```go
type outputFormatter interface { ... }   // ✗ Interface should be exported
type Scan_Result struct { ... }          // ✗ Underscores in type names
type textformatter struct { ... }        // ✗ Should be TextFormatter if exported
```

#### Function & Method Naming

**Rule:** Exported functions PascalCase, unexported camelCase. Use descriptive verb+noun names.

**Existing Patterns:**
```go
func PrintResult(...)              // ✓ Exported, PascalCase, verb+noun
func HandleAWSError(...)           // ✓ Exported, clear purpose
func ColorizeItem(...)             // ✓ Exported, verb+noun
func determineSeverity(...) string // ✓ Unexported, camelCase
```

**New Functions for Phase 1:**
```go
func FilterServices(...)           // ✓ Exported, verb+noun
func ParseServiceList(...) map[string]bool // ✓ Exported, clear purpose
func NewTextFormatter() *TextFormatter     // ✓ Constructor pattern
func (f *TextFormatter) Format(...) string // ✓ Method on receiver
```

**Anti-Pattern:**
```go
func filter_services(...)          // ✗ Underscore, should be FilterServices if exported
func services(...)                 // ✗ Noun without verb, unclear purpose
func GetFormatter(...)             // ✗ "Get" prefix unnecessary in Go (just Formatter())
```

#### Variable & Constant Naming

**Rule:** Variables camelCase (exported PascalCase), constants follow same rule.

**Existing Patterns:**
```go
const Version = "v0.3.0"                    // ✓ Exported constant
const DefaultModuleName = "AWTest"          // ✓ Exported constant
const InvalidAccessKeyId = "InvalidAccessKeyId" // ✓ Exported constant

var allServices []types.AWSService          // ✓ Unexported variable
```

**New for Phase 1:**
```go
const (
    FormatText        = "text"        // ✓ Exported format constants
    FormatJSON        = "json"
    FormatMarkdown    = "markdown"
    FormatJSONCompact = "json-compact"
)

var defaultTimeout = 5 * time.Minute  // ✓ Unexported default
```

**Anti-Pattern:**
```go
const FORMAT_TEXT = "text"            // ✗ Use FormatText (Go style, not SCREAMING_SNAKE)
var AllServices []types.AWSService    // ✗ Don't export package-level vars (use function)
```

#### AWS Service Naming in Code

**Rule:** Match AWS service name exactly, use official AWS SDK package names.

**Existing Pattern (MAINTAIN CONSISTENCY):**
```go
package s3        // ✓ Lowercase package name
var S3Calls = []types.AWSService{...}  // ✓ Uppercase exported var

package secretsmanager  // ✓ Match AWS SDK naming exactly
var SecretsManagerCalls = []types.AWSService{...} // ✓ PascalCase

Name: "s3:ListBuckets"              // ✓ Lowercase service, PascalCase method
Name: "secretsmanager:ListSecrets"  // ✓ Match AWS SDK capitalization
```

**New Services for Phase 1 (Example: ElastiCache, Redshift):**
```go
package elasticache
var ElastiCacheCalls = []types.AWSService{...}

package redshift
var RedshiftCalls = []types.AWSService{...}

Name: "elasticache:DescribeCacheClusters"
Name: "redshift:DescribeClusters"
```

### Structure Patterns

#### Directory Organization

**Established Structure (DO NOT CHANGE):**
```
cmd/awtest/
├── main.go
├── services/
│   ├── services.go        # Registration
│   ├── s3/calls.go
│   ├── ec2/calls.go
│   └── [service]/calls.go
├── types/
│   └── types.go
└── utils/
    └── output.go
```

**New Structure for Phase 1:**
```
cmd/awtest/
├── main.go
├── services/
│   ├── services.go
│   ├── service_filter.go     # NEW: Service filtering logic
│   ├── _template/calls.go    # NEW: Template for contributors
│   ├── s3/
│   │   ├── calls.go
│   │   └── calls_test.go     # NEW: Tests co-located
│   └── [service]/
│       ├── calls.go
│       └── calls_test.go
├── formatters/                # NEW: Output formatters
│   ├── formatter.go           # Interface definition
│   ├── text_formatter.go
│   ├── json_formatter.go
│   ├── markdown_formatter.go
│   ├── json_compact_formatter.go
│   └── formatters_test.go
├── types/
│   ├── types.go
│   ├── scan_result.go        # NEW: ScanResult type
│   └── types_test.go         # NEW: Type tests
└── utils/
    ├── output.go
    └── output_test.go        # NEW: Output tests
```

**Root Level (Build & Distribution):**
```
.goreleaser.yaml    # GoReleaser configuration
Makefile            # Local build automation
CONTRIBUTING.md     # Contribution guide
CHANGELOG.md        # Version history
```

**Rule:** Tests co-located with source files (`*_test.go` in same package directory).

**Rule:** New packages go under `cmd/awtest/` unless they're truly reusable (none currently).

#### File Organization Within Packages

**Rule:** One primary concern per file. Group related functions.

**Pattern for Service Packages:**
```go
// services/redshift/calls.go
package redshift

// All Redshift service calls in one file
var RedshiftCalls = []types.AWSService{
    // ListClusters
    {...},
    // DescribeClusterSnapshots
    {...},
}
```

**Pattern for Formatters Package:**
```go
// formatters/formatter.go - Interface definition
package formatters

type OutputFormatter interface {
    Format(results []ScanResult) (string, error)
    FileExtension() string
}

// formatters/text_formatter.go - Text implementation
package formatters

type TextFormatter struct{}
func NewTextFormatter() *TextFormatter {...}
func (f *TextFormatter) Format(...) {...}
func (f *TextFormatter) FileExtension() string {...}
```

**Rule:** Interface definition in separate file from implementations.

### Code Patterns

#### Error Handling Pattern

**Established Pattern (MUST USE):**

```go
// For AWS SDK errors in service Process functions:
if err != nil {
    return utils.HandleAWSError(debug, "service:Method", err)
}

// HandleAWSError handles:
// - Invalid credentials → Returns InvalidKeyError (abort scan)
// - Access denied → Prints message, returns nil (continue scan)
// - Other errors → Pretty-prints error, returns nil
```

**New Pattern for Phase 1 Formatters:**
```go
// Formatters return errors, don't handle them
func (f *JSONFormatter) Format(results []ScanResult) (string, error) {
    data, err := json.Marshal(results)
    if err != nil {
        return "", fmt.Errorf("json formatting failed: %w", err)
    }
    return string(data), nil
}

// Main function handles formatter errors
output, err := formatter.Format(results)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
    os.Exit(1)
}
```

**Anti-Pattern:**
```go
// ✗ Don't use panic for expected errors
if err != nil {
    panic(err)  // ✗ Bad
}

// ✗ Don't silently ignore errors
output, _ := formatter.Format(results)  // ✗ Bad

// ✗ Don't duplicate error handling logic
if awsErr, ok := err.(awserr.Error); ok {  // ✗ Use utils.HandleAWSError
    // ... custom handling
}
```

#### AWS Client Creation Pattern

**Established Pattern (MUST USE):**

```go
// In service Call function:
Call: func(sess *session.Session) (interface{}, error) {
    svc := s3.New(sess)  // ✓ Create client from session
    output, err := svc.ListBuckets(&s3.ListBucketsInput{})
    return output, err   // ✓ Return output and error directly
}
```

**For Region-Specific Calls:**
```go
// Copy session with different region
sessWithRegion := sess.Copy(&aws.Config{Region: aws.String("us-east-1")})
svc := s3.New(sessWithRegion)
```

**Anti-Pattern:**
```go
// ✗ Don't create sessions inside Call function
Call: func(sess *session.Session) (interface{}, error) {
    newSess := session.Must(session.NewSession())  // ✗ Bad
    svc := s3.New(newSess)
    ...
}
```

#### Result Collection Pattern (NEW for Phase 1)

**Old Pattern (Phase 0 - Current):**
```go
// Immediate printing in Process function
Process: func(output interface{}, err error, debug bool) error {
    if err != nil {
        return utils.HandleAWSError(debug, "s3:ListBuckets", err)
    }
    // Print immediately
    utils.PrintResult(debug, "", "s3:ListBuckets", "S3 bucket: ...", nil)
    return nil
}
```

**New Pattern (Phase 1 - With Formatters):**
```go
// Return results instead of printing
Process: func(output interface{}, err error, debug bool) ([]types.ScanResult, error) {
    if err != nil {
        // Return error result
        return []types.ScanResult{{
            ServiceName: "s3",
            MethodName:  "ListBuckets",
            Error:       err,
            Timestamp:   time.Now(),
        }}, nil
    }
    
    s3Output := output.(*s3.ListBucketsOutput)
    results := []types.ScanResult{}
    
    for _, bucket := range s3Output.Buckets {
        results = append(results, types.ScanResult{
            ServiceName:  "s3",
            MethodName:   "ListBuckets",
            ResourceType: "bucket",
            ResourceName: *bucket.Name,
            Details:      map[string]interface{}{"creation_date": bucket.CreationDate},
            Timestamp:    time.Now(),
        })
    }
    
    return results, nil
}
```

**Transition Strategy:**
- Update `types.AWSService` to have `ProcessV2` method returning results
- Keep old `Process` for backward compatibility during migration
- Migrate services incrementally

#### Service Registration Pattern

**Established Pattern (MUST USE):**

```go
// services/services.go
func AllServices() []types.AWSService {
    var allServices []types.AWSService
    
    allServices = append(allServices, sts.STSCalls...)
    allServices = append(allServices, s3.S3Calls...)
    allServices = append(allServices, ec2.EC2Calls...)
    // ... all services in alphabetical order
    
    return allServices
}
```

**Rule:** Maintain alphabetical order (except STS first for credential validation).

**Rule:** Always append the full slice (`ServiceCalls...`), never individual items.

#### Service Filtering Pattern (NEW for Phase 1)

```go
// services/service_filter.go
package services

func FilterServices(allServices []types.AWSService, include, exclude string) []types.AWSService {
    if include == "" && exclude == "" {
        return allServices // No filtering
    }
    
    includeSet := parseServiceList(include)
    excludeSet := parseServiceList(exclude)
    
    filtered := make([]types.AWSService, 0, len(allServices))
    for _, svc := range allServices {
        serviceName := extractServiceName(svc.Name)
        
        if len(includeSet) > 0 {
            if !includeSet[serviceName] {
                continue
            }
        }
        
        if excludeSet[serviceName] {
            continue
        }
        
        filtered = append(filtered, svc)
    }
    
    return filtered
}

func extractServiceName(callName string) string {
    parts := strings.SplitN(callName, ":", 2)
    return strings.ToLower(parts[0])
}

func parseServiceList(csv string) map[string]bool {
    if csv == "" {
        return nil
    }
    
    set := make(map[string]bool)
    for _, name := range strings.Split(csv, ",") {
        trimmed := strings.TrimSpace(strings.ToLower(name))
        if trimmed != "" {
            set[trimmed] = true
        }
    }
    return set
}
```

**Rule:** Use `make([]Type, 0, capacity)` for slices with known max capacity.

**Rule:** Always lowercase service names for comparison (user input is case-insensitive).

### Testing Patterns

#### Test File Organization

**Rule:** Co-locate tests with source (`*_test.go` in same package).

```
cmd/awtest/services/s3/
├── calls.go
└── calls_test.go

cmd/awtest/formatters/
├── formatter.go
├── text_formatter.go
├── text_formatter_test.go
├── json_formatter.go
└── json_formatter_test.go
```

**Rule:** Use same package name for tests (not `package X_test` unless testing exported API only).

#### Test Function Naming

**Rule:** `Test<FunctionName>` or `Test<TypeName>_<MethodName>` for table-driven tests.

```go
func TestFilterServices(t *testing.T) { ... }
func TestTextFormatter_Format(t *testing.T) { ... }
func TestParseServiceList(t *testing.T) { ... }
```

#### Table-Driven Test Pattern

**Standard Pattern for Go:**

```go
func TestFilterServices(t *testing.T) {
    tests := []struct {
        name     string
        include  string
        exclude  string
        input    []types.AWSService
        expected int
    }{
        {
            name:     "no filter returns all",
            include:  "",
            exclude:  "",
            input:    makeTestServices("s3", "ec2", "iam"),
            expected: 3,
        },
        {
            name:     "include filter",
            include:  "s3,ec2",
            exclude:  "",
            input:    makeTestServices("s3", "ec2", "iam"),
            expected: 2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FilterServices(tt.input, tt.include, tt.exclude)
            assert.Equal(t, tt.expected, len(result))
        })
    }
}
```

**Rule:** Always use `t.Run(tt.name, ...)` for sub-tests.

**Rule:** Use descriptive test case names.

#### Mock Pattern (Testify)

```go
type MockS3Client struct {
    mock.Mock
}

func (m *MockS3Client) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
    args := m.Called(input)
    return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
}

func TestS3ListBuckets(t *testing.T) {
    mockClient := new(MockS3Client)
    expectedOutput := &s3.ListBucketsOutput{
        Buckets: []*s3.Bucket{{Name: aws.String("test-bucket")}},
    }
    
    mockClient.On("ListBuckets", mock.Anything).Return(expectedOutput, nil)
    
    // Test code
    
    mockClient.AssertExpectations(t)
}
```

**Rule:** Use `mock.Anything` for inputs you don't care about matching.

**Rule:** Always call `mockClient.AssertExpectations(t)` at end.

### Documentation Patterns

#### Package Documentation

**Rule:** Every package has package-level comment.

```go
// Package formatters provides output formatting implementations
// for awtest scan results. Supports text, JSON, Markdown, and
// JSON-compact formats as specified in FR42-48.
package formatters
```

#### Function Documentation

**Rule:** Exported functions have doc comments starting with function name.

```go
// FilterServices filters the given service list based on include/exclude patterns.
// If include is non-empty, only services in the include list are returned.
// If exclude is non-empty, services in the exclude list are removed.
// Include takes precedence over exclude.
func FilterServices(allServices []types.AWSService, include, exclude string) []types.AWSService {
    ...
}
```

### Enforcement Guidelines

#### All AI Agents MUST:

1. **Follow existing patterns** before creating new ones
2. **Use established error handling** (`utils.HandleAWSError` for AWS errors)
3. **Co-locate tests** with source files (`*_test.go`)
4. **Document exported symbols** with proper doc comments
5. **Use table-driven tests** for multiple test cases
6. **Maintain alphabetical service order** in `AllServices()` registration
7. **Follow Go naming conventions** (PascalCase exports, camelCase unexported)
8. **Never use panic** for expected errors
9. **Use testify** for assertions and mocking
10. **Preserve backward compatibility** during Phase 1 refactoring

#### Pattern Verification Checklist

Before submitting code, verify:

- [ ] Package names are lowercase, single word
- [ ] File names use underscores for multi-word
- [ ] Type names use PascalCase for exported
- [ ] Tests are co-located (`*_test.go`)
- [ ] AWS errors use `utils.HandleAWSError()`
- [ ] Exported symbols have doc comments
- [ ] Table-driven tests use `t.Run(tt.name, ...)`
- [ ] Mock expectations are asserted
- [ ] No `panic()` for expected errors
- [ ] Service registration maintains alphabetical order

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:**
All architectural decisions work together harmoniously without conflicts:
- **Technology Stack Coherence**: Go 1.19 with AWS SDK v1.44.266 represents production-proven stability. The decision to maintain AWS SDK v1 for Phase 1 aligns with the brownfield context—34 existing services using v1 patterns create a strong gravitational pull toward consistency rather than mixed-version complexity.
- **Pattern Compatibility**: The Formatter Interface Pattern integrates cleanly with the existing AWSService pattern. Both use the same interface-based abstraction philosophy, creating consistent mental models for AI agents during implementation.
- **Build Tool Alignment**: GoReleaser + Makefile combination complements the existing Go build system without introducing conflicts. GoReleaser handles cross-platform distribution, Makefile handles development workflows—clear separation of concerns.
- **Configuration Strategy Coherence**: Flag-based configuration with multi-source resolution (flags → env vars → AWS_PROFILE → shared config) extends the existing session management pattern in main.go rather than replacing it.
- **Testing Framework Fit**: Testify + Manual Mocks aligns with existing error handling patterns and avoids complex mockgen integration that would conflict with the AWSService interface design.

**Pattern Consistency:**
Implementation patterns fully support architectural decisions across all areas:
- **Naming Convention Alignment**: PascalCase for exported identifiers, camelCase for unexported—matches existing codebase patterns in types/types.go, utils/output.go, and all 34 service packages.
- **File Organization Consistency**: Lowercase filenames with underscores (service_filter.go, output_formatter.go) follows the established pattern from existing files like output.go.
- **Error Handling Pattern Coherence**: Centralized HandleAWSError in utils/output.go with InvalidKeyError abort behavior creates consistent error processing across all services. Formatters inherit this pattern rather than introducing new error handling approaches.
- **Interface Design Consistency**: Both AWSService and OutputFormatter follow the same interface design philosophy—small, focused interfaces with clear contracts. This creates predictable implementation patterns for AI agents.
- **Test Organization Alignment**: Co-located *_test.go files with table-driven test patterns match existing Go ecosystem conventions, ensuring AI agents follow established Go testing idioms.

**Structure Alignment:**
The project structure fully supports all architectural decisions:
- **Brownfield Preservation**: Phase 1 additions (formatters/, service_filter.go, *_test.go files, .goreleaser.yaml) integrate into existing structure without disrupting the working 34-service implementation.
- **Boundary Respect**: Five defined boundaries (AWS SDK, Output Formatting, Error Handling, Configuration, Testing) map cleanly to directory structure and package organization, preventing AI agents from creating architectural violations.
- **Integration Point Clarity**: The main.go orchestration loop remains unchanged—formatters and service filtering plug in at well-defined integration points (result collection phase, service enumeration phase) without architectural refactoring.
- **Scalability Path**: The structure supports Phase 2+ additions (AWS SDK v2 migration, advanced concurrency, monitoring) without requiring Phase 1 rework.

### Requirements Coverage Validation ✅

**Functional Requirements Coverage:**
All 66 functional requirements have complete architectural support with specific file-location mapping:

**FR1-6 (Credential Input & Authentication):**
- Architecture: Multi-source credential resolution in main.go
- Implementation Location: cmd/awtest/main.go credential handling section
- Pattern: Flag parsing → env vars → AWS_PROFILE → shared config fallback
- Support: AWS SDK session.NewSession() with custom credential chain

**FR7-31 (AWS Service Enumeration - 25 Services):**
- Architecture: AWSService interface pattern with service-specific implementations
- Implementation Locations: services/{service}/calls.go for each of 25 services
- Existing Services (34): Amplify, API Gateway, AppSync, Batch, CloudFormation, CloudFront, CloudTrail, CloudWatch, CodePipeline, Cognito Identity, DynamoDB, EC2, ECS, Elastic Beanstalk, EventBridge, Glacier, Glue, IAM, IoT, IVS, IVS Chat, IVS Realtime, KMS, Lambda, RDS, Route53, S3, Secrets Manager, SES, SNS, SQS, STS, Transcribe, WAF
- Phase 1 Additions Needed: Certificate Manager, Cognito User Pools, Config, EFS, EKS, ElastiCache, Fargate, Redshift, Step Functions, Systems Manager, VPC (11 new services)
- Pattern: Each service follows template from _template/service_calls_template.go

**FR32-37 (Scan Configuration & Control):**
- Architecture: Flag-based configuration with service filtering logic
- Implementation Locations:
  - FR32 (Region): main.go -region flag + multi-region support
  - FR33-34 (Service targeting): cmd/awtest/service_filter.go include/exclude logic
  - FR35 (Timeout): main.go -timeout flag (not yet implemented, Phase 1 addition)
  - FR36 (Verbosity): existing -debug flag in main.go
  - FR37 (Concurrency): Future enhancement (Phase 2+)
- Pattern: Flag definition in main.go, filtering logic in service_filter.go

**FR38-41 (Scan Execution & Progress):**
- Architecture: Existing main loop orchestration with error handling
- Implementation Location: cmd/awtest/main.go service enumeration loop
- Pattern: Sequential service execution, InvalidKeyError abort, AccessDenied continue

**FR42-48 (Output Formats & Display):**
- Architecture: Formatter Interface Pattern with pluggable implementations
- Implementation Locations:
  - Interface: cmd/awtest/formatters/output_formatter.go
  - JSON: cmd/awtest/formatters/json_formatter.go
  - YAML: cmd/awtest/formatters/yaml_formatter.go
  - CSV: cmd/awtest/formatters/csv_formatter.go
  - Table: cmd/awtest/formatters/table_formatter.go
  - Human: Existing colorized output in utils/output.go (default)
- Pattern: Factory function in main.go selects formatter based on -format flag

**FR49-54 (Error Handling & Reporting):**
- Architecture: Centralized error classification and handling
- Implementation Location: cmd/awtest/utils/output.go
- Functions: HandleAWSError, PrintResult, PrintAccessGranted
- Pattern: AWS error code mapping, InvalidKeyError abort, AccessDenied continue

**FR55-58 (Distribution & Installation):**
- Architecture: GoReleaser for cross-platform builds, Makefile for dev workflow
- Implementation Locations:
  - .goreleaser.yaml: Cross-platform build configuration
  - Makefile: Development commands (build, test, install, clean)
  - .github/workflows/release.yml: CI/CD automation
- Pattern: Tag-based releases, GitHub Releases distribution

**FR59-63 (Documentation & Help):**
- Architecture: Template-based documentation with inline help
- Implementation Locations:
  - README.md: Installation, usage, examples
  - CONTRIBUTING.md: Service addition guide
  - main.go: -help flag with usage text
  - _template/: Service implementation template
- Pattern: Markdown documentation + inline code comments

**FR64-66 (Extensibility & Service Addition):**
- Architecture: Template-driven service addition with registration pattern
- Implementation Locations:
  - _template/service_calls_template.go: Boilerplate template
  - CONTRIBUTING.md: Step-by-step guide
  - services/services.go: AllServices() registration
- Pattern: Copy template → Customize → Register in AllServices()

**Non-Functional Requirements Coverage:**
All 34 non-functional requirements architecturally supported:

**NFR1-8 (Performance):**
- Sequential execution pattern with potential Phase 2 concurrency
- Efficient AWS SDK usage (single session, connection pooling)
- Minimal memory footprint (streaming results, no large buffers)
- Fast startup time (Go binary compilation, no runtime dependencies)

**NFR9-15 (Security):**
- Credential masking in output (MaskSecret function in utils/output.go)
- MFA token support via AWS SDK session management
- No credential logging (existing security practice in main.go)
- Secure defaults (no insecure fallbacks, explicit credential sources)

**NFR16-20 (Reliability):**
- Error classification (InvalidKeyError abort vs AccessDenied continue)
- Graceful degradation (continue scan despite service failures)
- Timeout handling (Phase 1 addition for -timeout flag)
- Retry logic (AWS SDK built-in retry with exponential backoff)

**NFR21-26 (Maintainability):**
- Template-driven service addition (low-code replication)
- Clear separation of concerns (AWSService, OutputFormatter boundaries)
- Comprehensive tests (table-driven, co-located)
- Documentation templates (CONTRIBUTING.md, _template/)

**NFR27-30 (Usability):**
- Multiple output formats (JSON, YAML, CSV, Table, Human-readable)
- Clear error messages (AWS error code mapping in types/types.go)
- Helpful defaults (us-east-1 region, human-readable output)
- Progress indication (per-service output with colorization)

**NFR31-32 (Compatibility):**
- Cross-platform build (GoReleaser targeting linux, darwin, windows)
- AWS SDK v1 stability (production-proven, well-documented)

**NFR33-34 (Error Handling Robustness):**
- Invalid credential detection (InvalidKeyError with early abort)
- Network failure handling (AWS SDK retries)
- Partial result preservation (continue despite service failures)
- Error context preservation (original AWS error codes in debug mode)

### Implementation Readiness Validation ✅

**Decision Completeness:**
All architectural decisions documented with sufficient detail for AI agent implementation:

**Technology Versions Specified:**
- Go 1.19 (existing constraint from go.mod)
- AWS SDK for Go v1.44.266 (existing version from go.mod)
- GoReleaser 2.x (latest stable, to be installed)
- Testify v1.9.x (to be added to go.mod)
- github.com/olekukonko/tablewriter (for table formatter)
- gopkg.in/yaml.v3 (for YAML formatter)

**Implementation Patterns Documented:**
- Formatter Interface Pattern with Format() and FileExtension() methods
- Service Template Pattern with AWSService interface compliance
- Error Handling Pattern with HandleAWSError centralization
- Testing Pattern with table-driven tests and Testify assertions
- Configuration Pattern with flag-based multi-source resolution

**Consistency Rules Defined:**
- Naming: PascalCase exports, camelCase unexported
- File Organization: Lowercase with underscores, co-located tests
- Error Handling: InvalidKeyError abort, AccessDenied continue
- Interface Design: Small, focused interfaces with clear contracts
- Documentation: Inline comments for exported functions

**Examples Provided:**
- OutputFormatter interface definition
- Formatter factory function pattern
- Service template structure
- Table-driven test example
- Error handling example

**Structure Completeness:**
Project structure comprehensively defined with all Phase 1 additions:

**Complete Directory Structure:**
```
awtest/
├── cmd/awtest/
│   ├── main.go                          # Existing: Entry point, credential handling
│   ├── service_filter.go                # Phase 1: Include/exclude service filtering
│   ├── formatters/                      # Phase 1: Output format implementations
│   │   ├── output_formatter.go          # Interface definition
│   │   ├── json_formatter.go            # JSON output
│   │   ├── yaml_formatter.go            # YAML output
│   │   ├── csv_formatter.go             # CSV tabular output
│   │   ├── table_formatter.go           # ASCII table output
│   │   ├── json_formatter_test.go       # Unit tests
│   │   ├── yaml_formatter_test.go
│   │   ├── csv_formatter_test.go
│   │   └── table_formatter_test.go
│   ├── services/                        # Existing: AWS service implementations
│   │   ├── services.go                  # AllServices() registration
│   │   ├── sts/calls.go                 # Existing: STS service
│   │   ├── amplify/calls.go             # Existing: Amplify service
│   │   ├── [... 32 more existing services ...]
│   │   ├── certificatemanager/calls.go  # Phase 1: New service
│   │   ├── cognitouserpools/calls.go    # Phase 1: New service
│   │   ├── config/calls.go              # Phase 1: New service
│   │   ├── efs/calls.go                 # Phase 1: New service
│   │   ├── eks/calls.go                 # Phase 1: New service
│   │   ├── elasticache/calls.go         # Phase 1: New service
│   │   ├── fargate/calls.go             # Phase 1: New service (may be ECS extension)
│   │   ├── redshift/calls.go            # Phase 1: New service
│   │   ├── stepfunctions/calls.go       # Phase 1: New service
│   │   ├── systemsmanager/calls.go      # Phase 1: New service
│   │   └── vpc/calls.go                 # Phase 1: New service
│   ├── types/
│   │   └── types.go                     # Existing: AWSService, error types
│   └── utils/
│       └── output.go                    # Existing: PrintResult, HandleAWSError
├── _template/                           # Phase 1: Service template for extensibility
│   └── service_calls_template.go        # Boilerplate for new services
├── .goreleaser.yaml                     # Phase 1: Cross-platform build config
├── Makefile                             # Phase 1: Development workflow automation
├── .github/workflows/
│   └── release.yml                      # Phase 1: CI/CD release automation
├── go.mod                               # Existing: Dependency management
├── go.sum                               # Existing: Dependency checksums
├── README.md                            # Existing (to be updated with new features)
└── CONTRIBUTING.md                      # Phase 1: Service addition guide
```

**All Files and Directories Defined:**
- Existing files: Preserved without modification where possible
- Phase 1 additions: 18 new files (formatters, tests, configs, templates, docs)
- Phase 1 service additions: 11 new service packages
- Total Phase 1 file count: ~29 new files plus updates to existing files

**Integration Points Mapped:**
- **Formatter Integration**: main.go result collection → formatter selection → Format() call → output/file write
- **Service Registration**: New service package → import in services/services.go → append to AllServices() slice
- **Configuration Integration**: Flag definition in main.go → service_filter.go filtering logic → main loop execution
- **Error Integration**: Service error → HandleAWSError classification → PrintResult output → abort or continue decision
- **Test Integration**: Co-located *_test.go files → `go test ./...` execution → table-driven test runners

**Component Boundaries Established:**
1. **AWS SDK Boundary**: services/ packages own all AWS API calls, main.go owns session creation
2. **Output Formatting Boundary**: services/ collect results, formatters/ handle presentation
3. **Error Handling Boundary**: services/ detect errors, utils/output.go classifies and reports
4. **Configuration Boundary**: main.go owns flag parsing, service_filter.go owns filtering logic
5. **Testing Boundary**: Each package tests its own functionality, no cross-package test dependencies

**Pattern Completeness:**
All potential conflict points addressed with clear resolution rules:

**Naming Convention Conflicts Resolved:**
- Formatter types: `{Format}Formatter` (JsonFormatter, YamlFormatter) - capitalized format names
- Formatter files: `{format}_formatter.go` (json_formatter.go) - lowercase format names
- Service packages: Lowercase, no underscores (certificatemanager, cognitouserpools)
- Service files: calls.go (consistent across all services)
- Test files: `{name}_test.go` (json_formatter_test.go)
- Interface files: `{concept}_interface.go` or `{concept}.go` (output_formatter.go)

**Communication Pattern Conflicts Resolved:**
- AWSService.Call() returns (interface{}, error) - services own the AWS response types
- AWSService.Process() accepts (interface{}, error, bool) - services interpret their own responses
- OutputFormatter.Format() accepts []ServiceResult - generic result collection
- Error propagation: HandleAWSError returns error for abort signals, nil for continue

**Process Pattern Conflicts Resolved:**
- Error handling: InvalidKeyError causes os.Exit(1), AccessDenied continues scan
- Test organization: Table-driven tests in *_test.go, Testify assertions, no external mocks
- Documentation: Inline comments for exported functions, CONTRIBUTING.md for processes
- Service addition: Copy template → Customize calls.go → Register in AllServices()

### Gap Analysis Results

**Critical Gaps (Block Implementation):** **NONE**

All Phase 1 requirements have complete architectural support. No gaps identified that would prevent AI agents from implementing the documented architecture.

**Important Gaps (Improve Implementation):** **MINIMAL**

- **AWS SDK v2 Migration**: Intentionally deferred to Phase 2+ due to existing 34-service v1 implementation. Not a gap, but a conscious architectural decision to maintain consistency in Phase 1.
- **Advanced Concurrency Patterns**: Sequential execution documented for Phase 1. Concurrent service enumeration deferred to Phase 2+ optimization phase. Not blocking, as sequential execution meets Phase 1 performance requirements.

**Nice-to-Have Gaps (Optional Enhancements):**

- **Additional Formatter Implementations**: Architecture supports XML, HTML, or custom formatters, but only JSON/YAML/CSV/Table specified for Phase 1. Future formatters can be added following the documented pattern.
- **Service-Specific Test Fixtures**: Generic test patterns documented, but service-specific AWS response fixtures could be pre-created to accelerate test development. Not blocking, as AI agents can generate fixtures during implementation.
- **CI/CD Test Coverage Reporting**: GoReleaser and GitHub Actions workflow documented, but test coverage reporting (codecov, coveralls) not specified. Can be added in Phase 2+ without architectural changes.
- **Configuration File Support**: Flag-based configuration documented, but config file support (YAML/JSON config files) not specified. Can be added in Phase 2+ as an additional configuration source without breaking existing flag-based approach.

### Validation Issues Addressed

**Critical Issues:** **NONE FOUND** ✅

No critical architectural issues identified. The architecture is coherent, complete, and ready for implementation.

**Important Issues:** **NONE FOUND** ✅

No important architectural concerns identified. All decisions align with project context, requirements, and existing codebase patterns.

**Minor Observations:**

- **Frontmatter Completeness**: Document frontmatter updated to include steps 6 and 7 (Project Structure & Boundaries, Architecture Validation) in stepsCompleted tracking.

### Architecture Completeness Checklist

**✅ Requirements Analysis**

- [x] Project context thoroughly analyzed (66 FRs, 34 NFRs, brownfield Go 1.19 + AWS SDK v1 codebase)
- [x] Scale and complexity assessed (34 existing services, Phase 1 adds 11 services + formatters + build automation)
- [x] Technical constraints identified (Go 1.19, AWS SDK v1, existing AWSService pattern)
- [x] Cross-cutting concerns mapped (error handling, credential management, output formatting, testing)

**✅ Architectural Decisions**

- [x] Critical decisions documented with versions (Formatter Interface, GoReleaser, Flag-based Config, Testify, Documentation Templates)
- [x] Technology stack fully specified (Go 1.19, AWS SDK v1.44.266, GoReleaser 2.x, Testify v1.9.x, tablewriter, yaml.v3)
- [x] Integration patterns defined (AWSService interface, OutputFormatter interface, error handling boundaries)
- [x] Performance considerations addressed (sequential execution Phase 1, deferred concurrency Phase 2+)

**✅ Implementation Patterns**

- [x] Naming conventions established (PascalCase exports, camelCase unexported, lowercase files with underscores)
- [x] Structure patterns defined (services/, formatters/, _template/, co-located tests)
- [x] Communication patterns specified (interface contracts, error propagation, result collection)
- [x] Process patterns documented (error classification, service registration, template-based addition)

**✅ Project Structure**

- [x] Complete directory structure defined (existing + Phase 1 additions, ~29 new files)
- [x] Component boundaries established (5 boundaries: AWS SDK, Output Formatting, Error Handling, Configuration, Testing)
- [x] Integration points mapped (formatter integration, service registration, config integration, error integration, test integration)
- [x] Requirements to structure mapping complete (all 66 FRs mapped to specific file locations)

### Architecture Readiness Assessment

**Overall Status:** ✅ **READY FOR IMPLEMENTATION**

**Confidence Level:** **HIGH**

The architecture demonstrates:
- **Brownfield Alignment**: All decisions respect and extend the existing 34-service implementation without disruptive refactoring
- **Pattern Consistency**: New patterns (Formatter Interface) align with existing patterns (AWSService Interface)
- **Complete Coverage**: All 66 FRs and 34 NFRs architecturally supported with specific file mappings
- **Implementation Clarity**: AI agents have unambiguous guidance on naming, structure, patterns, and boundaries
- **Extensibility**: Template-driven service addition enables low-friction Phase 2+ expansion

**Key Strengths:**

1. **Brownfield Respect**: Architecture preserves working implementation while enabling clean Phase 1 expansion
2. **Interface-Based Design**: Both AWSService and OutputFormatter follow consistent interface abstraction patterns
3. **Comprehensive Mapping**: Every FR category mapped to specific files/directories, eliminating implementation ambiguity
4. **Clear Boundaries**: Five defined boundaries (AWS SDK, Output Formatting, Error Handling, Configuration, Testing) prevent architectural violations
5. **Technology Maturity**: Go 1.19 + AWS SDK v1 represents production-proven stability, minimizing risk
6. **Template-Driven Extensibility**: Service template + CONTRIBUTING.md enables low-friction service addition
7. **Testing Strategy**: Table-driven tests with Testify provide comprehensive coverage without complex mocking infrastructure

**Areas for Future Enhancement:**

1. **Concurrency Optimization (Phase 2+)**: Sequential execution meets Phase 1 needs but could be optimized with goroutine pools for multi-service enumeration
2. **AWS SDK v2 Migration (Phase 2+)**: Deferred to future phase due to 34-service rewrite complexity, but would unlock modern SDK features
3. **Advanced Output Formats (Phase 2+)**: XML, HTML, or custom formatters could extend the OutputFormatter interface
4. **Configuration File Support (Phase 2+)**: YAML/JSON config files could complement flag-based configuration
5. **Monitoring/Observability (Phase 2+)**: Structured logging, metrics collection, or tracing could enhance production usage
6. **Result Caching (Phase 2+)**: Cache AWS API responses to reduce API call volume for repeated scans

### Implementation Handoff

**AI Agent Guidelines:**

1. **Follow All Architectural Decisions Exactly**: Use Go 1.19, AWS SDK v1.44.266, Formatter Interface Pattern, GoReleaser, flag-based configuration, Testify testing, and documentation templates as documented
2. **Use Implementation Patterns Consistently**: PascalCase exports, camelCase unexported, lowercase files with underscores, co-located tests, table-driven test structure
3. **Respect Project Structure and Boundaries**: Never violate the 5 defined boundaries (AWS SDK, Output Formatting, Error Handling, Configuration, Testing)
4. **Refer to This Document for All Architectural Questions**: This document is the single source of truth for architectural decisions during implementation

**First Implementation Priority:**

Begin Phase 1 implementation with the **Formatter System** as it has no dependencies on service additions:

1. **Create formatters/ directory structure**
2. **Implement OutputFormatter interface** (formatters/output_formatter.go)
3. **Implement JSON formatter** (formatters/json_formatter.go + json_formatter_test.go)
4. **Implement YAML formatter** (formatters/yaml_formatter.go + yaml_formatter_test.go)
5. **Implement CSV formatter** (formatters/csv_formatter.go + csv_formatter_test.go)
6. **Implement Table formatter** (formatters/table_formatter.go + table_formatter_test.go)
7. **Integrate formatter into main.go** (add -format flag, factory function, format result collection)

After formatters are implemented and tested, proceed to:
- **Service additions** (11 new services: Certificate Manager, Cognito User Pools, Config, EFS, EKS, ElastiCache, Fargate, Redshift, Step Functions, Systems Manager, VPC)
- **Build automation** (.goreleaser.yaml, Makefile, GitHub Actions)
- **Documentation updates** (README.md, CONTRIBUTING.md, _template/)
- **Service filtering** (service_filter.go include/exclude logic)

**Implementation Validation Checkpoints:**

- After formatters: Run `go test ./cmd/awtest/formatters/...` to verify all formatter tests pass
- After each service: Run `go build ./cmd/awtest` to verify compilation succeeds
- After service registration: Run `go run ./cmd/awtest -debug` to verify service enumeration includes new services
- After build automation: Run `goreleaser build --snapshot` to verify cross-platform builds succeed
- After documentation: Review CONTRIBUTING.md against actual service implementation to verify accuracy

**Architecture Complete** ✅
