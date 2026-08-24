package verticalsliceclosureeligibility

import _ "embed"

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
