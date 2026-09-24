package provenance

import "testing"

func completeOutcomeReceipt() SelfImprovementOutcomeReceipt {
	receipt := NewSelfImprovementOutcomeReceipt()
	receipt.ScenarioID = "outcome-receipt-contract"
	receipt.TrialIndex = 0
	receipt.SourceDigest = "sha256:source"
	receipt.FixtureDigest = "sha256:fixture"
	receipt.ToolchainDigest = "sha256:toolchain"
	receipt.EvaluatorDigest = "sha256:evaluator"
	receipt.CounterexampleRecovered = SelfImprovementOutcomeClosed
	receipt.ContractPreserved = SelfImprovementOutcomeClosed
	receipt.RegressionEvidence = SelfImprovementOutcomeClosed
	receipt.AdoptionAuthorized = SelfImprovementOutcomeClosed
	receipt.AdoptionAuthorityEvidenceDigest = "sha256:independent-authority"
	receipt.AdoptionAuthorityIndependent = true
	return receipt
}

func TestSelfImprovementOutcomeReceiptRejectsInferredAdoption(t *testing.T) {
	receipt := completeOutcomeReceipt()
	receipt.AdoptionAuthorityIndependent = false
	if err := receipt.Validate(); err == nil {
		t.Fatal("expected adoption authority independence to be required")
	}
}

func TestSelfImprovementOutcomeReceiptPreservesFirstUnknownStage(t *testing.T) {
	receipt := completeOutcomeReceipt()
	receipt.RegressionEvidence = SelfImprovementOutcomeUnknown
	if err := receipt.PreserveFirstUnknown(3, "regression_evidence", "replay receipt missing"); err != nil {
		t.Fatalf("preserve first unknown: %v", err)
	}
	if err := receipt.PreserveFirstUnknown(5, "adoption_authorized", "authority receipt missing"); err != nil {
		t.Fatalf("preserve later unknown: %v", err)
	}
	if receipt.MissingStageIndex != 3 || receipt.MissingStage != "regression_evidence" {
		t.Fatalf("first unknown stage was not preserved: %#v", receipt)
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("validate unknown receipt: %v", err)
	}
}

func TestSelfImprovementOutcomeReceiptRequiresUnknownCoordinates(t *testing.T) {
	receipt := completeOutcomeReceipt()
	receipt.CounterexampleRecovered = SelfImprovementOutcomeUnknown
	if err := receipt.Validate(); err == nil {
		t.Fatal("expected UNKNOWN coordinates to be required")
	}
}
