package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestBatchedBridgeOrigins16841(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	example := filepath.Join(root, "examples/local-field-bridge")
	f, err := os.Open(filepath.Join(example, "recorded-runtime.json.gz"))
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
	var r bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(t.TempDir(), "batch16841.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	args := []string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-shape", "field-local-columns", "--origin-offset", "0", "--origin-limit", "3", filepath.Join(example, "trace.workspace.json")}
	if code := run(args, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err, receipt.Error)
	}
	if receipt.OriginInput.ConsumerRows != 3 || receipt.OriginInput.ReturnedRows != 3 || receipt.OriginInput.ProducerExecutions != 0 || receipt.Result.Runtime.ModelCalls != 0 || receipt.Result.Runtime.FiniteTotal != 0 || len(receipt.Result.Runtime.Traces) != 3 {
		t.Fatal("batch transport scope", receipt.OriginInput)
	}
	for i, trace := range receipt.Result.Runtime.Traces {
		use := r.Traces[0].FieldUses[i]
		roots, _ := nativeBridgeRoots15217(r, 0, use.Assignment)
		var actual struct {
			Status string `json:"status"`
			Roots  string `json:"roots"`
			Count  int    `json:"root_count"`
		}
		if len(roots) != 1 || len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
			t.Fatal("batch result populations")
		}
		if actual.Status != "SCOPED_FIELD_LOCAL_ORIGINS" || actual.Count != 1 || actual.Roots != "urn:gooo:local-bridge:build|input|"+roots[0]+";" {
			t.Fatal("batch source identities", i, actual, roots)
		}
	}
	cross, report, err := bodyexecution.SelectRecordedBridgeLineagePage(r, bodyexecution.NativeOriginQuery{Kind: "field-uses", Offset: len(r.Traces[0].FieldUses) - 1, Limit: 2})
	if err != nil || len(cross) != 2 || cross[0].CaseIndex != 0 || cross[1].CaseIndex != 1 || report.ConsumerRows != 2 {
		t.Fatal("independent case namespace", err)
	}
	_, report, err = packageOriginInputsWithShape(raw, packageruntime.EntrySpec{PackagePath: "origin/bridgetrace", Activity: "Trace"}, bodyexecution.NativeOriginQuery{Kind: "field-uses", Limit: 128}, "field-local-columns")
	if err == nil || !strings.Contains(err.Error(), "32768") || report == nil || report.ProducerExecutions != 0 {
		t.Fatal("whole page bound", err)
	}
	if _, _, err := bodyexecution.SelectRecordedBridgeLineage(r, bodyexecution.NativeOriginQuery{Kind: "field-uses", Limit: 3}); err == nil {
		t.Fatal("single-use API changed")
	}
	if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
		f, e := os.Create(filepath.Join(dir, "batch16841.json.gz"))
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
	t.Log("three new saved initial-field queries ->one Gooo invocation3/3;cross-case namespace2/2;32KiB page enforced;old single-use API retained;producer/model0;input-only finite0")
}
