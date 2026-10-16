package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// Composition text is checked against the original JSON before accepting its
// decoded value. Syntax, byte limits and duplicate-field checks stay in decode.
func decodeCompositionTextData(raw []byte, output any, limit int) error {
	if len(raw) == 0 || len(raw) > limit || !utf8.Valid(raw) || !json.Valid(raw) {
		return decode(raw, output, limit)
	}
	if err := validateJSONTextEscapes(raw); err != nil {
		return err
	}
	return decode(raw, output, limit)
}

func decodeTextScalar(raw []byte) (string, error) {
	var text string
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '"' || !utf8.Valid(raw) || !json.Valid(raw) {
		return "", fmt.Errorf("Text requires a string of at most 1024 UTF-8 bytes")
	}
	if err := validateJSONTextEscapes(raw); err != nil {
		return "", err
	}
	if json.Unmarshal(raw, &text) != nil || len(text) > compositionTextMaxBytes {
		return "", fmt.Errorf("Text requires a string of at most 1024 UTF-8 bytes")
	}
	return text, nil
}

// The caller has checked JSON syntax and raw UTF-8. A backslash can therefore
// only begin a complete escape inside a string. Offsets refer to these bytes.
func validateJSONTextEscapes(raw []byte) error {
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
			next = jsonEscapeUnit(raw[i+8 : i+12])
		}
		switch status := TextEscapePolicy(jsonEscapeUnit(raw[i+2:i+6]), next); status {
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

// Four hexadecimal digits have already passed the JSON grammar check.
func jsonEscapeUnit(digits []byte) int64 {
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
