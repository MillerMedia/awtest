package types

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateSummary_MixedResults(t *testing.T) {
	startTime := time.Now().Add(-2 * time.Second)
	results := []ScanResult{
		{ServiceName: "S3", MethodName: "s3:ListBuckets", ResourceName: "bucket-1"},
		{ServiceName: "S3", MethodName: "s3:ListBuckets", ResourceName: "bucket-2"},
		{ServiceName: "EC2", Error: errors.New("access denied")},
		{ServiceName: "IAM", MethodName: "iam:ListUsers", ResourceName: "user-1"},
	}

	summary := GenerateSummary(results, startTime)

	if summary.TotalServices != 3 {
		t.Errorf("TotalServices = %d, want 3", summary.TotalServices)
	}
	if summary.AccessibleServices != 2 {
		t.Errorf("AccessibleServices = %d, want 2", summary.AccessibleServices)
	}
	if summary.AccessDeniedServices != 1 {
		t.Errorf("AccessDeniedServices = %d, want 1", summary.AccessDeniedServices)
	}
	if summary.TotalResources != 3 {
		t.Errorf("TotalResources = %d, want 3", summary.TotalResources)
	}
}

func TestGenerateSummary_EmptyResults(t *testing.T) {
	startTime := time.Now()
	summary := GenerateSummary([]ScanResult{}, startTime)

	if summary.TotalServices != 0 {
		t.Errorf("TotalServices = %d, want 0", summary.TotalServices)
	}
	if summary.AccessibleServices != 0 {
		t.Errorf("AccessibleServices = %d, want 0", summary.AccessibleServices)
	}
	if summary.AccessDeniedServices != 0 {
		t.Errorf("AccessDeniedServices = %d, want 0", summary.AccessDeniedServices)
	}
	if summary.TotalResources != 0 {
		t.Errorf("TotalResources = %d, want 0", summary.TotalResources)
	}
}

func TestGenerateSummary_AllErrors(t *testing.T) {
	startTime := time.Now()
	results := []ScanResult{
		{ServiceName: "S3", Error: errors.New("denied")},
		{ServiceName: "EC2", Error: errors.New("denied")},
		{ServiceName: "IAM", Error: errors.New("denied")},
	}

	summary := GenerateSummary(results, startTime)

	if summary.AccessibleServices != 0 {
		t.Errorf("AccessibleServices = %d, want 0", summary.AccessibleServices)
	}
	if summary.AccessDeniedServices != 3 {
		t.Errorf("AccessDeniedServices = %d, want 3", summary.AccessDeniedServices)
	}
	if summary.TotalResources != 0 {
		t.Errorf("TotalResources = %d, want 0", summary.TotalResources)
	}
}

func TestGenerateSummary_AllSuccesses(t *testing.T) {
	startTime := time.Now()
	results := []ScanResult{
		{ServiceName: "S3", ResourceName: "bucket-1"},
		{ServiceName: "EC2", ResourceName: "i-123"},
	}

	summary := GenerateSummary(results, startTime)

	if summary.AccessibleServices != 2 {
		t.Errorf("AccessibleServices = %d, want 2", summary.AccessibleServices)
	}
	if summary.AccessDeniedServices != 0 {
		t.Errorf("AccessDeniedServices = %d, want 0", summary.AccessDeniedServices)
	}
	if summary.TotalResources != 2 {
		t.Errorf("TotalResources = %d, want 2", summary.TotalResources)
	}
}

func TestGenerateSummary_Duration(t *testing.T) {
	startTime := time.Now().Add(-100 * time.Millisecond)
	summary := GenerateSummary([]ScanResult{}, startTime)

	if summary.ScanDuration <= 0 {
		t.Errorf("ScanDuration = %v, want > 0", summary.ScanDuration)
	}
}

func TestGenerateSummary_UniqueServiceCounting(t *testing.T) {
	startTime := time.Now()
	results := []ScanResult{
		{ServiceName: "S3", ResourceName: "bucket-1"},
		{ServiceName: "S3", ResourceName: "bucket-2"},
		{ServiceName: "S3", ResourceName: "bucket-3"},
		{ServiceName: "EC2", ResourceName: "i-123"},
	}

	summary := GenerateSummary(results, startTime)

	if summary.TotalServices != 2 {
		t.Errorf("TotalServices = %d, want 2 (unique services)", summary.TotalServices)
	}
	if summary.TotalResources != 4 {
		t.Errorf("TotalResources = %d, want 4", summary.TotalResources)
	}
}

func TestGenerateSummary_ServiceInBothAccessibleAndDenied(t *testing.T) {
	startTime := time.Now()
	results := []ScanResult{
		{ServiceName: "S3", ResourceName: "bucket-1"},
		{ServiceName: "S3", Error: errors.New("ListObjects denied")},
	}

	summary := GenerateSummary(results, startTime)

	if summary.TotalServices != 1 {
		t.Errorf("TotalServices = %d, want 1", summary.TotalServices)
	}
	if summary.AccessibleServices != 1 {
		t.Errorf("AccessibleServices = %d, want 1", summary.AccessibleServices)
	}
	if summary.AccessDeniedServices != 1 {
		t.Errorf("AccessDeniedServices = %d, want 1", summary.AccessDeniedServices)
	}
}

func TestGenerateSummary_Timestamp(t *testing.T) {
	startTime := time.Date(2026, 3, 3, 10, 30, 0, 0, time.UTC)
	summary := GenerateSummary([]ScanResult{}, startTime)

	if !summary.Timestamp.Equal(startTime) {
		t.Errorf("Timestamp = %v, want %v", summary.Timestamp, startTime)
	}
}
