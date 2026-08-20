package syntax

import (
	"errors"
	"fmt"
	"strings"
)

// ErrLatentFieldsUnsupported is returned when callers attempt to serialize a
// synthetic field carrier before the public grammar can represent fields.
var ErrLatentFieldsUnsupported = errors.New("latent entity fields are unsupported by the public syntax formatter")

// Format renders a syntax tree in the canonical .gooo source form.
// Formatting is semantic: source spans and original whitespace are not copied.
func Format(file *File) (string, error) {
	return FormatWithEntityFieldsSupport(file, CurrentEntityFieldsSupport())
}

// FormatWithEntityFieldsSupport renders a syntax tree with an explicit,
// profile-bound EntityFields mode.
func FormatWithEntityFieldsSupport(file *File, support EntityFieldsSupport) (string, error) {
	if err := support.Validate(); err != nil {
		return "", err
	}
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
			if err := formatDeclaration(&output, declaration, support); err != nil {
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
