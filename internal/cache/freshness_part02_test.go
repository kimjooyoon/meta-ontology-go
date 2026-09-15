package cache

import (
	"errors"
	"testing"
)

func TestGetFreshRejectsStaleArtifact(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key, oldFreshness := makeFreshKey(t, 1, 1)
	if err := cache.Put(key, []byte("projection")); err != nil {
		t.Fatal(err)
	}
	if data, _, err := cache.GetFresh(key, oldFreshness); err != nil || string(data) != "projection" {
		t.Fatalf("fresh read = %q, %v", data, err)
	}
	_, currentFreshness := makeFreshKey(t, 2, 1)
	if _, _, err := cache.GetFresh(key, currentFreshness); !errors.Is(err, ErrStale) {
		t.Fatalf("stale read error = %v, want ErrStale", err)
	}
	newKey, _ := makeFreshKey(t, 2, 1)
	if key.String() == newKey.String() {
		t.Fatal("dependency freshness did not change key")
	}
}
func TestInvalidateStaleByFreshness(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	oldKey, _ := makeFreshKey(t, 1, 1)
	provenanceKey, _ := makeFreshKey(t, 1, 2)
	currentKey, current := makeFreshKey(t, 2, 1)
	for _, key := range []Key{oldKey, provenanceKey, currentKey} {
		if err := cache.Put(key, []byte(key.String())); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := cache.InvalidateStale(StaleFilter{
		Namespace: "billing", HostStage: GoHostedStage, Current: current,
	})
	if err != nil || removed != 2 {
		t.Fatalf("stale invalidation = %d, %v", removed, err)
	}
	if _, _, err := cache.Get(currentKey); err != nil {
		t.Fatalf("current entry was removed: %v", err)
	}
	for _, key := range []Key{oldKey, provenanceKey} {
		if _, _, err := cache.Get(key); !errors.Is(err, ErrNotFound) {
			t.Fatalf("stale entry read = %v, want ErrNotFound", err)
		}
	}
}
func TestInvalidateStaleRequiresScopeAndFreshness(t *testing.T) {
	_, current := makeFreshKey(t, 1, 1)
	if _, err := (&Cache{}).InvalidateStale(StaleFilter{Current: current}); !errors.Is(err, ErrEmptyFilter) {
		t.Fatalf("empty stale filter error = %v", err)
	}
	if _, err := (&Cache{}).InvalidateStale(StaleFilter{Namespace: "billing"}); !errors.Is(err, ErrInvalidFreshness) {
		t.Fatalf("missing freshness error = %v", err)
	}
	if _, err := (&Cache{}).InvalidateStale(StaleFilter{
		Namespace: "billing", HostStage: HostStage("unknown"), Current: current,
	}); !errors.Is(err, ErrInvalidHostStage) {
		t.Fatalf("unknown stage error = %v", err)
	}
}
