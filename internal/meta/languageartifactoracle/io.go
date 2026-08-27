package languageartifactoracle

import (
	"encoding/json"
	"os"
)

func WriteReport(path string, report Report) error {
	if err := Validate(report); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
