package bidir

import (
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestFieldTypeFailureRetainsLocationAndCause(t *testing.T) {
	for _, exact := range []bool{false, true} {
		field := Field{ID: "urn:type-origin:field", Parent: "urn:type-origin:entity", Name: "value",
			TypeRef: TypeRef{Name: "Missing"}, Presence: FieldPresenceRequired, Cardinality: FieldCardinalityOne,
			Span: SourceSpan{File: "origin.gooo", Start: 1, End: 70}}
		want := field.Span
		if exact {
			field.TypeRefSpan = SourceSpan{File: "origin.gooo", Start: 40, End: 47}
			want = field.TypeRefSpan
		}
		_, err := normalizeField(field, field.Parent, EntityKind, semantic.NewTypeRegistry())
		var located *EntityFieldsError
		if !errors.As(err, &located) || located.Code != EntityFieldsUnknownTypeDiagnostic || located.Span != want ||
			!errors.Is(err, semantic.ErrUnknownType) || !errors.Is(err, ErrInvalidField) {
			t.Fatal("field failure must retain its type cause and available source range", exact, err)
		}
	}
}
