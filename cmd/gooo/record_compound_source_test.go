package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestRecordCompoundExampleSourceCoherence(t *testing.T) {
	for _, path := range []string{"../../examples/compound-updates/fields.gooo", "../../examples/local-field-bridge/trace.gooo", "../../examples/native-field-lineage/lineage.gooo"} {
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"check", "--semantic", "--json", path}, {"roundtrip", "--json", path}} {
			var out, diagnostics bytes.Buffer
			if code := run(args, &out, &diagnostics); code != exitOK {
				t.Fatal(path, code, diagnostics.String(), out.String())
			}
			var report jsonReport
			if err := json.Unmarshal(out.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
			if args[0] == "roundtrip" && (report.Equivalent == nil || !*report.Equivalent || report.GetPut == nil || !*report.GetPut || report.PutGet == nil || !*report.PutGet || report.OriginalSemanticHash != report.RoundTrippedSemanticHash) {
				t.Fatal("source/BX coherence", path, report)
			}
		}
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatal("readonly source query changed bytes", path, err)
		}
	}
}

func TestRecordCompoundExampleWorkspaceResolution(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "resolve", "--json", "../../examples/compound-updates/fields.workspace.json"}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Decision != "PASS" || receipt.Result == nil {
		t.Fatal("workspace source resolution", err)
	}
}
