package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	receipt "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
)

func run(args []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if options.caseID == "suite" {
		return writeJSON(options.output, receipt.RunSuite(options.subjectSHA), stdout, stderr)
	}
	input, err := inputFor(options)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(options.output, receipt.Evaluate(input), stdout, stderr)
}

func inputFor(options options) (receipt.Input, error) {
	if options.caseID != "" {
		if options.caseID != "equivalent" && options.caseID != "semantic-change" && options.caseID != "indeterminate" {
			return receipt.Input{}, fmt.Errorf("unknown fixed case %q", options.caseID)
		}
		return receipt.CaseInput(options.caseID, options.subjectSHA), nil
	}
	before, err := os.ReadFile(options.before)
	if err != nil {
		return receipt.Input{}, fmt.Errorf("read before source: %w", err)
	}
	after, err := os.ReadFile(options.after)
	if err != nil {
		return receipt.Input{}, fmt.Errorf("read after source: %w", err)
	}
	return receipt.Input{CaseID: "file-pair", SubjectSHA: options.subjectSHA, Before: before, After: after}, nil
}

func writeJSON(path string, value any, stdout, stderr io.Writer) int {
	file, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		_ = file.Close()
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := file.Close(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, path)
	return 0
}
