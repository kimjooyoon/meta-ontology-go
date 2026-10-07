package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestPackageConstructionReceiptFeedsGoooExplainer(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..")
	path := filepath.Join(root, "docs", "research", "domain-tools-20261007", "documentation-model.json")
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--construction-receipt", path,
		filepath.Join(root, "examples", "assembly-explainer", "gooo.workspace.json")}, &stdout, &stderr)
	if code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	o := receipt.ConstructionInput
	if receipt.Decision != "OBSERVED" || o == nil || len(o.Rows) != 2 || o.ModelCalls != 0 {
		t.Fatal("construction input was not bound", receipt)
	}
	c := o.Rows[1].Counts
	if c.Matched != 3 || c.Total != 3 || c.Best != 3 || c.Scored != 2 || c.Budget != 8 {
		t.Fatal("counts differ from the actual saved model run", c)
	}
	first := o.Rows[0].Counts
	if first.Matched != 2 || first.Total != 3 || first.Best != 2 || first.Scored != 1 {
		t.Fatal("first attempt used field counts or a future successful candidate", first)
	}
	r := receipt.Result.Runtime
	if !r.RuntimeReplayed || r.FiniteTotal != 0 || r.ModelCalls != 0 || len(r.Traces) != 2 {
		t.Fatal("explainer observation became historical runtime success", r)
	}
	for i, operation := range []string{"CONTINUE_CANDIDATES", "OBSERVE_NEW_INPUTS"} {
		var actual map[string]string
		if err := json.Unmarshal(r.Traces[i].Deliveries[0].Actual, &actual); err != nil || actual["next_operation"] != operation {
			t.Fatal("Gooo did not interpret the real construction counts", err, actual)
		}
	}
}

func TestPackageConstructionInputRejectsConflictingModesAndEvidence(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--construction-receipt", "saved.json", "--inputs", "inputs.json", "manifest.json"}, &stdout, &stderr); code != exitUsage {
		t.Fatal("two input modes were accepted", code)
	}
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "research", "domain-tools-20261007", "assembly-model.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved packageExecutionReceipt
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	saved.Result.Composition.Steps[0].Generation.Report.RecordAssembly.Total++
	changed, _ := json.Marshal(saved)
	if _, _, err := packageConstructionInputs(context.Background(), changed, packageruntime.EntrySpec{}); err == nil {
		t.Fatal("altered construction evidence was interpreted")
	}
	if _, _, err := packageConstructionInputs(context.Background(), []byte(`{"schema":"x","schema":"y"}`), packageruntime.EntrySpec{}); err == nil {
		t.Fatal("duplicate envelope keys were accepted")
	}
}
