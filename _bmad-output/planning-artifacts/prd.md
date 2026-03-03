---
stepsCompleted: ['step-01-init', 'step-02-discovery', 'step-02b-vision', 'step-02c-executive-summary', 'step-03-success', 'step-04-journeys', 'step-05-domain', 'step-06-innovation', 'step-07-project-type', 'step-08-scoping', 'step-09-functional', 'step-10-nonfunctional', 'step-11-polish']
inputDocuments:
  - product-brief-awtest-2026-02-27.md
  - README.md
workflowType: 'prd'
briefCount: 1
researchCount: 0
brainstormingCount: 0
projectDocsCount: 1
classification:
  projectType: 'CLI Tool / Security Tool'
  domain: 'Cybersecurity / Offensive Security / Pentesting'
  complexity: 'Medium-High'
  projectContext: 'brownfield'
---

# Product Requirements Document - awtest

**Author:** Kn0ck0ut
**Date:** 2026-02-28

## Executive Summary

AWTest is an open-source CLI tool establishing the industry standard for AWS credential enumeration in security assessments. When security professionals—pentesters, red teamers, bug bounty hunters, or incident responders—discover AWS credentials during engagements, awtest provides comprehensive, fast enumeration across AWS services, revealing accessible resources that manual testing would miss or take hours to uncover.

The current phase focuses on **comprehensive AWS service coverage expansion**, systematically adding high-value services across compute/containers (ECS, EKS, Fargate, Batch), databases (RDS, DynamoDB, ElastiCache, Redshift), security/identity (Secrets Manager, KMS, Certificate Manager, Cognito), storage (EBS, EFS, Glacier), networking (VPC, Route53, CloudFront, API Gateway), management (CloudFormation, CloudTrail, Config, Systems Manager), and application services (SQS, Kinesis, Step Functions, EventBridge). This builds on existing foundational service coverage (S3, EC2, Lambda, IAM, SNS, CloudWatch) to ensure pentesters achieve completeness confidence—certain they haven't missed critical findings due to time constraints or forgotten services.

Future phases deliver multi-threaded concurrent enumeration (Phase 2), structured output formats for LLM/reporting integration (Phase 3), and attack path analysis with automated risk scoring (Phase 4). Long-term vision: awtest becomes as automatic as running `nmap` for network scanning—the universally recognized default tool whenever AWS credentials are discovered.

### What Makes This Special

**Speed + Breadth = Immediate Value.** Users see comprehensive enumeration results across dozens of AWS services in seconds, discovering resources they wouldn't have manually checked. That instant breadth of coverage is the hook—the "aha!" moment where pentesters realize they can trust the tool to deliver completeness without burning engagement time.

**Frictionless By Design.** Single command execution, no setup, works immediately with explicit credentials or active AWS CLI profiles. Installation via `brew` or `go install` takes seconds. Cross-platform Go binary runs everywhere pentesters work.

**Practitioner-Built for Real Workflows.** Designed by a working pentester who understands what matters during time-sensitive engagements: resource-level visibility (not just permission checks), clear output showing what's accessible, error handling distinguishing access-denied from service-unavailable states.

**Solves the Completeness Problem.** Incomplete credential enumeration means critical security findings go unreported—database backups, production infrastructure, IAM privilege escalation paths missed because testers ran out of time or didn't check specific services. AWTest solves this through comprehensive, fast coverage giving pentesters confidence they've assessed the full blast radius of discovered credentials.

## Project Classification

**Project Type:** CLI Tool / Security Tool
**Domain:** Cybersecurity / Offensive Security / Pentesting
**Complexity:** Medium-High (AWS API integration across 200+ services, security domain expertise, community-driven open-source development)
**Project Context:** Brownfield—enhancing existing working tool with expanded service coverage and future phases for performance optimization, advanced output formats, and intelligence features

## Success Criteria

### User Success

**Discovery Moment:** Users find resources in AWS services they wouldn't have manually checked—RDS instances, DynamoDB tables, Secrets Manager entries, S3 buckets, KMS keys—creating actionable threads to pull during their engagement. The "aha!" moment is seeing awtest surface an overlooked service that becomes a critical security finding.

