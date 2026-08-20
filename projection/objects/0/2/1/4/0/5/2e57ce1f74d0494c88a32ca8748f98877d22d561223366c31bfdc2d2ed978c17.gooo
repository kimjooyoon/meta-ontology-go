package main

import (
	"encoding/json"
)

func encodeSelectiveCIShadowOutput(output selectiveCIShadowOutput) ([]byte, error) {
	output = sealSelectiveCIShadowOutput(output)
	data, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
