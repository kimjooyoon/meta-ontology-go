package generation

import "testing"

const ledgerZeroDigest = "0000000000000000000000000000000000000000000000000000000000000000"

func TestExecutionIndicatorLedgerProvenanceIsDecisionBound(t *testing.T) {
	manifest := ExecutionManifest{
		Decision:                      ExecutionDecisionFixedPoint,
		IndicatorDecisionLedgerDigest: "sha256:" + ledgerZeroDigest,
		IndicatorDecisionLedgerCount:  3,
	}
	if err := validateExecutionIndicatorLedgerProvenance(manifest); err != nil {
		t.Fatalf("validateExecutionIndicatorLedgerProvenance() error = %v", err)
	}
	manifest.Decision = ExecutionDecisionRejected
	if err := validateExecutionIndicatorLedgerProvenance(manifest); err == nil {
		t.Fatal("rejected manifest accepted executable ledger provenance")
	}
	manifest.Decision = ExecutionDecisionFixedPoint
	manifest.IndicatorDecisionLedgerDigest = ""
	if err := validateExecutionIndicatorLedgerProvenance(manifest); err == nil {
		t.Fatal("executable manifest accepted missing ledger provenance")
	}
}
