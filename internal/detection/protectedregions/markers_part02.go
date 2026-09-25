package protectedregions

import (
	"fmt"
	"strconv"
)

func parseAttributes(input string) (map[string]string, error) {
	attrs := make(map[string]string)
	for index := 0; index < len(input); {
		for index < len(input) && (input[index] == ' ' || input[index] == '\t') {
			index++
		}
		if index == len(input) {
			break
		}
		keyStart := index
		for index < len(input) && isAttributeKeyChar(input[index]) {
			index++
		}
		if keyStart == index || index >= len(input) || input[index] != '=' {
			return nil, fmt.Errorf("invalid marker attribute near byte %d", keyStart)
		}
		key := input[keyStart:index]
		index++
		if index >= len(input) || input[index] != '"' {
			return nil, fmt.Errorf("attribute %q must use a quoted value", key)
		}
		valueStart := index
		index++
		escaped := false
		for index < len(input) {
			if escaped {
				escaped = false
				index++
				continue
			}
			if input[index] == '\\' {
				escaped = true
				index++
				continue
			}
			if input[index] == '"' {
				break
			}
			index++
		}
		if index >= len(input) {
			return nil, fmt.Errorf("unterminated attribute %q", key)
		}
		index++
		value, err := strconv.Unquote(input[valueStart:index])
		if err != nil {
			return nil, fmt.Errorf("invalid attribute %q: %w", key, err)
		}
		if _, exists := attrs[key]; exists {
			return nil, fmt.Errorf("duplicate attribute %q", key)
		}
		attrs[key] = value
	}
	return attrs, nil
}
func isAttributeKeyChar(value byte) bool {
	return value == '_' || value == '-' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}
