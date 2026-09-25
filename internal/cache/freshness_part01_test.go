package cache

import (
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
		Inputs: map[string]any{"source": "main.gooo"}, OptionsDigest: mustOptionsDigest(map[string]any{"mode": "fast"}),
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
