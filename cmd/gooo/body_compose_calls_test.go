package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestBodyComposePureCallsEntryPolicyAndReplay(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "saved")
	base := []string{"body-compose", "--source", "../../examples/pure-activity-calls/main.gooo.fixture",
		"--cases", "../../examples/assembly-policy/cases.json"}
	args := append(base[:len(base):len(base)], "--entry", "Diagnose", "--assembly-policy",
		"../../examples/pure-activity-calls/policy.gooo.fixture", "--policy-activity", "Explain", "--out", directory, "--repeat", "2")
	var out, diagnostics bytes.Buffer
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 4 || len(result.RuntimeHistory) != 2 ||
		len(result.Composition.Steps[0].Generation.Report.RecordAssembly.Control.Decisions) != 4 {
		t.Fatal("entry/policy calls did not reach native execution", result.Runtime)
	}
	out.Reset()
	if code := run(append(base, "--composition", filepath.Join(directory, "composition.json")), &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.GeneratedNow || result.Runtime.FinitePassed != 4 {
		t.Fatal("saved call graph replay", err)
	}
}
