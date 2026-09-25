package impactcoverage

import (
	"encoding/json"
	"fmt"
	"io"
)

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	if delimiter != '{' {
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	seen := make(map[string]struct{})
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := key.(string)
		if !ok {
			return fmt.Errorf("JSON object key is not a string")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate JSON field %q", name)
		}
		seen[name] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}
