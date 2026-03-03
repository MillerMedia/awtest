package formatters

import (
	"encoding/json"
	"fmt"

	"github.com/MillerMedia/awtest/cmd/awtest/types"
)

// jsonScanResult is the JSON-serializable representation of a ScanResult.
// Uses camelCase field names per Go JSON conventions and NFR20.
type jsonScanResult struct {
	ServiceName  string                 `json:"serviceName"`
	MethodName   string                 `json:"methodName"`
	ResourceType string                 `json:"resourceType"`
	ResourceName string                 `json:"resourceName"`
	Details      map[string]interface{} `json:"details"`
	Error        string                 `json:"error,omitempty"`
	Timestamp    string                 `json:"timestamp"`
}

// JSONFormatter formats scan results as JSON output.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSONFormatter instance.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format converts scan results to a pretty-printed JSON string.
func (f *JSONFormatter) Format(results []types.ScanResult) (string, error) {
	jsonResults := make([]jsonScanResult, 0, len(results))
	for _, r := range results {
		jr := jsonScanResult{
			ServiceName:  r.ServiceName,
			MethodName:   r.MethodName,
			ResourceType: r.ResourceType,
			ResourceName: r.ResourceName,
			Details:      r.Details,
			Timestamp:    r.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}
		if r.Error != nil {
			jr.Error = r.Error.Error()
		}
		jsonResults = append(jsonResults, jr)
	}

	data, err := json.MarshalIndent(jsonResults, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json formatting failed: %w", err)
	}
	return string(data), nil
}

// FileExtension returns "json" for JSON formatted output.
func (f *JSONFormatter) FileExtension() string {
	return "json"
}
