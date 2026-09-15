package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readJSON[T any](filename string) (T, error) {
	value := new(T)
	raw, err := os.ReadFile(filename)
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
			return fmt.Errorf("FAIL_CLOSED: toolchain format/fix replay mismatch")
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	return os.WriteFile(output, raw, 0o644)
}

func requireExternal(root string, paths ...string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			return err
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("path must remain outside repository: %s", path)
		}
	}
	return nil
}
