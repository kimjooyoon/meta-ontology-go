package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func TestPackageExecuteAndReplayImportedBodyCalls(t *testing.T) {
	manifest := "../../examples/package-body-calls/gooo.workspace.json"
	cases := "../../examples/package-body-calls/cases.json"
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", cases, manifest}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var envelope struct {
		Result workspaceexecution.Result `json:"result"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result.Runtime.FinitePassed != 8 || len(envelope.Result.Program.PureCalls.Sites) != 2 ||
		len(envelope.Result.Program.Activities) != 2 {
		t.Fatal("package body call execution", out.String())
	}
	receipt := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(receipt, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := run([]string{"package", "replay", "--json", "--receipt", receipt, "--cases", cases, manifest}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil || envelope.Result.Runtime.FinitePassed != 8 ||
		envelope.Result.Replay == nil || envelope.Result.Replay.ModelCalls != 0 {
		t.Fatal("imported body calls did not replay", err, out.String())
	}
}
