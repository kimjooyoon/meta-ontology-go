package semantic

import (
	"reflect"
	"strings"
	"testing"
)

func TestInferencePathCanonicalReplayAndNoMutation(t *testing.T) {
	first := inferenceEdgeFixture(InferenceDeterministicDerivation, "canonical-first")
	second := inferenceEdgeFixture(InferenceDerivedProjection, "canonical-second")
	left := InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{first, second},
		Evidence: []InferenceEvidence{inferenceEvidenceFixture(first), inferenceEvidenceFixture(second)},
	}
	right := InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{second, first},
		Evidence: []InferenceEvidence{inferenceEvidenceFixture(second), inferenceEvidenceFixture(first)},
	}
	leftBefore, rightBefore := left, right
	if left.Canonical() != right.Canonical() || left.StableHash() != right.StableHash() {
		t.Fatal("canonical replay changed with insertion order")
	}
	if !reflect.DeepEqual(left, leftBefore) || !reflect.DeepEqual(right, rightBefore) {
		t.Fatal("canonicalization mutated the input record sets")
	}
	empty := InferencePathV1{Version: InferencePathSchemaVersion}
	canonical := empty.Canonical()
	for _, marker := range []string{"edges\t0", "claims\t0", "evidence-records\t0"} {
		if !strings.Contains(canonical, marker) {
			t.Fatalf("canonical empty set omitted explicit marker %q: %s", marker, canonical)
		}
	}
	if left.Canonical() != left.Canonical() {
		t.Fatal("two clean canonical runs diverged")
	}
}
