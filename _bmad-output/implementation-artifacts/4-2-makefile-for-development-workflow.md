# Story 4.2: Makefile for Development Workflow

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **developer working on awtest**,
I want **a Makefile with common development commands**,
so that **I can quickly build, test, install, and clean the project without remembering complex go commands**.

## Acceptance Criteria

1. Create `Makefile` in repository root
2. Add `VERSION` variable with default "dev" (overridable: `make VERSION=0.4.0`)
3. Add `BUILD_DATE` variable using shell date command: `$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")`
4. Add `LDFLAGS` variable embedding version and build date via `-X main.Version` and `-X main.BuildDate`
5. Implement `build` target: `go build $(LDFLAGS) -o awtest ./cmd/awtest`
6. Implement `test` target: `go test -v -race -coverprofile=coverage.out ./...`
7. Implement `test-coverage` target: `go tool cover -html=coverage.out` to view coverage report in browser
8. Implement `lint` target: `golangci-lint run` (document golangci-lint installation requirement)
9. Implement `install` target: `go install $(LDFLAGS) ./cmd/awtest`
10. Implement `clean` target: `rm -f awtest coverage.out`
11. Implement `snapshot` target: `goreleaser build --snapshot --clean` for local multi-platform testing
12. Add `.PHONY` declarations for all targets
13. Add `help` target showing available commands and descriptions
14. `make` (no target) defaults to `help` target
15. Verify: `make build` produces `awtest` binary with embedded version
16. Verify: `make test` runs all tests with race detector
17. Verify: `make install` installs to `$GOPATH/bin`
18. Verify: `make clean` removes build artifacts
19. Add `coverage.out` to `.gitignore`
20. All existing tests pass — no regressions

## Tasks / Subtasks

