package assuranceeligibility

import "reflect"

func RunSuite(base Input) Suite {
	suite := Suite{Schema: SuiteSchema, SubjectSHA: base.SubjectSHA, DenominatorID: SuiteDenominator,
		Decision: DecisionEligible, Resolution: ResolutionExact, Cases: make([]CaseResult, len(definitions))}
	suite.DenominatorDigest = digestJSON(definitions)
	for index, definition := range definitions {
		input, _ := CaseInput(base, definition.ID)
		report := Evaluate(input)
		passed := report.Decision == definition.ExpectedDecision &&
			report.Resolution == definition.ExpectedResolution &&
			report.EnforcementEffect == definition.ExpectedEffect && report.Reason == definition.ExpectedReason
		suite.Cases[index] = CaseResult{Definition: definition, Passed: passed, Report: report}
		if passed { suite.Passed++ }
		switch report.Resolution {
		case ResolutionExact: suite.ExactExpected++
		case ResolutionUnknown: suite.UnknownExpected++
		case ResolutionInvariant: suite.InvariantExpected++
		}
	}
	suite.Total = len(definitions)
	suite.CoverageBPS = suite.Passed * 10000 / suite.Total
	if suite.Passed != suite.Total || suite.ExactExpected != 1 ||
		suite.UnknownExpected != 10 || suite.InvariantExpected != 9 {
		suite.Decision, suite.Resolution = DecisionFailClosed, ResolutionInvariant
	}
	return sealSuite(suite)
}

func Validate(report Report, input Input) bool {
	return report.Schema == ReportSchema && reflect.DeepEqual(report, Evaluate(input))
}

func ValidateSuite(suite Suite, input Input) bool {
	return suite.Schema == SuiteSchema && reflect.DeepEqual(suite, RunSuite(input))
}
