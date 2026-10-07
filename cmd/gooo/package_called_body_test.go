package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageCalledBodyConstructionPolicyReplayAndExplanation(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := "../../examples/called-body-construction/"
	var out, diagnostics bytes.Buffer
	args := []string{"package", "execute", "--json", "--cases", root + "cases.json",
		"--assembly-policy-workspace", "../../examples/package-assembly-policy/gooo.workspace.json", root + "gooo.workspace.json"}
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	r := receipt.Result
	if r == nil || r.Runtime.FinitePassed != 4 || len(r.Composition.Preparations) != 1 ||
		len(r.Program.Activities) != 1 || r.AssemblyPolicy == nil || r.Runtime.InputSeparation.Status != "PASS" ||
		r.Runtime.InputSeparation.DisjointCasesPassed != 4 || len(r.Runtime.Traces[3].Calls) != 1 ||
		string(r.Runtime.Traces[3].Calls[0].Inputs[0]) != "9007199254740994" {
		t.Fatal("called package body was not constructed", out.String())
	}
	record := r.Composition.Preparations[0].Generation.Report.RecordAssembly
	if record == nil || record.Passed != 5 || record.Control == nil || len(record.Control.Decisions) != 4 {
		t.Fatal("helper construction lost finite checks or Gooo policy", record)
	}
	path := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(path, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	checkCalledPackageReplay(t, root, path)
	out.Reset()
	if code := run([]string{"package", "execute", "--json", "--construction-receipt", path,
		"../../examples/assembly-explainer/gooo.workspace.json"}, &out, &diagnostics); code != exitOK {
		t.Fatal("Gooo could not explain called-body attempts", code, diagnostics.String(), out.String())
	}
	var explained packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &explained); err != nil || explained.ConstructionInput == nil ||
		len(explained.ConstructionInput.Rows) != 4 || explained.ConstructionInput.ModelCalls != 0 ||
		explained.ConstructionInput.Rows[3].Counts.Matched != 5 {
		t.Fatal("helper attempt observations lost", err, out.String())
	}
}

func checkCalledPackageReplay(t *testing.T, root, path string) {
	t.Helper()
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "replay", "--json", "--receipt", path, "--cases", root + "cases.json",
		root + "gooo.workspace.json"}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var replay packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &replay); err != nil || replay.Result.Runtime.FinitePassed != 4 ||
		replay.Result.Replay.ModelCalls != 0 || len(replay.Result.Composition.Preparations) != 1 ||
		replay.Result.Runtime.InputSeparation.DisjointCasesPassed != 4 {
		t.Fatal("called construction not retained on replay", err, out.String())
	}
}
