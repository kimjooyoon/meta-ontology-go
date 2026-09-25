package cache

import (
	"errors"
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
