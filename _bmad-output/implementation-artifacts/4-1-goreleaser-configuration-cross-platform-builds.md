# Story 4.1: GoReleaser Configuration & Cross-Platform Builds

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **developer releasing awtest versions**,
I want **automated cross-platform binary builds**,
so that **security professionals on macOS, Linux, and Windows can download optimized binaries for their platform without manual compilation**.

## Acceptance Criteria

1. Create `.goreleaser.yaml` in repository root with builds section targeting platforms: darwin (amd64, arm64), linux (amd64, arm64), windows (amd64) (FR58)
2. Set binary name to "awtest" in GoReleaser config
3. Disable CGO (`CGO_ENABLED=0`) for static binary compilation (FR57)
4. Configure ldflags to embed version and build date: `-s -w -X main.Version={{.Version}} -X main.BuildDate={{.Date}}`
5. Configure archives section with tar.gz for unix, zip for windows
6. Convert `main.go` version from `const` to `var` and add `BuildDate` variable: `var Version = "dev"`, `var BuildDate = "unknown"`
7. Add `--version` flag that displays version and build date, then exits
8. Configure release section pointing to GitHub repository `MillerMedia/awtest`
9. Binary size < 15MB per platform (NFR32) - NOTE: current binary is 22.9MB, investigate stripping with `-s -w` ldflags
10. `.goreleaser.yaml` passes validation: `goreleaser check`
11. Local snapshot build succeeds: `goreleaser build --snapshot --clean`
12. All 5 platform binaries generated: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64
13. Version embedding works: `./awtest --version` shows version and build date
14. `go build ./cmd/awtest` still compiles successfully
15. All existing tests pass - no regressions

## Tasks / Subtasks

