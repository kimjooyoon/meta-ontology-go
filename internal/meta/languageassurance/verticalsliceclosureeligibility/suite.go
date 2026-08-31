package verticalsliceclosureeligibility

import (
	"fmt"
	"reflect"
)

func suiteContract() Suite {
	suite := Suite{Schema: SuiteSchema, DenominatorID: SuiteDenominator,
		Cases: make([]CaseResult, len(definitions))}
	for index, definition := range definitions {
		suite.Cases[index].Definition = definition
	}
	suite.DenominatorDigest = digestJSON(suite.Cases)
	return suite
}

func RunSuite(subjectSHA string) Suite {
	suite := suiteContract()
	suite.SubjectSHA, suite.Decision, suite.Resolution = subjectSHA, DecisionEligible, ResolutionExact
	for index, definition := range definitions {
		input, _ := CaseInput(definition.ID, subjectSHA)
		report := Evaluate(input)
		passed := report.Decision == definition.ExpectedDecision &&
			report.Resolution == definition.ExpectedResolution &&
			report.EnforcementEffect == definition.ExpectedEffect &&
			report.Reason == definition.ExpectedReason &&
			report.RepositoryWrites == 0 && report.PromotionApplied == 0
		suite.Cases[index] = CaseResult{Definition: definition, Passed: passed, Report: report}
		if passed {
			suite.CasesPassed++
		}
		switch {
		case report.Decision == DecisionEligible && report.Resolution == ResolutionExact:
			suite.EligibleExact++
		case report.Resolution == ResolutionUnknown:
			suite.UnknownFailClosed++
		case report.Resolution == ResolutionInvariant:
			suite.InvariantFailClosed++
		}
	}
	suite.CasesTotal = len(definitions)
	if suite.CasesPassed != suite.CasesTotal {
		suite.Decision, suite.Resolution = DecisionFailClosed, ResolutionInvariant
	}
	suite.CoverageBPS = suite.CasesPassed * 10_000 / suite.CasesTotal
	return sealSuite(suite)
}

func ValidateSuite(suite Suite, subjectSHA string) error {
	if suite.Schema != SuiteSchema || !reflect.DeepEqual(suite, RunSuite(subjectSHA)) {
		return fmt.Errorf("vertical slice eligibility suite does not replay")
	}
	return nil
}
