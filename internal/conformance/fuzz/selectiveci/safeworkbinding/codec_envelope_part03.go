package safeworkbinding

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"unicode"
	"unicode/utf8"
)

func readString(value jsonValue) (string, Reason) {
	switch value.kind {
	case jsonNullValue:
		return "", ReasonNullValue
	case jsonStringValue:
		if value.text == "" {
			return "", ReasonEmptyValue
		}
		return value.text, ReasonNone
	default:
		return "", ReasonInvalidSchema
	}
}
func validateStableID(value string) bool {
	if !utf8.ValidString(value) {
		return false
	}
	if value == "" {
		return false
	}
	if len(value) > 256 {
		return false
	}
	if strings.IndexFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) >= 0 {
		return false
	}
	identity, err := semantic.ParseIdentity(value)
	if err != nil {
		return false
	}
	return identity.String() == value
}
func validateDigest(value string) bool {
	payload, found := strings.CutPrefix(value, "sha256:")
	return found && len(payload) == 64 && strings.Trim(payload, "0123456789abcdef") == ""
}
