package semantic

import (
	"testing"
)

func inferenceTestDigest(value string) string { return StableHashString("inference-test/" + value) }
func inferenceEdgeFixture(kind InferenceKind, suffix string) InferenceEdge {
	semantic := inferenceTestDigest("semantic/" + suffix)
	before := SnapshotDigests{Source: inferenceTestDigest("source-before/" + suffix), Semantic: semantic}
	after := SnapshotDigests{Source: inferenceTestDigest("source-after/" + suffix), Semantic: semantic}
	controls := InferenceControls{}
	authority := AuthorityBinding{}
	phase := PhasePlacement{Ordinal: 1}
	switch kind {
	case InferenceAuthoritativeDeclaration:
		phase.Phase, authority.Layer, authority.Effect = PhaseDeclaration, AuthoritySource, AuthorityDeclare
	case InferenceDeterministicDerivation:
		phase.Phase, authority.Layer, authority.Effect = PhaseDerivation, AuthoritySemantic, AuthorityDerive
	case InferenceDerivedProjection:
		phase.Phase, authority.Layer, authority.Effect = PhaseProjection, AuthorityDerived, AuthorityProject
		controls.Profile = ProfileBinding{ID: "gooo.test.profile.v1", Version: "1", Digest: inferenceTestDigest("profile")}
	case InferenceObservationCandidate:
		phase.Phase, authority.Layer, authority.Effect = PhaseObservation, AuthorityCandidate, AuthorityObserve
		controls.CatalogDigest = inferenceTestDigest("catalog")
	case InferenceAcceptedLift:
		phase.Phase, authority.Layer, authority.Effect = PhaseLift, AuthoritySemantic, AuthorityLift
		controls.PolicyDigest = inferenceTestDigest("policy")
	case InferenceIndependentVerification:
		phase.Phase, authority.Layer, authority.Effect = PhaseVerification, AuthorityVerification, AuthorityVerify
		controls.PolicyDigest = inferenceTestDigest("policy")
	}
	evidenceID := MustIdentity("inference-test://evidence/" + suffix)
	edge := InferenceEdge{
		InferenceRecord: InferenceRecord{
			RecordID:  MustIdentity("inference-test://record/" + suffix),
			SubjectID: MustIdentity("inference-test://subject/" + suffix),
			ObjectID:  MustIdentity("inference-test://object/" + suffix),
			Rule: RuleBinding{
				ID: MustIdentity("inference-test://rule/v1"), Version: "1", Digest: inferenceTestDigest("rule"),
			},
			Phase: phase, Before: before, After: after, Authority: authority, Controls: controls,
			Evidence: []EvidenceReference{{ID: evidenceID, Digest: inferenceTestDigest("payload/" + suffix)}},
		},
	}
	if kind == InferenceAuthoritativeDeclaration {
		edge.SourceRoots = []ID{MustIdentity("inference-test://source/root/" + suffix)}
	}
	if kind == InferenceAcceptedLift {
		edge.AcceptanceReceipt = evidenceID
	}
	edge.Kind = kind
	return edge
}
func inferenceEvidenceFixture(edge InferenceEdge) InferenceEvidence {
	ref := edge.Evidence[0]
	return InferenceEvidence{
		ID: ref.ID, Digest: ref.Digest, Before: edge.Before, After: edge.After, Controls: edge.Controls,
		SourceBacked: edge.Kind == InferenceAcceptedLift,
		Independent:  edge.Kind == InferenceIndependentVerification,
	}
}
func inferenceBundle(edge InferenceEdge) InferencePathV1 {
	return InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{edge},
		Evidence: []InferenceEvidence{inferenceEvidenceFixture(edge)},
	}
}
func assertInferencePathRejected(t *testing.T, path InferencePathV1) {
	t.Helper()
	if err := path.Validate(); err == nil {
		t.Fatal("invalid inference path was accepted")
	}
}
