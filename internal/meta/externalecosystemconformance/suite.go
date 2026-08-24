package externalecosystemconformance

import "encoding/json"

func RunSuite(subject string, capsule Capsule, evidence Evidence) Suite {
	suite := Suite{
		Schema: SuiteSchema, SubjectSHA: subject,
		CasesTotal: len(caseIDs), RepositoryWrites: evidence.RepositoryWrites,
		ExternalExecutions: evidence.ExternalExecutions,
	}
	for _, caseID := range caseIDs {
		report := RunCase(subject, caseID, capsule, evidence)
		expectedDecision, expectedResolution := expectedCase(caseID)
		passed := report.Decision == expectedDecision && report.Resolution == expectedResolution
		suite.Cases = append(suite.Cases, SuiteCase{
			CaseID: caseID, ExpectedDecision: expectedDecision,
			ExpectedResolution: expectedResolution, ActualDecision: report.Decision,
			ActualResolution: report.Resolution, Passed: passed,
		})
		if passed {
			suite.CasesPassed++
		}
		switch {
		case report.Decision == DecisionReferenceBound && report.Resolution == ResolutionExact:
			suite.ReferenceBoundExact++
		case report.Decision == DecisionFailClosed && report.Resolution == ResolutionUnknown:
			suite.UnknownFailClosed++
		case report.Decision == DecisionFailClosed && report.Resolution == ResolutionInvariant:
			suite.InvariantFailClosed++
		}
	}
	suite.CoverageBPS = suite.CasesPassed * 10000 / suite.CasesTotal
	if suite.CasesPassed == suite.CasesTotal {
		suite.Decision = DecisionReferenceBound
		suite.Resolution = ResolutionExact
	} else {
		suite.Decision = DecisionFailClosed
		suite.Resolution = ResolutionInvariant
	}
	suite.SuiteDigest = ""
	raw, _ := json.Marshal(suite)
	suite.SuiteDigest = digest(raw)
	return suite
}
