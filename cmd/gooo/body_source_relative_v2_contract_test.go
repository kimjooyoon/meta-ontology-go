package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestSourceRelativeV2DivisionCLIExportsThenChecksOneRanking(t *testing.T) {
	model, err := contractdecision.NewSourceRelativeV2([contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "v2.json")
	if err = os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	source := "../../examples/body-codegen/source-division-layout-cases.gooo.fixture"
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-context", "--model", name, "--activity", "Choose", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var before bodycodegen.TypedPathContextExport
	if err = json.Unmarshal(out.Bytes(), &before); err != nil {
		t.Fatal(err)
	}
	if before.Context == nil || before.ContractCases == nil || before.ModelCompatibility == nil ||
		before.ModelPredictions != 0 || before.CandidateTests != 0 || before.Context.FeatureVersion != contractdecision.SourceRelativeV2FeatureVersion ||
		before.Context.Schema != "gooo/compiler-source-relative-contract-model-context/v2" || before.ContractCases.Count != 6 || before.ModelCompatibility.Model.ResidentTensorBytes != 76136 {
		t.Fatal("V2 did not export original authored requirements without inference")
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"body-codegen", "--json", "--activity", "Choose", "--path-model", name, "--path-step-attempts", "1", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var result bodycodegen.Result
	if err = json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p == nil || p.FunctionalCompleteness != 100 || p.ContractRanking == nil || p.ContractRanking.Calls != 1 || p.Conditions.Passed != 3 || len(p.ContractProgress) < 2 || p.ContractRanking.Schema != "gooo/source-relative-requirement-path-ranking/v2" {
		t.Fatal("V2 division did not complete through one ranked finite search")
	}
}

func TestSourceDivisionWithoutModelRemainsDeterministic(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Choose", "../../examples/body-codegen/source-division-layout-cases.gooo.fixture"}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p == nil || p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 || p.Search.Selection.ModelCalls != 0 {
		t.Fatal("model-free division path differs")
	}
}
