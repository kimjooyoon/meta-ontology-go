package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadJSON[T any](path string) (T, error) {
	var value T
	raw, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return value, err
	}
	return value, nil
}

func writeOutputs(interopPath, bindingPath string, interop, binding any) error {
	if err := writeJSON(interopPath, interop); err != nil {
		return err
	}
	return writeJSON(bindingPath, binding)
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
