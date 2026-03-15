# Story 11.1: Version Update Check Flag

Status: done

## Story

As a security practitioner,
I want to check if a newer version of awtest is available from the command line,
So that I always have the latest features and security fixes without manually checking GitHub.

## Acceptance Criteria

1. **Update available:** `awtest --check-update` queries `https://api.github.com/repos/MillerMedia/awtest/releases/latest`, compares the `tag_name` (stripped of `v` prefix) against the running binary's `Version`, and prints the new version plus upgrade instructions.
2. **Up to date:** When the running version matches or exceeds the latest release, prints `awtest vX.Y.Z is up to date`.
3. **Homebrew detection:** When the binary path (via `os.Executable()`) contains `/Cellar/` or `/homebrew/`, upgrade instructions show `brew upgrade awtest`.
4. **Binary download fallback:** When not detected as Homebrew, upgrade instructions show the GitHub releases URL: `https://github.com/MillerMedia/awtest/releases/latest`.
5. **Network failure:** On any HTTP error (timeout, DNS, no internet, non-200 status, JSON parse failure), prints `Warning: Unable to check for updates` to stderr and exits with code 0.
6. **Timeout:** The HTTP request uses a 5-second timeout (`http.Client{Timeout: 5 * time.Second}`). The check never blocks longer than 5 seconds.
7. **Exit behavior:** `--check-update` runs the check and exits immediately (like `--version`), before any AWS credential validation or scanning.
8. **Dev version skip:** When `Version == "dev"` (local build), prints `Development build — skipping update check` and exits 0.

## Tasks / Subtasks

- [x] Task 1: Create `cmd/awtest/update_check.go` (AC: 1, 2, 3, 4, 5, 6, 8)
  - [x] 1.1 Define `checkForUpdate(currentVersion string)` function
  - [x] 1.2 HTTP GET to GitHub Releases API with 5s timeout, `User-Agent: awtest/{version}`
  - [x] 1.3 Parse JSON response — only need `tag_name` field (use anonymous struct, not full API model)
  - [x] 1.4 Strip `v` prefix from both `tag_name` and `currentVersion` for comparison
  - [x] 1.5 Semver comparison: split on `.`, compare major/minor/patch as integers
  - [x] 1.6 Detect Homebrew install via `os.Executable()` path check
  - [x] 1.7 Print appropriate output to stdout (update available, up to date, or dev build)
  - [x] 1.8 Print warnings to stderr on error, return nil error (graceful failure)
- [x] Task 2: Add `--check-update` flag to `main.go` (AC: 7)
  - [x] 2.1 Add `flag.Bool("check-update", false, "Check if a newer version is available")`
  - [x] 2.2 Handle after `flag.Parse()`, before AWS session — same block as `--version`
  - [x] 2.3 Call `checkForUpdate(Version)` and `os.Exit(0)`
- [x] Task 3: Create `cmd/awtest/update_check_test.go` (AC: all)
  - [x] 3.1 Test: update available — mock HTTP response with newer tag_name
  - [x] 3.2 Test: up to date — mock HTTP response with same version
  - [x] 3.3 Test: network error — mock HTTP failure, verify graceful exit
  - [x] 3.4 Test: dev build skip — verify "dev" version short-circuits
  - [x] 3.5 Test: malformed JSON — verify graceful failure
  - [x] 3.6 Test: Homebrew path detection — verify brew upgrade instructions
  - [x] 3.7 Test: semver comparison logic (1.0.0 < 1.0.1, 1.0.0 < 1.1.0, 1.0.0 < 2.0.0)
  - [x] 3.8 Use `httptest.NewServer` for HTTP mocking (stdlib, no external deps)

## Dev Notes

### Architecture & Implementation Constraints

- **No new dependencies.** Use Go stdlib only: `net/http`, `encoding/json`, `os`, `strings`, `strconv`, `fmt`. The project has zero HTTP client dependencies today — keep it that way.
- **Go 1.19 compatibility.** Do not use any APIs added after Go 1.19 (e.g., no `errors.Join`, no `slices` package, no `os.IsTerminal`).
- **File placement:** New files go in `cmd/awtest/` alongside `main.go`, `speed.go`, `worker_pool.go`. Follow existing `snake_case.go` naming.
- **Flag convention:** Use hyphenated flag name `--check-update` (matches `--output-file`, `--exclude-services` pattern).
- **Exit pattern:** Mirror the `--version` flag pattern exactly (lines 69-72 of main.go): check after `flag.Parse()`, print, `os.Exit(0)`.
- **Output convention:** User-facing output to stdout. Warnings/errors to stderr. Matches existing project pattern.

