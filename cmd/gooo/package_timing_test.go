package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageTimingIsSeparateAndReplayObservesCurrentIntervals(t *testing.T) {
	root := "../../examples/assembly-explainer/"
	var out, diagnostics bytes.Buffer
	args := []string{"package", "execute", "--json", "--cases", root + "cases.json", root + "gooo.workspace.json"}
	if code := run(args, &out, &diagnostics); code != exitOK || diagnostics.Len() != 0 {
		t.Fatal("default execution gained a diagnostic", code, diagnostics.String())
	}
	saved := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(saved, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	args = []string{"package", "replay", "--timing", "--json", "--receipt", saved,
		"--cases", root + "cases.json", root + "gooo.workspace.json"}
	out.Reset()
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var receipt packageExecutionReceipt
	var timing packageWallDiagnostic
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(diagnostics.Bytes(), &timing); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 10 ||
		timing.Decision != receipt.Decision || timing.Operation != "replay" || timing.Wall.Status != "OBSERVED" ||
		timing.SourceSHA256 != receipt.Result.SourceSHA256 || timing.CasesDigest != receipt.CasesDigest ||
		timing.ManifestDigest != receipt.ManifestDigest || len(timing.Wall.Phases) < 10 {
		t.Fatal("diagnostic lost current source/inputs or changed execution", timing)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["wall"] != nil || envelope["timing"] != nil || bytes.Contains(out.Bytes(), []byte("package-wall-diagnostic")) {
		t.Fatal("optional view entered the execution receipt")
	}
}
