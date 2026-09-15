package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	meta "github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactresolutionexperiment"
)

func run(args []string) int {
	options, err := parseOptions(args)
	if err != nil {
		return fail(err)
	}
	payload, err := os.ReadFile(options.input)
	if err != nil {
		return fail(err)
	}
	var input meta.Input
	if err := json.Unmarshal(payload, &input); err != nil {
		return fail(err)
	}
	report := meta.Evaluate(input)
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fail(err)
	}
	encoded = append(encoded, '\n')
	if options.check != "" {
		expected, readErr := os.ReadFile(options.check)
		if readErr != nil || !bytes.Equal(encoded, expected) {
			return 1
		}
		return 0
	}
	if err := os.WriteFile(options.output, encoded, 0o644); err != nil {
		return fail(err)
	}
	if report.Decision != "PASS" {
		return 1
	}
	return 0
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, err)
	return 2
}
