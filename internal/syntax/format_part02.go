package syntax

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// FormatSource parses and formats a complete source file when it is valid.
func FormatSource(filename, source string) (string, Diagnostics, error) {
	file, diagnostics := ParseFile(filename, source)
	if diagnostics.HasErrors() {
		return "", diagnostics, diagnostics.Error()
	}
	formatted, err := Format(file)
	return formatted, diagnostics, err
}

// FormatSourceWithEntityFieldsSupport parses and formats a complete source
// file with an explicit, profile-bound EntityFields mode.
func FormatSourceWithEntityFieldsSupport(filename, source string, support EntityFieldsSupport) (string, Diagnostics, error) {
	file, diagnostics := ParseFileWithEntityFieldsSupport(filename, source, support)
	if diagnostics.HasErrors() {
		return "", diagnostics, diagnostics.Error()
	}
	formatted, err := FormatWithEntityFieldsSupport(file, support)
	return formatted, diagnostics, err
}
func formatDeclaration(output *strings.Builder, declaration Declaration, support EntityFieldsSupport) error {
	switch value := declaration.(type) {
	case *EntityDecl:
		return formatEntity(output, value, support)
	case *ActivityDecl:
		return formatActivity(output, value)
	case *PolicyDecl:
		return formatPolicy(output, value)
	default:
		return fmt.Errorf("unsupported declaration %T", declaration)
	}
}
func formatEntity(output *strings.Builder, entity *EntityDecl, support EntityFieldsSupport) error {
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
	if entity.FieldsPresent || len(entity.Fields) != 0 {
		switch support.State {
		case EntityFieldsDeferred:
			return ErrLatentFieldsUnsupported
		case EntityFieldsSupported:
			for _, field := range entity.Fields {
				if field.TextBound.IsSet() {
					capabilities, err := support.Profile.Capabilities()
					if err != nil || !capabilities.TextBound {
						return fmt.Errorf("max_utf8_bytes requires the EntityFields V6 profile")
					}
				}
			}
			return formatEntityFields(output, entity.Fields)
		default:
			return ErrEntityFieldsUnknownState
		}
	}
	return nil
}
