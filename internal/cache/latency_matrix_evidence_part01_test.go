package cache

import (
	"fmt"
	"os"
	"testing"
)

const (
	cacheLatencyHelperEnv  = "GOOO_CACHE_LATENCY_HELPER"
	cacheLatencyRootEnv    = "GOOO_CACHE_LATENCY_ROOT"
	cacheLatencyReleaseEnv = "GOOO_CACHE_LATENCY_RELEASE"
	cacheLatencyReadyEnv   = "GOOO_CACHE_LATENCY_READY"
	cacheLatencyResultEnv  = "GOOO_CACHE_LATENCY_RESULT"
	cacheLatencyValue      = "latency-matrix-projection"
)

var cacheLatencyMatrix = []int{1, 2, 4, 8, 16, 32}

type cacheLatencySample struct {
	Hit                 bool
	DurationNanoseconds int64
	Data                string
	Error               string
}

// TestCacheLatencyEvidenceMatrix records p50/p95 hit and miss evidence for
// each supported worker and process cardinality. These values are evidence
// only: this test deliberately has no timing threshold or performance claim.
func TestCacheLatencyEvidenceMatrix(t *testing.T) {
	requireClassifiedCacheTest(t, "TestCacheLatencyEvidenceMatrix", CacheTestClassSlowObservation)
	if os.Getenv(cacheLatencyHelperEnv) == "1" {
		runCacheLatencyHelper(t)
		return
	}
	for _, cardinality := range cacheLatencyMatrix {
		cardinality := cardinality
		t.Run(fmt.Sprintf("workers/%d", cardinality), func(t *testing.T) {
			recordLatencyMatrix(t, "workers", cardinality,
				measureWorkerLatency(t, cardinality, true),
				measureWorkerLatency(t, cardinality, false))
		})
		t.Run(fmt.Sprintf("processes/%d", cardinality), func(t *testing.T) {
			recordLatencyMatrix(t, "processes", cardinality,
				measureProcessLatency(t, cardinality, true),
				measureProcessLatency(t, cardinality, false))
		})
	}
}
func recordLatencyMatrix(t *testing.T, dimension string, cardinality int,
	hitSamples, missSamples []cacheLatencySample) {
	t.Helper()
	assertLatencySamples(t, hitSamples, true, cardinality)
	assertLatencySamples(t, missSamples, false, cardinality)
	hitDurations := durationsForLatencySamples(hitSamples, true)
	missDurations := durationsForLatencySamples(missSamples, false)
	t.Logf("dimension=%s cardinality=%d hit_count=%d miss_count=%d p50_hit=%s p95_hit=%s p50_miss=%s p95_miss=%s; timing is evidence only",
		dimension, cardinality, len(hitDurations), len(missDurations),
		latencyPercentile(hitDurations, 50), latencyPercentile(hitDurations, 95),
		latencyPercentile(missDurations, 50), latencyPercentile(missDurations, 95))
}
