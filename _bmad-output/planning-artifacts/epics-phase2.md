---
stepsCompleted: ['step-01-validate-prerequisites', 'step-02-design-epics', 'step-03-create-stories', 'step-04-final-validation']
inputDocuments:
  - prd-phase2.md
  - architecture-phase2.md
  - prd.md
  - architecture.md
---

# awtest Phase 2 - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for awtest Phase 2, decomposing the requirements from the PRD Phase 2 and Architecture Phase 2 into implementable stories. Phase 1 context (prd.md, architecture.md) used for reference.

## Requirements Inventory

### Functional Requirements

FR67: System executes service enumeration concurrently using a configurable concurrent execution engine
FR68: Users can select a named speed preset via `--speed` flag (safe, fast, insane)
FR69: Users can override speed preset concurrency with `--concurrency=N` numeric flag (1-20)
FR70: System maps speed presets to concurrency levels (safe=1, fast=5, insane=20)
FR71: System produces identical results regardless of speed preset or concurrency level
FR72: System collects results from concurrent services safely, preventing data corruption under parallel access
FR73: System sorts final output by service name for deterministic ordering across all speed presets
FR74: System retries AWS API calls that receive throttling responses (429/503) with exponential backoff
FR75: System applies backoff independently per service — throttling on one service does not block others
FR76: System varies retry timing across concurrent services to distribute retry load and prevent simultaneous retry storms
FR77: System distinguishes between throttling (retry), access denied (skip), and service errors (report)
FR78: System delivers complete results even when throttling occurs at any concurrency level
FR79: System displays in-place updating progress during concurrent scans showing services completed vs. total
FR80: System replaces progress display with summary statistics upon scan completion
FR81: System writes progress to stderr to avoid interfering with stdout formatted output
FR82: System suppresses progress display when `--quiet` flag is set
FR83: System suppresses progress display when stdout is not a TTY (piped output)
FR84: System preserves existing per-service sequential progress reporting in `--speed=safe` mode
FR85: System enumerates ECR container repositories, images, and repository policies
FR86: System enumerates AWS Organizations accounts, organizational units, and service control policies
FR87: System enumerates GuardDuty detectors, findings, and suppression filters
FR88: System enumerates Security Hub findings, enabled products, and compliance status
FR89: System enumerates CodeBuild projects and build configurations
FR90: System enumerates CodeCommit repositories and branches
FR91: System enumerates OpenSearch domains, access policies, and encryption configurations
FR92: System enumerates SageMaker notebook instances, endpoints, models, and training jobs
FR93: System enumerates AWS Backup vaults, backup plans, recovery points, and vault access policies
FR94: System enumerates Athena workgroups, saved queries, and query execution history
FR95: System enumerates Macie findings, classification jobs, and sensitive data discovery results
FR96: System maintains read-only operation guarantee under all concurrency levels
FR97: System prevents credential data from leaking through concurrent error paths or concurrent process crashes
FR98: System completes or terminates all in-progress service scans within 1 second of timeout, preserving partial results collected before cancellation
FR99: System drains in-progress service scans before formatting output on timeout or cancellation
FR100: System validates `--speed` flag accepts only valid presets (safe, fast, insane)
FR101: System resolves `--concurrency` and `--speed` conflict deterministically — `--concurrency` overrides preset
FR102: System displays speed preset and effective concurrency level in scan output header
FR103: All existing Phase 1 flags remain unchanged and compatible with new Phase 2 flags
FR104: README reflects 63 total services with updated service categories
FR105: README documents speed presets with OPSEC tradeoff guidance
FR106: CONTRIBUTING.md documents concurrent testing requirements for new service additions
FR107: Service template updated for concurrent-safe implementation patterns
FR108: System enumerates CodeDeploy applications, deployment groups, and deployment configurations
FR109: System enumerates Direct Connect connections, virtual interfaces, and gateways
FR110: System enumerates EMR clusters, instance groups, and security configurations
FR111: System enumerates Kinesis streams, shard details, and consumer applications
FR112: System enumerates MediaConvert jobs, queues, and presets
FR113: System enumerates Neptune DB clusters, instances, and parameter groups

### NonFunctional Requirements

