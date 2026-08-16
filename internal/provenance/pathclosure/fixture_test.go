package pathclosure_test

import (
	"reflect"
	"testing"

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
		InferenceRecord: semantic.InferenceRecord{
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
		},
		Kind: kind,
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

func completeInferenceFixture() inferenceFixture {
	source := fixtureID("node/source-declaration")
	declared := fixtureID("node/declared-activity")
	derived := fixtureID("node/deterministic-derivation")
	projected := fixtureID("node/generated-projection")
	verified := fixtureID("node/independent-verification")
	declaration, declarationEvidence := manualInferenceEdge(
		semantic.InferenceAuthoritativeDeclaration, "01-declaration", source, declared,
	)
	derivation, derivationEvidence := manualInferenceEdge(
		semantic.InferenceDeterministicDerivation, "02-derivation", declared, derived,
	)
	projection, projectionEvidence := manualInferenceEdge(
		semantic.InferenceDerivedProjection, "03-projection", derived, projected,
	)
	verification, verificationEvidence := manualInferenceEdge(
		semantic.InferenceIndependentVerification, "04-verification", projected, verified,
	)
	edges := []semantic.InferenceEdge{declaration, derivation, projection, verification}
	evidence := []semantic.InferenceEvidence{
		declarationEvidence, derivationEvidence, projectionEvidence, verificationEvidence,
	}
	return inferenceFixture{
		path:  semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence},
		edges: edges, evidence: evidence,
	}
}

func exactEdgeSequence(t *testing.T, edges []semantic.InferenceEdge) []string {
	t.Helper()
	sequence := make([]string, 0, len(edges))
	for _, edge := range edges {
		if err := edge.Validate(); err != nil {
			t.Fatalf("edge %s is invalid: %v", edge.RecordID, err)
		}
		sequence = append(sequence, edge.Canonical())
	}
	return sequence
}

func assertExactRecordSequence(t *testing.T, got, want []semantic.InferenceEdge) {
	t.Helper()
	gotSequence := exactEdgeSequence(t, got)
	wantSequence := exactEdgeSequence(t, want)
	if !reflect.DeepEqual(gotSequence, wantSequence) {
		t.Fatalf("record sequence changed:\n got: %#v\nwant: %#v", gotSequence, wantSequence)
	}
}

func reorderedFixture(fixture inferenceFixture) semantic.InferencePathV1 {
	return semantic.InferencePathV1{
		Version: semantic.InferencePathSchemaVersion,
		Edges:   []semantic.InferenceEdge{fixture.edges[2], fixture.edges[0], fixture.edges[3], fixture.edges[1]},
		Evidence: []semantic.InferenceEvidence{
			fixture.evidence[2], fixture.evidence[0], fixture.evidence[3], fixture.evidence[1],
		},
	}
}

func clonePath(path semantic.InferencePathV1) semantic.InferencePathV1 {
	clone := semantic.InferencePathV1{Version: path.Version}
	clone.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	clone.Claims = append([]semantic.SemanticChangeClaim(nil), path.Claims...)
	clone.Evidence = append([]semantic.InferenceEvidence(nil), path.Evidence...)
	for index := range clone.Edges {
		clone.Edges[index].SourceRoots = append([]semantic.ID(nil), path.Edges[index].SourceRoots...)
		clone.Edges[index].Evidence = append([]semantic.EvidenceReference(nil), path.Edges[index].Evidence...)
	}
	return clone
}
