package toolchainrelease

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOutcomeDeltaReleaseProfileRetainsComparedRecords(t *testing.T) {
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"matching", "changed", "command-failure"} {
		t.Run(mode, func(t *testing.T) {
			input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "fixture"}}
			calls := 0
			run := outcomeDeltaTestCommand(t, root, mode, &calls)
			err := runOutcomeDeltaSmoke("fixture-compiler", t.TempDir(), input, run)
			if (err == nil) != (mode == "matching") || calls != map[string]int{"matching": 2, "changed": 1, "command-failure": 1}[mode] {
				t.Fatal("unexpected comparison result", err, calls)
			}
			name := "fixture-outcome-delta-change.json"
			if mode == "command-failure" {
				name = "fixture-outcome-delta-change.failed-output"
			}
			if _, err := os.ReadFile(filepath.Join(input.OutputDir, name)); err != nil {
				t.Fatal("command output was not retained", err)
			}
		})
	}
}

func outcomeDeltaTestCommand(t *testing.T, root, mode string, calls *int) func(string, []string, string, ...string) ([]byte, error) {
	t.Helper()
	return func(dir string, env []string, binary string, args ...string) ([]byte, error) {
		*calls++
		if dir != root || len(env) != 0 || binary != "fixture-compiler" || len(args) != 6 ||
			args[0] != "body-outcomes-delta" || args[1] != "--before" || args[3] != "--after" || args[5] != "--json" {
			t.Fatal("unexpected operation", dir, binary, args)
		}
		for _, path := range []string{args[2], args[4]} {
			raw, err := os.ReadFile(path)
			if err != nil || !bytes.Contains(raw, []byte("9007199254740993")) {
				t.Fatal("original exact-integer input missing", err)
			}
		}
		if mode == "command-failure" {
			return []byte("failed comparison"), errors.New("exit 1")
		}
		if mode == "changed" {
			return []byte(`{"counts":{}}`), nil
		}
		name := "change.json"
		if *calls == 2 {
			name = "saved-replay-delta.json"
		}
		return os.ReadFile(filepath.Join(root, outcomeDeltaObservations, name))
	}
}

func TestOutcomeDeltaReleaseProfileOnPackagedCompiler(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "local-native"}}
	if output := os.Getenv("GOOO_OUTCOME_SMOKE_OUTPUT"); output != "" {
		input.OutputDir = output
		if err := os.MkdirAll(output, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := smokeOutcomeDelta(binary, t.TempDir(), input); err != nil {
		t.Fatal(err)
	}
}