NFR35: `--speed=insane` completes full scan across all services in under 30 seconds for credentials with access to 20+ services
NFR36: `--speed=fast` completes full scan in under 60 seconds for credentials with access to 20+ services
NFR37: `--speed=safe` maintains equivalent performance to Phase 1 sequential scanning (no regression)
NFR38: Concurrent scan time achieves at least 80% of linear scaling efficiency when doubling worker count, up to the point of rate limiting
NFR39: Memory consumption remains under 100MB at `--speed=insane` with all services running concurrently
NFR40: In-place progress updates render at minimum 2 updates per second during concurrent scans without flickering or tearing
NFR41: Output formatting (JSON, YAML, CSV, table, text) adds less than 100ms overhead after all services complete
NFR42: Read-only operation guarantee (NFR7) holds unconditionally under all concurrency levels — verified by concurrent test suite
NFR43: Credential values never appear in concurrent crash stack traces, error logs, or concurrent error paths
NFR44: AWS session sharing across goroutines uses verified thread-safe patterns — no credential corruption under concurrent access
NFR45: No credential data is transmitted outside AWS API endpoints regardless of concurrency level (NFR12 extended)
NFR46: Individual service failures in concurrent execution do not affect other concurrent services — fault isolation per goroutine
NFR47: Concurrent scan results are identical to sequential scan results for the same credentials — verified by automated comparison tests
NFR48: Concurrent process crashes in individual services are recovered without crashing the process or losing results from other services
NFR49: Context cancellation (timeout) cleanly terminates all concurrent operations within 1 second — no resource leaks
NFR50: Exponential backoff retries converge — throttled services eventually succeed or fail definitively, never retry infinitely
NFR51: Rate limit backoff adds maximum 15 seconds total delay per service (3 retries: 100ms, 500ms, 2.5s + jitter, then fail)
NFR52: AWS session concurrent usage verified safe through explicit concurrent integration tests
NFR53: All new services follow existing AWSService interface pattern — no changes to core enumeration engine required (NFR26 maintained)
NFR54: New services register in AllServices() maintaining alphabetical ordering convention
NFR55: Concurrent execution compatible with all existing output formatters without formatter modifications
NFR56: Concurrency implementation encapsulated in worker pool module — service implementations remain concurrency-unaware
NFR57: New service additions require no concurrency-specific code — the worker pool handles parallelism transparently
NFR58: Code coverage exceeds 70% for concurrent execution paths including race condition detection tests
NFR59: Concurrent vs. sequential comparison test suite runs as part of standard `make test`

### Additional Requirements

**From Architecture Phase 2:**

- safeScan wrapper function for panic recovery and error classification — foundation for all concurrent execution
- Worker pool with buffered channel dispatch and fixed goroutine pool sized by speed preset
- Mutex-protected shared slice for result collection with post-scan sort by service name
- Per-service inline exponential backoff with jitter (base 1s, 2x multiplier, +/-50% jitter, max 15s cap, max 3 retries)
- Atomic counter + ticker goroutine for progress reporting at 2+ Hz
- TTY detection dependency: golang.org/x/term required for IsTerminal() on Go 1.19
- New concurrency files in cmd/awtest/: worker_pool.go, safe_scan.go, backoff.go, progress.go, speed.go
- 17 new service files in services/ directory following existing template pattern
- go test -race required after any concurrency-related changes
- Services must remain concurrency-unaware — never import sync or sync/atomic
- All service executions must go through safeScan wrapper — no direct calls from workers
- Error classification: throttling (retry with backoff), access denied (skip), service error (report)
- Graceful shutdown: on context cancellation, drain in-progress services with 1s deadline before formatting output
- make test-race convenience target recommended for Makefile
- Starter template not applicable — brownfield project extending Phase 1 foundation
- No new external dependencies for core concurrency (Go stdlib only, except golang.org/x/term)

**Implementation Sequence (from Architecture):**
1. safeScan wrapper (foundation)
2. Worker pool with channel dispatch (core engine)
3. Speed preset resolution and flag validation
4. Progress reporting
5. Rate limit backoff
6. New service additions (can parallelize with above)
7. Documentation updates

**Service count adjustment:** PRD referenced 57 services (46 + 11). Actual Phase 2 target is 63 services (46 Phase 1 + 17 new). Documentation references to "57 services" should be updated to "63 services."

### FR Coverage Map

