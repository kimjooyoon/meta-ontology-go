package metrictransition

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func Documents(result Result) ([]byte, []byte, error) {
	state, err := json.MarshalIndent(result.State, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	ledger, err := json.MarshalIndent(result.Ledger, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return append(state, '\n'), append(ledger, '\n'), nil
}

func WriteResult(result Result, statePath, ledgerPath string) error {
	state, ledger, err := Documents(result)
	if err != nil {
		return err
	}
	if err := writeDocument(statePath, state); err != nil {
		return err
	}
	return writeDocument(ledgerPath, ledger)
}

func writeDocument(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
