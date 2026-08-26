package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestEntityFieldsFieldlessPublicBehaviorAndHashesRemainUnchanged(t *testing.T) {
	document := billingDocument()
	publicModel, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	supportedModel, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(publicModel, supportedModel) || SemanticFingerprint(publicModel) != SemanticFingerprint(supportedModel) {
		t.Fatal("fieldless public model changed when support was locally injected")
	}
	if err := CheckGetPut(document); err != nil {
		t.Fatal(err)
	}
	if err := validateEntityFieldsSupport(EntityFieldsSupport{State: "UNKNOWN"}); err == nil {
		t.Fatal("sanity check did not reject unknown support")
	}
}
func TestEntityFieldsSupportedBXPreservesIdentityOrderAndPutGet(t *testing.T) {
	document := latentDocument()
	support := supportedEntityFieldsForTest()
	model, err := getWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	fields := model.Nodes[0].Fields
	if len(fields) != 2 || fields[0].ID != document.Declarations[0].Fields[0].ID || fields[1].ID != document.Declarations[0].Fields[1].ID {
		t.Fatalf("field order or ID changed: %#v", fields)
	}
	for _, field := range fields {
		if field.Parent != "billing://entity/order" && field.Parent != "billing://entity/payment" {
			t.Fatalf("field parent changed: %#v", field)
		}
		if field.TypeRefUse.ResolvedID != ID(semantic.BuiltinStringTypeID) || field.Origin != FieldOriginSource || !field.Span.Valid() {
			t.Fatalf("field type or provenance was not retained: %#v", field)
		}
	}
	updated := model.Clone()
	updated.Nodes[0].Fields[0].Name = "Renamed Order Number"
	updated.Nodes[0].Fields[0].TypeRef = TypeRef{ID: semantic.BuiltinStringTypeID}
	written, err := putWithEntityFieldsSupport(document, updated, support)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := getWithEntityFieldsSupport(written, support)
	if err != nil {
		t.Fatal(err)
	}
	if observed.Nodes[0].Fields[0].ID != model.Nodes[0].Fields[0].ID || observed.Nodes[0].Fields[0].Name != "Renamed Order Number" {
		t.Fatalf("Put-Get lost field identity or presentation: %#v", observed.Nodes[0].Fields[0])
	}
	if !SemanticEquivalent(updated, observed) {
		t.Fatal("supported field Put-Get changed semantic meaning")
	}
	second, err := putWithEntityFieldsSupport(document, updated, support)
	if err != nil || !reflect.DeepEqual(second, written) {
		t.Fatalf("supported replay was not deterministic: %v %#v %#v", err, second, written)
	}
}
