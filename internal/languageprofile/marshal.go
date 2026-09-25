package languageprofile

import "encoding/json"

func Marshal(receipt Receipt) ([]byte, error) {
	if err := Validate(receipt); err != nil {
		return nil, err
	}
	payload, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
