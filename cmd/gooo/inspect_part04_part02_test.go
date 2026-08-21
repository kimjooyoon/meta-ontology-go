package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func lowerInspectFixtureIR(t *testing.T) semantic.IR {
	t.Helper()
	file, diagnostics := syntax.Parse(sourceOrderA)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	return ir
}
