package experimentportfolio

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
)

func ReadInput(path string) (Input, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Input{}, err
	}
	var input Input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&input)
	return input, err
}

func ReadReport(path string) (Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	var report Report
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&report)
	return report, err
}

func WriteReceipt(path string, receipt Receipt) error {
	return writeJSON(path, receipt)
}

func WriteReport(path string, report Report) error {
	return writeJSON(path, report)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func Equal(left, right Report) bool { return reflect.DeepEqual(left, right) }
