package bindingcoverage

import (
	"bytes"
	"testing"
)

func TestPermutationCanonicalEquality(t *testing.T) {
	first := fixtureInput()
	second := fixtureInput()
	second.RequiredBindings = reverseBindings(second.RequiredBindings)
	second.Partitions = reversePartitions(second.Partitions)
	second.PrecedenceRegistry = reversePrecedence(second.PrecedenceRegistry)
	left, err := EncodeJSON(Observe(first))
	if err != nil {
		t.Fatal(err)
	}
	right, err := EncodeJSON(Observe(second))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("permutations differ:\n%s\n%s", left, right)
	}
}
func TestZeroPrecedenceRankIsValid(t *testing.T) {
	input := fixtureInput()
	input.PrecedenceRegistry[0].Rank = 0
	got := Observe(input)
	if got.Decision != DecisionExact || got.Reason != ReasonComplete {
		t.Fatalf("got %s/%s, want EXACT/COMPLETE", got.Decision, got.Reason)
	}
}
