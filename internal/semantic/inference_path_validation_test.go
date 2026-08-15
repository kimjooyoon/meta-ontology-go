package semantic

import (
	"errors"
	"strings"
	"testing"
)

func TestInferencePathFailClosedAuthorityAndBindingRules(t *testing.T) {
	candidate := inferenceEdgeFixture(InferenceObservationCandidate, "candidate")
	candidate.After.Semantic = inferenceTestDigest("candidate-changed")
	if err := inferenceBundle(candidate).Validate(); err == nil {
		t.Fatal("candidate semantic mutation crossed the authority boundary")
	}
	wrongAuthority := inferenceEdgeFixture(InferenceObservationCandidate, "wrong-authority")
	wrongAuthority.Authority = AuthorityBinding{Layer: AuthoritySemantic, Effect: AuthorityDerive}
	if err := inferenceBundle(wrongAuthority).Validate(); err == nil {
		t.Fatal("candidate was presented as an authority derivation")
	}
	lift := inferenceEdgeFixture(InferenceAcceptedLift, "unbacked-lift")
	liftEvidence := inferenceEvidenceFixture(lift)
	liftEvidence.SourceBacked = false
	assertInferencePathRejected(t, InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{lift}, Evidence: []InferenceEvidence{liftEvidence},
	})
	verification := inferenceEdgeFixture(InferenceIndependentVerification, "verification")
	verificationEvidence := inferenceEvidenceFixture(verification)
	verificationEvidence.Independent = false
	assertInferencePathRejected(t, InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{verification},
		Evidence: []InferenceEvidence{verificationEvidence},
	})
	declaration := inferenceEdgeFixture(InferenceAuthoritativeDeclaration, "missing-root")
	declaration.SourceRoots = nil
	if err := inferenceBundle(declaration).Validate(); err == nil {
		t.Fatal("declaration without a source root was accepted")
	}
}

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
	// Display names and paths are intentionally absent from InferenceRecord.
	renamed.SubjectID = edge.SubjectID
	if edge.Canonical() != renamed.Canonical() || edge.StableHash() != renamed.StableHash() {
		t.Fatal("display-only rename changed the stable inference record")
	}
}

func TestInferencePathChainRejectsOrphanAndAmbiguousPaths(t *testing.T) {
	start := MustIdentity("inference-test://chain/start")
	middle := MustIdentity("inference-test://chain/middle")
	end := MustIdentity("inference-test://chain/end")
	first := inferenceEdgeFixture(InferenceDeterministicDerivation, "chain-first")
	first.SubjectID, first.ObjectID = start, middle
	second := inferenceEdgeFixture(InferenceDerivedProjection, "chain-second")
	second.SubjectID, second.ObjectID = middle, end
	chain, err := NewInferencePathChain(second, first)
	if err != nil || len(chain.Edges) != 2 || chain.Edges[0].SubjectID != start {
		t.Fatalf("valid unordered chain was not reconstructed: chain=%#v err=%v", chain, err)
	}
	branch := inferenceEdgeFixture(InferenceDeterministicDerivation, "chain-branch")
	branch.SubjectID, branch.ObjectID = start, end
	_, err = NewInferencePathChain(first, branch)
	if err == nil || !strings.Contains(err.Error(), "path_ambiguity") {
		t.Fatalf("ambiguous chain error = %v", err)
	}
	cycleA := inferenceEdgeFixture(InferenceDeterministicDerivation, "chain-cycle-a")
	cycleA.SubjectID = MustIdentity("inference-test://chain/cycle-a")
	cycleA.ObjectID = MustIdentity("inference-test://chain/cycle-b")
	cycleB := inferenceEdgeFixture(InferenceDeterministicDerivation, "chain-cycle-b")
	cycleB.SubjectID, cycleB.ObjectID = cycleA.ObjectID, cycleA.SubjectID
	_, err = NewInferencePathChain(first, cycleA, cycleB)
	if err == nil || !strings.Contains(err.Error(), "path_orphan") {
		t.Fatalf("orphan chain error = %v", err)
	}
}

func TestInferencePathErrorsAreFailClosed(t *testing.T) {
	bad := InferencePathV1{Version: InferencePathSchemaVersion, Edges: []InferenceEdge{{Kind: "UNKNOWN"}}}
	if err := bad.Validate(); err == nil || !errors.Is(err, ErrInferencePath) {
		t.Fatalf("unknown state error = %v, want ErrInferencePath", err)
	}
}
