package main

import (
	"errors"
	"fmt"
	"strings"

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

func sameCLIField(projected generator.Field, semanticField generator.Field) bool {
	return projected.ID == semanticField.ID && projected.Parent == semanticField.Parent &&
		projected.Name == semanticField.Name && projected.TypeRefID == semanticField.TypeRefID &&
		projected.Presence == semanticField.Presence && projected.Cardinality == semanticField.Cardinality &&
		projected.Source == semanticField.Source
}

func classifyCLIEntityFieldsModelError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "duplicate field ID"), strings.Contains(message, "collides with"):
		return fmt.Errorf("GOOO-EF-V1-ID-COLLISION: %w", err)
	case strings.Contains(message, "parent"):
		return fmt.Errorf("GOOO-EF-V1-WRONG-PARENT: %w", err)
	case strings.Contains(message, "type ref"), strings.Contains(message, "unknown semantic type"), strings.Contains(message, "ambiguous semantic type"):
		return fmt.Errorf("GOOO-EF-V1-UNKNOWN-TYPE: %w", err)
	case strings.Contains(message, "span"):
		return fmt.Errorf("GOOO-EF-V1-INCOMPLETE-FIELD: %w", err)
	default:
		return fmt.Errorf("GOOO-EF-V1-INCOMPLETE-FIELD: %w", err)
	}
}

