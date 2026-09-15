package provenance

import (
	"fmt"
	"os"
	"path/filepath"
)

func materializeLedger(path string, data []byte, points ledgerMaterializeFaultPoints) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create provenance directory: %w", err)
	}
	if points.create != "" {
		if err := failStorageOperation(points.create); err != nil {
			return fmt.Errorf("create recovery ledger: %w", err)
		}
	}
	temporary, err := os.CreateTemp(directory, ".provenance-ledger-*")
	if err != nil {
		return fmt.Errorf("create recovery ledger: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := writeFullAt(temporary, data, points.write); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write recovery ledger: %w", err)
	}
	if err := syncFile(temporary, points.sync); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync recovery ledger: %w", err)
	}
	if err := closeFile(temporary, points.close); err != nil {
		return fmt.Errorf("close recovery ledger: %w", err)
	}
	if err := failStorageOperation(points.rename); err != nil {
		return fmt.Errorf("rename recovery ledger: %w", err)
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("rename recovery ledger: %w", err)
	}
	if err := syncDirectory(directory, points.directorySync); err != nil {
		return fmt.Errorf("sync recovery directory: %w", err)
	}
	return nil
}
