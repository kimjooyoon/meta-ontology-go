package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Fresh synthetic native obligations, published with the source before CI runs.
// No original fitted artifacts or retained study inputs are loaded.
func TestSourceRelativeV4NativeConstructionAndDeletedModelReplayCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("new zero-weight native integration runs in CI")
	}
	source, model := freshV4Fixture(t)
	root := t.TempDir()
	caller, evaluation := filepath.Join(root, "caller.json"), filepath.Join(root, "evaluation.json")
	if err := os.WriteFile(caller, []byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":193},"expected":{"Main":216}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	inputs := [8]int64{-899, -10, 8, 145, 151, 291, 65537, 9007199254740993}
	rows := make([]map[string]map[string]int64, 0, len(inputs))
	for _, input := range inputs {
		output := input + 23
		if input < 149 {
			output = input / 9
		}
		rows = append(rows, map[string]map[string]int64{"inputs": {"Main": input}, "expected": {"Main": output}})
	}
	raw, err := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": rows})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evaluation, raw, 0600); err != nil {
		t.Fatal(err)
	}
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	call := func(args ...string) bodyConstructOutput {
		t.Helper()
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(code, diagnostics.String())
		}
		var result bodyConstructOutput
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	out := filepath.Join(root, "construction")
	first := call("body-construct", "--source", source, "--entry", "Main", "--model", model, "--attempts", "4",
		"--construction-cases", caller, "--cases", evaluation, "--go-bin", goBin, "--out", out)
	if first.Evaluation.Runtime.FinitePassed != len(inputs) || first.Evaluation.Runtime.FiniteTotal != len(inputs) {
		t.Fatal("V4 finite native output obligations differ")
	}
	calls := 0
	for _, preparation := range first.Construction.Initial.Preparations {
		if p := preparation.Generation.Report.BodyPaths; p != nil && p.ContractRanking != nil {
			calls += p.ContractRanking.Calls
		}
	}
	if calls != 1 {
		t.Fatal("V4 construction did not make exactly one prediction")
	}
	if err := os.Remove(model); err != nil {
		t.Fatal(err)
	}
	replay := call("body-construct", "--source", filepath.Join(out, "original.gooo"), "--construction",
		filepath.Join(out, "construction.json"), "--cases", evaluation, "--go-bin", goBin)
	if replay.GeneratedNow || !replay.Evaluation.ConstructionReplayed || replay.Evaluation.NewModelCalls != 0 ||
		replay.Evaluation.Runtime.FinitePassed != len(inputs) ||
		replay.Construction.Selected.GeneratedSHA256 != first.Construction.Selected.GeneratedSHA256 {
		t.Fatal("V4 saved replay required the model or changed generated output")
	}
}
