package cache

import (
	"flag"
	"testing"
)

// CacheTestClassSlowObservation is the explicit class for cache tests that
// exercise persistence, contention, large fixtures, or cross-process paths.
// The default test run leaves this class closed. Scheduled observation runs
// open it with: go test ./internal/cache -args -cache-test-class=slow-observation
const CacheTestClassSlowObservation = "slow-observation"

var cacheTestClass = flag.String("cache-test-class", "", "run one explicit internal/cache test class")

type cacheTestManifestEntry struct {
	Name  string
	Class string
}

var cacheTestManifest = []cacheTestManifestEntry{
	{Name: "TestCacheLatencyEvidenceMatrix", Class: CacheTestClassSlowObservation},
	{Name: "TestCacheSameKeyCrossProcessStampede", Class: CacheTestClassSlowObservation},
	{Name: "TestIncrementalCacheMutationMatrix", Class: CacheTestClassSlowObservation},
}

func requireClassifiedCacheTest(t *testing.T, name, class string) {
	t.Helper()
	entry, ok := lookupCacheTestManifestEntry(name)
	if !ok {
		t.Fatalf("test %q is not registered in the cache observation manifest", name)
	}
	if entry.Class != class {
		t.Fatalf("test %q registered as class %q, got %q", name, entry.Class, class)
	}
	if !cacheTestClassEnabled(*cacheTestClass, class) {
		t.Skipf("cache test class %q is closed; open with -cache-test-class=%s", class, class)
	}
}

func cacheTestClassEnabled(requested, class string) bool {
	return requested != "" && requested == class
}

func lookupCacheTestManifestEntry(name string) (cacheTestManifestEntry, bool) {
	for _, entry := range cacheTestManifest {
		if entry.Name == name {
			return entry, true
		}
	}
	return cacheTestManifestEntry{}, false
}

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
