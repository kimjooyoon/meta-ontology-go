package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalVersionShapeDefaults(t *testing.T) {
	flags := map[string]string{"--origin-receipt": "saved.json", "--origin-shape": "local-version-columns"}
	query, err := packageOriginQuery(flags)
	if err != nil || query.Kind != "local-writes" || query.Limit != 32 {
		t.Fatal(query, err)
	}
	flags["--origin-kind"] = "local-uses"
	if _, err := packageOriginQuery(flags); err == nil {
		t.Fatal("local-use rows accepted as write identities")
	}
}

func TestGoooLocalVersionPolicyCases(t *testing.T) {
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", "../../examples/local-version-origins/policy-cases.json", "../../examples/local-version-origins/gooo.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("new Gooo policy cases", code, err, diag.String(), out.String())
	}
	runtime := receipt.Result.Runtime
	if receipt.Decision != "PASS" || runtime.FiniteTotal != 13 || runtime.FinitePassed != 13 || runtime.ModelCalls != 0 {
		t.Fatal("controlled policy outcomes", receipt.Decision, runtime.FiniteTotal, runtime.FinitePassed, runtime.ModelCalls)
	}
}

func TestGoooLocalVersionTailPage(t *testing.T) {
	saved := localVersionSnapshot(t)
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-shape", "local-version-columns", "--origin-offset", "16", "--origin-limit", "16", "../../examples/local-version-origins/gooo.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal("new tail page", code, err, diag.String(), out.String())
	}
	p := receipt.OriginInput
	if p.Query.Offset != 16 || p.TotalRows != 32 || p.ReturnedRows != 16 || p.Truncated || p.NextOffset != nil || p.ConsumerRows != 1 || p.ProducerExecutions != 0 || p.ProducerModelCalls != 0 || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("tail page scope", p)
	}
	var actual struct {
		Declarations int  `json:"page_declarations"`
		Calls        int  `json:"page_write_calls"`
		Versions     int  `json:"page_completed_versions"`
		Complete     bool `json:"page_complete"`
	}
	if err := json.Unmarshal(receipt.Result.Runtime.Traces[0].Deliveries[0].Actual, &actual); err != nil || actual.Declarations != 4 || actual.Calls != 6 || actual.Versions != 16 || actual.Complete {
		t.Fatal("last page is still a partial population", actual, err)
	}
}

func TestGoooOriginalFileVersionPolicy(t *testing.T) {
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", "../../examples/local-version-origins/original-file-policy-cases.json", "../../examples/local-version-origins/gooo.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("original-file policy inputs", code, err, diag.String(), out.String())
	}
	runtime := receipt.Result.Runtime
	if receipt.Decision != "PASS" || runtime.FiniteTotal != 7 || runtime.FinitePassed != 7 || runtime.ModelCalls != 0 {
		t.Fatal("original-file scope outcomes", receipt.Decision, runtime.FiniteTotal, runtime.FinitePassed, runtime.ModelCalls)
	}
}

func TestGoooOriginalFileVersionPage(t *testing.T) {
	saved := localVersionSnapshot(t)
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-shape", "local-version-columns", "--origin-offset", "8", "--origin-limit", "8", "../../examples/local-version-origins/gooo.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal("original-file page", code, err, diag.String(), out.String())
	}
	var actual struct {
		Scope        string `json:"declaration_scope"`
		Files        string `json:"file_sources_sha256"`
		Declarations int    `json:"page_declarations"`
		Calls        int    `json:"page_write_calls"`
		Versions     int    `json:"page_completed_versions"`
		Complete     bool   `json:"page_complete"`
	}
	if err := json.Unmarshal(receipt.Result.Runtime.Traces[0].Deliveries[0].Actual, &actual); err != nil || actual.Scope != "original_file" || actual.Files == "" || actual.Declarations != 4 || actual.Calls != 3 || actual.Versions != 8 || actual.Complete {
		t.Fatal("one case of zero values retains original identities", actual, err)
	}
	if p := receipt.OriginInput; p.InputShape != "gooo/local-version-columns/v2" || p.ProducerExecutions != 0 || p.ProducerModelCalls != 0 || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("read-only original page scope", p)
	}
}

func localVersionSnapshot(t *testing.T) string {
	t.Helper()
	f, err := os.Open("../../examples/local-version-origins/recorded-runtime.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(t.TempDir(), "local-versions.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return saved
}

func TestGoooRecordedGenerationClaim(t *testing.T) {
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--origin-receipt", localVersionSnapshot(t), "--origin-shape", "local-version-columns", "--origin-offset", "31", "--origin-limit", "1", "../../examples/local-version-origins/generation.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(code, err, diag.String(), out.String())
	}
	var claim struct {
		Relation   string `json:"relation_iri"`
		Case       int    `json:"case_index"`
		Assignment int    `json:"assignment_index"`
		Frame      int    `json:"frame"`
		Original   string `json:"original_activity_id"`
		Package    string `json:"package_path"`
		Filename   string `json:"filename"`
		Completed  bool   `json:"completed"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(receipt.Result.Runtime.Traces[0].Deliveries[0].Actual, &claim); err != nil || claim.Relation != "http://www.w3.org/ns/prov#wasGeneratedBy" || claim.Case != 3 || claim.Assignment != 7 || claim.Frame != 0 || claim.Original != "urn:gooo:version-producer:combine" || claim.Package != "origin/versionproducer" || claim.Filename != "main.gooo" || !claim.Completed || claim.Status != "RECORDED_GENERATION_CLAIM" {
		t.Fatal(claim, err)
	}
	if receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.ProducerModelCalls != 0 || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("recorded generation reran its producer or model")
	}
}

func TestGoooGenerationCompletionRules(t *testing.T) {
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", "../../examples/local-version-origins/generation-cases.json", "../../examples/local-version-origins/generation.workspace.json"}, &out, &diag)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal(code, err, diag.String(), out.String())
	}
	r := receipt.Result.Runtime
	if receipt.Decision != "PASS" || r.FiniteTotal != 2 || r.FinitePassed != 2 || r.ModelCalls != 0 {
		t.Fatal("completed and incomplete generation claims", receipt.Decision, r.FiniteTotal, r.FinitePassed, r.ModelCalls)
	}
}
