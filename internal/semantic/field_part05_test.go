package semantic

import (
	"testing"
)

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
