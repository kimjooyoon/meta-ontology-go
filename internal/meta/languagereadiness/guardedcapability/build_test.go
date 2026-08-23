package guardedcapability

import "testing"

func TestExactCapabilitySeparatesImplementationFromEvent(t *testing.T) {
	receipt := Build(exactSource(t))
	if err := ValidateForHead(receipt, receipt.Source.CurrentHeadSHA); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionPass || receipt.Summary.Satisfied != 8 ||
		receipt.Summary.Total != 8 || receipt.Summary.ReadinessBPS != 10000 {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func TestUnknownAncestryLowersResolution(t *testing.T) {
	source := exactSource(t)
	source.AncestryObserved = false
	receipt := Build(source)
	if receipt.Decision != DecisionFailClosed || receipt.Reason != ReasonUnknown ||
		receipt.Resolution != ResolutionLower || receipt.Summary.Unresolved != 1 {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func TestImplementationDriftRejectsCapability(t *testing.T) {
	source := exactSource(t)
	source.CurrentGuardTree = "tree-drift"
	receipt := Build(source)
	if receipt.Decision != DecisionFailClosed || receipt.Reason != ReasonRejected ||
		receipt.Summary.NotSatisfied != 1 {
		t.Fatalf("receipt = %#v", receipt)
	}
}
