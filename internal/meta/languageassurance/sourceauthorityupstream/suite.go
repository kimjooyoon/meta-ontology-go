package sourceauthorityupstream

import (
	"context"
	"strings"
)

type denominatorCase struct {
	ID          string `json:"id"`
	Observation string `json:"observation"`
	Resolution  string `json:"resolution"`
	Enforcement string `json:"enforcement"`
	Reason      string `json:"reason"`
}

var fixedDenominator = []denominatorCase{
	{CaseExact, ObservationSatisfied, ResolutionExact, EnforcementAllow, ReasonSourceSnapshotExact},
	{CaseDigestMismatch, ObservationUnknown, ResolutionInvariantOnly, EnforcementBlock, ReasonSourceDigestMismatch},
	{CaseAuthorityMismatch, ObservationUnknown, ResolutionInvariantOnly, EnforcementBlock, ReasonAuthorityScopeMismatch},
}

func RunSuite(ctx context.Context, subjectSHA string, fetcher Fetcher) Suite {
	policy := GomacroPolicy()
	request := GomacroRequest(subjectSHA)
	digestPolicy := policy
	digestPolicy.ExpectedDigest = "sha256:" + strings.Repeat("0", 64)
	authorityRequest := request
	authorityRequest.Authority.Repository = "cosmos72/not-gomacro"
	receipts := map[string]Receipt{
		CaseExact: Observe(ctx, policy, request, fetcher),
		CaseDigestMismatch: Observe(ctx, digestPolicy, request, fetcher),
		CaseAuthorityMismatch: Observe(ctx, policy, authorityRequest, fetcher),
	}
	suite := Suite{Schema: SuiteSchema, SubjectSHA: subjectSHA, DenominatorID: DenominatorID, DenominatorDigest: digestValue(fixedDenominator)}
	for _, expected := range fixedDenominator {
		receipt := receipts[expected.ID]
		passed := matches(expected, receipt)
		suite.Cases = append(suite.Cases, CaseResult{
			ID: expected.ID, ExpectedObservation: expected.Observation, ExpectedResolution: expected.Resolution,
			ExpectedEnforcement: expected.Enforcement, ExpectedReason: expected.Reason, Passed: passed, Receipt: receipt,
		})
		if passed {
			suite.Summary.CasesPassed++
		}
		suite.RepositoryWrites += receipt.RepositoryWrites
		suite.PromotionCreditBPS += receipt.PromotionCreditBPS
	}
	suite.Summary.CasesTotal = len(fixedDenominator)
	suite.Summary.ExactAllow = 1
	suite.Summary.FailClosed = 2
	suite.Summary.CoverageBPS = suite.Summary.CasesPassed * 10000 / suite.Summary.CasesTotal
	suite.Decision, suite.Resolution, suite.Reason = "BLOCK", ResolutionInvariantOnly, ReasonConformanceCaseMismatch
	if suite.Summary.CasesPassed == suite.Summary.CasesTotal && suite.RepositoryWrites == 0 && suite.PromotionCreditBPS == 0 {
		suite.Decision, suite.Resolution, suite.Reason = "PASS", ResolutionExact, ReasonConformanceExact
	}
	unsigned := suite
	unsigned.SuiteDigest = ""
	suite.SuiteDigest = digestValue(unsigned)
	return suite
}

func matches(expected denominatorCase, receipt Receipt) bool {
	return receipt.Observation == expected.Observation && receipt.Resolution == expected.Resolution &&
		receipt.Enforcement == expected.Enforcement && receipt.Reason == expected.Reason &&
		receipt.RepositoryWrites == 0 && receipt.PromotionCreditBPS == 0
}