### GitHub Releases API Details

- **Endpoint:** `GET https://api.github.com/repos/MillerMedia/awtest/releases/latest`
- **No authentication required** for public repos (rate limit: 60 requests/hour unauthenticated)
- **Response fields needed:** Only `tag_name` (string, e.g., `"v1.0.0"`)
- **Set `User-Agent` header** — GitHub API requires it, returns 403 without one
- **Accept header:** `application/vnd.github+json` (recommended but not required)

### Semver Comparison Strategy

Do NOT import a semver library. Simple integer comparison is sufficient since this project uses strict `X.Y.Z` versioning:

```go
func isNewer(latest, current string) bool {
    // Strip "v" prefix from both
    // Split on ".", parse each segment as int
    // Compare major, then minor, then patch
}
```

Edge cases to handle:
- `v` prefix on tag_name (GitHub tags use `v1.0.0`)
- Version set to `"dev"` for local builds (skip check entirely)
- Malformed version strings (treat as "unable to compare", don't crash)

### Homebrew Detection

```go
func isHomebrewInstall() bool {
    exe, err := os.Executable()
    if err != nil { return false }
    return strings.Contains(exe, "/Cellar/") || strings.Contains(exe, "/homebrew/")
}
```

This works for both Intel (`/usr/local/Cellar/`) and Apple Silicon (`/opt/homebrew/Cellar/`) Homebrew installations.

### Expected Output Examples

**Update available (Homebrew):**
```
awtest v1.1.0 is available (you have v1.0.0)

Upgrade with:
  brew upgrade awtest
```

**Update available (binary):**
```
awtest v1.1.0 is available (you have v1.0.0)

Download the latest release:
  https://github.com/MillerMedia/awtest/releases/latest
```

**Up to date:**
```
awtest v1.0.0 is up to date
```

**Dev build:**
```
Development build — skipping update check
```

**Network error (stderr):**
```
Warning: Unable to check for updates
```

### Testing Strategy

- Use `httptest.NewServer` (Go stdlib) to mock the GitHub API — no external mocking libraries needed
- Extract the API URL as a package-level variable (e.g., `var releaseURL = "https://..."`) so tests can override it
- Table-driven tests for semver comparison covering: equal, patch bump, minor bump, major bump, downgrade, malformed
- Test `isHomebrewInstall()` by checking the detection logic with known paths
- Capture stdout/stderr in tests using `os.Pipe()` or buffer redirection
- Run with `-race` flag (standard for this project)

### Project Structure Notes

- New files align with `cmd/awtest/` convention — no new packages needed
- This is a standalone feature with no cross-cutting concerns to other components
- No changes needed to formatters, services, types, or utils packages
- GoReleaser ldflags already inject `Version` — no changes to `.goreleaser.yaml` needed

### References

- [Source: _bmad-output/planning-artifacts/epics-phase3.md#Story 11.1]
- [Source: _bmad-output/planning-artifacts/prd-phase3.md#FR210-FR214]
- [Source: _bmad-output/planning-artifacts/prd-phase3.md#NFR200]
- [Source: cmd/awtest/main.go — version handling at lines 21-24, 57, 69-72]
- [Source: .goreleaser.yaml — ldflags at lines 20-23]
- [Source: .github/workflows/sync-public.yml — public repo is MillerMedia/awtest]
- [Source: go.mod — Go 1.19, module path github.com/MillerMedia/awtest]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- All 9 tests pass with `-race` flag
- Full regression suite passes (all packages)
- Build compiles cleanly with `go build`

### Completion Notes List

- Created `update_check.go` with `checkForUpdate()`, `isNewer()`, and `isHomebrewInstall()` functions
- Zero new dependencies — pure Go stdlib (`net/http`, `encoding/json`, `strconv`, `strings`)
- Added `--check-update` flag to `main.go` following the exact `--version` pattern (after flag.Parse, before AWS session)
- Created 9 tests covering: semver comparison (16 table-driven cases), update available, up to date, network error, dev build, malformed JSON, non-200 status, User-Agent header
- All tests use `httptest.NewServer` for HTTP mocking (stdlib)
- `releaseURL` is a package-level var so tests can override without modifying production code
- Graceful failure on all error paths — never returns non-nil error, never crashes

### Change Log

- 2026-03-14: Implemented version update check feature (Story 11.1)

### File List

- cmd/awtest/update_check.go (NEW)
- cmd/awtest/update_check_test.go (NEW)
- cmd/awtest/main.go (MODIFIED — added --check-update flag and handler)
