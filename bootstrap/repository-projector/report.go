package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeEvidence(work string, report evidence) error {
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	name := filepath.Join(work, "evidence.json")
	return os.WriteFile(name, append(encoded, '\n'), 0o644)
}

func requireBlockingZero(report evidence) error {
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}
