package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func classifyEntityFieldsModelError(err error, span SourceSpan) error {
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
func validateEntityFieldsProfileField(field Field, registry semantic.TypeRegistry) error {
	normalized, err := normalizeField(field, field.Parent, EntityKind, registry)
	if err != nil {
		return classifyEntityFieldsTypeError(field, err)
	}
	typeID, err := resolvedFieldTypeID(normalized, normalized.TypeRef, registry)
	if err != nil {
		return classifyEntityFieldsTypeError(field, err)
	}
	if typeID != semantic.BuiltinStringTypeID {
		return entityFieldsProfileError(field, EntityFieldsUnsupportedTypeDiagnostic, fmt.Sprintf("resolved type %q is not in the bound profile", typeID))
	}
	if normalized.Presence != FieldPresenceRequired || normalized.Cardinality != FieldCardinalityOne {
		return entityFieldsProfileError(field, EntityFieldsUnsupportedShapeDiagnostic, "only required × one is in the bound profile")
	}
	return nil
}
func classifyEntityFieldsTypeError(field Field, err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, semantic.ErrAmbiguousType.Error()):
		return entityFieldsProfileError(field, EntityFieldsAmbiguousTypeDiagnostic, message)
	case strings.Contains(message, semantic.ErrUnknownType.Error()):
		return entityFieldsProfileError(field, EntityFieldsUnknownTypeDiagnostic, message)
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