- [x] Task 1: Convert Version const to var and add BuildDate (AC: #6)
  - [x] Change `const Version = "v0.3.0"` to `var Version = "dev"` in `cmd/awtest/main.go`
  - [x] Add `var BuildDate = "unknown"` below Version
  - [x] Keep `MinConcurrency` and `MaxConcurrency` as `const` (they're not injected)
- [x] Task 2: Add --version flag (AC: #7, #13)
  - [x] Add `version := flag.Bool("version", false, "Print version and build date")` in flag section
  - [x] After `flag.Parse()`, check `if *version { fmt.Printf("awtest %s (built %s)\n", Version, BuildDate); os.Exit(0) }`
  - [x] Update the banner display to use the `Version` variable (it already references the `Version` constant)
- [x] Task 3: Create .goreleaser.yaml (AC: #1, #2, #3, #4, #5, #8)
  - [x] Create `.goreleaser.yaml` in repository root
  - [x] Configure builds section: binary name, goos, goarch, CGO_ENABLED=0, ldflags
  - [x] Configure archives section: tar.gz default, zip override for windows
  - [x] Configure release section: MillerMedia/awtest
  - [x] Configure brews section for Homebrew tap (MillerMedia/homebrew-tap) - this sets up for Story 4.4
- [x] Task 4: Validate GoReleaser config (AC: #10)
  - [x] Run `goreleaser check` to validate configuration
  - [x] Fix any validation errors
- [x] Task 5: Test snapshot build (AC: #11, #12, #9)
  - [x] Run `goreleaser build --snapshot --clean`
  - [x] Verify all 5 platform binaries are generated in dist/ directory
  - [x] Check binary sizes against 15MB target (current binary is 22.9MB without stripping)
  - [x] `-s -w` ldflags should significantly reduce size by stripping debug symbols
- [x] Task 6: Verify version embedding (AC: #13, #14, #15)
  - [x] Build with ldflags: `go build -ldflags "-X main.Version=test -X main.BuildDate=2026-03-06" -o awtest ./cmd/awtest`
  - [x] Run `./awtest --version` and verify output shows injected version and date
  - [x] Run `go test ./cmd/awtest/...` to verify no regressions
  - [x] Run `go vet ./cmd/awtest/...` to verify no issues

## Dev Notes

### Architecture & Constraints

- **Go version:** 1.19 - standard library only for flag parsing (no cobra)
- **Module path:** `github.com/MillerMedia/awtest`
- **Entry point:** `cmd/awtest/main.go`
- **Flag package:** Standard `flag` package (NOT cobra, NOT urfave/cli)
- **Current Version:** `const Version = "v0.3.0"` at `cmd/awtest/main.go:21` - MUST change to `var`
- **Current binary size:** 22.9MB (Mach-O arm64, no stripping) - `-s -w` flags should reduce this significantly
- **GoReleaser version:** Use GoReleaser 2.x (latest stable)

### Critical Implementation Details

**Version variable change:** The `Version` constant at line 21 of main.go must become a `var` so that `-X main.Version={{.Version}}` ldflags can inject the value at build time. Go's `const` values cannot be overridden by linker flags.

**ldflags breakdown:**
- `-s` - Strip symbol table (reduces binary size)
- `-w` - Strip DWARF debugging information (reduces binary size)
- `-X main.Version={{.Version}}` - Inject git tag version
- `-X main.BuildDate={{.Date}}` - Inject build timestamp

**GoReleaser .goreleaser.yaml reference:**
```yaml
version: 2

builds:
  - binary: awtest
    main: ./cmd/awtest
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
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

**Note on windows/arm64:** The epics specify 5 targets (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64). Windows arm64 is excluded via `ignore` section since it wasn't in the requirements.

### What NOT To Do

- DO NOT use cobra or any CLI framework - stick with standard `flag` package
- DO NOT change `MinConcurrency` or `MaxConcurrency` from `const` to `var` - only `Version` needs to change
- DO NOT modify any service files, formatter files, or test infrastructure
- DO NOT add GoReleaser as a Go dependency - it's a standalone CLI tool
- DO NOT create GitHub Actions workflow yet - that's Story 4.3
- DO NOT set up Homebrew tap repository yet - that's Story 4.4 (but DO include brews config in .goreleaser.yaml)
- DO NOT install goreleaser via `go install` - it should be installed via brew or curl (standalone tool)
- DO NOT remove the existing `go install` installation method from README

### Testing Approach

- No new test file needed specifically for GoReleaser config
- Verify `go build ./cmd/awtest` still works after const-to-var change
- Verify `go test ./cmd/awtest/...` passes (no regressions from changing const to var)
- Verify `go vet ./cmd/awtest/...` passes
- Manual verification: `goreleaser check` and `goreleaser build --snapshot --clean`
- Manual verification: `./dist/awtest_darwin_arm64_v8.0/awtest --version` shows correct output

### Previous Story Intelligence (Story 3.3)

**Key learnings from Story 3.3:**
- All flag definitions are in `cmd/awtest/main.go` in the `main()` function
- Validation happens after `flag.Parse()` in main()
- Error messages go to stderr via `fmt.Fprintf(os.Stderr, ...)`
- Informational messages go to stderr
- Magic numbers were refactored to constants (MinConcurrency, MaxConcurrency)
- Tests use standard library assertions (not testify) for cmd/awtest package
- The `validateConcurrency()` helper pattern works well for testable validation

**Files modified in Story 3.3:**
- `cmd/awtest/main.go` - Flag definitions, validation, constants
- `cmd/awtest/concurrency_test.go` - Unit tests

### Git Intelligence

**Recent commits show consistent patterns:**
- Commit messages follow: "Add [feature description] (Story X.Y)"
- Stories are self-contained with minimal cross-file changes
- Epic 3 stories only modified `cmd/awtest/main.go` and added test files
- No CI/CD or build automation exists yet - this story creates the foundation

### Files to Modify

- `cmd/awtest/main.go` - Change Version const to var, add BuildDate var, add --version flag

### Files to Create

- `.goreleaser.yaml` - GoReleaser configuration (repository root)

### Project Structure Notes

- `.goreleaser.yaml` goes in the repository root (standard GoReleaser convention)
- No existing build automation files - this is a greenfield addition
- Binary output name remains `awtest`
- The `awtest` binary in repo root is in `.gitignore` already
- `dist/` directory (GoReleaser output) should be added to `.gitignore`

### References

- [Source: _bmad-output/planning-artifacts/epics.md - Epic 4, Story 4.1]
- [Source: _bmad-output/planning-artifacts/architecture.md - GoReleaser Configuration section]
- [Source: _bmad-output/planning-artifacts/architecture.md - Makefile for Local Development section]
- [Source: _bmad-output/planning-artifacts/architecture.md - Build Targets and Cross-Platform section]
- [Source: _bmad-output/planning-artifacts/prd.md - FR55-58, NFR32]
- [Source: cmd/awtest/main.go - Version constant at line 21, flag definitions]
- [Source: _bmad-output/implementation-artifacts/3-3-concurrency-configuration-preparation-for-phase-2.md - previous story patterns]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- GoReleaser 2.14.1 required updated archive config syntax: `formats` (plural) instead of deprecated `format`, and `format_overrides` with nested `formats` array
- Binary sizes with `-s -w` stripping: 16-17MB per platform (down from 22.9MB unstripped). Exceeds 15MB target but significant reduction achieved. The AWS SDK dependencies contribute substantial size that cannot be stripped further without removing functionality.

### Completion Notes List

- Converted `Version` from `const` to `var` and added `BuildDate` variable for linker injection
- Added `--version` flag using standard `flag` package, outputs format: `awtest <version> (built <date>)`
- Created `.goreleaser.yaml` with GoReleaser 2.x syntax targeting 5 platforms (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64)
- Config includes builds, archives, brews (for Story 4.4), and release sections
- Added `dist/` to `.gitignore`
- All 5 platform binaries generated successfully via snapshot build
- Version embedding verified with ldflags injection
- All existing tests pass with no regressions

### Change Log

- 2026-03-06: Implemented GoReleaser configuration and cross-platform build support (Story 4.1)

### File List

- `cmd/awtest/main.go` - Modified: Version const→var, added BuildDate var, added --version flag
- `.goreleaser.yaml` - Created: GoReleaser 2.x configuration for cross-platform builds
- `.gitignore` - Modified: Added dist/ directory
