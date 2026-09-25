package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

func produce(cfg config, stdout io.Writer) error {
	report, err := build(cfg)
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(report, "", "  ")
	data = append(data, '\n')
	if err := os.WriteFile(cfg.output, data, 0o644); err != nil {
		return err
	}
	printSummary(stdout, report)
	return nil
}

func consume(cfg config, stdout io.Writer) error {
	data, err := os.ReadFile(cfg.check)
	if err != nil {
		return err
	}
	actual := toolchainusecases.Report{}
	if err := json.Unmarshal(data, &actual); err != nil {
		return err
	}
	if err := toolchainusecases.Validate(actual, cfg.head); err != nil {
		return err
	}
	expected, err := build(cfg)
	if err != nil {
		return err
	}
	expectedData, _ := json.MarshalIndent(expected, "", "  ")
	if !bytes.Equal(data, append(expectedData, '\n')) {
		return fmt.Errorf("FAIL_CLOSED: executable use case replay mismatch")
	}
	printSummary(stdout, actual)
	return nil
}
