package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
)

func readReadiness(path string) ([]byte, readinessartifact.Receipt, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, readinessartifact.Receipt{}, err
	}
	var receipt readinessartifact.Receipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return nil, readinessartifact.Receipt{}, err
	}
	return raw, receipt, nil
}

func readBaselineReference(path string) (readinessartifact.BaselineReference, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return readinessartifact.BaselineReference{}, err
	}
	var reference readinessartifact.BaselineReference
	if err := json.Unmarshal(raw, &reference); err != nil {
		return readinessartifact.BaselineReference{}, err
	}
	return reference, nil
}

func encode(value readinessartifact.ImprovementArtifact) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func writeArtifact(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func compareArtifact(path string, actual []byte) error {
	expected, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(expected, actual) {
		return fmt.Errorf("FAIL_CLOSED: improvement artifact replay mismatch")
	}
	return nil
}
