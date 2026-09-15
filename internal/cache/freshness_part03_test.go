package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestIncompleteObjectIsRejectedAndRepairedAtomically(t *testing.T) {
	root := t.TempDir()
	cache, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	object, err := cache.objectPath(key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(object), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(object, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(object, dataFileName), []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cache.Get(key); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("incomplete object error = %v, want ErrCorrupt", err)
	}
	if err := cache.Put(key, []byte("recovered")); err != nil {
		t.Fatal(err)
	}
	data, _, err := cache.Get(key)
	if err != nil || string(data) != "recovered" {
		t.Fatalf("recovered object = %q, %v", data, err)
	}
}
