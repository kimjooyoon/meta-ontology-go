package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestGoooCandidateFieldSources16703(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata/field-fill16541-runtime.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	raw, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	var recorded bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &recorded); err != nil {
		t.Fatal(err)
	}
	rows, _, err := bodyexecution.SelectRecordedNativeOrigins(recorded, bodyexecution.NativeOriginQuery{Kind: "field-input-locations", Limit: 64})
	if err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(t.TempDir(), "field-fill16703.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	args := []string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-kind", "field-input-locations", "--origin-limit", "64", filepath.Join(root, "examples/field-source-map/gooo.workspace.json")}
	if code := run(args, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err, receipt.Error)
	}
	if len(receipt.Result.Runtime.Traces) != len(rows) || receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.ProducerModelCalls != 0 || receipt.Result.Runtime.ModelCalls != 0 || receipt.Result.Runtime.FiniteTotal != 0 {
		t.Fatal("saved field fill scope")
	}
	candidates, templates, unresolved := 0, 0, 0
	for i, trace := range receipt.Result.Runtime.Traces {
		var fact bodyexecution.CompositionFieldFileFact
		if err := json.Unmarshal(rows[i], &fact); err != nil {
			t.Fatal(err)
		}
		var actual struct {
			Status     string `json:"status"`
			Kind       string `json:"anchor_kind"`
			Mapped     bool   `json:"mapped"`
			Assembly   bool   `json:"assembly_origin"`
			Candidate  string `json:"candidate_id"`
			Expression string `json:"expression"`
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
			t.Fatal("Gooo field fill response")
		}
		wantKind := map[string]string{"target": "FIELD_TARGET", "value": "FIELD_VALUE", "declared_candidate": "ASSEMBLY_CANDIDATE", "retained_template": "ASSEMBLY_TEMPLATE", "derived_grammar": "ASSEMBLY_GRAMMAR"}[fact.Category]
		assembly := fact.Category == "declared_candidate" || fact.Category == "retained_template" || fact.Category == "derived_grammar"
		status := "UNRESOLVED"
		if fact.Mapping == "MAPPED" {
			status = "ORIGINAL_FILE"
		}
		if actual.Kind != wantKind || actual.Status != status || actual.Mapped != (fact.Mapping == "MAPPED") || actual.Assembly != assembly || actual.Candidate != fact.CandidateID || actual.Expression != fact.Expression {
			t.Fatal("Gooo candidate classification", fact, actual)
		}
		if actual.Kind == "ASSEMBLY_CANDIDATE" {
			candidates++
		}
		if actual.Kind == "ASSEMBLY_TEMPLATE" {
			templates++
		}
		if fact.Category == "value" && !actual.Mapped {
			unresolved++
		}
	}
	if candidates != 2 || templates == 0 || unresolved != 2 {
		t.Fatal("candidate/template/unresolved populations", candidates, templates, unresolved)
	}
	if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
		f, e := os.Create(filepath.Join(dir, "reader16703.json.gz"))
		if e != nil {
			t.Fatal(e)
		}
		z := gzip.NewWriter(f)
		if _, e = z.Write(out.Bytes()); e != nil {
			t.Fatal(e)
		}
		if e = z.Close(); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("new assembly-origin input cohort:Gooo field anchors%d/%d;candidate2,template%d,cross-fragment unresolved values2;producer executions0;model0;input-only finite0", len(rows), len(rows), templates)
}
