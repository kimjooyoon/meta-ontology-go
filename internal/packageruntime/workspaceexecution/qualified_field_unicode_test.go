package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestUnicodeFieldAliasesRetainByteRangesThroughNativeExecution(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "field352/app", Activity: "Decide"},
		Packages: []packageruntime.PackageSpec{
			{Path: "field352/counts", Name: "counts", Sources: []packageruntime.Source{{Filename: "계수.gooo", Content: `package counts
namespace field352_counts
entity 값 id "urn:gooo:type:integer"
entity 문장 id "urn:gooo:type:string"
`}}},
			{Path: "field352/flags", Name: "flags", Sources: []packageruntime.Source{{Filename: "상태.gooo", Content: `package flags
namespace field352_flags
entity 값 id "urn:gooo:type:boolean"
`}}},
			{Path: "field352/app", Name: "app", Imports: []string{"field352/counts", "field352/flags"}, Sources: []packageruntime.Source{
				{Filename: "원장.gooo", Content: `package app
namespace field352_app
import 자료 "field352/counts"
entity Ledger id "urn:field352:ledger" fields {
 field amount id "urn:field352:amount" type 자료.값 required one
 field note id "urn:field352:note" type 자료.문장 optional one
}
activity Pass(Ledger) -> Ledger id "urn:field352:pass" computes "return input"
`},
				{Filename: "판정.gooo", Content: `package app
namespace field352_app
import 자료 "field352/flags"
entity Claim id "urn:field352:claim" fields {
 field eligible id "urn:field352:eligible" type 자료.값 required one
}
activity Decide(Ledger) -> Claim id "urn:field352:decide" computes "return Claim{eligible: input.amount >= 43}"
bind Pass.result -> Decide.input
`},
			}},
		}}
	before, _ := json.Marshal(manifest)
	formatted, diagnostics, err := syntax.FormatSourceWithEntityFieldsSupport("원장.gooo",
		manifest.Packages[2].Sources[0].Content, syntax.EntityFieldsV4Support())
	if err != nil || diagnostics.HasErrors() || !strings.Contains(formatted, "type 자료.값 required one") {
		t.Fatal("Unicode qualified type spelling did not format", err, diagnostics)
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"field352/app:Pass":{"amount":42}},"expected":{"field352/app:Pass":{"amount":42},"field352/app:Decide":{"eligible":false}}},
 {"inputs":{"field352/app:Pass":{"amount":43,"note":"기준 충족"}},"expected":{"field352/app:Pass":{"amount":43,"note":"기준 충족"},"field352/app:Decide":{"eligible":true}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 4 || result.Runtime.FiniteTotal != 4 ||
		!result.Runtime.RuntimeReplayed || result.Runtime.ModelCalls != 0 {
		t.Fatal("Unicode source native outputs differ", result.Runtime)
	}
	want := map[string]string{"urn:field352:amount": "urn:gooo:type:integer",
		"urn:field352:note": "urn:gooo:type:string", "urn:field352:eligible": "urn:gooo:type:boolean"}
	references, edges := 0, 0
	for _, pkg := range result.Program.Workspace.Packages {
		for _, source := range pkg.Sources {
			for _, ref := range source.FieldTypes {
				references++
				if ref.TypeID != want[ref.FieldID] || !strings.HasPrefix(ref.Spelling, "자료.") ||
					len(ref.Spelling) <= utf8.RuneCountInString(ref.Spelling) {
					t.Fatal("Unicode field reference identity or spelling differs", ref)
				}
				matched := false
				for _, original := range manifest.Packages[2].Sources {
					if original.Filename == source.Filename {
						matched = ref.Start >= 0 && ref.End <= len(original.Content) && ref.Start < ref.End &&
							original.Content[ref.Start:ref.End] == ref.Spelling
					}
				}
				if !matched {
					t.Fatal("byte range does not locate the original Unicode reference", ref)
				}
			}
		}
		for _, edge := range pkg.Bindings {
			if edge.Source != nil {
				edges++
				if edge.Source.Filename != "판정.gooo" || edge.Source.ProducerID != "urn:field352:pass" ||
					edge.Source.ConsumerID != "urn:field352:decide" || edge.EntityID != "urn:field352:ledger" ||
					manifest.Packages[2].Sources[1].Content[edge.Source.Start:edge.Source.End] != "bind Pass.result -> Decide.input" {
					t.Fatal("Unicode file binding lost source identity", edge)
				}
			}
		}
	}
	after, _ := json.Marshal(manifest)
	if references != 3 || edges != 1 || string(before) != string(after) {
		t.Fatal("Unicode source references or original manifest changed", references, edges)
	}
	observation, _ := json.Marshal(map[string]any{"source_references": references, "source_reference_total": 3,
		"cross_file_bindings": edges, "unicode_source_filenames": 4, "input_cases": len(suite.Cases),
		"finite_passed": result.Runtime.FinitePassed, "finite_total": result.Runtime.FiniteTotal,
		"runtime_replayed": result.Runtime.RuntimeReplayed, "model_calls": result.Runtime.ModelCalls,
		"original_byte_ranges_verified": true, "input_unchanged": true, "source_sha256": result.SourceSHA256})
	t.Log("GOOO_UNICODE_FIELD_ALIAS_OBSERVATION=" + string(observation))
}