**Speed Confidence:** Standard scans complete in under 2 minutes, exhaustive scans in under 5 minutes. Users trust the tool to deliver comprehensive results fast enough that they don't second-guess running it on every discovered credential.

**Workflow Integration:** Users run awtest every time they discover AWS credentials during engagements—it becomes the automatic first step, not an optional extra. Output feeds directly into exploitation workflows and client reporting without manual reformatting.

**Completeness Confidence:** Users are certain they haven't missed critical findings due to time constraints or forgotten services. The tool checks services they didn't know existed or wouldn't remember to test manually.

### Business Success

**6-Month Community Validation:**
- **100 GitHub stars** — concrete validation that the tool is recognized and trusted by the security community
- **Active community engagement** — users filing issues for service requests, submitting PRs, recommending the tool to peers
- **External validation** — mentions in blog posts, security tool roundups, or conference talks demonstrating industry awareness

**12-Month Industry Recognition:**
- **Established as standard tool** in offensive security toolkit — referenced in pentesting training materials, blog posts, security assessments
- **Download growth** via brew and go install indicating real-world adoption beyond GitHub stars
- **Community contributions** — multiple external contributors adding service coverage, documentation, or features

**Sustainability:**
- Community donations via GitHub Sponsors or Buy Me a Coffee to support ongoing development
- Purely community-driven model with optional financial support from users who find value

### Technical Success

**Blazing Fast Performance:**
- **Almost instant execution** even with comprehensive service coverage across all AWS services
- Multi-threaded concurrent enumeration eliminates bottlenecks—doesn't slow down like network scanners
- **Speed + completeness** as the killer differentiator: users get exhaustive results in seconds, not minutes

**Service Coverage:**
- **Comprehensive AWS service enumeration** across compute, databases, security/identity, storage, networking, management, and application services
- Service list stays current with AWS new service releases
- Each service provides resource-level visibility, not just permission checks

**Reliability:**
- **Zero false positives** in resource detection—what the tool reports as accessible is actually accessible
- **Robust error handling** distinguishes access-denied from service-unavailable states
- Works reliably across all AWS regions without API failures or edge cases

**Maintainability:**
- Clean, extensible codebase making adding new AWS services straightforward
- Community contributors can add service coverage following established patterns
- Cross-platform compatibility (macOS, Linux, Windows) maintained as Go binary

### Measurable Outcomes

**Within 3 Months:**
- Expanded service coverage deployed covering compute/containers, databases, security/identity services
- Scan speed maintained or improved even with increased service count
- 25+ GitHub stars indicating early adoption

**Within 6 Months:**
- 100 GitHub stars achieved
- First external contributor submits PR for service coverage
- Tool mentioned in at least one blog post or security community discussion

**Within 12 Months:**
- awtest referenced as standard tool in pentesting training or security assessment guides
- Download metrics (brew/go install) show consistent month-over-month growth
- Multi-threaded concurrent enumeration (Phase 2) deployed delivering sub-2-minute scans

## Product Scope

### MVP - Minimum Viable Product (Current Phase)

**Comprehensive AWS Service Coverage Expansion:**

Add high-value services across all major AWS categories:
- **Compute & Container Services:** ECS, EKS, Fargate, Batch
- **Database Services:** RDS, DynamoDB, ElastiCache, Redshift
- **Security & Identity Services:** Secrets Manager, KMS, Certificate Manager, Cognito
- **Storage Services:** EBS, EFS, Glacier, Storage Gateway
- **Networking Services:** VPC details, Route53, CloudFront, API Gateway
- **Management Services:** CloudFormation, CloudTrail, Config, Systems Manager
- **Application Services:** SQS, Kinesis, Step Functions, EventBridge

**Service Addition Prioritization:**
1. Security impact (services commonly holding sensitive data or privileged access)
2. Usage frequency (services pentesters encounter most often)
3. Community requests via GitHub issues

**Maintained Core Functionality:**
- Single-command execution with explicit credentials or AWS CLI profile
- Clean, readable output showing accessible resources per service
- Error handling distinguishing access-denied vs. service-unavailable states
- Cross-platform compatibility (installable via brew, go install)
- Resource-level enumeration (not just permission checks)

