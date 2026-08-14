package semantic

import (
	"errors"
	"strings"
	"testing"
)

func testStringField(parent ID, suffix string) Field {
	return Field{
		ID:          MustIdentity("billing://field/" + suffix),
		Parent:      parent,
		Name:        suffix,
		TypeRef:     TypeRef{ID: BuiltinStringTypeID},
		Presence:    Required,
		Cardinality: One,
	}
}

func testEntityWithFields(t *testing.T, id ID, fields ...Field) Node {
	t.Helper()
	node := mustEntity(t, id, Namespace("billing"), "Order")
	node.Fields = fields
	return node
}

func TestFieldRequiresExplicitIdentityAndEntityParent(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	field := testStringField(entityID, "name")
	graph := NewGraph()
	if err := graph.AddNode(testEntityWithFields(t, entityID, field)); err != nil {
		t.Fatalf("valid entity field rejected: %v", err)
	}
	node, ok := graph.Node(entityID)
	if !ok || len(node.Fields) != 1 {
		t.Fatalf("stored fields = %#v, ok=%v", node.Fields, ok)
	}
	if node.Fields[0].ID != field.ID || node.Fields[0].Parent != entityID {
		t.Fatalf("stored field identity = %#v", node.Fields[0])
	}
	if err := graph.ValidateWithTypes(NewTypeRegistry()); err != nil {
		t.Fatalf("valid typed field graph rejected: %v", err)
	}

	badID := field
	badID.ID = "not an absolute identity"
	if err := (&Graph{}).AddNode(testEntityWithFields(t, entityID, badID)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("invalid field ID error = %v, want ErrInvalidField", err)
	}

	activity := mustActivity(t, MustIdentity("billing://activity/pay"), Namespace("billing"), "Pay")
	activity.Fields = []Field{field}
	if err := (&Graph{}).AddNode(activity); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("activity field error = %v, want ErrInvalidField", err)
	}
}

func TestFieldIDsAndOwnerNamesAreValidatedWithoutCrossParentNameCollisions(t *testing.T) {
	firstID := MustIdentity("billing://entity/first")
	secondID := MustIdentity("billing://entity/second")

	duplicateID := testEntityWithFields(t, firstID,
		testStringField(firstID, "same-id"),
		testStringField(firstID, "same-id"),
	)
	if err := (&Graph{}).AddNode(duplicateID); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("duplicate field ID error = %v, want ErrInvalidField", err)
	}

	nameCollision := testEntityWithFields(t, firstID,
		testStringField(firstID, "first"),
		func() Field {
			field := testStringField(firstID, "second")
			field.Name = "first"
			return field
		}(),
	)
	if err := (&Graph{}).AddNode(nameCollision); !errors.Is(err, ErrNameCollision) {
		t.Fatalf("same-owner name collision error = %v, want ErrNameCollision", err)
	}

	aliasCollision := testEntityWithFields(t, firstID,
		testStringField(firstID, "alias-first"),
		func() Field {
			field := testStringField(firstID, "alias-second")
			field.Aliases = []string{"alias-first"}
			return field
		}(),
	)
	if err := (&Graph{}).AddNode(aliasCollision); !errors.Is(err, ErrNameCollision) {
		t.Fatalf("same-owner alias collision error = %v, want ErrNameCollision", err)
	}

	graph := NewGraph()
	firstField := testStringField(firstID, "name")
	secondField := testStringField(secondID, "other-name")
	secondField.Name = "name"
	if err := graph.AddNode(testEntityWithFields(t, firstID, firstField)); err != nil {
		t.Fatal(err)
	}
	secondNode := testEntityWithFields(t, secondID, secondField)
	secondNode.Name = "Receipt"
	if err := graph.AddNode(secondNode); err != nil {
		t.Fatalf("cross-parent equal field names rejected: %v", err)
	}
}

