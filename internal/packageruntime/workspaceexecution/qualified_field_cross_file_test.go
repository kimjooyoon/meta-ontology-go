package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestQualifiedFieldAliasesAndCrossFileBindingsReachNativeExecution(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "field270/app", Activity: "Decide"},
		Packages: []packageruntime.PackageSpec{
			{Path: "field270/counts", Name: "counts", Sources: []packageruntime.Source{{Filename: "counts.gooo", Content: `package counts
namespace field270_counts
entity Count id "urn:gooo:type:integer"
entity Label id "urn:gooo:type:string"
`}}},
			{Path: "field270/flags", Name: "flags", Sources: []packageruntime.Source{{Filename: "flags.gooo", Content: `package flags
namespace field270_flags
entity Count id "urn:gooo:type:boolean"
`}}},
			{Path: "field270/app", Name: "app", Imports: []string{"field270/counts", "field270/flags"}, Sources: []packageruntime.Source{
				{Filename: "ledger.gooo", Content: `package app
namespace field270_app
import data "field270/counts"
entity Ledger id "urn:field270:ledger" fields {
 field amount id "urn:field270:amount" type data.Count required one
 field note id "urn:field270:note" type data.Label optional one
}
activity Pass(Ledger) -> Ledger id "urn:field270:pass" computes "return input"
`},
				{Filename: "claim.gooo", Content: `package app
namespace field270_app
import data "field270/flags"
entity Claim id "urn:field270:claim" fields {
 field eligible id "urn:field270:eligible" type data.Count required one
}
activity Decide(Ledger) -> Claim id "urn:field270:decide" computes "return Claim{eligible: input.amount >= 31}"
bind Pass.result -> Decide.input
`},
			}},
		}}
	before, _ := json.Marshal(manifest)
	public, err := packageruntime.Describe(manifest)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"urn:field270:amount": "urn:gooo:type:integer", "urn:field270:note": "urn:gooo:type:string", "urn:field270:eligible": "urn:gooo:type:boolean"}
	interfaceFields := 0
	for _, pkg := range public.Packages {
		for _, declaration := range pkg.Declarations {
			for _, field := range declaration.Fields {
				if field.TypeID != want[field.ID] {
					t.Fatal("interface field resolved to another type", field)
				}
				interfaceFields++
			}
		}
	}
	if interfaceFields != 3 {
		t.Fatal("interface lost fields", interfaceFields)
	}
	formatted, diagnostics, err := syntax.FormatSourceWithEntityFieldsSupport("ledger.gooo", manifest.Packages[2].Sources[0].Content, syntax.EntityFieldsV4Support())
	if err != nil || diagnostics.HasErrors() || !strings.Contains(formatted, "type data.Count required one") {
		t.Fatal("qualified spelling did not format", err, diagnostics)
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"field270/app:Pass":{"amount":0}},"expected":{"field270/app:Pass":{"amount":0},"field270/app:Decide":{"eligible":false}}},
 {"inputs":{"field270/app:Pass":{"amount":31,"note":""}},"expected":{"field270/app:Pass":{"amount":31,"note":""},"field270/app:Decide":{"eligible":true}}},
 {"inputs":{"field270/app:Pass":{"amount":9007199254740995,"note":"한글"}},"expected":{"field270/app:Pass":{"amount":9007199254740995,"note":"한글"},"field270/app:Decide":{"eligible":true}}}
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
	edges := 0
	for _, pkg := range result.Program.Workspace.Packages {
		for _, edge := range pkg.Bindings {
			if edge.Source == nil {
				continue
			}
			edges++
			if edge.EntityID != "urn:field270:ledger" || edge.Source.ProducerID != "urn:field270:pass" || edge.Source.ConsumerID != "urn:field270:decide" || edge.Source.Filename != "claim.gooo" || manifest.Packages[2].Sources[1].Content[edge.Source.Start:edge.Source.End] != "bind Pass.result -> Decide.input" {
				t.Fatal("cross-file binding lost source or stable IDs", edge)
			}
		}
	}
	if edges != 1 {
		t.Fatal("cross-file binding evidence absent", edges)
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
	observation, _ := json.Marshal(map[string]any{"source_references": references, "source_reference_total": 3, "cross_file_bindings": edges, "interface_fields": interfaceFields, "canonical_field_types": want, "formatted_qualified_spelling": true,
		"finite_passed": result.Runtime.FinitePassed, "finite_total": result.Runtime.FiniteTotal,
		"runtime_replayed": result.Runtime.RuntimeReplayed, "model_calls": result.Runtime.ModelCalls,
		"input_cases": len(suite.Cases), "activity_outputs": 6, "input_unchanged": true,
		"same_named_scalar_ids": ids, "source_sha256": result.SourceSHA256})
	t.Log("GOOO_QUALIFIED_FIELD_CROSS_FILE_OBSERVATION=" + string(observation))
}
