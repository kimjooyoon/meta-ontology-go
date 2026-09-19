package main

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestGeneratorProjectsRuntimeBindingsUntilExplicitPlanDelivery(t *testing.T) {
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