- [x] Task 1: Create Makefile with variables (AC: #1, #2, #3, #4)
  - [x] Create `Makefile` in repository root
  - [x] Define `VERSION ?= dev` (overridable via command line)
  - [x] Define `BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")`
  - [x] Define `LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"`
  - [x] Include `-s -w` in LDFLAGS to match GoReleaser stripping behavior
- [x] Task 2: Implement build, install, clean targets (AC: #5, #9, #10)
  - [x] `build` target: `go build $(LDFLAGS) -o awtest ./cmd/awtest`
  - [x] `install` target: `go install $(LDFLAGS) ./cmd/awtest`
  - [x] `clean` target: `rm -f awtest coverage.out` and optionally `rm -rf dist/`
- [x] Task 3: Implement test targets (AC: #6, #7)
  - [x] `test` target: `go test -v -race -coverprofile=coverage.out ./...`
  - [x] `test-coverage` target: `go tool cover -html=coverage.out`
- [x] Task 4: Implement lint and snapshot targets (AC: #8, #11)
  - [x] `lint` target: `golangci-lint run` with comment noting installation requirement
  - [x] `snapshot` target: `goreleaser build --snapshot --clean`
- [x] Task 5: Implement help target and .PHONY (AC: #12, #13, #14)
  - [x] `help` target using self-documenting pattern (grep for `##` comments)
  - [x] Set `.DEFAULT_GOAL := help`
  - [x] Add `.PHONY` declarations for all targets
- [x] Task 6: Update .gitignore (AC: #19)
  - [x] Add `coverage.out` to `.gitignore`
- [x] Task 7: Verification (AC: #15, #16, #17, #18, #20)
  - [x] Run `make build` and verify binary has embedded version via `./awtest --version`
  - [x] Run `make test` and verify all tests pass with race detector
  - [x] Run `make clean` and verify artifacts removed
  - [x] Run `make` (no args) and verify help output
  - [x] Run existing `go test ./...` to confirm no regressions

## Dev Notes

### Architecture & Constraints

- **Go version:** 1.19 — standard library only for flag parsing (no cobra)
- **Module path:** `github.com/MillerMedia/awtest`
- **Entry point:** `cmd/awtest/main.go`
- **Version variables:** `var Version = "dev"` and `var BuildDate = "unknown"` at `cmd/awtest/main.go:20-23` (already converted from const in Story 4.1)
- **GoReleaser config:** `.goreleaser.yaml` already exists (created in Story 4.1) — uses `-s -w -X main.Version={{.Version}} -X main.BuildDate={{.Date}}`
- **Binary output:** `awtest` in repo root (already in `.gitignore`)
- **GoReleaser output:** `dist/` directory (already in `.gitignore`)

### Architecture Reference Makefile

The architecture document specifies this exact Makefile structure:

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

**Additions beyond architecture reference (from epics AC):**
- Add `-s -w` to LDFLAGS (strip debug symbols, consistent with `.goreleaser.yaml`)
- Add `test-coverage` target for HTML report
- Add `snapshot` target for GoReleaser local builds
- Add `help` target as default
- Add `.DEFAULT_GOAL := help`

### Self-Documenting Help Pattern

Use the standard Makefile self-documenting pattern:

```makefile
.DEFAULT_GOAL := help

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
```

Each target gets a `## description` comment that the help target extracts. Example:
```makefile
build: ## Build awtest binary with version embedding
```

### Critical Implementation Details

**LDFLAGS must include `-s -w`:** The GoReleaser config already uses `-s -w` for stripping. The Makefile LDFLAGS should match to ensure consistent binary behavior between `make build` and `goreleaser build`. Without `-s -w`, `make build` produces ~23MB binaries vs ~17MB from GoReleaser.

**Tab characters required:** Makefile recipes MUST use literal tab characters, not spaces. This is the #1 Makefile error.

**`-race` flag on test:** The `-race` flag is Go's race condition detector. It requires CGO on some platforms. If `make test` fails due to CGO issues, the developer should note this but NOT remove `-race` — it's specified in the acceptance criteria.

**`go tool cover -html`:** This opens the coverage report in the default browser. It requires `coverage.out` to exist (from a previous `make test` run).

### What NOT To Do

- DO NOT add a `vet` target — `go vet` is run implicitly by `go test` since Go 1.10
- DO NOT add a `fmt` target — not in the acceptance criteria
- DO NOT modify `cmd/awtest/main.go` — no changes needed, Version/BuildDate vars already exist
- DO NOT modify `.goreleaser.yaml` — already configured correctly
- DO NOT create CONTRIBUTING.md yet — that's Story 5.2
- DO NOT add complex conditional logic or platform detection — keep the Makefile simple
- DO NOT use GNU Make extensions that break on BSD make (macOS) — stick to POSIX-compatible syntax where possible
- DO NOT add a `run` target — not in acceptance criteria
- DO NOT add Docker targets — not in scope

### Testing Approach

- No new Go test files needed — this story only creates a Makefile
- Verification is manual: run each `make` target and confirm output
- `make build && ./awtest --version` should show "awtest dev (built <timestamp>)"
- `make VERSION=test-ver build && ./awtest --version` should show "awtest test-ver (built <timestamp>)"
- `make test` should run all existing tests with `-race -v`
- `make clean` should remove `awtest` binary and `coverage.out`
- `make` with no args should display help

### Previous Story Intelligence (Story 4.1)

**Key learnings from Story 4.1:**
- Version variables are `var Version = "dev"` and `var BuildDate = "unknown"` at `cmd/awtest/main.go:20-23`
- `--version` flag already exists, output format: `awtest <version> (built <date>)`
- GoReleaser 2.x syntax used (e.g., `formats` plural, not `format`)
- Binary sizes with `-s -w`: 16-17MB per platform (down from 22.9MB without stripping)
- `dist/` and `/awtest` already in `.gitignore`
- `.goreleaser.yaml` already includes `brews` section for Story 4.4

**Files modified in Story 4.1:**
- `cmd/awtest/main.go` — Version const→var, added BuildDate, added --version flag
- `.goreleaser.yaml` — Created
- `.gitignore` — Added `dist/`

### Git Intelligence

**Recent commit patterns:**
- Commit messages: `"Add [feature] (Story X.Y)"`
- Stories are self-contained with minimal cross-file changes
- Story 4.1 was the first build automation story — established `.goreleaser.yaml` pattern
- No existing Makefile, CONTRIBUTING.md, or CI/CD workflows exist yet

### Files to Create

- `Makefile` — Development workflow automation (repository root)

### Files to Modify

- `.gitignore` — Add `coverage.out`

### Project Structure Notes

- `Makefile` goes in repository root (standard convention, alongside `.goreleaser.yaml`)
- Aligns with architecture doc file structure: root-level `Makefile` for local build automation
- No conflicts with existing files

### References

- [Source: _bmad-output/planning-artifacts/epics.md - Epic 4, Story 4.2, lines 934-963]
- [Source: _bmad-output/planning-artifacts/architecture.md - "Makefile for Local Development" section, lines 577-603]
- [Source: _bmad-output/planning-artifacts/architecture.md - "Build & Distribution Architecture" decision, lines 528-530]
- [Source: _bmad-output/planning-artifacts/architecture.md - Root Level file structure, lines 1190-1195]
- [Source: _bmad-output/implementation-artifacts/4-1-goreleaser-configuration-cross-platform-builds.md - Previous story patterns]
- [Source: .goreleaser.yaml - Existing GoReleaser config with ldflags pattern]
- [Source: cmd/awtest/main.go:20-23 - Version and BuildDate variables]
- [Source: .gitignore - Current entries: /awtest, /dist/]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no issues.

### Completion Notes List

- Created Makefile with all 8 targets: build, test, test-coverage, lint, install, clean, snapshot, help
- LDFLAGS include `-s -w` matching GoReleaser config for consistent binary sizes
- Self-documenting help pattern using `grep -E` with `##` comments on each target
- `.DEFAULT_GOAL := help` makes `make` with no args show help
- All verifications passed: `make build` produces binary with embedded version, `make test` runs with race detector, `make clean` removes artifacts, VERSION override works
- Added `coverage.out` to `.gitignore`
- All existing tests pass — no regressions

### File List

- `Makefile` — Created: Development workflow automation with build, test, lint, install, clean, snapshot, and help targets
- `.gitignore` — Modified: Added `coverage.out` entry

### Code Review Fixes (AI)

- **Makefile**: Added dependency checks for `golangci-lint` and `goreleaser` to provide helpful installation instructions if missing.
