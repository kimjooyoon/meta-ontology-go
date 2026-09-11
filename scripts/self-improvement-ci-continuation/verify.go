package main

import (
	"fmt"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcontinuation"
)

func runVerify(program continuation.PolicyProgram, settings options) error {
	if settings.reportPath == "" {
		return fmt.Errorf("-report is required for verify mode")
	}
	var report continuation.Report
	if err := readJSON(settings.reportPath, &report); err != nil {
		return err
	}
	verification := continuation.Verify(program, report.Request, report.Resolution)
	if settings.check && (!verification.Verified || verification.IndependentReplayComparisons != 1) {
		return fmt.Errorf("continuation verification failed: %#v", verification)
	}
	return writeJSON(settings.outputPath, verification)
}
