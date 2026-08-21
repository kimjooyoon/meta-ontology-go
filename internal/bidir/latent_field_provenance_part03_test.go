package bidir

import (
	"reflect"
	"strings"
	"testing"
)

func TestLatentFieldNonSourceOriginsFailClosedWithoutWrite(t *testing.T) {
	for _, origin := range []FieldOrigin{FieldOriginGenerated, FieldOriginSynthesized, FieldOriginDeferred, FieldOriginUnsupported} {
		t.Run(string(origin), func(t *testing.T) {
			document := latentDocument()
			model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
			if err != nil {
				t.Fatal(err)
			}
			updated := model.Clone()
			updated.Nodes[0].Fields[0].Origin = origin
			written, err := putWithEntityFieldsSupport(document, updated, supportedEntityFieldsForTest())
			if err == nil || !strings.Contains(err.Error(), "not representable") {
				t.Fatalf("origin %s readback result = %v", origin, err)
			}
			if !reflect.DeepEqual(written, document) {
				t.Fatal("non-source origin rejection changed the source document")
			}
		})
	}
}
func latentDocument() Document {
	return Document{
		Package: "billing", Namespace: "billing",
		Declarations: []Declaration{
			{Kind: EntityKind, ID: "billing://entity/order", Name: "Order", Fields: []Field{
				{ID: "billing://field/order-number", Parent: "billing://entity/order", Name: "Order Number", Aliases: []string{"legacy-number", "order-no"}, TypeRef: TypeRef{Namespace: "gooo", Name: "string"}, TypeRefUse: TypeRefUse{Form: TypeRefFormLookup, Spelling: "string", ResolvedID: "urn:gooo:type:string", Span: SourceSpan{File: "latent.gooo", Start: 31, End: 37}}, Origin: FieldOriginSource, Presence: FieldPresenceRequired, Cardinality: FieldCardinalityOne, Span: SourceSpan{File: "latent.gooo", Start: 10, End: 50}, IDSpan: SourceSpan{File: "latent.gooo", Start: 10, End: 18}, NameSpan: SourceSpan{File: "latent.gooo", Start: 19, End: 30}, TypeRefSpan: SourceSpan{File: "latent.gooo", Start: 31, End: 37}, PresenceSpan: SourceSpan{File: "latent.gooo", Start: 38, End: 46}, CardinalitySpan: SourceSpan{File: "latent.gooo", Start: 47, End: 50}},
				{ID: "billing://field/amount", Parent: "billing://entity/order", Name: "Amount", Aliases: []string{"total"}, TypeRef: TypeRef{Name: "string"}, TypeRefUse: TypeRefUse{Form: TypeRefFormLookup, Spelling: "gooo:string", ResolvedID: "urn:gooo:type:string", Span: SourceSpan{File: "latent.gooo", Start: 81, End: 92}}, Origin: FieldOriginSource, Presence: FieldPresenceRequired, Cardinality: FieldCardinalityOne, Span: SourceSpan{File: "latent.gooo", Start: 60, End: 102}, IDSpan: SourceSpan{File: "latent.gooo", Start: 60, End: 68}, NameSpan: SourceSpan{File: "latent.gooo", Start: 69, End: 80}, TypeRefSpan: SourceSpan{File: "latent.gooo", Start: 81, End: 92}, PresenceSpan: SourceSpan{File: "latent.gooo", Start: 93, End: 97}, CardinalitySpan: SourceSpan{File: "latent.gooo", Start: 98, End: 102}},
			}},
			{Kind: EntityKind, ID: "billing://entity/payment", Name: "Payment", Fields: []Field{{ID: "billing://field/receipt", Parent: "billing://entity/payment", Name: "Amount", TypeRef: TypeRef{Name: "string"}, TypeRefUse: TypeRefUse{Form: TypeRefFormLookup, Spelling: "string", ResolvedID: "urn:gooo:type:string", Span: SourceSpan{File: "latent.gooo", Start: 131, End: 137}}, Origin: FieldOriginSource, Presence: FieldPresenceRequired, Cardinality: FieldCardinalityOne, Span: SourceSpan{File: "latent.gooo", Start: 110, End: 150}, IDSpan: SourceSpan{File: "latent.gooo", Start: 110, End: 118}, NameSpan: SourceSpan{File: "latent.gooo", Start: 119, End: 130}, TypeRefSpan: SourceSpan{File: "latent.gooo", Start: 131, End: 137}, PresenceSpan: SourceSpan{File: "latent.gooo", Start: 138, End: 146}, CardinalitySpan: SourceSpan{File: "latent.gooo", Start: 147, End: 150}}}},
		},
	}
}
func customTypeDocument(typeRef TypeRef, typeRefUse TypeRefUse) Document {
	return Document{
		Package: "custom", Namespace: "custom",
		Declarations: []Declaration{{Kind: EntityKind, ID: "custom://entity/item", Name: "Item", Span: SourceSpan{File: "custom.gooo", Start: 1, End: 40}, Fields: []Field{{
			ID: "custom://field/value", Parent: "custom://entity/item", Name: "Value", TypeRef: typeRef, TypeRefUse: typeRefUse, Origin: FieldOriginSource,
			Presence: FieldPresenceRequired, Cardinality: FieldCardinalityOne,
			Span: SourceSpan{File: "custom.gooo", Start: 1, End: 50}, IDSpan: SourceSpan{File: "custom.gooo", Start: 1, End: 19}, NameSpan: SourceSpan{File: "custom.gooo", Start: 20, End: 25}, TypeRefSpan: typeRefUse.Span, PresenceSpan: SourceSpan{File: "custom.gooo", Start: 36, End: 43}, CardinalitySpan: SourceSpan{File: "custom.gooo", Start: 44, End: 49},
		}}}},
	}
}
