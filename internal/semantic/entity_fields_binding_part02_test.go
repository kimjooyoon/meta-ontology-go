package semantic

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

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
