package cache

import (
	"errors"
	"testing"
)

func TestPutRejectsTamperedFullKey(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	key.Digest = HashBytes([]byte("tampered"))
	if err := cache.Put(key, []byte("projection")); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("tampered key error = %v, want ErrInvalidKey", err)
	}
}
