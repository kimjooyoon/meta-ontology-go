package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestLatentSyntaxFieldsLowerToSemanticIRInDeclarationOrder(t *testing.T) {
	file := latentSyntaxFile()
	support := supportedEntityFieldsForTest()
	document, err := documentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal(err)
	}
	if got := document.Declarations[0].Fields; len(got) != 2 || got[0].ID != "billing://field/order-number" || got[1].ID != "billing://field/amount" {
		t.Fatalf("field order or identity changed: %#v", got)
	}
	if got := document.Declarations[0].Fields; got[0].TypeRef.Name != "string" || got[1].TypeRef.Namespace != "gooo" || got[1].TypeRef.Name != "string" {
		t.Fatalf("type reference lookup spelling changed: %#v", got)
	}

	ir, err := lowerWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal(err)
	}
	order, ok := ir.Graph.Node(semantic.MustIdentity("billing://entity/order"))
	if !ok || len(order.Fields) != 2 {
		t.Fatalf("semantic order fields missing: %#v", order)
	}
	if order.Fields[0].ID != semantic.MustIdentity("billing://field/order-number") || order.Fields[1].ID != semantic.MustIdentity("billing://field/amount") {
		t.Fatalf("semantic field order changed: %#v", order.Fields)
	}
	if order.Fields[0].Parent != order.ID || order.Fields[1].Parent != order.ID {
		t.Fatalf("field parent identity was not explicit: %#v", order.Fields)
	}
	if order.Fields[0].TypeRef.ID != semantic.BuiltinStringTypeID || order.Fields[1].TypeRef.ID != semantic.BuiltinStringTypeID {
		t.Fatalf("registered string type was not resolved: %#v", order.Fields)
	}
	if order.Fields[0].Presence != semantic.Required || order.Fields[1].Presence != semantic.Required || order.Fields[0].Cardinality != semantic.One || order.Fields[1].Cardinality != semantic.One {
		t.Fatalf("field presence/cardinality changed: %#v", order.Fields)
	}
	if order.Fields[0].Span.File != "latent.gooo" || order.Fields[0].Span.Start.Offset != 10 || order.Fields[1].Span.End.Offset != 102 {
		t.Fatalf("field spans were not lowered: %#v", order.Fields)
	}

	payment, ok := ir.Graph.Node(semantic.MustIdentity("billing://entity/payment"))
	if !ok || len(payment.Fields) != 1 || payment.Fields[0].ID != semantic.MustIdentity("billing://field/receipt") {
		t.Fatalf("second entity fields were not lowered: %#v", payment)
	}
}
