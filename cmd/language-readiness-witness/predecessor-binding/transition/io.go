package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorbinding"
)

func readReport(path string) (predecessorbinding.Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return predecessorbinding.Report{}, err
	}
	var report predecessorbinding.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		return predecessorbinding.Report{}, err
	}
	return report, nil
}

func encode(value predecessorbinding.BindingTransition) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	return append(raw, '\n'), err
}

func writeOrCheck(value config, raw []byte) error {
	if value.output != "" {
		return os.WriteFile(value.output, raw, 0o644)
	}
	expected, err := os.ReadFile(value.check)
	if err != nil {
		return err
	}
	if !bytes.Equal(expected, raw) {
		return fmt.Errorf("binding transition replay mismatch")
	}
	return nil
}