| FR | Epic | Description |
|---|---|---|
| FR67 | Epic 1 | Concurrent execution engine |
| FR68 | Epic 1 | --speed flag (safe/fast/insane) |
| FR69 | Epic 1 | --concurrency=N override |
| FR70 | Epic 1 | Speed preset to concurrency mapping |
| FR71 | Epic 1 | Identical results regardless of speed |
| FR72 | Epic 1 | Thread-safe result collection |
| FR73 | Epic 1 | Deterministic output ordering |
| FR74 | Epic 1 | Exponential backoff on throttling |
| FR75 | Epic 1 | Per-service independent backoff |
| FR76 | Epic 1 | Jitter to prevent retry storms |
| FR77 | Epic 1 | Error classification (throttle/denied/error) |
| FR78 | Epic 1 | Complete results under throttling |
| FR79 | Epic 1 | In-place progress display |
| FR80 | Epic 1 | Progress replaced by summary on completion |
| FR81 | Epic 1 | Progress writes to stderr |
| FR82 | Epic 1 | Suppress progress in quiet mode |
| FR83 | Epic 1 | Suppress progress in non-TTY |
| FR84 | Epic 1 | Sequential progress in safe mode |
| FR85 | Epic 2 | ECR enumeration |
| FR86 | Epic 2 | Organizations enumeration |
| FR87 | Epic 2 | GuardDuty enumeration |
| FR88 | Epic 2 | Security Hub enumeration |
| FR89 | Epic 3 | CodeBuild enumeration |
| FR90 | Epic 3 | CodeCommit enumeration |
| FR91 | Epic 3 | OpenSearch enumeration |
| FR92 | Epic 3 | SageMaker enumeration |
| FR93 | Epic 3 | Backup enumeration |
| FR94 | Epic 3 | Athena enumeration |
| FR95 | Epic 3 | Macie enumeration |
| FR96 | Epic 1 | Read-only guarantee under concurrency |
| FR97 | Epic 1 | No credential leakage in concurrent paths |
| FR98 | Epic 1 | Timeout termination within 1s |
| FR99 | Epic 1 | Drain in-progress scans before output |
| FR100 | Epic 1 | Speed flag validation |
| FR101 | Epic 1 | Concurrency overrides speed preset |
| FR102 | Epic 1 | Display speed/concurrency in header |
| FR103 | Epic 1 | Phase 1 flag compatibility |
| FR104 | Epic 5 | README reflects 63 services |
| FR105 | Epic 5 | Speed preset OPSEC guidance |
| FR106 | Epic 5 | CONTRIBUTING.md concurrent testing docs |
| FR107 | Epic 5 | Service template concurrent-safe update |
| FR108 | Epic 4 | CodeDeploy enumeration |
| FR109 | Epic 4 | DirectConnect enumeration |
| FR110 | Epic 4 | EMR enumeration |
| FR111 | Epic 4 | Kinesis enumeration |
| FR112 | Epic 4 | MediaConvert enumeration |
| FR113 | Epic 4 | Neptune enumeration |

## Epic List

### Epic 1: Concurrent Speed Scanning
Users can scan all services in seconds using `--speed` presets (safe/fast/insane), with reliable rate limit handling, real-time progress visibility, and concurrent safety guarantees. This is the headline Phase 2 feature — transforming awtest from comprehensive to *fast*.
**FRs covered:** FR67-84, FR96-103
**NFRs addressed:** NFR35-52, NFR55-59

### Epic 2: Critical Security Service Expansion
Users discover container registries with embedded secrets (ECR), organizational account mapping (Organizations), detection coverage gaps (GuardDuty), and aggregated compliance posture (Security Hub). These are the highest-value security services for pentesters and red teamers.
**FRs covered:** FR85-88
**NFRs addressed:** NFR53-54

### Epic 3: High-Priority Service Expansion
Users enumerate build secrets (CodeBuild), source code access (CodeCommit), search clusters (OpenSearch), ML infrastructure (SageMaker), backup vaults and recovery points (Backup), analytics queries (Athena), and sensitive data maps (Macie).
**FRs covered:** FR89-95
**NFRs addressed:** NFR53-54

### Epic 4: Infrastructure & Data Service Expansion
Users discover deployment pipelines (CodeDeploy), network connections (DirectConnect), big data clusters (EMR), streaming data (Kinesis), media processing (MediaConvert), and graph databases (Neptune).
**FRs covered:** FR108-113
**NFRs addressed:** NFR53-54

### Epic 5: Documentation & Contributor Enablement
README reflects 63 services with speed preset and OPSEC tradeoff guidance. CONTRIBUTING.md documents concurrent testing requirements. Service template updated for concurrent-safe patterns. Contributors can confidently add new services to the concurrent architecture.
**FRs covered:** FR104-107

