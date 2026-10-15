package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestNamedSignatureExportCLI(t *testing.T) {
	args := []string{"--activity", "Difference", "--inputs-for", "source/control:DescribeInputs",
		"--inputs-shape", "named-input-declaration-columns", "../../examples/named-inputs/main.gooo"}
	options, err := parseBodyOriginOptions(args)
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(options.file)
	if err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := writeTypedBodyOrigins(options, source, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var envelope struct {
		Inputs []map[string]struct{ Names []string }
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil || len(envelope.Inputs) != 1 {
		t.Fatal(err)
	}
	names := envelope.Inputs[0]["source/control:DescribeInputs"].Names
	if len(names) != 2 || names[0] != "minuend" || names[1] != "subtrahend" {
		t.Fatal(names)
	}
	for _, extra := range [][]string{{"--field", "urn:field"}, {"--include-control"}, {"--format", "text"}, {"--fill-candidates"}} {
		if _, err := parseBodyOriginOptions(append(append([]string{}, args...), extra...)); err == nil {
			t.Fatal("incompatible signature options", extra)
		}
	}
}
