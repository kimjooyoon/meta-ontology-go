package syntax

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Format renders a syntax tree in the canonical .gooo source form.
// Formatting is semantic: source spans and original whitespace are not copied.
func Format(file *File) (string, error) {
	if file == nil || file.Package == nil || file.Namespace == nil {
		return "", fmt.Errorf("package and namespace declarations are required")
	}
	if err := validateIdentifier(file.Package.Name, "package name"); err != nil {
		return "", err
	}
	if err := validateIdentifier(file.Namespace.Name, "namespace name"); err != nil {
		return "", err
	}
	if file.Decls != nil && file.Declarations != nil && !sameDeclarations(file.Decls, file.Declarations) {
		return "", fmt.Errorf("file declaration aliases conflict")
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	var output strings.Builder
	fmt.Fprintf(&output, "package %s\nnamespace %s\n", file.Package.Name, file.Namespace.Name)
	if len(declarations) > 0 {
		output.WriteByte('\n')
		for index, declaration := range declarations {
			if index > 0 {
				output.WriteByte('\n')
			}
			if err := formatDeclaration(&output, declaration); err != nil {
				return "", err
			}
		}
	}
	output.WriteByte('\n')
	return output.String(), nil
}

func sameDeclarations(left, right []Declaration) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// FormatSource parses and formats a complete source file when it is valid.
func FormatSource(filename, source string) (string, Diagnostics, error) {
	file, diagnostics := ParseFile(filename, source)
	if diagnostics.HasErrors() {
		return "", diagnostics, diagnostics.Error()
	}
	formatted, err := Format(file)
	return formatted, diagnostics, err
}

func formatDeclaration(output *strings.Builder, declaration Declaration) error {
	switch value := declaration.(type) {
	case *EntityDecl:
		return formatEntity(output, value)
	case *ActivityDecl:
		return formatActivity(output, value)
	default:
		return fmt.Errorf("unsupported declaration %T", declaration)
	}
}

func formatEntity(output *strings.Builder, entity *EntityDecl) error {
	if entity == nil {
		return fmt.Errorf("nil entity declaration")
	}
	if err := validateIdentifier(entity.Name, "entity name"); err != nil {
		return err
	}
	if !utf8.ValidString(entity.ID) {
		return fmt.Errorf("entity id is not valid UTF-8")
	}
	fmt.Fprintf(output, "entity %s id %s", entity.Name, quoteString(entity.ID))
	return nil
}

func formatActivity(output *strings.Builder, activity *ActivityDecl) error {
	if activity == nil {
		return fmt.Errorf("nil activity declaration")
	}
	if err := validateIdentifier(activity.Name, "activity name"); err != nil {
		return err
	}
	parameters, err := activityParameters(activity)
	if err != nil {
		return err
	}
	result, err := activityResult(activity)
	if err != nil {
		return err
	}
	output.WriteString("activity ")
	output.WriteString(activity.Name)
	output.WriteByte('(')
	for index, parameter := range parameters {
		if index > 0 {
			output.WriteString(", ")
		}
		output.WriteString(parameter.Name)
	}
	output.WriteString(") -> ")
	output.WriteString(result)
	return nil
}

func activityParameters(activity *ActivityDecl) ([]NameRef, error) {
	if activity.Inputs != nil && activity.Parameters != nil && !sameNames(activity.Inputs, activity.Parameters) {
		return nil, fmt.Errorf("activity %s has conflicting input aliases", activity.Name)
	}
	parameters := activity.Parameters
	if parameters == nil {
		parameters = activity.Inputs
	}
	for _, parameter := range parameters {
		if err := validateIdentifier(parameter.Name, "activity parameter"); err != nil {
			return nil, err
		}
	}
	return parameters, nil
}

func activityResult(activity *ActivityDecl) (string, error) {
	result := activity.Result.Name
	if result != "" && activity.Output != "" && result != activity.Output {
		return "", fmt.Errorf("activity %s has conflicting result aliases", activity.Name)
	}
	if result == "" {
		result = activity.Output
	}
	if err := validateIdentifier(result, "activity result"); err != nil {
		return "", err
	}
	return result, nil
}

func sameNames(left, right []NameRef) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Name != right[index].Name {
			return false
		}
	}
	return true
}

func validateIdentifier(value, label string) error {
	if value == "" || !utf8.ValidString(value) {
		return fmt.Errorf("%s is not a valid identifier", label)
	}
	first := true
	for _, character := range value {
		if first && !isIdentifierStart(character) {
			return fmt.Errorf("%s is not a valid identifier", label)
		}
		if !first && !isIdentifierContinue(character) {
			return fmt.Errorf("%s is not a valid identifier", label)
		}
		first = false
	}
	if _, keyword := keywordKinds[value]; keyword {
		return fmt.Errorf("%s uses reserved keyword %q", label, value)
	}
	return nil
}

func quoteString(value string) string {
	var output strings.Builder
	output.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"':
			output.WriteString(`\"`)
		case '\\':
			output.WriteString(`\\`)
		case '\n':
			output.WriteString(`\n`)
		case '\r':
			output.WriteString(`\r`)
		case '\t':
			output.WriteString(`\t`)
		default:
			if character < 0x20 || character == 0x7f {
				writeUnicodeEscape(&output, character)
			} else {
				output.WriteRune(character)
			}
		}
	}
	output.WriteByte('"')
	return output.String()
}

func writeUnicodeEscape(output *strings.Builder, character rune) {
	const hexadecimal = "0123456789abcdef"
	output.WriteString(`\u`)
	output.WriteByte(hexadecimal[(character>>12)&0xf])
	output.WriteByte(hexadecimal[(character>>8)&0xf])
	output.WriteByte(hexadecimal[(character>>4)&0xf])
	output.WriteByte(hexadecimal[character&0xf])
}
