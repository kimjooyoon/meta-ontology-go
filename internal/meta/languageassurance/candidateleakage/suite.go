package candidateleakage

var definitions = []Definition{
	{ID: "isolated-candidate", ExpectedDecision: DecisionPass, ExpectedResolution: ResolutionExact,
		ExpectedEffect: EffectNone, ExpectedReason: ReasonCandidateIsolated},
	{ID: "authorized-transition", ExpectedDecision: DecisionPass, ExpectedResolution: ResolutionExact,
		ExpectedEffect: EffectNone, ExpectedReason: ReasonExactPromotionBound},
	{ID: "denied-fixed-point-leak", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionExact,
		ExpectedEffect: EffectBlock, ExpectedReason: ReasonLeakageDetected, ExpectedLeakage: 1},
	{ID: "fail-closed-allow-leak", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionExact,
		ExpectedEffect: EffectBlock, ExpectedReason: ReasonLeakageDetected, ExpectedLeakage: 1},
	{ID: "unknown-promotion", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant,
		ExpectedEffect: EffectBlock, ExpectedReason: ReasonDecisionUnknown, ExpectedUnknown: 1},
	{ID: "digest-mismatch", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant,
		ExpectedEffect: EffectBlock, ExpectedReason: ReasonDigestBindingMismatch, ExpectedUnknown: 1},
}

func DenominatorContract() Denominator {
	cases := append([]Definition(nil), definitions...)
	denominator := Denominator{ID: DenominatorID, Cases: cases}
	denominator.Digest = digestJSON(denominator)
	return denominator
}

func CaseInput(id, subjectSHA string) (Input, bool) {
	input := baseInput(subjectSHA)
	switch id {
	case "isolated-candidate":
	case "authorized-transition":
		input.Promotion.Decision = PromotionAuthorized
		makeOfficialPositive(&input, OfficialAllow)
	case "denied-fixed-point-leak":
		makeOfficialPositive(&input, OfficialFixedPoint)
	case "fail-closed-allow-leak":
		input.Promotion.Decision = PromotionFailClosed
		makeOfficialPositive(&input, OfficialAllow)
	case "unknown-promotion":
		input.Promotion.Decision = "UNKNOWN"
		input.Promotion.Resolution = ResolutionInvariant
	case "digest-mismatch":
		input.Promotion.Decision = PromotionAuthorized
		input.Promotion.CandidateDigest = digestText("different-candidate:" + subjectSHA)
		makeOfficialPositive(&input, OfficialAllow)
	default:
		return Input{}, false
	}
	return input, true
}

func RunSuite(subjectSHA string) Suite {
	denominator := DenominatorContract()
	suite := Suite{Schema: SuiteSchema, SubjectSHA: subjectSHA, DenominatorID: denominator.ID,
		DenominatorDigest: denominator.Digest, Decision: DecisionPass, Resolution: ResolutionExact}
	for _, definition := range denominator.Cases {
		input, _ := CaseInput(definition.ID, subjectSHA)
		report := Evaluate(input)
		passed := matches(definition, report) && replayEqual(report, Evaluate(input))
		suite.Cases = append(suite.Cases, CaseResult{Definition: definition, Passed: passed, Report: report})
		countCase(&suite.Summary, report, passed)
	}
	suite.Summary.CasesTotal = len(denominator.Cases)
	if suite.Summary.CasesPassed != suite.Summary.CasesTotal {
		suite.Decision, suite.Resolution = DecisionFailClosed, ResolutionInvariant
	}
	suite.Summary.CoverageBPS = suite.Summary.CasesPassed * 10_000 / suite.Summary.CasesTotal
	return sealSuite(suite)
}
