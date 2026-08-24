package sourceauthority

import "testing"

func TestContractCannotMintReadiness(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("load contract: %v", err)
	}
	if contract.AdoptionState != "CONTRACT_ONLY" {
		t.Fatalf("adoption state = %q", contract.AdoptionState)
	}
	if contract.ReadinessCredit != 0 {
		t.Fatalf("readiness credit = %d", contract.ReadinessCredit)
	}
}

func TestUnknownEvidenceLowersResolutionAndBlocks(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("load contract: %v", err)
	}
	assertUnknownBlock(t, contract.UnknownEvidence)
	assertUnknownBlock(t, contract.EmptyDenominator)
}

func assertUnknownBlock(t *testing.T, mode FailureMode) {
	t.Helper()
	if mode.Observation != "UNKNOWN" {
		t.Fatalf("observation = %q", mode.Observation)
	}
	if mode.Resolution != "INVARIANT_ONLY" {
		t.Fatalf("resolution = %q", mode.Resolution)
	}
	if mode.Enforcement != "BLOCK" {
		t.Fatalf("enforcement = %q", mode.Enforcement)
	}
}
