package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBodyConstructSourceConditionsAndSavedNativeReplay(t *testing.T) {
	root := filepath.Join(t.TempDir(), "conditions")
	base := "../../examples/body-codegen/source-condition-"
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	args := []string{"body-construct", "--source", base + "cases.gooo.fixture", "--entry", "Main",
		"--construction-cases", base + "construction-cases.json", "--cases", base + "evaluation-cases.json",
		"--attempts", "8", "--go-bin", goBin, "--out", root}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyConstructOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Evaluation.Runtime.FinitePassed != 11 {
		t.Fatal(result, err)
	}
	p := result.Construction.Initial.Preparations[0].Generation.Report.BodyPaths
	if p.Conditions == nil || p.Conditions.Passed != 3 || p.Search.Selection.ModelCalls != 0 {
		t.Fatal("native construction lost condition observations", p)
	}
	stdout.Reset()
	stderr.Reset()
	args = []string{"body-construct", "--source", filepath.Join(root, "original.gooo"),
		"--construction", filepath.Join(root, "construction.json"), "--cases", base + "evaluation-cases.json", "--go-bin", goBin}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.GeneratedNow ||
		!result.Evaluation.ConstructionReplayed || result.Evaluation.NewModelCalls != 0 || result.Evaluation.Runtime.FinitePassed != 11 {
		t.Fatal("condition native replay differs", err, result)
	}
}
