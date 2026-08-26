package syntax

import (
	"errors"
	"testing"
)

func TestEntityFieldsSupportContractIsDeferredAndProfileBound(t *testing.T) {
	support := CurrentEntityFieldsSupport()
	if support.State != EntityFieldsDeferred || !support.State.Valid() {
		t.Fatalf("support state = %q, want DEFERRED", support.State)
	}
	wantProfile := EntityFieldsProfile{
		ID: EntityFieldsProfileID, Version: EntityFieldsProfileVersion, Digest: EntityFieldsProfileDigest,
	}
	if support.Profile != wantProfile {
		t.Fatalf("profile = %#v", support.Profile)
	}
	if err := support.Validate(); err != nil {
		t.Fatalf("checked-in support contract is invalid: %v", err)
	}
	if err := (EntityFieldsSupport{State: EntityFieldsSupported, Profile: support.Profile}).Validate(); err != nil {
		t.Fatalf("SUPPORTED is not an exhaustive valid state: %v", err)
	}
}
func TestEntityFieldsSupportRejectsUnknownStateAndProfileMismatch(t *testing.T) {
	profile := CurrentEntityFieldsSupport().Profile
	cases := []struct {
		name    string
		support EntityFieldsSupport
		want    error
	}{
		{
			name: "unknown state", support: EntityFieldsSupport{State: "UNKNOWN", Profile: profile},
			want: ErrEntityFieldsUnknownState,
		},
		{
			name: "unbound profile", support: EntityFieldsSupport{State: EntityFieldsDeferred},
			want: ErrEntityFieldsProfileMismatch,
		},
		{
			name: "profile identity",
			support: EntityFieldsSupport{State: EntityFieldsDeferred, Profile: EntityFieldsProfile{
				ID: "other", Version: 1, Digest: profile.Digest,
			}},
			want: ErrEntityFieldsProfileMismatch,
		},
		{
			name: "profile version",
			support: EntityFieldsSupport{State: EntityFieldsDeferred, Profile: EntityFieldsProfile{
				ID: profile.ID, Version: 2, Digest: profile.Digest,
			}},
			want: ErrEntityFieldsProfileMismatch,
		},
		{
			name: "profile digest",
			support: EntityFieldsSupport{State: EntityFieldsDeferred, Profile: EntityFieldsProfile{
				ID: profile.ID, Version: 1, Digest: "wrong",
			}},
			want: ErrEntityFieldsProfileMismatch,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := test.support.Validate(); !errors.Is(err, test.want) {
				t.Fatalf("validation error = %v, want %v", err, test.want)
			}
		})
	}
}
