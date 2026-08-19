package cache

import (
	"errors"
	"sync"
	"testing"
)

func TestCacheInvalidateKeyConcurrentIsSingleWriter(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if err := cache.Put(key, []byte("stale")); err != nil {
		t.Fatal(err)
	}
	const workers = 16
	start := make(chan struct{})
	type result struct {
		removed bool
		err     error
	}
	results := make(chan result, workers)
	var wait sync.WaitGroup
	wait.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wait.Done()
			<-start
			removed, err := cache.InvalidateKey(key)
			results <- result{removed: removed, err: err}
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	removed := 0
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent invalidation error = %v", result.err)
		}
		if result.removed {
			removed++
		}
	}
	if removed != 1 {
		t.Fatalf("successful invalidations = %d, want 1", removed)
	}
	if removedAgain, err := cache.InvalidateKey(key); err != nil || removedAgain {
		t.Fatalf("repeated invalidation = %v, %v; want false, nil", removedAgain, err)
	}
	if _, _, err := cache.Get(key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("invalidated entry read error = %v, want ErrNotFound", err)
	}
}
