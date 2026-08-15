package bidir

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestLatentFieldSourceTypeRefUseSurvivesRegistryRename(t *testing.T) {
	oldRegistry := semantic.TypeRegistry{}
	const typeID semantic.ID = "urn:custom:type:field"
	if err := oldRegistry.Register(semantic.TypeDef{ID: typeID, Namespace: "custom", Name: "old-name"}); err != nil {
		t.Fatal(err)
	}
	document := customTypeDocument(TypeRef{Name: "old-name", Namespace: "custom"}, TypeRefUse{Form: TypeRefFormLookup, Spelling: "custom:old-name", Span: SourceSpan{File: "custom.gooo", Start: 20, End: 29}})
	model, err := GetWithTypes(document, oldRegistry)
	if err != nil {
		t.Fatal(err)
	}
	firstWrite, err := PutWithTypes(document, model, oldRegistry)
	if err != nil {
		t.Fatal(err)
	}

	renamedRegistry := semantic.TypeRegistry{}
	if err := renamedRegistry.Register(semantic.TypeDef{ID: typeID, Namespace: "custom", Name: "new-name"}); err != nil {
		t.Fatal(err)
	}
	written, err := PutWithTypes(firstWrite, model, renamedRegistry)
	if err != nil {
		t.Fatal(err)
	}
	field := written.Declarations[0].Fields[0]
	if field.TypeRefUse.Form != TypeRefFormLookup || field.TypeRefUse.Spelling != "custom:old-name" || field.TypeRefUse.ResolvedID != ID(typeID) {
		t.Fatalf("registry rename rewrote source TypeRefUse: %#v", field.TypeRefUse)
	}

	stableDocument := customTypeDocument(TypeRef{ID: typeID}, TypeRefUse{Form: TypeRefFormStableID, Spelling: string(typeID), ResolvedID: ID(typeID), Span: SourceSpan{File: "custom.gooo", Start: 20, End: 40}})
	stableModel, err := GetWithTypes(stableDocument, renamedRegistry)
	if err != nil {
		t.Fatal(err)
	}
	stableWritten, err := PutWithTypes(stableDocument, stableModel, renamedRegistry)
	if err != nil {
		t.Fatal(err)
	}
	stableUse := stableWritten.Declarations[0].Fields[0].TypeRefUse
	if stableUse.Form != TypeRefFormStableID || stableUse.Spelling != string(typeID) || stableUse.ResolvedID != ID(typeID) {
		t.Fatalf("stable-ID source TypeRefUse changed: %#v", stableUse)
	}
}

func TestLatentFieldExplicitTypeIDCannotBeMaskedByStaleTypeRefUse(t *testing.T) {
	registry := semantic.NewTypeRegistry()
	const changedTypeID semantic.ID = "urn:custom:type:changed"
	if err := registry.Register(semantic.TypeDef{ID: changedTypeID, Namespace: "custom", Name: "changed"}); err != nil {
		t.Fatal(err)
	}
	document := latentDocument()
	model, err := GetWithTypes(document, registry)
	if err != nil {
		t.Fatal(err)
	}
	updated := model.Clone()
	updated.Nodes[0].Fields[0].TypeRef = TypeRef{ID: changedTypeID}
	// Keep the old source presentation metadata deliberately stale.
	if updated.Nodes[0].Fields[0].TypeRefUse.ResolvedID != ID(semantic.BuiltinStringTypeID) {
		t.Fatalf("fixture did not retain the original resolved TypeRef ID: %#v", updated.Nodes[0].Fields[0].TypeRefUse)
	}
	if SemanticEquivalent(model, updated) {
		t.Fatal("stale TypeRefUse masked an explicit semantic TypeRef.ID change")
	}
	if SemanticFingerprint(model) == SemanticFingerprint(updated) {
		t.Fatal("stale TypeRefUse masked an explicit semantic fingerprint change")
	}

	written, err := PutWithTypes(document, updated, registry)
	if err == nil || !strings.Contains(err.Error(), "disagrees") {
		t.Fatalf("stale TypeRefUse Put result = %v", err)
	}
	putErr, ok := err.(*PutError)
	if !ok || !putErr.NoWrite {
		t.Fatalf("stale TypeRefUse Put was not transactional: %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("stale TypeRefUse rejection changed the source document")
	}
}

func TestLatentFieldReadbackRejectsAggregateAndSemanticOnlyProvenanceWithoutWrite(t *testing.T) {
	document := latentDocument()
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}

	aggregate := model.Clone()
	aggregate.Nodes[0].Fields[0].IDSpan = SourceSpan{}
	written, err := Put(document, aggregate)
	if err == nil || !strings.Contains(err.Error(), "exact source provenance") {
		t.Fatalf("aggregate-span readback result = %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("aggregate-span rejection changed the source document")
	}

	semanticOnly := model.Clone()
	semanticOnly.Nodes[0].Fields[0].Origin = FieldOriginSynthesized
	semanticOnly.Nodes[0].Fields[0].TypeRefUse = TypeRefUse{}
	written, err = Put(document, semanticOnly)
	if err == nil || !strings.Contains(err.Error(), "not representable") {
		t.Fatalf("semantic-only readback result = %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("semantic-only rejection changed the source document")
	}
}

func TestLatentFieldNonSourceOriginsFailClosedWithoutWrite(t *testing.T) {
	for _, origin := range []FieldOrigin{FieldOriginGenerated, FieldOriginSynthesized, FieldOriginDeferred, FieldOriginUnsupported} {
		t.Run(string(origin), func(t *testing.T) {
			document := latentDocument()
			model, err := Get(document)
			if err != nil {
				t.Fatal(err)
			}
			updated := model.Clone()
			updated.Nodes[0].Fields[0].Origin = origin
			written, err := Put(document, updated)
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
				{ID: "billing://field/amount", Parent: "billing://entity/order", Name: "Amount", Aliases: []string{"total"}, TypeRef: TypeRef{Name: "string"}, TypeRefUse: TypeRefUse{Form: TypeRefFormLookup, Spelling: "gooo:string", ResolvedID: "urn:gooo:type:string", Span: SourceSpan{File: "latent.gooo", Start: 81, End: 92}}, Origin: FieldOriginSource, Presence: FieldPresenceOptional, Cardinality: FieldCardinalityMany, Span: SourceSpan{File: "latent.gooo", Start: 60, End: 102}, IDSpan: SourceSpan{File: "latent.gooo", Start: 60, End: 68}, NameSpan: SourceSpan{File: "latent.gooo", Start: 69, End: 80}, TypeRefSpan: SourceSpan{File: "latent.gooo", Start: 81, End: 92}, PresenceSpan: SourceSpan{File: "latent.gooo", Start: 93, End: 97}, CardinalitySpan: SourceSpan{File: "latent.gooo", Start: 98, End: 102}},
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
			Span: SourceSpan{File: "custom.gooo", Start: 1, End: 40}, IDSpan: SourceSpan{File: "custom.gooo", Start: 1, End: 19}, NameSpan: SourceSpan{File: "custom.gooo", Start: 20, End: 25}, TypeRefSpan: typeRefUse.Span, PresenceSpan: SourceSpan{File: "custom.gooo", Start: 30, End: 37}, CardinalitySpan: SourceSpan{File: "custom.gooo", Start: 38, End: 40},
		}}}},
	}
}
