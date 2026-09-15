package metarecognition

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func uniqueValues(caseID, kind string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("replay case %q has empty %s", caseID, kind)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("replay case %q has duplicate %s %q", caseID, kind, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}
func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	return requireEOF(decoder)
}
func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		return expectDelimiter(decoder, ']')
	case '{':
		keys := make(map[string]struct{})
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, exists := keys[name]; exists {
				return fmt.Errorf("duplicate JSON field %q", name)
			}
			keys[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		return expectDelimiter(decoder, '}')
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}
