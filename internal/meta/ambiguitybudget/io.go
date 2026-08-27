package ambiguitybudget

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func DecodeContract(raw []byte) (Contract, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var contract Contract
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Contract{}, fmt.Errorf("contract has trailing JSON")
		}
		return Contract{}, err
	}
	return contract, nil
}

func WriteReceipt(path string, receipt Receipt) error {
	if err := Validate(receipt); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
