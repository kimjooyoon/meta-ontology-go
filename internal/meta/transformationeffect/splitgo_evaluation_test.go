package transformationeffect

import (
	"reflect"
	"testing"
)

func TestProjectSplitGoReportPassesExactFixedDenominator(t *testing.T) {
	report := splitGoTestReport("PASS", "PASS", splitGoTestIndicatorIDs)
	receipts, resolution, reasons, err := projectSplitGoReport(report, splitGoTestIndicatorIDs)
	if err != nil {
		t.Fatal(err)
	}
	if resolution != "EXACT" || len(reasons) != 0 {
		t.Fatalf("resolution=%q reasons=%v, want EXACT with no reasons", resolution, reasons)
	}
	if len(receipts) != len(splitGoTestIndicatorIDs) {
		t.Fatalf("receipts=%d, want %d", len(receipts), len(splitGoTestIndicatorIDs))
	}
	for index, receipt := range receipts {
		if got := splitGoReceiptField(receipt, "IndicatorID", "ID"); got != splitGoTestIndicatorIDs[index] {
			t.Fatalf("receipt[%d] id=%q, want %q", index, got, splitGoTestIndicatorIDs[index])
		}
		if got := splitGoReceiptField(receipt, "Verdict"); got != "PASS" {
			t.Fatalf("receipt[%d] verdict=%q, want PASS", index, got)
		}
	}
}

func TestProjectSplitGoReportLowersUnknownTopDecision(t *testing.T) {
	report := splitGoTestReport("UNRECOGNIZED", "PASS", splitGoTestIndicatorIDs)
	receipts, resolution, reasons, err := projectSplitGoReport(report, splitGoTestIndicatorIDs)
	if err != nil {
		t.Fatal(err)
	}
	if resolution != "LOWER_RESOLUTION" {
		t.Fatalf("resolution=%q, want LOWER_RESOLUTION", resolution)
	}
	if !reflect.DeepEqual(reasons, []string{"TOP_LEVEL_DECISION_UNKNOWN"}) {
		t.Fatalf("reasons=%v, want TOP_LEVEL_DECISION_UNKNOWN", reasons)
	}
	for index, receipt := range receipts {
		if got := splitGoReceiptField(receipt, "Verdict"); got != "UNKNOWN" {
			t.Fatalf("receipt[%d] verdict=%q, want UNKNOWN", index, got)
		}
	}
}
