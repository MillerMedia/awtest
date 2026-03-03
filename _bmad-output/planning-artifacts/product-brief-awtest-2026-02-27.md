---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments:
  - README.md
date: 2026-02-27
author: Kn0ck0ut
---

# Product Brief: awtest

## Executive Summary

AWTest is an open-source CLI tool that enables pentesters and red teams to rapidly enumerate the permissions and accessible resources of discovered AWS credentials. In an engagement where time is critical, AWTest eliminates the manual, service-by-service enumeration process by automatically scanning across the breadth of AWS services and reporting what the credentials can actually access — not just whether permission exists, but what specific resources (buckets, instances, functions, etc.) are exposed. The vision extends toward attack path analysis, helping practitioners understand the real-world risk and exploitability of found credentials.

---

## Core Vision

### Problem Statement

When pentesters or red teamers discover AWS credentials during an engagement, determining the scope of access is a slow, manual, and error-prone process. AWS has 200+ services, and there is no fast, reliable way to enumerate what a set of credentials can do across all of them. This leads to incomplete assessments, missed critical findings, and wasted time during time-sensitive operations.

### Problem Impact

Incomplete credential enumeration means critical security findings go unreported. A set of keys with access to database backups, production infrastructure, or IAM privilege escalation paths may never be discovered if the tester doesn't check the right services. For pentesters and red teams mid-operation, every minute spent on manual enumeration is a minute not spent on deeper exploitation and finding real risk.

### Why Existing Solutions Fall Short

The current approach is largely manual — run `sts get-caller-identity`, infer what the credentials might be for based on context, and then run individual AWS CLI commands hoping to hit the right services. While some tools exist in this space, there has been no robust, open-source solution that is easy to install, fast to run, and comprehensive in coverage. Most existing options require significant setup, lack breadth of service coverage, or stop at permission detection without showing actual accessible resources.

### Proposed Solution

AWTest provides a single-command CLI experience: install via brew or `go install`, provide credentials (or use the active AWS CLI profile), and get immediate, comprehensive enumeration across AWS services. The tool shows not just permissions but actual resources — specific S3 buckets, EC2 instances, Lambda functions, and more. The roadmap extends to concurrent multi-threaded enumeration for maximum speed and attack path analysis that surfaces the highest-risk exploits possible with the discovered credentials.

### Key Differentiators

- **Open source and frictionless** — Clone, install, run. No setup, no configuration, no accounts. Available via brew and `go install`
- **Resource-level visibility** — Goes beyond "has permission" to show exactly what resources are accessible
- **Practitioner-built** — Designed by a working pentester who understands what matters in the output during real engagements
- **Speed-first architecture** — Built in Go for performance, with a roadmap toward concurrent enumeration across all services simultaneously
- **Attack path vision** — Evolving toward automated risk analysis that surfaces the highest-impact exploits possible with found credentials
- **Right timing** — AI-accelerated development is putting more infrastructure in the hands of people without deep DevOps expertise, expanding the attack surface for misconfigured credentials

## Target Users

### Primary Users

**Alex, the Engagement Pentester**

Alex is a security consultant doing external penetration tests, sometimes solo, sometimes as part of a team. During a typical engagement, Alex finds AWS credentials in a variety of ways — exposed in public GitHub repos, leaked in configuration files, discovered in S3 buckets, or extracted from compromised systems. The clock is always ticking: engagements have fixed timelines, and every hour spent manually enumerating AWS services is an hour not spent exploiting findings or writing the report.

When Alex finds credentials, the immediate need is: *"What can these do?"* Running `aws sts get-caller-identity` gives basic info, but understanding the full scope requires checking dozens or hundreds of AWS services manually. This is where awtest comes in — Alex runs a single command, and within seconds gets a comprehensive view of accessible resources across S3, EC2, Lambda, IAM, and more. The output feeds directly into the next phase: exploiting high-value targets or documenting the finding for the client report. As Alex's workflow evolves, having structured output (JSON, Markdown) that can be fed into LLMs or reporting tools makes the entire process even more efficient.

**Success looks like:** Finding credentials mid-engagement, running awtest, and immediately knowing whether they've uncovered a critical issue (access to production databases, admin IAM permissions) or a low-value finding (read-only access to a logging bucket). No guesswork, no missed opportunities, no wasted time.

