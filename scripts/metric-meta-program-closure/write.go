package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeOutputs(directory string, receipt, verification []byte) error {
	if directory == "" {
		return fmt.Errorf("explicit output directory is required")
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "closure.json"), receipt, 0o644); err != nil {
		return fmt.Errorf("write closure: %w", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "verification.json"), verification, 0o644); err != nil {
		return fmt.Errorf("write verification: %w", err)
	}
	return nil
}
