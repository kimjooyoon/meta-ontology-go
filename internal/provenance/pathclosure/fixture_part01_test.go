package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type inferenceFixture struct {
	path     semantic.InferencePathV1
	edges    []semantic.InferenceEdge
	evidence []semantic.InferenceEvidence
}

func fixtureDigest(value string) string {
	return semantic.StableHashString("pathclosure-test/" + value)
}
func fixtureID(value string) semantic.ID {
	return semantic.MustIdentity("pathclosure-test://" + value)
}
func manualInferenceEdge(
	kind semantic.InferenceKind, label string, subject, object semantic.ID,
) (semantic.InferenceEdge, semantic.InferenceEvidence) {
	phase := semantic.PhasePlacement{Ordinal: 1}
	authority := semantic.AuthorityBinding{}
	controls := semantic.InferenceControls{}
	switch kind {
	case semantic.InferenceAuthoritativeDeclaration:
		phase.Phase = semantic.PhaseDeclaration
		authority.Layer, authority.Effect = semantic.AuthoritySource, semantic.AuthorityDeclare
	case semantic.InferenceDeterministicDerivation:
		phase.Phase = semantic.PhaseDerivation
		authority.Layer, authority.Effect = semantic.AuthoritySemantic, semantic.AuthorityDerive
	case semantic.InferenceDerivedProjection:
		phase.Phase = semantic.PhaseProjection
		authority.Layer, authority.Effect = semantic.AuthorityDerived, semantic.AuthorityProject
		controls.Profile = semantic.ProfileBinding{
			ID: "pathclosure.profile.v1", Version: "1", Digest: fixtureDigest("profile"),
		}
	case semantic.InferenceIndependentVerification:
		phase.Phase = semantic.PhaseVerification
		authority.Layer, authority.Effect = semantic.AuthorityVerification, semantic.AuthorityVerify
		controls.PolicyDigest = fixtureDigest("policy")
	default:
		panic("unsupported manual path-closure fixture kind")
	}
	evidenceID := fixtureID("evidence/" + label)
	evidenceDigest := fixtureDigest("evidence-payload/" + label)
	edge := semantic.InferenceEdge{
		RecordID: fixtureID("record/" + label), SubjectID: subject, ObjectID: object,
		Rule:  semantic.RuleBinding{ID: fixtureID("rule/v1"), Version: "1", Digest: fixtureDigest("rule")},
		Phase: phase,
		Before: semantic.SnapshotDigests{
			Source: semantic.StableHashString("source-before/" + label), Semantic: fixtureDigest("semantic-before/" + label),
		},
		After: semantic.SnapshotDigests{
			Source: semantic.StableHashString("source-after/" + label), Semantic: fixtureDigest("semantic-after/" + label),
		},
		Authority: authority, Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: evidenceDigest}},
		Controls: controls,
		Kind:     kind,
	}
	if kind == semantic.InferenceAuthoritativeDeclaration {
		edge.SourceRoots = []semantic.ID{fixtureID("source-root/" + label)}
	}
	edgesEvidence := semantic.InferenceEvidence{
		ID: evidenceID, Digest: evidenceDigest, Before: edge.Before, After: edge.After, Controls: controls,
		SourceBacked: kind == semantic.InferenceAuthoritativeDeclaration,
		Independent:  kind == semantic.InferenceIndependentVerification,
	}
	return edge, edgesEvidence
}
