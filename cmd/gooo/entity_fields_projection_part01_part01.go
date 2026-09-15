package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// projectionIRFromBidirModel enriches the semantic projection with the BX
// source provenance that the semantic IR intentionally does not duplicate.
// The BX model is the only CLI-owned input that carries the five exact field
// subspans, so this adapter refuses to synthesize them from names or output.
func projectionIRFromBidirModel(ir semantic.IR, sourceModel bidir.Model) (generator.SemanticIR, error) {
	return projectionIRFromBidirModelWithSupport(ir, sourceModel, syntax.CurrentEntityFieldsSupport())
}
