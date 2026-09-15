package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectionKeyC2PresentationStableAndCorruptionMisses(t *testing.T) {
	first := projectionSpec()
	second := projectionSpec()
	first.BuildTags = []string{"linux", "windows"}
	second.BuildTags = []string{"windows", "linux", "linux"}
	second.OptionsDigest = mustOptionsDigest(map[string]any{"trim": true, "mode": "fast"})
	firstKey, err := NewProjectionKey(first)
	if err != nil {
		t.Fatal(err)
	}
	secondKey, err := NewProjectionKey(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstKey != secondKey {
		t.Fatal("presentation-only changes altered projection key")
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(firstKey, []byte("small")); err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(secondKey, []byte("larger-content")); err != nil {
		t.Fatal(err)
	}
	data, metadata, err := cache.Get(firstKey)
	if err != nil || string(data) != "small" || metadata.Size != int64(len(data)) {
		t.Fatalf("immutable false-hit result = %q, metadata=%+v, err=%v", data, metadata, err)
	}
	object, err := cache.objectPath(firstKey)
	if err != nil {
		t.Fatal(err)
	}
	metadata.Size++
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(object, metaFileName), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cache.Get(firstKey); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("content-size mismatch = %v, want ErrCorrupt", err)
	}
}
