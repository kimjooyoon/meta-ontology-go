package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoooFaultOriginReaderUsesSavedFaults(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "fault-origins")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check", filepath.Join(root, "describe.gooo")}, &stdout, &stderr); code != exitOK {
		t.Fatal("source check", code, stderr.String())
	}
	runtime := savedFaultOriginRuntime(t, "repair-outputs/local-division-faults.json")
	raw, err := json.Marshal(runtime)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "runtime.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	args := []string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "faults", "--origin-activity-id", "urn:localdivision:divide", filepath.Join(root, "gooo.workspace.json")}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal("Gooo fault reader", code, stderr.String(), stdout.String())
	}
	if dir := os.Getenv("GOOO_READ_OBSERVATION_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "gooo-fault-reader.json"), stdout.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	assertGoooFaultDescriptions(t, receipt)
}

func assertGoooFaultDescriptions(t *testing.T, receipt packageExecutionReceipt) {
	t.Helper()
	if receipt.Decision != "OBSERVED" || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal("input-only reader result", receipt)
	}
	r, origins := receipt.Result.Runtime, receipt.OriginInput
	if len(r.Traces) != 2 || r.FinitePassed != 0 || r.FiniteTotal != 0 || r.ModelCalls != 0 || origins.TotalRows != 2 || origins.ProducerExecutions != 0 || origins.ProducerModelCalls != 0 {
		t.Fatal("reader scope", r.Outcomes, origins)
	}
	for i, want := range []struct{ form, view string }{{"LOCAL_UPDATE", "SELECTED_GOOO_RHS"}, {"EXPRESSION", "GO_PROJECTION"}} {
		if len(r.Traces[i].Deliveries) != 1 {
			t.Fatal("extra reader delivery")
		}
		delivery := r.Traces[i].Deliveries[0]
		var value struct {
			Form     string `json:"form"`
			View     string `json:"source_view"`
			Case     int    `json:"case_index"`
			Activity string `json:"activity_id"`
			Root     string `json:"root_activity_id"`
		}
		if err := json.Unmarshal(delivery.Actual, &value); err != nil || value.Form != want.form || value.View != want.view || value.Case != i+2 || value.Activity != "urn:localdivision:divide" || value.Root != "urn:localdivision:main" || delivery.Passed != nil || delivery.Expected != nil {
			t.Fatal("Gooo interpretation differs", err, value)
		}
	}
}
