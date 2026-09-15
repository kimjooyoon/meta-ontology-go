package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// projectionIRFromBidirModelWithSupport is package-private so focused CLI
// tests can exercise latent SUPPORTED projection. Production callers use the
// checked-in support state through projectionIRFromBidirModel.
func projectionIRFromBidirModelWithSupport(ir semantic.IR, sourceModel bidir.Model, support syntax.EntityFieldsSupport) (generator.SemanticIR, error) {
	if err := validateCLIEntityFieldsSupport(support); err != nil {
		return generator.SemanticIR{}, err
	}
	if semanticIRHasFields(ir) && support.State == syntax.EntityFieldsDeferred {
		return generator.SemanticIR{}, errors.New("parse.entity-fields-deferred")
	}
	model, err := projectionIR(ir)
	if err != nil {
		return generator.SemanticIR{}, err
	}
	if err := sourceModel.Validate(); err != nil {
		return generator.SemanticIR{}, classifyCLIEntityFieldsModelError(err)
	}
	byID := make(map[string]bidir.Node, len(sourceModel.Nodes))
	for _, node := range sourceModel.Nodes {
		byID[string(node.ID)] = node
	}
	for entityIndex := range model.Entities {
		entity := &model.Entities[entityIndex]
		sourceNode, ok := byID[entity.ID]
		if !ok {
			if len(entity.Fields) != 0 {
				return generator.SemanticIR{}, fmt.Errorf("bidir model omitted entity %q fields", entity.ID)
			}
			continue
		}
		if sourceNode.Kind != bidir.EntityKind || len(sourceNode.Fields) != len(entity.Fields) {
			return generator.SemanticIR{}, fmt.Errorf("bidir model field cardinality for entity %q does not match semantic IR", entity.ID)
		}
		fields := make([]generator.Field, len(sourceNode.Fields))
		for fieldIndex, sourceField := range sourceNode.Fields {
			semanticField := entity.Fields[fieldIndex]
			projected, err := projectionBidirField(sourceField)
			if err != nil {
				return generator.SemanticIR{}, fmt.Errorf("entity %q field %d: %w", entity.ID, fieldIndex, err)
			}
			if !sameCLIField(projected, semanticField) {
				return generator.SemanticIR{}, fmt.Errorf("entity %q field %d disagrees with semantic authority", entity.ID, fieldIndex)
			}
			fields[fieldIndex] = projected
		}
		entity.Fields = fields
	}
	if semanticIRHasFields(ir) {
		switch support.State {
		case syntax.EntityFieldsSupported:
			if err := validateCLIProjectedFields(model); err != nil {
				return generator.SemanticIR{}, err
			}
		default:
			return generator.SemanticIR{}, fmt.Errorf("GOOO-EF-V1-UNKNOWN-STATE: %q", support.State)
		}
	}
	return model, nil
}
