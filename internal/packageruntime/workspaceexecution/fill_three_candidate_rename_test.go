package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestThreeCandidateFillRenamePreservesTemplateAndDormantIds(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "scope83/app", Activity: "Build"},
		Packages: []packageruntime.PackageSpec{
			{Path: "scope83/helpers", Name: "helpers", Sources: []packageruntime.Source{{Filename: "helper.gooo", Content: `package helpers
namespace scope83_helpers
entity Integer id "urn:scope83:integer"
activity Boost(Integer) -> Integer id "urn:scope83:boost" computes "return input+17"
`}}},
			{Path: "scope83/app", Name: "app", Imports: []string{"scope83/helpers"}, Sources: []packageruntime.Source{
				{Filename: "assemble.gooo", Content: `package app
namespace scope83_app
import fees "scope83/helpers"
entity Integer id "urn:scope83:integer"
activity Build(Integer) -> Integer id "urn:scope83:build" computes "if input < -2 { let fees = input; _ = fees }; let base = __GOOO_BODY_HOLE_base__; return fees.Boost(base)+__GOOO_BODY_HOLE_bias__" assembling {
 source_fill intent "Fill two expressions; a local alias in another block does not own this call." {
  hole "base"
  hole "bias"
  candidate "plain" { fill "base" "input" fill "bias" "0" }
  candidate "nested" { fill "base" "fees.\u0042oost(input)" fill "bias" "fees.Boost(0)" }
  candidate "offset" { fill "base" "fees.Boost(input+2)" fill "bias" "fees.Boost(1)" }
 }
 case "5" -> "56"
 holdout_case "9" -> "60"
}
`},
				{Filename: "dormant.gooo", Content: `package app
namespace scope83_app
import fees "scope83/helpers"
activity Carry(Integer) -> Integer computes "return fees.Boost(input)+5"
`},
			}},
			{Path: "scope83/peer", Name: "peer", Sources: []packageruntime.Source{{Filename: "peer.gooo", Content: `package peer
namespace scope83_peer
entity Integer id "urn:scope83:integer"
activity Boost(Integer) -> Integer id "urn:scope83:peer-boost" computes "return input+99"
activity Read(Integer) -> Integer computes "return Boost(input)"
`}}},
		}}
	before, _ := json.Marshal(manifest)
	r, err := RenameWorkspaceActivity(context.Background(), manifest, "scope83/helpers", "urn:scope83:boost", "가산")
	if err != nil {
		t.Fatal(err)
	}
	if r.ResolvedReferences != 6 || r.UpdatedReferences != 6 || len(r.Files) != 3 || r.EntryUpdated || r.IdentityPinned {
		t.Fatal("template, nested fills or dormant consumer have wrong source coverage", r)
	}
	source := r.Manifest.Packages[1].Sources[0].Content
	for _, expected := range []string{"fees.가산(base)", `fill "base" "fees.가산(input)"`, `fill "bias" "fees.가산(0)"`, `fill "base" "fees.가산(input+2)"`, `fill "bias" "fees.가산(1)"`,
		"let fees = input", `candidate "plain" { fill "base" "input" fill "bias" "0" }`} {
		if !strings.Contains(source, expected) {
			t.Fatal("precise fill source range not retained", expected, source)
		}
	}
	if r.Manifest.Packages[2].Sources[0].Content != manifest.Packages[2].Sources[0].Content ||
		r.ProgramExecutions != 0 || r.ModelPredictions != 0 {
		t.Fatal("peer source changed or construction executed")
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("proposal mutated input manifest")
	}
	counts := map[string]int{}
	for _, file := range r.Files {
		for _, edit := range file.Edits {
			counts[edit.Kind]++
		}
	}
	raw, _ := json.Marshal(map[string]any{"references": r.UpdatedReferences, "reference_total": r.ResolvedReferences,
		"changed_files": len(r.Files), "source_files": 4, "packages": 3, "declared_candidates": 3, "holes": 2,
		"edit_kinds": counts, "input_unchanged": string(before) == string(after), "peer_unchanged": true,
		"program_executions": r.ProgramExecutions, "model_predictions": r.ModelPredictions,
		"input_image_digest": r.InputImageDigest, "output_image_digest": r.OutputImageDigest})
	t.Log("THREE_CANDIDATE_FILL_RENAME_OBSERVATION", string(raw))
}
