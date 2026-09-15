package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

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
