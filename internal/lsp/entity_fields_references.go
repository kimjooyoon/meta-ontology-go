package lsp

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// appendValueProgramFieldReference records the explicit field use-site carried
// by the profile's value program. A field ID is never inferred from a name.
func appendValueProgramFieldReference(result *ParseResult, source string, activity *syntax.ActivityDecl) error {
	const prefix = "field.read:"
	if !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, prefix) {
		return nil
	}
	payload := strings.TrimPrefix(activity.ValueProgram, prefix)
	name, id, ok := strings.Cut(payload, "=")
	if !ok || name == "" || id == "" || activity.ValueProgramSpan.IsEmpty() {
		return nil
	}
	rangeValue, err := syntaxRange(source, activity.ValueProgramSpan)
	if err != nil {
		return err
	}
	result.References = append(result.References, Reference{Name: name, ID: id, Range: rangeValue})
	return nil
}
