package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeEvidence(path string, report cutoverEvidence) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func requirePassing(report cutoverEvidence) error {
	for _, item := range report.Indicators {
		if item.Blocking && item.Value > item.Limit {
			return fmt.Errorf("blocking indicator %s=%d", item.ID, item.Value)
		}
	}
	return nil
}
