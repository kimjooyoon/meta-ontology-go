package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func newIdentityRenameWorkspace() packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "rename/app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "rename/left", Name: "left", Sources: []packageruntime.Source{{Filename: "left.gooo", Content: `package left
namespace rename_left
entity Integer id "urn:identityrename:integer"
entity Boolean id "urn:identityrename:boolean"
activity Check(Integer,Integer) -> Boolean id "urn:identityrename:left:check" computes "return input0 == input1"
activity Local(Integer) -> Boolean computes "return Check(input,input)"
`}}},
			{Path: "rename/right", Name: "right", Sources: []packageruntime.Source{{Filename: "right.gooo", Content: `package right
namespace rename_right
entity Integer id "urn:identityrename:integer"
entity Boolean id "urn:identityrename:boolean"
activity Check(Integer,Integer) -> Boolean id "urn:identityrename:right:check" computes "return input0 != input1"
`}}},
			{Path: "rename/app", Name: "app", Imports: []string{"rename/left", "rename/right"}, Sources: []packageruntime.Source{
				{Filename: "main.gooo", Content: `package app
namespace rename_app
import rules "rename/left"
activity Main(Integer) -> Integer computes "let note = \"Check()\"; if note == \"\" { return input }; if rules.\u0043heck(input,input) { return input+9 }; return input"
activity Candidate(Integer) -> Integer computes "return input-3" assembling {
 choice "direction" operand_order at "0" intent "subtract three"
 criteria { predicate "check" activity "rules.Check" input "[5]" holdout_input "[6]" }
 attempts "2"
}
activity Accept(Boolean) -> Boolean computes "return input"
bind rules.Check.result -> Accept.input
`},
				{Filename: "dormant.gooo", Content: "package app\nnamespace rename_app\nimport rules \"rename/left\"\nactivity Draft(Integer) -> Integer computes `if rules.Check(input,input) { return input+10 }; return 0`\n"},
				{Filename: "other.gooo", Content: "package app\nnamespace rename_app\nimport rules \"rename/right\"\nactivity Other(Integer) -> Boolean computes `return rules.Check(input,input)`\n"},
			}},
		}}
}

func TestWorkspaceIdentityRenameIncludesDormantAndAliasScopedReferences(t *testing.T) {
	manifest := newIdentityRenameWorkspace()
	original, _ := json.Marshal(manifest)
	r, err := RenameWorkspaceActivity(context.Background(), manifest, "rename/left", "urn:identityrename:left:check", "조건판정")
	if err != nil {
		t.Fatal(err)
	}
	if r.ResolvedReferences != 5 || r.UpdatedReferences != 5 || len(r.Files) != 3 || r.IdentityPinned || r.EntryUpdated ||
		r.ActivityID != "urn:identityrename:left:check" || r.ModelPredictions != 0 || r.ProgramExecutions != 0 {
		t.Fatal("workspace source coverage or stable identity changed", r)
	}
	checkIdentityRenameSourceLocality(t, r, manifest)
	afterOriginal, _ := json.Marshal(manifest)
	if string(afterOriginal) != string(original) {
		t.Fatal("source proposal mutated the input manifest")
	}
	raw, _ := json.Marshal(map[string]any{"activity_id": r.ActivityID, "resolved_references": r.ResolvedReferences,
		"updated_references": r.UpdatedReferences, "changed_files": len(r.Files), "source_files": 5, "packages": 3,
		"input_image_digest": r.InputImageDigest, "output_image_digest": r.OutputImageDigest,
		"dormant_consumers_updated": true, "file_local_alias_other_target_preserved": true,
		"program_executions": r.ProgramExecutions, "model_predictions": r.ModelPredictions})
	t.Log("WORKSPACE_IDENTITY_RENAME_OBSERVATION", string(raw))
}

func checkIdentityRenameSourceLocality(t *testing.T, r WorkspaceActivityRename, original packageruntime.Manifest) {
	t.Helper()
	left := r.Manifest.Packages[0].Sources[0].Content
	main := r.Manifest.Packages[2].Sources[0].Content
	dormant := r.Manifest.Packages[2].Sources[1].Content
	if !strings.Contains(left, `activity 조건판정(Integer,Integer) -> Boolean id "urn:identityrename:left:check"`) ||
		!strings.Contains(left, "return 조건판정(input,input)") || !strings.Contains(main, "rules.조건판정(input,input)") ||
		!strings.Contains(main, `activity "rules.조건판정"`) || !strings.Contains(main, "bind rules.조건판정.result -> Accept.input") ||
		!strings.Contains(main, `\"Check()\"`) || !strings.Contains(dormant, "rules.조건판정(input,input)") {
		t.Fatal("declaration, escaped/direct calls, criterion, binding or literal locality changed")
	}
	if r.Manifest.Packages[1].Sources[0].Content != original.Packages[1].Sources[0].Content ||
		r.Manifest.Packages[2].Sources[2].Content != original.Packages[2].Sources[2].Content {
		t.Fatal("same-name activity or a file-local alias to another package changed")
	}
}
