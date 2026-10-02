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
	if len(data) == 0 || len(data) > 1<<20 || !utf8.Valid(data) {
		return nil, fmt.Errorf("receipt must contain 1..1048576 valid UTF-8 bytes")
	}
	scan := json.NewDecoder(bytes.NewReader(data))
	scan.UseNumber()
	if err := checkJSONValue(scan, 0); err != nil {
		return nil, err
	}
	if _, err := scan.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing receipt data")
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
