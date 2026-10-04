package assemblyspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

// CanonicalValue preserves exact integer spellings and normalizes object order.
func CanonicalValue(raw string) (string, error) {
	if len(raw) == 0 || len(raw) > 32<<10 || !utf8.ValidString(raw) {
		return "", fmt.Errorf("value requires 1..32 KiB UTF-8 JSON")
	}
	keys := json.NewDecoder(bytes.NewBufferString(raw))
	keys.UseNumber()
	if err := checkValueKeys(keys); err != nil {
		return "", err
	}
	d := json.NewDecoder(bytes.NewBufferString(raw))
	d.UseNumber()
	var value any
	if err := d.Decode(&value); err != nil {
		return "", err
	}
	if d.Decode(new(any)) != io.EOF {
		return "", fmt.Errorf("one JSON value required")
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

// The syntax/IR carrier stays standard-library-only, including in small
// consumer modules that do not load or depend on a model runtime.
func checkValueKeys(d *json.Decoder) error {
	token, err := d.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	seen := make(map[string]bool)
	for d.More() {
		if delimiter == '{' {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return fmt.Errorf("JSON object requires distinct string keys")
			}
			seen[name] = true
		}
		if err := checkValueKeys(d); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}
