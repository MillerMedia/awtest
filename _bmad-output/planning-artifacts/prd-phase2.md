---
stepsCompleted: ['step-01-init', 'step-02-discovery', 'step-02b-vision', 'step-02c-executive-summary', 'step-03-success', 'step-04-journeys', 'step-05-domain', 'step-06-innovation', 'step-07-project-type', 'step-08-scoping', 'step-09-functional', 'step-10-nonfunctional', 'step-11-polish', 'step-12-complete']
inputDocuments:
  - product-brief-awtest-2026-02-27.md
  - prd.md
  - architecture.md
  - epics.md
  - sprint-status.yaml
workflowType: 'prd'
briefCount: 1
researchCount: 0
brainstormingCount: 0
projectDocsCount: 4
classification:
  projectType: 'CLI Tool / Security Tool'
  domain: 'Cybersecurity / Offensive Security / Pentesting'
  complexity: 'Medium-High'
  projectContext: 'brownfield'
---

# Product Requirements Document - awtest (Phase 2)

**Author:** Kn0ck0ut
**Date:** 2026-03-07

## Executive Summary

AWTest Phase 2 transforms the tool from comprehensive to *fast* while expanding coverage to 57 AWS services. With Phase 1 complete — 46 services, 5 output formats, service filtering, Homebrew distribution, and full documentation — awtest already delivers the broadest credential enumeration coverage in the open-source security toolkit. Phase 2 delivers a dual-impact release: concurrent goroutine-based enumeration targeting sub-30-second scans, combined with 11 new high-value security services (ECR, Organizations, GuardDuty, Security Hub, CodeBuild, CodeCommit, OpenSearch, SageMaker, Backup, Athena, Macie).

The concurrent enumeration infrastructure is already in place — context-aware service calls, cancellation support, timeout handling, concurrency flag with validation (1-20 workers), and comprehensive test scaffolding. Phase 2 implements the goroutine worker pool and introduces named speed presets (`--speed=safe/fast/insane`) modeled after nmap's timing templates, giving practitioners intuitive control over the speed-vs-OPSEC tradeoff.

The result: "57 AWS services scanned in under 10 seconds." Speed + breadth becomes the combination that makes awtest the undisputed default — as automatic and fast as `nmap` for network scanning.

### What Makes This Special

**Speed as a Force Multiplier.** Phase 1 proved awtest's coverage is comprehensive. Phase 2 makes that coverage feel instant. For practitioners in time-constrained engagements, the difference between "results in 90 seconds" and "results in 8 seconds" changes how they use the tool: from "run it and go do something else" to "run it and immediately act on findings."

**nmap-Style Speed Control.** Named presets (`safe`, `fast`, `insane`) communicate the speed-vs-risk tradeoff intuitively. Practitioners already understand this mental model from nmap's `-T0` through `-T5`. Safe mode preserves sequential behavior for stealth-sensitive engagements; insane mode delivers maximum speed for time-critical scenarios.

**Security-Focused Service Expansion.** The 11 new services target what pentesters and red teamers actually need: container registries with embedded secrets (ECR), organizational account mapping (Organizations), detection coverage gaps (GuardDuty/Security Hub), source code access (CodeCommit), build secrets (CodeBuild), and sensitive data maps (Macie). These are the services that tools like Pacu and CloudFox prioritize.

## Project Classification

**Project Type:** CLI Tool / Security Tool
**Domain:** Cybersecurity / Offensive Security / Pentesting
**Complexity:** Medium-High (goroutine concurrency patterns, race condition prevention, rate limit resilience, 11 new AWS service integrations)
**Project Context:** Brownfield — Phase 2 builds on fully completed Phase 1 (5 epics, 29 stories, all done)

## Success Criteria

### User Success

**Speed Revelation:** Users run awtest with `--speed=fast` or `--speed=insane` and get comprehensive results across 57 services in seconds — a transformative experience compared to sequential scanning. The "aha" moment: results appear before the user can context-switch to another task.

