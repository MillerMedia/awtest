---
stepsCompleted: [1, 2, 3]
inputDocuments:
  - prd.md
  - architecture.md
---

# awtest - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for awtest, decomposing the requirements from the PRD, UX Design if it exists, and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

**FR1:** Users can provide AWS Access Key ID and Secret Access Key via command-line flags
**FR2:** Users can authenticate using active AWS CLI profile credentials without providing explicit credentials
**FR3:** Users can use temporary AWS STS credentials for enumeration
**FR4:** System validates credentials before starting enumeration
**FR5:** System displays Account ID and User ARN for authenticated credentials
**FR6:** System never logs or outputs full credential values in scan results

**FR7:** System enumerates accessible S3 buckets
**FR8:** System enumerates EC2 instances and associated resources
**FR9:** System enumerates RDS database instances
**FR10:** System enumerates Lambda functions
**FR11:** System enumerates IAM users, roles, and policies
**FR12:** System enumerates DynamoDB tables
**FR13:** System enumerates Secrets Manager secrets
**FR14:** System enumerates KMS encryption keys
**FR15:** System enumerates ECS/EKS/Fargate container services
**FR16:** System enumerates ElastiCache clusters
**FR17:** System enumerates Redshift data warehouses
**FR18:** System enumerates CloudFormation stacks
**FR19:** System enumerates CloudWatch log groups and streams
**FR20:** System enumerates SNS topics and SQS queues
**FR21:** System enumerates API Gateway endpoints
**FR22:** System enumerates CloudFront distributions
**FR23:** System enumerates Route53 hosted zones
**FR24:** System enumerates Step Functions state machines
**FR25:** System enumerates EventBridge event buses
**FR26:** System enumerates Certificate Manager certificates
**FR27:** System enumerates Cognito user pools
**FR28:** System enumerates EBS volumes and EFS file systems
**FR29:** System enumerates Glacier vaults
**FR30:** System enumerates CloudTrail trails and Config recorders
**FR31:** System enumerates Systems Manager parameters

**FR32:** Users can specify target AWS region for enumeration
**FR33:** Users can select specific services to scan
**FR34:** Users can exclude specific services from scans
**FR35:** Users can set maximum scan timeout duration
**FR36:** Users can enable verbose debug logging
**FR37:** Users can configure concurrency level for parallel service scanning

**FR38:** System provides resource-level details for each discovered resource (not just permission existence)
**FR39:** System categorizes findings by severity level (info, warning, critical)
**FR40:** System reports scan metadata including timestamp, region, and scan duration
**FR41:** System distinguishes between accessible resources and access-denied services

**FR42:** Users receive scan results in human-readable text format by default
**FR43:** Users can export scan results in structured JSON format
**FR44:** Users can export scan results in Markdown format optimized for LLM consumption
**FR45:** Users can write scan results to specified file
**FR46:** Users can suppress informational messages and output only findings (quiet mode)
**FR47:** System displays real-time scan progress during execution
**FR48:** System provides findings summary at scan completion

**FR49:** System distinguishes between access-denied errors and service-unavailable errors
**FR50:** System handles AWS API throttling without crashing
**FR51:** System reports invalid or revoked credentials with clear messaging
**FR52:** System continues enumeration when individual service checks fail
**FR53:** System provides actionable error messages including service name and error type
**FR54:** System handles region-unavailable scenarios gracefully

**FR55:** Users can install the tool via Homebrew package manager
**FR56:** Users can install the tool via go install command
**FR57:** System runs as single static binary without external dependencies
**FR58:** System supports macOS (Intel and ARM), Linux (amd64 and arm64), and Windows (amd64) platforms

**FR59:** System completes standard scans in under 2 minutes
**FR60:** System completes exhaustive scans across all services in under 5 minutes
**FR61:** System produces zero false positives in resource detection
**FR62:** System maintains memory footprint under 100MB during execution
**FR63:** System executes multiple service checks concurrently without blocking

**FR64:** Contributors can add new AWS service enumeration following documented patterns
**FR65:** System provides clear service implementation template for new services
**FR66:** System validates contributed service modules for consistency with existing patterns

### NonFunctional Requirements

**NFR1:** Standard scans complete in under 2 minutes for credentials with access to 20+ services
**NFR2:** Exhaustive scans across all supported services complete in under 5 minutes
**NFR3:** Tool startup (initialization to first API call) completes in under 1 second
**NFR4:** Concurrent service enumeration (Phase 2) reduces scan time by at least 60% compared to sequential execution
**NFR5:** Memory consumption remains under 100MB during execution regardless of scan scope
**NFR6:** Tool handles AWS API rate limiting gracefully without user intervention, implementing exponential backoff automatically

**NFR7:** Tool operates in read-only mode—never creates, modifies, or deletes AWS resources
**NFR8:** Credential values (access keys, secrets) are never logged to stdout, stderr, or files
**NFR9:** Only Account ID and User ARN are displayed in scan output for identification purposes
**NFR10:** Tool supports temporary STS credentials and respects credential expiration
**NFR11:** Command-line credential parameters (--aki, --sak) are cleared from shell history recommendations in documentation
**NFR12:** No credential data is transmitted outside AWS API endpoints (no analytics, telemetry, or external calls)

**NFR13:** Tool produces zero false positives—any reported accessible resource is actually accessible
**NFR14:** Individual service enumeration failures do not crash the entire scan
**NFR15:** Tool continues enumeration across all services even when some services return errors
**NFR16:** Error messages distinguish between access-denied, service-unavailable, invalid-credentials, and throttling scenarios
**NFR17:** Tool handles network interruptions without data corruption or incomplete state
**NFR18:** Concurrent operations (Phase 2) use proper synchronization to prevent race conditions

**NFR19:** Tool integrates with AWS SDK for Go following official SDK best practices and patterns
**NFR20:** JSON output format conforms to standard JSON schema for programmatic parsing
**NFR21:** Markdown output format is optimized for LLM consumption with clear structure and semantic sections
**NFR22:** Tool respects AWS CLI profile configuration (region, output format preferences) when using profile credentials
**NFR23:** Exit codes follow UNIX conventions (0 for success, non-zero for errors) for script integration

**NFR24:** Codebase follows Go standard project layout and idiomatic Go patterns
**NFR25:** Each AWS service implementation follows consistent interface pattern for community contributions
**NFR26:** New service additions require no changes to core enumeration engine
**NFR27:** Code coverage exceeds 70% with unit tests for all service modules
**NFR28:** Contribution documentation enables new contributors to add services without maintainer guidance
**NFR29:** Pull requests are reviewed and merged within 48 hours for quality contributions

**NFR30:** Single static binary runs without external dependencies on all supported platforms
**NFR31:** Tool supports macOS (Intel x86_64 and Apple Silicon arm64), Linux (amd64 and arm64), and Windows (amd64)
**NFR32:** Cross-compilation produces platform-specific binaries under 15MB each
**NFR33:** Tool behavior is consistent across all platforms (no platform-specific feature variations)
**NFR34:** Installation via Homebrew, go install, and direct binary download all produce identical functional tool

### Additional Requirements

**Architecture Technical Requirements:**

- **Language & Runtime:** Go 1.19 (brownfield constraint)
- **AWS SDK:** AWS SDK for Go v1.44.266 (existing version, Phase 1 maintains v1)
- **Build Tooling:** GoReleaser 2.x + Makefile for cross-platform builds and distribution
- **Testing Framework:** Testify v1.9.x for assertions and table-driven tests
- **Output Dependencies:**
  - github.com/olekukonko/tablewriter (for table formatter)
  - gopkg.in/yaml.v3 (for YAML formatter)

