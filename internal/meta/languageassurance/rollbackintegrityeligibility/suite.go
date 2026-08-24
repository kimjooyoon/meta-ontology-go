package rollbackintegrityeligibility

import _ "embed"

//go:embed evidence/assurance.json
var assuranceEvidence []byte

//go:embed evidence/report-a.json
var reportAEvidence []byte

//go:embed evidence/report-b.json
var reportBEvidence []byte

var definitions = []Definition{
	{ID: "exact", ExpectedDecision: DecisionEligible, ExpectedResolution: ResolutionExact, ExpectedReason: ReasonEligible},
	{ID: "unavailable", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedReason: ReasonUnavailable},
	{ID: "digest-mismatch", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant, ExpectedReason: ReasonDigestMismatch},
	{ID: "subject-unknown", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedReason: ReasonSubjectUnknown},
}

func EmbeddedInput(subjectSHA string) Input {
	return Input{SubjectSHA: subjectSHA,
		Assurance: Capsule{Name: "language-assurance", ArtifactID: AssuranceArtifactID,
			ArtifactDigest: AssuranceArtifactDigest, CapsuleDigest: AssuranceCapsuleDigest,
			Payload: append([]byte(nil), assuranceEvidence...)},
		ShadowReportA: Capsule{Name: "rollback-integrity-shadow-replay-a", ArtifactID: ShadowArtifactID,
			ArtifactDigest: ShadowArtifactDigest, CapsuleDigest: ShadowReportACapsuleDigest,
			Payload: append([]byte(nil), reportAEvidence...)},
		ShadowReportB: Capsule{Name: "rollback-integrity-shadow-replay-b", ArtifactID: ShadowArtifactID,
			ArtifactDigest: ShadowArtifactDigest, CapsuleDigest: ShadowReportBCapsuleDigest,
			Payload: append([]byte(nil), reportBEvidence...)}}
}

func CaseInput(id, subjectSHA string) (Input, bool) {
	input := EmbeddedInput(subjectSHA)
	switch id {
	case "exact":
	case "unavailable":
		input.Assurance.Payload = nil
	case "digest-mismatch":
		input.ShadowReportB.Payload = append(input.ShadowReportB.Payload, '\n')
	case "subject-unknown":
		input.SubjectSHA = EvidenceSubjectSHA
	default:
		return Input{}, false
	}
	return input, true
}

func denominatorContract() Suite {
	suite := Suite{Schema: SuiteSchema, DenominatorID: DenominatorID, Cases: make([]CaseResult, len(definitions))}
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
		passed := report.Decision == definition.ExpectedDecision && report.Resolution == definition.ExpectedResolution &&
			report.Reason == definition.ExpectedReason && report.RepositoryWrites == 0 && report.PromotionApplied == 0
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