## Epic 1: Concurrent Speed Scanning

Users can scan all services in seconds using `--speed` presets (safe/fast/insane), with reliable rate limit handling, real-time progress visibility, and concurrent safety guarantees. This is the headline Phase 2 feature — transforming awtest from comprehensive to *fast*.

### Story 1.1: Speed Preset & Concurrency Flag Resolution

As a pentester,
I want to select `--speed=safe/fast/insane` or `--concurrency=N` to control scan parallelism,
So that I can choose the right speed-vs-OPSEC tradeoff for my engagement.

**Acceptance Criteria:**

**Given** a user runs awtest with `--speed=fast`
**When** the scan begins
**Then** the effective concurrency level is set to 5 workers
**And** the scan output header displays "Speed: fast (concurrency: 5)"

**Given** a user runs awtest with `--speed=insane`
**When** the scan begins
**Then** the effective concurrency level is set to 20 workers

**Given** a user runs awtest with `--speed=safe`
**When** the scan begins
**Then** the effective concurrency level is 1 (sequential, identical to Phase 1 behavior)

**Given** a user runs awtest with `--concurrency=10 --speed=fast`
**When** flags are resolved
**Then** `--concurrency` overrides the speed preset (effective concurrency = 10)

**Given** a user runs awtest with `--speed=invalid`
**When** flags are validated
**Then** the tool exits with an error listing valid presets (safe, fast, insane)

**Given** a user runs awtest without `--speed` or `--concurrency`
**When** flags are resolved
**Then** the default is `--speed=safe` (concurrency=1), preserving Phase 1 behavior

**Given** all Phase 1 flags are used alongside `--speed`
**When** flags are parsed
**Then** all existing flags work unchanged (backward compatibility)

*FRs: FR68-70, FR100-103 | New file: speed.go, speed_test.go*

### Story 1.2: safeScan Wrapper with Panic Recovery & Error Classification

As a pentester,
I want each service scan wrapped with panic recovery and error classification,
So that one service crash doesn't lose results from other services, and errors are handled appropriately.

**Acceptance Criteria:**

**Given** a service scan panics during execution
**When** the panic is caught by safeScan
**Then** the panic is converted to an error result for that service
**And** no credential data appears in the error message or stack trace
**And** other services are unaffected

**Given** a service returns an AWS throttling response (HTTP 429 / RequestLimitExceeded)
**When** safeScan classifies the error
**Then** it is classified as "throttle" (eligible for retry)

**Given** a service returns an access denied response
**When** safeScan classifies the error
**Then** it is classified as "denied" (service skipped, no error reported)

**Given** a service returns any other error
**When** safeScan classifies the error
**Then** it is classified as "error" (included in results as error entry)

**Given** safeScan wraps a service execution
**When** the service completes successfully
**Then** the result is returned normally with no overhead

**Given** the system operates under concurrent execution
**When** credential data exists in goroutine state
**Then** no credentials appear in panic stack traces or error logs (NFR43)

*FRs: FR77, FR96-97 | New file: safe_scan.go, safe_scan_test.go*

### Story 1.3: Concurrent Worker Pool Execution

As a pentester,
I want services to scan concurrently using a worker pool,
So that scans complete in seconds instead of minutes at higher speed presets.

**Acceptance Criteria:**

**Given** a user runs awtest with `--speed=fast` (concurrency=5)
**When** the scan executes
**Then** 5 worker goroutines process services concurrently via buffered channel dispatch
**And** all service executions go through the safeScan wrapper

**Given** concurrent workers complete services in arbitrary order
**When** all workers finish
**Then** results are sorted by service name for deterministic output
**And** output is identical to `--speed=safe` for the same credentials (FR71)

**Given** concurrent workers collect results
**When** multiple workers append results simultaneously
**Then** results are protected by mutex — no data corruption (FR72)

**Given** the scan timeout is reached during concurrent execution
**When** context cancellation propagates to workers
**Then** in-progress services drain within 1 second (FR98)
**And** partial results collected before cancellation are preserved (FR99)
**And** output is formatted with available results

**Given** a user runs awtest with `--speed=safe` (concurrency=1)
**When** the scan executes
**Then** behavior is identical to Phase 1 sequential scanning (no regression)

