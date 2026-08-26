package bindingcoverage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

func sortedStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
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
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter == '{' {
		return scanJSONObject(decoder)
	}
	if delimiter == '[' {
		return scanJSONArray(decoder)
	}
	return fmt.Errorf("unexpected JSON delimiter")
}
func scanJSONObject(decoder *json.Decoder) error {
	seen := map[string]struct{}{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := token.(string)
		if !ok {
			return fmt.Errorf("object key is not a string")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate object field %q", name)
		}
		seen[name] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}
func scanJSONArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}
