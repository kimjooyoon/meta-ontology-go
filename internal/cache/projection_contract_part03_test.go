package cache

import (
	"errors"
	"testing"
)

func TestProjectionKeyC3UnknownIdentityFailsClosed(t *testing.T) {
	spec := projectionSpec()
	spec.DependencyRoot = ""
	if _, err := NewProjectionKey(spec); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("missing dependency root = %v, want ErrUnknownFreshness", err)
	}
	var unknown map[string]string
	if _, err := NewFreshness(FreshnessSpec{
		Dependencies: unknown, Provenance: map[string]any{"known": true},
	}); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("nil dependency input = %v, want ErrUnknownFreshness", err)
	}
	if _, err := NewFreshness(FreshnessSpec{
		Dependencies: map[string]any{"known": true}, Provenance: nil,
	}); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("nil provenance input = %v, want ErrUnknownFreshness", err)
	}
	if _, err := NewProjectionKey(projectionSpec()); err != nil {
		t.Fatal(err)
	}
}
func TestProjectionKeyC3OpaqueOptionsFailClosed(t *testing.T) {
	spec := projectionSpec()
	spec.Options = map[string]any{"mode": "unsafe"}
	if _, err := NewProjectionKey(spec); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("opaque options = %v, want ErrUnknownFreshness", err)
	}
	if _, err := DigestOptions(nil); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("nil options = %v, want ErrUnknownFreshness", err)
	}
	first, err := DigestOptions(map[string]any{"mode": "fast", "trim": true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := DigestOptions(map[string]any{"trim": true, "mode": "fast"})
	if err != nil || first != second {
		t.Fatalf("options presentation changed digest: %s != %s (%v)", first, second, err)
	}
}
func TestProjectionKeyC3EntryInfoCannotAliasIdentity(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.PutWithInfo(key, []byte("projection"), EntryInfo{Projection: "other"}); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("metadata-only projection alias = %v, want ErrInvalidKey", err)
	}
	variant := projectionSpec()
	variant.ArtifactKind = "docs"
	variant.Projection = "ir"
	variantKey, err := NewProjectionKey(variant)
	if err != nil {
		t.Fatal(err)
	}
	if variantKey == key {
		t.Fatal("artifact/projection mutation retained identity")
	}
}
