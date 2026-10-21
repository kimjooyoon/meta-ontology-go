package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRecordedUseRetentionControlledCases(t *testing.T) {
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", "../../examples/recorded-use-origins/retention-cases.json",
		"../../examples/recorded-use-origins/retention-counts.workspace.json"}, &out, &diagnostics)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("retention cases", code, err, receipt.Error, diagnostics.String())
	}
	runtime := receipt.Result.Runtime
	if receipt.Decision != "PASS" || runtime.FinitePassed != 6 || runtime.FiniteTotal != 6 || runtime.ModelCalls != 0 {
		t.Fatal("retention case denominator or model scope", receipt.Decision, runtime.FinitePassed, runtime.FiniteTotal, runtime.ModelCalls)
	}
}

func TestRecordedUseRetentionKeepsSavedOriginScope(t *testing.T) {
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--origin-receipt", "../../examples/recorded-use-origins/full-producer-receipt.json.gz",
		"--origin-shape", "local-use-columns", "--origin-offset", "15", "../../examples/recorded-use-origins/retention.workspace.json"}, &out, &diagnostics)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal("saved retention", code, err, receipt.Error, diagnostics.String())
	}
	var measure struct {
		Receipt      string `json:"receipt_sha256"`
		Retained     int    `json:"retained_uses"`
		Dropped      int    `json:"dropped_uses"`
		Attempted    int    `json:"attempted_uses"`
		BasisPoints  int    `json:"retention_basis_points"`
		PageComplete bool   `json:"page_complete"`
		Status       string `json:"status"`
	}
	runtime := receipt.Result.Runtime
	if len(runtime.Traces) != 1 || len(runtime.Traces[0].Deliveries) != 1 {
		t.Fatal("one selected retention result required")
	}
	if err := json.Unmarshal(runtime.Traces[0].Deliveries[0].Actual, &measure); err != nil {
		t.Fatal(err)
	}
	origin := receipt.OriginInput
	if receipt.Decision != "OBSERVED" || measure.Receipt != origin.ReceiptSHA256 ||
		measure.Receipt != "sha256:a15f33ca09f99265aa9b838dcd4628a9dd74a8b77dcff93e1156241738a5441a" ||
		measure.Retained != origin.TotalRows ||
		measure.Retained != 32 || measure.Dropped != 0 || measure.Attempted != 32 || measure.BasisPoints != 10000 ||
		measure.PageComplete || measure.Status != "RECORDED_USE_RETENTION" || origin.ReturnedRows != 1 ||
		origin.ProducerExecutions != 0 || origin.ProducerModelCalls != 0 || runtime.ModelCalls != 0 {
		t.Fatal("saved retention changed original scope or overstated page coverage", measure, origin)
	}
}
