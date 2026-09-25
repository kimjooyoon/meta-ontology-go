package languageprofileexperiment

import (
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
	err = json.Unmarshal(data, &input)
	return input, err
}

func ReadReport(path string) (Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	var report Report
	err = json.Unmarshal(data, &report)
	return report, err
}

func WriteReport(path string, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func Equal(left, right Report) bool { return reflect.DeepEqual(left, right) }
