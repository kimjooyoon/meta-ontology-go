package sourceauthority

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
)

//go:embed contract.json
var contractBytes []byte

func Decode(raw []byte) (Contract, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var contract Contract
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, fmt.Errorf("decode source authority contract: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Contract{}, fmt.Errorf("decode source authority contract: trailing value")
	}
	return contract, nil
}

func Load() (Contract, error) {
	contract, err := Decode(contractBytes)
	if err != nil {
		return Contract{}, err
	}
	if err := Validate(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func Snapshot() []byte {
	return append([]byte(nil), contractBytes...)
}
