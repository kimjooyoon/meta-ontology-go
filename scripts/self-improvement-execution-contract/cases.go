package main

import (
	"errors"
	"fmt"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
)

func runCases(program contract.PolicyProgram, settings options) error {
	report := contract.BuildCanonicalCaseReport(program)
	if settings.check {
		if report.CaseDenominator != 9 || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || !report.ReplayEqual || report.LiveExecutionCount != 0 || report.CanonicalExecutionCount != 0 || report.ExecutionGrants != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 {
			return errors.New("canonical execution contract case check failed")
		}
		for _, current := range report.Cases {
			if !current.Pass {
				return fmt.Errorf("canonical case %q failed", current.ID)
			}
		}
	}
	return writeJSON(settings.outputPath, report)
}
