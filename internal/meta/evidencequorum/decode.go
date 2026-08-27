package evidencequorum

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
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return value, fmt.Errorf("trailing JSON input")
	}
	return value, nil
}

func DecodeContract(raw []byte) (Contract, error) { return decodeStrict[Contract](raw) }

func DecodeReceipt(raw []byte) (Receipt, error) { return decodeStrict[Receipt](raw) }
