package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestQualifiedFieldAliasesCarryScalarIDsThroughNativeBindings(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "field185/app", Activity: "Decide"},
		Packages: []packageruntime.PackageSpec{
			{Path: "field185/counts", Name: "counts", Sources: []packageruntime.Source{{Filename: "counts.gooo", Content: `package counts
namespace field185_counts
entity Count id "urn:gooo:type:integer"
entity Label id "urn:gooo:type:string"
`}}},
			{Path: "field185/flags", Name: "flags", Sources: []packageruntime.Source{{Filename: "flags.gooo", Content: `package flags
namespace field185_flags
entity Count id "urn:gooo:type:boolean"
`}}},
			{Path: "field185/app", Name: "app", Imports: []string{"field185/counts", "field185/flags"}, Sources: []packageruntime.Source{
				{Filename: "ledger.gooo", Content: `package app
namespace field185_app
import data "field185/counts"
entity Ledger id "urn:field185:ledger" fields {
 field amount id "urn:field185:amount" type data.Count required one
 field note id "urn:field185:note" type data.Label optional one
}
activity Pass(Ledger) -> Ledger id "urn:field185:pass" computes "return input"
`},
				{Filename: "claim.gooo", Content: `package app
namespace field185_app
import data "field185/flags"
entity Claim id "urn:field185:claim" fields {
 field eligible id "urn:field185:eligible" type data.Count required one
}
activity Decide(Ledger) -> Claim id "urn:field185:decide" computes "return Claim{eligible: input.amount > 0}"
bind Pass.result -> Decide.input
`},
			}},
		}}
	before, _ := json.Marshal(manifest)
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"field185/app:Pass":{"amount":0}},"expected":{"field185/app:Pass":{"amount":0},"field185/app:Decide":{"eligible":false}}},
 {"inputs":{"field185/app:Pass":{"amount":6,"note":""}},"expected":{"field185/app:Pass":{"amount":6,"note":""},"field185/app:Decide":{"eligible":true}}},
 {"inputs":{"field185/app:Pass":{"amount":9007199254740995,"note":"한글"}},"expected":{"field185/app:Pass":{"amount":9007199254740995,"note":"한글"},"field185/app:Decide":{"eligible":true}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FiniteTotal == 0 || result.Runtime.FinitePassed != result.Runtime.FiniteTotal ||
		!result.Runtime.RuntimeReplayed || result.Runtime.ModelCalls != 0 {
		t.Fatal("native scalar field alias behavior differs", result.Runtime)
	}
	references := 0
	for _, pkg := range result.Program.Workspace.Packages {
		for _, source := range pkg.Sources {
			references += len(source.FieldTypes)
		}
	}
	if references != 3 || strings.Contains(result.Program.Source, "type data.") {
		t.Fatal("workspace did not lower all three original field lookups", references)
	}
	ids := map[string]string{}
	for _, alias := range result.Program.EntityAliases {
		if alias.Name == "Count" {
			ids[alias.EntityID] = alias.LoweredName
		}
	}
	if len(ids) != 2 || ids["urn:gooo:type:integer"] == ids["urn:gooo:type:boolean"] {
		t.Fatal("same display name merged distinct scalar IDs", ids)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("construction rewrote original source")
	}
	observation, _ := json.Marshal(map[string]any{"source_references": references, "source_reference_total": 3,
		"finite_passed": result.Runtime.FinitePassed, "finite_total": result.Runtime.FiniteTotal,
		"runtime_replayed": result.Runtime.RuntimeReplayed, "model_calls": result.Runtime.ModelCalls,
		"input_cases": len(suite.Cases), "activity_outputs": 6, "input_unchanged": true,
		"same_named_scalar_ids": ids, "source_sha256": result.SourceSHA256})
	t.Log("GOOO_QUALIFIED_FIELD_NATIVE_OBSERVATION=" + string(observation))
}
