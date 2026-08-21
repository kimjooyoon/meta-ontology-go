package semantic

import (
	"testing"
)

func assertEntityFieldsSemanticFixture(t *testing.T, binding EntityFieldsBinding) {
	t.Helper()
	if binding.State != EntityFieldsDeferred && binding.State != EntityFieldsSupported {
		t.Fatalf("fixture received unexpected state %q", binding.State)
	}
	entityID := MustIdentity("billing://entity/order")
	field := testStringField(entityID, "order-number")
	entity := testEntityWithFields(t, entityID, field)
	graph := NewGraph()
	if err := graph.AddNode(entity); err != nil {
		t.Fatalf("latent field fixture rejected: %v", err)
	}
	if err := graph.ValidateWithTypes(NewTypeRegistry()); err != nil {
		t.Fatalf("latent field fixture failed semantic validation: %v", err)
	}
	if got := graph.Nodes()[0].Fields[0].ID; got != field.ID {
		t.Fatalf("fixture field ID = %s, want %s", got, field.ID)
	}
}