**Success Threshold for MVP:**
- Users discover resources in services they wouldn't have manually checked
- Scan speed remains under 5 minutes for exhaustive scans even with expanded coverage
- Community feedback indicates comprehensive coverage across common pentesting scenarios

### Growth Features (Post-MVP)

**Phase 2: Multi-threaded Concurrent Enumeration**
- Parallel service scanning for blazing-fast execution
- Target: 95% of scans complete in under 2 minutes regardless of service count
- Leverage Go's native concurrency for bottleneck-free performance

**Phase 3: Advanced Output Formats**
- JSON, Markdown, CSV structured exports
- LLM-friendly output for automated finding summarization
- Integration with reporting tools for client deliverables
- SIEM/logging platform compatibility for defensive use cases

**Phase 4: Attack Path Analysis & Intelligence Layer**
- Automated risk scoring showing highest-impact exploits possible with found credentials
- Privilege escalation path detection
- Data exfiltration risk assessment
- Blast radius visualization
- Remediation priority recommendations

### Vision (Future)

**Industry Standard Status:**
- awtest becomes as automatic as running `nmap` for network scanning—universally recognized default tool whenever AWS credentials are discovered
- Referenced in pentesting certifications, training programs, and security assessment frameworks

**Community-Driven Ecosystem:**
- Plugin architecture for custom service checks or organization-specific resources
- Contributor model sustains service coverage as AWS releases new services
- Integration partnerships with pentesting platforms and security training providers

**Defensive Tooling Expansion:**
- Blue team spin-off for credential auditing and blast radius assessment
- CI/CD pipeline integration for continuous credential exposure monitoring
- Incident response workflows for rapid "how bad is this?" assessment during credential leaks

**Thought Leadership:**
- Conference talks at Black Hat, DEF CON, BSides establishing awtest as industry standard
- Blog posts, training materials, book chapters cementing the tool's place in security workflows
- GitHub stars and downloads position awtest as go-to AWS security enumeration tool

## User Journeys

### Alex, the Engagement Pentester — Completeness Under Time Pressure

**Opening Scene:** Alex is mid-engagement at a fintech client. Day 3 of a 5-day pentest. They've just found AWS access keys hardcoded in a public GitHub repo the client forgot to scrub. The clock is ticking—only 2 days left to assess, exploit, and write the report.

**Rising Action:** Alex runs `aws sts get-caller-identity` and sees the keys are valid. Now comes the tedious part: manually checking what these keys can access. S3? EC2? RDS? There are 200+ AWS services. Doing this manually would burn hours they don't have. Instead, Alex runs `awtest --aki=<key> --sak=<secret>` and waits.

**Climax:** Within 90 seconds, awtest outputs comprehensive results. Alex sees the keys have access to an RDS instance they wouldn't have thought to check manually—it's named `legacy-backup-db` and contains customer PII. This is a critical finding. Without awtest, Alex would have missed it or run out of time before checking RDS.

**Resolution:** Alex documents the RDS access as a high-severity finding in their report, demonstrates the exposure to the client, and moves on to exploitation. The engagement deliverable is thorough because awtest ensured completeness. The client gets real value, and Alex's reputation as a thorough pentester is reinforced.

### Riley, the Bug Bounty Hunter — Speed + Impact Demonstration

**Opening Scene:** Riley discovers exposed AWS credentials in a web app's client-side JavaScript while hunting on a private bug bounty program. The credentials are sitting in plain sight in a minified bundle. Riley knows this could be a high-payout finding—but only if they can demonstrate real impact. "Found credentials" alone won't maximize the bounty.

**Rising Action:** Riley needs to quickly show what an attacker could do with these keys. Time is critical—other hunters might find this too. They clone awtest via `brew install awtest`, paste the credentials, and run the scan. The tool enumerates across dozens of services while Riley drafts the initial report structure.

**Climax:** AWtest reveals access to S3 buckets containing user uploads, a Secrets Manager entry with database credentials, and a Lambda function with environment variables exposing API keys. Riley now has concrete proof of impact: "Attacker can access customer data, database credentials, and third-party API keys." This transforms a medium-severity finding into a critical-severity report.

**Resolution:** Riley submits the report with comprehensive evidence of accessible resources. The bounty program awards a high payout because the impact is clearly demonstrated, not just theoretical. AWtest turned "I found credentials" into "here's everything an attacker can access" in under 2 minutes, maximizing both speed and payout.

