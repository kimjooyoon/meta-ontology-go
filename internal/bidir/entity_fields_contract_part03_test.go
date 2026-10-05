package bidir

import (
	"context"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"strings"
	"testing"
)

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

func TestEntityFieldsV2AddsBooleanWhileV1RemainsBounded(t *testing.T) {
	const source = `package records
namespace records
entity Gate id "records://gate" fields {
  field enabled id "records://gate/enabled" type boolean required one
}`
	v1 := syntax.EntityFieldsV1Support()
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("gate.gooo", source, v1)
	if diagnostics.HasErrors() {
		t.Fatal("field syntax should parse before profile type validation", diagnostics)
	}
	if _, err := LowerContextWithEntityFieldsSupport(context.Background(), file, v1); err == nil ||
		!strings.Contains(err.Error(), EntityFieldsUnsupportedTypeDiagnostic) {
		t.Fatalf("V1 accepted Boolean field: %v", err)
	}
	v2 := syntax.EntityFieldsV2Support()
	file, diagnostics = syntax.ParseFileWithEntityFieldsSupport("gate.gooo", source, v2)
	if diagnostics.HasErrors() {
		t.Fatal("V2 parse", diagnostics)
	}
	ir, err := LowerContextWithEntityFieldsSupport(context.Background(), file, v2)
	if err != nil {
		t.Fatal("V2 lowering", err)
	}
	entity, found := ir.Graph.NodeByName(ir.Namespace, "Gate")
	if !found || len(entity.Fields) != 1 || entity.Fields[0].TypeRef.ID != semantic.BuiltinBooleanTypeID {
		t.Fatalf("resolved Boolean field missing from semantic graph: %+v", entity)
	}
}

func TestEntityFieldsV3AddsIntegerFields(t *testing.T) {
	const source = `package records
namespace records
entity Counter id "records://counter" fields {
  field total id "records://counter/total" type integer required one
}`
	support := syntax.EntityFieldsV3Support()
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("counter.gooo", source, support)
	if diagnostics.HasErrors() {
		t.Fatal("V3 field syntax", diagnostics)
	}
	ir, err := LowerContextWithEntityFieldsSupport(context.Background(), file, support)
	if err != nil {
		t.Fatal("V3 integer lowering", err)
	}
	entity, found := ir.Graph.NodeByName(ir.Namespace, "Counter")
	if !found || len(entity.Fields) != 1 || entity.Fields[0].TypeRef.ID != semantic.BuiltinIntegerTypeID {
		t.Fatalf("resolved integer field missing from semantic graph: %+v", entity)
	}
}
func assertEntityFieldsDeferred(t *testing.T, err error, span SourceSpan) {
	t.Helper()
	var deferred *EntityFieldsError
	if !errors.As(err, &deferred) || deferred.Code != EntityFieldsDeferredDiagnostic || !errors.Is(err, ErrEntityFieldsDeferred) || deferred.Span != span {
		t.Fatalf("deferred error = %v, want source-backed %s at %#v", err, EntityFieldsDeferredDiagnostic, span)
	}
}
