package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
	if field.TypeRef.ID == "" {
		return generator.Field{}, errors.New("GOOO-EF-V1-INCOMPLETE-FIELD: field has no authoritative TypeRef.ID")
	}
	return generator.Field{
		ID: string(field.ID), Parent: string(field.Parent), Name: field.Name, Aliases: append([]string(nil), field.Aliases...),
		TypeRefID: string(field.TypeRef.ID), Presence: string(field.Presence), Cardinality: string(field.Cardinality), Origin: string(field.Origin),
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