**Architectural Patterns:**

- **Formatter Interface Pattern:** OutputFormatter interface with Format() and FileExtension() methods for pluggable output formats
- **Service Template Pattern:** AWSService interface compliance for consistent service implementations
- **Error Handling Pattern:** Centralized HandleAWSError in utils/output.go
- **Configuration Pattern:** Flag-based multi-source credential resolution (flags → env vars → AWS_PROFILE → shared config)

**Project Structure Constraints:**

- **Brownfield Project:** Existing 34 AWS services already implemented
- **Phase 1 Service Additions (11 new services):**
  - Certificate Manager (ACM)
  - Cognito User Pools
  - AWS Config
  - EFS (Elastic File System)
  - EKS (Elastic Kubernetes Service)
  - ElastiCache
  - Fargate (may be ECS extension)
  - Redshift
  - Step Functions
  - Systems Manager (SSM)
  - VPC (Virtual Private Cloud)

**Architectural Boundaries (5 defined):**
1. AWS SDK Boundary - services/ packages own AWS API calls, main.go owns session creation
2. Output Formatting Boundary - services/ collect results, formatters/ handle presentation
3. Error Handling Boundary - services/ detect errors, utils/output.go classifies and reports
4. Configuration Boundary - main.go owns flag parsing, service_filter.go owns filtering logic
5. Testing Boundary - Each package tests its own functionality, no cross-package test dependencies

**Implementation Priorities (from Architecture):**

1. **First Priority:** Formatter System (no dependencies on service additions)
   - Create formatters/ directory structure
   - Implement OutputFormatter interface
   - Implement JSON, YAML, CSV, Table formatters with tests
   - Integrate formatters into main.go with -format flag

2. **Second Priority:** Service Additions (11 new services)

3. **Third Priority:** Build Automation (.goreleaser.yaml, Makefile, GitHub Actions)

4. **Fourth Priority:** Documentation Updates (README.md, CONTRIBUTING.md, _template/)

5. **Fifth Priority:** Service Filtering (service_filter.go include/exclude logic)

**Naming Conventions:**
- PascalCase for exported identifiers
- camelCase for unexported identifiers
- Lowercase filenames with underscores (e.g., service_filter.go, output_formatter.go)
- Co-located tests: *_test.go files
- Service packages: lowercase, no underscores (e.g., certificatemanager, cognitouserpools)

### FR Coverage Map

**Epic 1: Output Format System**
- FR42: Users receive scan results in human-readable text format by default
- FR43: Users can export scan results in structured JSON format
- FR44: Users can export scan results in Markdown format optimized for LLM consumption
- FR45: Users can write scan results to specified file
- FR46: Users can suppress informational messages and output only findings (quiet mode)
- FR47: System displays real-time scan progress during execution
- FR48: System provides findings summary at scan completion

**Epic 2: AWS Service Coverage Expansion (11 New Services)**
- FR13: System enumerates Secrets Manager secrets (existing - baseline)
- FR15: System enumerates ECS/EKS/Fargate container services (EKS & Fargate new)
- FR16: System enumerates ElastiCache clusters (new)
- FR17: System enumerates Redshift data warehouses (new)
- FR26: System enumerates Certificate Manager certificates (new)
- FR27: System enumerates Cognito user pools (new)
- FR28: System enumerates EBS volumes and EFS file systems (EFS new)
- FR30: System enumerates CloudTrail trails and Config recorders (Config new)
- FR31: System enumerates Systems Manager parameters (new)
- FR24: System enumerates Step Functions state machines (new)
- Note: VPC enumeration is implicit infrastructure discovery (new)

**Epic 3: Scan Configuration & Control**
- FR32: Users can specify target AWS region for enumeration
- FR33: Users can select specific services to scan
- FR34: Users can exclude specific services from scans
- FR35: Users can set maximum scan timeout duration
- FR36: Users can enable verbose debug logging
- FR37: Users can configure concurrency level for parallel service scanning

**Epic 4: Build Automation & Distribution**
- FR55: Users can install the tool via Homebrew package manager
- FR56: Users can install the tool via go install command
- FR57: System runs as single static binary without external dependencies
- FR58: System supports macOS (Intel and ARM), Linux (amd64 and arm64), and Windows (amd64) platforms

**Epic 5: Documentation & Community Contribution Framework**
- FR64: Contributors can add new AWS service enumeration following documented patterns
- FR65: System provides clear service implementation template for new services
- FR66: System validates contributed service modules for consistency with existing patterns

**Brownfield - Already Implemented (34 Existing Services):**
- FR1-6: Credential Input & Authentication (existing in main.go)
- FR7-12, FR14, FR18-23, FR25, FR29: Existing AWS services (S3, EC2, RDS, Lambda, IAM, DynamoDB, KMS, CloudFormation, CloudWatch, SNS, SQS, API Gateway, CloudFront, Route53, EventBridge, Glacier, etc.)
- FR38-41: Resource Discovery & Reporting (existing in service Process() methods)
- FR49-54: Error Handling & Status Communication (existing in utils/output.go)
- FR59-63: Performance & Reliability (existing architecture patterns)

**Cross-Cutting NFRs (Applied Across All Epics):**
- NFR1-6: Performance targets enforced through architecture and testing
- NFR7-12: Security practices enforced through code patterns
- NFR13-18: Reliability patterns enforced through error handling
- NFR19-23: Integration standards enforced through AWS SDK usage
- NFR24-29: Maintainability enforced through Go patterns and testing
- NFR30-34: Portability enforced through GoReleaser build configuration

## Epic List

### Epic 1: Output Format System
**Goal:** Security professionals can export scan results in multiple structured formats (JSON, Markdown, CSV, Table) for reporting tools, LLM analysis, and automated workflows.

**User Value:** Enables integration with reporting platforms, LLM-based analysis, spreadsheet investigation, and automated security workflows. Transforms awtest from a terminal-only tool to a data source for the entire security toolchain.

**FRs covered:** FR42, FR43, FR44, FR45, FR46, FR47, FR48

