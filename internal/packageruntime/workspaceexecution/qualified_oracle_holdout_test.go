package workspaceexecution

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func importedHoldoutQuestionWorkspace() packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "questions/price", Activity: "Price"},
		Packages: []packageruntime.PackageSpec{
			{Path: "questions/rates", Name: "rates", Sources: []packageruntime.Source{{Filename: "rates.gooo", Content: `package rates
namespace question_rates
entity Integer id "urn:question31:integer"
activity Expected(Integer) -> Integer id "urn:question31:expected" computes "return 31-input"
`}}},
			{Path: "questions/price", Name: "price", Imports: []string{"questions/rates"}, Sources: []packageruntime.Source{{Filename: "price.gooo", Content: `package price
namespace question_price
import rates "questions/rates"
activity Price(Integer) -> Integer computes "return input-31" assembling {
 choice "direction" operand_order at "0" intent "deduct input from thirty one"
 case "31" -> "0"
 holdout_case "42" -> "-11"
 observe { input "-8" oracle "rates.Expected" max_candidates "2" max_rounds "1" reuse_outputs "true" resolve_unique "true" }
 attempts "2"
}
`}}},
		}}
}

func TestImportedOracleQuestionKeepsHoldoutSeparateAfterWorkspaceRename(t *testing.T) {
	ctx := context.Background()
	rename, err := RenameWorkspaceActivity(ctx, importedHoldoutQuestionWorkspace(), "questions/rates", "urn:question31:expected", "기준값")
	if err != nil {
		t.Fatal(err)
	}
	program, err := Prepare(rename.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(ctx, "workspace.gooo", []byte(program.Source), program.Entry.LoweredName)
	if err != nil {
		t.Fatal(err)
	}
	inspectImportedOracleHoldout(t, rename, result)
}

func inspectImportedOracleHoldout(t *testing.T, rename WorkspaceActivityRename, result bodycodegen.Result) {
	t.Helper()
	p := result.Report.BodyPaths
	if p == nil || p.Observation == nil || p.Holdout == nil {
		t.Fatal("question or separate evaluation observation missing")
	}
	o := p.Observation
	answerCount := 0
	for _, round := range o.Rounds {
		if round.Observation != nil {
			answerCount++
		}
	}
	raw, _ := json.Marshal(map[string]any{"oracle": o, "holdout": p.Holdout,
		"activity_id": rename.ActivityID, "references": rename.UpdatedReferences, "reference_total": rename.ResolvedReferences,
		"changed_files": len(rename.Files), "answered_source_questions": answerCount,
		"rename_program_executions": rename.ProgramExecutions, "rename_model_predictions": rename.ModelPredictions,
		"native_processes": 0, "model_predictions": 0})
	t.Log("IMPORTED_ORACLE_HOLDOUT_OBSERVATION", string(raw))
	if rename.UpdatedReferences != 1 || rename.ResolvedReferences != 1 || len(rename.Files) != 2 ||
		o.Status != "ONE_SURVIVING_CANDIDATE" || o.OracleActivityID != "urn:question31:expected" ||
		len(o.Rounds) != 2 || answerCount != 1 || o.OracleEvaluations != 2 || o.EffectiveCases != 2 ||
		o.Rounds[0].Observation.Input != -8 || o.Rounds[0].Observation.Expected != 39 ||
		p.Holdout.Total != 1 || p.Holdout.Passed != 1 {
		t.Fatal("imported source answer, final ranking or separate evaluation changed")
	}
}
