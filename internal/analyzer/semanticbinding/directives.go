package semanticbinding

import (
	"go/ast"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func parseFields(input string, shape directiveFields) (map[string]string, error) {
	allowed := make(map[string]struct{}, len(shape.allowed))
	for _, name := range shape.allowed {
		allowed[name] = struct{}{}
	}
	fields := make(map[string]string, len(shape.required))
	for index := 0; index < len(input); {
		index = skipSpace(input, index)
		if index == len(input) {
			break
		}
		start := index
		for index < len(input) && input[index] != '=' && input[index] != ' ' && input[index] != '\t' {
			index++
		}
		if start == index || index == len(input) || input[index] != '=' {
			return nil, &Error{
				Code:              CodeMalformedDirective,
				Message:           "fields must use key=\"value\" syntax",
				FullSuiteFallback: true,
			}
		}
		name := input[start:index]
		if _, ok := allowed[name]; !ok {
			return nil, &Error{
				Code:              CodeUnknownField,
				Message:           "unknown directive field " + strconv.Quote(name),
				FullSuiteFallback: true,
			}
		}
		if _, duplicate := fields[name]; duplicate {
			return nil, &Error{
				Code:              CodeDuplicateField,
				Message:           "duplicate directive field " + strconv.Quote(name),
				FullSuiteFallback: true,
			}
		}
		index++
		if index >= len(input) || input[index] != '"' {
			return nil, &Error{
				Code:              CodeMalformedDirective,
				Message:           "directive field values must be double quoted",
				FullSuiteFallback: true,
			}
		}
		index++
		valueStart := index
		for index < len(input) && input[index] != '"' {
			if input[index] == '\\' || input[index] == '\n' || input[index] == '\r' {
				return nil, &Error{
					Code:              CodeMalformedDirective,
					Message:           "directive field values cannot contain escapes or newlines",
					FullSuiteFallback: true,
				}
			}
			index++
		}
		if index == len(input) {
			return nil, &Error{
				Code:              CodeMalformedDirective,
				Message:           "unterminated directive field value",
				FullSuiteFallback: true,
			}
		}
		value := input[valueStart:index]
		if value == "" || value != strings.TrimSpace(value) {
			return nil, &Error{
				Code:              CodeMalformedDirective,
				Message:           "directive field values must not be empty or padded",
				FullSuiteFallback: true,
			}
		}
		fields[name] = value
		index++
		if index < len(input) && input[index] != ' ' && input[index] != '\t' {
			return nil, &Error{
				Code:              CodeMalformedDirective,
				Message:           "directive fields must be separated by whitespace",
				FullSuiteFallback: true,
			}
		}
	}
	for _, name := range shape.required {
		if _, ok := fields[name]; !ok {
			return nil, &Error{
				Code:              CodeMissingField,
				Message:           "missing directive field " + strconv.Quote(name),
				FullSuiteFallback: true,
			}
		}
	}
	return fields, nil
}

func skipSpace(value string, index int) int {
	for index < len(value) && (value[index] == ' ' || value[index] == '\t') {
		index++
	}
	return index
}

func ignoredGeneratedMarker(body string) bool {
	switch strings.TrimSpace(body) {
	case "generated:start", "generated:end", "slot:start", "slot:end":
		return true
	default:
		return false
	}
}

func normalizeIdentity(raw string) (string, error) {
	if raw != strings.TrimSpace(raw) {
		return "", bindingError(CodeInvalidIdentity, Span{}, "identity values may not be padded")
	}
	id, err := semantic.ParseIdentity(raw)
	if err != nil {
		return "", bindingError(CodeInvalidIdentity, Span{}, err.Error())
	}
	return id.String(), nil
}

func validateDirective(current directive) (directive, error) {
	names := make([]string, 0, len(current.fields))
	for name := range current.fields {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		value := current.fields[name]
		switch name {
		case "id", "subject", "pressure":
			canonical, err := normalizeIdentity(value)
			if err != nil {
				return directive{}, withErrorSpan(err, current.span)
			}
			current.fields[name] = canonical
		case "role":
			if !Role(value).valid() {
				return directive{}, bindingError(CodeInvalidRole, current.span, "unknown binding role")
			}
		default:
			return directive{}, bindingError(CodeUnknownField, current.span, "unknown directive field")
		}
	}
	return current, nil
}

func ensureRegistered(current directive, registered map[string]struct{}) error {
	if registered == nil {
		return nil
	}
	for _, name := range []string{"id", "subject", "pressure"} {
		value, present := current.fields[name]
		if !present {
			continue
		}
		if _, ok := registered[value]; !ok {
			return bindingError(CodeUnregisteredID, current.span, "identity is not present in the explicit registry")
		}
	}
	return nil
}

func bindingError(code Code, span Span, message string) *Error {
	return &Error{Code: code, Span: span, Message: message, FullSuiteFallback: true}
}

func withErrorSpan(err error, span Span) error {
	if typed, ok := err.(*Error); ok {
		copy := *typed
		copy.Span = span
		return &copy
	}
	return bindingError(CodeMalformedDirective, span, err.Error())
}
