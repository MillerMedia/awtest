package types

import "time"

// ScanSummary holds aggregate statistics about a completed scan.
type ScanSummary struct {
	TotalServices        int
	AccessibleServices   int
	AccessDeniedServices int
	TotalResources       int
	ScanDuration         time.Duration
	Timestamp            time.Time
}

// GenerateSummary computes aggregate scan statistics from results.
// A service can appear in both accessible and denied counts if it has
// mixed success/error results (e.g., S3 ListBuckets succeeds but ListObjects fails).
func GenerateSummary(results []ScanResult, startTime time.Time) ScanSummary {
	serviceMap := make(map[string]bool)
	accessibleMap := make(map[string]bool)
	deniedMap := make(map[string]bool)
	resourceCount := 0

	for _, r := range results {
		serviceMap[r.ServiceName] = true
		if r.Error != nil {
			deniedMap[r.ServiceName] = true
		} else {
			accessibleMap[r.ServiceName] = true
			resourceCount++
		}
	}

	return ScanSummary{
		TotalServices:        len(serviceMap),
		AccessibleServices:   len(accessibleMap),
		AccessDeniedServices: len(deniedMap),
		TotalResources:       resourceCount,
		ScanDuration:         time.Since(startTime).Truncate(time.Millisecond),
		Timestamp:            startTime,
	}
}
