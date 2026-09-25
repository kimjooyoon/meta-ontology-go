package cache

import (
	"errors"
	"testing"
)

func TestCacheReceiptBindsProjectionAndContent(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	receipt := cacheReceiptFixture(key, "binding")
	if err := receipt.ValidateForKey(key); err != nil {
		t.Fatal(err)
	}
	if err := receipt.ValidateForData([]byte("projection")); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*CacheReceipt){
		"domain":     func(r *CacheReceipt) { r.Domain = "other" },
		"version":    func(r *CacheReceipt) { r.KeyVersion = "other" },
		"host stage": func(r *CacheReceipt) { r.HostStage = GoooHostedStage },
		"projection": func(r *CacheReceipt) { r.Projection = "other" },
		"artifact":   func(r *CacheReceipt) { r.ArtifactKind = "other" },
		"options":    func(r *CacheReceipt) { r.OptionsDigest = mustOptionsDigest(map[string]any{"mode": "other"}) },
		"content":    func(r *CacheReceipt) { r.ContentDigest = HashBytes([]byte("other")) },
		"size":       func(r *CacheReceipt) { r.Size++ },
		"rebuild":    func(r *CacheReceipt) { r.Reconstructable = false },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := receipt
			mutate(&mutated)
			if name == "content" || name == "size" {
				if err := mutated.ValidateForData([]byte("projection")); !errors.Is(err, ErrInvalidReceipt) {
					t.Fatalf("mutated data receipt = %v, want ErrInvalidReceipt", err)
				}
				return
			}
			if err := mutated.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("mutated identity receipt = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}
func TestCacheReceiptC1C4OptionsDigestFailsClosed(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	receipt := cacheReceiptFixture(key, "options")
	missing := receipt
	missing.OptionsDigest = ""
	if err := missing.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("missing options digest = %v, want ErrInvalidReceipt", err)
	}
	variant := projectionSpec()
	variant.OptionsDigest = mustOptionsDigest(map[string]any{"mode": "other"})
	variantKey, err := NewProjectionKey(variant)
	if err != nil {
		t.Fatal(err)
	}
	if variantKey == key {
		t.Fatal("options mutation retained projection identity")
	}
	receipt.OptionsDigest = variantKey.OptionsDigest
	if err := receipt.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("mismatched options digest = %v, want ErrInvalidReceipt", err)
	}
}
