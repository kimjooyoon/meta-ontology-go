package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoooReadsCompoundTitleOrigin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recorded-runtime.json")
	if err := os.WriteFile(path, readCompoundRuntimeFixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--origin-receipt", path,
		"--origin-shape", "field-lineage-columns", "--origin-offset", "11", "--origin-limit", "1",
		"../../examples/native-field-lineage/gooo.workspace.json"}, &out, &diagnostics)
	if dir := os.Getenv("GOOO_COMPOUND_READER_OUTPUT"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "title-lineage.json"), out.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err)
	}
	r := receipt.Result.Runtime
	if receipt.Decision != "OBSERVED" || r.FiniteTotal != 0 || r.ModelCalls != 0 || len(r.Traces) != 1 || receipt.OriginInput.ProducerExecutions != 0 {
		t.Fatal("title reader scope")
	}
	var actual struct {
		Status, Roots string
		RootCount     int `json:"root_count"`
	}
	if len(r.Traces[0].Deliveries) != 1 || json.Unmarshal(r.Traces[0].Deliveries[0].Actual, &actual) != nil {
		t.Fatal("title reader output")
	}
	if actual.Status != "SCOPED_FIELD_ORIGINS" || actual.RootCount != 1 || actual.Roots != "urn:gooo:compound:summarize|input0|urn:gooo:compound:title;" {
		t.Fatal("title origin", actual)
	}
}
