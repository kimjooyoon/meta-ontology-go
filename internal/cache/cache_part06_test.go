package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

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
	}{{billing, EntryInfo{ArtifactType: "go"}}, {other, EntryInfo{}}, {versioned, EntryInfo{Projection: "default"}}} {
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
