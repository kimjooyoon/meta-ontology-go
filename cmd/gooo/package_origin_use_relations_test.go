package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordedUsageShapeDefaults(t *testing.T) {
	flags := map[string]string{"--origin-receipt": "saved.json", "--origin-shape": "local-use-columns"}
	query, err := packageOriginQuery(flags)
	if err != nil || query.Kind != "local-uses" || query.Limit != 1 {
		t.Fatal(query, err)
	}
	flags["--origin-kind"] = "local-writes"
	if _, err := packageOriginQuery(flags); err == nil {
		t.Fatal("incompatible origin kind accepted")
	}
}

func TestRecordedUsageControlledPolicy(t *testing.T) {
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", "../../examples/recorded-use-origins/policy-cases.json", "../../examples/recorded-use-origins/gooo.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("new usage policy", code, err, receipt.Error, diag.String())
	}
	if r := receipt.Result.Runtime; receipt.Decision != "PASS" || r.FiniteTotal != 12 || r.FinitePassed != 12 || r.ModelCalls != 0 {
		t.Fatal("controlled usage outcomes", receipt.Decision, r.FiniteTotal, r.FinitePassed, r.ModelCalls)
	}
}

func TestRecordedUsageScopesSavedJSONBytes(t *testing.T) {
	raw, err := os.ReadFile(localVersionSnapshot(t))
	if err != nil {
		t.Fatal(err)
	}
	var previous string
	for _, input := range [][]byte{raw, append(append([]byte(nil), raw...), '\n')} {
		file := filepath.Join(t.TempDir(), "saved-usage.json")
		if err := os.WriteFile(file, input, 0600); err != nil {
			t.Fatal(err)
		}
		var out, diag bytes.Buffer
		code := run([]string{"package", "execute", "--json", "--origin-receipt", file, "--origin-shape", "local-use-columns", "--origin-offset", "15", "../../examples/recorded-use-origins/gooo.workspace.json"}, &out, &diag)
		var receipt packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil || receipt.OriginInput == nil {
			t.Fatal("scoped recorded use", code, err, receipt.Error, diag.String())
		}
		var claim struct {
			Receipt      string `json:"receipt_sha256"`
			Scope        string `json:"receipt_scope"`
			Generation   string `json:"generation_relation_iri"`
			Usage        string `json:"use_relation_iri"`
			Case         int    `json:"case_index"`
			Use          int    `json:"use_index"`
			Frame        int    `json:"frame"`
			Assignment   int    `json:"assignment_index"`
			VersionFrame int    `json:"version_frame"`
			Name         string `json:"variable_name"`
			Status       string `json:"status"`
		}
		value := receipt.Result.Runtime.Traces[0].Deliveries[0].Actual
		if err := json.Unmarshal(value, &claim); err != nil || claim.Receipt != workspaceDigest(input) || claim.Receipt == previous || claim.Scope != "saved_json_bytes" || claim.Case != 1 || claim.Use != 7 || claim.Frame != 0 || claim.Assignment != 7 || claim.VersionFrame != 0 || claim.Name != "total" || claim.Status != "RECORDED_GENERATION_AND_USAGE" {
			t.Fatal("saved scope or original reference changed", claim, err)
		}
		if claim.Generation != "http://www.w3.org/ns/prov#wasGeneratedBy" || claim.Usage != "http://www.w3.org/ns/prov#used" || bytes.Contains(value, []byte("wasDerivedFrom")) {
			t.Fatal("relation meaning changed", claim)
		}
		p := receipt.OriginInput
		if p.ReceiptSHA256 != claim.Receipt || p.Query.Limit != 1 || p.ReturnedRows != 1 || p.TotalRows != 32 || p.ConsumerRows != 1 || p.ProducerExecutions != 0 || p.ProducerModelCalls != 0 || receipt.Result.Runtime.ModelCalls != 0 {
			t.Fatal("recorded lookup executed producer/model or lost scope", p)
		}
		previous = claim.Receipt
	}
}
