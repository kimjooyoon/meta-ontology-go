package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func cliConditionModel(t *testing.T) string {
	t.Helper()
	m, err := conditiondecision.New([conditiondecision.ParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	raw = append([]byte(strings.Repeat(" ", 70<<10)), raw...)
	name := filepath.Join(t.TempDir(), "condition.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestConditionModelCLIContextAndCodegenUseSameFeatures(t *testing.T) {
	model := cliConditionModel(t)
	source := "../../examples/body-codegen/source-condition-cases.gooo.fixture"
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-context", "--model", model, "--activity", "Choose", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var before bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(out.Bytes(), &before); err != nil {
		t.Fatal(err)
	}
	if before.ModelPredictions != 0 || before.CandidateTests != 0 || before.ModelCompatibility.Status != "READY_FOR_RANKING" || len(before.Inputs) != 3 || before.Inputs[0].Features == nil {
		t.Fatal("CLI lost typed condition arrays")
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"body-codegen", "--json", "--activity", "Choose", "--path-model", model, "--path-step-attempts", "1", "--path-feedback-rounds", "2", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p.Conditions == nil || p.Conditions.Passed != 3 || p.FunctionalCompleteness != 100 || p.Search.Selection.ModelCalls != 1 || len(p.ConditionProgress) == 0 {
		t.Fatal("CLI did not use condition model", p)
	}
	for i, input := range before.Inputs {
		if input.InputSHA != "sha256:"+p.ConditionProgress[0].Ranking.FeatureSHA[i] {
			t.Fatal("CLI preflight differs from inference")
		}
	}
}

func TestConditionModelCLIConstructAndSavedNativeReplay(t *testing.T) {
	model := cliConditionModel(t)
	root := filepath.Join(t.TempDir(), "construction")
	base := "../../examples/body-codegen/source-condition-"
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	args := []string{"body-construct", "--source", base + "cases.gooo.fixture", "--entry", "Main", "--model", model,
		"--construction-cases", base + "construction-cases.json", "--cases", base + "evaluation-cases.json", "--attempts", "8", "--go-bin", goBin, "--out", root}
	var out, diagnostics bytes.Buffer
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var result bodyConstructOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Construction.Initial.Preparations[0].Generation.Report.BodyPaths
	if result.Evaluation.Runtime.FinitePassed != 11 || p.Conditions.Passed != 3 || p.Search.Selection.ModelCalls != 1 || len(p.ConditionProgress) == 0 || p.ModelRetention.ArtifactSHA256 == "" {
		t.Fatal("native construction lost own model", result)
	}
	if err := os.Remove(model); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostics.Reset()
	args = []string{"body-construct", "--source", filepath.Join(root, "original.gooo"), "--construction", filepath.Join(root, "construction.json"), "--cases", base + "evaluation-cases.json", "--go-bin", goBin}
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.GeneratedNow || !result.Evaluation.ConstructionReplayed || result.Evaluation.NewModelCalls != 0 || result.Evaluation.Runtime.FinitePassed != 11 {
		t.Fatal("native saved replay required model or changed answers", err, result)
	}
}
