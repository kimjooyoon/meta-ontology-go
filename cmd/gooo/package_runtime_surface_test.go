package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimeSurfaceOwnsExpectationAndPreservesObservedPaths(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "runtime-package-surface")
	manifest, cases := filepath.Join(root, "gooo.workspace.json"), filepath.Join(root, "cases.json")
	raw, receipt, code := runtimeSurfaceCommand(t, []string{"package", "execute", "--json", "--cases", cases, manifest})
	if code != exitOK || receipt.Decision != "PASS" {
		t.Fatal("runtime surface cases failed", code, receipt.Decision)
	}
	saved := filepath.Join(t.TempDir(), "surface.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	_, replay, code := runtimeSurfaceCommand(t, []string{"package", "replay", "--json", "--receipt", saved, "--cases", cases, manifest})
	for _, observed := range []packageExecutionReceipt{receipt, replay} {
		if code != exitOK || observed.Result == nil || observed.Decision != "PASS" {
			t.Fatal("runtime surface replay failed", code, observed)
		}
		runtime := observed.Result.Runtime
		if runtime.FiniteTotal != 14 || runtime.FinitePassed != 14 || runtime.ModelCalls != 0 ||
			!runtime.RuntimeReplayed || !runtime.ProjectionReplayed {
			t.Fatal("surface lost its independent finite expectations or replay", runtime)
		}
	}
}

func runtimeSurfaceCommand(t *testing.T, args []string) ([]byte, packageExecutionReceipt, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err, stderr.String())
	}
	return stdout.Bytes(), receipt, code
}
