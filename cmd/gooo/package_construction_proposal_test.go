package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestPackageConstructionRetainsRecordedModelProposal(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--construction-receipt",
		filepath.Join(root, "docs", "research", "workspace-replay-20261007", "model.json"),
		filepath.Join(root, "examples", "assembly-explainer", "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal("saved model construction did not reach Gooo", code, stdout.String(), stderr.String())
	}
	var result packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.ConstructionInput == nil || len(result.ConstructionInput.Rows) != 2 || result.ConstructionInput.ModelCalls != 0 {
		t.Fatal("construction observation scope differs", result)
	}
	first, selected := result.ConstructionInput.Rows[0], result.ConstructionInput.Rows[1]
	if first.CandidateMask == nil || *first.CandidateMask != 7 || !first.Proposed || first.Selected || first.Counts.Matched != 3 || first.Counts.Total != 5 {
		t.Fatal("incomplete recorded model proposal was not identified", first)
	}
	if selected.CandidateMask == nil || *selected.CandidateMask != 3 || selected.Proposed || !selected.Selected || selected.Counts.Matched != 5 {
		t.Fatal("finite continuation was confused with the model proposal", selected)
	}
}
