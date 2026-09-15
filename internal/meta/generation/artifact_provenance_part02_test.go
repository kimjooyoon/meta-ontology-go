package generation

import (
	"strings"
	"testing"
)

func TestArtifactProvenanceRejectsCanonicalLedgerMismatch(t *testing.T) {
	plan := actionableReceiptPlan()
	execution := BuildExecutionManifest(plan)
	receipts := VerifyReceipts(plan, passingReceipts(plan))
	receipts.IndicatorDecisionLedgerDigest = "sha256:" + strings.Repeat("f", 64)
	receipts = finishReceiptReport(receipts)
	envelope := BindArtifactProvenance(plan, execution, receipts)
	if envelope.Decision != ArtifactProvenanceDecisionRejected ||
		envelope.Reason != ArtifactProvenanceReasonMismatch ||
		envelope.Summary.Fail != 1 {
		t.Fatalf("ledger mismatch did not reject provenance: %+v", envelope)
	}
}