### Jordan, the Incident Responder — Crisis Mode Blast Radius Assessment

**Opening Scene:** It's 2 AM. Jordan's phone buzzes with a PagerDuty alert: AWS credentials for a service account were accidentally committed to a public GitHub repo 6 hours ago. The security team needs to know: How bad is this? What can these credentials access? Do they need to wake up the entire engineering team, or can this wait until morning?

**Rising Action:** Jordan pulls up the leaked credentials from the GitHub commit history. Instead of manually checking permissions or waiting for the morning team to assess, they run awtest against the exposed keys to get an immediate blast radius assessment. The scan runs while Jordan prepares the incident report template.

**Climax:** AWtest shows the credentials have read-only access to CloudWatch logs and a single S3 bucket containing application logs—no databases, no Secrets Manager, no production infrastructure. This is low-severity exposure. Jordan now has confidence to make a risk-based decision: rotate the credentials first thing in the morning, no need for emergency escalation.

**Resolution:** Jordan documents the incident with clear blast radius evidence from awtest's output, rotates the credentials at 9 AM, and closes the incident. The team avoided unnecessary emergency response because Jordan could quickly assess "how bad is this?" with authoritative data. AWtest enabled risk-based incident response instead of panic-driven escalation.

### Sam, the Open-Source Contributor — From User to Contributor

**Opening Scene:** Sam is a security engineer who uses awtest regularly during cloud security audits. During a recent engagement, they discovered credentials with access to AWS AppSync—but awtest didn't check that service. Sam had to manually enumerate AppSync, which broke their workflow. They think: "I could add this service to awtest so nobody else hits this gap."

**Rising Action:** Sam forks the awtest repo, reads the contribution guidelines, and examines existing service implementations to understand the pattern. The Go codebase is clean and well-structured—adding a new service follows a clear template. Sam implements AppSync enumeration using the AWS SDK, writes tests, and validates it works against their test credentials.

**Climax:** Sam submits a PR adding AppSync support. Within a day, a maintainer reviews it, suggests minor improvements, and merges it. Sam's contribution is now part of the tool that thousands of security professionals use. The next version of awtest includes AppSync coverage because Sam filled the gap they encountered.

**Resolution:** Other pentesters and security engineers benefit from Sam's contribution without knowing who added it—they just see awtest now checks AppSync. Sam feels ownership in the tool and continues contributing when they encounter other service gaps. The community-driven model works: users become contributors, and awtest's coverage expands organically through real-world usage.

## Development Phases & Feature Roadmap

### Phase 1: Comprehensive Coverage + Concurrent Enumeration

**Scope:** Service coverage expansion with multi-threaded performance optimization (combined from originally separate phases based on stronger release impact).

**Service Coverage Expansion:**
- Compute/containers: ECS, EKS, Fargate, Batch
- Databases: RDS, DynamoDB, ElastiCache, Redshift
- Security/identity: Secrets Manager, KMS, Certificate Manager, Cognito
- Storage: EBS, EFS, Glacier, Storage Gateway
- Networking: VPC, Route53, CloudFront, API Gateway
- Management: CloudFormation, CloudTrail, Config, Systems Manager
- Application: SQS, Kinesis, Step Functions, EventBridge

**Concurrent Enumeration Architecture:**
- Goroutine per service with worker pool pattern (10-20 concurrent workers)
- Target: 95% of scans under 2 minutes
- Memory footprint maintained under 100MB
- Graceful timeout handling and cancellation

**Success Criteria:**
- Users discover resources in previously unchecked services
- Sub-2-minute scan times achieved through concurrency
- Zero performance degradation with expanded service count

### Phase 2: Advanced Output Formats + Distribution Expansion

**Structured Output Formats:**
- JSON (structured and compact variants)
- Markdown (LLM-optimized)
- CSV for spreadsheet analysis

**Distribution Channels:**
- APT (Debian/Ubuntu)
- Yum/DNF (RHEL/Fedora/CentOS)
- Direct binary downloads via GitHub releases
- Docker image for CI/CD integration
- Snap/Flatpak for universal Linux support

