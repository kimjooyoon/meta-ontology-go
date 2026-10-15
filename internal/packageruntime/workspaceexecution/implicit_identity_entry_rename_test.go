package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestImplicitWorkspaceEntryRenamePinsInventoryIdentity(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "dispatch/core", Activity: "Sum"},
		Packages: []packageruntime.PackageSpec{{Path: "dispatch/core", Name: "core", Sources: []packageruntime.Source{
			{Filename: "sum.gooo", Content: "package core\nnamespace dispatch_core\nentity Integer id \"urn:dispatch:integer\"\nactivity Sum(Integer) -> Integer computes `return input+4`\n"},
			{Filename: "carry.gooo", Content: "package core\nnamespace dispatch_core\nactivity Carry(Integer) -> Integer computes `return Sum(input)+2`\n"},
		}}}}
	before, _ := json.Marshal(manifest)
	id, err := workspaceActivityIdentity("dispatch_core", "Sum", "")
	if err != nil {
		t.Fatal(err)
	}
	r, err := RenameWorkspaceActivity(context.Background(), manifest, "dispatch/core", id, "합계")
	if err != nil {
		t.Fatal(err)
	}
	if !r.IdentityPinned || !r.EntryUpdated || r.Manifest.Entry.Activity != "합계" || r.ActivityID != id ||
		r.ResolvedReferences != 1 || r.UpdatedReferences != 1 || len(r.Files) != 2 ||
		!strings.Contains(r.Manifest.Packages[0].Sources[0].Content, `activity 합계(Integer) -> Integer id "`+id+`"`) ||
		!strings.Contains(r.Manifest.Packages[0].Sources[1].Content, "return 합계(input)+2") {
		t.Fatal("implicit ID, workspace entry or dormant consumer was not retained")
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) || r.ProgramExecutions != 0 || r.ModelPredictions != 0 {
		t.Fatal("proposal mutated or executed input")
	}
	raw, _ := json.Marshal(map[string]any{"activity_id": id, "identity_pinned": r.IdentityPinned,
		"entry_updated": r.EntryUpdated, "references": r.UpdatedReferences, "reference_total": r.ResolvedReferences,
		"changed_files": len(r.Files), "source_files": 2, "packages": 1, "program_executions": r.ProgramExecutions,
		"model_predictions": r.ModelPredictions, "input_image_digest": r.InputImageDigest, "output_image_digest": r.OutputImageDigest})
	t.Log("IMPLICIT_WORKSPACE_ENTRY_RENAME_OBSERVATION", string(raw))
}
