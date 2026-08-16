package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
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

func measureWorkerLatency(t *testing.T, workers int, prepopulate bool) []cacheLatencySample {
	t.Helper()
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if prepopulate {
		if err := cache.Put(key, []byte(cacheLatencyValue)); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	results := make(chan cacheLatencySample, workers)
	for index := 0; index < workers; index++ {
		go func() {
			<-start
			started := time.Now()
			data, _, hit, err := cache.GetOrCompute(context.Background(), key, func() ([]byte, error) {
				time.Sleep(2 * time.Millisecond)
				return []byte(cacheLatencyValue), nil
			})
			sample := cacheLatencySample{Hit: hit, DurationNanoseconds: time.Since(started).Nanoseconds(), Data: string(data)}
			if err != nil {
				sample.Error = err.Error()
			}
			results <- sample
		}()
	}
	close(start)
	samples := make([]cacheLatencySample, 0, workers)
	for index := 0; index < workers; index++ {
		samples = append(samples, <-results)
	}
	return samples
}

func measureProcessLatency(t *testing.T, processes int, prepopulate bool) []cacheLatencySample {
	t.Helper()
	root := t.TempDir()
	cache, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if prepopulate {
		if err := cache.Put(key, []byte(cacheLatencyValue)); err != nil {
			t.Fatal(err)
		}
	}
	release := filepath.Join(root, "release")
	commands := make([]*exec.Cmd, 0, processes)
	outputs := make([]*strings.Builder, 0, processes)
	readyFiles := make([]string, 0, processes)
	for index := 0; index < processes; index++ {
		ready := filepath.Join(root, fmt.Sprintf("ready-%d", index))
		result := filepath.Join(root, fmt.Sprintf("result-%d", index))
		output := &strings.Builder{}
		command := exec.Command(os.Args[0], "-test.run", "^TestCacheLatencyEvidenceMatrix$", "-cache-test-class="+CacheTestClassSlowObservation)
		command.Env = append(os.Environ(),
			cacheLatencyHelperEnv+"=1", cacheLatencyRootEnv+"="+root,
			cacheLatencyReleaseEnv+"="+release, cacheLatencyReadyEnv+"="+ready,
			cacheLatencyResultEnv+"="+result)
		command.Stdout, command.Stderr = output, output
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		commands = append(commands, command)
		outputs = append(outputs, output)
		readyFiles = append(readyFiles, ready)
	}
	if err := waitForFiles(readyFiles); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(release, []byte("go"), 0o600); err != nil {
		t.Fatal(err)
	}
	samples := make([]cacheLatencySample, 0, processes)
	for index, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatalf("latency helper %d failed: %v\n%s", index, err, outputs[index].String())
		}
		raw, err := os.ReadFile(filepath.Join(root, fmt.Sprintf("result-%d", index)))
		if err != nil {
			t.Fatal(err)
		}
		var sample cacheLatencySample
		if err := json.Unmarshal(raw, &sample); err != nil {
			t.Fatal(err)
		}
		samples = append(samples, sample)
	}
	if _, _, err := cache.Get(key); err != nil {
		t.Fatalf("final process-matrix object is not readable: %v", err)
	}
	return samples
}

func runCacheLatencyHelper(t *testing.T) {
	t.Helper()
	cache, err := Open(os.Getenv(cacheLatencyRootEnv))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv(cacheLatencyReadyEnv), []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := waitForFiles([]string{os.Getenv(cacheLatencyReleaseEnv)}); err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	started := time.Now()
	data, _, hit, err := cache.GetOrCompute(context.Background(), key, func() ([]byte, error) {
		time.Sleep(2 * time.Millisecond)
		return []byte(cacheLatencyValue), nil
	})
	sample := cacheLatencySample{Hit: hit, DurationNanoseconds: time.Since(started).Nanoseconds(), Data: string(data)}
	if err != nil {
		sample.Error = err.Error()
	}
	raw, err := json.Marshal(sample)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv(cacheLatencyResultEnv), raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertLatencySamples(t *testing.T, samples []cacheLatencySample, wantHit bool, wantCount int) {
	t.Helper()
	if len(samples) != wantCount {
		t.Fatalf("latency samples=%d, want %d", len(samples), wantCount)
	}
	misses := 0
	for index, sample := range samples {
		if sample.Error != "" || sample.Data != cacheLatencyValue || sample.DurationNanoseconds <= 0 {
			t.Fatalf("sample[%d]=%+v, want valid result", index, sample)
		}
		if sample.Hit != wantHit {
			if !wantHit && sample.Hit {
				continue
			}
			t.Fatalf("sample[%d] hit=%v, want %v", index, sample.Hit, wantHit)
		}
		if !sample.Hit {
			misses++
		}
	}
	if !wantHit && misses == 0 {
		t.Fatal("cold process/worker matrix produced no miss evidence")
	}
}

func durationsForLatencySamples(samples []cacheLatencySample, hit bool) []time.Duration {
	result := make([]time.Duration, 0, len(samples))
	for _, sample := range samples {
		if sample.Hit == hit {
			result = append(result, time.Duration(sample.DurationNanoseconds))
		}
	}
	return result
}

func latencyPercentile(samples []time.Duration, percentile int) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	rank := (len(sorted)*percentile + 99) / 100
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}
