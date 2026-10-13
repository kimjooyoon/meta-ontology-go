package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestFilledHoleRejectsLocalImportAliasAtItsActualScope(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "scope94/app", Activity: "Total"},
		Packages: []packageruntime.PackageSpec{
			{Path: "scope94/helpers", Name: "helpers", Sources: []packageruntime.Source{{Filename: "helper.gooo", Content: `package helpers
namespace scope94_helpers
entity Integer id "urn:scope94:integer"
activity Add(Integer) -> Integer id "urn:scope94:add" computes "return input+29"
`}}},
			{Path: "scope94/app", Name: "app", Imports: []string{"scope94/helpers"}, Sources: []packageruntime.Source{{Filename: "total.gooo", Content: `package app
namespace scope94_app
import fees "scope94/helpers"
activity Total(Integer) -> Integer computes "let fees = input; return __GOOO_BODY_HOLE_amount__" assembling {
 source_fill intent "A same-scope local value makes the imported spelling unresolved." {
  hole "amount"
  candidate "input" { fill "amount" "input" }
  candidate "call" { fill "amount" "fees.Add(input)" }
 }
 case "6" -> "35"
}
`}}},
		}}
	before, _ := json.Marshal(manifest)
	r, err := RenameWorkspaceActivity(context.Background(), manifest, "scope94/helpers", "urn:scope94:add", "가산")
	if err == nil || !strings.Contains(err.Error(), `call name "fees" is shadowed by a local value`) ||
		len(r.Files) != 0 || r.Schema != "" || r.ProgramExecutions != 0 || r.ModelPredictions != 0 {
		t.Fatal("candidate escaped its actual lexical scope or returned a partial proposal", r, err)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("failed proposal mutated input")
	}
	raw, _ := json.Marshal(map[string]any{"outcome": "UNRESOLVED_LOCAL_ALIAS", "error": err.Error(),
		"partial_proposal_files": len(r.Files), "input_unchanged": true, "declared_candidates": 2, "holes": 1,
		"program_executions": r.ProgramExecutions, "model_predictions": r.ModelPredictions})
	t.Log("FILLED_HOLE_LOCAL_ALIAS_OBSERVATION", string(raw))
}