**Integration Features:**
- `--output-format` flag (text, json, markdown, json-compact)
- `--output-file` for file output
- `--quiet` mode for programmatic usage

**Success Criteria:**
- LLM tools successfully consume Markdown output
- Reporting tools integrate JSON output without custom parsing
- Installation available across all major package managers

### Phase 3: Intelligence Layer + Attack Path Analysis

**Risk Analysis Features:**
- Automated risk scoring by credential blast radius
- Privilege escalation path detection
- Data exfiltration opportunity identification
- Lateral movement possibility mapping
- Remediation priority recommendations

**Visualization:**
- Blast radius visualization showing credential impact
- Attack path graphs from discovered permissions
- Risk heat maps across AWS services

**Success Criteria:**
- Users prioritize findings based on automated risk scores
- Attack paths surface privilege escalation opportunities
- Blast radius visualizations inform incident response decisions

## Functional Requirements

### Credential Input & Authentication

- **FR1:** Users can provide AWS Access Key ID and Secret Access Key via command-line flags
- **FR2:** Users can authenticate using active AWS CLI profile credentials without providing explicit credentials
- **FR3:** Users can use temporary AWS STS credentials for enumeration
- **FR4:** System validates credentials before starting enumeration
- **FR5:** System displays Account ID and User ARN for authenticated credentials
- **FR6:** System never logs or outputs full credential values in scan results

### AWS Service Enumeration

- **FR7:** System enumerates accessible S3 buckets
- **FR8:** System enumerates EC2 instances and associated resources
- **FR9:** System enumerates RDS database instances
- **FR10:** System enumerates Lambda functions
- **FR11:** System enumerates IAM users, roles, and policies
- **FR12:** System enumerates DynamoDB tables
- **FR13:** System enumerates Secrets Manager secrets
- **FR14:** System enumerates KMS encryption keys
- **FR15:** System enumerates ECS/EKS/Fargate container services
- **FR16:** System enumerates ElastiCache clusters
- **FR17:** System enumerates Redshift data warehouses
- **FR18:** System enumerates CloudFormation stacks
- **FR19:** System enumerates CloudWatch log groups and streams
- **FR20:** System enumerates SNS topics and SQS queues
- **FR21:** System enumerates API Gateway endpoints
- **FR22:** System enumerates CloudFront distributions
- **FR23:** System enumerates Route53 hosted zones
- **FR24:** System enumerates Step Functions state machines
- **FR25:** System enumerates EventBridge event buses
- **FR26:** System enumerates Certificate Manager certificates
- **FR27:** System enumerates Cognito user pools
- **FR28:** System enumerates EBS volumes and EFS file systems
- **FR29:** System enumerates Glacier vaults
- **FR30:** System enumerates CloudTrail trails and Config recorders
- **FR31:** System enumerates Systems Manager parameters

### Scan Configuration & Control

- **FR32:** Users can specify target AWS region for enumeration
- **FR33:** Users can select specific services to scan
- **FR34:** Users can exclude specific services from scans
- **FR35:** Users can set maximum scan timeout duration
- **FR36:** Users can enable verbose debug logging
- **FR37:** Users can configure concurrency level for parallel service scanning

### Resource Discovery & Reporting

- **FR38:** System provides resource-level details for each discovered resource (not just permission existence)
- **FR39:** System categorizes findings by severity level (info, warning, critical)
- **FR40:** System reports scan metadata including timestamp, region, and scan duration
- **FR41:** System distinguishes between accessible resources and access-denied services

### Output Formats & Presentation

- **FR42:** Users receive scan results in human-readable text format by default
- **FR43:** Users can export scan results in structured JSON format
- **FR44:** Users can export scan results in Markdown format optimized for LLM consumption
- **FR45:** Users can write scan results to specified file
- **FR46:** Users can suppress informational messages and output only findings (quiet mode)
- **FR47:** System displays real-time scan progress during execution
- **FR48:** System provides findings summary at scan completion

### Error Handling & Status Communication

- **FR49:** System distinguishes between access-denied errors and service-unavailable errors
- **FR50:** System handles AWS API throttling without crashing
- **FR51:** System reports invalid or revoked credentials with clear messaging
- **FR52:** System continues enumeration when individual service checks fail
- **FR53:** System provides actionable error messages including service name and error type
- **FR54:** System handles region-unavailable scenarios gracefully

