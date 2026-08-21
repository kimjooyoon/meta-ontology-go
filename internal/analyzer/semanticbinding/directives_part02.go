package semanticbinding

import (
	"strconv"
)

func parseFields(input string, shape directiveFields) (map[string]string, error) {
	allowed := allowedFieldSet(shape.allowed)
	fields := make(map[string]string, len(shape.required))
	for index := 0; index < len(input); {
		index = skipSpace(input, index)
		if index == len(input) {
			break
		}
		name, next, err := parseFieldName(input, index, allowed, fields)
		if err != nil {
			return nil, err
		}
		value, next, err := parseFieldValue(input, next)
		if err != nil {
			return nil, err
		}
		fields[name] = value
		index, err = consumeFieldSeparator(input, next)
		if err != nil {
			return nil, err
		}
	}
	return fields, requireFields(fields, shape.required)
}
func allowedFieldSet(names []string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(names))
	for _, name := range names {
		allowed[name] = struct{}{}
	}
	return allowed
}
func parseFieldName(input string, start int, allowed map[string]struct{}, fields map[string]string) (string, int, error) {
	index := start
	for index < len(input) && input[index] != '=' && input[index] != ' ' && input[index] != '\t' {
		index++
	}
	if start == index || index == len(input) || input[index] != '=' {
		return "", 0, directiveError(CodeMalformedDirective, "fields must use key=\"value\" syntax")
	}
	name := input[start:index]
	if _, ok := allowed[name]; !ok {
		return "", 0, directiveError(CodeUnknownField, "unknown directive field "+strconv.Quote(name))
	}
	if _, duplicate := fields[name]; duplicate {
		return "", 0, directiveError(CodeDuplicateField, "duplicate directive field "+strconv.Quote(name))
	}
	return name, index + 1, nil
}
