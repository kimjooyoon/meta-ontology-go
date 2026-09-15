package languagedelivery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func DecodeContract(data []byte) (Contract, error) {
	contract, err := decodeStrict[Contract](data)
	if err != nil {
		return Contract{}, err
	}
	if err := ValidateContract(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func DecodeManifest(data []byte) (SourceManifest, error) {
	manifest, err := decodeStrict[SourceManifest](data)
	if err != nil {
		return SourceManifest{}, err
	}
	return manifest, nil
}

func decodeStrict[T any](data []byte) (T, error) {
	var value T
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return value, fmt.Errorf("trailing JSON content")
	}
	return value, nil
}
