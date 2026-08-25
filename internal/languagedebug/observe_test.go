package languagedebug

import (
	"encoding/json"
	"testing"
)

func TestObservePausesAtDeterministicEvent(t *testing.T) {
	receipt := Observe(executionJSON(t), "ACTIVITY_INVOKED")
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionPass || len(receipt.Trace) != 3 || receipt.RemainingEvents != 1 {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func TestObserveRejectsMissingBreakpoint(t *testing.T) {
	receipt := Observe(executionJSON(t), "MISSING")
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionFailClosed || receipt.Resolution != ResolutionExact {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func executionJSON(t *testing.T) []byte {
	t.Helper()
	execution := executionReceipt{
		Schema: "gooo/source-execution-receipt/v1", Decision: "PASS", Resolution: "EXACT",
		Filename: "main.gooo", SourceDigest: testDigest('a'), SemanticDigest: testDigest('b'),
		Entry: json.RawMessage(`{"activity":"PayOrder"}`), Digest: testDigest('c'),
		Events: []Event{{1, "SOURCE_PARSED", "a"}, {2, "SEMANTIC_LOWERED", "b"},
			{3, "ACTIVITY_INVOKED", "PayOrder"}, {4, "ENTITY_PRODUCED", "Receipt"}},
		Diagnostics: []json.RawMessage{}, Effects: Effects{},
	}
	data, err := json.Marshal(execution)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func testDigest(value byte) string {
	data := make([]byte, 64)
	for index := range data {
		data[index] = value
	}
	return "sha256:" + string(data)
}
