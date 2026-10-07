package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBodyComposeGoooPolicyRunsAndReplays(t *testing.T) {
	var out, diagnostics bytes.Buffer
	args := []string{"body-compose", "--source", "../../examples/body-codegen/record-field-assembly.gooo.fixture",
		"--cases", "../../examples/body-codegen/record-field-assembly-cases.json",
		"--assembly-policy", "../../examples/assembly-explainer/main.gooo.fixture", "--policy-activity", "Explain"}
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	r := result.Composition.Steps[0].Generation.Report.RecordAssembly
	if r.Control == nil || len(r.Control.Decisions) != 8 || r.Control.Decisions[0].Operation != "CONTINUE_CANDIDATES" ||
		result.Runtime.FinitePassed != 14 || result.Runtime.ModelCalls != 0 {
		t.Fatal("policy construction did not reach native execution", r, result.Runtime)
	}
	path := filepath.Join(t.TempDir(), "composition.json")
	raw, err := json.Marshal(result.Composition)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := run(append(args[:5:5], "--composition", path), &out, &diagnostics); code != exitOK {
		t.Fatal("embedded policy replay", code, diagnostics.String())
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.GeneratedNow || result.Runtime.FinitePassed != 14 {
		t.Fatal("saved policy program did not replay", err)
	}
}

func TestBodyComposePolicyFlagsArePairedAndGenerationOnly(t *testing.T) {
	base := []string{"body-compose", "--source", "source", "--cases", "cases"}
	for _, extra := range [][]string{
		{"--assembly-policy", "policy"}, {"--policy-activity", "Explain"},
		{"--assembly-policy", "policy", "--policy-activity", "Explain", "--composition", "saved"},
	} {
		var out, diagnostics bytes.Buffer
		if code := run(append(append([]string(nil), base...), extra...), &out, &diagnostics); code != exitUsage {
			t.Fatal("invalid policy options", extra, code, diagnostics.String())
		}
	}
}
