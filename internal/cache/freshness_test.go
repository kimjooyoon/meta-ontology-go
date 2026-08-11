package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func makeFreshKey(t *testing.T, dependency, provenance int) (Key, Freshness) {
	t.Helper()
	spec := FreshnessSpec{
		Dependencies: map[string]any{"revision": dependency},
		Provenance:   map[string]any{"source": provenance},
	}
	freshness, err := NewFreshness(spec)
	if err != nil {
		t.Fatal(err)
	}
	key, err := NewKey(KeySpec{
		Version: "v1", Namespace: "billing", ToolVersion: "compiler-1",
		Inputs: map[string]any{"source": "main.gooo"}, Options: map[string]any{"mode": "fast"},
		Freshness: spec,
	})
	if err != nil {
		t.Fatal(err)
	}
	return key, freshness
}

func TestFreshnessCanonicalAndDistinct(t *testing.T) {
	first, err := NewFreshness(FreshnessSpec{
		Dependencies: map[string]any{"a": 1, "b": "two"},
		Provenance:   map[string]any{"commit": "abc", "files": []string{"a", "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewFreshness(FreshnessSpec{
		Dependencies: map[string]any{"b": "two", "a": 1},
		Provenance:   map[string]any{"files": []string{"a", "b"}, "commit": "abc"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !first.Equal(second) || !first.Valid() {
		t.Fatalf("map ordering changed freshness: %+v != %+v", first, second)
	}
	changed, err := NewFreshness(FreshnessSpec{
		Dependencies: map[string]any{"a": 2, "b": "two"},
		Provenance:   map[string]any{"commit": "abc", "files": []string{"a", "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if changed.DependencyDigest == first.DependencyDigest || changed.ProvenanceDigest != first.ProvenanceDigest {
		t.Fatalf("dependency change did not alter only dependency identity: %+v", changed)
	}
}

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
