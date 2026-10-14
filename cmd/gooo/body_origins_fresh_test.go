package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTypedOriginExample13013(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	root := filepath.Join(sourceRoot, "examples/typed-value-origins")
	call := func(name string, args ...string) []byte {
		t.Helper()
		var output, diagnostics bytes.Buffer
		if code := run(args, &output, &diagnostics); code != exitOK {
			t.Fatal(name, code, diagnostics.String(), output.String())
		}
		if dir := os.Getenv("GOOO_TYPED_OUTPUT"); dir != "" {
			if err := os.WriteFile(filepath.Join(dir, name+".json"), output.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
		}
		return append([]byte(nil), output.Bytes()...)
	}
	producer := call("producer13013", "package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json"))
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(producer, &receipt); err != nil || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 2 || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("native record-copy example", err, receipt)
	}
	var rows []map[string]json.RawMessage
	for i, activity := range []string{"Compute", "Select", "Select"} {
		args := []string{"body-origins", "--activity", activity, "--inputs-for", "typed/ancestry:Inputs"}
		if i == 2 {
			args = append(args, "--include-control")
		}
		args = append(args, filepath.Join(root, "main.gooo"))
		raw := call([]string{"compute-columns13013", "select-data13013", "select-control13013"}[i], args...)
		var input struct {
			Schema string                       `json:"schema"`
			Inputs []map[string]json.RawMessage `json:"inputs"`
		}
		if err := json.Unmarshal(raw, &input); err != nil || len(input.Inputs) != 1 {
			t.Fatal("column input", err)
		}
		rows = append(rows, input.Inputs[0])
	}
	input := struct {
		Schema string                       `json:"schema"`
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}{Schema: "gooo/body-composition-inputs/v1", Inputs: rows}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ancestry-inputs13013.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	consumer := call("consumer13013", "package", "execute", "--json", "--inputs", path, filepath.Join(root, "ancestry.workspace.json"))
	var observed packageExecutionReceipt
	if err := json.Unmarshal(consumer, &observed); err != nil || observed.Result == nil || observed.Decision != "OBSERVED" {
		t.Fatal("Gooo graph consumer", err, observed.Error)
	}
	runtime := observed.Result.Runtime
	if len(runtime.Traces) != 3 || runtime.ModelCalls != 0 || runtime.FiniteTotal != 0 {
		t.Fatal("consumer observation scope", runtime)
	}
	want := []string{"input0:urn:gooo:typed-origins:right;", "input0:urn:gooo:typed-origins:right;input1:;", "input0:urn:gooo:typed-origins:right;input1:;input2:;"}
	for i, trace := range runtime.Traces {
		var actual string
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || actual != want[i] {
			t.Fatal("Gooo-owned ancestry", i, actual, want[i])
		}
	}
	t.Log("actual new Gooo record-copy example2/2; Gooo graph traversal3/3 external comparisons, input-only OBSERVED finite0; model calls0")
}
