package cache

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCachePutGetMetadataAndImmutableFirstWriter(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if err := cache.PutWithInfo(key, []byte("first"), EntryInfo{ArtifactType: "go", Projection: "source"}); err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(key, []byte("second")); err != nil {
		t.Fatal(err)
	}
	data, metadata, err := cache.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first" || metadata.ArtifactType != "go" || metadata.Projection != "source" {
		t.Fatalf("unexpected cache entry: %q %+v", data, metadata)
	}
	if ok, err := cache.Has(key); err != nil || !ok {
		t.Fatalf("Has = %v, %v", ok, err)
	}
	if metadataOnly, err := cache.GetMetadata(key); err != nil || metadataOnly.Key != key.String() {
		t.Fatalf("GetMetadata = %+v, %v", metadataOnly, err)
	}
	object, err := cache.objectPath(key)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{dataFileName, metaFileName} {
		if _, err := os.Stat(filepath.Join(object, name)); err != nil {
			t.Fatalf("committed object missing %s: %v", name, err)
		}
	}
}

func TestCacheMissCorruptionAndRecovery(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if _, _, err := cache.Get(key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("miss error = %v, want ErrNotFound", err)
	}
	if ok, err := cache.Has(key); err != nil || ok {
		t.Fatalf("Has on miss = %v, %v", ok, err)
	}
	if err := cache.Put(key, []byte("valid")); err != nil {
		t.Fatal(err)
	}
	object, err := cache.objectPath(key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(object, dataFileName), []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cache.Get(key); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("corrupt read error = %v, want ErrCorrupt", err)
	}
	var computes atomic.Int32
	data, _, hit, err := cache.GetOrCompute(context.Background(), key, func() ([]byte, error) {
		computes.Add(1)
		return []byte("rebuilt"), nil
	})
	if err != nil || hit || string(data) != "rebuilt" || computes.Load() != 1 {
		t.Fatalf("recovery = %q, hit=%v, computes=%d, err=%v", data, hit, computes.Load(), err)
	}
	if _, _, hit, err := cache.GetOrCompute(context.Background(), key, func() ([]byte, error) {
		computes.Add(1)
		return []byte("unexpected"), nil
	}); err != nil || !hit || computes.Load() != 1 {
		t.Fatalf("recovered hit = %v, computes=%d, err=%v", hit, computes.Load(), err)
	}
}

func TestCacheDetectsMetadataTampering(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if err := cache.PutWithInfo(key, []byte("valid"), EntryInfo{Projection: "go"}); err != nil {
		t.Fatal(err)
	}
	object, err := cache.objectPath(key)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := cache.GetMetadata(key)
	if err != nil {
		t.Fatal(err)
	}
	metadata.Projection = "tampered"
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(object, metaFileName), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cache.Get(key); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("tampered metadata error = %v, want ErrCorrupt", err)
	}
}

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

func TestCacheInvalidationClearAndTemporaryCleanup(t *testing.T) {
	root := t.TempDir()
	cache, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	billing := makeTestKey(t, "v1", "billing")
	other := makeTestKey(t, "v1", "other")
	versioned := makeTestKey(t, "v2", "billing")
	for _, item := range []struct {
		key  Key
		info EntryInfo
	}{{billing, EntryInfo{ArtifactType: "go"}}, {other, EntryInfo{}}, {versioned, EntryInfo{Projection: "docs"}}} {
		if err := cache.PutWithInfo(item.key, []byte(item.key.Namespace), item.info); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := cache.Invalidate(InvalidationFilter{Namespace: "billing"})
	if err != nil || removed != 2 {
		t.Fatalf("namespace invalidation = %d, %v", removed, err)
	}
	if _, _, err := cache.Get(other); err != nil {
		t.Fatalf("unmatched entry was removed: %v", err)
	}
	if _, err := cache.Invalidate(InvalidationFilter{}); !errors.Is(err, ErrEmptyFilter) {
		t.Fatalf("empty filter error = %v", err)
	}
	object, err := cache.objectPath(other)
	if err != nil {
		t.Fatal(err)
	}
	shard := filepath.Dir(object)
	temporary, err := os.MkdirTemp(shard, ".stale.tmp-")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(temporary); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale temporary entry remains: %v", err)
	}
	if err := cache.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cache.Get(other); !errors.Is(err, ErrNotFound) {
		t.Fatalf("clear read error = %v, want ErrNotFound", err)
	}
}

func TestCacheContextAndSizeErrors(t *testing.T) {
	cache, err := Open(t.TempDir(), Options{MaxEntrySize: 3})
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if err := cache.Put(key, []byte("four")); !errors.Is(err, ErrEntryTooLarge) {
		t.Fatalf("large write error = %v, want ErrEntryTooLarge", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, _, err := cache.GetOrCompute(ctx, key, func() ([]byte, error) {
		t.Fatal("cancelled compute was called")
		return nil, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled compute error = %v", err)
	}
	if _, _, _, err := cache.GetOrCompute(context.Background(), key, nil); err == nil {
		t.Fatal("nil compute function was accepted")
	}
}
