package generator

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

func validateSupportedField(entity Entity, index int, field Field, used map[string]string, seenNames map[string]string, sourceURI string, previousStart int, hasPrevious bool) error {
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
	if field.GoType != "" && field.GoType != "string" {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "Go type is derived from the profile and cannot be overridden")
	}
	if !isGoIdentifier(field.Name) {
		return entityFieldsError(entityFieldsGoNameCollisionDiagnostic, field, "Field.Name is not a valid Go identifier")
	}
	if previous, exists := seenNames[field.Name]; exists {
		return entityFieldsError(entityFieldsGoNameCollisionDiagnostic, field, fmt.Sprintf("Go name is already used by field %q", previous))
	}
	seenNames[field.Name] = field.ID
	if field.TypeRefID != entityFieldsStringTypeID {
		return entityFieldsError(entityFieldsUnsupportedTypeDiagnostic, field, fmt.Sprintf("resolved type %q is not in the bound profile", field.TypeRefID))
	}
	if field.Presence != "required" || field.Cardinality != "one" {
		return entityFieldsError(entityFieldsUnsupportedShapeDiagnostic, field, "only required × one is in the bound profile")
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
