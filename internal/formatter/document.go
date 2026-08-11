package formatter

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// DeclarationKind identifies the currently supported .gooo declaration forms.
type DeclarationKind string

const (
	EntityDeclaration   DeclarationKind = "entity"
	ActivityDeclaration DeclarationKind = "activity"
)

// Declaration is a syntax-neutral semantic declaration. Entity IDs are
// authoritative. Activity IDs are optional because the initial surface
// grammar derives them from namespace and display name.
type Declaration struct {
	Kind   DeclarationKind
	Name   string
	ID     string
	Inputs []string
	Output string
}

// Document is the minimal semantic view consumed by the formatter.
type Document struct {
	Package      string
	Namespace    string
	Declarations []Declaration
}

// Clone returns a detached document so adapters can safely reuse their AST.
func (d Document) Clone() Document {
	clone := Document{Package: d.Package, Namespace: d.Namespace}
	clone.Declarations = make([]Declaration, len(d.Declarations))
	for index, declaration := range d.Declarations {
		clone.Declarations[index] = declaration
		clone.Declarations[index].Inputs = append([]string(nil), declaration.Inputs...)
	}
	return clone
}

func (d Document) validate() Diagnostics {
	diagnostics := make(Diagnostics, 0)
	if !isIdentifier(d.Package) {
		diagnostics = appendInvalid(diagnostics, "package must be a non-empty identifier")
	}
	if !isIdentifier(d.Namespace) {
		diagnostics = appendInvalid(diagnostics, "namespace must be a non-empty identifier")
	}
	names := make(map[string]struct{}, len(d.Declarations))
	entityNames := make(map[string]struct{}, len(d.Declarations))
	entityIDs := make(map[string]struct{}, len(d.Declarations))
	for _, declaration := range d.Declarations {
		if _, exists := names[declaration.Name]; exists || declaration.Name == "" {
			diagnostics = appendInvalid(diagnostics, "declaration names must be unique and non-empty")
		}
		names[declaration.Name] = struct{}{}
		if declaration.Kind == EntityDeclaration {
			entityNames[declaration.Name] = struct{}{}
			diagnostics = validateEntity(diagnostics, declaration, entityIDs)
		}
	}
	activityIDs := make(map[string]struct{}, len(d.Declarations))
	for _, declaration := range d.Declarations {
		if declaration.Kind == ActivityDeclaration {
			diagnostics = validateActivity(diagnostics, declaration, d.Namespace, entityNames, entityIDs, activityIDs)
			continue
		}
		if declaration.Kind != EntityDeclaration {
			diagnostics = appendInvalid(diagnostics, "unsupported declaration kind "+string(declaration.Kind))
		}
	}
	return diagnostics
}

func validateEntity(diagnostics Diagnostics, declaration Declaration, ids map[string]struct{}) Diagnostics {
	if !isIdentifier(declaration.Name) || strings.TrimSpace(declaration.ID) == "" || hasWhitespace(declaration.ID) {
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
	}
	if declaration.ID != "" && declaration.ID != activityID {
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
	if value == "" {
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
	return true
}

func hasWhitespace(value string) bool {
	for _, r := range value {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func defaultActivityID(namespace, name string) string {
	return namespace + "://activity/" + kebab(name)
}

func kebab(value string) string {
	var result strings.Builder
	for index, r := range value {
		if unicode.IsUpper(r) && index > 0 {
			result.WriteByte('-')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// SemanticFingerprint returns a canonical identity-and-relation view. It
// deliberately excludes display names and declaration order.
func (d Document) SemanticFingerprint() string {
	if d.validate().HasErrors() {
		return ""
	}
	entities := make(map[string]string, len(d.Declarations))
	records := make([]string, 0, len(d.Declarations))
	for _, declaration := range d.Declarations {
		if declaration.Kind == EntityDeclaration {
			entities[declaration.Name] = declaration.ID
			records = append(records, "entity|"+declaration.ID)
		}
	}
	for _, declaration := range d.Declarations {
		if declaration.Kind != ActivityDeclaration {
			continue
		}
		activityID := declaration.ID
		if activityID == "" {
			activityID = defaultActivityID(d.Namespace, declaration.Name)
		}
		inputs := make([]string, len(declaration.Inputs))
		for index, input := range declaration.Inputs {
			inputs[index] = entities[input]
		}
		output := entities[declaration.Output]
		records = append(records, "activity|"+activityID+"|used="+strings.Join(inputs, ",")+"|generated="+output)
	}
	sort.Strings(records)
	return fmt.Sprintf("package=%s|namespace=%s|%s", d.Package, d.Namespace, strings.Join(records, ";"))
}
