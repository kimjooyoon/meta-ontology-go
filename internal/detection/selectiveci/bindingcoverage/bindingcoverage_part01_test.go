package bindingcoverage

import (
	"testing"
)

func TestExactFixture(t *testing.T) {
	input := fixtureInput()
	got := Observe(input)
	if got.Decision != DecisionExact || got.Reason != ReasonComplete {
		t.Fatalf("got %s/%s, want EXACT/COMPLETE", got.Decision, got.Reason)
	}
	if got.RequiredBindingCount != 9 || got.MatchCoveredCount != 9 || got.MismatchCoveredCount != 9 || got.PartitionCount != 18 {
		t.Fatalf("coverage counts = %d/%d/%d/%d, want 9/9/9/18", got.RequiredBindingCount, got.MatchCoveredCount, got.MismatchCoveredCount, got.PartitionCount)
	}
	if got.EndpointReferenceCount != 18 || got.DeterministicWorkUnits != 45 || got.InputBytes == 0 {
		t.Fatalf("work/input bytes = %d/%d, want work 45 and nonzero input", got.DeterministicWorkUnits, got.InputBytes)
	}
	if got.InputDigest == "" {
		t.Fatal("canonical input digest is empty")
	}
	if len(got.MissingMatchBindingIDs) != 0 || len(got.MissingMismatchBindingIDs) != 0 {
		t.Fatal("exact fixture reported missing coverage")
	}
	if got.CanonicalDigest != got.StableDigest() {
		t.Fatalf("digest = %q, stable digest = %q", got.CanonicalDigest, got.StableDigest())
	}
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
