package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func resolvedFieldTypeID(field Field, ref semantic.TypeRef, registry semantic.TypeRegistry) (semantic.ID, error) {
	typeID, err := resolvedTypeID(ref, registry)
	if err != nil {
		return "", err
	}
	if field.TypeRefUse.ResolvedID != "" {
		id, err := semantic.ParseIdentity(string(field.TypeRefUse.ResolvedID))
		if err != nil {
			return "", err
		}
		if typeID != id {
			return "", fmt.Errorf("field TypeRef ID %s disagrees with source presentation ID %s", typeID, id)
		}
	}
	return typeID, nil
}
func validateSourceField(field Field, owner ID, registry semantic.TypeRegistry) error {
	if field.Origin != FieldOriginSource {
		return fmt.Errorf("%w: field %q has non-source origin %q", ErrUnrepresentableField, field.ID, field.Origin)
	}
	if err := validateExactFieldSpans(field); err != nil {
		return fmt.Errorf("%w: field %q: %v", ErrUnrepresentableField, field.ID, err)
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
		if !item.span.Valid() {
			return fmt.Errorf("%w: %s span is not exact source provenance", ErrUnrepresentableField, item.name)
		}
	}
	normalizedUse, err := normalizeTypeRefUse(field.TypeRefUse)
	if err != nil {
		return fmt.Errorf("%w: type reference presentation: %v", ErrUnrepresentableField, err)
	}
	if normalizedUse.Form == "" || normalizedUse.Spelling == "" {
		return fmt.Errorf("%w: field %q has no original type reference presentation", ErrUnrepresentableField, field.ID)
	}
	if normalizedUse.Spelling != field.TypeRefUse.Spelling {
		return fmt.Errorf("%w: field %q type reference spelling is not normalized", ErrUnrepresentableField, field.ID)
	}
	if normalizedUse.Span != field.TypeRefSpan {
		return fmt.Errorf("%w: field %q type reference span does not match its source subspan", ErrUnrepresentableField, field.ID)
	}
	normalized, err := normalizeField(field, owner, EntityKind, registry)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnrepresentableField, err)
	}
	if field.TypeRefUse.ResolvedID != "" && field.TypeRefUse.ResolvedID != normalized.TypeRefUse.ResolvedID {
		return fmt.Errorf("%w: field %q source TypeRef resolved ID changed", ErrUnrepresentableField, field.ID)
	}
	return nil
}
