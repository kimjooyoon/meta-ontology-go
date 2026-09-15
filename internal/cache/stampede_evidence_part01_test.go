package cache

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	stampedeHelperEnv          = "GOOO_CACHE_STAMPEDE_HELPER"
	stampedeRootEnv            = "GOOO_CACHE_STAMPEDE_ROOT"
	stampedeReleaseEnv         = "GOOO_CACHE_STAMPEDE_RELEASE"
	stampedeReadyEnv           = "GOOO_CACHE_STAMPEDE_READY"
	stampedeComputeEnv         = "GOOO_CACHE_STAMPEDE_COMPUTE"
	stampedeResultEnv          = "GOOO_CACHE_STAMPEDE_RESULT"
	stampedeExpectedProjection = "cross-process-stable"
)

func TestCacheSameKeyCrossProcessStampede(t *testing.T) {
	requireClassifiedCacheTest(t, "TestCacheSameKeyCrossProcessStampede", CacheTestClassSlowObservation)
	if os.Getenv(stampedeHelperEnv) == "1" {
		runStampedeHelper(t)
		return
	}
	root := t.TempDir()
	release := filepath.Join(root, "release")
	commands, outputs, readyFiles := startStampedeHelpers(t, root, release)
	if err := waitForFiles(readyFiles); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(release, []byte("go"), 0o600); err != nil {
		t.Fatal(err)
	}
	for index, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatalf("helper %d failed: %v\n%s", index, err, outputs[index].String())
		}
	}
	hits, misses := readStampedeResults(t, root, len(commands))
	cache, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	data, _, err := cache.Get(makeTestKey(t, "v1", "billing"))
	if err != nil || string(data) != stampedeExpectedProjection {
		t.Fatalf("cross-process durable result = %q, err=%v", data, err)
	}
	computeFiles, err := filepath.Glob(filepath.Join(root, "compute-*"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cross-process same-key callers=%d hits=%d misses=%d recomputations=%d; final object verified", len(commands), hits, misses, len(computeFiles))
	if len(computeFiles) == 0 || len(computeFiles) > len(commands) {
		t.Fatalf("cross-process recomputation count=%d, callers=%d", len(computeFiles), len(commands))
	}
}
