package pressureshadow

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}
func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	if delimiter == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	} else if delimiter == '{' {
		seen := map[string]struct{}{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name := key.(string)
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate JSON key %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	_, err = decoder.Token()
	return err
}
