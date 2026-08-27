package languageresourcebudgetconsumer

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadInput(path string) (Input, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Input{}, fmt.Errorf("CONSUMER_INPUT_READ_FAILED")
	}
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return Input{}, fmt.Errorf("CONSUMER_INPUT_INVALID")
	}
	return input, nil
}

func WriteReport(path string, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("CONSUMER_REPORT_ENCODE_FAILED")
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("CONSUMER_REPORT_WRITE_FAILED")
	}
	return nil
}
