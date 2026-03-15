---
stepsCompleted: ['quick-prd']
inputDocuments:
  - prd-phase2.md
  - architecture-phase2.md
  - epics-phase2.md
  - sprint-status.yaml
workflowType: 'prd'
classification:
  projectType: 'CLI Tool / Security Tool'
  domain: 'Cybersecurity / Offensive Security / Pentesting'
  complexity: 'Medium'
  projectContext: 'brownfield'
---

# Product Requirements Document - awtest (Phase 3: Polish & Distribution)

**Author:** Kn0ck0ut
**Date:** 2026-03-14

## Executive Summary

With Phase 1 (output formats, 46 services, filtering, Homebrew distribution) and Phase 2 (concurrent scanning, 63 services, rate limiting) shipped as v1.0.0, awtest is production-ready. Phase 3 focuses on polish and reach: renaming the project (name TBD), adding a version update check so users know when new releases land, signing macOS binaries to eliminate Gatekeeper friction, and getting the tool listed across security tool directories and awesome-lists to drive adoption.

These are not feature-heavy changes — they're the difference between a working tool and a professional one that users trust and discover organically.

### What Makes This Special

**Frictionless Install.** macOS code signing eliminates the "Apple could not verify" warning that currently requires a manual `xattr` workaround. First impressions matter — a blocked binary on first run erodes trust.

**Update Awareness.** A `--check-update` flag lets users know when a new version is available without forcing anything. Homebrew users get `brew upgrade`; binary users get a clear path to the latest release.

**Discoverability.** The best tool nobody knows about is useless. Listing awtest in security tool directories, awesome-lists, and potentially Homebrew core puts it where practitioners look when evaluating tools.

**Identity.** The rename (details TBD) gives the project a distinct, memorable name that reflects what it does — sifting through AWS services to reveal credential reach.

## Project Classification

**Project Type:** CLI Tool / Security Tool
**Domain:** Cybersecurity / Offensive Security / Pentesting
**Complexity:** Medium (no new architecture; build pipeline changes, documentation, community outreach)
**Project Context:** Brownfield — Phase 3 builds on completed Phase 1 + Phase 2 (v1.0.0 released)

## Success Criteria

### User Success

- macOS users install via Homebrew and run the tool without any Gatekeeper warnings or manual workarounds
- Users can check for updates with a single flag and see clear upgrade instructions
- The tool is discoverable through standard security tool directories and search

### Technical Success

- macOS binaries are signed and notarized, passing Gatekeeper without `xattr` workarounds
- Version check queries GitHub releases API and compares semver, with graceful failure on network issues
- CI/CD pipeline handles signing automatically on release
- Rename (when executed) updates all 1800+ import paths, module path, binary name, Homebrew tap, and documentation atomically

### Business Success

- Listed on 3+ security tool directories or awesome-lists
- Homebrew core formula submitted (if eligibility requirements met)
- GitHub stars trajectory increases through discoverability improvements

## Functional Requirements

### FR200: Project Rename (Deferred — name TBD)

FR200: The project binary, module path, GitHub repository, Homebrew formula, GoReleaser config, README, CONTRIBUTING.md, ASCII banner, all Go imports, service templates, and CI/CD workflows are renamed atomically from "awtest" to the new name.

FR201: The ASCII startup banner is redesigned to reflect the new project name and branding.

FR202: The sync workflow updates to push to the new public repository name.

FR203: Homebrew tap is updated so both old and new formula names work during a transition period (old name prints deprecation notice).

### FR210: Version Update Check

FR210: A `--check-update` flag queries the GitHub Releases API for the latest release tag and compares it to the running binary's version.

FR211: If a newer version is available, the tool prints the new version number and upgrade instructions (Homebrew command for Homebrew users, download URL for binary users).

FR212: If the current version is up to date, the tool prints a confirmation message.

FR213: The version check fails gracefully on network errors (timeout, DNS failure, no internet) — prints a warning and exits cleanly, never blocks normal operation.

FR214: The version check respects a reasonable timeout (5 seconds max) to avoid stalling.

### FR220: macOS Code Signing

FR220: macOS binaries (amd64 and arm64) are code-signed with an Apple Developer ID certificate during the GoReleaser build process.

FR221: Signed binaries are notarized with Apple's notarization service so Gatekeeper allows execution without warnings.

FR222: The signing and notarization process is automated in the CI/CD release pipeline (no manual steps).

FR223: Non-macOS builds (Linux, Windows) are unaffected by signing changes.

### FR230: Public Distribution & Discoverability

FR230: The tool is submitted to relevant security tool directories and awesome-lists, including but not limited to: awesome-aws, awesome-security, awesome-pentest, and similar curated lists.

FR231: The README is optimized for discoverability — clear one-line description, badges, feature highlights, and installation instructions that work as a standalone landing page.

FR232: A Homebrew core formula is submitted if the project meets eligibility requirements (notable, 50+ stars, etc.).

FR233: The GitHub repository has proper topics/tags set (aws, security, enumeration, pentest, cloud-security, etc.).

## Non-Functional Requirements

NFR200: Version check adds no more than 5 seconds to execution time (with timeout).

NFR201: Code signing does not increase CI/CD build time by more than 5 minutes.

NFR202: The rename must be atomic — no intermediate state where the tool is partially renamed.

NFR203: All distribution listings must accurately represent the tool's capabilities without exaggeration.
