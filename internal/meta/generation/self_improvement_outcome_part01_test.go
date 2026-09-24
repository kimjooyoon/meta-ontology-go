package generation

import "testing"

func TestValidateSelfImprovementOutcomeDoesNotAuthorizeFromProvenance(t *testing.T) {
	receipt := SelfImprovementOutcomeReceipt{
		SchemaVersion:                   SelfImprovementOutcomeSchema,
		ScenarioID:                      "generation-outcome",
		TrialIndex:                     0,
		CounterexampleRecovered:         SelfImprovementOutcomeClosed,
		ContractPreserved:               SelfImprovementOutcomeClosed,
		RegressionEvidence:              SelfImprovementOutcomeClosed,
		AdoptionAuthorized:              SelfImprovementOutcomeClosed,
		AdoptionAuthorityEvidenceDigest: "sha256:authority",
	}
	if err := ValidateSelfImprovementOutcome(receipt); err == nil {
		t.Fatal("expected independent authority evidence marker")
	}
}
