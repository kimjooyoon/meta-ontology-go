package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func checkpointReferenceRenameWorkspace() packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "income/rates", Activity: "Reference"},
		Packages: []packageruntime.PackageSpec{
			{Path: "income/rates", Name: "rates", Sources: []packageruntime.Source{
				{Filename: "rates.gooo", Content: "package rates\nnamespace income_rates\nentity Integer id \"urn:income:integer\"\nactivity Reference(Integer) -> Integer computes `return input*2`\n"},
				{Filename: "local.gooo", Content: "package rates\nnamespace income_rates\nactivity Local(Integer) -> Integer computes `return Reference(input)`\n"},
			}},
			{Path: "income/peer", Name: "peer", Sources: []packageruntime.Source{{Filename: "peer.gooo", Content: "package peer\nnamespace income_peer\nentity Integer id \"urn:income:integer\"\nactivity Reference(Integer) -> Integer id \"urn:income:peer\" computes `return input*3`\n"}}},
			{Path: "income/client", Name: "client", Imports: []string{"income/rates", "income/peer"}, Sources: []packageruntime.Source{
				{Filename: "draft.gooo", Content: `package client
namespace income_client
import fees "income/rates"
activity Quote(Integer) -> Integer computes "let text=\"Reference()\"; if text==\"\" { return input }; return fees.\u0052eference(input)"
activity Draft(Integer) -> Integer computes "return fees.Reference(input)-8" assembling {
 choice "direction" operand_order at "0" intent "deduct eight from the reference amount"
 baseline "return fees.Reference(input)-8"
 picked "direction" -> "layout_forward"
 case "4" -> "0"
 observe { input "9" oracle "fees.Reference" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}
activity Sink(Integer) -> Integer computes "return input"
bind fees.Reference.result -> Sink.input
`},
				{Filename: "peer_client.gooo", Content: "package client\nnamespace income_client\nimport fees \"income/peer\"\nactivity Peer(Integer) -> Integer computes `return fees.Reference(input)`\n"},
			}},
		}}
}

func TestWorkspaceRenameRetainsCheckpointAndPinsImplicitEntry(t *testing.T) {
	manifest := checkpointReferenceRenameWorkspace()
	original, _ := json.Marshal(manifest)
	id, err := workspaceActivityIdentity("income_rates", "Reference", "")
	if err != nil {
		t.Fatal(err)
	}
	r, err := RenameWorkspaceActivity(context.Background(), manifest, "income/rates", id, "기준액")
	if err != nil {
		t.Fatal(err)
	}
	if !r.IdentityPinned || !r.EntryUpdated || r.Manifest.Entry.Activity != "기준액" || r.ActivityID != id ||
		r.ResolvedReferences != 6 || r.UpdatedReferences != 6 || len(r.Files) != 3 {
		t.Fatal("implicit identity, entry or reference counts changed")
	}
	checkCheckpointRenameLocality(t, r, manifest)
	untouched, _ := json.Marshal(manifest)
	if string(untouched) != string(original) || r.ProgramExecutions != 0 || r.ModelPredictions != 0 {
		t.Fatal("proposal mutated or executed input")
	}
	observation, _ := json.Marshal(map[string]any{"activity_id": id, "references": r.UpdatedReferences,
		"reference_total": r.ResolvedReferences, "changed_files": len(r.Files), "packages": 3, "source_files": 5,
		"entry_updated": r.EntryUpdated, "identity_pinned": r.IdentityPinned,
		"input_image_digest": r.InputImageDigest, "output_image_digest": r.OutputImageDigest,
		"program_executions": r.ProgramExecutions, "model_predictions": r.ModelPredictions})
	t.Log("WORKSPACE_CHECKPOINT_RENAME_OBSERVATION", string(observation))
}

func checkCheckpointRenameLocality(t *testing.T, r WorkspaceActivityRename, original packageruntime.Manifest) {
	t.Helper()
	decl := r.Manifest.Packages[0].Sources[0].Content
	local := r.Manifest.Packages[0].Sources[1].Content
	client := r.Manifest.Packages[2].Sources[0].Content
	if !strings.Contains(decl, `activity 기준액(Integer) -> Integer id "`+r.ActivityID+`"`) ||
		!strings.Contains(local, "return 기준액(input)") || !strings.Contains(client, "fees.기준액(input)") ||
		!strings.Contains(client, `baseline "return fees.기준액(input)-8"`) ||
		!strings.Contains(client, `picked "direction" -> "layout_forward"`) ||
		!strings.Contains(client, `oracle "fees.기준액"`) || !strings.Contains(client, "bind fees.기준액.result -> Sink.input") ||
		!strings.Contains(client, `\"Reference()\"`) {
		t.Fatal("declaration, escaped call, checkpoint, oracle, binding or prose locality changed")
	}
	if r.Manifest.Packages[1].Sources[0].Content != original.Packages[1].Sources[0].Content ||
		r.Manifest.Packages[2].Sources[1].Content != original.Packages[2].Sources[1].Content {
		t.Fatal("same-name peer or file-local alias to another package changed")
	}
}
