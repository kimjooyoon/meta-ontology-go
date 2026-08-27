package semanticresolution

import "testing"

func TestPartialObservationDescendsDirectlyToInvariantOnly(t *testing.T) {
	got := ResolvePartialObservation(PartialObservation{Required: 3, Observed: 2, Reason: "REQUIRED_INPUT_UNOBSERVED"})
	if got.Decision != DecisionLowerResolution || got.ToResolution != ResolutionInvariantOnly || got.Unknown == nil {
		t.Fatalf("transition = %+v", got)
	}
	if got.Unknown.Stage != StagePartialObservation || got.Unknown.Step != 1 || got.Unknown.Reason != "REQUIRED_INPUT_UNOBSERVED" {
		t.Fatalf("unknown = %+v", got.Unknown)
	}
}

func TestCanonicalLatticeReceiptIsClosedAndReplayable(t *testing.T) {
	receipt := BuildLatticeReceipt("examples/semantic-resolution-lattice/main.gooo", "sha256:test")
	if err := ValidateLatticeReceipt(receipt); err != nil {
		t.Fatal(err)
	}
}

func TestMutationAuthorityFailsClosed(t *testing.T) {
	got := ResolvePartialObservation(PartialObservation{Required: 3, Observed: 2, Reason: "REQUIRED_INPUT_UNOBSERVED", MutationAuthority: true})
	if got.Decision != DecisionFailClosed || got.Reason != "MUTATION_AUTHORITY_PRESENT" {
		t.Fatalf("transition = %+v", got)
	}
}