**Given** concurrent execution at `--speed=insane` (20 workers)
**When** all services complete
**Then** memory consumption remains under 100MB (NFR39)

**Given** the scan runs with any concurrency level
**When** services execute
**Then** all operations remain strictly read-only (FR96, NFR42)

*FRs: FR67, FR71-73, FR98-99 | New file: worker_pool.go, worker_pool_test.go | Tests must include go test -race*

### Story 1.4: Rate Limit Resilience with Exponential Backoff

As a pentester,
I want automatic retry with exponential backoff when AWS throttles API calls,
So that I get complete results even at high concurrency without manual intervention.

**Acceptance Criteria:**

**Given** a service receives an AWS throttling response (429/RequestLimitExceeded)
**When** the backoff retry logic executes
**Then** the service retries with exponential backoff (base 1s, 2x multiplier, +/-50% jitter)
**And** maximum 3 retries before marking service as rate-limited error
**And** maximum total delay per service is 15 seconds (NFR51)

**Given** one service is being throttled
**When** it enters backoff retry
**Then** other concurrent services continue executing unblocked (FR75)
**And** backoff state is per-service, not global

**Given** multiple services retry simultaneously
**When** jitter is applied to retry timing
**Then** retry attempts are spread across time to prevent retry storms (FR76)

**Given** a throttled service eventually succeeds after retry
**When** results are collected
**Then** the result appears normal — no indication of throttling to the user

**Given** a throttled service exhausts all 3 retries
**When** the final retry fails
**Then** the service is included in results as a rate-limited error
**And** the error message indicates rate limiting

**Given** backoff is in progress and context cancellation occurs
**When** the retry loop checks context between retries
**Then** the retry is abandoned and partial results are preserved

**Given** any concurrency level from 1 to 20
**When** throttling occurs on some services
**Then** the system delivers complete results for non-throttled services (FR78)

*FRs: FR74-78 | New file: backoff.go, backoff_test.go*

### Story 1.5: Concurrent Progress Reporting

As a pentester,
I want real-time progress during concurrent scans,
So that I know the scan is active and how many services have completed.

**Acceptance Criteria:**

**Given** a user runs awtest with `--speed=fast` or `--speed=insane` in a TTY terminal
**When** the scan is in progress
**Then** an in-place updating progress display shows "Scanning... 15/63 services complete" on stderr
**And** the display updates at minimum 2 Hz without flickering (NFR40)

**Given** a concurrent scan completes
**When** all services finish
**Then** the progress display is replaced by summary statistics (FR80)

**Given** a user runs awtest with `--quiet` flag
**When** the scan executes
**Then** no progress display is shown (FR82)

**Given** awtest output is piped to a file or another command (non-TTY)
**When** the scan executes
**Then** no progress display is shown (FR83)

**Given** a user runs awtest with `--speed=safe` (sequential mode)
**When** the scan executes
**Then** the existing per-service "Scanning [service_name]..." output is preserved (FR84)
**And** no in-place progress counter is used

**Given** progress is displayed during concurrent scans
**When** output is written
**Then** progress writes to stderr only — stdout contains only formatted scan results (FR81)

**Given** a concurrent scan with 63 services at `--speed=insane`
**When** services complete at different times
**Then** the progress counter increments atomically as each service finishes, regardless of completion order

*FRs: FR79-84 | New file: progress.go, progress_test.go | Dependency: golang.org/x/term for TTY detection*

## Epic 2: Critical Security Service Expansion

Users discover container registries with embedded secrets (ECR), organizational account mapping (Organizations), detection coverage gaps (GuardDuty), and aggregated compliance posture (Security Hub). These are the highest-value security services for pentesters and red teamers.

### Story 2.1: ECR Container Registry Enumeration

As a pentester,
I want to enumerate ECR container repositories, images, and repository policies,
So that I can discover container images with embedded secrets and overly permissive registry access.

**Acceptance Criteria:**

**Given** credentials with ECR read access
**When** the scan executes
**Then** all ECR repositories are listed with repository names and URIs
**And** images within each repository are enumerated
**And** repository policies are retrieved showing access permissions

**Given** credentials without ECR access
**When** the scan executes
**Then** ECR is skipped silently (access denied handling)

**Given** the ECR service is registered
**When** GetServices() is called
**Then** ECR appears in alphabetical order among all services

**Given** the scan runs at any concurrency level
**When** ECR executes
**Then** the service contains no sync primitives and is concurrency-unaware (NFR57)

