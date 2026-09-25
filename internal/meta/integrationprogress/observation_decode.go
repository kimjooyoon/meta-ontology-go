package integrationprogress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func DecodeObservation(raw []byte) (Observation, error) {
	var observation Observation
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&observation); err != nil {
		return Observation{}, fmt.Errorf("decode integration progress observation: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Observation{}, fmt.Errorf("integration progress observation has trailing content")
	}
	return observation, nil
}
