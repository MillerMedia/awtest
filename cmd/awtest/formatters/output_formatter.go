package formatters

import "github.com/MillerMedia/awtest/cmd/awtest/types"

// OutputFormatter defines the interface for all output formatters.
// Implementations of this interface can format scan results into various output formats
// such as JSON, YAML, CSV, or table formats.
type OutputFormatter interface {
	// Format takes scan results and returns formatted output string.
	// Returns an error if formatting fails.
	Format(results []types.ScanResult) (string, error)

	// FileExtension returns the file extension for this format (e.g., "json", "yaml", "txt").
	FileExtension() string
}
