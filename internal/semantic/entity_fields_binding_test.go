package semantic

import (
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

func TestEntityFieldsBindingFailsClosedForUnknownStateAndProfileMismatch(t *testing.T) {
	support := syntax.CurrentEntityFieldsSupport()
	cases := []struct {
		name    string
		binding EntityFieldsBinding
		want    error
	}{
		{
			name:    "unknown state",
			binding: EntityFieldsBinding{State: "UNKNOWN", Profile: semanticEntityFieldsProfile(support.Profile)},
			want:    ErrEntityFieldsUnknownState,
		},
		{
			name: "profile ID mismatch",
			binding: EntityFieldsBinding{
				State:   EntityFieldsDeferred,
				Profile: EntityFieldsProfile{ID: "other", Version: support.Profile.Version, Digest: support.Profile.Digest},
			},
			want: ErrEntityFieldsProfileMismatch,
		},
		{
			name: "profile version mismatch",
			binding: EntityFieldsBinding{
				State:   EntityFieldsDeferred,
				Profile: EntityFieldsProfile{ID: support.Profile.ID, Version: support.Profile.Version + 1, Digest: support.Profile.Digest},
			},
			want: ErrEntityFieldsProfileMismatch,
		},
		{
			name: "profile digest mismatch",
			binding: EntityFieldsBinding{
				State:   EntityFieldsDeferred,
				Profile: EntityFieldsProfile{ID: support.Profile.ID, Version: support.Profile.Version, Digest: "wrong"},
			},
			want: ErrEntityFieldsProfileMismatch,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := test.binding.Validate(); !errors.Is(err, test.want) {
				t.Fatalf("binding error = %v, want %v", err, test.want)
			}
		})
	}
}

func semanticEntityFieldsBinding(support syntax.EntityFieldsSupport) EntityFieldsBinding {
	return EntityFieldsBinding{State: EntityFieldsState(support.State), Profile: semanticEntityFieldsProfile(support.Profile)}
}

func semanticEntityFieldsProfile(profile syntax.EntityFieldsProfile) EntityFieldsProfile {
	return EntityFieldsProfile{ID: profile.ID, Version: profile.Version, Digest: profile.Digest}
}

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
