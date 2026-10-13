package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func qualifiedOracleRenameWorkspace() packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "module/questions", Activity: "Fee"},
		Packages: []packageruntime.PackageSpec{
			{Path: "module/reference", Name: "reference", Sources: []packageruntime.Source{{Filename: "reference.gooo", Content: `package reference
namespace module_reference
entity Integer id "urn:moduleoracle:integer"
activity Reference(Integer) -> Integer id "urn:moduleoracle:fixed" computes "return 17-input"
`}}},
			{Path: "module/questions", Name: "questions", Imports: []string{"module/reference"}, Sources: []packageruntime.Source{{Filename: "questions.gooo", Content: `package questions
namespace module_questions
import math "module/reference"
activity Fee(Integer) -> Integer computes "return input-17" assembling {
 choice "direction" operand_order at "0" intent "deduct input from seventeen"
 case "17" -> "0"
 observe { input "22" oracle "math.Reference" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}
`}}},
		}}
}

func TestQualifiedOracleRenameFeedsFiniteWorkspaceConstruction(t *testing.T) {
	ctx := context.Background()
	manifest := qualifiedOracleRenameWorkspace()
	rename, err := RenameWorkspaceActivity(ctx, manifest, "module/reference", "urn:moduleoracle:fixed", "기준계산")
	if err != nil {
		t.Fatal(err)
	}
	if rename.UpdatedReferences != 1 || rename.ResolvedReferences != 1 || len(rename.Files) != 2 ||
		!strings.Contains(rename.Manifest.Packages[1].Sources[0].Content, `oracle "math.기준계산"`) ||
		rename.ProgramExecutions != 0 || rename.ModelPredictions != 0 {
		t.Fatal("qualified source oracle was not retained by package and ID")
	}
	program, err := Prepare(rename.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	if program.PureCalls == nil || len(program.PureCalls.Sites) != 1 || program.PureCalls.Sites[0].Surface != "observe:oracle" {
		t.Fatal("imported oracle is missing from the prepared dependency closure")
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(ctx, "workspace.gooo", []byte(program.Source), program.Entry.LoweredName)
	if err != nil {
		t.Fatal(err)
	}
	checkQualifiedOracleConstruction(t, rename, result)
}

func checkQualifiedOracleConstruction(t *testing.T, rename WorkspaceActivityRename, result bodycodegen.Result) {
	t.Helper()
	if result.Report.BodyPaths == nil || result.Report.BodyPaths.Observation == nil {
		t.Fatal("finite source question has no observation")
	}
	o := result.Report.BodyPaths.Observation
	if o.Status != "ONE_SURVIVING_CANDIDATE" || o.OracleActivityID != "urn:moduleoracle:fixed" ||
		len(o.Rounds) != 2 || o.Rounds[0].Observation == nil || o.Rounds[1].Observation != nil ||
		o.OracleEvaluations != 2 || o.EffectiveCases != 2 {
		t.Fatal("qualified oracle did not answer the original case and one source question")
	}
	raw, _ := json.Marshal(map[string]any{"oracle_activity_id": o.OracleActivityID, "source_references": rename.UpdatedReferences,
		"reference_total": rename.ResolvedReferences, "changed_files": len(rename.Files), "ranking_rounds": len(o.Rounds),
		"answered_source_questions": 1,
		"oracle_evaluations":        o.OracleEvaluations, "effective_cases": o.EffectiveCases, "status": o.Status,
		"rename_program_executions": rename.ProgramExecutions, "rename_model_predictions": rename.ModelPredictions,
		"native_processes": 0, "model_predictions": 0})
	t.Log("QUALIFIED_ORACLE_RENAME_CONSTRUCTION_OBSERVATION", string(raw))
}
