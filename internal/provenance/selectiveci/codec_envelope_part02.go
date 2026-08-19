package selectiveci

import (
	"encoding/json"
)

func DecodeInput(data []byte) (Input, error) {
	var value Input
	if err := decodeStrict(data, &wireInput{}); err != nil {
		return Input{}, err
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return Input{}, err
	}
	return value, nil
}
func EncodeReceipt(value Receipt) ([]byte, error) { return json.Marshal(value) }
func DecodeReceipt(data []byte) (Receipt, error) {
	var value Receipt
	if err := json.Unmarshal(data, &value); err != nil {
		return Receipt{}, err
	}
	return value, nil
}