### Secondary Users

**Red Teams** — Internal security teams conducting longer-term adversarial simulations. They discover credentials during lateral movement through networks, CI/CD pipelines, and code repositories. Like pentesters, they need fast enumeration but often operate in more constrained environments where stealthy, efficient tooling matters.

**Bug Bounty Hunters** — Independent researchers finding exposed credentials in web apps, public repos, or misconfigurations. They need to quickly demonstrate the impact of their finding to maximize payout — awtest helps them go from "I found keys" to "here's everything an attacker could access" in seconds.

**Cloud Security Engineers / Blue Team** — Defensive practitioners using awtest to audit their own organization's credential exposure. They run it against service accounts, IAM users, and CI/CD credentials to understand blast radius *before* an attacker does. Their focus is on risk assessment and remediation prioritization.

**Incident Responders** — Security teams reacting to credential leaks (exposed in logs, stolen in breaches, accidentally committed to repos). They need to answer "how bad is this?" immediately to determine response urgency and scope key rotation efforts.

### User Journey

**Discovery:** Users find awtest through GitHub, security community recommendations, blog posts, or conference talks. The appeal is immediate: open source, easy install via brew or `go install`, no setup required.

**Onboarding:** First use is frictionless — install the tool, run `awtest` with credentials (or use the active AWS CLI profile), and get results. No configuration files, no API keys to register, no learning curve. The tool just works.

**Core Usage:** During engagements or audits, users discover AWS credentials and immediately run awtest. The tool enumerates permissions and accessible resources across AWS services in parallel, providing clear output showing what's accessible. Users take this output and either exploit findings (offensive use) or assess risk and remediate (defensive use). As workflows mature, users may pipe awtest's structured output into LLMs, reporting tools, or automation pipelines.

**Success Moment:** The "aha!" moment happens when awtest reveals something the user wouldn't have found manually — a production RDS instance accessible via discovered credentials, an S3 bucket with backup data, or IAM permissions that enable privilege escalation. The tool surfaces findings that would have been missed in a time-constrained manual enumeration process.

**Long-term:** awtest becomes a standard tool in the user's workflow. It's the first thing they reach for after finding AWS credentials, just like nmap for network scanning or Burp Suite for web app testing. The tool's speed, coverage, and ease of use make it indispensable for anyone dealing with AWS credential assessment.

## Success Metrics

**User Success Outcomes:**

- **Speed:** Complete comprehensive AWS service enumeration in under 2 minutes for standard scans, under 5 minutes for exhaustive scans across all supported services
- **Completeness:** Detect accessible resources across services users wouldn't think to check manually, surfacing findings that would otherwise be missed
- **Trust and adoption:** Users run awtest every time they discover AWS credentials during engagements — it becomes the default first step, not an optional extra
- **Workflow integration:** Users successfully incorporate awtest output into their reporting, exploitation, and analysis workflows

**Observable User Behaviors:**
- Users complete a full credential enumeration scan within the target timeframe
- Users discover resources in services they didn't manually test (measured by service coverage in output)
- Return usage: users run the tool repeatedly across multiple engagements
- Community engagement: users file issues, submit PRs, recommend the tool to peers, reference it in blog posts or reports

### Business Objectives

**Community Growth (6-12 months):**
- Build an active open-source community around awtest
- Establish awtest as a recognized standard tool in the offensive security toolkit
- Drive awareness through security conferences, blog posts, and community recommendations
- Create sustainable community-driven development with regular contributions

**Adoption Metrics:**
- GitHub stars as a proxy for tool recognition and trust in the security community
- Download volume via brew and `go install` as a measure of actual usage
- Community contributions: issues filed, PRs submitted, forks created
- External validation: mentions in blog posts, conference talks, security training materials, pentesting reports

**Sustainability:**
- Community donations to support ongoing development (via GitHub Sponsors, Buy Me a Coffee, etc.)
- Purely community-driven with optional financial support from users who find value

### Key Performance Indicators

**Adoption KPIs:**
- GitHub stars: Track growth over time as a community trust indicator
- Installation metrics: Downloads via brew tap and `go install` (if trackable via package manager stats)
- Repository engagement: Watchers, forks, issue activity, PR submissions

