package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"
)

func run(args []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if options.caseID == "suite" {
		return writeJSON(options.output, runSuite(options.subjectSHA, options.observedCheckoutSHA, options.effectsBefore, options.effectsAfter, options.output), stdout, stderr)
	}
	input, err := inputFor(options)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(options.output, evaluate(input, options.output), stdout, stderr)
}

func inputFor(options options) (producer.Input, error) {
	if options.caseID != "" {
		for _, definition := range producer.Denominator() {
			if definition.ID == options.caseID {
				return producer.Input{CaseID: definition.ID, SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA, BeforePath: definition.BeforePath, AfterPath: definition.AfterPath, EffectsBeforePath: options.effectsBefore, EffectsAfterPath: options.effectsAfter, OutputPath: options.output}, nil
			}
		}
		return producer.Input{}, fmt.Errorf("unknown fixed case %q", options.caseID)
	}
	return producer.Input{CaseID: "file-pair", SubjectSHA: options.subjectSHA, ObservedCheckoutSHA: options.observedCheckoutSHA, BeforePath: options.before, AfterPath: options.after, EffectsBeforePath: options.effectsBefore, EffectsAfterPath: options.effectsAfter, OutputPath: options.output}, nil
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
