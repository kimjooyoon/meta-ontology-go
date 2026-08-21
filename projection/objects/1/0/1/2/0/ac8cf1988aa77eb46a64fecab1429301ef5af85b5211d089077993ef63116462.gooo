package workfrontier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func joinR4(values []string) string {
	var result strings.Builder
	for index, value := range values {
		if index != 0 {
			result.WriteString("\x00")
		}
		result.WriteString(value)
	}
	return result.String()
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
