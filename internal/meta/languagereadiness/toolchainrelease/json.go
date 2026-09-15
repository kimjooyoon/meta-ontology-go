package toolchainrelease

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func decodeStrict[T any](raw []byte) (T, error) {
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, fmt.Errorf("unexpected trailing JSON")
	}
	return value, nil
}
