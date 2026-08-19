package semantic

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func TestEntityFieldsDeferredProfileBindsExactSemanticFixture(t *testing.T) {
	support := syntax.CurrentEntityFieldsSupport()
	if support.State != syntax.EntityFieldsDeferred {
		t.Fatalf("syntax support state = %q, want DEFERRED", support.State)
	}
	if err := support.Validate(); err != nil {
		t.Fatalf("syntax support contract is invalid: %v", err)
	}
	binding := semanticEntityFieldsBinding(support)
	if binding.Profile != CurrentEntityFieldsProfile() {
		t.Fatalf("semantic profile = %#v, want %#v", binding.Profile, CurrentEntityFieldsProfile())
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("exact deferred binding rejected: %v", err)
	}
	assertEntityFieldsSemanticFixture(t, binding)
}
func TestEntityFieldsSupportedBindingRunsTheSameSemanticClosure(t *testing.T) {
	deferred := syntax.CurrentEntityFieldsSupport()
	supported := syntax.EntityFieldsSupport{State: syntax.EntityFieldsSupported, Profile: deferred.Profile}
	if err := supported.Validate(); err != nil {
		t.Fatalf("syntax SUPPORTED branch is not contract-valid: %v", err)
	}
	binding := semanticEntityFieldsBinding(supported)
	if err := binding.Validate(); err != nil {
		t.Fatalf("exact supported binding rejected: %v", err)
	}
	assertEntityFieldsSemanticFixture(t, binding)
}
