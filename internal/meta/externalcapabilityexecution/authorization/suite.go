package authorization

import "slices"

func exactInput(input Input) Input {
	input.Foundation = Foundation{
		Schema: FoundationSchema, Available: true,
		SubjectSHA: input.Invocation.SubjectSHA, ProducerRunID: "fixture-run",
		ArtifactID: 1, ArchiveDigest: digestValue("fixture-archive"),
		PolicySourceDigest:    input.Policy.SourceDigest,
		PolicyGeneratedDigest: input.Policy.GeneratedDigest,
	}
	return input
}

func expectedCase(caseID string) (string, string) {
	if caseID == "exact" {
		return DecisionAuthorized, ResolutionExact
	}
	if slices.Contains(caseIDs[1:8], caseID) {
		return DecisionFailClosed, ResolutionUnknown
	}
	return DecisionDenied, ResolutionExact
}

func RunSuite(input Input) Suite {
	suite := Suite{Schema: SuiteSchema, SubjectSHA: input.Invocation.SubjectSHA,
		Total: len(caseIDs)}
	for _, caseID := range caseIDs {
		receipt := runCase(input, caseID)
		expectedDecision, expectedResolution := expectedCase(caseID)
		passed := receipt.Decision == expectedDecision && receipt.Resolution == expectedResolution
		suite.Cases = append(suite.Cases, SuiteCase{CaseID: caseID,
			ExpectedDecision: expectedDecision, ExpectedResolution: expectedResolution,
			ActualDecision: receipt.Decision, ActualResolution: receipt.Resolution, Passed: passed})
		if passed {
			suite.Passed++
		}
		switch expectedDecision {
		case DecisionAuthorized:
			suite.AuthorizedCases++
		case DecisionFailClosed:
			suite.UnknownCases++
		case DecisionDenied:
			suite.DeniedCases++
		}
	}
	suite.CoverageBPS = suite.Passed * 10000 / suite.Total
	if suite.Passed == suite.Total && suite.Total == SuiteDenominator {
		suite.Decision, suite.Resolution = DecisionPass, ResolutionExact
	} else {
		suite.Decision, suite.Resolution = DecisionFailClosed, ResolutionExact
	}
	suite.SuiteDigest = digestValue(suite)
	return suite
}