*FR: FR85 | New file: services/ecr/calls.go + tests*

### Story 2.2: AWS Organizations Enumeration

As a pentester,
I want to enumerate AWS Organizations accounts, OUs, and service control policies,
So that I can map the organizational account structure and identify cross-account access opportunities.

**Acceptance Criteria:**

**Given** credentials with Organizations read access
**When** the scan executes
**Then** all member accounts are listed with account IDs and names
**And** organizational units (OUs) are enumerated with hierarchy
**And** service control policies (SCPs) are retrieved

**Given** credentials without Organizations access (non-management account)
**When** the scan executes
**Then** Organizations is skipped silently

**Given** the Organizations service is registered
**When** GetServices() is called
**Then** Organizations appears in alphabetical order

*FR: FR86 | New file: services/organizations/calls.go + tests*

### Story 2.3: GuardDuty Enumeration

As a pentester,
I want to enumerate GuardDuty detectors, findings, and suppression filters,
So that I can identify detection coverage gaps and understand what security monitoring is active.

**Acceptance Criteria:**

**Given** credentials with GuardDuty read access
**When** the scan executes
**Then** all GuardDuty detectors are listed with status and configuration
**And** recent findings are enumerated with severity and type
**And** suppression filters are retrieved showing what's being suppressed

**Given** credentials without GuardDuty access
**When** the scan executes
**Then** GuardDuty is skipped silently

**Given** the GuardDuty service is registered
**When** GetServices() is called
**Then** GuardDuty appears in alphabetical order

*FR: FR87 | New file: services/guardduty/calls.go + tests*

### Story 2.4: Security Hub Enumeration

As a pentester,
I want to enumerate Security Hub findings, enabled products, and compliance status,
So that I can understand the aggregated security posture and identify compliance gaps.

**Acceptance Criteria:**

**Given** credentials with Security Hub read access
**When** the scan executes
**Then** enabled security products/integrations are listed
**And** recent findings are enumerated with severity and compliance status
**And** compliance standards status is retrieved

**Given** credentials without Security Hub access
**When** the scan executes
**Then** Security Hub is skipped silently

**Given** the Security Hub service is registered
**When** GetServices() is called
**Then** Security Hub appears in alphabetical order

*FR: FR88 | New file: services/securityhub/calls.go + tests*

## Epic 3: High-Priority Service Expansion

Users enumerate build secrets (CodeBuild), source code access (CodeCommit), search clusters (OpenSearch), ML infrastructure (SageMaker), backup vaults and recovery points (Backup), analytics queries (Athena), and sensitive data maps (Macie).

### Story 3.1: CodeBuild Enumeration

As a pentester,
I want to enumerate CodeBuild projects and build configurations,
So that I can discover build environment variables containing secrets and access build history.

**Acceptance Criteria:**

**Given** credentials with CodeBuild read access
**When** the scan executes
**Then** all build projects are listed with project names
**And** build environment variables are enumerated (key names, not secret values)
**And** recent build history is retrieved

**Given** credentials without CodeBuild access
**When** the scan executes
**Then** CodeBuild is skipped silently

*FR: FR89 | New file: services/codebuild/calls.go + tests*

### Story 3.2: CodeCommit Enumeration

As a pentester,
I want to enumerate CodeCommit repositories and branches,
So that I can discover source code repositories accessible with these credentials.

**Acceptance Criteria:**

**Given** credentials with CodeCommit read access
**When** the scan executes
**Then** all repositories are listed with names and clone URLs
**And** branches within each repository are enumerated

**Given** credentials without CodeCommit access
**When** the scan executes
**Then** CodeCommit is skipped silently

*FR: FR90 | New file: services/codecommit/calls.go + tests*

### Story 3.3: OpenSearch Enumeration

As a pentester,
I want to enumerate OpenSearch domains, access policies, and encryption configurations,
So that I can discover search clusters with sensitive data and identify access control weaknesses.

**Acceptance Criteria:**

**Given** credentials with OpenSearch read access
**When** the scan executes
**Then** all OpenSearch domains are listed with domain names and endpoints
**And** access policies are retrieved for each domain
**And** encryption configuration is reported (at-rest and in-transit)

**Given** credentials without OpenSearch access
**When** the scan executes
**Then** OpenSearch is skipped silently

*FR: FR91 | New file: services/opensearch/calls.go + tests*

