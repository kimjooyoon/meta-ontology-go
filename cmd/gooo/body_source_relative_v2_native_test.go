package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

// New zero-weight API regression; it never reuses the fitted quality model.
func TestSourceRelativeV2PublishedSDKNativeConstructionAndReplay(t *testing.T) {
	m, err := contractdecision.NewSourceRelativeV2([contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	model := filepath.Join(root, "zero-v2.json")
	if err = os.WriteFile(model, raw, 0600); err != nil {
		t.Fatal(err)
	}
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	base := "../../examples/body-codegen/source-division-layout-"
	call := func(args ...string) bodyConstructOutput {
		t.Helper()
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(code, out.String(), diagnostics.String())
		}
		var r bodyConstructOutput
		if err := json.Unmarshal(out.Bytes(), &r); err != nil {
			t.Fatal(err)
		}
		return r
	}
	out := filepath.Join(root, "construction")
	first := call("body-construct", "--source", base+"cases.gooo.fixture", "--entry", "Main", "--model", model, "--attempts", "4",
		"--construction-cases", base+"construction.json", "--cases", base+"evaluation.json", "--go-bin", goBin, "--out", out)
	if first.Evaluation.Runtime.FinitePassed != 8 || first.Evaluation.Runtime.FiniteTotal != 8 {
		t.Fatal("V2 native expectations differ")
	}
	calls := 0
	for _, step := range first.Construction.Initial.Preparations {
		p := step.Generation.Report.BodyPaths
		if p != nil && p.ContractRanking != nil {
			calls += p.ContractRanking.Calls
		}
	}
	if calls != 1 {
		t.Fatal("V2 session called the model after its initial ranking", calls)
	}
	if err = os.Remove(model); err != nil {
		t.Fatal(err)
	}
	replay := call("body-construct", "--source", filepath.Join(out, "original.gooo"), "--construction", filepath.Join(out, "construction.json"),
		"--cases", base+"evaluation.json", "--go-bin", goBin)
	if replay.GeneratedNow || !replay.Evaluation.ConstructionReplayed || replay.Evaluation.NewModelCalls != 0 ||
		replay.Evaluation.Runtime.FinitePassed != 8 || replay.Evaluation.Runtime.FiniteTotal != 8 ||
		replay.Construction.Selected.GeneratedSHA256 != first.Construction.Selected.GeneratedSHA256 {
		t.Fatal("V2 saved program identity or model-free replay differs")
	}
}
