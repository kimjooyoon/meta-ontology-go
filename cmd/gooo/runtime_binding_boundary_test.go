package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestGeneratorRejectsRuntimeBindingsWithoutAPlan(t *testing.T) {
	file, diagnostics := syntax.ParseFile("binding.gooo", sourceWithRuntimeBinding)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("binding parse diagnostics=%v file=%#v", diagnostics, file)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := projectionIR(ir); err == nil || err.Error() != errRuntimeBindingsUnsupportedByGenerator.Error() {
		t.Fatalf("generator boundary error=%v", err)
	}
}
