package main

import (
	"errors"
	"fmt"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
)

func runCases(program grant.PolicyProgram, settings options) error {
	report, err := grant.BuildCanonicalCaseReport(program)
	if err != nil {
		return err
	}
	if settings.check {
		if err := grant.ValidateCanonicalCases(report); err != nil {
			return err
		}
		if report.CaseDenominator != 9 || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || report.CanonicalGrantedCases != 2 || report.CanonicalExecutionCount != 0 || report.GrantConsumedUses != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.FallbackAccepted != 0 {
			return errors.New("canonical execution grant metrics are not exact")
		}
		for _, current := range report.Cases {
			if !current.Pass {
				return fmt.Errorf("canonical case %q failed", current.ID)
			}
		}
	}
	return writeJSON(settings.outputPath, report)
}
