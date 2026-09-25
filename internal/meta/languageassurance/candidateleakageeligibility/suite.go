package candidateleakageeligibility

import _ "embed"

//go:embed evidence/assurance.json
var assuranceEvidence []byte

//go:embed evidence/shadow.json
var shadowEvidence []byte

var definitions = []Definition{
	{ID: "exact", ExpectedDecision: DecisionEligible, ExpectedResolution: ResolutionExact,
		ExpectedReason: ReasonEligible},
	{ID: "unavailable", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown,
		ExpectedReason: ReasonUnavailable},
	{ID: "digest-mismatch", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant,
		ExpectedReason: ReasonDigestMismatch},
}

func EmbeddedInput(subjectSHA string) Input {
	return Input{SubjectSHA: subjectSHA,
		Assurance: Capsule{Name: "language-assurance", ArtifactID: AssuranceArtifactID,
			ArtifactDigest: AssuranceArtifactDigest, CapsuleDigest: AssuranceCapsuleDigest,
			Payload: append([]byte(nil), assuranceEvidence...)},
		Shadow: Capsule{Name: "candidate-leakage-shadow", ArtifactID: ShadowArtifactID,
			ArtifactDigest: ShadowArtifactDigest, CapsuleDigest: ShadowCapsuleDigest,
			Payload: append([]byte(nil), shadowEvidence...)}}
}

func CaseInput(id, subjectSHA string) (Input, bool) {
	input := EmbeddedInput(subjectSHA)
	switch id {
	case "exact":
	case "unavailable":
		input.Assurance.Payload = nil
	case "digest-mismatch":
		input.Assurance.Payload = append(input.Assurance.Payload, '\n')
	default:
		return Input{}, false
	}
	return input, true
}

func denominatorContract() Suite {
	suite := Suite{Schema: SuiteSchema, DenominatorID: DenominatorID,
		Cases: make([]CaseResult, len(definitions))}
	for index, definition := range definitions {
		suite.Cases[index].Definition = definition
	}
	suite.DenominatorDigest = digestJSON(suite.Cases)
	return suite
}

func RunSuite(subjectSHA string) Suite {
	suite := denominatorContract()
	suite.SubjectSHA, suite.Decision, suite.Resolution = subjectSHA, DecisionEligible, ResolutionExact
	for index, definition := range definitions {
		input, _ := CaseInput(definition.ID, subjectSHA)
		report := Evaluate(input)
		passed := report.Decision == definition.ExpectedDecision && report.Resolution ==
			definition.ExpectedResolution && report.Reason == definition.ExpectedReason && report.RepositoryWrites == 0 &&
			report.PromotionApplied == 0
		suite.Cases[index] = CaseResult{Definition: definition, Passed: passed, Report: report}
		if passed {
			suite.CasesPassed++
		}
	}
	suite.CasesTotal = len(definitions)
	if suite.CasesPassed != suite.CasesTotal {
		suite.Decision, suite.Resolution = DecisionFailClosed, ResolutionInvariant
	}
	suite.CoverageBPS = suite.CasesPassed * 10_000 / suite.CasesTotal
	return sealSuite(suite)
}
