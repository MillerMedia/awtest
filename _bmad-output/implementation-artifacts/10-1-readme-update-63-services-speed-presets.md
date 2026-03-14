# Story 10.1: README Update with 63 Services & Speed Presets

Status: done

<!-- Generated: 2026-03-13 by BMAD Create Story Workflow -->
<!-- Epic: 10 - Documentation & Contributor Enablement (Phase 2 Epic 5) -->
<!-- FRs: FR104, FR105 | Source: epics-phase2.md#Story 5.1 -->
<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user evaluating awtest,
I want the README to reflect all 63 services with speed preset documentation and OPSEC guidance,
So that I understand the tool's full capabilities and can make informed speed-vs-stealth decisions.

## Acceptance Criteria

1. **AC1:** The README intro line reflects **63 AWS services** (not 46) and the updated API call count.

2. **AC2:** The Features section includes `--speed` presets (`safe`, `fast`, `insane`) as a headline feature with concurrency mappings (safe=1, fast=5, insane=20).

3. **AC3:** The Flags table includes the `--speed` flag with valid values (`safe`, `fast`, `insane`), default (`safe`), and description explaining it controls scan parallelism via named presets.

4. **AC4:** The Flags table updates the `--concurrency` flag description to note it overrides the speed preset's concurrency value when both are specified.

5. **AC5:** A new **Speed Presets & OPSEC** section documents:
   - Speed preset to concurrency mapping table (safe=1, fast=5, insane=20)
   - OPSEC tradeoff guidance: safe = low profile (sequential, minimal CloudTrail density), fast = moderate parallelism, insane = high visibility (dense CloudTrail footprint, may trigger GuardDuty/alerting)
   - `--concurrency=N` as a power-user override
   - Default behavior (no `--speed` flag = `safe` = Phase 1 sequential behavior)

6. **AC6:** The Supported AWS Services section header reflects **63 services** and the accurate total API call count.

7. **AC7:** The service list is updated with all 17 new Phase 2 services organized into appropriate categories:
   - **Security & Identity** additions: ECR, GuardDuty, Macie, Organizations, Security Hub
   - **Databases** additions: Neptune, OpenSearch
   - **Compute & Containers** additions: EMR
   - **Developer Tools** additions: CodeBuild, CodeCommit, CodeDeploy
   - **Storage** additions: Backup
   - **Analytics** (new category): Athena, Kinesis
   - **Networking** additions: Direct Connect
   - **Media & ML** additions: MediaConvert, SageMaker

8. **AC8:** Usage examples include `--speed` flag usage (e.g., `awtest --speed=fast --format=json --output-file=results.json`).

9. **AC9:** Real-world use cases are updated to reference speed presets where appropriate (e.g., pentesting scenario uses `--speed=insane`, incident response uses `--speed=safe`).

10. **AC10:** No structural changes to sections that don't need updating (Installation, Output Formats, Contributing, License).

11. **AC11:** All service API calls listed in the service table match the actual `Name:` fields in the codebase's `calls.go` files — no stale or incorrect API call names.

## Tasks / Subtasks

- [x] Task 1: Count actual services and API calls from codebase (AC: 1, 6, 11)
  - [x] Count unique service directories under `cmd/awtest/services/` (should be 63) — Confirmed: 63 service directories (64 total minus _template)
  - [x] Count unique AWSService `Name:` fields across all `calls.go` files for accurate API call total — Confirmed: 117 API calls
  - [x] Verify all new Phase 2 services are registered in `services/services.go` — All 63 services registered in AllServices()

- [x] Task 2: Update README intro and Features section (AC: 1, 2)
  - [x] Update intro line: "46 AWS services" → "63 AWS services" and update API call count
  - [x] Add speed preset feature line: `--speed` presets (safe/fast/insane) with OPSEC tradeoff control
  - [x] Update "Concurrent Scanning" feature line to reference `--speed` presets alongside `--concurrency`

