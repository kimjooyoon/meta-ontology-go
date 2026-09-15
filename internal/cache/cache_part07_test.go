package cache

import (
	"context"
	"errors"
	"testing"
)

func TestCacheContextAndSizeErrors(t *testing.T) {
	cache, err := Open(t.TempDir(), Options{MaxEntrySize: 3})
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if err := cache.Put(key, []byte("four")); !errors.Is(err, ErrEntryTooLarge) {
		t.Fatalf("large write error = %v, want ErrEntryTooLarge", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, _, err := cache.GetOrCompute(ctx, key, func() ([]byte, error) {
		t.Fatal("cancelled compute was called")
		return nil, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled compute error = %v", err)
	}
	if _, _, _, err := cache.GetOrCompute(context.Background(), key, nil); err == nil {
		t.Fatal("nil compute function was accepted")
	}
}
