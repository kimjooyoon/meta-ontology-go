package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoooOriginCoverage11983(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	dir := filepath.Join(root, "examples/derived-origin")
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(dir, "coverage-cases.json"), filepath.Join(dir, "coverage.workspace.json")}, &output, &diagnostics); code != exitOK {
		t.Fatal("Gooo source-owned coverage", code, diagnostics.String(), output.String())
	}
	saveGrammarOutput11903(t, "coverage11983.json", output.Bytes())
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil {
		t.Fatal(err)
	}
	runtime := receipt.Result.Runtime
	if runtime.FinitePassed != 8 || runtime.FiniteTotal != 8 || runtime.ModelCalls != 0 || !runtime.RuntimeReplayed {
		t.Fatal("coverage output population", runtime.FinitePassed, runtime.FiniteTotal)
	}
	if receipt.Result.Composition.Plan.EntryActivity == "" {
		t.Fatal("coverage entry identity missing")
	}
	t.Log("Gooo source owns coverage math8/8:actual whole/prefix pages,missing origin,repeated write fragments,partial recording,unaligned/order contradictions,empty population; no original-producer rerun; models0")
}