- [x] Task 3: Update Flags table (AC: 3, 4)
  - [x] Add `--speed` flag row: values `safe`, `fast`, `insane`; default `safe`; description explains named speed presets controlling concurrency
  - [x] Update `--concurrency` flag description to note it overrides speed preset when both specified

- [x] Task 4: Add Speed Presets & OPSEC section (AC: 5)
  - [x] Add new section after Flags table (or integrate into existing Usage section)
  - [x] Include preset-to-concurrency mapping table (safe=1, fast=5, insane=20)
  - [x] Document OPSEC tradeoffs: CloudTrail density implications for each preset
  - [x] Note `--concurrency=N` override behavior
  - [x] Note default behavior preserves Phase 1 sequential scanning

- [x] Task 5: Update service list to 63 services (AC: 6, 7, 11)
  - [x] Update section header: "46 services, 77 API calls" → "63 services, 117 API calls"
  - [x] Add ECR to Security & Identity: DescribeRepositories, ListImages, GetRepositoryPolicy
  - [x] Add GuardDuty to Security & Identity: ListDetectors, GetFindings, ListFilters
  - [x] Add Macie to Security & Identity: ListClassificationJobs, ListFindings, DescribeBuckets
  - [x] Add Organizations to Security & Identity: ListAccounts, ListOrganizationalUnits, ListPolicies
  - [x] Add Security Hub to Security & Identity: GetEnabledStandards, GetFindings, ListEnabledProductsForImport
  - [x] Add Neptune to Databases: DescribeDBClusters, DescribeDBInstances, DescribeDBClusterParameterGroups
  - [x] Add OpenSearch to Databases: ListDomains, DescribeDomainAccessPolicies, DescribeDomainEncryption
  - [x] Add EMR to Compute & Containers: ListClusters, ListInstanceGroups, ListSecurityConfigurations
  - [x] Add CodeBuild to Developer Tools: ListProjects, ListProjectEnvironmentVariables, ListBuilds
  - [x] Add CodeCommit to Developer Tools: ListRepositories, ListBranches
  - [x] Add CodeDeploy to Developer Tools: ListApplications, ListDeploymentGroups, ListDeploymentConfigs
  - [x] Add Backup to Storage: ListBackupVaults, ListBackupPlans, ListRecoveryPointsByBackupVault, GetBackupVaultAccessPolicy
  - [x] Create Analytics category: Athena (ListWorkGroups, ListNamedQueries, ListQueryExecutions), Kinesis (ListStreams, ListShards, ListStreamConsumers)
  - [x] Add Direct Connect to Networking: DescribeConnections, DescribeVirtualInterfaces, DescribeDirectConnectGateways
  - [x] Add MediaConvert to Media & ML: ListQueues, ListJobs, ListPresets
  - [x] Add SageMaker to Media & ML: ListNotebookInstances, ListEndpoints, ListModels, ListTrainingJobs
  - [x] Verify all API call names match actual `Name:` fields in codebase

- [x] Task 6: Update usage examples with speed presets (AC: 8)
  - [x] Add example: `awtest --speed=fast` — fast scan with moderate parallelism
  - [x] Add example: `awtest --speed=insane --format=json --output-file=results.json` — maximum speed with JSON output
  - [x] Update existing examples to show `--speed` alongside other flags where appropriate

- [x] Task 7: Update real-world use cases (AC: 9)
  - [x] Pentesting example: add `--speed=insane` to the command, update "90 seconds" timing reference to "seconds"
  - [x] Bug Bounty example: add `--speed=insane` to demonstrate speed advantage
  - [x] Incident Response example: add `--speed=safe` to demonstrate controlled scanning, note OPSEC rationale

- [x] Task 8: Verify no unintended changes (AC: 10)
  - [x] Confirm Installation section unchanged
  - [x] Confirm Output Formats section unchanged (except any new service references if applicable)
  - [x] Confirm Contributing section unchanged
  - [x] Confirm License section unchanged
  - [x] Confirm all existing links and badges still correct

