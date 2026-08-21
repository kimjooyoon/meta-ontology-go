package formatter

import (
	"unicode"
	"unicode/utf8"
)

func validateEntity(diagnostics Diagnostics, declaration Declaration, ids map[string]struct{}) Diagnostics {
	if !isIdentifier(declaration.Name) || declaration.ID == "" || !utf8.ValidString(declaration.ID) {
		return appendInvalid(diagnostics, "entity requires an identifier and a stable semantic ID")
	}
	if _, exists := ids[declaration.ID]; exists {
		return appendInvalid(diagnostics, "semantic IDs must be unique")
	}
	ids[declaration.ID] = struct{}{}
	return diagnostics
}
func validateActivity(diagnostics Diagnostics, declaration Declaration, namespace string, entityNames, entityIDs, activityIDs map[string]struct{}) Diagnostics {
	if !isIdentifier(declaration.Name) || declaration.Output == "" {
		return appendInvalid(diagnostics, "activity requires an identifier and one output")
	}
	activityID := declaration.ID
	if activityID == "" {
		activityID = defaultActivityID(namespace, declaration.Name)
	} else if declaration.ID != defaultActivityID(namespace, declaration.Name) {
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityError, Code: CodeUnsupportedIdentity, Message: "activity identity cannot be represented by the initial surface grammar"})
	}
	if _, exists := entityIDs[activityID]; exists {
		diagnostics = appendInvalid(diagnostics, "semantic IDs must be unique")
	}
	if _, exists := activityIDs[activityID]; exists {
		diagnostics = appendInvalid(diagnostics, "semantic IDs must be unique")
	}
	activityIDs[activityID] = struct{}{}
	for _, input := range append(append([]string(nil), declaration.Inputs...), declaration.Output) {
		if !isIdentifier(input) {
			diagnostics = appendInvalid(diagnostics, "activity references must be identifiers")
		}
		if _, exists := entityNames[input]; !exists {
			diagnostics = appendInvalid(diagnostics, "activity reference must name a declared entity")
		}
	}
	return diagnostics
}
func appendInvalid(diagnostics Diagnostics, message string) Diagnostics {
	return append(diagnostics, Diagnostic{Severity: SeverityError, Code: CodeInvalidDocument, Message: message})
}
func isIdentifier(value string) bool {
	if value == "" || !utf8.ValidString(value) {
		return false
	}
	for index, r := range value {
		if index == 0 && r != '_' && !unicode.IsLetter(r) {
			return false
		}
		if index > 0 && r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return !isReservedKeyword(value)
}
func isReservedKeyword(value string) bool {
	switch value {
	case "package", "namespace", "entity", "id", "activity":
		return true
	default:
		return false
	}
}
func defaultActivityID(namespace, name string) string {
	return namespace + "://activity/" + kebab(name)
}
