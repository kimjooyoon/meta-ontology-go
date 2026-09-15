package cache

import (
	"errors"
	"testing"
)

func TestPartialFreshnessSeparatesProvenance(t *testing.T) {
	first := partialSpec("body", 1, 1)
	second := partialSpec("body", 1, 2)
	firstKey, err := NewPartialKey(first)
	if err != nil {
		t.Fatal(err)
	}
	secondKey, err := NewPartialKey(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstKey == secondKey {
		t.Fatal("provenance change did not change partial key")
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(firstKey, []byte("old")); err != nil {
		t.Fatal(err)
	}
	current, err := NewFreshness(second.KeySpec.Freshness)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := cache.GetFresh(firstKey, current); !errors.Is(err, ErrStale) {
		t.Fatalf("stale partial read = %v, want ErrStale", err)
	}
}
func TestNewPartialKeyRejectsEmptyPart(t *testing.T) {
	if _, err := NewPartialKey(PartialSpec{KeySpec: KeySpec{Namespace: "billing"}}); err == nil {
		t.Fatal("empty partial name was accepted")
	}
}