**Control Confidence:** Users understand the speed-vs-risk tradeoff intuitively through named presets (safe/fast/insane). Practitioners coming from nmap immediately grasp the mental model. They choose the right speed for their scenario — stealth engagement vs. time-crunch vs. lab environment.

**Result Trust:** Even at `--speed=insane`, the tool delivers complete, accurate results. Built-in exponential backoff handles AWS throttling transparently. Users never wonder if results are incomplete because of rate limiting.

### Business Success

**"Best Tool" Positioning:** Phase 2 establishes awtest as the fastest, most comprehensive AWS credential enumeration tool available. Speed + breadth becomes the undeniable differentiator.

**Community Signal:**
- Phase 2 release generates visible community engagement — blog posts, tweets, security tool roundup mentions
- GitHub stars trajectory accelerates post-release as speed story spreads
- Docker pulls provide a new adoption metric beyond brew/go install

**Conference-Ready:** Phase 2 produces a tool worthy of demo at BSides/DEF CON — the speed demo ("watch 57 services enumerate in 8 seconds") is inherently impressive and shareable.

### Technical Success

**Concurrency Correctness:** Goroutine worker pool produces identical results to sequential scanning — no race conditions, no dropped results, no data corruption. Verified through comprehensive testing comparing sequential vs. concurrent output.

**Rate Limit Resilience:** Exponential backoff with retry handles AWS API throttling at all concurrency levels. Tool degrades gracefully — slows down rather than failing or producing incomplete results.

**Preset Mapping:** Named speed presets map to well-tested concurrency configurations. The `--concurrency` numeric flag remains available as a power-user override.

**Memory Discipline:** Memory footprint stays under 100MB even at `--speed=insane` with all 57 services running concurrently.

### Measurable Outcomes

**Within 1 Month:**
- Concurrent scanning implemented and tested at all speed presets
- `--speed=insane` completes full scan in under 30 seconds (target)
- `--speed=fast` completes full scan in under 60 seconds
- Zero result discrepancies between `--speed=safe` and `--speed=insane`

**Within 3 Months:**
- Phase 2 release tagged and distributed via Homebrew and direct download
- 11 new services shipped and tested
- First community feedback on speed improvements validates the positioning

**Within 6 Months:**
- awtest referenced as "fastest AWS enumeration tool" in security community
- Docker pulls as measurable secondary adoption metric

## Product Scope

### MVP - Minimum Viable Product (Phase 2 Core)

**MVP Approach:** Dual-impact release — concurrent enumeration (speed) combined with 11 high-value service additions (breadth). "57 AWS services scanned in under 10 seconds" is a compelling headline that reinforces both differentiators.

**1. Concurrent Enumeration Engine**
- Goroutine worker pool with `--speed` presets (safe/fast/insane)
- `--concurrency=N` power-user override (1-20 workers)
- Per-service exponential backoff with jitter for rate limit resilience
- Thread-safe result collection, deterministic output ordering by service name
- In-place progress reporting (`"Scanning... 15/57 services complete"`)

**2. New Service Additions (11 services)**

*Critical Priority:*
- **ECR** (Elastic Container Registry) — container image enumeration, repository policies
- **Organizations** — account structure, OUs, service control policies
- **GuardDuty** — detectors, findings, filters (detection coverage gaps)
- **Security Hub** — aggregated findings, enabled products, compliance status

*High Priority:*
- **CodeBuild** — build projects, environment variables, build history
- **CodeCommit** — repositories, branches (source code access)
- **OpenSearch** — domains, access policies, encryption config
- **SageMaker** — notebook instances, endpoints, models, training jobs
- **Backup** — backup vaults, plans, recovery points, vault policies
- **Athena** — workgroups, saved queries, query history
- **Macie** — findings, classification jobs, sensitive data map

**3. Updated Documentation**
- README updated with 57 services, speed presets, new service categories
- CONTRIBUTING.md updated with concurrent testing requirements
- Service template updated for concurrent-safe patterns

