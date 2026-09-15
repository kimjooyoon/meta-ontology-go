package pathclosure

import (
	"encoding/json"
	"fmt"
	"io"
)

func walkR4JSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); ok {
		switch delimiter {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				key, tokenErr := decoder.Token()
				if tokenErr != nil {
					return tokenErr
				}
				name, ok := key.(string)
				if !ok {
					return fmt.Errorf("JSON object key is not a string")
				}
				if _, exists := seen[name]; exists {
					return fmt.Errorf("duplicate JSON field %q", name)
				}
				seen[name] = struct{}{}
				if err := walkR4JSON(decoder); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := walkR4JSON(decoder); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
		}
	}
	return nil
}
func requireR4EOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("strict JSON: trailing value")
		}
		return fmt.Errorf("strict JSON: trailing data: %w", err)
	}
	return nil
}
