package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestImportedRecordConstructorFollowsItsDeclaredModule(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "constructor612/app", Activity: "Multiply"},
		Packages: []packageruntime.PackageSpec{
			{Path: "constructor612/money", Name: "money", Sources: []packageruntime.Source{{Filename: "money.gooo", Content: `package money
namespace constructor612_money
entity Value id "urn:constructor612:money" fields {
 field value id "urn:constructor612:money-value" type "urn:gooo:type:integer" required one
}
activity Keep(Value) -> Value id "urn:constructor612:keep-money" computes "return input"
`}}},
			{Path: "constructor612/count", Name: "count", Sources: []packageruntime.Source{{Filename: "count.gooo", Content: `package count
namespace constructor612_count
entity Value id "urn:constructor612:quantity" fields {
 field value id "urn:constructor612:quantity-value" type "urn:gooo:type:integer" required one
}
activity Keep(Value) -> Value id "urn:constructor612:keep-quantity" computes "return input"
`}}},
			{Path: "constructor612/app", Name: "app", Imports: []string{"constructor612/money", "constructor612/count"}, Sources: []packageruntime.Source{{Filename: "main.gooo", Content: `package app
namespace constructor612_app
import price "constructor612/money"
import quantity "constructor612/count"
activity Multiply(price.Value, quantity.Value) -> price.Value id "urn:constructor612:multiply" computes "return price.Value{value: input0.value * input1.value}"
bind price.Keep.result -> Multiply.input0
bind quantity.Keep.result -> Multiply.input1
`}}},
		}}
	before, _ := json.Marshal(manifest)
	program, err := Prepare(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(program.Source, "return price.Value{") {
		t.Fatal("the imported record constructor retained an unresolved alias in the executable graph")
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"constructor612/money:Keep":{"value":13},"constructor612/count:Keep":{"value":8}},"expected":{"constructor612/money:Keep":{"value":13},"constructor612/count:Keep":{"value":8},"constructor612/app:Multiply":{"value":104}}},
 {"inputs":{"constructor612/money:Keep":{"value":-9},"constructor612/count:Keep":{"value":3}},"expected":{"constructor612/money:Keep":{"value":-9},"constructor612/count:Keep":{"value":3},"constructor612/app:Multiply":{"value":-27}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 6 || result.Runtime.FiniteTotal != 6 || !result.Runtime.RuntimeReplayed || result.Runtime.ModelCalls != 0 {
		t.Fatal("module constructor differs from source calculation", result.Runtime)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("constructor lowering rewrote source")
	}
}
