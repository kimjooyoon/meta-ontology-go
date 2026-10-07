package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageReplayThenGoooConstructionInterpretation(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..")
	workspace := filepath.Join(root, "examples", "workspace-input-observations")
	saved := filepath.Join(root, "docs", "research", "workspace-inputs-20261008", "execution.json")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "replay", "--json", "--receipt", saved,
		"--cases", filepath.Join(workspace, "cases.json"), filepath.Join(workspace, "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal("saved program did not replay", code, stdout.String(), stderr.String())
	}
	var replay packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &replay); err != nil {
		t.Fatal(err)
	}
	s := replay.Result.Runtime.InputSeparation
	if replay.Result.Runtime.FinitePassed != 8 || s.OverlappingInputs != 2 || s.DisjointInputs != 1 || s.DisjointCasesPassed != 1 {
		t.Fatal("replay lost earlier construction inputs", s)
	}
	path := filepath.Join(t.TempDir(), "replay.json")
	if err := os.WriteFile(path, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := run([]string{"package", "execute", "--json", "--construction-receipt", path,
		filepath.Join(root, "examples", "assembly-explainer", "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal("Gooo could not interpret the replayed program", code, stdout.String(), stderr.String())
	}
	var interpretation packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &interpretation); err != nil {
		t.Fatal(err)
	}
	o := interpretation.ConstructionInput
	if interpretation.Decision != "OBSERVED" || o == nil || len(o.Rows) != 2 || o.ModelCalls != 0 || interpretation.Result.Runtime.FiniteTotal != 0 {
		t.Fatal("construction interpretation mixed runtime and selection observations", interpretation)
	}
	for i, operation := range []string{"OBSERVE_NEW_INPUTS", "USE_OBSERVED_CANDIDATE"} {
		var value map[string]string
		if err := json.Unmarshal(interpretation.Result.Runtime.Traces[i].Deliveries[0].Actual, &value); err != nil || value["next_operation"] != operation {
			t.Fatal("Gooo next operation differs", i, value, err)
		}
	}
}
