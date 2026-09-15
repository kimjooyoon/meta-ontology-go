package coupling

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestNoWriteDoesNotMutateInput(t *testing.T) {
	input := testCorpus()[3].Input
	before, err := CanonicalInputBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	_ = Evaluate(input)
	after, err := CanonicalInputBytes(input)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("oracle wrote to its input")
	}
}
func reverseBindings(values []CodeBinding) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseChanges(values []CodeChange) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseReceipts(values []CouplingReceipt) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseResources(values []ExternalResourceReceipt) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseEdges(values []semantic.InferenceEdge) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseClaims(values []semantic.SemanticChangeClaim) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseEvidence(values []semantic.InferenceEvidence) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
