package semanticresolution

import "testing"

const latticeTestSource = `package resolutionlattice
namespace meta
activity CaseExactObservation(PartialObservation) -> Claim computes "resolution-lattice.case;id=exact-observation;required=3;observed=3;reason=OBSERVATION_COMPLETE;repository_writes=0;mutation_authority=false;claim_id=claim-exact-observation;claim_state=DISCHARGED"
activity CasePartialInvariantDescent(PartialObservation) -> Claim computes "resolution-lattice.case;id=partial-invariant-descent;required=3;observed=2;reason=REQUIRED_INPUT_UNOBSERVED;repository_writes=0;mutation_authority=false;claim_id=claim-invariant-fallback;claim_state=OPEN"
activity CaseMalformedObservation(PartialObservation) -> Claim computes "resolution-lattice.case;id=malformed-observation;required=3;observed=4;reason=OBSERVATION_CARDINALITY_INVALID;repository_writes=0;mutation_authority=false;claim_id=claim-exact-under-missing-evidence;claim_state=REFUTED"
activity CaseMutationAuthority(PartialObservation) -> Claim computes "resolution-lattice.case;id=mutation-authority;required=3;observed=2;reason=REQUIRED_INPUT_UNOBSERVED;repository_writes=0;mutation_authority=true;claim_id=claim-write-free-descent;claim_state=DISCHARGED"
`

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
	receipt, err := BuildLatticeReceipt("examples/semantic-resolution-lattice/main.gooo", "sha256:test", "package resolutionlattice\nnamespace meta\n")
	if err == nil {
		t.Fatal("incomplete Gooo source unexpectedly produced a receipt")
	}
	receipt, err = BuildLatticeReceipt("examples/semantic-resolution-lattice/main.gooo", "sha256:test", latticeTestSource)
	if err != nil {
		t.Fatal(err)
	}
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
