package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInvalidateKeyIsVersionedAndRejectsStaleIdentity(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	v1 := makeTestKey(t, "v1", "billing")
	v2 := makeTestKey(t, "v2", "billing")
	if err := cache.Put(v1, []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(v2, []byte("v2")); err != nil {
		t.Fatal(err)
	}
	removed, err := cache.InvalidateKey(v2)
	if err != nil || !removed {
		t.Fatalf("v2 invalidation = %v, %v", removed, err)
	}
	if _, _, err := cache.Get(v1); err != nil {
		t.Fatalf("v1 was invalidated with v2: %v", err)
	}
	if _, _, err := cache.Get(v2); !errors.Is(err, ErrNotFound) {
		t.Fatalf("v2 read after invalidation = %v, want ErrNotFound", err)
	}
	stale := v1
	stale.Version = "v0"
	if _, err := cache.InvalidateKey(stale); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("stale version invalidation = %v, want ErrInvalidKey", err)
	}
	if _, _, err := cache.Get(v1); err != nil {
		t.Fatalf("stale key changed v1: %v", err)
	}
	if removed, err := cache.InvalidateKey(v2); err != nil || removed {
		t.Fatalf("repeated v2 invalidation = %v, %v", removed, err)
	}
}

func TestVersionedCacheRejectsCorruptAndFailedWrites(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
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
		t.Fatalf("corrupt entry read = %v, want ErrCorrupt", err)
	}
	if removed, err := cache.InvalidateKey(key); err != nil || !removed {
		t.Fatalf("corrupt entry invalidation = %v, %v", removed, err)
	}
	limited, err := Open(t.TempDir(), Options{MaxEntrySize: 3})
	if err != nil {
		t.Fatal(err)
	}
	failedKey := makeTestKey(t, "v2", "billing")
	if err := limited.Put(failedKey, []byte("too-large")); !errors.Is(err, ErrEntryTooLarge) {
		t.Fatalf("failed write = %v, want ErrEntryTooLarge", err)
	}
	failedObject, err := limited.objectPath(failedKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(failedObject); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed write left partial entry: %v", err)
	}
	if _, _, err := limited.Get(failedKey); !errors.Is(err, ErrNotFound) {
		t.Fatalf("failed write read = %v, want ErrNotFound", err)
	}
}

func TestInvalidateSkipsMetadataFromAnotherObject(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	corruptKey := makeTestKey(t, "v1", "billing")
	validKey := makeTestKey(t, "v2", "billing")
	for _, key := range []Key{corruptKey, validKey} {
		if err := cache.Put(key, []byte(key.String())); err != nil {
			t.Fatal(err)
		}
	}
	copyMetadata(t, cache, validKey, corruptKey)
	removed, err := cache.Invalidate(InvalidationFilter{Namespace: "billing"})
	if err != nil || removed != 1 {
		t.Fatalf("bound invalidation = %d, %v; want one valid removal", removed, err)
	}
	assertCorruptObjectPreserved(t, cache, corruptKey)
	if _, _, err := cache.Get(validKey); !errors.Is(err, ErrNotFound) {
		t.Fatalf("valid object after invalidation = %v, want ErrNotFound", err)
	}
}

func TestInvalidateStaleSkipsMetadataFromAnotherObject(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	corruptKey, _ := makeFreshKey(t, 1, 1)
	validKey, _ := makeFreshKey(t, 2, 1)
	_, current := makeFreshKey(t, 3, 1)
	for _, key := range []Key{corruptKey, validKey} {
		if err := cache.Put(key, []byte(key.String())); err != nil {
			t.Fatal(err)
		}
	}
	copyMetadata(t, cache, validKey, corruptKey)
	removed, err := cache.InvalidateStale(StaleFilter{
		Namespace: "billing", HostStage: GoHostedStage, Current: current,
	})
	if err != nil || removed != 1 {
		t.Fatalf("bound stale invalidation = %d, %v; want one valid removal", removed, err)
	}
	assertCorruptObjectPreserved(t, cache, corruptKey)
	if _, _, err := cache.Get(validKey); !errors.Is(err, ErrNotFound) {
		t.Fatalf("valid stale object = %v, want ErrNotFound", err)
	}
}

func copyMetadata(t *testing.T, cache *Cache, source, destination Key) {
	t.Helper()
	sourcePath, err := cache.objectPath(source)
	if err != nil {
		t.Fatal(err)
	}
	destinationPath, err := cache.objectPath(destination)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(sourcePath, metaFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destinationPath, metaFileName), raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertCorruptObjectPreserved(t *testing.T, cache *Cache, key Key) {
	t.Helper()
	object, err := cache.objectPath(key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(object); err != nil {
		t.Fatalf("metadata-mismatched object was removed: %v", err)
	}
	if _, _, err := cache.Get(key); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("metadata-mismatched object read = %v, want ErrCorrupt", err)
	}
}
