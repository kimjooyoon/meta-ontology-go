package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func readJSON[T any](path string) (T, error) {
	value := new(T)
	raw, err := os.ReadFile(path)
	if err != nil {
		return *value, err
	}
	if err := json.Unmarshal(raw, value); err != nil {
		return *value, err
	}
	return *value, nil
}

func writeOrCheck(output, check string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if check != "" {
		existing, err := os.ReadFile(check)
		if err != nil {
			return err
		}
		if !bytes.Equal(existing, raw) {
			return fmt.Errorf("FAIL_CLOSED: diagnostic provenance replay mismatch")
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	return os.WriteFile(output, raw, 0o644)
}
