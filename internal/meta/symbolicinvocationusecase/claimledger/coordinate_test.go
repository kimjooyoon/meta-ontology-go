package claimledger

import "testing"

func TestClaimRequiresAStageAndStep(t *testing.T) {
	claim := inScope("artifact", "artifact", "NON_NULL", "FOUNDATION", "EMIT", "", nil)
	contract := testContract([]ClaimSpec{claim}, ExpectedMetrics{FixedClaimTotal: 1})
	if _, err := Project(contract, []byte(`{}`), "abc"); err == nil {
		t.Fatal("expected an imprecise process coordinate to fail")
	}
}
