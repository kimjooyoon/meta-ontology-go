package transformationeffect

import (
	"encoding/json"
	"testing"
)

func TestProjectSplitGoReportRejectsIncompleteIndicatorSet(t *testing.T) {
	last := len(splitGoTestIndicatorIDs) - 1
	report := splitGoTestReport("PASS", "PASS", splitGoTestIndicatorIDs[:last])
	receipts, resolution, reasons, err := projectSplitGoReport(report, splitGoTestIndicatorIDs)
	if err != nil {
		t.Fatal(err)
	}
	if resolution != "LOWER_RESOLUTION" || len(reasons) == 0 {
		t.Fatalf("resolution=%q reasons=%v, want explained lower resolution", resolution, reasons)
	}
	for index, receipt := range receipts {
		if got := splitGoReceiptField(receipt, "Verdict"); got != "UNKNOWN" {
			t.Fatalf("receipt[%d] verdict=%q, want UNKNOWN", index, got)
		}
	}
}

func TestProjectSplitGoReportPreservesExplicitFailure(t *testing.T) {
	report := splitGoTestReport("FAIL", "PASS", splitGoTestIndicatorIDs)
	var decoded map[string]any
	if err := json.Unmarshal(report, &decoded); err != nil {
		t.Fatal(err)
	}
	indicators := decoded["indicators"].([]any)
	indicators[2].(map[string]any)["verdict"] = "FAIL"
	report, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	receipts, resolution, reasons, err := projectSplitGoReport(report, splitGoTestIndicatorIDs)
	if err != nil {
		t.Fatal(err)
	}
	if resolution != "EXACT" || len(reasons) != 0 {
		t.Fatalf("resolution=%q reasons=%v, want exact failure", resolution, reasons)
	}
	if got := splitGoReceiptField(receipts[2], "Verdict"); got != "FAIL" {
		t.Fatalf("failed receipt verdict=%q, want FAIL", got)
	}
}
