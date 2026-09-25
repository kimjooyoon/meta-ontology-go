package generation

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestArtifactProvenanceBindsAllStagesAndReplays(t *testing.T) {
	plan := actionableReceiptPlan()
	execution := BuildExecutionManifest(plan)
	receipts := VerifyReceipts(plan, passingReceipts(plan))
	envelope := BindArtifactProvenance(plan, execution, receipts)
	if envelope.Decision != ArtifactProvenanceDecisionBound ||
		envelope.Summary != (ArtifactProvenanceSummary{Pass: 4}) ||
		envelope.PromotionAuthorized {
		t.Fatalf("unexpected artifact provenance: %+v", envelope)
	}
	if envelope.IndicatorDecisionLedgerDigest != plan.IndicatorDecisionLedger.Digest ||
		envelope.IndicatorDecisionLedgerCount != len(plan.IndicatorDecisionLedger.Entries) {
		t.Fatal("artifact provenance lost indicator ledger identity")
	}
	payload, err := EncodeArtifactProvenance(envelope)
	if err != nil {
		t.Fatal(err)
	}
	var replay ArtifactProvenance
	if err := json.Unmarshal(payload, &replay); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(envelope, replay) {
		t.Fatal("artifact provenance replay changed evidence")
	}
}
