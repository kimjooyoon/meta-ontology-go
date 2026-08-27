package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func run(args []string) error {
	value, err := parseOptions(args)
	if err != nil {
		return err
	}
	contract, observation, err := loadInput(value)
	if err != nil {
		return err
	}
	report, err := workgraphEvaluate(contract, observation)
	if err != nil {
		return err
	}
	if value.resourceOut != "" {
		if err := writeJSON(value.resourceOut, observation.Resource); err != nil {
			return err
		}
	}
	if err := writeJSON(value.output, report); err != nil {
		return err
	}
	if report.Decision != value.expect {
		return fmt.Errorf("decision = %s, want %s", report.Decision, value.expect)
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
