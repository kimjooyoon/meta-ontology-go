package languageproofartifactverifier

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
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return value, fmt.Errorf("trailing JSON data")
	}
	return value, nil
}
