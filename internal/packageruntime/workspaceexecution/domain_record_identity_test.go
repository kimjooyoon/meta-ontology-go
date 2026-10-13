package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestDomainRecordIdentitySeparatesEqualLayoutsBeforeNativeExecution(t *testing.T) {
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "domain438/app", Activity: "Multiply"},
		Packages: []packageruntime.PackageSpec{
			{Path: "domain438/money", Name: "money", Sources: []packageruntime.Source{{Filename: "money.gooo", Content: `package money
namespace domain438_money
entity Value id "urn:domain438:money" fields {
 field value id "urn:domain438:money-value" type "urn:gooo:type:integer" required one
}
activity Keep(Value) -> Value id "urn:domain438:keep-money" computes "return input"
`}}},
			{Path: "domain438/count", Name: "count", Sources: []packageruntime.Source{{Filename: "count.gooo", Content: `package count
namespace domain438_count
entity Value id "urn:domain438:quantity" fields {
 field value id "urn:domain438:quantity-value" type "urn:gooo:type:integer" required one
}
activity Keep(Value) -> Value id "urn:domain438:keep-quantity" computes "return input"
`}}},
			{Path: "domain438/app", Name: "app", Imports: []string{"domain438/money", "domain438/count"}, Sources: []packageruntime.Source{{Filename: "invoice.gooo", Content: `package app
namespace domain438_app
import price "domain438/money"
import quantity "domain438/count"
entity Total id "urn:domain438:total" fields {
 field value id "urn:domain438:total-value" type "urn:gooo:type:integer" required one
}
activity Multiply(price.Value, quantity.Value) -> Total id "urn:domain438:multiply" computes "return Total{value: input0.value * input1.value}"
bind price.Keep.result -> Multiply.input0
bind quantity.Keep.result -> Multiply.input1
`}}},
		}}
	before, _ := json.Marshal(manifest)
	image, err := packageruntime.Build(manifest)
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range image.Packages {
		if pkg.Path == "domain438/app" {
			if len(pkg.Bindings) != 2 || pkg.Bindings[0].EntityID == pkg.Bindings[1].EntityID {
				t.Fatal("different domain IDs became one binding type", pkg.Bindings)
			}
		}
	}
	wrong := manifest
	wrong.Packages = append([]packageruntime.PackageSpec(nil), manifest.Packages...)
	wrong.Packages[2].Sources = append([]packageruntime.Source(nil), manifest.Packages[2].Sources...)
	wrong.Packages[2].Sources[0].Content = strings.ReplaceAll(wrong.Packages[2].Sources[0].Content,
		"bind price.Keep.result -> Multiply.input0", "bind quantity.Keep.result -> Multiply.input0")
	declined, err := packageruntime.Build(wrong)
	if err == nil || !strings.Contains(err.Error(), "PACKAGE_BINDING_TYPE_MISMATCH") || declined.Schema != "" {
		t.Fatal("quantity output was accepted as a money input", declined, err)
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"domain438/money:Keep":{"value":29},"domain438/count:Keep":{"value":4}},"expected":{"domain438/money:Keep":{"value":29},"domain438/count:Keep":{"value":4},"domain438/app:Multiply":{"value":116}}},
 {"inputs":{"domain438/money:Keep":{"value":-3},"domain438/count:Keep":{"value":5}},"expected":{"domain438/money:Keep":{"value":-3},"domain438/count:Keep":{"value":5},"domain438/app:Multiply":{"value":-15}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 6 || result.Runtime.FiniteTotal != 6 ||
		!result.Runtime.RuntimeReplayed || result.Runtime.ModelCalls != 0 {
		t.Fatal("source-authored domain calculation differs", result.Runtime)
	}
	aliases := map[string]string{}
	for _, alias := range result.Program.EntityAliases {
		if alias.Name == "Value" {
			aliases[alias.EntityID] = alias.LoweredName
		}
	}
	if len(aliases) != 2 || aliases["urn:domain438:money"] == aliases["urn:domain438:quantity"] {
		t.Fatal("equal-layout domains lost separate native names", aliases)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("domain construction rewrote source")
	}
	observation, _ := json.Marshal(map[string]any{"domain_ids": 2, "same_named_record_layouts": 2,
		"typed_bindings": 2, "wrong_domain_bindings_declined": 1, "partial_images_accepted": 0,
		"input_cases": len(suite.Cases), "activity_outputs": 6, "finite_passed": result.Runtime.FinitePassed,
		"finite_total": result.Runtime.FiniteTotal, "runtime_replayed": result.Runtime.RuntimeReplayed,
		"model_calls": result.Runtime.ModelCalls, "input_unchanged": true,
		"lowered_domain_names": aliases, "source_sha256": result.SourceSHA256})
	t.Log("GOOO_DOMAIN_RECORD_IDENTITY_OBSERVATION=" + string(observation))
}
