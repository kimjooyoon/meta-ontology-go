package languageproofartifactverifier

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteReport(path string, report Report) error {
	if err := Validate(report); err != nil {
		return fmt.Errorf("validate proof-carrying report: %w", err)
	}
	raw, err := json.MarshalIndent(report, "", "  ")
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
	return decodeStrict[Report](raw)
}
