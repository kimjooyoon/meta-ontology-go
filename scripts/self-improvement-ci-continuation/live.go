package main

import (
	"fmt"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcontinuation"
)

func runLive(program continuation.PolicyProgram, settings options) error {
	if settings.inputPath == "" {
		return fmt.Errorf("-input is required for live mode")
	}
	var input continuation.ContinuationInput
	if err := readJSON(settings.inputPath, &input); err != nil {
		return err
	}
	report := continuation.BuildReport(program, input)
	if settings.check {
		if err := continuation.ValidateReport(report); err != nil {
			return err
		}
		if !report.Verification.Verified {
			return fmt.Errorf("continuation report independent verification failed")
		}
	}
	return writeJSON(settings.outputPath, report)
}
