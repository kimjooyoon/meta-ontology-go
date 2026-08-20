package cache

import (
	"testing"
)

func TestCacheObservationManifest(t *testing.T) {
	want := []cacheTestManifestEntry{
		{Name: "TestCacheLatencyEvidenceMatrix", Class: CacheTestClassSlowObservation},
		{Name: "TestCacheSameKeyCrossProcessStampede", Class: CacheTestClassSlowObservation},
		{Name: "TestIncrementalCacheMutationMatrix", Class: CacheTestClassSlowObservation},
	}
	if len(cacheTestManifest) != len(want) {
		t.Fatalf("cache observation manifest has %d entries, want %d", len(cacheTestManifest), len(want))
	}
	seen := make(map[string]bool, len(cacheTestManifest))
	for index, entry := range cacheTestManifest {
		if seen[entry.Name] {
			t.Fatalf("cache observation manifest duplicates %q", entry.Name)
		}
		seen[entry.Name] = true
		if entry != want[index] {
			t.Fatalf("cache observation manifest entry %d = %+v, want %+v", index, entry, want[index])
		}
		if entry.Class == "" {
			t.Fatalf("cache observation manifest entry %q has no class", entry.Name)
		}
		if cacheTestClassEnabled("", entry.Class) {
			t.Fatalf("cache observation manifest entry %q is enabled by default", entry.Name)
		}
		if !cacheTestClassEnabled(CacheTestClassSlowObservation, entry.Class) {
			t.Fatalf("cache observation manifest entry %q is not enabled by its class", entry.Name)
		}
	}
	if *cacheTestClass != "" && *cacheTestClass != CacheTestClassSlowObservation {
		t.Fatalf("unsupported cache test class %q", *cacheTestClass)
	}
}
func TestCacheObservationClassSelection(t *testing.T) {
	for _, test := range []struct {
		requested string
		class     string
		want      bool
	}{
		{requested: "", class: CacheTestClassSlowObservation, want: false},
		{requested: CacheTestClassSlowObservation, class: CacheTestClassSlowObservation, want: true},
		{requested: "other", class: CacheTestClassSlowObservation, want: false},
	} {
		if got := cacheTestClassEnabled(test.requested, test.class); got != test.want {
			t.Fatalf("cache test class selection requested=%q class=%q = %v, want %v", test.requested, test.class, got, test.want)
		}
	}
}