**Success Threshold:**
- `--speed=insane` delivers sub-30-second full scans
- Zero result discrepancies vs. sequential mode
- Rate limiting handled transparently without user intervention

### Growth Features (Post-MVP)

**Additional Services (documented, not yet implemented):**
- RAM (Resource Access Manager) — cross-account resource sharing
- Transit Gateway — network topology mapping
- EMR — Hadoop/Spark cluster enumeration
- CodeArtifact — package repository enumeration
- Inspector — vulnerability findings without scanning
- FSx — file system enumeration (Windows File Server, Lustre, ONTAP)
- Bedrock — AI models, custom model jobs
- DocumentDB — MongoDB-compatible cluster enumeration
- Neptune — graph database enumeration

**Distribution Expansion:**
- Docker image for CI/CD integration
- APT packages for Debian/Ubuntu
- Yum/DNF packages for RHEL/Fedora/CentOS

**Modernization & Configuration:**
- AWS SDK v1 to v2 migration (major refactor across all services)
- YAML/JSON config file support with persistent scan profiles
- Project rename (awtest → awscan) — if decided

### Vision (Future)

**Phase 3: Intelligence Layer + Attack Path Analysis**
- Automated risk scoring by credential blast radius
- Privilege escalation path detection
- Data exfiltration opportunity identification
- Lateral movement mapping
- Blast radius visualization
- Remediation priority recommendations

### Risk Mitigation Strategy

**Technical Risks:**
- *Concurrency correctness:* Mitigated by existing infrastructure (context support, cancellation, timeout handling) and comprehensive test suite comparing sequential vs. concurrent output
- *Rate limiting at scale:* Mitigated by per-service backoff with jitter, named presets communicating risk level, and safe mode as default
- *AWS session thread safety:* AWS SDK v1 session.Session documented as safe for concurrent use — verified with explicit concurrent tests

**Market Risks:**
- *Speed claims need validation:* Real-world scan timing across different AWS account sizes and credential permission levels needed before marketing sub-30-second claims
- *Service additions must maintain quality:* Each new service follows established template pattern with table-driven tests — quality bar maintained through consistency

**Resource Risks:**
- *11 services is ambitious:* Each service follows a well-established template (30-60 min implementation each). If resource-constrained, Critical tier (4 services) ships first, High tier follows
- *Concurrent engine is the hard part:* If concurrency takes longer than expected, service additions can ship independently in a point release

## User Journeys

### Alex, the Engagement Pentester — Speed as a Weapon

**Opening Scene:** Alex is on day 2 of a 5-day fintech engagement. They've already found three sets of AWS credentials — hardcoded in a Lambda environment variable, leaked in a .env file on a staging server, and extracted from a developer's compromised laptop. In Phase 1, Alex would have run awtest three times sequentially, each taking 90+ seconds. That's 5 minutes of waiting just on enumeration — time that compounds when you're juggling multiple credential sets across an engagement.

**Rising Action:** Alex runs `awtest --aki=<key1> --sak=<secret1> --speed=insane --format=json --output-file=creds1.json`. Before they can even tab over to start documenting the first credential set's provenance, awtest is done. Eight seconds. 57 services enumerated. Alex immediately pipes the JSON into their analysis workflow, then runs the second set. Eight more seconds. Then the third. In under 30 seconds total, Alex has comprehensive enumeration across all three credential sets — work that would have taken nearly 5 minutes before.

**Climax:** The speed transforms Alex's workflow from "batch and wait" to "rapid-fire triage." Instead of running awtest and context-switching to something else while it scans, Alex stays in flow — discover credentials, enumerate, assess, move on. The third credential set reveals access to a production RDS instance and an S3 bucket with database backups. Alex catches it immediately because they're still in the credential investigation mindset, not distracted by waiting.

**Resolution:** Alex documents all three credential sets with comprehensive enumeration evidence in a fraction of the time. The engagement report is thorough because awtest's speed kept Alex in the zone — no context-switching, no lost threads. The client gets three critical findings instead of the one Alex might have had time to fully investigate at sequential speeds.

