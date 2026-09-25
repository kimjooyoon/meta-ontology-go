package semantic

import (
	"testing"
)

func TestInferencePathRejectsMissingDuplicateAndStaleBindings(t *testing.T) {
	derived := inferenceEdgeFixture(InferenceDeterministicDerivation, "derived")
	for _, edit := range []func(*InferenceEdge){
		func(e *InferenceEdge) { e.Rule.ID = "" },
		func(e *InferenceEdge) { e.Rule.Version = "" },
		func(e *InferenceEdge) { e.Rule.Digest = "" },
	} {
		bad := derived
		edit(&bad)
		if err := inferenceBundle(bad).Validate(); err == nil {
			t.Fatal("derived edge with incomplete rule binding was accepted")
		}
	}
	edge := inferenceEdgeFixture(InferenceDeterministicDerivation, "stale")
	evidence := inferenceEvidenceFixture(edge)
	evidence.After.Semantic = inferenceTestDigest("stale-snapshot")
	assertInferencePathRejected(t, InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{edge}, Evidence: []InferenceEvidence{evidence},
	})
	projection := inferenceEdgeFixture(InferenceDerivedProjection, "profile")
	profileEvidence := inferenceEvidenceFixture(projection)
	profileEvidence.Controls.Profile.Digest = inferenceTestDigest("wrong-profile")
	assertInferencePathRejected(t, InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{projection},
		Evidence: []InferenceEvidence{profileEvidence},
	})
	duplicate := inferenceEdgeFixture(InferenceDeterministicDerivation, "duplicate")
	duplicateEvidence := inferenceEvidenceFixture(duplicate)
	duplicateEdge := duplicate
	duplicateEdge.Kind = InferenceDerivedProjection
	duplicateEdge.Phase = PhasePlacement{Phase: PhaseProjection, Ordinal: 1}
	duplicateEdge.Authority = AuthorityBinding{Layer: AuthorityDerived, Effect: AuthorityProject}
	duplicateEdge.Controls.Profile = ProfileBinding{ID: "profile", Version: "1", Digest: inferenceTestDigest("profile")}
	duplicateEdge.Evidence = []EvidenceReference{{ID: duplicateEvidence.ID, Digest: duplicateEvidence.Digest}}
	assertInferencePathRejected(t, InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{duplicate, duplicateEdge},
		Evidence: []InferenceEvidence{duplicateEvidence, duplicateEvidence},
	})
}
func TestInferencePathStableIdentityIgnoresDisplayRenames(t *testing.T) {
	edge := inferenceEdgeFixture(InferenceDeterministicDerivation, "rename")
	renamed := edge

	renamed.SubjectID = edge.SubjectID
	if edge.Canonical() != renamed.Canonical() || edge.StableHash() != renamed.StableHash() {
		t.Fatal("display-only rename changed the stable inference record")
	}
}
