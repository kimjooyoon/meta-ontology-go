package main

import (
	"encoding/json"
	"fmt"
	"io"

	receipts "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/changedsurfacereceipt"
)

func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	var value any
	if opts.caseID == "suite" {
		value = receipts.EvaluateSuite(opts.subjectSHA)
	} else {
		value = receipts.Evaluate(receipts.CaseInput(opts.caseID, opts.subjectSHA))
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	raw = append(raw, '\n')
	if err := writeOutput(opts.output, raw); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "changed-surface-receipt: case=%s\n", opts.caseID)
	return 0
}