**Performance KPIs:**
- Scan speed: 95% of standard scans complete in under 2 minutes
- Service coverage: Number of AWS services actively checked (expand over time as AWS releases new services)
- Resource detection accuracy: Tools successfully enumerates accessible resources without false positives

**Community Health KPIs:**
- Active contributors: Number of unique contributors submitting code, documentation, or issue reports
- Response time: How quickly issues and PRs are addressed
- External mentions: Blog posts, security tool roundups, conference references
- Donation support: Optional financial contributions from the community

## MVP Scope

### Core Features

**Comprehensive AWS Service Coverage**

Expand enumeration coverage across all high-value and commonly-used AWS services. Current implementation covers foundational services (S3, EC2, Lambda, IAM, SNS, CloudWatch, etc.) based on boto3 API exploration. The next phase systematically adds:

- **Compute & Container Services:** ECS, EKS, Fargate, Batch
- **Database Services:** RDS, DynamoDB, ElastiCache, Redshift
- **Security & Identity Services:** Secrets Manager, KMS, Certificate Manager, Cognito
- **Storage Services:** EBS, EFS, Glacier, Storage Gateway
- **Networking Services:** VPC details, Route53, CloudFront, API Gateway
- **Management Services:** CloudFormation, CloudTrail, Config, Systems Manager
- **Application Services:** SQS, Kinesis, Step Functions, EventBridge

Service additions prioritized by:
1. Security impact (services commonly holding sensitive data or privileged access)
2. Usage frequency (services pentesters encounter most often)
3. Community requests and issue reports

Each service implementation provides resource-level enumeration (not just permission checks) to surface what's actually accessible.

**Maintained Core Functionality:**
- Single-command execution with explicit credentials or AWS CLI profile
- Clean, readable output showing accessible resources per service
- Error handling for access-denied vs. service-unavailable states
- Cross-platform compatibility (installable via brew, go install)

### Out of Scope for MVP

**Deferred to Future Phases:**

- **Multi-threading/Concurrent Enumeration (Phase 2):** Parallel service scanning for sub-2-minute execution across all services
- **Advanced Output Formats (Phase 3):** JSON, Markdown, or structured exports for LLM/reporting tool integration
- **Attack Path Analysis (Phase 4):** Automated risk scoring, privilege escalation detection, blast radius assessment
- **Stealth/Evasion Features:** Rate limiting, randomization, or obfuscation techniques for operational security
- **Interactive Mode:** Real-time selection of services to scan or drill-down exploration
- **Cloud Provider Expansion:** Azure, GCP support (AWS-only for foreseeable future)

### MVP Success Criteria

**Coverage Validation:**
- Users report discovering resources in services they wouldn't have checked manually
- Community feedback indicates comprehensive coverage across common pentesting scenarios
- Service list stays current with AWS new service releases

**Adoption Signals:**
- GitHub issues requesting specific service additions (validates real-world usage)
- PRs submitted by community members adding new services (community engagement)
- Tool continues to be "run every time" for discovered credentials (trust maintained)

**Performance Threshold:**
- Even with expanded coverage, scans complete within acceptable timeframes (under 5 minutes for exhaustive scans)
- No degradation in output quality or accuracy with broader service coverage

### Future Vision

**Phase 2: Performance Optimization (Multi-threading)**
Concurrent enumeration across all AWS services simultaneously. Target: 95% of scans complete in under 2 minutes regardless of service count. Go's native concurrency makes this a natural evolution of the existing architecture.

**Phase 3: Integration & Automation (Advanced Outputs)**
Structured output formats (JSON, Markdown, CSV) that feed directly into:
- LLM analysis for automated finding summarization
- Reporting tools for client deliverables
- SIEM/logging platforms for defensive use cases
- CI/CD pipelines for continuous credential auditing

**Phase 4: Intelligence Layer (Attack Path Analysis)**
Move from "what can these credentials access?" to "what's the worst thing an attacker could do with these?"
- Privilege escalation path detection
- Data exfiltration risk scoring
- Lateral movement opportunity identification
- Blast radius visualization
- Remediation priority recommendations

**Long-term Ecosystem Vision:**
- Community-driven service coverage maintained through contributor model
- Plugin architecture for custom service checks or organization-specific resources
- Integration partnerships with pentesting platforms, security training providers
- Potential defensive tooling spin-off for blue team credential auditing
- Conference talks, training materials, book chapters establishing awtest as industry standard
