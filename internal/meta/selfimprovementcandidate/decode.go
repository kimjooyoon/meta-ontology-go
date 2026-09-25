package selfimprovementcandidate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func decodeSource(raw []byte) (sourceObservation, error) {
	result := sourceObservation{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("decode observation: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return result, fmt.Errorf("decode observation: trailing content")
	}
	return result, nil
}

func sourceDigest(source sourceObservation) string {
	source.Digest = ""
	return digestJSON(source)
}
