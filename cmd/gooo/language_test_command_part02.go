package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagetest"
)

func writeLanguageTestReceipt(receipt languagetest.Receipt, jsonMode bool, stdout, stderr io.Writer) int {
	if jsonMode {
		payload, err := languagetest.Marshal(receipt)
		if err != nil {
			fmt.Fprintf(stderr, "gooo test: %v\n", err)
			return exitFailure
		}
		_, _ = stdout.Write(payload)
	} else {
		fmt.Fprintf(stdout, "language tests: %s passed=%d failed=%d declared=%d\n",
			receipt.Decision, receipt.Summary.Passed, receipt.Summary.Failed, receipt.Summary.Declared)
	}
	if receipt.Decision == languagetest.DecisionPass {
		return exitOK
	}
	return exitFailure
}
