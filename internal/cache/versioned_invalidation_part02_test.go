package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

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
