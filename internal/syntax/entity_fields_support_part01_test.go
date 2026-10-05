package syntax

import (
	"crypto/sha256"
	"encoding/hex"
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

func TestEntityFieldsV2ProfileIsDistinctAndValidated(t *testing.T) {
	v1 := EntityFieldsV1Support()
	v2 := EntityFieldsV2Support()
	if err := v1.Validate(); err != nil {
		t.Fatal("V1 support changed", err)
	}
	if err := v2.Validate(); err != nil {
		t.Fatal("V2 support invalid", err)
	}
	if v1.Profile == v2.Profile || v2.Profile.ID != EntityFieldsV2ProfileID || v2.Profile.Version != 2 {
		t.Fatalf("V2 profile did not create a distinct contract: v1=%+v v2=%+v", v1.Profile, v2.Profile)
	}
	canonical := "id=gooo.entityfields.go-projection.v2\ntypes=urn:gooo:type:string,urn:gooo:type:boolean\npresence=required\ncardinality=one\nordering=source\n"
	digest := sha256.Sum256([]byte(canonical))
	if hex.EncodeToString(digest[:]) != EntityFieldsV2ProfileDigest {
		t.Fatal("V2 profile digest does not match its canonical contract")
	}
}

func TestEntityFieldsV3ProfileBindsIntegerType(t *testing.T) {
	v3 := EntityFieldsV3Support()
	if err := v3.Validate(); err != nil {
		t.Fatal("V3 support invalid", err)
	}
	canonical := "id=gooo.entityfields.go-projection.v3\ntypes=urn:gooo:type:string,urn:gooo:type:boolean,urn:gooo:type:integer\npresence=required\ncardinality=one\nordering=source\n"
	digest := sha256.Sum256([]byte(canonical))
	if hex.EncodeToString(digest[:]) != EntityFieldsV3ProfileDigest {
		t.Fatal("V3 profile digest does not match its canonical contract")
	}
}
