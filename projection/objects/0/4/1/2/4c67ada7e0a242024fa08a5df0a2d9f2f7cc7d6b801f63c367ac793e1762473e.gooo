package roundtrip

import (
	"fmt"
	"strconv"
)

func parseAttributes(input string) (map[string]string, error) {
	attributes := make(map[string]string)
	for index := 0; index < len(input); {
		index = skipSpaces(input, index)
		if index == len(input) {
			break
		}
		keyStart := index
		for index < len(input) && attributeKeyChar(input[index]) {
			index++
		}
		if keyStart == index || index == len(input) || input[index] != '=' {
			return nil, fmt.Errorf("expected quoted attribute near byte %d", keyStart)
		}
		key := input[keyStart:index]
		index++
		if index == len(input) || input[index] != '"' {
			return nil, fmt.Errorf("attribute %q must use a quoted value", key)
		}
		valueStart := index
		index++
		escaped, closed := false, false
		for index < len(input) {
			if escaped {
				escaped = false
				index++
				continue
			}
			switch input[index] {
			case '\\':
				escaped = true
			case '"':
				index++
				closed = true
				goto valueDone
			}
			index++
		}
	valueDone:
		if !closed {
			return nil, fmt.Errorf("unterminated attribute %q", key)
		}
		value, err := strconv.Unquote(input[valueStart:index])
		if err != nil {
			return nil, fmt.Errorf("invalid attribute %q: %w", key, err)
		}
		if _, exists := attributes[key]; exists {
			return nil, fmt.Errorf("duplicate attribute %q", key)
		}
		attributes[key] = value
	}
	return attributes, nil
}
