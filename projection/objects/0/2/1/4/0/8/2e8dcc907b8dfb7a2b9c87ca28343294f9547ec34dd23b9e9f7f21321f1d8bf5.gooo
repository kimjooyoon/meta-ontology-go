package bidir

import (
	"strings"
	"unicode"
)

func slug(value string) string {
	var builder strings.Builder
	lastDash := false
	var previous rune
	for _, current := range value {
		if unicode.IsUpper(current) && unicode.IsLower(previous) && builder.Len() > 0 && !lastDash {
			builder.WriteByte('-')
		}
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			builder.WriteRune(unicode.ToLower(current))
			lastDash = false
		} else if builder.Len() > 0 && !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
		previous = current
	}
	return strings.Trim(builder.String(), "-")
}
