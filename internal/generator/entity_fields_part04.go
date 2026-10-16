package generator

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func validateSupportedField(entity Entity, index int, field Field, used, seenNames map[string]string,
	sourceURI string, previousStart int, hasPrevious bool, capabilities syntax.EntityFieldsCapabilities) error {
	context := fmt.Sprintf("entity %q field %d", entity.ID, index)
	if field.ID == "" || field.Name == "" || field.Parent == "" || field.TypeRefID == "" || field.Presence == "" || field.Cardinality == "" {
		return entityFieldsError(entityFieldsIncompleteDiagnostic, field, context+" is missing a required identity, name, type, shape, or parent")
	}
	canonicalID, err := canonicalEntityFieldIdentity(field.ID)
	if err != nil || canonicalID != field.ID {
		return entityFieldsError(entityFieldsIncompleteDiagnostic, field, "field ID must be a canonical absolute identity")
	}
	if previous, exists := used[field.ID]; exists {
		return entityFieldsError(entityFieldsIDCollisionDiagnostic, field, fmt.Sprintf("identity is already used by %s", previous))
	}
	used[field.ID] = "field"
	if field.Parent != entity.ID {
		return entityFieldsError(entityFieldsWrongParentDiagnostic, field, fmt.Sprintf("parent %q does not match entity %q", field.Parent, entity.ID))
	}
	if len(field.Aliases) != 0 {
		return entityFieldsError(entityFieldsUnsupportedShapeDiagnostic, field, "aliases are not part of the bound projection profile")
	}
	if field.Origin != "" && field.Origin != "Source" {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, fmt.Sprintf("field origin %q is not source-authoritative", field.Origin))
	}
	if field.GoName != "" && field.GoName != field.Name {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "Go name must be exactly Field.Name")
	}
	expectedGoType := "string"
	if field.TypeRefID == entityFieldsBooleanTypeID && capabilities.Boolean {
		expectedGoType = "bool"
	} else if field.TypeRefID == entityFieldsIntegerTypeID && capabilities.Integer {
		expectedGoType = "int64"
	}
	if field.Cardinality == "many" && capabilities.Many {
		expectedGoType = "[]" + expectedGoType
	}
	if field.Presence == "optional" && capabilities.Optional {
		expectedGoType = "*" + expectedGoType
	}
	if field.GoType != "" && field.GoType != expectedGoType {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "Go type is derived from the profile and cannot be overridden")
	}
	if !isGoIdentifier(field.Name) {
		return entityFieldsError(entityFieldsGoNameCollisionDiagnostic, field, "Field.Name is not a valid Go identifier")
	}
	if previous, exists := seenNames[field.Name]; exists {
		return entityFieldsError(entityFieldsGoNameCollisionDiagnostic, field, fmt.Sprintf("Go name is already used by field %q", previous))
	}
	seenNames[field.Name] = field.ID
	if field.TypeRefID != entityFieldsStringTypeID && !(capabilities.Boolean && field.TypeRefID == entityFieldsBooleanTypeID) && !(capabilities.Integer && field.TypeRefID == entityFieldsIntegerTypeID) {
		return entityFieldsError(entityFieldsUnsupportedTypeDiagnostic, field, fmt.Sprintf("resolved type %q is not in the bound profile", field.TypeRefID))
	}
	if field.TextBound.Validate() != nil || field.TextBound.IsSet() && (!capabilities.TextBound || field.TypeRefID != entityFieldsStringTypeID) {
		return entityFieldsError(entityFieldsUnsupportedTypeDiagnostic, field, "max_utf8_bytes requires Text under V6")
	}
	validPresence := field.Presence == "required" || capabilities.Optional && field.Presence == "optional"
	if !validPresence || field.Cardinality != "one" && !(capabilities.Many && field.Cardinality == "many") {
		return entityFieldsError(entityFieldsUnsupportedShapeDiagnostic, field, "the bound profile does not support this presence/cardinality shape")
	}
	if err := validateFieldSource(entity, field, sourceURI, previousStart, hasPrevious); err != nil {
		return err
	}
	return nil
}
func canonicalEntityFieldIdentity(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("empty or spaced identity")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || strings.Contains(value, "://") && parsed.Host == "" {
		return "", fmt.Errorf("identity is not absolute")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Host != "" {
		parsed.Host = strings.ToLower(parsed.Host)
	}
	return parsed.String(), nil
}
