# Contributing to awtest

Thank you for your interest in contributing to awtest! The most common contribution is adding support for new AWS services.

## Adding a New AWS Service

A complete service implementation template is provided at `cmd/awtest/services/_template/`. Follow these steps:

1. **Create your service directory:** `cmd/awtest/services/yourservice/`
2. **Copy the template:** Copy `cmd/awtest/services/_template/calls.go.tmpl` to `cmd/awtest/services/yourservice/calls.go`
3. **Replace all placeholders** using the reference table in `cmd/awtest/services/_template/README.md`
4. **Implement Call()** with the actual AWS SDK v1 API call (use `WithContext` variant)
5. **Implement Process()** to extract and format discovered resources
6. **Register the service** in `cmd/awtest/services/services.go` (maintain alphabetical order)
7. **Tidy dependencies:** `go mod tidy` (required if importing a new SDK package)
8. **Build:** `go build ./cmd/awtest`
9. **Test:** `make test`

See `cmd/awtest/services/_template/README.md` for detailed instructions and `example_calls.go.reference` for an annotated real-world example.

## Service Validation Checklist

Before submitting a new service, verify:

- [ ] Package name is lowercase with no underscores
- [ ] Exported variable follows `DISPLAYNAMECalls` naming convention
- [ ] `Call()` uses `WithContext` variant for timeout support
- [ ] `Process()` handles errors via `utils.HandleAWSError`
- [ ] `Process()` returns proper `ScanResult` entries with all fields populated
- [ ] `Process()` calls `utils.PrintResult()` for console output
- [ ] Service is registered in `services.go` in alphabetical order
- [ ] `go build ./cmd/awtest` succeeds
- [ ] `make test` passes with no regressions

## Development Requirements

- **Go:** 1.19+
- **AWS SDK:** v1 (`github.com/aws/aws-sdk-go`) -- do not use SDK v2
