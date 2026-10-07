package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageConstructionExplainsImportedExternalPlans(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "package-imports")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"),
		"--body-plans", filepath.Join(root, "body-plans.json"), filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal("external construction failed", code, stdout.String(), stderr.String())
	}
	path := filepath.Join(t.TempDir(), "execution.json")
	if err := os.WriteFile(path, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := run([]string{"package", "execute", "--json", "--construction-receipt", path,
		filepath.Join("..", "..", "examples", "assembly-explainer", "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal("external construction interpretation failed", code, stdout.String(), stderr.String())
	}
	var observed packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Decision != "OBSERVED" || observed.ConstructionInput == nil || len(observed.ConstructionInput.Rows) != 4 || observed.ConstructionInput.ModelCalls != 0 || observed.Result.Runtime.FiniteTotal != 0 {
		t.Fatal("construction explanation lost its observation scope", observed)
	}
	for i, row := range observed.ConstructionInput.Rows {
		if row.Profile != "external_fill" || row.View != "scored_set" || row.Counts.Total != 3 || row.Counts.Budget != 2 || row.InputIndex == nil || *row.InputIndex != i {
			t.Fatal("external plan was labeled as source-owned or miscounted", row)
		}
		want := "OBSERVE_NEW_INPUTS"
		if i%2 == 1 {
			want = "USE_OBSERVED_CANDIDATE"
		}
		var output map[string]string
		if err := json.Unmarshal(observed.Result.Runtime.Traces[i].Deliveries[0].Actual, &output); err != nil || output["next_operation"] != want {
			t.Fatal("Gooo explanation differs", i, output, err)
		}
	}
}