## Dev Notes

### This is a Documentation-Only Story

This story modifies **only README.md**. No Go code changes. No tests to write or run. The acceptance criteria focus on content accuracy and completeness.

### Accurate Service Count: 63 Services

Phase 1 shipped 46 services. Phase 2 added 17 new services:
1. Athena
2. Backup
3. CodeBuild
4. CodeCommit
5. CodeDeploy
6. Direct Connect
7. ECR
8. EMR
9. GuardDuty
10. Kinesis
11. Macie
12. MediaConvert
13. Neptune
14. OpenSearch
15. Organizations
16. SageMaker
17. Security Hub

Total: 46 + 17 = **63 services**

### API Call Count — Must Be Verified from Code

The README currently says "77 API calls" for 46 services. The developer MUST count the actual `Name:` fields across all `calls.go` files to determine the accurate total. Each Phase 2 service typically has 3 API calls, but some have 2 (CodeCommit) or 4 (Backup, SageMaker).

**To count accurately:**
```bash
grep -rn 'Name:' cmd/awtest/services/*/calls.go | grep -v 'MethodName\|ServiceName\|ResourceName\|ModuleName\|DefaultModuleName' | grep '"[a-z]' | wc -l
```

Expected: approximately 128 API calls (77 original + ~51 new).

### Speed Preset Implementation — Already in Codebase

The speed presets are already implemented in `cmd/awtest/speed.go`:
- `safe` = 1 worker (sequential, Phase 1 behavior)
- `fast` = 5 workers (moderate parallelism)
- `insane` = 20 workers (maximum parallelism)

The `--concurrency` flag overrides the preset's concurrency value when explicitly set.

### OPSEC Guidance — Critical for Security Tool README

The OPSEC tradeoffs section is NOT optional. This is a security tool used by pentesters and red teamers who need to make informed decisions about API call density:

- **safe:** Sequential scanning. Minimal CloudTrail footprint. Same profile as a user clicking around the AWS console. Appropriate for stealth engagements, red team operations, and scanning your own production infrastructure.
- **fast:** 5 concurrent workers. Moderate CloudTrail density — more events in a shorter window than sequential, but within normal operational patterns. Appropriate for time-sensitive pentests where speed matters more than stealth.
- **insane:** 20 concurrent workers. Dense CloudTrail burst. All 63 services hammered simultaneously. Will create a visible spike in API call patterns. Appropriate for lab environments, time-critical bug bounty, and situations where OPSEC is not a concern.

### Service Category Organization

The service list uses collapsible `<details>` tag for readability. Phase 2 services should be integrated into the existing category tables, not listed separately. New category needed:

- **Analytics** (new): Athena, Kinesis

Services that could go in multiple categories should follow the most intuitive mapping:
- SageMaker → Media & ML (aligns with existing ML services like Rekognition)
- EMR → Compute & Containers (big data processing = compute)
- OpenSearch → Databases (search/data store)
- ECR → Security & Identity (container registry policies are security-relevant)
- Macie → Security & Identity (data classification = security)

### Existing README Structure to Preserve

```
1. Header/badges
2. One-line description
3. Features list
4. Installation (Homebrew, Go Install, Binary)
5. Usage examples
6. Flags table
7. Output Formats
8. Real-World Use Cases
9. Supported AWS Services (collapsible)
10. Contributing
11. Support the Project
12. License
```

The speed presets section should go between the Flags table and Output Formats, or as a subsection within Usage.

### Anti-Patterns to Avoid

- **DO NOT** change the ASCII art header or badge URLs
- **DO NOT** modify the Installation section
- **DO NOT** change the Output Formats section (formatters are unchanged in Phase 2)
- **DO NOT** claim specific scan timing (e.g., "under 10 seconds") — timing varies by account size and credential permissions
- **DO NOT** list services in a flat table without categories — maintain the categorized organization
- **DO NOT** forget to update the `<summary>` tag text inside the `<details>` element
- **DO NOT** change the Contributing or License sections