**Implementation Priority:** First (Architecture Priority #1)
- Has no dependencies on service additions
- Immediately benefits all 34 existing services
- Foundation for Epic 2 service output

**Technical Notes:**
- Formatter Interface Pattern: OutputFormatter with Format() and FileExtension()
- Result Collection Phase: Collect ScanResult structs instead of immediate printing
- Formatting Phase: Pass results to selected formatter based on -format flag
- Backward compatibility: Text formatter replicates existing PrintResult() behavior

---

### Epic 2: AWS Service Coverage Expansion
**Goal:** Pentesters and security professionals discover accessible resources across 11 additional high-value AWS services (Certificate Manager, Cognito User Pools, Config, EFS, EKS, ElastiCache, Fargate, Redshift, Step Functions, Systems Manager, VPC) that they wouldn't manually check during time-constrained engagements.

**User Value:** Comprehensive AWS coverage is awtest's primary differentiator. These 11 services are frequently encountered during security assessments and often contain critical findings (database credentials in Secrets Manager, PII in ElastiCache, infrastructure configs in Systems Manager, container orchestration in EKS).

**FRs covered:** FR13, FR15, FR16, FR17, FR24, FR26, FR27, FR28, FR30, FR31, plus VPC infrastructure discovery

**Implementation Priority:** Second (Architecture Priority #2)
- Builds on Epic 1's formatter system
- Core product value delivery
- Each service follows existing AWSService interface pattern

**Services to Add:**
1. Certificate Manager (ACM) - SSL/TLS certificates
2. Cognito User Pools - User authentication/authorization
3. AWS Config - Configuration compliance and tracking
4. EFS (Elastic File System) - Network file storage
5. EKS (Elastic Kubernetes Service) - Kubernetes clusters
6. ElastiCache - In-memory caching (Redis/Memcached)
7. Fargate - Serverless container compute (may extend ECS)
8. Redshift - Data warehouse clusters
9. Step Functions - Workflow orchestration state machines
10. Systems Manager (SSM) - Parameter store and operational data
11. VPC - Virtual Private Cloud infrastructure discovery

**Technical Notes:**
- Each service implements AWSService interface (Name, Call(), Process(), ModuleName)
- Follow existing service template pattern from _template/
- Register in services/services.go AllServices() function
- Co-located tests using table-driven pattern

---

### Epic 3: Scan Configuration & Control
**Goal:** Security professionals can customize scans for specific scenarios - targeting specific services, excluding noisy services, setting timeouts, and controlling scan scope to match engagement constraints.

**User Value:** Power users can optimize scans for specific use cases (fast triage scans targeting only critical services, comprehensive audits excluding known-noisy services, time-boxed scans for quick assessments). Enables workflow flexibility during time-sensitive engagements.

**FRs covered:** FR32, FR33, FR34, FR35, FR36, FR37

**Implementation Priority:** Third (After formatters and services are in place)
- Builds on Epic 2's expanded service coverage
- Enhances existing 34 services + 11 new services from Epic 2

**Technical Notes:**
- New flags: -services, -exclude-services, -timeout, -concurrency
- Service filtering logic in service_filter.go
- Include/exclude patterns for service targeting
- Concurrency flag prepared for Phase 2 optimization

---

### Epic 4: Build Automation & Distribution
**Goal:** Security professionals can easily install awtest via Homebrew on macOS/Linux, ensuring they always have the latest version with cross-platform binaries optimized for their system.

**User Value:** Frictionless installation ("brew install awtest") removes adoption barriers. Automatic updates ensure users benefit from Epic 2's new services without manual builds. Cross-platform binaries ensure consistent tool availability across pentesting environments.

**FRs covered:** FR55, FR56, FR57, FR58

**Implementation Priority:** Fourth (Architecture Priority #3)
- Packages completed Epic 1 formatters + Epic 2 services + Epic 3 configuration
- Enables community growth through easy distribution

**Technical Notes:**
- GoReleaser configuration (.goreleaser.yaml)
- Makefile for local development (build, test, install, clean)
- GitHub Actions workflow for automated releases
- Homebrew tap: MillerMedia/homebrew-tap
- Cross-platform builds: macOS (Intel/ARM), Linux (amd64/arm64), Windows (amd64)
- Version embedding via ldflags

---

### Epic 5: Documentation & Community Contribution Framework
**Goal:** Open-source contributors can add new AWS services to awtest following clear templates and contribution guidelines, ensuring the tool stays current as AWS releases new services.

**User Value:** Sustainability through community contributions. When a pentester encounters an AWS service awtest doesn't cover, they can add it themselves following documented patterns. Community-driven model ensures awtest evolves with AWS's service portfolio without requiring core maintainer intervention for every addition.

**FRs covered:** FR64, FR65, FR66

**Implementation Priority:** Fifth (Architecture Priority #4)
- Documents patterns established in Epics 1-3
- Enables future community growth

**Technical Notes:**
- CONTRIBUTING.md with step-by-step service addition guide
- _template/service_calls_template.go boilerplate template
- README.md updates with new features and examples
- Service addition validation checklist

---

## Epic 1: Output Format System

**Goal:** Security professionals can export scan results in multiple structured formats (JSON, Markdown, CSV, Table) for reporting tools, LLM analysis, and automated workflows.

### Story 1.1: Formatter Interface & Result Collection

As a **developer implementing awtest enhancements**,
I want **a clean formatter interface and result collection system**,
So that **scan results can be formatted in multiple output formats without coupling service logic to presentation**.

**Acceptance Criteria:**

**Given** the existing service enumeration architecture
**When** implementing the formatter system foundation
**Then** create the OutputFormatter interface in cmd/awtest/formatters/output_formatter.go with Format(results []ScanResult) (string, error) and FileExtension() string methods
**And** define the ScanResult struct in cmd/awtest/types/types.go with ServiceName, MethodName, ResourceType, ResourceName, Details map[string]interface{}, Error, and Timestamp fields
**And** modify the main.go service enumeration loop to collect results in a []ScanResult slice instead of immediate printing
**And** update service Process() methods to return ScanResult instead of printing directly
**And** write unit tests for result collection logic achieving >70% coverage
**And** ensure backward compatibility - existing service behavior is preserved
**And** verify compilation succeeds with `go build ./cmd/awtest`

### Story 1.2: JSON Output Formatter

As a **security professional using awtest in automated workflows**,
I want **JSON-formatted scan results**,
So that **I can programmatically parse output for SIEM integration, reporting tools, and custom analysis scripts**.

**Acceptance Criteria:**

**Given** the OutputFormatter interface from Story 1.1
**When** implementing JSON output format
**Then** create JsonFormatter struct in cmd/awtest/formatters/json_formatter.go implementing OutputFormatter interface
**And** Format() method produces valid JSON conforming to standard JSON schema (NFR20)
**And** JSON includes all ScanResult fields (service, method, resource type, resource name, details, timestamp)
**And** JSON structure uses camelCase field names following Go JSON conventions
**And** FileExtension() returns "json"
**And** handle empty results gracefully (return valid empty JSON array)
**And** handle errors in ScanResult.Error field by including error message in JSON output
**And** write table-driven unit tests in json_formatter_test.go covering: valid results, empty results, results with errors, timestamp formatting
**And** verify JSON output can be parsed by standard JSON tools (jq, Python json.loads)
**And** test passes: `go test ./cmd/awtest/formatters/...`

### Story 1.3: YAML Output Formatter

As a **security professional creating readable reports**,
I want **YAML-formatted scan results**,
So that **I can produce human-readable structured output for documentation and reports**.

**Acceptance Criteria:**

**Given** the OutputFormatter interface from Story 1.1
**When** implementing YAML output format
**Then** create YamlFormatter struct in cmd/awtest/formatters/yaml_formatter.go implementing OutputFormatter interface
**And** add gopkg.in/yaml.v3 dependency to go.mod
**And** Format() method produces valid YAML with proper indentation and structure
**And** YAML includes all ScanResult fields with human-readable formatting
**And** FileExtension() returns "yaml"
**And** handle empty results gracefully (return valid empty YAML)
**And** handle errors in ScanResult.Error field by including error message in YAML output
**And** write table-driven unit tests in yaml_formatter_test.go covering: valid results, empty results, results with errors, special characters in resource names
**And** verify YAML output can be parsed by standard YAML tools
**And** test passes: `go test ./cmd/awtest/formatters/...`

### Story 1.4: CSV Output Formatter

As a **security professional analyzing scan results in spreadsheets**,
I want **CSV-formatted scan results**,
So that **I can import findings into Excel/Google Sheets for filtering, sorting, and pivot table analysis**.

**Acceptance Criteria:**

**Given** the OutputFormatter interface from Story 1.1
**When** implementing CSV output format
**Then** create CsvFormatter struct in cmd/awtest/formatters/csv_formatter.go implementing OutputFormatter interface
**And** Format() method produces valid CSV with header row: Service,Method,ResourceType,ResourceName,Details,Timestamp,Error
**And** CSV escapes special characters (commas, quotes, newlines) following RFC 4180
**And** Details field flattens map[string]interface{} to comma-separated key:value pairs
**And** FileExtension() returns "csv"
**And** handle empty results gracefully (return CSV with header row only)
**And** handle errors in ScanResult.Error field by populating Error column
**And** write table-driven unit tests in csv_formatter_test.go covering: valid results, empty results, results with special characters requiring escaping, results with complex Details maps
**And** verify CSV output can be imported into Excel/Google Sheets without errors
**And** test passes: `go test ./cmd/awtest/formatters/...`

### Story 1.5: Table Output Formatter

As a **security professional viewing scan results in terminal**,
I want **ASCII table-formatted scan results**,
So that **I can quickly scan findings in a structured, readable table layout**.

**Acceptance Criteria:**

**Given** the OutputFormatter interface from Story 1.1
**When** implementing ASCII table output format
**Then** create TableFormatter struct in cmd/awtest/formatters/table_formatter.go implementing OutputFormatter interface
**And** add github.com/olekukonko/tablewriter dependency to go.mod
**And** Format() method produces ASCII table with columns: Service, Method, Resource Type, Resource Name, Timestamp
**And** table uses borders, proper alignment, and column wrapping for readability
**And** FileExtension() returns "txt"
**And** handle empty results gracefully (return "No results found" message)
**And** handle errors in ScanResult.Error field by including error indicator in table row
**And** limit table width to 120 characters for terminal readability
**And** write table-driven unit tests in table_formatter_test.go covering: valid results, empty results, long resource names requiring wrapping, results with errors
**And** verify table renders correctly in 80-column and 120-column terminals
**And** test passes: `go test ./cmd/awtest/formatters/...`

### Story 1.6: Format Selection & File Output

As a **security professional running awtest**,
I want **to select output format via command-line flag and optionally save to file**,
So that **I can choose the format that best fits my workflow and save results for later analysis**.

**Acceptance Criteria:**

**Given** all formatters implemented (Stories 1.2-1.5)
**When** integrating formatters into main.go
**Then** add -format flag accepting values: text, json, yaml, csv, table (FR42-44)
**And** add -output-file flag accepting file path (FR45)
**And** implement formatter factory function that returns appropriate formatter based on -format flag
**And** default format is "text" (human-readable, current behavior) when -format not specified (FR42)
**And** after service enumeration loop completes, pass collected results to selected formatter
**And** if -output-file specified, write formatted output to file with appropriate extension from FileExtension()
**And** if -output-file not specified, write formatted output to stdout
**And** preserve existing colorized output for "text" format when outputting to terminal
**And** handle file write errors gracefully with clear error messages
**And** validate -format flag value, error if invalid format specified
**And** write integration tests verifying: each format produces expected output, file output writes correctly, invalid format errors appropriately
**And** verify: `go run ./cmd/awtest -format=json` produces JSON output
**And** verify: `go run ./cmd/awtest -format=yaml -output-file=results.yaml` creates results.yaml file
**And** test compilation: `go build ./cmd/awtest`

### Story 1.7: Progress Tracking & Summary Reporting

As a **security professional running comprehensive scans**,
I want **real-time progress indicators and a findings summary**,
So that **I can see scan progress and quickly understand the overall results without reading every line**.

**Acceptance Criteria:**

**Given** the formatter system integrated from Story 1.6
**When** implementing progress and summary features
**Then** add -quiet flag to suppress informational messages, showing only findings (FR46)
**And** display real-time progress during scan showing: "Scanning [service_name]..." for each service (FR47)
**And** progress messages write to stderr (not stdout) so they don't interfere with formatted output
**And** suppress progress messages when -quiet flag is set
**And** after scan completes, display summary report showing: total services scanned, services with accessible resources, services with access denied, total resources discovered, scan duration, timestamp (FR48)
**And** summary report respects -quiet flag (suppress when quiet mode enabled)
**And** summary formatting adapts to selected output format (plain text for text/table, structured for JSON/YAML/CSV)
**And** write unit tests for progress tracking and summary generation
**And** verify: `go run ./cmd/awtest -quiet -format=json` shows only JSON output, no progress messages
**And** verify: `go run ./cmd/awtest` shows progress messages and summary report
**And** verify scan metadata includes timestamp and duration (FR40)
**And** test passes: `go test ./cmd/awtest/...`

---

## Epic 2: AWS Service Coverage Expansion

**Goal:** Pentesters and security professionals discover accessible resources across 11 additional high-value AWS services (Certificate Manager, Cognito User Pools, Config, EFS, EKS, ElastiCache, Fargate, Redshift, Step Functions, Systems Manager, VPC) that they wouldn't manually check during time-constrained engagements.

### Story 2.1: Certificate Manager (ACM) Service Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Certificate Manager certificates**,
So that **I can discover SSL/TLS certificates accessible with the credentials, which may reveal domains, internal infrastructure, and expiration risks**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern from 34 existing services
**When** implementing Certificate Manager enumeration
**Then** create cmd/awtest/services/certificatemanager/ directory
**And** implement calls.go with ListCertificates() API call using AWS SDK v1.44.266 ACM client
**And** implement AWSService interface with Name="Certificate Manager", Call(), Process(), ModuleName="CertificateManager"
**And** Call() method creates ACM client, calls ListCertificates, returns certificate list
**And** Process() method displays each certificate: ARN, DomainName, Status, InUseBy count
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("Certificate Manager", "certificates")
**And** register service in services/services.go AllServices() function in alphabetical order
**And** write table-driven tests in calls_test.go covering: valid certificates, empty results, access denied, API errors
**And** verify service follows naming conventions: package certificatemanager (lowercase, no underscores)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/certificatemanager/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "Certificate Manager" in enumeration output
**And** FR26 requirement fulfilled: System enumerates Certificate Manager certificates

### Story 2.2: Cognito User Pools Service Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Cognito user pools**,
So that **I can discover user authentication databases accessible with the credentials, which may contain user PII and authentication configurations**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing Cognito User Pools enumeration
**Then** create cmd/awtest/services/cognitouserpools/ directory
**And** implement calls.go with ListUserPools() API call using AWS SDK v1.44.266 Cognito Identity Provider client
**And** implement AWSService interface with Name="Cognito User Pools", Call(), Process(), ModuleName="CognitoUserPools"
**And** Call() method creates Cognito client, calls ListUserPools with MaxResults=60, handles pagination
**And** Process() method displays each user pool: Id, Name, CreationDate, LastModifiedDate, UserPoolTags
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("Cognito User Pools", "user pools")
**And** register service in services/services.go AllServices() after cognitoidentity, before dynamodb
**And** write table-driven tests in calls_test.go covering: valid user pools, pagination, empty results, access denied
**And** verify package naming: cognitouserpools (lowercase, no spaces/underscores)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/cognitouserpools/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "Cognito User Pools" in output
**And** FR27 requirement fulfilled: System enumerates Cognito user pools

### Story 2.3: AWS Config Service Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate AWS Config recorders and rules**,
So that **I can discover configuration compliance tracking accessible with the credentials, revealing monitored resources and compliance rules**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing AWS Config enumeration
**Then** create cmd/awtest/services/config/ directory
**And** implement calls.go with DescribeConfigurationRecorders() and DescribeConfigRules() API calls using AWS SDK v1.44.266 ConfigService client
**And** implement AWSService interface with Name="AWS Config", Call(), Process(), ModuleName="Config"
**And** Call() method creates Config client, calls both DescribeConfigurationRecorders and DescribeConfigRules
**And** Process() method displays configuration recorders: Name, RoleARN, RecordingGroup, and config rules: ConfigRuleName, ConfigRuleState, Source
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("AWS Config", "configuration recorders and rules")
**And** register service in services/services.go AllServices() after codepipeline, before cognitoidentity
**And** write table-driven tests in calls_test.go covering: valid config data, empty recorders, empty rules, access denied
**And** verify package naming: config (lowercase, single word)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/config/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "AWS Config" in output
**And** FR30 requirement partially fulfilled: System enumerates Config recorders (CloudTrail already exists)

### Story 2.4: EFS (Elastic File System) Service Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate EFS file systems**,
So that **I can discover network-attached storage accessible with the credentials, which may contain sensitive data and mount targets**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing EFS enumeration
**Then** create cmd/awtest/services/efs/ directory
**And** implement calls.go with DescribeFileSystems() API call using AWS SDK v1.44.266 EFS client
**And** implement AWSService interface with Name="EFS", Call(), Process(), ModuleName="EFS"
**And** Call() method creates EFS client, calls DescribeFileSystems
**And** Process() method displays each file system: FileSystemId, Name, LifeCycleState, SizeInBytes, NumberOfMountTargets, Encrypted
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("EFS", "file systems")
**And** register service in services/services.go AllServices() after ecs, before elasticbeanstalk
**And** write table-driven tests in calls_test.go covering: valid file systems, encrypted vs unencrypted, empty results, access denied
**And** verify package naming: efs (lowercase)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/efs/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "EFS" in output
**And** FR28 requirement partially fulfilled: System enumerates EFS file systems (EBS already exists)

### Story 2.5: EKS (Elastic Kubernetes Service) Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate EKS clusters**,
So that **I can discover Kubernetes clusters accessible with the credentials, revealing container orchestration infrastructure and potential lateral movement targets**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing EKS enumeration
**Then** create cmd/awtest/services/eks/ directory
**And** implement calls.go with ListClusters() and DescribeCluster() API calls using AWS SDK v1.44.266 EKS client
**And** implement AWSService interface with Name="EKS", Call(), Process(), ModuleName="EKS"
**And** Call() method creates EKS client, calls ListClusters to get cluster names, then DescribeCluster for each cluster
**And** Process() method displays each cluster: Name, Arn, Status, Version, Endpoint, RoleArn, ResourcesVpcConfig
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("EKS", "clusters")
**And** register service in services/services.go AllServices() after efs, before elasticbeanstalk
**And** write table-driven tests in calls_test.go covering: valid clusters, multiple clusters, empty results, access denied
**And** verify package naming: eks (lowercase)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/eks/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "EKS" in output
**And** FR15 requirement partially fulfilled: System enumerates EKS container services (ECS already exists, Fargate in Story 2.7)

### Story 2.6: ElastiCache Clusters Service Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate ElastiCache clusters**,
So that **I can discover Redis/Memcached caching databases accessible with the credentials, which often contain session data, cached credentials, and PII**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing ElastiCache enumeration
**Then** create cmd/awtest/services/elasticache/ directory
**And** implement calls.go with DescribeCacheClusters() API call using AWS SDK v1.44.266 ElastiCache client
**And** implement AWSService interface with Name="ElastiCache", Call(), Process(), ModuleName="ElastiCache"
**And** Call() method creates ElastiCache client, calls DescribeCacheClusters with ShowCacheNodeInfo=true
**And** Process() method displays each cluster: CacheClusterId, Engine (Redis/Memcached), EngineVersion, CacheNodeType, CacheClusterStatus, NumCacheNodes, PreferredAvailabilityZone
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("ElastiCache", "clusters")
**And** register service in services/services.go AllServices() after eks, before elasticbeanstalk
**And** write table-driven tests in calls_test.go covering: Redis clusters, Memcached clusters, empty results, access denied
**And** verify package naming: elasticache (lowercase)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/elasticache/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "ElastiCache" in output
**And** FR16 requirement fulfilled: System enumerates ElastiCache clusters

### Story 2.7: Fargate Tasks Service Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Fargate tasks**,
So that **I can discover serverless container workloads accessible with the credentials, revealing running containers and task definitions**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing Fargate enumeration
**Then** create cmd/awtest/services/fargate/ directory
**And** implement calls.go with ListClusters() and ListTasks() API calls using AWS SDK v1.44.266 ECS client (Fargate uses ECS API)
**And** implement AWSService interface with Name="Fargate", Call(), Process(), ModuleName="Fargate"
**And** Call() method creates ECS client, calls ListClusters, then ListTasks with launchType=FARGATE filter
**And** Process() method displays each Fargate task: TaskArn, ClusterArn, LaunchType=FARGATE, LastStatus, DesiredStatus, TaskDefinitionArn
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("Fargate", "tasks")
**And** register service in services/services.go AllServices() after eventbridge, before glacier
**And** write table-driven tests in calls_test.go covering: Fargate tasks in clusters, empty clusters, access denied
**And** verify package naming: fargate (lowercase)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/fargate/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "Fargate" in output
**And** FR15 requirement fully fulfilled: System enumerates Fargate container services (ECS exists, EKS in Story 2.5)

### Story 2.8: Redshift Clusters Service Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Redshift data warehouse clusters**,
So that **I can discover data warehouses accessible with the credentials, which often contain large volumes of analytics data and business intelligence**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing Redshift enumeration
**Then** create cmd/awtest/services/redshift/ directory
**And** implement calls.go with DescribeClusters() API call using AWS SDK v1.44.266 Redshift client
**And** implement AWSService interface with Name="Redshift", Call(), Process(), ModuleName="Redshift"
**And** Call() method creates Redshift client, calls DescribeClusters
**And** Process() method displays each cluster: ClusterIdentifier, NodeType, ClusterStatus, MasterUsername, DBName, Endpoint, Encrypted, NumberOfNodes
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("Redshift", "clusters")
**And** register service in services/services.go AllServices() after rds, before route53
**And** write table-driven tests in calls_test.go covering: valid clusters, encrypted vs unencrypted, empty results, access denied
**And** verify package naming: redshift (lowercase)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/redshift/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "Redshift" in output
**And** FR17 requirement fulfilled: System enumerates Redshift data warehouses

### Story 2.9: Step Functions State Machines Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Step Functions state machines**,
So that **I can discover workflow orchestration accessible with the credentials, revealing automated processes and integration patterns**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing Step Functions enumeration
**Then** create cmd/awtest/services/stepfunctions/ directory
**And** implement calls.go with ListStateMachines() API call using AWS SDK v1.44.266 Step Functions (SFN) client
**And** implement AWSService interface with Name="Step Functions", Call(), Process(), ModuleName="StepFunctions"
**And** Call() method creates SFN client, calls ListStateMachines with MaxResults=100, handles pagination
**And** Process() method displays each state machine: StateMachineArn, Name, Type (STANDARD/EXPRESS), Status, CreationDate
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("Step Functions", "state machines")
**And** register service in services/services.go AllServices() after sqs, before transcribe
**And** write table-driven tests in calls_test.go covering: STANDARD state machines, EXPRESS state machines, pagination, empty results, access denied
**And** verify package naming: stepfunctions (lowercase, no spaces)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/stepfunctions/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "Step Functions" in output
**And** FR24 requirement fulfilled: System enumerates Step Functions state machines

### Story 2.10: Systems Manager (SSM) Parameters Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate Systems Manager parameters**,
So that **I can discover configuration parameters and secrets accessible with the credentials, which often contain database credentials, API keys, and infrastructure configs**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing Systems Manager enumeration
**Then** create cmd/awtest/services/systemsmanager/ directory
**And** implement calls.go with DescribeParameters() API call using AWS SDK v1.44.266 SSM client
**And** implement AWSService interface with Name="Systems Manager", Call(), Process(), ModuleName="SystemsManager"
**And** Call() method creates SSM client, calls DescribeParameters with MaxResults=50, handles pagination
**And** Process() method displays each parameter: Name, Type (String/StringList/SecureString), Description, LastModifiedDate, Version (DO NOT retrieve values - read-only enumeration)
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("Systems Manager", "parameters")
**And** register service in services/services.go AllServices() after stepfunctions, before transcribe
**And** write table-driven tests in calls_test.go covering: String parameters, SecureString parameters, pagination, empty results, access denied
**And** verify package naming: systemsmanager (lowercase, no spaces/underscores)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/systemsmanager/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "Systems Manager" in output
**And** FR31 requirement fulfilled: System enumerates Systems Manager parameters
**And** verify NFR7 compliance: Read-only operation, parameter values NOT retrieved

### Story 2.11: VPC Infrastructure Service Enumeration

As a **security professional assessing AWS credentials**,
I want **awtest to enumerate VPC infrastructure**,
So that **I can discover network infrastructure accessible with the credentials, revealing VPCs, subnets, security groups, and network topology**.

**Acceptance Criteria:**

**Given** the existing AWSService interface pattern
**When** implementing VPC enumeration
**Then** create cmd/awtest/services/vpc/ directory
**And** implement calls.go with DescribeVpcs(), DescribeSubnets(), and DescribeSecurityGroups() API calls using AWS SDK v1.44.266 EC2 client
**And** implement AWSService interface with Name="VPC", Call(), Process(), ModuleName="VPC"
**And** Call() method creates EC2 client, calls DescribeVpcs, DescribeSubnets, DescribeSecurityGroups
**And** Process() method displays VPCs: VpcId, CidrBlock, IsDefault, State, and counts of Subnets and Security Groups per VPC
**And** handle access-denied errors using utils.HandleAWSError
**And** handle empty results with utils.PrintAccessGranted("VPC", "infrastructure")
**And** register service in services/services.go AllServices() after waf (last in alphabetical order)
**And** write table-driven tests in calls_test.go covering: valid VPCs with subnets, default VPC, empty results, access denied
**And** verify package naming: vpc (lowercase)
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/services/vpc/...` passes
**And** verify: `go run ./cmd/awtest -debug` shows "VPC" in output
**And** implicit VPC infrastructure enumeration requirement fulfilled (not explicit FR, but architectural requirement)

---

## Epic 3: Scan Configuration & Control

**Goal:** Security professionals can customize scans for specific scenarios - targeting specific services, excluding noisy services, setting timeouts, and controlling scan scope to match engagement constraints.

### Story 3.1: Service Filtering (Include/Exclude Services)

As a **security professional running targeted scans**,
I want **to include or exclude specific AWS services from enumeration**,
So that **I can run fast triage scans targeting only critical services or comprehensive audits excluding known-noisy services during time-sensitive engagements**.

**Acceptance Criteria:**

**Given** the expanded service coverage from Epic 2 (45 total services)
**When** implementing service filtering
**Then** create cmd/awtest/service_filter.go with filtering logic
**And** add -services flag accepting comma-separated service names (e.g., -services=s3,ec2,iam) (FR33)
**And** add -exclude-services flag accepting comma-separated service names to exclude (e.g., -exclude-services=cloudwatch,cloudtrail) (FR34)
**And** implement ParseServiceFilter() function that takes flag values and returns filtered service list
**And** support case-insensitive service name matching (e.g., "S3", "s3", "S 3" all match S3)
**And** support partial matching (e.g., "cognito" matches both "Cognito Identity" and "Cognito User Pools")
**And** validate service names - error if non-existent service specified
**And** handle conflicting flags appropriately: if both -services and -exclude-services specified, apply -services filter first, then apply exclusions
**And** modify main.go to filter services.AllServices() based on parsed filter before enumeration loop
**And** display "Scanning N services..." message showing filtered service count
**And** when no services match filter, error clearly: "No services matched filter criteria"
**And** write unit tests in service_filter_test.go covering: exact matching, case-insensitive matching, partial matching, include-only, exclude-only, include+exclude combination, invalid service names, no matches
**And** write integration tests verifying: -services filters correctly, -exclude-services filters correctly, combined filters work
**And** verify: `go run ./cmd/awtest -services=s3,ec2 -debug` only scans S3 and EC2
**And** verify: `go run ./cmd/awtest -exclude-services=cloudwatch,cloudtrail` skips CloudWatch and CloudTrail
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/...` passes
**And** FR33 and FR34 requirements fulfilled: Users can select/exclude specific services

### Story 3.2: Timeout Configuration

As a **security professional with time-constrained engagements**,
I want **to set a maximum scan timeout**,
So that **I can ensure scans complete within my engagement window and don't hang indefinitely on slow API calls**.

**Acceptance Criteria:**

**Given** the service enumeration architecture
**When** implementing timeout configuration
**Then** add -timeout flag accepting duration value (e.g., -timeout=5m, -timeout=300s) with default 5 minutes (FR35)
**And** implement timeout context in main.go using context.WithTimeout()
**And** pass timeout context to each service Call() method
**And** modify AWSService.Call() signature to accept context.Context parameter
**And** update all existing service Call() implementations to use context for AWS SDK calls
**And** when timeout is reached, cancel remaining service enumerations gracefully
**And** display timeout warning: "Scan timeout reached after [duration]. N services not scanned."
**And** list services that were not scanned due to timeout
**And** ensure partial results from completed services are still output
**And** exit code 0 if timeout occurs (not an error, just incomplete scan)
**And** handle context cancellation in AWS SDK calls - terminate API calls cleanly
**And** write unit tests for timeout logic covering: timeout before first service, timeout mid-scan, timeout after all services, no timeout
**And** write integration tests verifying: -timeout flag works, partial results preserved, timeout message displayed
**And** verify: `go run ./cmd/awtest -timeout=10s` stops after 10 seconds
**And** verify: `go run ./cmd/awtest -timeout=1h` allows 1 hour scan time
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/...` passes
**And** FR35 requirement fulfilled: Users can set maximum scan timeout duration
**And** update all 45 service implementations to accept context parameter (backward compatible change)

### Story 3.3: Concurrency Configuration (Preparation for Phase 2)

As a **security professional wanting fast scans**,
I want **the ability to configure concurrent service scanning**,
So that **future Phase 2 concurrent enumeration can be controlled and I can prepare for blazing-fast execution**.

**Acceptance Criteria:**

**Given** the current sequential service enumeration
**When** preparing concurrency configuration for Phase 2
**Then** add -concurrency flag accepting integer value (default 1 for sequential) (FR37)
**And** validate -concurrency value: must be >= 1 and <= 20 (reasonable worker pool limit)
**And** when -concurrency=1 (default), maintain existing sequential behavior (no changes)
**And** when -concurrency > 1, display message: "Concurrent enumeration (--concurrency > 1) will be available in Phase 2. Running sequentially."
**And** store concurrency value in config for future Phase 2 implementation
**And** document -concurrency flag in help text with "Phase 2 feature" note
**And** write unit tests for concurrency flag parsing covering: default value, valid values (1-20), invalid values (<1, >20), non-integer values
**And** verify flag parsing does not break current sequential execution
**And** verify: `go run ./cmd/awtest -concurrency=10` runs successfully (sequentially) with Phase 2 message
**And** verify: `go run ./cmd/awtest -concurrency=0` errors: "Concurrency must be >= 1"
**And** verify: `go run ./cmd/awtest -concurrency=50` errors: "Concurrency must be <= 20"
**And** verify: `go build ./cmd/awtest` compiles successfully
**And** verify: `go test ./cmd/awtest/...` passes
**And** FR37 requirement partially fulfilled: Concurrency configuration prepared, full implementation in Phase 2
**And** architecture prepared for Phase 2 goroutine pool implementation

---

## Epic 4: Build Automation & Distribution

**Goal:** Security professionals can easily install awtest via Homebrew on macOS/Linux, ensuring they always have the latest version with cross-platform binaries optimized for their system.

### Story 4.1: GoReleaser Configuration & Cross-Platform Builds

As a **developer releasing awtest versions**,
I want **automated cross-platform binary builds**,
So that **security professionals on macOS, Linux, and Windows can download optimized binaries for their platform without manual compilation**.

**Acceptance Criteria:**

**Given** the need for cross-platform distribution
**When** implementing GoReleaser configuration
**Then** create .goreleaser.yaml in repository root
**And** configure builds section targeting platforms: darwin (amd64, arm64), linux (amd64, arm64), windows (amd64) (FR58)
**And** set binary name to "awtest"
**And** disable CGO (CGO_ENABLED=0) for static binary compilation (FR57)
**And** configure ldflags to embed version and build date: -s -w -X main.Version={{.Version}} -X main.BuildDate={{.Date}}
**And** configure archives section with tar.gz for unix, zip for windows
**And** add main.go version variables: var Version = "dev", var BuildDate = "unknown"
**And** configure release section pointing to GitHub repository MillerMedia/awtest
**And** ensure binary size < 15MB per platform (NFR32)
**And** write .goreleaser.yaml validation: `goreleaser check`
**And** test local snapshot build: `goreleaser build --snapshot --clean`
**And** verify all 5 platform binaries generated: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64
**And** verify each binary is statically linked (no external dependencies) using `ldd` on Linux, `otool -L` on macOS
**And** verify version embedding: `./awtest --version` shows version and build date
**And** verify binaries run on target platforms without errors
**And** document GoReleaser installation in CONTRIBUTING.md
**And** FR57 and FR58 requirements fulfilled: Single static binary, cross-platform support

### Story 4.2: Makefile for Development Workflow

As a **developer working on awtest**,
I want **a Makefile with common development commands**,
So that **I can quickly build, test, install, and clean the project without remembering complex go commands**.

**Acceptance Criteria:**

**Given** the need for streamlined development workflow
**When** implementing Makefile
**Then** create Makefile in repository root
**And** add VERSION variable with default "dev" (overridable: make VERSION=0.4.0)
**And** add BUILD_DATE variable using shell date command: $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
**And** add LDFLAGS variable embedding version and build date
**And** implement "build" target: `go build $(LDFLAGS) -o awtest ./cmd/awtest`
**And** implement "test" target: `go test -v -race -coverprofile=coverage.out ./...`
**And** implement "test-coverage" target: `go tool cover -html=coverage.out` to view coverage report
**And** implement "lint" target: `golangci-lint run` (document golangci-lint installation requirement)
**And** implement "install" target: `go install $(LDFLAGS) ./cmd/awtest`
**And** implement "clean" target: `rm -f awtest coverage.out`
**And** implement "snapshot" target: `goreleaser build --snapshot --clean` for local multi-platform testing
**And** add .PHONY declarations for all targets
**And** add help target showing available commands and descriptions
**And** verify: `make build` produces awtest binary with embedded version
**And** verify: `make test` runs all tests with race detector
**And** verify: `make install` installs to $GOPATH/bin
**And** verify: `make clean` removes build artifacts
**And** verify: `make` (no target) shows help menu
**And** document Makefile usage in CONTRIBUTING.md
**And** Makefile enables efficient local development workflow

### Story 4.3: GitHub Actions Release Automation

As a **developer releasing awtest versions**,
I want **automated GitHub releases on git tag push**,
So that **cross-platform binaries, checksums, and release notes are automatically generated and published without manual intervention**.

**Acceptance Criteria:**

**Given** the GoReleaser configuration from Story 4.1
**When** implementing GitHub Actions automation
**Then** create .github/workflows/release.yml workflow file
**And** trigger workflow on tag push matching pattern: v* (e.g., v0.4.0, v1.0.0)
**And** checkout repository with fetch-depth: 0 for full git history (GoReleaser needs tags)
**And** setup Go environment with go-version matching go.mod (Go 1.19)
**And** run tests before release: `go test ./...` to ensure quality
**And** install GoReleaser using goreleaser-action
**And** run GoReleaser with GITHUB_TOKEN for release creation
**And** configure GoReleaser to generate SHA256 checksums for all binaries
**And** configure GoReleaser to auto-generate release notes from git commits
**And** test workflow locally using `act` (GitHub Actions local runner) if possible
**And** create test release workflow (.github/workflows/test.yml) running on pull requests: build + test only, no release
**And** test workflow triggers: Create test tag, push, verify workflow runs
**And** verify release artifacts include: binaries for all 5 platforms, checksums file, release notes
**And** verify release appears on GitHub Releases page
**And** document release process in CONTRIBUTING.md: "To release: git tag v0.x.y && git push origin v0.x.y"
**And** GitHub Actions enables automated, reliable releases

### Story 4.4: Homebrew Tap Setup & Distribution

As a **security professional on macOS or Linux**,
I want **to install awtest via Homebrew**,
So that **I can use `brew install awtest` for frictionless installation and automatic updates**.

**Acceptance Criteria:**

**Given** the GoReleaser and GitHub Actions from Stories 4.1-4.3
**When** implementing Homebrew distribution
**Then** create separate GitHub repository: MillerMedia/homebrew-tap for Homebrew formulas
**And** configure GoReleaser brews section in .goreleaser.yaml
**And** set brew name to "awtest"
**And** set brew repository to owner: MillerMedia, name: homebrew-tap
**And** configure brew description: "AWS credential enumeration for security assessments"
**And** configure brew homepage: "https://github.com/MillerMedia/awtest"
**And** configure brew install stanza: `bin.install "awtest"`
**And** on release, GoReleaser automatically updates homebrew-tap repository with new formula version
**And** test Homebrew installation locally: `brew install MillerMedia/tap/awtest`
**And** verify: `brew install MillerMedia/tap/awtest` downloads and installs latest release
**And** verify: `awtest --version` shows correct version after brew install
**And** verify: `brew upgrade awtest` upgrades to newer version
**And** verify: `brew uninstall awtest` removes binary cleanly
**And** go install remains supported: `go install github.com/MillerMedia/awtest/cmd/awtest@latest` (FR56)
**And** test go install: verify it downloads, builds, and installs to $GOPATH/bin
**And** update README.md with installation instructions for both Homebrew and go install
**And** installation section includes: brew install (macOS/Linux), go install (all platforms), direct binary download from releases
**And** FR55 and FR56 requirements fulfilled: Homebrew installation, go install support
**And** frictionless installation enables community adoption and growth

---

## Epic 5: Documentation & Community Contribution Framework

**Goal:** Open-source contributors can add new AWS services to awtest following clear templates and contribution guidelines, ensuring the tool stays current as AWS releases new services.

### Story 5.1: Service Implementation Template & Documentation

As a **community contributor wanting to add an AWS service**,
I want **a complete service implementation template with examples**,
So that **I can quickly add new services following established patterns without reading the entire codebase**.

**Acceptance Criteria:**

**Given** the established AWSService interface pattern from 34 existing services
**When** creating service implementation template
**Then** create _template/ directory in repository root
**And** create _template/service_calls_template.go with complete boilerplate code
**And** template includes package declaration: `package SERVICENAME`
**And** template includes AWS SDK import: `"github.com/aws/aws-sdk-go/service/AWSSERVICE"`
**And** template includes types and utils imports
**And** template defines AWSService implementation with Name, Call(), Process(), ModuleName
**And** template includes Call() function with AWS client creation, API call, error handling
**And** template includes Process() function with result iteration, utils.PrintResult(), utils.HandleAWSError()
**And** template includes service registration export: `var SERVICECalls = []types.AWSService{...}`
**And** template includes comprehensive code comments explaining each section
**And** template includes TODO markers for customization: `// TODO: Replace SERVICENAME with actual service name`
**And** template follows all naming conventions: PascalCase exports, camelCase unexported, lowercase package
**And** create _template/README.md explaining template usage step-by-step
**And** include real-world example in _template/: copy existing S3 service as reference implementation
**And** verify template compiles after placeholder replacement: test by creating a sample service
**And** document template structure in CONTRIBUTING.md
**And** FR65 requirement fulfilled: Clear service implementation template provided

### Story 5.2: CONTRIBUTING.md Guide for Service Addition

As a **community contributor wanting to contribute**,
I want **comprehensive contribution guidelines**,
So that **I can add AWS services, submit quality pull requests, and understand project standards without maintainer guidance**.

**Acceptance Criteria:**

**Given** the complete architecture from Epics 1-4
**When** creating contribution documentation
**Then** create CONTRIBUTING.md in repository root
**And** include "Adding a New AWS Service" section with step-by-step guide:
  1. Copy _template/service_calls_template.go to cmd/awtest/services/SERVICENAME/calls.go
  2. Replace SERVICENAME placeholders with actual service name
  3. Implement AWS SDK API calls in Call() function
  4. Implement result processing in Process() function
  5. Register service in cmd/awtest/services/services.go AllServices() in alphabetical order
  6. Write table-driven tests in calls_test.go
  7. Run `make test` to verify tests pass
  8. Run `go build ./cmd/awtest` to verify compilation
  9. Test manually with `go run ./cmd/awtest -debug`
  10. Submit pull request with service name in title
**And** include "Code Standards" section documenting:
  - Naming conventions (PascalCase, camelCase, lowercase files)
  - Error handling patterns (utils.HandleAWSError)
  - Test requirements (>70% coverage, table-driven tests)
  - Documentation requirements (inline comments for exported functions)
**And** include "Development Workflow" section:
  - Prerequisites: Go 1.19+, make, golangci-lint, GoReleaser
  - Setup: `go mod download`
  - Build: `make build`
  - Test: `make test`
  - Lint: `make lint`
**And** include "Pull Request Process" section:
  - PR title format: "Add [Service Name] enumeration"
  - PR description template with checklist
  - Review process expectations (48-hour review SLA per NFR29)
  - Testing requirements before merge
**And** include "Service Validation Checklist":
  - [ ] Service follows AWSService interface pattern
  - [ ] Package name is lowercase, no underscores
  - [ ] Error handling uses utils.HandleAWSError
  - [ ] Tests achieve >70% coverage
  - [ ] Service registered in AllServices() alphabetically
  - [ ] Compiles successfully
  - [ ] Manual testing completed
**And** include "Release Process" section (for maintainers):
  - Tag creation: `git tag v0.x.y`
  - Push tag: `git push origin v0.x.y`
  - GitHub Actions handles release automation
**And** verify CONTRIBUTING.md accuracy by following it to add a test service
**And** FR64 and FR66 requirements fulfilled: Documented patterns, validation checklist for consistency

### Story 5.3: README.md Updates with New Features

As a **security professional discovering awtest**,
I want **comprehensive README documentation**,
So that **I understand awtest's capabilities, installation options, usage examples, and how to get started immediately**.

**Acceptance Criteria:**

**Given** all features implemented in Epics 1-4
**When** updating README documentation
**Then** update README.md with current project state
**And** update "Installation" section with three methods:
  1. Homebrew: `brew install MillerMedia/tap/awtest` (macOS/Linux)
  2. Go Install: `go install github.com/MillerMedia/awtest/cmd/awtest@latest`
  3. Direct Download: Link to GitHub Releases with platform-specific binaries
**And** update "Features" section highlighting:
  - 45 AWS services enumerated (34 existing + 11 new from Epic 2)
  - Multiple output formats (text, JSON, YAML, CSV, table)
  - Service targeting and exclusion
  - Configurable timeouts
  - Cross-platform support
**And** update "Usage" section with examples:
  - Basic scan: `awtest`
  - With credentials: `awtest -access-key-id=AKI... -secret-access-key=...`
  - JSON output: `awtest -format=json -output-file=results.json`
  - Target specific services: `awtest -services=s3,ec2,iam`
  - Exclude services: `awtest -exclude-services=cloudwatch,cloudtrail`
  - Set timeout: `awtest -timeout=10m`
**And** add "Output Formats" section with format descriptions and use cases:
  - text: Human-readable terminal output (default)
  - json: Programmatic parsing, SIEM integration
  - yaml: Readable structured reports
  - csv: Spreadsheet analysis
  - table: Structured terminal view
**And** add "AWS Services Covered" section listing all 45 services by category:
  - Compute & Containers (EC2, Lambda, ECS, EKS, Fargate, Batch)
  - Databases (RDS, DynamoDB, ElastiCache, Redshift)
  - Security & Identity (IAM, Secrets Manager, KMS, Certificate Manager, Cognito)
  - Storage (S3, EBS, EFS, Glacier)
  - Networking (VPC, API Gateway, CloudFront, Route53)
  - Management (CloudFormation, CloudWatch, CloudTrail, Config, Systems Manager)
  - Application Services (SNS, SQS, Step Functions, EventBridge)
**And** add "Contributing" section linking to CONTRIBUTING.md
**And** add "License" section (if not already present)
**And** update badges: build status, go version, latest release
**And** include real-world usage examples from user journey scenarios (Alex, Riley, Jordan from PRD)
**And** verify all README links work (installation links, GitHub links, documentation links)
**And** verify all code examples execute correctly
**And** README provides complete getting-started experience for new users
