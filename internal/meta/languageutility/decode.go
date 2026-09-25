package languageutility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func DecodeContract(raw []byte) (Contract, error) {
	value, err := decodeStrict[Contract](raw)
	if err == nil {
		err = ValidateContract(value)
	}
	return value, err
}

func DecodeObservation(raw []byte) (Observation, error) {
	return decodeStrict[Observation](raw)
}

func decodeStrict[T any](raw []byte) (T, error) {
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return value, fmt.Errorf("trailing JSON value")
	}
	return value, nil
}
