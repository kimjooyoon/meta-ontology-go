package syntax

import (
	"fmt"
	"unicode/utf8"
)

func validateIdentifier(value, label string) error {
	if value == "" || !utf8.ValidString(value) {
		return fmt.Errorf("%s is not a valid identifier", label)
	}
	first := true
	for _, character := range value {
		if first && !isIdentifierStart(character) {
			return fmt.Errorf("%s is not a valid identifier", label)
		}
		if !first && !isIdentifierContinue(character) {
			return fmt.Errorf("%s is not a valid identifier", label)
		}
		first = false
	}
	if _, keyword := keywordKinds[value]; keyword {
		return fmt.Errorf("%s uses reserved keyword %q", label, value)
	}
	return nil
}
