package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func cliDeclaredContractModel(t *testing.T) string {
	t.Helper()
	m, err := contractdecision.New([contractdecision.ParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "contract.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestDeclaredContractModelCLIContextAndCodegen(t *testing.T) {
	model := cliDeclaredContractModel(t)
	source := "../../examples/body-codegen/source-condition-cases.gooo.fixture"
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-context", "--model", model, "--activity", "Choose", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var before bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(out.Bytes(), &before); err != nil {
		t.Fatal(err)
	}
	if before.ModelPredictions != 0 || before.CandidateTests != 0 || before.ContractCases == nil || before.ContractCases.Count != 7 {
		t.Fatal("CLI did not export the authored examples")
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"body-codegen", "--json", "--activity", "Choose", "--path-model", model, "--path-step-attempts", "1", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p.FunctionalCompleteness != 100 || p.ContractRanking == nil || p.ContractRanking.Calls != 1 || p.ContractRanking.CaseSHA != before.ContractCases.CaseSHA || len(p.ContractProgress) < 2 || p.Conditions.Passed != 3 {
		t.Fatal("CLI did not perform one declared-case ranking")
	}
}

func TestDeclaredContractModelCLINativeConstructionAndModelFreeReplay(t *testing.T) {
	model := cliDeclaredContractModel(t)
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
	if result.Evaluation.Runtime.FinitePassed != 11 || p.ContractRanking == nil || p.ContractRanking.Calls != 1 || p.ModelRetention.ModelSchema != contractdecision.Schema {
		t.Fatal("native construction lost the declared-case decision")
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
		t.Fatal("saved native construction required another model call", err)
	}
}