func TestFieldParentMismatchAndMoveAreRejectedAtomically(t *testing.T) {
	firstID := MustIdentity("billing://entity/first")
	secondID := MustIdentity("billing://entity/second")
	field := testStringField(firstID, "name")

	graph := NewGraph()
	if err := graph.AddNode(testEntityWithFields(t, firstID, field)); err != nil {
		t.Fatal(err)
	}
	before := graph.Canonical()
	mismatched := testStringField(secondID, "mismatch")
	mismatched.ID = field.ID
	if err := graph.AddNode(testEntityWithFields(t, secondID, mismatched)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("moved field error = %v, want ErrInvalidField", err)
	}
	if graph.Canonical() != before {
		t.Fatal("rejected moved field mutated graph")
	}

	parentMismatch := field
	parentMismatch.Parent = secondID
	if err := (&Graph{}).AddNode(testEntityWithFields(t, firstID, parentMismatch)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("parent mismatch error = %v, want ErrInvalidField", err)
	}
}

func TestFieldTypeRegistryResolvesStableIDsAndRejectsUnknownAmbiguousOrInvalidRefs(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	registry := NewTypeRegistry()
	valid := testEntityWithFields(t, entityID, testStringField(entityID, "name"))
	graph := NewGraph()
	if err := graph.AddNode(valid); err != nil {
		t.Fatal(err)
	}
	if err := graph.ValidateWithTypes(registry); err != nil {
		t.Fatalf("built-in string type rejected: %v", err)
	}

	unknown := testStringField(entityID, "unknown")
	unknown.TypeRef = TypeRef{ID: MustIdentity("billing://type/unknown")}
	unknownGraph := NewGraph()
	if err := unknownGraph.AddNode(testEntityWithFields(t, entityID, unknown)); err != nil {
		t.Fatal(err)
	}
	if err := unknownGraph.ValidateWithTypes(registry); !errors.Is(err, ErrUnknownType) {
		t.Fatalf("unknown type error = %v, want ErrUnknownType", err)
	}

	if err := registry.Register(TypeDef{ID: MustIdentity("alt://type/string"), Namespace: Namespace("alt"), Name: "string"}); err != nil {
		t.Fatal(err)
	}
	ambiguous := testStringField(entityID, "ambiguous")
	ambiguous.TypeRef = TypeRef{Name: "string"}
	ambiguousGraph := NewGraph()
	if err := ambiguousGraph.AddNode(testEntityWithFields(t, entityID, ambiguous)); err != nil {
		t.Fatal(err)
	}
	if err := ambiguousGraph.ValidateWithTypes(registry); !errors.Is(err, ErrAmbiguousType) {
		t.Fatalf("ambiguous type error = %v, want ErrAmbiguousType", err)
	}

	invalid := testStringField(entityID, "invalid")
	invalid.TypeRef = TypeRef{ID: "not an identity"}
	if err := (&Graph{}).AddNode(testEntityWithFields(t, entityID, invalid)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("invalid type ref error = %v, want ErrInvalidField", err)
	}

	lookup := testStringField(entityID, "lookup")
	lookup.TypeRef = TypeRef{Name: BuiltinStringTypeName, Namespace: BuiltinTypeNamespace}
	lookupGraph := NewGraph()
	if err := lookupGraph.AddNode(testEntityWithFields(t, entityID, lookup)); err != nil {
		t.Fatal(err)
	}
	normalized, err := lookupGraph.NormalizedWithTypes(NewTypeRegistry())
	if err != nil {
		t.Fatalf("lookup TypeRef normalization failed: %v", err)
	}
	got := normalized.Nodes()[0].Fields[0].TypeRef
	if got.ID != BuiltinStringTypeID || got.Name != "" || got.Namespace != "" {
		t.Fatalf("lookup metadata was not reduced to stable ID: %#v", got)
	}
}

