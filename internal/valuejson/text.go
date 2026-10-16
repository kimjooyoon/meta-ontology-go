// Package valuejson decodes value transport without normalizing malformed Text.
package valuejson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

func DecodeText(raw []byte, maximum uint32) (string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '"' || !utf8.Valid(raw) || !json.Valid(raw) {
		return "", fmt.Errorf("Text requires a string of at most %d UTF-8 bytes", maximum)
	}
	if err := CheckEscapesInValidJSON(raw); err != nil {
		return "", err
	}
	var text string
	if json.Unmarshal(raw, &text) != nil || uint64(len(text)) > uint64(maximum) {
		return "", fmt.Errorf("Text requires a string of at most %d UTF-8 bytes", maximum)
	}
	return text, nil
}

// CheckEscapesInValidJSON requires complete JSON syntax and valid raw UTF-8.
// Callers check those first; byte offsets always refer to the original input.
// The unchanged Gooo projection classifies Unicode units and surrogate pairs.
func CheckEscapesInValidJSON(raw []byte) error {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		if raw[i+1] != 'u' {
			i++
			continue
		}
		next := int64(-1)
		if i+12 <= len(raw) && raw[i+6] == '\\' && raw[i+7] == 'u' {
			next = escapeUnit(raw[i+8 : i+12])
		}
		switch status := TextEscapePolicy(escapeUnit(raw[i+2:i+6]), next); status {
		case "SURROGATE_PAIR":
			i += 6
		case "SCALAR_ESCAPE":
		default:
			return fmt.Errorf("Text JSON escape at byte %d: %s", i, status)
		}
		i += 5
	}
	return nil
}

func escapeUnit(digits []byte) int64 {
	var unit int64
	for _, digit := range digits {
		unit *= 16
		switch {
		case digit <= '9':
			unit += int64(digit - '0')
		case digit <= 'F':
			unit += int64(digit-'A') + 10
		default:
			unit += int64(digit-'a') + 10
		}
	}
	return unit
}
