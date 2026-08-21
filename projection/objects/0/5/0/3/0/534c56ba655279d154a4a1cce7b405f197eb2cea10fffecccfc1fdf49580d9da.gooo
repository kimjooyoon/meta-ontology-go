package bidir

import (
	"errors"
	"reflect"
	"testing"
)

func TestEntityFieldsBidirContractIsExhaustivelyBound(t *testing.T) {
	current := CurrentEntityFieldsSupport()
	cases := []struct {
		name    string
		support EntityFieldsSupport
		wantErr error
	}{
		{name: "deferred", support: current},
		{name: "supported", support: supportedEntityFieldsForTest()},
		{name: "unknown-state", support: EntityFieldsSupport{State: "UNKNOWN", Profile: current.Profile}, wantErr: ErrEntityFieldsUnknownState},
		{name: "unbound", support: EntityFieldsSupport{State: EntityFieldsDeferred}, wantErr: ErrEntityFieldsUnboundProfile},
		{name: "profile-id", support: EntityFieldsSupport{State: EntityFieldsDeferred, Profile: EntityFieldsProfile{ID: "other", Version: current.Profile.Version, Digest: current.Profile.Digest}}, wantErr: ErrEntityFieldsProfileMismatch},
		{name: "profile-version", support: EntityFieldsSupport{State: EntityFieldsDeferred, Profile: EntityFieldsProfile{ID: current.Profile.ID, Version: current.Profile.Version + 1, Digest: current.Profile.Digest}}, wantErr: ErrEntityFieldsProfileMismatch},
		{name: "profile-digest", support: EntityFieldsSupport{State: EntityFieldsDeferred, Profile: EntityFieldsProfile{ID: current.Profile.ID, Version: current.Profile.Version, Digest: "wrong"}}, wantErr: ErrEntityFieldsProfileDigest},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := validateEntityFieldsSupport(test.support)
			if test.wantErr == nil {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
func TestEntityFieldsDeferredPublicBoundariesAreSourceBackedAndTransactional(t *testing.T) {
	document := latentDocument()
	before := document
	model, err := Get(document)
	assertEntityFieldsDeferred(t, err, document.Declarations[0].Fields[0].Span)
	if !reflect.DeepEqual(model, Model{}) || !reflect.DeepEqual(document, before) {
		t.Fatalf("deferred Get changed state: model=%#v document=%#v", model, document)
	}

	supported, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	written, err := Put(document, supported)
	assertEntityFieldsDeferred(t, err, document.Declarations[0].Fields[0].Span)
	if !reflect.DeepEqual(written, document) {
		t.Fatal("ordinary exported Put accepted an injected SUPPORTED model")
	}
}
