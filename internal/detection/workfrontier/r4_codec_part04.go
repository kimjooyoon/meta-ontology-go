package workfrontier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func joinR4(values []string) string {
	result := ""
	for index, value := range values {
		if index != 0 {
			result += "\x00"
		}
		result += value
	}
	return result
}
func r4FieldPresent(fields map[string]json.RawMessage, name string) bool {
	raw, ok := fields[name]
	return ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
func rejectR4DuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanR4JSONValue(decoder); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}
