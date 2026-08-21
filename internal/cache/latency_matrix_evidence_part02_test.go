package cache

import (
	"context"
	"testing"
	"time"
)

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
	for range workers {
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
	for range workers {
		samples = append(samples, <-results)
	}
	return samples
}
