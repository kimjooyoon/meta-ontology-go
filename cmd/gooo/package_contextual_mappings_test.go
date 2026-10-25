package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestContextualMappingClaimsExecuteAndReplayWithoutAuthority(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "contextual-mappings")
	manifest, cases := filepath.Join(root, "gooo.workspace.json"), filepath.Join(root, "cases.json")
	raw := runContextMappingExample(t, []string{"package", "execute", "--json", "--cases", cases, manifest})
	saved := filepath.Join(t.TempDir(), "mapping-claims.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	runContextMappingExample(t, []string{"package", "replay", "--json", "--receipt", saved, "--cases", cases, manifest})
}

func runContextMappingExample(t *testing.T, args []string) []byte {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil {
		t.Fatal("contextual mapping example did not complete", receipt)
	}
	runtime := receipt.Result.Runtime
	if runtime.FinitePassed != 40 || runtime.FiniteTotal != 40 || !runtime.RuntimeReplayed ||
		!runtime.ProjectionReplayed || runtime.ModelCalls != 0 {
		t.Fatal("mapping claims lost their finite/source replay obligations", runtime)
	}
	return stdout.Bytes()
}
