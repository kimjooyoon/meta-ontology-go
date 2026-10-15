package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestQualifiedRecordSourceFillUsesActualSlotScopeAndSeparateNativeCases(t *testing.T) {
	manifest := constructorFill687Manifest()
	before, _ := json.Marshal(manifest)
	wrong := constructorFill687Manifest()
	wrong.Packages[1].Sources[0].Content = strings.ReplaceAll(wrong.Packages[1].Sources[0].Content,
		"return __GOOO_BODY_HOLE_result__", "let 資料 = input.value; return __GOOO_BODY_HOLE_result__")
	partial, err := Prepare(wrong)
	if err == nil || !strings.Contains(err.Error(), "shadowed by a local value") || partial.Schema != "" {
		t.Fatal("source candidate ignored a local value at its actual hole", partial, err)
	}

	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"constructor687/app:Add":{"value":-6,"note":"negative"}},"expected":{"constructor687/app:Add":{"value":-6,"note":"negative"}}},
 {"inputs":{"constructor687/app:Add":{"value":5,"note":"native"}},"expected":{"constructor687/app:Add":{"value":8,"note":"資料.Value{preview}"}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	plan := bodycodegen.IRBodyFillPlan{Schema: "gooo/body-codegen-ir-fill-plan/v3-record", Intent: "Construct the imported value.",
		Holes: []bodycodegen.IRBodyFillHole{{ID: "result"}}, Candidates: []bodycodegen.IRBodyFillCandidate{
			{ID: "up", Fills: map[string]string{"result": `資料.Value{value: input.value + 3, note: "資料.Value{preview}"}`}},
			{ID: "down", Fills: map[string]string{"result": `資料.Value{value: input.value - 3, note: "資料.Value{preview}"}`}},
		}, ValueCases: []assemblyspec.ValueCase{{Inputs: `[{"value":18,"note":"select"}]`, Expected: `{"value":21,"note":"資料.Value{preview}"}`}}}
	planBefore, _ := json.Marshal(plan)
	external := wrong
	external.Packages = append([]packageruntime.PackageSpec(nil), wrong.Packages...)
	external.Packages[1].Sources = append([]packageruntime.Source(nil), wrong.Packages[1].Sources...)
	external.Packages[1].Sources[0].Content = strings.Split(wrong.Packages[1].Sources[0].Content, " assembling {")[0] + "\n"
	rejected, err := ExecuteWorkspaceWithOptions(context.Background(), external, suite,
		ExecuteOptions{BodyFillPlans: map[string]bodycodegen.IRBodyFillPlan{"constructor687/app:Add": plan}})
	if err == nil || !strings.Contains(err.Error(), "shadowed by a local value") || rejected.Schema != "" {
		t.Fatal("external candidate ignored the source hole's local scope", rejected, err)
	}
	planAfter, _ := json.Marshal(plan)
	if string(planBefore) != string(planAfter) {
		t.Fatal("constructor resolution rewrote the caller's plan")
	}

	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.BodyFills) != 1 || result.Runtime.FinitePassed != 2 || result.Runtime.FiniteTotal != 2 || !result.Runtime.RuntimeReplayed || result.Runtime.ModelCalls != 0 {
		t.Fatal("source fill did not reach separate native cases", result.Runtime)
	}
	fill := result.BodyFills[0].Generation.Report.BodyFill
	if fill == nil || fill.SelectedCandidateID != "up" || fill.TestCasesPassed != 1 || fill.TestCasesTotal != 1 ||
		fill.HoldoutCasesPassed != 1 || fill.HoldoutCasesTotal != 1 || fill.LocalModelPredictions != 0 || fill.ExternalProviderCalls != 0 {
		t.Fatal("candidate selection or held-out observation differs", fill)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("source constructor lowering rewrote the original program")
	}
	observation, _ := json.Marshal(map[string]any{"source_fill_candidates": len(fill.CandidateScores), "selected_candidate": fill.SelectedCandidateID,
		"selection_passed": fill.TestCasesPassed, "selection_total": fill.TestCasesTotal,
		"holdout_passed": fill.HoldoutCasesPassed, "holdout_total": fill.HoldoutCasesTotal,
		"native_input_cases": len(suite.Cases), "native_passed": result.Runtime.FinitePassed, "native_total": result.Runtime.FiniteTotal,
		"source_and_external_shadow_declines": 2, "unrelated_block_shadow_preserved": true,
		"partial_results_accepted": 0, "source_and_external_plan_unchanged": true,
		"model_calls":      result.Runtime.ModelCalls + fill.LocalModelPredictions + fill.ExternalProviderCalls,
		"runtime_replayed": result.Runtime.RuntimeReplayed, "source_sha256": result.SourceSHA256})
	t.Log("GOOO_QUALIFIED_RECORD_FILL_OBSERVATION=" + string(observation))
}

func constructorFill687Manifest() packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "constructor687/app", Activity: "Add"},
		Packages: []packageruntime.PackageSpec{
			{Path: "constructor687/value", Name: "value", Sources: []packageruntime.Source{{Filename: "value.gooo", Content: `package value
namespace constructor687_value
entity Value id "urn:constructor687:value" fields {
 field value id "urn:constructor687:number" type "urn:gooo:type:integer" required one
 field note id "urn:constructor687:note" type "urn:gooo:type:string" required one
}
`}}},
			{Path: "constructor687/app", Name: "app", Imports: []string{"constructor687/value"}, Sources: []packageruntime.Source{{Filename: "구성.gooo", Content: `package app
namespace constructor687_app
import 資料 "constructor687/value"
activity Add(資料.Value) -> 資料.Value id "urn:constructor687:add" computes "if input.value < 0 { let 資料 = input.value; return Value{value: 資料, note: input.note} }; return __GOOO_BODY_HOLE_result__" assembling {
 source_fill intent "Construct the imported value with the declared increment." {
  hole "result"
  candidate "up" { fill "result" "資料.Value{value: input.value + 3, note: \"資料.Value{preview}\"}" }
  candidate "down" { fill "result" "資料.Value{value: input.value - 3, note: \"資料.Value{preview}\"}" }
 }
 value_case "[{\"value\":18,\"note\":\"select\"}]" -> "{\"value\":21,\"note\":\"資料.Value{preview}\"}"
 holdout_value_case "[{\"value\":41,\"note\":\"held\"}]" -> "{\"value\":44,\"note\":\"資料.Value{preview}\"}"
}
`}}},
		}}
}
