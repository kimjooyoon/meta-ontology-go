package generator

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func supportedEntityFieldsForTest() entityFieldsSupport {
	support := checkedEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	return support
}
func entityFieldsFixture() SemanticIR {
	return SemanticIR{
		Package: "entityfieldsgen",
		Entities: []Entity{{
			ID: "urn:gooo:entity:order", Name: "Order", GoName: "Order",
			Source: SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 100, Line: 8, Column: 1}},
			Fields: []Field{
				{
					ID: "urn:gooo:field:order-number", Parent: "urn:gooo:entity:order", Name: "OrderNumber",
					TypeRefID: entityFieldsStringTypeID, Presence: "required", Cardinality: "one",
					Source:          SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 20, Line: 4, Column: 5}, End: Position{Offset: 38, Line: 4, Column: 23}},
					IDSpan:          SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 20, Line: 4, Column: 5}, End: Position{Offset: 22, Line: 4, Column: 7}},
					NameSpan:        SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 23, Line: 4, Column: 8}, End: Position{Offset: 25, Line: 4, Column: 10}},
					TypeRefSpan:     SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 26, Line: 4, Column: 11}, End: Position{Offset: 28, Line: 4, Column: 13}},
					PresenceSpan:    SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 29, Line: 4, Column: 14}, End: Position{Offset: 32, Line: 4, Column: 17}},
					CardinalitySpan: SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 33, Line: 4, Column: 18}, End: Position{Offset: 37, Line: 4, Column: 22}},
				},
				{
					ID: "urn:gooo:field:customer-name", Parent: "urn:gooo:entity:order", Name: "CustomerName",
					TypeRefID: entityFieldsStringTypeID, Presence: "required", Cardinality: "one",
					Source:          SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 40, Line: 5, Column: 5}, End: Position{Offset: 58, Line: 5, Column: 23}},
					IDSpan:          SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 40, Line: 5, Column: 5}, End: Position{Offset: 42, Line: 5, Column: 7}},
					NameSpan:        SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 43, Line: 5, Column: 8}, End: Position{Offset: 45, Line: 5, Column: 10}},
					TypeRefSpan:     SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 46, Line: 5, Column: 11}, End: Position{Offset: 48, Line: 5, Column: 13}},
					PresenceSpan:    SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 49, Line: 5, Column: 14}, End: Position{Offset: 52, Line: 5, Column: 17}},
					CardinalitySpan: SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 53, Line: 5, Column: 18}, End: Position{Offset: 57, Line: 5, Column: 22}},
				},
			},
		}},
		Activities: []Activity{{
			ID: "urn:gooo:activity:load-order", Name: "LoadOrder", GoName: "LoadOrder",
			Outputs: []Port{{ID: "urn:gooo:entity:order", EntityID: "urn:gooo:entity:order", Name: "order", GoName: "order", GoType: "Order"}},
			Slots:   []Slot{{ID: "urn:gooo:slot:load-order", Default: "return Order{}"}},
		}},
	}
}
func supportedEntityFieldsResult(t *testing.T, ir SemanticIR, previous []byte) Result {
	t.Helper()
	result, err := New(Options{}).generateWithEntityFieldsSupport(ir, previous, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	return result
}
