package evidencequorum

import (
	"encoding/json"
	"os"
)

func WriteReceipt(path string, value Receipt) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func WriteReport(path string, value Report) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func LoadReport(path string) (Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	var value Report
	if err := json.Unmarshal(raw, &value); err != nil {
		return Report{}, err
	}
	return value, nil
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestJSON(report)
}
