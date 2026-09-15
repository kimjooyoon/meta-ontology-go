package semanticbinding

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strconv"
	"strings"
)

func parseFieldValue(input string, index int) (string, int, error) {
	if index >= len(input) || input[index] != '"' {
		return "", 0, directiveError(CodeMalformedDirective, "directive field values must be double quoted")
	}
	index++
	valueStart := index
	for index < len(input) && input[index] != '"' {
		if input[index] == '\\' || input[index] == '\n' || input[index] == '\r' {
			return "", 0, directiveError(CodeMalformedDirective, "directive field values cannot contain escapes or newlines")
		}
		index++
	}
	if index == len(input) {
		return "", 0, directiveError(CodeMalformedDirective, "unterminated directive field value")
	}
	value := input[valueStart:index]
	if value == "" || value != strings.TrimSpace(value) {
		return "", 0, directiveError(CodeMalformedDirective, "directive field values must not be empty or padded")
	}
	return value, index + 1, nil
}
func consumeFieldSeparator(input string, index int) (int, error) {
	if index < len(input) && input[index] != ' ' && input[index] != '\t' {
		return 0, directiveError(CodeMalformedDirective, "directive fields must be separated by whitespace")
	}
	return index, nil
}
func requireFields(fields map[string]string, required []string) error {
	for _, name := range required {
		if _, ok := fields[name]; !ok {
			return directiveError(CodeMissingField, "missing directive field "+strconv.Quote(name))
		}
	}
	return nil
}
func directiveError(code Code, message string) error {
	return &Error{Code: code, Message: message, FullSuiteFallback: true}
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
