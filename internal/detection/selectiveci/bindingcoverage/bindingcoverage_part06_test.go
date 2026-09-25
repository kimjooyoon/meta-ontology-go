package bindingcoverage

import (
	"testing"
)

func TestCanonicalFixtureDigest(t *testing.T) {
	got := Observe(fixtureInput())
	t.Logf("binding coverage fixture digest=%s counts=%d/%d/%d/%d work=%d input_bytes=%d", got.CanonicalDigest, got.RequiredBindingCount, got.MatchCoveredCount, got.MismatchCoveredCount, got.PartitionCount, got.DeterministicWorkUnits, got.InputBytes)
	if got.CanonicalDigest != got.StableDigest() {
		t.Fatal("fixture digest changed after canonicalization")
	}
}
func assertMissing(t *testing.T, got []string, want string) {
	t.Helper()
	if len(got) != 1 || got[0] != want {
		t.Fatalf("missing IDs = %v, want [%s]", got, want)
	}
}
