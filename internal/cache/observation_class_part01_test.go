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
