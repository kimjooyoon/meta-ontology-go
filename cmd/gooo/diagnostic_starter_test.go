package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticStarterExecutesAndReplays(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	for _, module := range []string{"", "example.org/team/diagnostics"} {
		t.Run(module, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "diagnostic")
			args := []string{"init", "--template", "diagnostic"}
			if module != "" {
				args = append(args, "--module", module)
			}
			var stdout, stderr bytes.Buffer
			if code := run(append(args, root), &stdout, &stderr); code != exitOK {
				t.Fatal(code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), "Created Gooo diagnostic starter") {
				t.Fatal(stdout.String())
			}
			checkDiagnosticStarterExecution(t, root)
		})
	}
}

func checkDiagnosticStarterExecution(t *testing.T, root string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	manifest := filepath.Join(root, "gooo.workspace.json")
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), manifest}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PASS" || receipt.Result.Runtime.FinitePassed != 8 || receipt.Result.Runtime.FiniteTotal != 8 {
		t.Fatal("starter did not satisfy its eight output expectations", receipt)
	}
	path := filepath.Join(root, "execution.json")
	if err := os.WriteFile(path, stdout.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	moved := root + "-moved"
	if err := os.Rename(root, moved); err != nil {
		t.Fatal(err)
	}
	root = moved
	manifest = filepath.Join(root, "gooo.workspace.json")
	path = filepath.Join(root, "execution.json")
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"package", "replay", "--json", "--receipt", path, "--inputs", filepath.Join(root, "inputs.json"), manifest}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "OBSERVED" || receipt.Result.Replay == nil || receipt.Result.Replay.ModelCalls != 0 || receipt.Result.Runtime.FiniteTotal != 0 {
		t.Fatal("starter replay did not preserve input-only scope", receipt)
	}
	if !strings.Contains(stdout.String(), "missing branch result [repair-and-replay]") {
		t.Fatal("source-owned diagnostic output was lost", stdout.String())
	}
}
