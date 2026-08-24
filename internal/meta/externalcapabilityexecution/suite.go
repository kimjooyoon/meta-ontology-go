package externalcapabilityexecution

func RunSuite(subject string) Suite {
	suite := Suite{
		Schema: SuiteSchema, SubjectSHA: subject, Total: len(caseIDs),
		RepositoryWrites: 0, ExternalExecutions: 0,
	}
	for _, caseID := range caseIDs {
		report := RunCase(subject, caseID)
		expectedDecision, expectedResolution := expectedCase(caseID)
		passed := report.Decision == expectedDecision && report.Resolution == expectedResolution
		suite.Cases = append(suite.Cases, SuiteCase{
			CaseID: caseID, ExpectedDecision: expectedDecision,
			ExpectedResolution: expectedResolution, ActualDecision: report.Decision,
			ActualResolution: report.Resolution, Passed: passed,
		})
		if passed {
			suite.Passed++
		}
		switch expectedResolution {
		case ResolutionExact:
			suite.ExactExpected++
		case ResolutionUnknown:
			suite.UnknownExpected++
		case ResolutionInvariant:
			suite.InvariantExpected++
		}
	}
	suite.CoverageBPS = suite.Passed * 10000 / suite.Total
	if suite.Passed == suite.Total {
		suite.Decision, suite.Resolution = DecisionExecutable, ResolutionExact
	} else {
		suite.Decision, suite.Resolution = DecisionFailClosed, ResolutionInvariant
	}
	suite.SuiteDigest = ""
	suite.SuiteDigest = digestValue(suite)
	return suite
}
