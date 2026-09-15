package cache

import (
	"errors"
	"testing"
)

func assertReplay(t *testing.T, cache *Cache, receipt CacheReceipt, label string) {
	t.Helper()
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("%s replay = %v, want ErrReceiptReplay", label, err)
	}
}
