package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeSurfaceDeclarationChangeRequiresIndependentExpectation(t *testing.T) {
	for _, replacement := range []string{"unreviewed", "kernel"} {
		t.Run(replacement, func(t *testing.T) { runtimeSurfaceDeclarationMutation(t, replacement) })
	}
}

func runtimeSurfaceDeclarationMutation(t *testing.T, replacement string) {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"contract.gooo", "inspect.gooo", "gooo.workspace.json", "cases.json"} {
		raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "runtime-package-surface", name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "inspect.gooo" {
			before := string(raw)
			raw = []byte(strings.Replace(before, `module + "/internal/snapshotkernel"`, `module + "/internal/`+replacement+`"`, 1))
			if before == string(raw) {
				t.Fatal("expected package declaration was not changed")
			}
		}
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	_, receipt, code := runtimeSurfaceCommand(t, []string{"package", "execute", "--json",
		"--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json")})
	if code != exitOK || receipt.Decision != "PROGRESS" || receipt.Result == nil ||
		receipt.Result.Runtime.FiniteTotal != 14 || receipt.Result.Runtime.FinitePassed != 0 {
		t.Fatal("changed source expectation escaped the independent oracle", code, receipt.Decision)
	}
}
