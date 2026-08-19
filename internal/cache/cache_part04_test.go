package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheSameKeyComputesOnceConcurrently(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	const workers = 16
	start := make(chan struct{})
	type result struct {
		hit  bool
		err  error
		data string
	}
	results := make(chan result, workers)
	var computes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wait.Done()
			<-start
			data, _, hit, err := cache.GetOrCompute(context.Background(), key, func() ([]byte, error) {
				computes.Add(1)
				time.Sleep(5 * time.Millisecond)
				return []byte("shared"), nil
			})
			results <- result{hit: hit, err: err, data: string(data)}
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	misses := 0
	for result := range results {
		if result.err != nil || result.data != "shared" {
			t.Fatalf("concurrent result = %+v", result)
		}
		if !result.hit {
			misses++
		}
	}
	if computes.Load() != 1 || misses != 1 {
		t.Fatalf("computes=%d misses=%d, want one each", computes.Load(), misses)
	}
}
