package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func measureProcessLatency(t *testing.T, processes int, prepopulate bool) []cacheLatencySample {
	t.Helper()
	root := t.TempDir()
	cache, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	key := makeTestKey(t, "v1", "billing")
	if prepopulate {
		if err := cache.Put(key, []byte(cacheLatencyValue)); err != nil {
			t.Fatal(err)
		}
	}
	release := filepath.Join(root, "release")
	commands := make([]*exec.Cmd, 0, processes)
	outputs := make([]*strings.Builder, 0, processes)
	readyFiles := make([]string, 0, processes)
	for index := range processes {
		ready := filepath.Join(root, fmt.Sprintf("ready-%d", index))
		result := filepath.Join(root, fmt.Sprintf("result-%d", index))
		output := &strings.Builder{}
		command := exec.Command(os.Args[0], "-test.run", "^TestCacheLatencyEvidenceMatrix$", "-cache-test-class="+CacheTestClassSlowObservation)
		command.Env = append(os.Environ(),
			cacheLatencyHelperEnv+"=1", cacheLatencyRootEnv+"="+root,
			cacheLatencyReleaseEnv+"="+release, cacheLatencyReadyEnv+"="+ready,
			cacheLatencyResultEnv+"="+result)
		command.Stdout, command.Stderr = output, output
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		commands = append(commands, command)
		outputs = append(outputs, output)
		readyFiles = append(readyFiles, ready)
	}
	if err := waitForFiles(readyFiles); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(release, []byte("go"), 0o600); err != nil {
		t.Fatal(err)
	}
	samples := make([]cacheLatencySample, 0, processes)
	for index, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatalf("latency helper %d failed: %v\n%s", index, err, outputs[index].String())
		}
		raw, err := os.ReadFile(filepath.Join(root, fmt.Sprintf("result-%d", index)))
		if err != nil {
			t.Fatal(err)
		}
		var sample cacheLatencySample
		if err := json.Unmarshal(raw, &sample); err != nil {
			t.Fatal(err)
		}
		samples = append(samples, sample)
	}
	if _, _, err := cache.Get(key); err != nil {
		t.Fatalf("final process-matrix object is not readable: %v", err)
	}
	return samples
}
