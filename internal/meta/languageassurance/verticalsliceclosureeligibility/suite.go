package verticalsliceclosureeligibility

import (
	_ "embed"
	"encoding/json"
)

//go:embed evidence/assurance.json
var assuranceEvidence []byte

//go:embed evidence/shadow.json
var shadowEvidence []byte

var definitions = []Definition{
	{ID: "exact", ExpectedDecision: DecisionEligible, ExpectedResolution: ResolutionExact, ExpectedEffect: EffectNone, ExpectedReason: ReasonEligible},
	{ID: "assurance-unavailable", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedEffect: EffectBlock, ExpectedReason: ReasonUnavailable},
	{ID: "shadow-unavailable", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedEffect: EffectBlock, ExpectedReason: ReasonUnavailable},
	{ID: "assurance-digest-mismatch", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant, ExpectedEffect: EffectBlock, ExpectedReason: ReasonDigestMismatch},
	{ID: "shadow-digest-mismatch", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant, ExpectedEffect: EffectBlock, ExpectedReason: ReasonDigestMismatch},
	{ID: "unknown-top-decision", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedEffect: EffectBlock, ExpectedReason: ReasonDecisionUnknown},
	{ID: "semantic-link-mismatch", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant, ExpectedEffect: EffectBlock, ExpectedReason: ReasonLinkMismatch},
	{ID: "observed-write", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant, ExpectedEffect: EffectBlock, ExpectedReason: ReasonWriteObserved},
}

func EmbeddedInput(subjectSHA string) Input {
	return Input{SubjectSHA: subjectSHA,
		Assurance: Capsule{Name: AssuranceName, ArtifactID: AssuranceArtifactID,
			ArchiveDigest: AssuranceArchiveDigest, CapsuleDigest: AssuranceCapsuleDigest,
			Payload: append([]byte(nil), assuranceEvidence...)},
		Shadow: Capsule{Name: ShadowName, ArtifactID: ShadowArtifactID,
			ArchiveDigest: ShadowArchiveDigest, CapsuleDigest: ShadowCapsuleDigest,
			Payload: append([]byte(nil), shadowEvidence...)}}
}

func CaseInput(id, subjectSHA string) (Input, bool) {
	input := EmbeddedInput(subjectSHA)
	switch id {
	case "exact":
	case "assurance-unavailable":
		input.Assurance.Payload = nil
	case "shadow-unavailable":
		input.Shadow.Payload = nil
	case "assurance-digest-mismatch":
		input.Assurance.Payload = append(input.Assurance.Payload, '
')
	case "shadow-digest-mismatch":
		input.Shadow.Payload = append(input.Shadow.Payload, '
')
	case "unknown-top-decision":
		rewriteShadow(&input, func(value *shadowCapsule) { value.Decision = "UNKNOWN" })
	case "semantic-link-mismatch":
		rewriteShadow(&input, func(value *shadowCapsule) { value.AssuranceDigest = "sha256:unknown" })
	case "observed-write":
		rewriteShadow(&input, func(value *shadowCapsule) { value.RepositoryWrites = 1 })
	default:
		return Input{}, false
	}
	return input, true
}

func rewriteShadow(input *Input, change func(*shadowCapsule)) {
	var value shadowCapsule
	_ = json.Unmarshal(input.Shadow.Payload, &value)
	change(&value)
	input.Shadow.Payload, _ = json.MarshalIndent(value, "", "  ")
}

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
		passed := report.Decision == definition.ExpectedDecision && report.Resolution == definition.ExpectedResolution &&
			report.EnforcementEffect == definition.ExpectedEffect && report.Reason == definition.ExpectedReason &&
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
	if suite.Schema != SuiteSchema || suite.SuiteDigest == "" ||
		suite.DenominatorDigest != suiteContract().DenominatorDigest {
		return json.InvalidUnmarshalError(nil)
	}
	replay := RunSuite(subjectSHA)
	if suite.SuiteDigest != replay.SuiteDigest {
		return json.InvalidUnmarshalError(nil)
	}
	return nil
}