### Story 3.4: SageMaker Enumeration

As a pentester,
I want to enumerate SageMaker notebook instances, endpoints, models, and training jobs,
So that I can discover ML infrastructure and potential data access through notebooks.

**Acceptance Criteria:**

**Given** credentials with SageMaker read access
**When** the scan executes
**Then** notebook instances are listed with status and instance types
**And** endpoints are enumerated with configuration
**And** models and training jobs are listed

**Given** credentials without SageMaker access
**When** the scan executes
**Then** SageMaker is skipped silently

*FR: FR92 | New file: services/sagemaker/calls.go + tests*

### Story 3.5: AWS Backup Enumeration

As a pentester,
I want to enumerate Backup vaults, plans, recovery points, and vault access policies,
So that I can discover backup data stores and identify cross-account backup access.

**Acceptance Criteria:**

**Given** credentials with Backup read access
**When** the scan executes
**Then** backup vaults are listed with names and recovery point counts
**And** backup plans are enumerated
**And** recovery points are listed per vault
**And** vault access policies are retrieved

**Given** credentials without Backup access
**When** the scan executes
**Then** Backup is skipped silently

*FR: FR93 | New file: services/backup/calls.go + tests*

### Story 3.6: Athena Enumeration

As a pentester,
I want to enumerate Athena workgroups, saved queries, and query execution history,
So that I can discover analytics queries that may reveal data access patterns and S3 data locations.

**Acceptance Criteria:**

**Given** credentials with Athena read access
**When** the scan executes
**Then** workgroups are listed with configuration
**And** saved (named) queries are enumerated
**And** recent query execution history is retrieved

**Given** credentials without Athena access
**When** the scan executes
**Then** Athena is skipped silently

*FR: FR94 | New file: services/athena/calls.go + tests*

### Story 3.7: Macie Enumeration

As a pentester,
I want to enumerate Macie findings, classification jobs, and sensitive data discovery results,
So that I can identify where sensitive data has been detected across S3 buckets.

**Acceptance Criteria:**

**Given** credentials with Macie read access
**When** the scan executes
**Then** classification jobs are listed with status and scope
**And** findings are enumerated with severity and data type
**And** sensitive data discovery results are retrieved

**Given** credentials without Macie access
**When** the scan executes
**Then** Macie is skipped silently

*FR: FR95 | New file: services/macie/calls.go + tests*

## Epic 4: Infrastructure & Data Service Expansion

Users discover deployment pipelines (CodeDeploy), network connections (DirectConnect), big data clusters (EMR), streaming data (Kinesis), media processing (MediaConvert), and graph databases (Neptune).

### Story 4.1: CodeDeploy Enumeration

As a pentester,
I want to enumerate CodeDeploy applications, deployment groups, and deployment configurations,
So that I can discover deployment pipelines and understand how code reaches production.

**Acceptance Criteria:**

**Given** credentials with CodeDeploy read access
**When** the scan executes
**Then** all applications are listed with names
**And** deployment groups are enumerated per application
**And** deployment configurations are retrieved

**Given** credentials without CodeDeploy access
**When** the scan executes
**Then** CodeDeploy is skipped silently

*FR: FR108 | New file: services/codedeploy/calls.go + tests*

### Story 4.2: Direct Connect Enumeration

As a pentester,
I want to enumerate Direct Connect connections, virtual interfaces, and gateways,
So that I can discover dedicated network links between AWS and on-premises infrastructure.

**Acceptance Criteria:**

**Given** credentials with Direct Connect read access
**When** the scan executes
**Then** all connections are listed with connection IDs, bandwidth, and location
**And** virtual interfaces are enumerated with VLAN and BGP details
**And** Direct Connect gateways are listed

**Given** credentials without Direct Connect access
**When** the scan executes
**Then** Direct Connect is skipped silently

*FR: FR109 | New file: services/directconnect/calls.go + tests*

### Story 4.3: EMR Cluster Enumeration

As a pentester,
I want to enumerate EMR clusters, instance groups, and security configurations,
So that I can discover big data processing infrastructure and potential data access paths.

**Acceptance Criteria:**

**Given** credentials with EMR read access
**When** the scan executes
**Then** all clusters are listed with names, status, and release labels
**And** instance groups are enumerated per cluster
**And** security configurations are retrieved

**Given** credentials without EMR access
**When** the scan executes
**Then** EMR is skipped silently