### Riley, the Bug Bounty Hunter — Insane Mode in a Race

**Opening Scene:** Riley spots AWS credentials in a web application's client-side JavaScript at 11 PM. The bug bounty program is competitive — other hunters are active, and duplicate reports pay nothing. Every second between "found credentials" and "submitted report with impact evidence" matters. Riley needs to demonstrate blast radius *fast*.

**Rising Action:** Riley runs `awtest --aki=<key> --sak=<secret> --speed=insane`. They chose insane deliberately — these aren't their credentials, rate limiting the target's AWS account isn't their concern, and speed-to-report is everything. The scan completes in under 10 seconds. Riley sees access to three S3 buckets, a Secrets Manager entry, and two Lambda functions with environment variables.

**Climax:** While a competitor might still be manually running `aws s3 ls` and `aws lambda list-functions`, Riley already has the full picture. The Secrets Manager entry contains database credentials — that's the critical escalation path. Riley screenshots the awtest output, documents the attack chain (exposed JS credentials → Secrets Manager → database access), and submits the report within 5 minutes of initial discovery.

**Resolution:** Riley's report lands first. The bounty program awards a critical-severity payout because the impact chain is clearly demonstrated. The speed differential between awtest at `--speed=insane` and manual enumeration was the difference between a $5,000 payout and a "duplicate" rejection. Riley adds awtest to their permanent toolkit — it's now the first thing that runs after any credential discovery.

### Jordan, the Incident Responder — Safe Mode for Production

**Opening Scene:** 3 AM. PagerDuty fires: a service account's AWS credentials were committed to a public GitHub repo 4 hours ago. Jordan needs to assess blast radius immediately — but these are *production* credentials for their own infrastructure. Aggressive API scanning could trigger CloudTrail alerts, rate limit their own services, or — worst case — trip automated security responses that make the incident worse.

**Rising Action:** Jordan runs `awtest --aki=<key> --sak=<secret> --speed=safe`. Safe mode. Sequential scanning. Jordan knows this takes longer, but it's deliberate — a controlled, low-profile enumeration that won't create noise in their own monitoring systems. While the scan runs, Jordan pulls up CloudTrail to check for unauthorized access during the 4-hour exposure window.

**Climax:** The safe-mode scan completes in about 90 seconds. Results show the credentials have access to CloudWatch logs, a single S3 bucket with application logs, and Systems Manager parameters (read-only). No databases, no Secrets Manager, no IAM write access. This is low-severity exposure. Jordan now has confidence to make a risk-based decision: rotate credentials at 9 AM, no emergency escalation needed. If the blast radius had been worse — production databases, IAM admin — Jordan would have escalated immediately.

**Resolution:** Jordan documents the incident with authoritative blast radius evidence, rotates credentials in the morning, and closes the incident. The team avoided a 3 AM all-hands fire drill because Jordan could assess "how bad is this?" with controlled, production-safe scanning. The `--speed=safe` preset gave Jordan confidence that the assessment itself wouldn't create secondary problems.

### Sam, the Contributor — Concurrent-Aware Service Addition

**Opening Scene:** Sam uses awtest regularly and notices it doesn't cover AWS AppSync. They want to contribute it — but Phase 2's concurrent architecture means a new service needs to work correctly under parallel execution, not just sequential. Sam reads the updated CONTRIBUTING.md and sees the new testing requirements.

**Rising Action:** Sam follows the service template to implement AppSync enumeration. The implementation itself is familiar — same AWSService interface, same Call/Process pattern. But now Sam also needs to verify the service works correctly when running alongside 19 other concurrent services. Sam writes the standard table-driven tests, then runs the concurrent test suite: `make test-concurrent` verifies that AppSync results are identical whether run in safe, fast, or insane mode.

**Climax:** Sam's PR includes concurrent safety verification. The maintainer reviews it, notes that Sam properly used thread-safe patterns for result collection, and merges within a day. Sam's AppSync service works perfectly at `--speed=insane` alongside all other services — no race conditions, no result corruption.

