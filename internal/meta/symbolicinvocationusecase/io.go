package symbolicinvocationusecase

import "encoding/json"

func DecodeInput(data []byte) (Input, error) {
	var input Input
	err := json.Unmarshal(data, &input)
	return input, err
}

func Marshal(report Report) ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
