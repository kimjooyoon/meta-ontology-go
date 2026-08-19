package impactgraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// DecodeJSON decodes exactly one strict graph document.
func DecodeJSON(data []byte) (Graph, error) {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return Graph{}, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var raw *Graph
	if err := decoder.Decode(&raw); err != nil {
		return Graph{}, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}
	if raw == nil {
		return Graph{}, fmt.Errorf("%w: expected one graph object", ErrInvalidDocument)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Graph{}, fmt.Errorf("%w: multiple JSON values", ErrInvalidDocument)
		}
		return Graph{}, fmt.Errorf("%w: trailing data: %v", ErrInvalidDocument, err)
	}
	return raw.Normalized()
}
func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	return nil
}
