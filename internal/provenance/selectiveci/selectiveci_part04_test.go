package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"testing"
	"testing/quick"
)

func TestPermutationsHaveIdenticalCanonicalReceipt(t *testing.T) {
	left := Evaluate(completeFixture().input)
	rightFixture := cloneFixture(completeFixture())
	reverseEdges := append([]semantic.InferenceEdge(nil), rightFixture.input.InferencePath.Edges...)
	for leftIndex, rightIndex := 0, len(reverseEdges)-1; leftIndex < rightIndex; leftIndex, rightIndex = leftIndex+1, rightIndex-1 {
		reverseEdges[leftIndex], reverseEdges[rightIndex] = reverseEdges[rightIndex], reverseEdges[leftIndex]
	}
	rightFixture.input.InferencePath.Edges = reverseEdges
	reverseEvidence := append([]semantic.InferenceEvidence(nil), rightFixture.input.InferencePath.Evidence...)
	for leftIndex, rightIndex := 0, len(reverseEvidence)-1; leftIndex < rightIndex; leftIndex, rightIndex = leftIndex+1, rightIndex-1 {
		reverseEvidence[leftIndex], reverseEvidence[rightIndex] = reverseEvidence[rightIndex], reverseEvidence[leftIndex]
	}
	rightFixture.input.InferencePath.Evidence = reverseEvidence
	right := Evaluate(rightFixture.input)
	if left.Canonical() != right.Canonical() || left.Digest != right.Digest {
		t.Fatalf("permutation changed receipt:\nleft=%s\nright=%s", left.Canonical(), right.Canonical())
	}
}
func TestPermutationProperty(t *testing.T) {
	property := func(keys []uint8) bool {
		fixture := cloneFixture(completeFixture())
		order := []uint8{0, 0, 0}
		copy(order, keys)
		sort.SliceStable(fixture.input.InferencePath.Edges, func(left, right int) bool {
			return order[left] < order[right]
		})
		sort.SliceStable(fixture.input.InferencePath.Evidence, func(left, right int) bool {
			return order[left] < order[right]
		})
		return Evaluate(fixture.input).Canonical() == Evaluate(completeFixture().input).Canonical()
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 64}); err != nil {
		t.Fatal(err)
	}
}
