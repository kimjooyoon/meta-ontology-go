package operationconformance

import (
	"os"
	"testing"
)

func TestContractOracleDenominator(t *testing.T) {
	raw, err := os.ReadFile("../../../examples/source-splitter-conformance/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := EvaluateContract(raw)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Matched != 18 || receipt.Total != 18 || receipt.PassCases != 6 ||
		receipt.FailCases != 6 || receipt.UnknownCases != 6 {
		t.Fatalf("receipt=%+v", receipt)
	}
}

func TestUnknownEvidenceCannotPass(t *testing.T) {
	raw, err := os.ReadFile("../../../examples/source-splitter-conformance/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	report := Evaluate(raw, SplitGoEvidence{ExpectedHeadSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", OperationID: OperationID})
	if report.Decision != DecisionBlock || report.Resolution != "LOWER_RESOLUTION" ||
		report.Summary.UnknownCount != 6 || report.Summary.PassCount != 0 {
		t.Fatalf("report=%+v", report)
	}
	if err := Validate(report, raw); err != nil {
		t.Fatal(err)
	}
}
