package syntax

import (
	"errors"
	"reflect"
	"strings"
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

func TestDeferredEntityFieldsRejectsWithoutPartialASTOrWrite(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order" fields {
    field Number id "billing://field/number" type string required one
}

activity PayOrder(Order) -> Order`
	filename := "entity-fields.gooo"
	original := source
	want := Diagnostic{
		Severity: SeverityError,
		Code:     DiagEntityFieldsDeferred,
		Message:  "entity fields are deferred and unsupported by the public syntax",
		Span: Span{
			Filename: filename,
			Start:    Position{Offset: 75, Line: 3, Column: 42},
			End:      Position{Offset: 81, Line: 3, Column: 48},
		},
	}

	file, diagnostics := ParseFile(filename, source)
	if file != nil || !reflect.DeepEqual(diagnostics, Diagnostics{want}) {
		t.Fatalf("deferred parse result = %#v, %#v; want no AST and %#v", file, diagnostics, want)
	}
	formatted, formatDiagnostics, err := FormatSource(filename, source)
	if formatted != "" || !reflect.DeepEqual(formatDiagnostics, diagnostics) ||
		err == nil || err.Error() != want.String() {
		t.Fatalf("deferred format result = %q, %#v, %v", formatted, formatDiagnostics, err)
	}
	secondFile, secondDiagnostics := ParseFile(filename, source)
	if secondFile != nil || !reflect.DeepEqual(secondDiagnostics, diagnostics) || source != original {
		t.Fatal("deferred rejection was not deterministic or source was changed")
	}
	if !strings.Contains(source, "fields {") {
		t.Fatal("test source lost its EntityFields marker")
	}
}

func TestFieldlessBillingSyntaxRemainsSupportedWhileFieldsAreDeferred(t *testing.T) {
	source := `package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
`
	file, diagnostics := ParseFile("billing.gooo", source)
	if len(diagnostics) != 0 || file == nil || len(file.Declarations) != 4 {
		t.Fatalf("billing parse = %#v, %#v", file, diagnostics)
	}
	formatted, err := Format(file)
	want := "package billing\nnamespace billing\n\n" +
		"entity Order id \"billing://entity/order\"\n" +
		"entity PaymentMethod id \"billing://entity/payment-method\"\n" +
		"entity Payment id \"billing://entity/payment\"\n" +
		"activity PayOrder(Order, PaymentMethod) -> Payment\n"
	if err != nil || formatted != want {
		t.Fatalf("billing format = %q, %v", formatted, err)
	}
}
