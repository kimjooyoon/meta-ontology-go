package cache

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

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
