package denominatorevolution

import "testing"

func TestCanonicalContractProducesExactReadOnlyReport(t *testing.T) {
	report := Evaluate(Input{Contract: CanonicalContract(), HeadSHA: "0123456789012345678901234567890123456789", Source: []byte(canonicalSource)})
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.FixedDenominatorNumerator != 5 || report.Summary.FixedDenominatorDenominator != 5 {
		t.Fatalf("fixed denominator = %d/%d, want 5/5", report.Summary.FixedDenominatorNumerator, report.Summary.FixedDenominatorDenominator)
	}
	if report.Summary.LegalAdvanceNumerator != 1 || report.Summary.UnauthorizedRejectionNumerator != 1 || report.Summary.UnknownPredecessorNumerator != 1 {
		t.Fatalf("case summary = %+v", report.Summary)
	}
	if !guardrailsConform(report.Summary.Guardrails) || report.RepositoryWrites != 0 || report.MutationAuthority {
		t.Fatalf("unsafe summary = %+v", report.Summary)
	}
	if report.Cases[0].Receipt == nil || !guardrailsConform(report.Cases[0].Receipt.Guardrails) {
		t.Fatalf("receipt guardrails = %+v", report.Cases[0].Receipt)
	}
}

func TestContractDriftFailsClosed(t *testing.T) {
	contract := CanonicalContract()
	contract.Policy.NoAggregateEstimates = false
	report := Evaluate(Input{Contract: contract, HeadSHA: "0123456789012345678901234567890123456789", Source: []byte(canonicalSource)})
	if report.Decision != "FAIL_CLOSED" || report.Reason != "DENOMINATOR_EVOLUTION_CONTRACT_DRIFT" {
		t.Fatalf("report = %+v", report)
	}
}

const canonicalSource = `package denominatorevolution
namespace denominatorevolution
entity FixedDenominator id "gooo://denominator-evolution/entity/fixed-denominator"
entity DenominatorVersion id "gooo://denominator-evolution/entity/denominator-version"
entity ChangeReason id "gooo://denominator-evolution/entity/change-reason"
entity PredecessorEvidence id "gooo://denominator-evolution/entity/predecessor-evidence"
entity MigrationReceipt id "gooo://denominator-evolution/entity/migration-receipt"
entity ClaimTransition id "gooo://denominator-evolution/entity/claim-transition"
entity IndependentDecision id "gooo://denominator-evolution/entity/independent-decision"
activity DeclareFixedDenominator(FixedDenominator) -> DenominatorVersion
activity ProposeDenominatorChange(DenominatorVersion) -> ChangeReason
activity BindPredecessorDigest(DenominatorVersion) -> PredecessorEvidence
activity RecordChangeReasons(ChangeReason) -> MigrationReceipt
activity IssueMigrationReceipt(PredecessorEvidence) -> MigrationReceipt
activity TransitionClaim(MigrationReceipt) -> ClaimTransition
activity IndependentlyDecide(ClaimTransition) -> IndependentDecision
`
