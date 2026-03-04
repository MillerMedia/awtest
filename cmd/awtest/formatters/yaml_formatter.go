package formatters

import (
	"fmt"
	"time"

	"github.com/MillerMedia/awtest/cmd/awtest/types"
	"gopkg.in/yaml.v3"
)

// yamlScanResult is the YAML-serializable representation of a ScanResult.
type yamlScanResult struct {
	ServiceName  string                 `yaml:"serviceName"`
	MethodName   string                 `yaml:"methodName"`
	ResourceType string                 `yaml:"resourceType"`
	ResourceName string                 `yaml:"resourceName"`
	Details      map[string]interface{} `yaml:"details"`
	Error        string                 `yaml:"error,omitempty"`
	Timestamp    string                 `yaml:"timestamp"`
}

// YAMLFormatter formats scan results as YAML output.
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAMLFormatter instance.
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format converts scan results to YAML for human-readable structured output.
func (f *YAMLFormatter) Format(results []types.ScanResult) (string, error) {
	yamlResults := make([]yamlScanResult, 0, len(results))
	for _, r := range results {
		yr := yamlScanResult{
			ServiceName:  r.ServiceName,
			MethodName:   r.MethodName,
			ResourceType: r.ResourceType,
			ResourceName: r.ResourceName,
			Details:      map[string]interface{}{},
			Timestamp:    r.Timestamp.Format(time.RFC3339),
		}
		if r.Details != nil {
			if marshalErr := tryMarshalYAML(r.Details); marshalErr == "" {
				yr.Details = r.Details
			} else {
				yr.Error = fmt.Sprintf("details serialization error: %v", marshalErr)
			}
		}
		if r.Error != nil {
			if yr.Error != "" {
				yr.Error = r.Error.Error() + "; " + yr.Error
			} else {
				yr.Error = r.Error.Error()
			}
		}
		yamlResults = append(yamlResults, yr)
	}

	data, err := yaml.Marshal(yamlResults)
	if err != nil {
		return "", fmt.Errorf("yaml formatting failed: %w", err)
	}
	return string(data), nil
}

// tryMarshalYAML attempts to marshal a value to YAML and returns the error message if it fails.
// Returns empty string on success. Uses recover() because yaml.v3 panics on unserializable
// types (e.g., func, chan) instead of returning an error.
func tryMarshalYAML(v interface{}) (errMsg string) {
	defer func() {
		if r := recover(); r != nil {
			errMsg = fmt.Sprintf("%v", r)
		}
	}()
	if _, err := yaml.Marshal(v); err != nil {
		return err.Error()
	}
	return ""
}

// FileExtension returns "yaml" for YAML formatted output.
func (f *YAMLFormatter) FileExtension() string {
	return "yaml"
}
