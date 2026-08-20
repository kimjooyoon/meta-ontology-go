package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"strings"
	"testing"
)

func TestLatentFieldValidationRejectsDeterministicallyWithoutPartialModel(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Document)
		want   string
	}{
		{name: "duplicate-global-id", mutate: func(document *Document) {
			document.Declarations[1].Fields[0].ID = document.Declarations[0].Fields[0].ID
		}, want: "duplicate field ID"},
		{name: "same-parent-name-alias-collision", mutate: func(document *Document) {
			document.Declarations[0].Fields[1].Name = "legacy-number"
		}, want: "field name"},
		{name: "unknown-type", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].TypeRef = semantic.TypeRef{Name: "missing"}
		}, want: "unknown semantic type"},
		{name: "invalid-presence", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].Presence = FieldPresence("sometimes")
		}, want: "unknown presence"},
		{name: "invalid-cardinality", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].Cardinality = FieldCardinality("unordered")
		}, want: "unknown cardinality"},
		{name: "non-entity-owner", mutate: func(document *Document) {
			document.Declarations[0].Kind = ActivityKind
		}, want: "only valid on Entity"},
		{name: "wrong-parent", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].Parent = "billing://entity/payment"
		}, want: "parent"},
		{name: "invalid-type-ref", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].TypeRef = semantic.TypeRef{ID: "not-an-identity"}
		}, want: "invalid semantic type reference"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			document := latentDocument()
			test.mutate(&document)
			beforeGet := document
			model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Get error = %v, want substring %q", err, test.want)
			}
			if !reflect.DeepEqual(model, Model{}) {
				t.Fatalf("failed Get returned partial model: %#v", model)
			}
			if !reflect.DeepEqual(document, beforeGet) {
				t.Fatal("failed Get mutated the source document")
			}
		})
	}
}
