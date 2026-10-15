package bodycodegen

import (
	"fmt"
	"strconv"
	"strings"
)

// Positions in computes are source bytes; Go AST positions use decoded bytes.
// Verify the complete decoded value before returning any mapped position.
func sourceLiteralPosition(raw, value string, wanted int, sourceToValue bool) (int, error) {
	if len(raw) < 2 || raw[0] != raw[len(raw)-1] || (raw[0] != '"' && raw[0] != '`') {
		return 0, fmt.Errorf("complete Gooo body literal required")
	}
	if raw[0] == '`' {
		if raw[1:len(raw)-1] != value {
			return 0, fmt.Errorf("raw body literal differs from decoded source")
		}
		if sourceToValue && wanted >= 1 && wanted <= len(raw)-1 {
			return wanted - 1, nil
		}
		if !sourceToValue && wanted >= 0 && wanted <= len(value) {
			return wanted + 1, nil
		}
		return 0, fmt.Errorf("body position outside literal")
	}
	var decoded strings.Builder
	position, mapped := 1, -1
	for position < len(raw)-1 {
		remaining := raw[position : len(raw)-1]
		character, _, tail, err := strconv.UnquoteChar(remaining, '"')
		if err != nil {
			return 0, fmt.Errorf("body literal position: %w", err)
		}
		next := position + len(remaining) - len(tail)
		if sourceToValue && wanted >= position && wanted < next {
			mapped = decoded.Len()
		}
		if !sourceToValue && wanted == decoded.Len() {
			mapped = position
		}
		decoded.WriteRune(character)
		position = next
	}
	if decoded.String() != value {
		return 0, fmt.Errorf("body literal differs from decoded source")
	}
	if sourceToValue && wanted == len(raw)-1 {
		mapped = len(value)
	}
	if !sourceToValue && wanted == len(value) {
		mapped = len(raw) - 1
	}
	if mapped < 0 {
		return 0, fmt.Errorf("body position is not a supported literal boundary")
	}
	return mapped, nil
}
