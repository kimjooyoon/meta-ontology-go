package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSourceRelativeV3FreshNativeConstructionAndDeletedModelReplay(t *testing.T) {
	source, model := freshV3Fixture(t)
	root := t.TempDir()
	caller, evaluation := filepath.Join(root, "caller.json"), filepath.Join(root, "evaluation.json")
	if err := os.WriteFile(caller, []byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":83},"expected":{"Main":96}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	inputs := [8]int64{-203, -13, 0, 53, 69, 105, 34567, 9007199254740993}
	rows := make([]map[string]map[string]int64, 0, 8)
	for _, input := range inputs {
		output := input + 13
		if input < 67 {
			output = input / 4
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
		var out, diagnostic bytes.Buffer
		if code := run(args, &out, &diagnostic); code != exitOK {
			t.Fatal(code, diagnostic.String())
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
	if first.Evaluation.Runtime.FinitePassed != 8 || first.Evaluation.Runtime.FiniteTotal != 8 {
		t.Fatal("V3 finite native outputs differ")
	}
	calls := 0
	for _, preparation := range first.Construction.Initial.Preparations {
		if p := preparation.Generation.Report.BodyPaths; p != nil && p.ContractRanking != nil {
			calls += p.ContractRanking.Calls
		}
	}
	if calls != 1 {
		t.Fatal("V3 re-predicted during construction")
	}
	if err := os.Remove(model); err != nil {
		t.Fatal(err)
	}
	replay := call("body-construct", "--source", filepath.Join(out, "original.gooo"), "--construction", filepath.Join(out, "construction.json"), "--cases", evaluation, "--go-bin", goBin)
	if replay.GeneratedNow || !replay.Evaluation.ConstructionReplayed || replay.Evaluation.NewModelCalls != 0 ||
		replay.Evaluation.Runtime.FinitePassed != 8 || replay.Construction.Selected.GeneratedSHA256 != first.Construction.Selected.GeneratedSHA256 {
		t.Fatal("V3 deleted-model replay or generated identity differs")
	}
}
