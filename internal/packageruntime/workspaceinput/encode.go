package workspaceinput

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Encode preserves authored values in the external manifest's existing format.
// The result includes its final newline and must reload within the same bound.
func Encode(definition Definition) ([]byte, error) {
	raw, err := json.MarshalIndent(definition, "", "  ")
	if err != nil {
		return nil, err
	}
	raw = append(raw, '\n')
	reloaded, err := Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("encode workspace manifest: %w", err)
	}
	if !reflect.DeepEqual(reloaded, definition) {
		return nil, fmt.Errorf("encoding changed authored workspace values")
	}
	return raw, nil
}
