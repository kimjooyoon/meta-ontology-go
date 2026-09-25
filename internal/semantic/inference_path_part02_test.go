package semantic

import (
	"strings"
	"testing"
)

func TestInferencePathKindsAreClosedAndEvidenceBound(t *testing.T) {
	kinds := []InferenceKind{
		InferenceAuthoritativeDeclaration, InferenceDeterministicDerivation, InferenceDerivedProjection,
		InferenceObservationCandidate, InferenceAcceptedLift, InferenceIndependentVerification,
	}
	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			edge := inferenceEdgeFixture(kind, strings.ToLower(string(kind)))
			if err := inferenceBundle(edge).Validate(); err != nil {
				t.Fatalf("valid edge rejected: %v", err)
			}
		})
	}
	if InferenceKind(NoSemanticDelta).Valid() {
		t.Fatal("semantic-change claim kind crossed into the inference-kind sum")
	}
	if InferenceKind("UNKNOWN").Valid() {
		t.Fatal("unknown inference kind was accepted")
	}
}
