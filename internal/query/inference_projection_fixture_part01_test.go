package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func inferenceQueryDigest(value string) string {
	return semantic.StableHashString("query-inference/" + value)
}
func inferenceQueryID(value string) semantic.ID {
	return semantic.MustIdentity("inference-query://" + value)
}
func inferenceQueryEdge(
	kind semantic.InferenceKind, suffix string, subject, object semantic.ID,
) (semantic.InferenceEdge, semantic.InferenceEvidence) {
	semanticDigest := inferenceQueryDigest("semantic/" + suffix)
	controls := semantic.InferenceControls{}
	phase := semantic.PhasePlacement{Ordinal: 1}
	authority := semantic.AuthorityBinding{}
	switch kind {
	case semantic.InferenceAuthoritativeDeclaration:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseDeclaration, semantic.AuthoritySource, semantic.AuthorityDeclare
	case semantic.InferenceDeterministicDerivation:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseDerivation, semantic.AuthoritySemantic, semantic.AuthorityDerive
	case semantic.InferenceDerivedProjection:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseProjection, semantic.AuthorityDerived, semantic.AuthorityProject
		controls.Profile = semantic.ProfileBinding{ID: "query-inference.profile", Version: "1", Digest: inferenceQueryDigest("profile")}
	case semantic.InferenceObservationCandidate:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseObservation, semantic.AuthorityCandidate, semantic.AuthorityObserve
		controls.CatalogDigest = inferenceQueryDigest("catalog")
	case semantic.InferenceAcceptedLift:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseLift, semantic.AuthoritySemantic, semantic.AuthorityLift
		controls.PolicyDigest = inferenceQueryDigest("policy")
	case semantic.InferenceIndependentVerification:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseVerification, semantic.AuthorityVerification, semantic.AuthorityVerify
		controls.PolicyDigest = inferenceQueryDigest("policy")
	default:
		panic("unsupported test inference kind")
	}
	evidenceID := inferenceQueryID("evidence/" + suffix)
	edge := semantic.InferenceEdge{
		RecordID: inferenceQueryID("record/" + suffix), SubjectID: subject, ObjectID: object,
		Rule:      semantic.RuleBinding{ID: inferenceQueryID("rule/v1"), Version: "1", Digest: inferenceQueryDigest("rule")},
		Phase:     phase,
		Before:    semantic.SnapshotDigests{Source: inferenceQueryDigest("source/" + suffix), Semantic: semanticDigest},
		After:     semantic.SnapshotDigests{Source: inferenceQueryDigest("source/" + suffix), Semantic: semanticDigest},
		Authority: authority, Controls: controls,
		Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: inferenceQueryDigest("evidence-payload/" + suffix)}},
		Kind:     kind,
	}
	if kind == semantic.InferenceAuthoritativeDeclaration {
		edge.SourceRoots = []semantic.ID{inferenceQueryID("source-root/" + suffix)}
	}
	if kind == semantic.InferenceAcceptedLift {
		edge.AcceptanceReceipt = evidenceID
	}
	evidence := semantic.InferenceEvidence{
		ID: evidenceID, Digest: edge.Evidence[0].Digest, Before: edge.Before, After: edge.After,
		SourceBacked: kind == semantic.InferenceAcceptedLift,
		Independent:  kind == semantic.InferenceIndependentVerification,
		Controls:     controls,
	}
	return edge, evidence
}
