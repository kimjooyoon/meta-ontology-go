package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func validateFieldParentStability(before, after Model) error {
	parents := make(map[ID]ID)
	for _, node := range before.Nodes {
		for _, field := range node.Fields {
			parents[field.ID] = field.Parent
		}
	}
	for _, node := range after.Nodes {
		for _, field := range node.Fields {
			if parent, exists := parents[field.ID]; exists && parent != field.Parent {
				return fmt.Errorf("%w: field %s cannot move from %s to %s", ErrInvalidField, field.ID, parent, field.Parent)
			}
		}
	}
	return nil
}
func fieldSemanticEqual(left, right Field) bool {
	leftSemantic, leftErr := left.semantic()
	rightSemantic, rightErr := right.semantic()
	if leftErr != nil || rightErr != nil {
		return false
	}
	registry := semantic.DefaultTypeRegistry()
	leftTypeID, leftErr := comparisonFieldTypeID(left, leftSemantic.TypeRef, registry)
	rightTypeID, rightErr := comparisonFieldTypeID(right, rightSemantic.TypeRef, registry)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return leftSemantic.ID == rightSemantic.ID &&
		leftSemantic.Parent == rightSemantic.Parent &&
		leftTypeID == rightTypeID &&
		leftSemantic.Presence == rightSemantic.Presence &&
		leftSemantic.Cardinality == rightSemantic.Cardinality &&
		leftSemantic.TextBound == rightSemantic.TextBound
}
func writeFieldSemanticFingerprint(builder *strings.Builder, field Field) {
	normalized, err := field.semantic()
	if err != nil {
		writeFingerprintPart(builder, "invalid:"+string(field.ID))
		return
	}
	typeID, err := comparisonFieldTypeID(field, normalized.TypeRef, semantic.DefaultTypeRegistry())
	if err != nil {
		writeFingerprintPart(builder, "invalid-type:"+string(field.ID))
		return
	}
	writeFingerprintPart(builder, string(normalized.ID))
	writeFingerprintPart(builder, string(normalized.Parent))
	writeFingerprintPart(builder, string(typeID))
	writeFingerprintPart(builder, string(normalized.Presence))
	writeFingerprintPart(builder, string(normalized.Cardinality))
	if normalized.TextBound.IsSet() {
		writeFingerprintPart(builder, normalized.TextBound.Canonical())
	}
}
func resolvedTypeID(ref semantic.TypeRef, registry semantic.TypeRegistry) (semantic.ID, error) {
	normalized, err := ref.Normalized()
	if err != nil {
		return "", err
	}
	if normalized.ID != "" {
		return normalized.ID, nil
	}
	definition, err := registry.Resolve(normalized)
	if err != nil {
		return "", err
	}
	return definition.ID, nil
}

// Comparison uses the identity retained by validated lowering. Lookup spelling
// belongs to the source view; it need not exist in the built-in-only registry.
// Admission still resolves the complete source reference with normalizeField.
func comparisonFieldTypeID(field Field, ref semantic.TypeRef, registry semantic.TypeRegistry) (semantic.ID, error) {
	if ref.ID == "" && field.TypeRefUse.ResolvedID != "" {
		ref = semantic.TypeRef{ID: semantic.ID(field.TypeRefUse.ResolvedID)}
	}
	return resolvedFieldTypeID(field, ref, registry)
}