func validateCLIEntityFieldsSupport(support syntax.EntityFieldsSupport) error {
	switch support.State {
	case syntax.EntityFieldsDeferred, syntax.EntityFieldsSupported:
		if err := support.Profile.Validate(); err != nil {
			return fmt.Errorf("GOOO-EF-V1-PROFILE-MISMATCH: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("GOOO-EF-V1-UNKNOWN-STATE: %q", support.State)
	}
}

func validateCLIProjectedFields(model generator.SemanticIR) error {
	for _, entity := range model.Entities {
		if len(entity.Fields) == 0 {
			continue
		}
		seenIDs := make(map[string]struct{}, len(entity.Fields))
		seenNames := make(map[string]struct{}, len(entity.Fields))
		var sourceURI string
		var previousStart int
		for index, field := range entity.Fields {
			if field.ID == "" || field.Parent == "" || field.Name == "" || field.TypeRefID == "" || field.Presence == "" || field.Cardinality == "" {
				return fmt.Errorf("GOOO-EF-V1-INCOMPLETE-FIELD: entity %q field %d", entity.ID, index)
			}
			if field.Parent != entity.ID {
				return fmt.Errorf("GOOO-EF-V1-WRONG-PARENT: field %q parent %q does not match %q", field.ID, field.Parent, entity.ID)
			}
			if _, exists := seenIDs[field.ID]; exists {
				return fmt.Errorf("GOOO-EF-V1-ID-COLLISION: field %q is duplicated", field.ID)
			}
			seenIDs[field.ID] = struct{}{}
			if _, exists := seenNames[field.Name]; exists {
				return fmt.Errorf("GOOO-EF-V1-GO-NAME-COLLISION: field name %q is duplicated", field.Name)
			}
			seenNames[field.Name] = struct{}{}
			if field.TypeRefID != "urn:gooo:type:string" {
				return fmt.Errorf("GOOO-EF-V1-UNSUPPORTED-TYPE: field %q type %q", field.ID, field.TypeRefID)
			}
			if field.Presence != "required" || field.Cardinality != "one" {
				return fmt.Errorf("GOOO-EF-V1-UNSUPPORTED-SHAPE: field %q", field.ID)
			}
			if err := validateCLIFieldSpans(entity, field, sourceURI, previousStart, index > 0); err != nil {
				return err
			}
			if sourceURI == "" {
				sourceURI = field.Source.URI
			}
			previousStart = field.Source.Start.Offset
		}
	}
	return nil
}

func validateCLIFieldSpans(entity generator.Entity, field generator.Field, sourceURI string, previousStart int, hasPrevious bool) error {
	if field.Source.URI == "" || field.Source.End.Offset <= field.Source.Start.Offset || field.Source.Start.Line <= 0 || field.Source.End.Line <= 0 {
		return fmt.Errorf("GOOO-EF-V1-INCOMPLETE-FIELD: field %q source span is missing", field.ID)
	}
	if sourceURI != "" && field.Source.URI != sourceURI {
		return fmt.Errorf("GOOO-EF-V1-UNREPRESENTABLE: field %q crosses source snapshots", field.ID)
	}
	if entity.Source.URI != "" && field.Source.URI != entity.Source.URI {
		return fmt.Errorf("GOOO-EF-V1-UNREPRESENTABLE: field %q does not match entity source snapshot", field.ID)
	}
	if hasPrevious && field.Source.Start.Offset <= previousStart {
		return fmt.Errorf("GOOO-EF-V1-ILLEGAL-REORDER: field %q is not source ordered", field.ID)
	}
	spans := []generator.SourceSpan{field.IDSpan, field.NameSpan, field.TypeRefSpan, field.PresenceSpan, field.CardinalitySpan}
	for _, span := range spans {
		if span.URI != field.Source.URI || span.End.Offset <= span.Start.Offset || span.Start.Offset < field.Source.Start.Offset || span.End.Offset > field.Source.End.Offset || span.Start.Line <= 0 || span.End.Line <= 0 {
			return fmt.Errorf("GOOO-EF-V1-UNREPRESENTABLE: field %q has an invalid source subspan", field.ID)
		}
	}
	for index := 1; index < len(spans); index++ {
		if spans[index-1].End.Offset > spans[index].Start.Offset {
			return fmt.Errorf("GOOO-EF-V1-ILLEGAL-REORDER: field %q source subspans overlap", field.ID)
		}
	}
	return nil
}

func projectionSemanticFields(node semantic.Node) ([]generator.Field, error) {
	if len(node.Fields) == 0 {
		return nil, nil
	}
	fields := make([]generator.Field, len(node.Fields))
	for index, field := range node.Fields {
		if field.Parent != node.ID {
			return nil, fmt.Errorf("entity %q field %d has parent %q", node.ID, index, field.Parent)
		}
		if field.TypeRef.ID == "" {
			return nil, fmt.Errorf("entity %q field %d has no resolved TypeRef.ID", node.ID, index)
		}
		fields[index] = generator.Field{
			ID: string(field.ID), Parent: string(field.Parent), Name: field.Name,
			Aliases: append([]string(nil), field.Aliases...), TypeRefID: string(field.TypeRef.ID),
			Presence: string(field.Presence), Cardinality: string(field.Cardinality), Source: generatorSpan(field.Span),
		}
	}
	return fields, nil
}

func projectionBidirField(field bidir.Field) (generator.Field, error) {
	typeRefID := field.TypeRef.ID
	if typeRefID == "" {
		typeRefID = semantic.ID(field.TypeRefUse.ResolvedID)
	}
	if typeRefID == "" {
		return generator.Field{}, errors.New("field has no resolved TypeRef.ID")
	}
	return generator.Field{
		ID: string(field.ID), Parent: string(field.Parent), Name: field.Name, Aliases: append([]string(nil), field.Aliases...),
		TypeRefID: string(typeRefID), Presence: string(field.Presence), Cardinality: string(field.Cardinality), Origin: string(field.Origin),
		Source: bidirGeneratorSpan(field.Span), IDSpan: bidirGeneratorSpan(field.IDSpan), NameSpan: bidirGeneratorSpan(field.NameSpan),
		TypeRefSpan: bidirGeneratorSpan(field.TypeRefSpan), PresenceSpan: bidirGeneratorSpan(field.PresenceSpan),
		CardinalitySpan: bidirGeneratorSpan(field.CardinalitySpan), NameSource: bidirGeneratorSpan(field.NameSpan),
	}, nil
}

func bidirGeneratorSpan(span bidir.SourceSpan) generator.SourceSpan {
	return generator.SourceSpan{URI: span.File, Start: generator.Position{Offset: span.Start, Line: span.StartLine, Column: span.StartColumn}, End: generator.Position{Offset: span.End, Line: span.EndLine, Column: span.EndColumn}}
}

func semanticIRHasFields(ir semantic.IR) bool {
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Entity && len(node.Fields) > 0 {
			return true
		}
	}
	return false
}
