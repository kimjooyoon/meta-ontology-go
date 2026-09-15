package generator

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
		if keyStart == index {
			return nil, fmt.Errorf("expected attribute name at byte %d", keyStart)
		}
		key := input[keyStart:index]
		if index >= len(input) || input[index] != '=' {
			return nil, fmt.Errorf("expected '=' after attribute %q", key)
		}
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
				index++
				break
			}
			index++
		}
		if index > len(input) || input[index-1] != '"' {
			return nil, fmt.Errorf("unterminated attribute %q", key)
		}
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
