package semanticbinding

import (
	"go/ast"
	"strings"
)

type directive struct {
	kind    string
	fields  map[string]string
	span    Span
	comment *ast.Comment
}
type directiveFields struct {
	allowed  []string
	required []string
}

func parseDirective(comment *ast.Comment, span Span) (directive, bool, error) {
	text := comment.Text
	if strings.HasPrefix(text, "/*") {
		if strings.Contains(text, "gooo:") {
			return directive{}, false, bindingError(CodeUnknownDirective, span, "block comments are not binding directives")
		}
		return directive{}, false, nil
	}
	if !strings.HasPrefix(text, "//") {
		return directive{}, false, nil
	}
	body := strings.TrimPrefix(text, "//")
	if !strings.HasPrefix(body, "gooo:") {
		return directive{}, false, nil
	}
	body = strings.TrimPrefix(body, "gooo:")
	if ignoredGeneratedMarker(body) {
		return directive{}, false, nil
	}
	name, fieldsText := firstWord(body)
	var shape directiveFields
	switch name {
	case "bind":
		shape = directiveFields{
			allowed:  []string{"id", "role"},
			required: []string{"id", "role"},
		}
	case "obligation":
		shape = directiveFields{
			allowed:  []string{"id", "subject", "pressure"},
			required: []string{"id", "subject", "pressure"},
		}
	default:
		return directive{}, false, bindingError(
			CodeUnknownDirective, span, "only bind and obligation directives are recognized",
		)
	}
	fields, err := parseFields(fieldsText, shape)
	if err != nil {
		return directive{}, false, withErrorSpan(err, span)
	}
	return directive{kind: name, fields: fields, span: span, comment: comment}, true, nil
}
func firstWord(value string) (string, string) {
	for index, character := range value {
		if character == ' ' || character == '\t' {
			return value[:index], value[index:]
		}
	}
	return value, ""
}
