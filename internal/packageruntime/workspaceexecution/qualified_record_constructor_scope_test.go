package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestFileLocalRecordConstructorsKeepSameAliasDomainsAndNativeResults(t *testing.T) {
	manifest := constructorScope641Manifest()
	before, _ := json.Marshal(manifest)
	program, err := Prepare(manifest)
	if err != nil {
		t.Fatal(err)
	}
	aliases := map[string]string{}
	for _, alias := range program.EntityAliases {
		aliases[alias.EntityID] = alias.LoweredName
	}
	if len(aliases) != 2 || aliases["urn:constructor641:money"] == aliases["urn:constructor641:quantity"] {
		t.Fatal("file-local constructor domains became one native name", aliases)
	}
	if !strings.Contains(program.Source, "資料.Value{value: 999}") || !strings.Contains(program.Source, "資料.Value{ignored}") {
		t.Fatal("constructor lowering changed a string or comment")
	}
	for _, variant := range []struct{ from, to, reason string }{
		{"return 資料.Value{", "return absent.Value{", "exported by an import in this source file"},
		{"return 資料.Value{", "return 資料.Missing{", "exported by an import in this source file"},
		{"// 資料.Value{ignored}", "let 資料 = input0.value", "shadowed by a local value"},
	} {
		wrong := constructorScope641Manifest()
		wrong.Packages[2].Sources[0].Content = strings.ReplaceAll(wrong.Packages[2].Sources[0].Content, variant.from, variant.to)
		partial, err := Prepare(wrong)
		if err == nil || !strings.Contains(err.Error(), variant.reason) || partial.Schema != "" {
			t.Fatal("invalid constructor scope returned a partial program", variant.reason, partial, err)
		}
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"constructor641/money:Keep":{"value":13,"note":"seed"},"constructor641/count:Keep":{"value":8}},"expected":{"constructor641/money:Keep":{"value":13,"note":"seed"},"constructor641/count:Keep":{"value":8},"constructor641/app:AddQuantity":{"value":10},"constructor641/app:Multiply":{"value":130,"note":"資料.Value{value: 999}"}}},
 {"inputs":{"constructor641/money:Keep":{"value":-9,"note":"negative"},"constructor641/count:Keep":{"value":3}},"expected":{"constructor641/money:Keep":{"value":-9,"note":"negative"},"constructor641/count:Keep":{"value":3},"constructor641/app:AddQuantity":{"value":5},"constructor641/app:Multiply":{"value":0,"note":"資料.Value{value: 999}"}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 8 || result.Runtime.FiniteTotal != 8 || !result.Runtime.RuntimeReplayed || result.Runtime.ModelCalls != 0 {
		t.Fatal("scoped constructors differ from source behavior", result.Runtime)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("constructor lowering changed original source")
	}
	observation, _ := json.Marshal(map[string]any{"file_local_alias_domains": 2, "domain_names": aliases,
		"qualified_constructor_sites": 3, "invalid_scope_declines": 3, "partial_programs_accepted": 0,
		"input_cases": len(suite.Cases), "finite_passed": result.Runtime.FinitePassed, "finite_total": result.Runtime.FiniteTotal,
		"model_calls": result.Runtime.ModelCalls, "runtime_replayed": result.Runtime.RuntimeReplayed,
		"source_unchanged": true, "strings_and_comments_preserved": true, "source_sha256": result.SourceSHA256})
	t.Log("GOOO_QUALIFIED_RECORD_CONSTRUCTOR_OBSERVATION=" + string(observation))
}

func constructorScope641Manifest() packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "constructor641/app", Activity: "Multiply"},
		Packages: []packageruntime.PackageSpec{
			{Path: "constructor641/money", Name: "money", Sources: []packageruntime.Source{{Filename: "money.gooo", Content: `package money
namespace constructor641_money
entity Value id "urn:constructor641:money" fields {
 field value id "urn:constructor641:money-value" type "urn:gooo:type:integer" required one
 field note id "urn:constructor641:money-note" type "urn:gooo:type:string" required one
}
activity Keep(Value) -> Value id "urn:constructor641:keep-money" computes "return input"
`}}},
			{Path: "constructor641/count", Name: "count", Sources: []packageruntime.Source{{Filename: "count.gooo", Content: `package count
namespace constructor641_count
entity Value id "urn:constructor641:quantity" fields {
 field value id "urn:constructor641:quantity-value" type "urn:gooo:type:integer" required one
}
activity Keep(Value) -> Value id "urn:constructor641:keep-quantity" computes "return input"
`}}},
			{Path: "constructor641/app", Name: "app", Imports: []string{"constructor641/money", "constructor641/count"}, Sources: []packageruntime.Source{
				{Filename: "돈.gooo", Content: `package app
namespace constructor641_app
import 資料 "constructor641/money"
import 数量 "constructor641/count"
activity Multiply(資料.Value, 数量.Value) -> 資料.Value id "urn:constructor641:multiply" computes "// 資料.Value{ignored}\nif input0.value < 0 { return 資料.Value{value: 0, note: \"資料.Value{value: 999}\"} }; return 資料.Value{value: input0.value * input1.value, note: \"資料.Value{value: 999}\"}"
bind 資料.Keep.result -> Multiply.input0
bind AddQuantity.result -> Multiply.input1
`},
				{Filename: "수량.gooo", Content: `package app
namespace constructor641_app
import 資料 "constructor641/count"
activity AddQuantity(資料.Value) -> 資料.Value id "urn:constructor641:add-quantity" computes "return 資料.Value{value: input.value + 2}"
bind 資料.Keep.result -> AddQuantity.input
`},
			}},
		}}
}
