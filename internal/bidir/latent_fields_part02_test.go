package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestLatentFieldsGetPutPreservesAliasesSpansAndOrder(t *testing.T) {
	document := latentDocument()
	before := document
	model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Nodes[0].Fields) != 2 || model.Nodes[0].Fields[0].Name != "Order Number" || model.Nodes[0].Fields[1].Name != "Amount" {
		t.Fatalf("model field order changed: %#v", model.Nodes[0].Fields)
	}
	written, err := putWithEntityFieldsSupport(document, model, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatalf("field carrier changed during Get-Put:\n got %#v\nwant %#v", written, document)
	}
	observed, err := getWithEntityFieldsSupport(written, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(observed, model) || !SemanticEquivalent(model, observed) {
		t.Fatalf("field readback changed model:\n got %#v\nwant %#v", observed, model)
	}
	if !reflect.DeepEqual(document, before) {
		t.Fatal("Get-Put mutated the source document")
	}

	clone := model.Clone()
	clone.Nodes[0].Fields[0].Aliases[0] = "changed"
	if model.Nodes[0].Fields[0].Aliases[0] == "changed" || document.Declarations[0].Fields[0].Aliases[0] == "changed" {
		t.Fatal("field aliases were not deep-cloned")
	}
	written.Declarations[0].Fields[0].Aliases[0] = "written-only"
	if model.Nodes[0].Fields[0].Aliases[0] == "written-only" || document.Declarations[0].Fields[0].Aliases[0] == "written-only" {
		t.Fatal("readback field aliases retained caller storage")
	}
}
func TestLatentFieldSemanticIdentityIgnoresPresentationAndSpan(t *testing.T) {
	left, err := getWithEntityFieldsSupport(latentDocument(), supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	right := left.Clone()
	right.Nodes[0].Fields[0].Name = "Renamed presentation"
	right.Nodes[0].Fields[0].Aliases = []string{"new alias"}
	right.Nodes[0].Fields[0].Span = SourceSpan{File: "other.gooo", Start: 100, End: 120}
	right.Nodes[0].Fields[0].TypeRef = semantic.TypeRef{ID: semantic.BuiltinStringTypeID}
	right.Nodes[0].Fields[0].TypeRefUse.Spelling = "renamed:display"
	right.Nodes[0].Fields[0].TypeRefUse.Form = TypeRefFormLookup
	if !SemanticEquivalent(left, right) || SemanticFingerprint(left) != SemanticFingerprint(right) {
		t.Fatal("field presentation or span changed semantic identity")
	}
	right.Nodes[0].Fields[0].Cardinality = FieldCardinalityMany
	if SemanticEquivalent(left, right) || SemanticFingerprint(left) == SemanticFingerprint(right) {
		t.Fatal("field semantic cardinality change was ignored")
	}
}