**Resolution:** The next awtest release includes AppSync coverage, and every user benefits — whether they run safe, fast, or insane. Sam's contribution was straightforward because the concurrent architecture was designed to make service additions safe by default. The template and testing infrastructure absorbed the concurrency complexity so contributors don't have to think about goroutines.

### Journey Requirements Summary

| Journey | Speed Preset | Key Capability Revealed |
|---------|-------------|------------------------|
| Alex (Pentester) | insane | Rapid-fire multi-credential triage; speed enables staying in flow |
| Riley (Bug Bounty) | insane | Speed-to-report competitive advantage; blast radius in seconds |
| Jordan (Incident Response) | safe | Production-safe controlled scanning; risk-appropriate speed selection |
| Sam (Contributor) | all | Concurrent-safe service template; automated concurrency verification |

## Domain-Specific Requirements

### Operational Security (OPSEC)

- Concurrent API calls create a denser CloudTrail footprint — more events in a shorter time window than sequential scanning
- `--speed=insane` is inherently "noisy" in environments with GuardDuty or custom CloudTrail alerting
- Speed preset documentation must clearly communicate the OPSEC tradeoffs: safe = low profile, fast = moderate, insane = high visibility
- Users in stealth-sensitive engagements (red team, covert pentesting) need to make informed speed choices

### Read-Only Safety Under Concurrency

- Phase 1's read-only guarantee (NFR7: never create, modify, or delete AWS resources) must hold unconditionally under concurrent execution
- Concurrent architecture must guarantee no unexpected API call patterns emerge from race conditions or goroutine scheduling
- All service Call() functions remain strictly read-only — no write operations regardless of execution mode

### Rate Limiting Constraints

- AWS throttles differently per service — IAM and STS are more aggressive than S3 or Lambda
- Per-account rate limit ceilings apply across all concurrent calls from the same credentials
- Exponential backoff must be per-service, not global — one throttled service must not block other concurrent services
- Retry logic must distinguish between throttling (429), service errors (503), and access denied (403) — only retry throttling

### Credential Safety Under Concurrency

- Multiple goroutines sharing the same AWS session — AWS SDK v1's session.Session must be verified thread-safe for concurrent use
- No credential data may leak through goroutine panics, stack traces, or concurrent error paths
- NFR8 (credentials never logged) must hold under all concurrent execution paths including error and panic recovery

## CLI Tool Specific Requirements

### Command Structure

