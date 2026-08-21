package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func cliEntityFieldsFixture(t *testing.T) (semantic.IR, bidir.Model) {
	t.Helper()
	const uri = "entity-fields.gooo"
	entityID := semantic.MustIdentity("billing://entity/order")
	stringID := semantic.MustIdentity("urn:gooo:type:string")
	entitySpan := semantic.Span{File: uri, Start: semantic.Position{Offset: 0, Line: 1, Column: 1}, End: semantic.Position{Offset: 160, Line: 8, Column: 1}}
	semanticField := func(id, name string, start, end int) semantic.Field {
		return semantic.Field{
			ID: semantic.ID(id), Parent: entityID, Name: name, TypeRef: semantic.TypeRef{ID: stringID}, Presence: semantic.Required, Cardinality: semantic.One,
			Span: semantic.Span{File: uri, Start: semantic.Position{Offset: start, Line: 2, Column: start + 1}, End: semantic.Position{Offset: end, Line: 2, Column: end + 1}},
		}
	}
	fields := []semantic.Field{semanticField("billing://field/order-number", "OrderNumber", 10, 60), semanticField("billing://field/customer-name", "CustomerName", 70, 120)}
	ir := semantic.NewIR("billing", "billing")
	if err := ir.AddNode(semantic.Node{ID: entityID, Kind: semantic.Entity, Namespace: "billing", Name: "Order", Fields: fields, Span: entitySpan}); err != nil {
		t.Fatal(err)
	}
	bidirField := func(id, name string, start, end int) bidir.Field {
		span := func(left, right int) bidir.SourceSpan {
			return bidir.SourceSpan{File: uri, Start: left, End: right, StartLine: 2, StartColumn: left + 1, EndLine: 2, EndColumn: right + 1}
		}
		nameStart := start + 6
		nameEnd := nameStart + len(name)
		typeStart := nameEnd + 1
		typeEnd := typeStart + 5
		presenceStart := typeEnd + 1
		presenceEnd := presenceStart + 8
		cardinalityStart := presenceEnd + 1
		cardinalityEnd := cardinalityStart + 4
		return bidir.Field{
			ID: bidir.ID(id), Parent: bidir.ID(entityID), Name: name, TypeRef: semantic.TypeRef{ID: stringID}, TypeRefUse: bidir.TypeRefUse{Form: bidir.TypeRefFormStableID, Spelling: string(stringID), ResolvedID: bidir.ID(stringID), Span: span(typeStart, typeEnd)}, Origin: bidir.FieldOriginSource, Presence: bidir.FieldPresenceRequired, Cardinality: bidir.FieldCardinalityOne,
			Span: span(start, end), IDSpan: span(start, start+5), NameSpan: span(nameStart, nameEnd), TypeRefSpan: span(typeStart, typeEnd), PresenceSpan: span(presenceStart, presenceEnd), CardinalitySpan: span(cardinalityStart, cardinalityEnd),
		}
	}
	model := bidir.Model{Package: "billing", Namespace: "billing", Nodes: []bidir.Node{{ID: bidir.ID(entityID), Kind: bidir.EntityKind, Namespace: "billing", Name: "Order", Fields: []bidir.Field{bidirField("billing://field/order-number", "OrderNumber", 10, 60), bidirField("billing://field/customer-name", "CustomerName", 70, 120)}, Span: bidir.SourceSpan{File: uri, Start: 0, End: 160, StartLine: 1, StartColumn: 1, EndLine: 8, EndColumn: 1}}}}
	return ir, model
}
