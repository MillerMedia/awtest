---
stepsCompleted: ['step-01-validate-prerequisites', 'step-02-design-epics', 'step-03-create-stories', 'step-04-final-validation']
inputDocuments:
  - prd-phase3.md
  - architecture-phase2.md
---

# awtest (Phase 3: Polish & Distribution) - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for awtest Phase 3, decomposing the requirements from the Phase 3 PRD into implementable stories focused on polish, distribution, and user experience improvements.

## Requirements Inventory

### Functional Requirements

FR200: Project renamed atomically — binary, module path, repo, Homebrew, imports, banner, CI/CD (DEFERRED — name TBD)
FR201: ASCII startup banner redesigned for new name (DEFERRED)
FR202: Sync workflow updated for new public repo name (DEFERRED)
FR203: Homebrew tap supports old+new name during transition (DEFERRED)
FR210: `--check-update` flag queries GitHub Releases API and compares to running version
FR211: Prints new version + upgrade instructions when update available
FR212: Prints confirmation when current version is up to date
FR213: Graceful failure on network errors — warning + clean exit
FR214: Version check respects 5-second timeout
FR220: macOS binaries code-signed with Apple Developer ID certificate
FR221: Signed binaries notarized with Apple notarization service
FR222: Signing/notarization automated in CI/CD release pipeline
FR223: Non-macOS builds unaffected
FR230: Tool submitted to security tool directories and awesome-lists
FR231: README optimized as standalone landing page
FR232: Homebrew core formula submitted if eligible
FR233: GitHub repository topics/tags set for discoverability

### NonFunctional Requirements

NFR200: Version check adds no more than 5 seconds to execution time
NFR201: Code signing adds no more than 5 minutes to CI/CD build time
NFR202: Rename must be atomic — no partial rename state
NFR203: Distribution listings accurately represent capabilities

### Additional Requirements

- Existing CI/CD pipeline (sync-public.yml → release.yml → GoReleaser) must continue to work
- Classic PAT with repo+workflow scopes already configured as GH_PAT secret
- Apple Developer Program membership required for code signing (user must enroll if not already)
- Homebrew core submission requires 50+ GitHub stars and notable usage

### FR Coverage Map

| FR | Epic | Story |
|----|------|-------|
| FR210-214 | Epic 11 | Story 11.1 |
| FR220-223 | Epic 12 | Story 12.1, 12.2 |
| FR230-233 | Epic 13 | Story 13.1, 13.2, 13.3 |
| FR200-203 | Epic 14 | Deferred — stories TBD when name is finalized |

## Epic List

- **Epic 11:** Version Update Check (Phase 3 Epic 1)
- **Epic 12:** macOS Code Signing & Notarization (Phase 3 Epic 2)
- **Epic 13:** Public Distribution & Discoverability (Phase 3 Epic 3)
- **Epic 14:** Project Rename (Phase 3 Epic 4) — DEFERRED

---

## Epic 11: Version Update Check (Phase 3 Epic 1)

**Goal:** Enable users to easily check if a newer version of awtest is available, with clear upgrade instructions tailored to their installation method.

### Story 11.1: Version Update Check Flag

As a security practitioner,
I want to check if a newer version of awtest is available from the command line,
So that I always have the latest features and security fixes without manually checking GitHub.

**Acceptance Criteria:**

**Given** the user runs `awtest --check-update`
**When** the tool queries the GitHub Releases API for the latest release
**Then** it compares the latest release tag against the running binary's version
**And** if a newer version exists, it prints the new version number and upgrade instructions
**And** if the current version is already the latest, it prints "awtest vX.Y.Z is up to date"

**Given** the user runs `awtest --check-update` with no network connectivity
**When** the GitHub API request fails (timeout, DNS, no internet)
**Then** the tool prints a warning ("Unable to check for updates") and exits with code 0
**And** the check never blocks for more than 5 seconds

**Given** the user installed via Homebrew
**When** an update is available
**Then** the upgrade instructions show `brew upgrade awtest`

**Given** the user installed via binary download
**When** an update is available
**Then** the upgrade instructions show the GitHub releases download URL

