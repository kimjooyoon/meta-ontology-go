package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestRecordedRDFControlledCases(t *testing.T) {
	const root = "../../examples/recorded-use-origins/"
	for batch := 1; batch <= 6; batch++ {
		name := fmt.Sprintf("rdf-cases/rdf-cases-%d.json", batch)
		t.Run(name, func(t *testing.T) { checkRecordedRDFCases(t, root+name, root+"rdf.workspace.json", 2) })
	}
	for _, helper := range []struct {
		name  string
		count int
	}{{"Decimal", 12}, {"IsSHA256", 10}, {"RDFLiteral", 10}} {
		t.Run(helper.name, func(t *testing.T) {
			checkRecordedRDFCases(t, root+"rdf-cases/"+helper.name+"-cases.json", root+helper.name+".workspace.json", helper.count)
		})
	}
}

func checkRecordedRDFCases(t *testing.T, cases, workspace string, count int) {
	t.Helper()
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", cases, workspace}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("Gooo RDF source", code, err, receipt.Error, diag.String())
	}
	r := receipt.Result.Runtime
	if receipt.Decision != "PASS" || r.FiniteTotal != count || r.FinitePassed != count || r.ModelCalls != 0 {
		t.Fatal("finite source-owned RDF outcomes", receipt.Decision, r.FiniteTotal, r.FinitePassed, r.ModelCalls)
	}
}

func TestRecordedRDFSavedUse(t *testing.T) {
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--origin-receipt", recordedExplanationReceipt(t),
		"--origin-shape", "local-use-columns", "--origin-offset", "0",
		"../../examples/recorded-use-origins/rdf.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal("saved RDF source", code, err, receipt.Error, diag.String())
	}
	var claim struct {
		Status string   `json:"status"`
		Scope  string   `json:"scope"`
		Call   string   `json:"call_iri"`
		Value  string   `json:"value_iri"`
		Usage  string   `json:"usage_iri"`
		Lines  []string `json:"lines"`
	}
	r, origin := receipt.Result.Runtime, receipt.OriginInput
	if len(r.Traces) != 1 || len(r.Traces[0].Deliveries) != 1 {
		t.Fatal("one retained use required")
	}
	if err := json.Unmarshal(r.Traces[0].Deliveries[0].Actual, &claim); err != nil {
		t.Fatal(err)
	}
	const savedSHA = "sha256:a15f33ca09f99265aa9b838dcd4628a9dd74a8b77dcff93e1156241738a5441a"
	const base = "urn:gooo:saved:" + savedSHA + ":case:0"
	if receipt.Decision != "OBSERVED" || claim.Status != "RDF_READY" || claim.Scope != "saved_json_bytes" ||
		claim.Call != base+":frame:1" || claim.Value != base+":version:1" || claim.Usage != base+":use:0" || len(claim.Lines) != 26 {
		t.Fatal("saved-document identities changed", claim)
	}
	if origin.ProducerExecutions != 0 || origin.ProducerModelCalls != 0 || r.ModelCalls != 0 ||
		origin.ReceiptSHA256 != savedSHA {
		t.Fatal("saved source scope changed or producer/model ran", origin)
	}
}
