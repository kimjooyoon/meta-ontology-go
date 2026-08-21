package bidir

import (
	"reflect"
	"strings"
	"testing"
)

func TestLatentFieldPutRejectsParentMoveWithoutWrite(t *testing.T) {
	document := latentDocument()
	model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	updated := model.Clone()
	field := updated.Nodes[0].Fields[0]
	updated.Nodes[0].Fields = updated.Nodes[0].Fields[1:]
	updated.Nodes[1].Fields = append(updated.Nodes[1].Fields, Field{
		ID: field.ID, Parent: "billing://entity/payment", Name: field.Name, Aliases: field.Aliases,
		TypeRef: field.TypeRef, Presence: field.Presence, Cardinality: field.Cardinality, Span: field.Span,
	})
	written, err := putWithEntityFieldsSupport(document, updated, supportedEntityFieldsForTest())
	if err == nil || !strings.Contains(err.Error(), "cannot move") {
		t.Fatalf("parent move result = %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("parent move changed the source document")
	}
}
func TestLatentFieldPutRejectsSemanticAdditionWithoutFieldSpan(t *testing.T) {
	document := latentDocument()
	document.Declarations[0].Span = SourceSpan{File: "latent.gooo", Start: 1, End: 2}
	model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	updated := model.Clone()
	updated.Nodes[0].Fields = append(updated.Nodes[0].Fields, Field{
		ID: "billing://field/no-span", Parent: "billing://entity/order", Name: "No Span",
		TypeRef: TypeRef{Name: "string"}, Presence: FieldPresenceRequired, Cardinality: FieldCardinalityOne,
	})
	written, err := putWithEntityFieldsSupport(document, updated, supportedEntityFieldsForTest())
	if err == nil || !strings.Contains(err.Error(), "field \"billing://field/no-span\" semantic change has no source span") {
		t.Fatalf("missing field provenance result = %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("missing field provenance changed the source document")
	}
}
