package main

import (
	"fmt"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcontinuation"
)

func runCases(program continuation.PolicyProgram, settings options) error {
	report, err := continuation.BuildCanonicalCaseReport(program)
	if err != nil {
		return err
	}
	if settings.check {
		if err := continuation.ValidateCanonicalCases(report); err != nil {
			return err
		}
		if report.CaseDenominator != 9 || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || !report.ReplayEqual || report.Metrics.LiveGrantDecision != 0 || report.Metrics.LiveGrants != 0 || report.Metrics.LiveExecutionCount != 0 || report.Metrics.GrantConsumedUses != 0 || report.Metrics.RepositoryWrites != 0 || report.Metrics.LocalTestExecutions != 0 {
			return fmt.Errorf("continuation canonical metrics are not exact")
		}
	}
	return writeJSON(settings.outputPath, report)
}
