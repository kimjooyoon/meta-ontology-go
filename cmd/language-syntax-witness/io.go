package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
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
	actual := languagesyntax.Report{}
	if err := json.Unmarshal(data, &actual); err != nil {
		return err
	}
	if err := languagesyntax.Validate(actual, cfg.head); err != nil {
		return err
	}
	expected, err := build(cfg)
	if err != nil {
		return err
	}
	expectedData, _ := json.MarshalIndent(expected, "", "  ")
	if !bytes.Equal(data, append(expectedData, '\n')) {
		return fmt.Errorf("FAIL_CLOSED: language syntax replay mismatch")
	}
	printSummary(stdout, actual)
	return nil
}

func printSummary(output io.Writer, report languagesyntax.Report) {
	fmt.Fprintf(output, "language-syntax: decision=%s resolution=%s satisfied=%d/%d gooo_files=%d gooo_lines=%d writes=%d\n",
		report.Decision, report.Resolution, report.Summary.Satisfied, report.Summary.Total,
		len(report.Source.GoooFiles), report.Summary.GoooLines, report.RepositoryWrites)
}
