package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCachePutGetMetadataAndImmutableFirstWriter(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if err := cache.PutWithInfo(key, []byte("first"), EntryInfo{ArtifactType: "go", Projection: "default"}); err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(key, []byte("second")); err != nil {
		t.Fatal(err)
	}
	data, metadata, err := cache.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first" || metadata.ArtifactType != "go" || metadata.Projection != "default" {
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