### Installation & Distribution

- **FR55:** Users can install the tool via Homebrew package manager
- **FR56:** Users can install the tool via go install command
- **FR57:** System runs as single static binary without external dependencies
- **FR58:** System supports macOS (Intel and ARM), Linux (amd64 and arm64), and Windows (amd64) platforms

### Performance & Reliability

- **FR59:** System completes standard scans in under 2 minutes
- **FR60:** System completes exhaustive scans across all services in under 5 minutes
- **FR61:** System produces zero false positives in resource detection
- **FR62:** System maintains memory footprint under 100MB during execution
- **FR63:** System executes multiple service checks concurrently without blocking

### Community Contribution & Extensibility

- **FR64:** Contributors can add new AWS service enumeration following documented patterns
- **FR65:** System provides clear service implementation template for new services
- **FR66:** System validates contributed service modules for consistency with existing patterns

## Non-Functional Requirements

### Performance

- **NFR1:** Standard scans complete in under 2 minutes for credentials with access to 20+ services
- **NFR2:** Exhaustive scans across all supported services complete in under 5 minutes
- **NFR3:** Tool startup (initialization to first API call) completes in under 1 second
- **NFR4:** Concurrent service enumeration (Phase 2) reduces scan time by at least 60% compared to sequential execution
- **NFR5:** Memory consumption remains under 100MB during execution regardless of scan scope
- **NFR6:** Tool handles AWS API rate limiting gracefully without user intervention, implementing exponential backoff automatically

### Security

- **NFR7:** Tool operates in read-only mode—never creates, modifies, or deletes AWS resources
- **NFR8:** Credential values (access keys, secrets) are never logged to stdout, stderr, or files
- **NFR9:** Only Account ID and User ARN are displayed in scan output for identification purposes
- **NFR10:** Tool supports temporary STS credentials and respects credential expiration
- **NFR11:** Command-line credential parameters (--aki, --sak) are cleared from shell history recommendations in documentation
- **NFR12:** No credential data is transmitted outside AWS API endpoints (no analytics, telemetry, or external calls)

### Reliability

- **NFR13:** Tool produces zero false positives—any reported accessible resource is actually accessible
- **NFR14:** Individual service enumeration failures do not crash the entire scan
- **NFR15:** Tool continues enumeration across all services even when some services return errors
- **NFR16:** Error messages distinguish between access-denied, service-unavailable, invalid-credentials, and throttling scenarios
- **NFR17:** Tool handles network interruptions without data corruption or incomplete state
- **NFR18:** Concurrent operations (Phase 2) use proper synchronization to prevent race conditions

### Integration

- **NFR19:** Tool integrates with AWS SDK for Go following official SDK best practices and patterns
- **NFR20:** JSON output format conforms to standard JSON schema for programmatic parsing
- **NFR21:** Markdown output format is optimized for LLM consumption with clear structure and semantic sections
- **NFR22:** Tool respects AWS CLI profile configuration (region, output format preferences) when using profile credentials
- **NFR23:** Exit codes follow UNIX conventions (0 for success, non-zero for errors) for script integration

### Maintainability

- **NFR24:** Codebase follows Go standard project layout and idiomatic Go patterns
- **NFR25:** Each AWS service implementation follows consistent interface pattern for community contributions
- **NFR26:** New service additions require no changes to core enumeration engine
- **NFR27:** Code coverage exceeds 70% with unit tests for all service modules
- **NFR28:** Contribution documentation enables new contributors to add services without maintainer guidance
- **NFR29:** Pull requests are reviewed and merged within 48 hours for quality contributions

### Portability

- **NFR30:** Single static binary runs without external dependencies on all supported platforms
- **NFR31:** Tool supports macOS (Intel x86_64 and Apple Silicon arm64), Linux (amd64 and arm64), and Windows (amd64)
- **NFR32:** Cross-compilation produces platform-specific binaries under 15MB each
- **NFR33:** Tool behavior is consistent across all platforms (no platform-specific feature variations)
- **NFR34:** Installation via Homebrew, go install, and direct binary download all produce identical functional tool
