package semantic

import (
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
