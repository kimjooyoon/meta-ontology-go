package workspaceexecution

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestConstructionExternalRecordPlanRetainsExactTypedCases(t *testing.T) {
	source := `package records
namespace records
entity Input id "records://input" fields { field value id "records://input/value" type integer required one }
entity Output id "records://output" fields { field result id "records://output/result" type integer required one }
activity Build(Input) -> Output computes "return Output{result: __GOOO_BODY_HOLE_result__}"
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "example/records", Activity: "Build"},
		Packages: []packageruntime.PackageSpec{{Path: "example/records", Name: "records",
			Sources: []packageruntime.Source{{Filename: "records.gooo", Content: source}}}}}
	plan := bodycodegen.IRBodyFillPlan{Schema: "gooo/body-codegen-ir-fill-plan/v3-record", Intent: "Increment the exact integer field.",
		Holes: []bodycodegen.IRBodyFillHole{{ID: "result"}},
		Candidates: []bodycodegen.IRBodyFillCandidate{
			{ID: "add", Fills: map[string]string{"result": "input.value + 1"}},
			{ID: "identity", Fills: map[string]string{"result": "input.value"}}},
		ValueCases: []assemblyspec.ValueCase{{Inputs: `[{"value":9007199254740993}]`, Expected: `{"result":9007199254740994}`}}}
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1",
		Cases: []bodyexecution.CompositionCase{{
			Inputs:   map[string]json.RawMessage{"example/records:Build": json.RawMessage(`{"value":9007199254740993}`)},
			Expected: map[string]json.RawMessage{"example/records:Build": json.RawMessage(`{"result":9007199254740994}`)}}}}
	prior, err := ExecuteWorkspaceWithOptions(context.Background(), manifest, suite,
		ExecuteOptions{BodyFillPlans: map[string]bodycodegen.IRBodyFillPlan{"example/records:Build": plan}})
	if err != nil || prior.Runtime.FinitePassed != 1 {
		t.Fatal("record plan did not execute", err)
	}
	rows, err := ObserveConstruction(context.Background(), savedWorkspaceResult(t, prior))
	if err != nil || len(rows) != 2 {
		t.Fatal("record plan did not reconstruct", err, rows)
	}
	for i, row := range rows {
		if row.Profile != "external_fill" || row.Counts.Total != 1 || row.Counts.Matched != 1-i || row.Counts.Best != 1 {
			t.Fatal("record field count or integer rounding replaced whole cases", row)
		}
	}
}
