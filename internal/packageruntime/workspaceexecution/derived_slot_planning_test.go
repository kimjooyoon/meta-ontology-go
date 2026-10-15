package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestDerivedSlotPlanningRetainsRenamedHelperAndHeldInput(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "derive129/app", Activity: "Builder"},
		Packages: []packageruntime.PackageSpec{
			{Path: "derive129/helpers", Name: "helpers", Sources: []packageruntime.Source{{Filename: "base.gooo", Content: `package helpers
namespace derive129_helpers
entity Integer id "urn:derive129:integer"
activity Base(Integer) -> Integer id "urn:derive129:base" computes "return input+83"
`}}},
			{Path: "derive129/app", Name: "app", Imports: []string{"derive129/helpers"}, Sources: []packageruntime.Source{
				{Filename: "builder.gooo", Content: `package app
namespace derive129_app
import fees "derive129/helpers"
activity Builder(Integer) -> Integer id "urn:derive129:builder" computes "if __GOOO_BODY_HOLE_condition__ { return fees.Base(__GOOO_BODY_HOLE_value__) }; return 0" assembling {
 source_fill intent "Derive slots through the dedicated fill API with a separate reserved input." {
  hole "condition"
  hole "value"
  derive assignments max_candidates "4" {
   hole "condition" grammar "integer-predicate/v1" max_expressions "2"
   hole "value" grammar "integer-offset-constant/v1" max_expressions "2"
  }
 }
 case "4" -> "87"
 holdout_case "-3" -> "80"
}
`},
				{Filename: "dormant.gooo", Content: `package app
namespace derive129_app
import fees "derive129/helpers"
activity Bare(Integer) -> Integer computes "return fees.Base(input)-1"
`},
			}},
		}}
	before, _ := json.Marshal(manifest)
	ctx := context.Background()
	rename, err := RenameWorkspaceActivity(ctx, manifest, "derive129/helpers", "urn:derive129:base", "기본값")
	if err != nil {
		t.Fatal(err)
	}
	if rename.ResolvedReferences != 2 || rename.UpdatedReferences != 2 || len(rename.Files) != 3 ||
		rename.IdentityPinned || rename.EntryUpdated || rename.ProgramExecutions != 0 || rename.ModelPredictions != 0 {
		t.Fatal("derived template or dormant helper reference changed", rename)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) || !strings.Contains(rename.Manifest.Packages[1].Sources[0].Content,
		"fees.기본값(__GOOO_BODY_HOLE_value__)") {
		t.Fatal("derived proposal changed inputs or lost original source range")
	}
	program, err := Prepare(rename.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := bodycodegen.SourceAssembly(ctx, "derived-slots-workspace.gooo", []byte(program.Source), program.Entry.LoweredName)
	if err != nil {
		t.Fatal(err)
	}
	result, err := bodycodegen.GenerateWithSourceIRBodyFill(ctx, "derived-slots-workspace.gooo", []byte(program.Source),
		program.Entry.LoweredName, spec, "", "", bodycodegen.IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	fill := result.Report.BodyFill
	if fill == nil || fill.CandidateGeneration == nil || fill.TestCasesPassed != 1 || fill.HoldoutCasesPassed != 1 ||
		fill.LocalModelPredictions != 0 || fill.ExternalProviderCalls != 0 {
		t.Fatal("renamed deterministic derived body did not satisfy separate cases", fill)
	}
	raw, _ := json.Marshal(map[string]any{"activity_id": rename.ActivityID, "references": rename.UpdatedReferences,
		"reference_total": rename.ResolvedReferences, "changed_files": len(rename.Files), "source_files": 3,
		"input_unchanged": true, "rename_program_executions": rename.ProgramExecutions,
		"rename_model_predictions": rename.ModelPredictions, "native_processes": 0,
		"construction_model_predictions": fill.LocalModelPredictions, "external_provider_calls": fill.ExternalProviderCalls,
		"selected_cases_passed": fill.TestCasesPassed, "selected_cases_total": 1,
		"held_cases_passed": fill.HoldoutCasesPassed, "held_cases_total": 1,
		"selected_candidate": fill.SelectedCandidateID, "candidate_generation": fill.CandidateGeneration})
	t.Log("DERIVED_SLOT_FILL_API_OBSERVATION", string(raw))
}
