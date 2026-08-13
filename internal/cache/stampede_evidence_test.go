package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func startStampedeHelpers(t *testing.T, root, release string) ([]*exec.Cmd, []*strings.Builder, []string) {
	t.Helper()
	commands := make([]*exec.Cmd, 0, 2)
	outputs := make([]*strings.Builder, 0, 2)
	readyFiles := make([]string, 0, 2)
	for index := 0; index < 2; index++ {
		ready := filepath.Join(root, fmt.Sprintf("ready-%d", index))
		output := &strings.Builder{}
		command := exec.Command(os.Args[0], "-test.run", "^TestCacheSameKeyCrossProcessStampede$")
		command.Env = append(os.Environ(),
			stampedeHelperEnv+"=1", stampedeRootEnv+"="+root, stampedeReleaseEnv+"="+release,
			stampedeReadyEnv+"="+ready, stampedeComputeEnv+"="+filepath.Join(root, fmt.Sprintf("compute-%d", index)),
			stampedeResultEnv+"="+filepath.Join(root, fmt.Sprintf("result-%d", index)),
		)
		command.Stdout, command.Stderr = output, output
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		commands = append(commands, command)
		outputs = append(outputs, output)
		readyFiles = append(readyFiles, ready)
	}
	return commands, outputs, readyFiles
}

func readStampedeResults(t *testing.T, root string, callers int) (int, int) {
	t.Helper()
	hits, misses := 0, 0
	for index := 0; index < callers; index++ {
		var result struct {
			Hit  bool   `json:"hit"`
			Data string `json:"data"`
		}
		raw, err := os.ReadFile(filepath.Join(root, fmt.Sprintf("result-%d", index)))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		if result.Data != stampedeExpectedProjection {
			t.Fatalf("helper %d returned %q, want %q", index, result.Data, stampedeExpectedProjection)
		}
		if result.Hit {
			hits++
		} else {
			misses++
		}
	}
	return hits, misses
}

func runStampedeHelper(t *testing.T) {
	t.Helper()
	cache, err := Open(os.Getenv(stampedeRootEnv))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv(stampedeReadyEnv), []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := waitForFiles([]string{os.Getenv(stampedeReleaseEnv)}); err != nil {
		t.Fatal(err)
	}
	data, _, hit, err := cache.GetOrCompute(context.Background(), makeTestKey(t, "v1", "billing"), func() ([]byte, error) {
		if err := os.WriteFile(os.Getenv(stampedeComputeEnv), []byte("compute"), 0o600); err != nil {
			return nil, err
		}
		time.Sleep(25 * time.Millisecond)
		return []byte(stampedeExpectedProjection), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(struct {
		Hit  bool   `json:"hit"`
		Data string `json:"data"`
	}{Hit: hit, Data: string(data)})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv(stampedeResultEnv), raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func waitForFiles(paths []string) error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		ready := true
		for _, path := range paths {
			if _, err := os.Stat(path); err != nil {
				ready = false
				break
			}
		}
		if ready {
			return nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %v", paths)
}
