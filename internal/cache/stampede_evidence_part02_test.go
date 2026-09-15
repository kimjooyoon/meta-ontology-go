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

func startStampedeHelpers(t *testing.T, root, release string) ([]*exec.Cmd, []*strings.Builder, []string) {
	t.Helper()
	commands := make([]*exec.Cmd, 0, 2)
	outputs := make([]*strings.Builder, 0, 2)
	readyFiles := make([]string, 0, 2)
	for index := range 2 {
		ready := filepath.Join(root, fmt.Sprintf("ready-%d", index))
		output := &strings.Builder{}
		command := exec.Command(os.Args[0], "-test.run", "^TestCacheSameKeyCrossProcessStampede$", "-cache-test-class="+CacheTestClassSlowObservation)
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
	for index := range callers {
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
