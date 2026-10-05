package main

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestEntityFieldsV2FlowsThroughCLIProjectionAndGenerator(t *testing.T) {
	const source = `package records
namespace records
entity Boolean id "records://boolean"
entity Gate id "records://gate" fields {
    field enabled id "records://gate/enabled" type boolean required one
}
activity Build(Boolean) -> Gate`
	support := syntax.EntityFieldsV2Support()
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("gate.gooo", source, support)
	if diagnostics.HasErrors() {
		t.Fatal("parse", diagnostics)
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, support)
	if err != nil {
		t.Fatal("lower", err)
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal("document", err)
	}
	model, err := bidir.GetWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal("bidir model", err)
	}
	projected, err := projectionIRFromBidirModelWithSupport(ir, model, support)
	if err != nil {
		t.Fatal("CLI projection", err)
	}
	generated, err := generator.GenerateWithEntityFieldsSupport(projected, nil, support)
	if err != nil {
		t.Fatal("Go generation", err)
	}
	if !strings.Contains(string(generated.Source), "enabled bool") {
		t.Fatalf("Boolean field was not emitted by the compiler:\n%s", generated.Source)
	}
}
