package formatters

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MillerMedia/awtest/cmd/awtest/types"
)

func TestJSONFormatter_Format(t *testing.T) {
	formatter := NewJSONFormatter()
	fixedTime := time.Date(2026, 3, 2, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		results   []types.ScanResult
		wantErr   bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name: "single valid result",
			results: []types.ScanResult{
				{
					ServiceName:  "S3",
					MethodName:   "s3:ListBuckets",
					ResourceType: "bucket",
					ResourceName: "test-bucket",
					Details:      map[string]interface{}{"region": "us-east-1"},
					Timestamp:    fixedTime,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				var parsed []jsonScanResult
				if err := json.Unmarshal([]byte(output), &parsed); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if len(parsed) != 1 {
					t.Fatalf("expected 1 result, got %d", len(parsed))
				}
				if parsed[0].ServiceName != "S3" {
					t.Errorf("expected S3, got %s", parsed[0].ServiceName)
				}
				if !strings.Contains(output, "serviceName") {
					t.Error("missing camelCase serviceName")
				}
				if !strings.Contains(output, "methodName") {
					t.Error("missing camelCase methodName")
				}
			},
		},
		{
			name:    "empty results returns empty array",
			results: []types.ScanResult{},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				if strings.TrimSpace(output) != "[]" {
					t.Errorf("expected [], got %s", output)
				}
			},
		},
		{
			name: "result with error",
			results: []types.ScanResult{
				{
					ServiceName: "RDS",
					MethodName:  "rds:DescribeDBInstances",
					Error:       errors.New("AccessDeniedException: not authorized"),
					Timestamp:   fixedTime,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				var parsed []jsonScanResult
				if err := json.Unmarshal([]byte(output), &parsed); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if parsed[0].Error != "AccessDeniedException: not authorized" {
					t.Errorf("error not serialized correctly: %s", parsed[0].Error)
				}
			},
		},
		{
			name: "timestamp in RFC3339 format",
			results: []types.ScanResult{
				{ServiceName: "EC2", Timestamp: fixedTime},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "2026-03-02T14:30:00Z") {
					t.Error("timestamp not in RFC3339 format")
				}
			},
		},
		{
			name: "multiple results",
			results: []types.ScanResult{
				{ServiceName: "S3", MethodName: "s3:ListBuckets", ResourceName: "bucket-1", Timestamp: fixedTime},
				{ServiceName: "EC2", MethodName: "ec2:DescribeInstances", ResourceName: "i-123", Timestamp: fixedTime},
				{ServiceName: "IAM", MethodName: "iam:ListUsers", ResourceName: "admin", Timestamp: fixedTime},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				var parsed []jsonScanResult
				if err := json.Unmarshal([]byte(output), &parsed); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if len(parsed) != 3 {
					t.Fatalf("expected 3 results, got %d", len(parsed))
				}
			},
		},
		{
			name: "nil details serialized correctly",
			results: []types.ScanResult{
				{ServiceName: "Lambda", Timestamp: fixedTime},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				var parsed []jsonScanResult
				if err := json.Unmarshal([]byte(output), &parsed); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if parsed[0].Details != nil {
					t.Error("nil details should remain nil")
				}
			},
		},
		{
			name: "successful result has no error field in JSON",
			results: []types.ScanResult{
				{
					ServiceName:  "S3",
					MethodName:   "s3:ListBuckets",
					ResourceType: "bucket",
					ResourceName: "test-bucket",
					Timestamp:    fixedTime,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				if strings.Contains(output, `"error"`) {
					t.Error("successful result should not contain error field (omitempty)")
				}
			},
		},
		{
			name: "camelCase field naming verified",
			results: []types.ScanResult{
				{
					ServiceName:  "S3",
					MethodName:   "s3:ListBuckets",
					ResourceType: "bucket",
					ResourceName: "test-bucket",
					Details:      map[string]interface{}{"key": "value"},
					Timestamp:    fixedTime,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				expectedFields := []string{"serviceName", "methodName", "resourceType", "resourceName", "details", "timestamp"}
				for _, field := range expectedFields {
					if !strings.Contains(output, `"`+field+`"`) {
						t.Errorf("missing camelCase field: %s", field)
					}
				}
				// Ensure PascalCase is NOT present
				unexpectedFields := []string{"ServiceName", "MethodName", "ResourceType", "ResourceName", "Timestamp"}
				for _, field := range unexpectedFields {
					if strings.Contains(output, `"`+field+`"`) {
						t.Errorf("found PascalCase field that should be camelCase: %s", field)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := formatter.Format(tt.results)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}

func TestJSONFormatter_FileExtension(t *testing.T) {
	formatter := NewJSONFormatter()
	if ext := formatter.FileExtension(); ext != "json" {
		t.Errorf("FileExtension() = %s, want json", ext)
	}
}

func TestNewJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()
	if formatter == nil {
		t.Error("NewJSONFormatter() returned nil")
	}
}

// Test that JSONFormatter satisfies OutputFormatter interface
func TestJSONFormatter_ImplementsInterface(t *testing.T) {
	var _ OutputFormatter = (*JSONFormatter)(nil)
}
