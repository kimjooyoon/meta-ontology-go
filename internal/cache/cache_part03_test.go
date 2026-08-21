package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCacheDetectsMetadataTampering(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if err := cache.PutWithInfo(key, []byte("valid"), EntryInfo{Projection: "default"}); err != nil {
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
