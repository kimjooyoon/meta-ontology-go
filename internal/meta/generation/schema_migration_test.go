package generation

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaMigrationRejectsPreviousReceiptBoundaries(t *testing.T) {
	plan := actionableReceiptPlan()
	receipts := passingReceipts(plan)
	report := VerifyReceipts(plan, receipts)
	bundle, _, _ := validationFailureBundle([]string{
		"go", "run", "./scripts/source-splitter", "-root", "<workspace>",
		"-subject", "fixture.go:1:Selected",
	})
	ledger := *plan.IndicatorDecisionLedger

	type schemaCase struct {
		name    string
		payload []byte
		current string
		old     string
		decode  func([]byte) error
	}
	marshal := func(value interface{}) []byte {
		t.Helper()
		payload, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal migration fixture: %v", err)
		}
		return payload
	}
	cases := []schemaCase{
		{
			name:    "operation receipt",
			payload: marshal(receipts[0]),
			current: OperationReceiptSchemaVersion,
			old:     "gooo/meta-operation-receipt/v2",
			decode: func(data []byte) error {
				var value OperationReceipt
				return json.Unmarshal(data, &value)
			},
		},
		{
			name:    "receipt report",
			payload: marshal(report),
			current: ReceiptReportSchemaVersion,
			old:     "gooo/meta-operation-receipt-report/v2",
			decode: func(data []byte) error {
				var value ReceiptReport
				return json.Unmarshal(data, &value)
			},
		},
		{
			name:    "observation bundle",
			payload: marshal(bundle),
			current: OperationObservationBundleSchema,
			old:     "gooo/meta-operation-observation-bundle/v1",
			decode: func(data []byte) error {
				var value OperationObservationBundle
				return json.Unmarshal(data, &value)
			},
		},
		{
			name:    "indicator ledger",
			payload: marshal(ledger),
			current: IndicatorDecisionLedgerSchemaVersion,
			old:     "gooo/indicator-decision-ledger/v2",
			decode: func(data []byte) error {
				var value IndicatorDecisionLedger
				return json.Unmarshal(data, &value)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			legacyPayload := []byte(strings.Replace(string(tc.payload), tc.current, tc.old, 1))
			err := tc.decode(legacyPayload)
			if err == nil {
				t.Fatal("previous schema version was accepted")
			}
			if !strings.Contains(err.Error(), "unsupported") {
				t.Fatalf("previous schema version failed without explicit rejection: %v", err)
			}
		})
	}
}

func TestSchemaMigrationDecodesAndVerifiesCurrentReceiptBoundaries(t *testing.T) {
	plan := actionableReceiptPlan()
	receipts := passingReceipts(plan)
	report := VerifyReceipts(plan, receipts)
	if report.Decision != ReceiptDecisionConformant {
		t.Fatalf("current receipt report did not verify: %+v", report)
	}

	receiptPayload, err := json.Marshal(receipts[0])
	if err != nil {
		t.Fatalf("marshal current operation receipt: %v", err)
	}
	var receipt OperationReceipt
	if err := json.Unmarshal(receiptPayload, &receipt); err != nil {
		t.Fatalf("current operation receipt did not decode: %v", err)
	}

	reportPayload, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal current receipt report: %v", err)
	}
	var decodedReport ReceiptReport
	if err := json.Unmarshal(reportPayload, &decodedReport); err != nil {
		t.Fatalf("current receipt report did not decode: %v", err)
	}

	ledgerPayload, err := json.Marshal(*plan.IndicatorDecisionLedger)
	if err != nil {
		t.Fatalf("marshal current indicator ledger: %v", err)
	}
	var decodedLedger IndicatorDecisionLedger
	if err := json.Unmarshal(ledgerPayload, &decodedLedger); err != nil {
		t.Fatalf("current indicator ledger did not decode: %v", err)
	}

	bundle, bundlePlan, manifest := validationFailureBundle([]string{
		"go", "run", "./scripts/source-splitter", "-root", "<workspace>",
		"-subject", "fixture.go:1:Selected",
	})
	if err := ValidateObservationBundle(bundle, bundlePlan, manifest); err != nil {
		t.Fatalf("current observation bundle did not verify: %v", err)
	}
	bundlePayload, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("marshal current observation bundle: %v", err)
	}
	var decodedBundle OperationObservationBundle
	if err := json.Unmarshal(bundlePayload, &decodedBundle); err != nil {
		t.Fatalf("current observation bundle did not decode: %v", err)
	}
	if err := ValidateObservationBundle(decodedBundle, bundlePlan, manifest); err != nil {
		t.Fatalf("decoded current observation bundle did not verify: %v", err)
	}
}
