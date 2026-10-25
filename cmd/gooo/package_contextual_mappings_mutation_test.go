package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestContextualMappingOracleDetectsAddedAuthority(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	for _, name := range []string{"mapping.gooo.fixture", "gooo.workspace.json", "cases.json"} {
		raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "contextual-mappings", name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "mapping.gooo.fixture" {
			for _, field := range []string{"identity_merge", "behavior_verified", "permission_granted"} {
				raw = bytes.ReplaceAll(raw, []byte(field+": false"), []byte(field+": true"))
			}
		}
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"),
		filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PROGRESS" || receipt.Result == nil {
		t.Fatal(receipt)
	}
	runtime := receipt.Result.Runtime
	if runtime.FinitePassed != 20 || runtime.FiniteTotal != 40 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("repeatable false authority escaped the independent expected records", runtime)
	}
}
