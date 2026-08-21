package pressurecoverage

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func validID(value string) bool {
	if value == "" || value != strings.TrimSpace(value) || len(value) > 256 ||
		!utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}