### Previous Story Intelligence

**From Story 5.3 (Phase 1 — README Updates):**
- README structure was established with collapsible service list, categorized tables, real-world use cases
- API call names in the README matched the `Name:` fields from the codebase

**From recent Phase 2 stories (9.1-9.6):**
- All 17 new services are implemented and registered in `services/services.go`
- Each service follows the standard AWSService pattern with Call/Process
- API call names follow the pattern `"service:APIMethod"` (e.g., `"neptune:DescribeDBClusters"`)

### Git Intelligence

Recent commits follow pattern: `"Add [service] enumeration with N API calls (Story X.Y)"` or `"Mark Story X.Y as done"`. This is a documentation story, so the commit message should be something like:
- `"Update README with 63 services, speed presets, and OPSEC guidance (Story 10.1)"`

### Project Structure Notes

**Files to MODIFY:**
```
README.md  # The only file modified in this story
```

**Files to REFERENCE (DO NOT MODIFY):**
```
cmd/awtest/speed.go                    # Speed preset implementation (safe/fast/insane mappings)
cmd/awtest/services/services.go        # AllServices() registry — verify service count
cmd/awtest/services/*/calls.go         # AWSService Name fields — verify API call names
```

### References

- [Source: epics-phase2.md#Story 5.1: README Update with 63 Services & Speed Presets] — BDD acceptance criteria
- [Source: prd-phase2.md#FR104] — README reflects 63 total services with updated service categories
- [Source: prd-phase2.md#FR105] — README documents speed presets with OPSEC tradeoff guidance
- [Source: architecture-phase2.md#Speed Preset Naming] — SpeedSafe, SpeedFast, SpeedInsane constants
- [Source: architecture-phase2.md#Concurrency Patterns] — Preset-to-concurrency mapping (safe=1, fast=5, insane=20)
- [Source: cmd/awtest/speed.go] — Speed preset resolution implementation
- [Source: cmd/awtest/services/services.go] — AllServices() with 63 service registrations
- [Source: README.md] — Current README (46 services, 77 API calls — needs update)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no blockers.

### Completion Notes List

- Verified 63 service directories (64 total minus `_template`) and 117 API call `Name:` fields from codebase
- All 63 services confirmed registered in `services/services.go` AllServices()
- Updated README intro: 46→63 services, 77→117 API calls
- Added Speed Presets feature line and updated Concurrent Scanning feature line
- Added `--speed` flag to Flags table with safe/fast/insane values and safe default
- Updated `--concurrency` description to note it overrides speed preset
- Added new "Speed Presets & OPSEC" section between Flags and Output Formats with concurrency mapping table, OPSEC tradeoff guidance, override behavior, and default behavior note
- Updated service list with all 17 new Phase 2 services in appropriate categories, plus new Analytics category
- Also corrected existing service API call listings to match actual `Name:` fields (AC11) — removed stale entries that didn't correspond to AWSService Name fields (e.g., IAM previously listed 5 calls but has 1 Name field, EC2 listed 4 but has 1, S3 listed 2 but has 1, etc.)
- Split ECS/Fargate into separate table rows since they are separate service directories
- Added speed preset usage examples (`--speed=fast`, `--speed=insane`)
- Updated all 3 real-world use cases with speed presets (pentest→insane, bounty→insane, IR→safe)
- Verified Installation, Output Formats, Contributing, License sections unchanged
- All existing tests pass — no regressions (documentation-only change)

### Change Log

- 2026-03-13: Updated README.md with 63 services, 117 API calls, speed preset documentation, OPSEC guidance, and 17 new Phase 2 services

### File List

- README.md (modified) — Updated with 63 services, speed presets, OPSEC section, new service categories
