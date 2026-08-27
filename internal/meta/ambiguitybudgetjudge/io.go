package ambiguitybudgetjudge

import (
	"encoding/json"
	"os"
)

func WriteResult(path string, result Result) error {
	raw, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
