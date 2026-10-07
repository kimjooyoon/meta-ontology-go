package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestRunPackageExecuteSupportsIndependentRecordNamespaces(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "package-record-namespaces")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr)
	if code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	result := receipt.Result
	if result == nil || result.Runtime.FinitePassed != 12 || result.Runtime.FiniteTotal != 12 || !result.Runtime.RuntimeReplayed || len(result.Program.EntityAliases) != 2 {
		t.Fatal("record namespace example did not execute", receipt)
	}
	if result.Runtime.ModelCalls != 0 || len(result.Composition.Plan.Records) != 2 || len(result.Program.Activities) != 3 {
		t.Fatal("record identities or deterministic route changed")
	}
}