---

## Epic 12: macOS Code Signing & Notarization (Phase 3 Epic 2)

**Goal:** Eliminate macOS Gatekeeper warnings by signing and notarizing the binary, providing a frictionless first-run experience.

### Story 12.1: Apple Developer Certificate Setup & GoReleaser Signing

As a macOS user,
I want the awtest binary to be signed with an Apple Developer certificate,
So that macOS does not block it on first run or show "cannot verify developer" warnings.

**Acceptance Criteria:**

**Given** the release pipeline runs GoReleaser
**When** macOS binaries (amd64 and arm64) are built
**Then** they are code-signed with an Apple Developer ID certificate
**And** Linux and Windows builds are unaffected by signing

**Given** the signing certificate and credentials
**When** stored as GitHub Actions secrets
**Then** the release pipeline accesses them securely without exposing credentials in logs

### Story 12.2: Apple Notarization in CI/CD

As a macOS user,
I want the signed binary to pass Apple's notarization check,
So that Gatekeeper allows execution without any manual `xattr` workaround.

**Acceptance Criteria:**

**Given** a signed macOS binary
**When** submitted to Apple's notarization service during the release pipeline
**Then** the notarization completes successfully and the binary is stapled

**Given** a user downloads the notarized binary via Homebrew or GitHub releases
**When** they run it for the first time
**Then** macOS Gatekeeper allows execution without warnings or quarantine

**Given** the notarization step
**When** it runs in CI/CD
**Then** it adds no more than 5 minutes to the total build time

---

## Epic 13: Public Distribution & Discoverability (Phase 3 Epic 3)

**Goal:** Get awtest listed across security tool directories and optimize the project's public presence to drive organic discovery and adoption.

### Story 13.1: README & Repository Optimization

As a potential user discovering awtest on GitHub,
I want a clear, professional README that immediately communicates what the tool does and how to install it,
So that I can evaluate and start using it within minutes.

**Acceptance Criteria:**

**Given** a user lands on the GitHub repository
**When** they read the README
**Then** they see: one-line description, key feature highlights, installation instructions, quick-start example, and output sample
**And** the README works as a standalone landing page

**Given** the GitHub repository
**When** viewed on GitHub
**Then** topics/tags include: aws, security, enumeration, pentest, cloud-security, credential-testing, golang, cli
**And** the repository description is set to a clear one-liner

### Story 13.2: Security Tool Directory Submissions

As the project maintainer,
I want awtest listed on curated security tool directories and awesome-lists,
So that practitioners discover it when researching AWS security tools.

**Acceptance Criteria:**

**Given** the tool is production-ready (v1.0.0+)
**When** PRs are submitted to relevant awesome-lists and directories
**Then** submissions are made to at minimum: awesome-aws, awesome-security, awesome-pentest
**And** each submission follows the list's contribution guidelines

**Given** each directory submission
**When** the description is written
**Then** it accurately represents the tool's capabilities without exaggeration

### Story 13.3: Homebrew Core Submission

As a macOS/Linux user,
I want to install awtest from Homebrew core (`brew install awtest`) without needing a custom tap,
So that installation follows the standard Homebrew workflow.

**Acceptance Criteria:**

**Given** the project meets Homebrew core eligibility (notable, 50+ stars, stable releases)
**When** a formula is submitted to homebrew/homebrew-core
**Then** it follows Homebrew's formula authoring guidelines
**And** the existing tap continues to work as a fallback

**Given** the project does not yet meet eligibility requirements
**When** this story is evaluated
**Then** it is deferred until eligibility criteria are met
**And** the criteria gap is documented

---

## Epic 14: Project Rename (Phase 3 Epic 4) — DEFERRED

**Goal:** Rename the project from "awtest" to the finalized new name, updating all references atomically.

*Stories will be created when the project name is finalized. The rename touches ~1800 Go import paths, module path, binary name, Homebrew tap, GoReleaser config, CI/CD workflows, README, CONTRIBUTING.md, ASCII banner, service templates, and the public GitHub repository.*

**Key constraint:** The rename must be atomic — no intermediate state where the tool is partially renamed.
