package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNamedInputReaderFromCompilerFacts(t *testing.T) {
	var rows []map[string]json.RawMessage
	for _, activity := range []string{"Credit", "Net"} {
		raw := namedReaderCommand(t, "export-"+activity, "body-origins", "--activity", activity,
			"--inputs-for", "source/control:DescribeInputs", "--inputs-shape", "named-input-declaration-columns",
			"../../examples/named-inputs/main.gooo")
		var envelope struct{ Inputs []map[string]json.RawMessage }
		if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.Inputs) != 1 {
			t.Fatal(err)
		}
		rows = append(rows, envelope.Inputs[0])
	}
	raw, err := json.Marshal(struct {
		Schema string                       `json:"schema"`
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}{"gooo/body-composition-inputs/v1", rows})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "inputs.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	receipt := namedReaderCommand(t, "native-receipt", "package", "execute", "--json", "--inputs", path,
		"../../examples/source-control-reader/named-inputs.workspace.json")
	assertNamedReaderResults(t, receipt)
}

func namedReaderCommand(t *testing.T, name string, args ...string) []byte {
	t.Helper()
	var out, diagnostics bytes.Buffer
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(name, code, diagnostics.String(), out.String())
	}
	if dir := os.Getenv("GOOO_NAMED_READER_RECEIPT"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, name+".json"), out.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return out.Bytes()
}
