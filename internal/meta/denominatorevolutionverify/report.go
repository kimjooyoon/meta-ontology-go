package denominatorevolutionverify

import (
	"encoding/json"
	"os"
)

func WriteVerification(filename string, value Verification) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(filename, raw, 0o644)
}
