package resourcevector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func canonicalRelativePath(root, value string) (string, bool) {
	root = strings.ReplaceAll(root, "\\", "/")
	value = strings.ReplaceAll(value, "\\", "/")
	if root == "" || value == "" || strings.ContainsAny(root+value, "\x00") {
		return "", false
	}
	root = strings.TrimSuffix(root, "/")
	if after, ok := strings.CutPrefix(value, root+"/"); ok {
		value = after
	} else if strings.HasPrefix(value, "/") {
		return "", false
	}
	if value == "" || strings.HasPrefix(value, "/") {
		return "", false
	}
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}
	return strings.Join(parts, "/"), true
}
func pressureKey(record PressureRecord) string {
	applicable := "false"
	if record.Applicable != nil && *record.Applicable {
		applicable = "true"
	}
	return record.ID + "\x00" + record.IndependenceGroupID + "\x00" + applicable
}
func present(fields map[string]json.RawMessage, name string) bool {
	raw, ok := fields[name]
	return ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}
