package main

import (
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