func TestFieldsPreserveSourceOrderNormalizeIdempotentlyAndExcludePresentationFromSemanticHash(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	first := testStringField(entityID, "first")
	first.Name = " First  Name "
	first.Aliases = []string{"Primary", " Primary "}
	first.Span = Span{File: "one.gooo", Start: Position{Offset: 4}, End: Position{Offset: 10}}
	second := testStringField(entityID, "second")
	second.Name = "Second Name"

	graph := NewGraph()
	if err := graph.AddNode(testEntityWithFields(t, entityID, first, second)); err != nil {
		t.Fatal(err)
	}
	normalized, err := graph.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	fields := normalized.Nodes()[0].Fields
	if len(fields) != 2 || fields[0].ID != first.ID || fields[1].ID != second.ID {
		t.Fatalf("field source order was not preserved: %#v", fields)
	}
	if fields[0].Name != "First Name" || len(fields[0].Aliases) != 1 || fields[0].Aliases[0] != "Primary" {
		t.Fatalf("field presentation was not normalized: %#v", fields[0])
	}
	again, err := normalized.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Canonical() != again.Canonical() {
		t.Fatal("field normalization is not idempotent")
	}

	left := fields[0]
	right := left
	right.Name = "Renamed"
	right.Aliases = []string{"Other"}
	right.Span = Span{File: "two.gooo", Start: Position{Offset: 20}, End: Position{Offset: 30}}
	if left.StableHash() != right.StableHash() {
		t.Fatal("field presentation or span changed semantic identity")
	}
	if left.Canonical() == right.Canonical() {
		t.Fatal("field canonical form discarded presentation metadata")
	}
}

func TestFieldChangesAreSemanticAndFactsRemainSeparate(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	activityID := MustIdentity("billing://activity/pay")
	base := testEntityWithFields(t, entityID, testStringField(entityID, "name"))
	changed := base
	changed.Fields = copyFields(base.Fields)
	changed.Fields[0].Cardinality = Many
	if base.StableHash() == changed.StableHash() {
		t.Fatal("field cardinality did not participate in semantic identity")
	}

	graph := NewGraph()
	if err := graph.AddNode(base); err != nil {
		t.Fatal(err)
	}
	if err := graph.AddNode(mustActivity(t, activityID, Namespace("billing"), "Pay")); err != nil {
		t.Fatal(err)
	}
	fact := NewCandidateFact(activityID, Used, entityID, "candidate evidence")
	if err := graph.AddCandidate(fact); err != nil {
		t.Fatal(err)
	}
	if len(graph.Facts()) != 0 || len(graph.Candidates()) != 1 {
		t.Fatalf("field addition crossed fact boundary: facts=%d candidates=%d", len(graph.Facts()), len(graph.Candidates()))
	}
	if strings.Contains(graph.SemanticCanonical(), "candidate evidence") {
		t.Fatal("candidate explanation leaked into semantic canonical form")
	}
}

func TestInvalidFieldAdditionDoesNotMutateOriginalIR(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	ir := NewIR("billing", Namespace("billing"))
	if err := ir.AddNode(mustEntity(t, entityID, Namespace("billing"), "Order")); err != nil {
		t.Fatal(err)
	}
	beforeCanonical := ir.Canonical()
	beforeHash := ir.StableHash()
	bad := testStringField(entityID, "bad")
	bad.Parent = MustIdentity("billing://entity/other")
	if err := ir.AddNode(testEntityWithFields(t, entityID, bad)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("invalid field error = %v, want ErrInvalidField", err)
	}
	if ir.Canonical() != beforeCanonical || ir.StableHash() != beforeHash {
		t.Fatal("rejected field addition mutated original IR")
	}
}

func TestFieldsAreLatentUntilDirectSemanticConstruction(t *testing.T) {
	// The public syntax/lowering packages are intentionally unchanged in this
	// slice. Their existing constructors still produce field-free Nodes; a
	// field can only be reached here through direct semantic IR construction.
	node, err := NewEntity(MustIdentity("billing://entity/order"), Namespace("billing"), "Order")
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Fields) != 0 {
		t.Fatal("field semantics became reachable from the existing node constructor")
	}
}
