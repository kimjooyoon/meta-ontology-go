package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func uniqueJSON(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := jsonValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("JSON contains trailing values")
	}
	return nil
}

func jsonValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, container := token.(json.Delim)
	if !container {
		return nil
	}
	switch delimiter {
	case '{':
		keys := map[string]bool{}
		for decoder.More() {
			token, err := decoder.Token()
			key, ok := token.(string)
			if err != nil || !ok || keys[key] {
				return fmt.Errorf("JSON object contains an invalid or duplicate key")
			}
			keys[key] = true
			if err := jsonValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := jsonValue(decoder); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter")
	}
	closing, err := decoder.Token()
	if err != nil || (delimiter == '{' && closing != json.Delim('}')) ||
		(delimiter == '[' && closing != json.Delim(']')) {
		return fmt.Errorf("JSON container is incomplete")
	}
	return nil
}

func decodeStrict(raw []byte, target any) error {
	if err := uniqueJSON(raw); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("JSON contains trailing values")
	}
	return nil
}
