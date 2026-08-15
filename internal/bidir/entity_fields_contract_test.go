package bidir

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestEntityFieldsFieldlessPublicBehaviorAndHashesRemainUnchanged(t *testing.T) {
	document := billingDocument()
	publicModel, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	supportedModel, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(publicModel, supportedModel) || SemanticFingerprint(publicModel) != SemanticFingerprint(supportedModel) {
		t.Fatal("fieldless public model changed when support was locally injected")
	}
	if err := CheckGetPut(document); err != nil {
		t.Fatal(err)
	}
	if err := validateEntityFieldsSupport(EntityFieldsSupport{State: "UNKNOWN"}); err == nil {
		t.Fatal("sanity check did not reject unknown support")
	}
}

func TestEntityFieldsSupportedBXPreservesIdentityOrderAndPutGet(t *testing.T) {
	document := latentDocument()
	support := supportedEntityFieldsForTest()
	model, err := getWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	fields := model.Nodes[0].Fields
	if len(fields) != 2 || fields[0].ID != document.Declarations[0].Fields[0].ID || fields[1].ID != document.Declarations[0].Fields[1].ID {
		t.Fatalf("field order or ID changed: %#v", fields)
	}
	for _, field := range fields {
		if field.Parent != "billing://entity/order" && field.Parent != "billing://entity/payment" {
			t.Fatalf("field parent changed: %#v", field)
		}
		if field.TypeRefUse.ResolvedID != ID(semantic.BuiltinStringTypeID) || field.Origin != FieldOriginSource || !field.Span.Valid() {
			t.Fatalf("field type or provenance was not retained: %#v", field)
		}
	}
	updated := model.Clone()
	updated.Nodes[0].Fields[0].Name = "Renamed Order Number"
	updated.Nodes[0].Fields[0].TypeRef = TypeRef{ID: semantic.BuiltinStringTypeID}
	written, err := putWithEntityFieldsSupport(document, updated, support)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := getWithEntityFieldsSupport(written, support)
	if err != nil {
		t.Fatal(err)
	}
	if observed.Nodes[0].Fields[0].ID != model.Nodes[0].Fields[0].ID || observed.Nodes[0].Fields[0].Name != "Renamed Order Number" {
		t.Fatalf("Put-Get lost field identity or presentation: %#v", observed.Nodes[0].Fields[0])
	}
	if !SemanticEquivalent(updated, observed) {
		t.Fatal("supported field Put-Get changed semantic meaning")
	}
	second, err := putWithEntityFieldsSupport(document, updated, support)
	if err != nil || !reflect.DeepEqual(second, written) {
		t.Fatalf("supported replay was not deterministic: %v %#v %#v", err, second, written)
	}
}

func TestEntityFieldsSupportedProfileRejectsUnsupportedInputs(t *testing.T) {
	support := supportedEntityFieldsForTest()
	cases := []struct {
		name   string
		mutate func(*Document)
		want   string
	}{
		{name: "optional-one", mutate: func(document *Document) { document.Declarations[0].Fields[0].Presence = FieldPresenceOptional }, want: EntityFieldsUnsupportedShapeDiagnostic},
		{name: "required-many", mutate: func(document *Document) { document.Declarations[0].Fields[0].Cardinality = FieldCardinalityMany }, want: EntityFieldsUnsupportedShapeDiagnostic},
		{name: "cross-kind-id", mutate: func(document *Document) {
			document.Declarations = append(document.Declarations, Declaration{Kind: ActivityKind, ID: "billing://activity/pay", Name: "Pay"})
			document.Declarations[0].Fields[0].ID = "billing://activity/pay"
		}, want: EntityFieldsIDCollisionDiagnostic},
		{name: "cross-snapshot", mutate: func(document *Document) { document.Declarations[0].Fields[0].NameSpan.File = "other.gooo" }, want: "cross source snapshots"},
		{name: "illegal-reorder", mutate: func(document *Document) {
			document.Declarations[0].Fields[0], document.Declarations[0].Fields[1] = document.Declarations[0].Fields[1], document.Declarations[0].Fields[0]
		}, want: EntityFieldsIllegalReorderDiagnostic},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			document := latentDocument()
			test.mutate(&document)
			before := document
			model, err := getWithEntityFieldsSupport(document, support)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
			if !reflect.DeepEqual(model, Model{}) || !reflect.DeepEqual(document, before) {
				t.Fatal("rejected field input produced partial model or mutated source")
			}
		})
	}

	unknown := latentDocument()
	unknown.Declarations[0].Fields[0].TypeRef = TypeRef{ID: "billing://type/missing"}
	if _, err := getWithEntityFieldsSupport(unknown, support); err == nil || !strings.Contains(err.Error(), EntityFieldsUnknownTypeDiagnostic) {
		t.Fatalf("unknown type error = %v", err)
	}
	customRegistry := semantic.NewTypeRegistry()
	customType := semantic.TypeDef{ID: "billing://type/custom", Namespace: "billing", Name: "Custom"}
	if err := customRegistry.Register(customType); err != nil {
		t.Fatal(err)
	}
	unsupported := latentDocument()
	unsupported.Declarations[0].Fields[0].TypeRef = TypeRef{ID: customType.ID}
	unsupported.Declarations[0].Fields[0].TypeRefUse = TypeRefUse{Form: TypeRefFormStableID, Spelling: string(customType.ID), ResolvedID: ID(customType.ID), Span: unsupported.Declarations[0].Fields[0].TypeRefSpan}
	if _, err := getWithTypesAndEntityFieldsSupport(unsupported, customRegistry, support); err == nil || !strings.Contains(err.Error(), EntityFieldsUnsupportedTypeDiagnostic) {
		t.Fatalf("unprofiled type error = %v", err)
	}
}

func assertEntityFieldsDeferred(t *testing.T, err error, span SourceSpan) {
	t.Helper()
	var deferred *EntityFieldsError
	if !errors.As(err, &deferred) || deferred.Code != EntityFieldsDeferredDiagnostic || !errors.Is(err, ErrEntityFieldsDeferred) || deferred.Span != span {
		t.Fatalf("deferred error = %v, want source-backed %s at %#v", err, EntityFieldsDeferredDiagnostic, span)
	}
}
