package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func normalizeField(field Field, owner ID, kind Kind, registry semantic.TypeRegistry) (Field, error) {
	if kind != EntityKind {
		return Field{}, fmt.Errorf("%w: fields are only valid on Entity nodes", ErrInvalidField)
	}
	spans := []struct {
		name string
		span SourceSpan
	}{
		{name: "field", span: field.Span},
		{name: "field ID", span: field.IDSpan},
		{name: "field name", span: field.NameSpan},
		{name: "field type", span: field.TypeRefSpan},
		{name: "field presence", span: field.PresenceSpan},
		{name: "field cardinality", span: field.CardinalitySpan},
	}
	for _, item := range spans {
		if err := item.span.Validate(); err != nil {
			return Field{}, fmt.Errorf("%w: %s span: %v", ErrInvalidField, item.name, err)
		}
	}
	if field.Origin != "" && !validFieldOrigin(field.Origin) {
		return Field{}, fmt.Errorf("%w: unknown field origin %q", ErrInvalidField, field.Origin)
	}
	use, err := normalizeTypeRefUse(field.TypeRefUse)
	if err != nil {
		return Field{}, fmt.Errorf("%w: type reference use: %v", ErrInvalidField, err)
	}
	normalized, err := field.semantic()
	if err != nil {
		return Field{}, err
	}
	if normalized.Parent != semantic.ID(owner) {
		return Field{}, fmt.Errorf("%w: field %s parent is %s, want %s", ErrInvalidField, normalized.ID, normalized.Parent, owner)
	}
	definition, err := resolveFieldType(normalized.TypeRef, use, registry)
	if err != nil {
		return Field{}, fmt.Errorf("%w: type ref: %v", ErrInvalidField, err)
	}
	result := field.clone()
	result.TypeRefUse = use
	result.TypeRefUse.ResolvedID = ID(definition.ID)
	return result, nil
}
