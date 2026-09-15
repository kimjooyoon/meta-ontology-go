package main

import (
	"encoding/json"
	"fmt"
	"io"

	activation "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/verticalsliceclosureactivation"
)

func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	input, err := readInput(opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	receipt := activation.Evaluate(input)
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := writeOutput(opts.output, append(raw, '\n')); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "vertical-slice-closure-activation: decision=%s resolution=%s applied=%d writes=%d\n", receipt.Decision, receipt.Resolution, receipt.TransitionApplied, receipt.RepositoryWrites)
	return 0
}
