package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
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

func writeOrCheck(output, check string, value metaff.Report) error {
	if check != "" {
		existing, err := readComparedReport(check)
		if err != nil {
			return err
		}
		return metaff.CompareReplayReports(existing, value, value.Source.ExpectedHeadSHA)
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
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
