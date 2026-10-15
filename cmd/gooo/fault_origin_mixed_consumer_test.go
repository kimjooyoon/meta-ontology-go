package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoooFaultOriginReaderDistinguishesFieldUpdates(t *testing.T) {
	runtime := savedFaultOriginRuntime(t, "mixed-outputs/mixed-division-faults.json")
	raw, err := json.Marshal(runtime)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "mixed-runtime.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	args := []string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "faults", "../../examples/fault-origins/gooo.workspace.json"}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal("mixed Gooo reader", code, stderr.String(), stdout.String())
	}
	if dir := os.Getenv("GOOO_READ_OBSERVATION_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "gooo-mixed-fault-reader.json"), stdout.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal("mixed receipt", err)
	}
	r := receipt.Result.Runtime
	if len(r.Traces) != 2 || r.FiniteTotal != 0 || r.FinitePassed != 0 || r.ModelCalls != 0 || receipt.OriginInput.ProducerExecutions != 0 {
		t.Fatal("mixed consumer scope")
	}
	for i, form := range []string{"FIELD_UPDATE", "LOCAL_UPDATE"} {
		var actual map[string]any
		if err := json.Unmarshal(r.Traces[i].Deliveries[0].Actual, &actual); err != nil || actual["form"] != form || actual["case_index"] != float64(i+2) {
			t.Fatal("mixed form", err, actual)
		}
	}
}