**Phase 2 Flag Additions:**

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--speed` | `safe`, `fast`, `insane` | `safe` | Named speed preset controlling concurrency level |
| `--concurrency` | `1-20` | (set by speed preset) | Power-user numeric override for worker count |

**Flag Interaction Rules:**
- `--speed` sets a default concurrency value: `safe=1`, `fast=5`, `insane=20` (tunable based on testing)
- `--concurrency` overrides the speed preset's concurrency value if both are specified
- `--speed=safe` is equivalent to current Phase 1 behavior (sequential)
- All existing Phase 1 flags remain unchanged and compatible

**Preset-to-Concurrency Mapping (initial, tunable):**

| Preset | Workers | Behavior |
|--------|---------|----------|
| `safe` | 1 | Sequential, current Phase 1 behavior |
| `fast` | 5 | Moderate parallelism, low rate limit risk |
| `insane` | 20 | Maximum parallelism, highest speed |

### Concurrent Progress Reporting

**In-place updating display** for concurrent execution:
- Single line, updated in place showing: `"Scanning... 15/57 services complete"`
- Updates as each service finishes, regardless of completion order
- On completion, replaced by summary statistics
- Progress writes to stderr (not stdout) so it doesn't interfere with formatted output or piping
- Suppressed when `--quiet` flag is set
- Suppressed when stdout is not a TTY (piped to file or another command)

**Sequential mode (safe):** Preserves existing `"Scanning [service_name]..."` per-service output for backward compatibility

### Output Assembly Under Concurrency

- Results arrive out of order as services complete at different speeds
- All results collected in thread-safe buffer before formatting
- Final output sorted by service name for deterministic ordering — same output regardless of which services finish first
- Formatting (JSON, YAML, CSV, table, text) only begins after all services complete (or timeout)
- This ensures output is identical between `--speed=safe` and `--speed=insane`

### Scripting Support

- Exit codes unchanged from Phase 1: 0 for success, non-zero for errors
- Timeout still exits 0 with partial results (not an error)
- `--speed` flag value validated — invalid preset returns error with valid options listed
- `--concurrency` and `--speed` conflict resolution is deterministic (concurrency overrides)
- All output formats produce identical structure regardless of speed preset

### Implementation Considerations

**Worker Pool Architecture:**
- Fixed-size goroutine pool sized by speed preset or `--concurrency` value
- Services dispatched to pool as work items, not one goroutine per service
- Bounded concurrency prevents goroutine explosion
- Context propagation from main timeout to all workers
- Graceful shutdown: on timeout or cancel, drain in-progress services before formatting output

**Thread Safety Requirements:**
- Result slice protected by mutex or channel-based collection
- Progress counter uses atomic operations
- AWS session shared read-only across goroutines (verified thread-safe in SDK v1)
- Error collection thread-safe — per-service errors don't corrupt other results

**Rate Limit Handling:**
- Per-service exponential backoff with jitter
- Initial backoff: 100ms, max backoff: 5s, max retries: 3
- Throttled service retries independently — doesn't block or slow other concurrent services
- Backoff state is per-goroutine, no shared retry state between services

## Functional Requirements

### Concurrent Enumeration

- **FR67:** System executes service enumeration concurrently using a configurable concurrent execution engine
- **FR68:** Users can select a named speed preset via `--speed` flag (safe, fast, insane)
- **FR69:** Users can override speed preset concurrency with `--concurrency=N` numeric flag (1-20)
- **FR70:** System maps speed presets to concurrency levels (safe=1, fast=5, insane=20)
- **FR71:** System produces identical results regardless of speed preset or concurrency level
- **FR72:** System collects results from concurrent services safely, preventing data corruption under parallel access
- **FR73:** System sorts final output by service name for deterministic ordering across all speed presets

### Rate Limit Resilience

- **FR74:** System retries AWS API calls that receive throttling responses (429/503) with exponential backoff
- **FR75:** System applies backoff independently per service — throttling on one service does not block others
- **FR76:** System varies retry timing across concurrent services to distribute retry load and prevent simultaneous retry storms
- **FR77:** System distinguishes between throttling (retry), access denied (skip), and service errors (report)
- **FR78:** System delivers complete results even when throttling occurs at any concurrency level

### Concurrent Progress Reporting

- **FR79:** System displays in-place updating progress during concurrent scans showing services completed vs. total
- **FR80:** System replaces progress display with summary statistics upon scan completion
- **FR81:** System writes progress to stderr to avoid interfering with stdout formatted output
- **FR82:** System suppresses progress display when `--quiet` flag is set
- **FR83:** System suppresses progress display when stdout is not a TTY (piped output)
- **FR84:** System preserves existing per-service sequential progress reporting in `--speed=safe` mode

### AWS Service Enumeration — Critical Priority

- **FR85:** System enumerates ECR container repositories, images, and repository policies
- **FR86:** System enumerates AWS Organizations accounts, organizational units, and service control policies
- **FR87:** System enumerates GuardDuty detectors, findings, and suppression filters
- **FR88:** System enumerates Security Hub findings, enabled products, and compliance status

### AWS Service Enumeration — High Priority

- **FR89:** System enumerates CodeBuild projects and build configurations
- **FR90:** System enumerates CodeCommit repositories and branches
- **FR91:** System enumerates OpenSearch domains, access policies, and encryption configurations
- **FR92:** System enumerates SageMaker notebook instances, endpoints, models, and training jobs
- **FR93:** System enumerates AWS Backup vaults, backup plans, recovery points, and vault access policies
- **FR94:** System enumerates Athena workgroups, saved queries, and query execution history
- **FR95:** System enumerates Macie findings, classification jobs, and sensitive data discovery results

### Concurrent Safety

- **FR96:** System maintains read-only operation guarantee under all concurrency levels
- **FR97:** System prevents credential data from leaking through concurrent error paths or concurrent process crashes
- **FR98:** System completes or terminates all in-progress service scans within 1 second of timeout, preserving partial results collected before cancellation
- **FR99:** System drains in-progress service scans before formatting output on timeout or cancellation

### Flag Interaction & Validation

- **FR100:** System validates `--speed` flag accepts only valid presets (safe, fast, insane)
- **FR101:** System resolves `--concurrency` and `--speed` conflict deterministically — `--concurrency` overrides preset
- **FR102:** System displays speed preset and effective concurrency level in scan output header
- **FR103:** All existing Phase 1 flags remain unchanged and compatible with new Phase 2 flags

### Documentation Updates

- **FR104:** README reflects 57 total services with updated service categories
- **FR105:** README documents speed presets with OPSEC tradeoff guidance
- **FR106:** CONTRIBUTING.md documents concurrent testing requirements for new service additions
- **FR107:** Service template updated for concurrent-safe implementation patterns

## Non-Functional Requirements

### Performance

- **NFR35:** `--speed=insane` completes full scan across all services in under 30 seconds for credentials with access to 20+ services
- **NFR36:** `--speed=fast` completes full scan in under 60 seconds for credentials with access to 20+ services
- **NFR37:** `--speed=safe` maintains equivalent performance to Phase 1 sequential scanning (no regression)
- **NFR38:** Concurrent scan time achieves at least 80% of linear scaling efficiency when doubling worker count, up to the point of rate limiting
- **NFR39:** Memory consumption remains under 100MB at `--speed=insane` with all services running concurrently
- **NFR40:** In-place progress updates render at minimum 2 updates per second during concurrent scans without flickering or tearing
- **NFR41:** Output formatting (JSON, YAML, CSV, table, text) adds less than 100ms overhead after all services complete

### Security

- **NFR42:** Read-only operation guarantee (NFR7) holds unconditionally under all concurrency levels — verified by concurrent test suite
- **NFR43:** Credential values never appear in concurrent crash stack traces, error logs, or concurrent error paths
- **NFR44:** AWS session sharing across goroutines uses verified thread-safe patterns — no credential corruption under concurrent access
- **NFR45:** No credential data is transmitted outside AWS API endpoints regardless of concurrency level (NFR12 extended)

### Reliability

- **NFR46:** Individual service failures in concurrent execution do not affect other concurrent services — fault isolation per goroutine
- **NFR47:** Concurrent scan results are identical to sequential scan results for the same credentials — verified by automated comparison tests
- **NFR48:** Concurrent process crashes in individual services are recovered without crashing the process or losing results from other services
- **NFR49:** Context cancellation (timeout) cleanly terminates all concurrent operations within 1 second — no resource leaks
- **NFR50:** Exponential backoff retries converge — throttled services eventually succeed or fail definitively, never retry infinitely
- **NFR51:** Rate limit backoff adds maximum 15 seconds total delay per service (3 retries: 100ms, 500ms, 2.5s + jitter, then fail)

### Integration

- **NFR52:** AWS session concurrent usage verified safe through explicit concurrent integration tests
- **NFR53:** All 11 new services follow existing AWSService interface pattern — no changes to core enumeration engine required (NFR26 maintained)
- **NFR54:** New services register in AllServices() maintaining alphabetical ordering convention
- **NFR55:** Concurrent execution compatible with all existing output formatters without formatter modifications

### Maintainability

- **NFR56:** Concurrency implementation encapsulated in worker pool module — service implementations remain concurrency-unaware
- **NFR57:** New service additions require no concurrency-specific code — the worker pool handles parallelism transparently
- **NFR58:** Code coverage exceeds 70% for concurrent execution paths including race condition detection tests
- **NFR59:** Concurrent vs. sequential comparison test suite runs as part of standard `make test`