*FR: FR110 | New file: services/emr/calls.go + tests*

### Story 4.4: Kinesis Enumeration

As a pentester,
I want to enumerate Kinesis streams, shard details, and consumer applications,
So that I can discover real-time data streams and understand data flow paths.

**Acceptance Criteria:**

**Given** credentials with Kinesis read access
**When** the scan executes
**Then** all streams are listed with names and status
**And** shard details are enumerated per stream
**And** registered consumer applications are listed

**Given** credentials without Kinesis access
**When** the scan executes
**Then** Kinesis is skipped silently

*FR: FR111 | New file: services/kinesis/calls.go + tests*

### Story 4.5: MediaConvert Enumeration

As a pentester,
I want to enumerate MediaConvert jobs, queues, and presets,
So that I can discover media processing infrastructure and associated S3 input/output locations.

**Acceptance Criteria:**

**Given** credentials with MediaConvert read access
**When** the scan executes
**Then** queues are listed with names and status
**And** recent jobs are enumerated
**And** presets are listed

**Given** credentials without MediaConvert access
**When** the scan executes
**Then** MediaConvert is skipped silently

*FR: FR112 | New file: services/mediaconvert/calls.go + tests*

### Story 4.6: Neptune Enumeration

As a pentester,
I want to enumerate Neptune DB clusters, instances, and parameter groups,
So that I can discover graph database infrastructure and access configurations.

**Acceptance Criteria:**

**Given** credentials with Neptune read access
**When** the scan executes
**Then** all DB clusters are listed with identifiers, status, and endpoints
**And** DB instances are enumerated per cluster
**And** cluster parameter groups are listed

**Given** credentials without Neptune access
**When** the scan executes
**Then** Neptune is skipped silently

*FR: FR113 | New file: services/neptune/calls.go + tests*

## Epic 5: Documentation & Contributor Enablement

README reflects 63 services with speed preset and OPSEC tradeoff guidance. CONTRIBUTING.md documents concurrent testing requirements. Service template updated for concurrent-safe patterns. Contributors can confidently add new services to the concurrent architecture.

### Story 5.1: README Update with 63 Services & Speed Presets

As a user evaluating awtest,
I want the README to reflect all 63 services with speed preset documentation and OPSEC guidance,
So that I understand the tool's full capabilities and can make informed speed-vs-stealth decisions.

**Acceptance Criteria:**

**Given** the README is updated
**When** a user reads the service list
**Then** all 63 services are listed with updated categories reflecting Phase 2 additions

**Given** the README is updated
**When** a user reads the speed preset section
**Then** `--speed=safe/fast/insane` is documented with concurrency mappings
**And** OPSEC tradeoffs are clearly explained (safe = low profile, fast = moderate, insane = high visibility)
**And** `--concurrency=N` override is documented

**Given** the README is updated
**When** a user reads usage examples
**Then** examples include `--speed` flag usage alongside existing flags

*FRs: FR104-105 | Modified file: README.md*

### Story 5.2: CONTRIBUTING.md Concurrent Testing Requirements

As a contributor,
I want CONTRIBUTING.md to document concurrent testing requirements,
So that I can write services that work correctly under parallel execution.

**Acceptance Criteria:**

**Given** a contributor reads CONTRIBUTING.md
**When** they review the testing section
**Then** the requirement to run `go test -race ./...` is documented
**And** concurrent comparison testing expectations are explained
**And** the rule that services must not import sync primitives is stated

**Given** a contributor reads CONTRIBUTING.md
**When** they review the service implementation section
**Then** it explains that services are concurrency-unaware by design
**And** the safeScan wrapper is referenced as the concurrency boundary

*FR: FR106 | Modified file: CONTRIBUTING.md*

### Story 5.3: Service Template Concurrent-Safe Update

As a contributor,
I want the service template updated for concurrent-safe patterns,
So that new services I create automatically follow the correct patterns for parallel execution.

**Acceptance Criteria:**

**Given** a contributor uses the service template
**When** they create a new service
**Then** the template follows the existing Call/Process pattern with no sync imports
**And** the template includes a comment noting services must remain concurrency-unaware
**And** the template test file includes a note about race detection testing

**Given** a new service is created from the template
**When** it is registered in GetServices()
**Then** alphabetical ordering is maintained
**And** the service works correctly at all concurrency levels without modification

*FR: FR107 | Modified file: services/_template/*
