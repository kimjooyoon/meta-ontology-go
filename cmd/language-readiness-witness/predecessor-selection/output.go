package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func encodeJSON(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	return append(raw, '\n'), err
}

func writeExclusive(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(raw); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func requireOutside(root, candidate string) error {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	candidatePath, err := filepath.Abs(candidate)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(rootPath, candidatePath)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("selection output must be outside repository")
	}
	return nil
}
