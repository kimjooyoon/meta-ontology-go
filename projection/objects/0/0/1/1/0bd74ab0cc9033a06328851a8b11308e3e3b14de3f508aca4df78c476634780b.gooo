package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"strings"
	"testing"
)

func TestLatentFieldExplicitTypeIDCannotBeMaskedByStaleTypeRefUse(t *testing.T) {
	registry := semantic.NewTypeRegistry()
	const changedTypeID semantic.ID = "urn:custom:type:changed"
	if err := registry.Register(semantic.TypeDef{ID: changedTypeID, Namespace: "custom", Name: "changed"}); err != nil {
		t.Fatal(err)
	}
	document := latentDocument()
	model, err := getWithTypesAndEntityFieldsSupport(document, registry, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	updated := model.Clone()
	updated.Nodes[0].Fields[0].TypeRef = TypeRef{ID: changedTypeID}

	if updated.Nodes[0].Fields[0].TypeRefUse.ResolvedID != ID(semantic.BuiltinStringTypeID) {
		t.Fatalf("fixture did not retain the original resolved TypeRef ID: %#v", updated.Nodes[0].Fields[0].TypeRefUse)
	}
	if SemanticEquivalent(model, updated) {
		t.Fatal("stale TypeRefUse masked an explicit semantic TypeRef.ID change")
	}
	if SemanticFingerprint(model) == SemanticFingerprint(updated) {
		t.Fatal("stale TypeRefUse masked an explicit semantic fingerprint change")
	}

	written, err := putWithTypesAndEntityFieldsSupport(document, updated, registry, supportedEntityFieldsForTest())
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
	model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}

	aggregate := model.Clone()
	aggregate.Nodes[0].Fields[0].IDSpan = SourceSpan{}
	written, err := putWithEntityFieldsSupport(document, aggregate, supportedEntityFieldsForTest())
	if err == nil || !strings.Contains(err.Error(), "exact source provenance") {
		t.Fatalf("aggregate-span readback result = %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("aggregate-span rejection changed the source document")
	}

	semanticOnly := model.Clone()
	semanticOnly.Nodes[0].Fields[0].Origin = FieldOriginSynthesized
	semanticOnly.Nodes[0].Fields[0].TypeRefUse = TypeRefUse{}
	written, err = putWithEntityFieldsSupport(document, semanticOnly, supportedEntityFieldsForTest())
	if err == nil || !strings.Contains(err.Error(), "not representable") {
		t.Fatalf("semantic-only readback result = %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("semantic-only rejection changed the source document")
	}
}
