package assemblyspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
)

// CanonicalValue preserves exact integer spellings and normalizes object order.
func CanonicalValue(raw string) (string, error) {
	if len(raw) == 0 || len(raw) > 32<<10 || !utf8.ValidString(raw) {
		return "", fmt.Errorf("value requires 1..32 KiB UTF-8 JSON")
	}
	if err := decision.RejectDuplicateJSONKeys([]byte(raw)); err != nil {
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
