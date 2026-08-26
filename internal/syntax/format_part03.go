package syntax

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func formatEntityFields(output *strings.Builder, fields []FieldDecl) error {
	output.WriteString(" fields {")
	for _, field := range fields {
		if err := validateFieldForFormatting(field); err != nil {
			return err
		}
		output.WriteString("\n    field ")
		output.WriteString(field.Name)
		output.WriteString(" id ")
		output.WriteString(quoteString(field.ID))
		output.WriteString(" type ")
		output.WriteString(formatTypeReference(field.TypeRef.Spelling))
		output.WriteByte(' ')
		output.WriteString(string(field.Presence))
		output.WriteByte(' ')
		output.WriteString(string(field.Cardinality))
	}
	if len(fields) != 0 {
		output.WriteByte('\n')
	}
	output.WriteByte('}')
	return nil
}
func validateFieldForFormatting(field FieldDecl) error {
	if field.Span.IsEmpty() || field.NameSpan.IsEmpty() || field.IDSpan.IsEmpty() || field.TypeRef.Span.IsEmpty() || field.PresenceSpan.IsEmpty() || field.CardinalitySpan.IsEmpty() {
		return fmt.Errorf("entity field %q is missing source spans", field.Name)
	}
	if err := validateIdentifier(field.Name, "field name"); err != nil {
		return err
	}
	if isEntityFieldsReservedWord(field.Name) {
		return fmt.Errorf("field name uses reserved EntityFields keyword %q", field.Name)
	}
	if !utf8.ValidString(field.ID) {
		return fmt.Errorf("field id is not valid UTF-8")
	}
	if field.TypeRef.Spelling == "" || !utf8.ValidString(field.TypeRef.Spelling) {
		return fmt.Errorf("field type reference is not valid UTF-8")
	}
	if isEntityFieldsReservedWord(field.TypeRef.Spelling) {
		return fmt.Errorf("field type reference uses reserved EntityFields keyword %q", field.TypeRef.Spelling)
	}
	if !field.Presence.Valid() || !field.Cardinality.Valid() {
		return fmt.Errorf("field %q has an invalid presence or cardinality", field.Name)
	}
	return nil
}
func formatTypeReference(spelling string) string {
	if err := validateIdentifier(spelling, "field type reference"); err == nil {
		return spelling
	}
	return quoteString(spelling)
}
