package counterexamplefirst

import "testing"

func TestCanonicalContractPinsTheExperimentDenominator(t *testing.T) {
	contract := CanonicalContract()
	if !ValidContract(contract) || contract.Fixed.Version != DenominatorVersion ||
		contract.Fixed.Cases != CaseCount || contract.Fixed.Indicators != IndicatorCount ||
		contract.Fixed.ClaimTransitions != TransitionCount || contract.Fixed.UnknownCoordinates != 1 {
		t.Fatalf("contract=%#v", contract)
	}
}
