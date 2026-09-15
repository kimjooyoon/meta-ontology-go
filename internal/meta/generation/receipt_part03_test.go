package generation

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestReceiptReportCarriesLedgerProvenanceAndReplaysStrictly(t *testing.T) {
	report := sourcepolicy.Report{
		Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{
			metric("floor", sourcepolicy.OperationSplitGo, true, true),
		},
	}
	plan := Build(strings.Repeat("3", 40), strings.Repeat("4", 40), report)
	verified := VerifyReceipts(plan, nil)
	if verified.IndicatorDecisionLedgerDigest != plan.IndicatorDecisionLedger.Digest ||
		verified.IndicatorDecisionLedgerCount != len(plan.IndicatorDecisionLedger.Entries) {
		t.Fatalf("receipt report lost indicator ledger provenance: %+v", verified)
	}
	encoded, err := json.Marshal(verified)
	if err != nil {
		t.Fatal(err)
	}
	var replay ReceiptReport
	if err := json.Unmarshal(encoded, &replay); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(verified, replay) {
		t.Fatal("receipt report strict replay changed evidence")
	}
}

func TestVerifyReceiptsRejectsForgedLedgerProvenance(t *testing.T) {
	plan := actionableReceiptPlan()
	forged := passingReceipts(plan)
	forged[0].IndicatorDecisionLedgerDigest = "sha256:" + strings.Repeat("f", 64)
	forged[0].ReceiptDigest = operationReceiptDigest(forged[0])
	report := VerifyReceipts(plan, forged)
	if report.Decision != ReceiptDecisionUnknown ||
		report.Reason != ReceiptReasonUnknownIndicator {
		t.Fatalf("forged ledger provenance did not fail closed: %+v", report)
	}
}
