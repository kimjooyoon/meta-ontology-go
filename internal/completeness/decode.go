package completeness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"
)

// Decode retains exact JSON numbers, rejects duplicate keys and trailing data,
// and checks the shared accounting contract without executing observations.
func Decode(data []byte) (*CompletenessReceipt, error) {
	if err := CheckJSON(data, 1<<20); err != nil {
		return nil, err
	}
	if err := checkShape(data, reflect.TypeFor[CompletenessReceipt]()); err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	d.UseNumber()
	var r CompletenessReceipt
	if err := d.Decode(&r); err != nil {
		return nil, err
	}
	if err := Validate(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CheckJSON checks bounded UTF-8 JSON, duplicate keys, depth and trailing data.
// It does not validate a schema or authenticate an observation.
func CheckJSON(data []byte, limit int) error {
	if len(data) == 0 || len(data) > limit || !utf8.Valid(data) {
		return fmt.Errorf("JSON must contain 1..%d valid UTF-8 bytes", limit)
	}
	scan := json.NewDecoder(bytes.NewReader(data))
	scan.UseNumber()
	if err := checkJSONValue(scan, 0); err != nil {
		return err
	}
	if _, err := scan.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}

func checkJSONValue(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("receipt JSON exceeds 64 nested levels")
	}
	token, err := d.Token()
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
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || keys[name] {
				return fmt.Errorf("duplicate or invalid JSON object key")
			}
			keys[name] = true
			if err := checkJSONValue(d, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := checkJSONValue(d, depth+1); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter")
	}
	_, err = d.Token()
	return err
}
