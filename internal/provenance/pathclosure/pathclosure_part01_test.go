package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestCompleteDeclarationDerivationProjectionVerificationPath(t *testing.T) {
	fixture := completeInferenceFixture()
	if err := fixture.path.Validate(); err != nil {
		t.Fatalf("complete path rejected: %v", err)
	}
	normalized, err := fixture.path.Normalized()
	if err != nil {
		t.Fatalf("complete path normalization failed: %v", err)
	}
	assertExactRecordSequence(t, normalized.Edges, fixture.edges)
	chain, err := semantic.NewInferencePathChain(fixture.edges...)
	if err != nil {
		t.Fatalf("complete chain rejected: %v", err)
	}
	assertExactRecordSequence(t, chain.Edges, fixture.edges)
	if got := []semantic.InferenceKind{
		chain.Edges[0].Kind, chain.Edges[1].Kind, chain.Edges[2].Kind, chain.Edges[3].Kind,
	}; !reflect.DeepEqual(got, []semantic.InferenceKind{
		semantic.InferenceAuthoritativeDeclaration,
		semantic.InferenceDeterministicDerivation,
		semantic.InferenceDerivedProjection,
		semantic.InferenceIndependentVerification,
	}) {
		t.Fatalf("chain kind sequence = %#v", got)
	}
}
func TestInsertionOrderReplayPreservesExactRecordSequence(t *testing.T) {
	fixture := completeInferenceFixture()
	reordered := reorderedFixture(fixture)
	original := clonePath(reordered)
	left, err := fixture.path.Normalized()
	if err != nil {
		t.Fatalf("original normalization failed: %v", err)
	}
	right, err := reordered.Normalized()
	if err != nil {
		t.Fatalf("reordered normalization failed: %v", err)
	}
	assertExactRecordSequence(t, right.Edges, left.Edges)
	if left.Canonical() != right.Canonical() || left.StableHash() != right.StableHash() {
		t.Fatal("insertion order changed the normalized path receipt")
	}
	if !reflect.DeepEqual(reordered, original) {
		t.Fatal("normalization mutated the insertion-ordered fixture")
	}
}
func TestMissingEdgeIsNotAnEmptySuccessfulChain(t *testing.T) {
	fixture := completeInferenceFixture()
	incomplete := append([]semantic.InferenceEdge(nil), fixture.edges[:2]...)
	incomplete = append(incomplete, fixture.edges[3])
	if err := (semantic.InferencePathV1{
		Version: semantic.InferencePathSchemaVersion, Edges: incomplete, Evidence: fixture.evidence,
	}).Validate(); err != nil {
		t.Fatalf("record-valid incomplete path was rejected before topology check: %v", err)
	}
	if _, err := semantic.NewInferencePathChain(incomplete...); err == nil {
		t.Fatal("missing edge produced a successful chain")
	}
}
func TestZeroRequirementsIsNotASuccessfulPath(t *testing.T) {
	if _, err := semantic.NewInferencePathChain(); err == nil {
		t.Fatal("zero requirements produced a successful chain")
	}
}
