package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestGeneratorEmitsDeterministicCompositionForExplicitBindings(t *testing.T) {
	file, diagnostics := syntax.ParseFile("binding.gooo", sourceWithRuntimeBinding)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("binding parse diagnostics=%v file=%#v", diagnostics, file)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	projected, err := projectionIR(ir)
	if err != nil || len(projected.Activities) == 0 {
		t.Fatalf("runtime binding projection err=%v projected=%#v", err, projected)
	}
	if _, err := buildRuntimePlanData([]byte(sourceWithRuntimeBinding), ir); err != nil {
		t.Fatalf("runtime plan build error=%v", err)
	}
	generated, err := generateWithDeadlineCore(file, nil, commandDeadline)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"func GoooComposeProduceConsume(input0 Order) Order",
		"runtime0_0 := Produce(input0)",
		"runtime1_0 := Consume(runtime0_0)",
		"return runtime1_0",
	} {
		if !strings.Contains(string(generated.result.Source), expected) {
			t.Fatalf("generated Go missing composition fragment %q:\n%s", expected, generated.result.Source)
		}
	}
	if strings.Contains(string(generated.result.Source), "GoooCompose") != strings.Contains(sourceWithRuntimeBinding, "bind ") {
		t.Fatal("composition emission must follow explicit source bindings only")
	}
}

func TestGeneratedRuntimePlanCarriesValidatedTypedOrder(t *testing.T) {
	file, diagnostics := syntax.ParseFile("binding.gooo", sourceWithRuntimeBinding)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("binding parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	typedPlan, err := bidir.CompileTypedPlan(document)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	data, err := buildRuntimePlanDataWithTypedPlan([]byte(sourceWithRuntimeBinding), ir, typedPlan)
	if err != nil {
		t.Fatal(err)
	}
	var got runtimePlanDocument
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.TypedPlanDigest != typedPlan.Digest() || len(got.ActivityOrder) != len(typedPlan.Activities) {
		t.Fatalf("typed runtime plan metadata = %#v, want digest %s and %d activities", got, typedPlan.Digest(), len(typedPlan.Activities))
	}
}
