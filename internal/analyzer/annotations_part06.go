package analyzer

import (
	"strconv"
	"strings"
)

func annotationFields(value string) []string {
	var fields []string
	var current strings.Builder
	inQuote, escaped := false, false
	flush := func() {
		if current.Len() > 0 {
			fields = append(fields, current.String())
			current.Reset()
		}
	}
	for _, character := range value {
		switch {
		case escaped:
			current.WriteRune(character)
			escaped = false
		case character == '\\' && inQuote:
			current.WriteRune(character)
			escaped = true
		case character == '"':
			inQuote = !inQuote
		case (character == ' ' || character == '\t' || character == '\n') && !inQuote:
			flush()
		default:
			current.WriteRune(character)
		}
	}
	flush()
	for index, field := range fields {
		if unquoted, err := strconv.Unquote(field); err == nil {
			fields[index] = unquoted
		}
	}
	return fields
}
