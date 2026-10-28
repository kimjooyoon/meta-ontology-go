package bidir

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func classifyEntityFieldsModelError(err error, span SourceSpan) error {
	if _, ok := errors.AsType[*EntityFieldsError](err); ok {
		return err
	}
	message := err.Error()
	switch {
	case strings.Contains(message, semantic.ErrAmbiguousType.Error()):
		return entityFieldsError(EntityFieldsAmbiguousTypeDiagnostic, message, span, semantic.ErrAmbiguousType)
	case strings.Contains(message, semantic.ErrUnknownType.Error()):
		return entityFieldsError(EntityFieldsUnknownTypeDiagnostic, message, span, semantic.ErrUnknownType)
	case strings.Contains(message, "parent is"):
		return entityFieldsError(EntityFieldsWrongParentDiagnostic, message, span, ErrInvalidField)
	default:
		return err
	}
}
func validateEntityFieldsProfileField(field Field, registry semantic.TypeRegistry, support EntityFieldsSupport) error {
	normalized, err := normalizeField(field, field.Parent, EntityKind, registry)
	if err != nil {
		return classifyEntityFieldsTypeError(field, err)
	}
	typeID, err := resolvedFieldTypeID(normalized, normalized.TypeRef, registry)
	if err != nil {
		return classifyEntityFieldsTypeError(field, err)
	}
	stringType := typeID == semantic.BuiltinStringTypeID
	capabilities, err := support.Profile.Capabilities()
	if err != nil {
		return err
	}
	booleanType := typeID == semantic.BuiltinBooleanTypeID && capabilities.Boolean
	integerType := typeID == semantic.BuiltinIntegerTypeID && capabilities.Integer
	if !stringType && !booleanType && !integerType {
		return entityFieldsProfileError(field, EntityFieldsUnsupportedTypeDiagnostic, fmt.Sprintf("resolved type %q is not in the bound profile", typeID))
	}
	if normalized.TextBound.IsSet() && (!capabilities.TextBound || !stringType) {
		return entityFieldsProfileError(field, EntityFieldsUnsupportedTypeDiagnostic,
			"max_utf8_bytes requires Text under the V6 profile")
	}
	validPresence := normalized.Presence == FieldPresenceRequired || capabilities.Optional && normalized.Presence == FieldPresenceOptional
	validCardinality := normalized.Cardinality == FieldCardinalityOne || capabilities.Many && normalized.Cardinality == FieldCardinalityMany
	if !validPresence || !validCardinality {
		return entityFieldsProfileError(field, EntityFieldsUnsupportedShapeDiagnostic, "the bound profile does not support this presence/cardinality shape")
	}
	return nil
}
func classifyEntityFieldsTypeError(field Field, err error) error {
	if _, ok := errors.AsType[*EntityFieldsError](err); ok {
		return err
	}
	message := err.Error()
	span := field.TypeRefSpan
	if !span.Valid() {
		span = field.Span
	}
	switch {
	case errors.Is(err, semantic.ErrAmbiguousType):
		return entityFieldsError(EntityFieldsAmbiguousTypeDiagnostic, fmt.Sprintf("field %q: %s", field.ID, message), span, err)
	case errors.Is(err, semantic.ErrUnknownType):
		return entityFieldsError(EntityFieldsUnknownTypeDiagnostic, fmt.Sprintf("field %q: %s", field.ID, message), span, err)
	default:
		return err
	}
}
func modelHasFields(nodes []Node) bool {
	for _, node := range nodes {
		if len(node.Fields) != 0 {
			return true
		}
	}
	return false
}
func firstModelFieldSpan(nodes []Node) SourceSpan {
	for _, node := range nodes {
		if len(node.Fields) != 0 {
			return node.Fields[0].Span
		}
	}
	return SourceSpan{}
}
func documentHasFields(document Document) bool {
	for _, declaration := range document.Declarations {
		if len(declaration.Fields) != 0 {
			return true
		}
	}
	return false
}
