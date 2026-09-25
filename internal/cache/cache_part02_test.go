package cache

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestCacheMissCorruptionAndRecovery(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if _, _, err := cache.Get(key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("miss error = %v, want ErrNotFound", err)
	}
	if ok, err := cache.Has(key); err != nil || ok {
		t.Fatalf("Has on miss = %v, %v", ok, err)
	}
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
		t.Fatalf("corrupt read error = %v, want ErrCorrupt", err)
	}
	var computes atomic.Int32
	data, _, hit, err := cache.GetOrCompute(context.Background(), key, func() ([]byte, error) {
		computes.Add(1)
		return []byte("rebuilt"), nil
	})
	if err != nil || hit || string(data) != "rebuilt" || computes.Load() != 1 {
		t.Fatalf("recovery = %q, hit=%v, computes=%d, err=%v", data, hit, computes.Load(), err)
	}
	if _, _, hit, err := cache.GetOrCompute(context.Background(), key, func() ([]byte, error) {
		computes.Add(1)
		return []byte("unexpected"), nil
	}); err != nil || !hit || computes.Load() != 1 {
		t.Fatalf("recovered hit = %v, computes=%d, err=%v", hit, computes.Load(), err)
	}
}
