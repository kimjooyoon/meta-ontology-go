package selfimprovementattestation

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func WriteReceipt(filename string, receipt ResolutionReceipt) error {
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}
