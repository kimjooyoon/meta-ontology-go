package bindingcoverage

import (
	"bytes"
	"testing"
)

func FuzzPermutationCanonical(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{1, 2, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		first := fixtureInput()
		second := fixtureInput()
		if len(data)%2 == 0 {
			second.RequiredBindings = reverseBindings(second.RequiredBindings)
		}
		if len(data)%3 == 0 {
			second.Partitions = reversePartitions(second.Partitions)
		}
		if len(data)%5 == 0 {
			second.PrecedenceRegistry = reversePrecedence(second.PrecedenceRegistry)
		}
		left, err := EncodeJSON(Observe(first))
		if err != nil {
			t.Fatal(err)
		}
		right, err := EncodeJSON(Observe(second))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(left, right) {
			t.Fatal("permutation changed canonical output")
		}
	})
}

func FuzzNoFalseExact(f *testing.F) {
	seed, err := EncodeInputJSON(fixtureInput())
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte(`{"schema_version":"bad"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		got := ClassifyJSON(data)
		if got.CanonicalDigest == "" || got.CanonicalDigest != got.StableDigest() {
			t.Fatal("result is not digest-bound")
		}
		if got.Decision != DecisionExact {
			return
		}
		input, err := DecodeJSON(data)
		if err != nil || input.SchemaVersion != SchemaVersion || len(input.RequiredBindings) == 0 {
			t.Fatal("EXACT result came from malformed or zero-denominator input")
		}
		match, mismatch := coveredPolarity(input)
		for _, binding := range input.RequiredBindings {
			if !match[binding.BindingID] || !mismatch[binding.BindingID] {
				t.Fatalf("EXACT result lacks both polarities for %s", binding.BindingID)
			}
		}
	})
}

func coveredPolarity(input Input) (map[string]bool, map[string]bool) {
	match := map[string]bool{}
	mismatch := map[string]bool{}
	for _, partition := range input.Partitions {
		if partition.Polarity == PolarityMatch {
			match[partition.BindingID] = true
		} else if partition.Polarity == PolarityMismatch {
			mismatch[partition.BindingID] = true
		}
	}
	return match, mismatch
}
