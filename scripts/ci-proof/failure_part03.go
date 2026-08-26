package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func writeFailureManifest(inputPath, outputPath string) error {
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read failure input: %w", err)
	}
	var input failureInput
	if err := json.Unmarshal(inputData, &input); err != nil {
		return fmt.Errorf("parse failure input: %w", err)
	}
	binding, err := readFailureBinding()
	if err != nil {
		return err
	}
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal failure manifest: %w", err)
	}
	if err := os.WriteFile(outputPath, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write failure manifest: %w", err)
	}
	return nil
}
