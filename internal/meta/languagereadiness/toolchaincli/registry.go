package toolchaincli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

func DecodeRegistry(raw []byte) (Registry, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	registry := Registry{}
	if err := decoder.Decode(&registry); err != nil {
		return Registry{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Registry{}, fmt.Errorf("toolchain CLI registry has trailing JSON")
	}
	if !reflect.DeepEqual(registry, expectedRegistry()) {
		return Registry{}, fmt.Errorf("toolchain CLI registry contract mismatch")
	}
	return registry, nil
}
