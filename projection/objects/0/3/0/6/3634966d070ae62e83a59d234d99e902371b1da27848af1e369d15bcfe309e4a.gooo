package fullsoundness

import (
	"math"
	"testing"
)

func TestResourceClasses(t *testing.T) {
	equal := soundInput()
	selectOnly(&equal, []string{id("command/guard"), id("command/impact"), id("command/pass")})
	if got := Evaluate(equal); got.Decision != DecisionSound || got.ResourceVector.Class != ResourceEqual {
		t.Fatalf("equal resources got %#v", got)
	}
	regressed := soundInput()
	findReceipt(regressed.SelectedResourceReceipts, id("command/impact")).CPUCoreNS = 20
	if got := Evaluate(regressed); got.Decision != DecisionSound || got.ResourceVector.Class != ResourceRegressed {
		t.Fatalf("regressed resources got %#v", got)
	}
	higherUtilization := soundInput()
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/guard")).CPUCoreNS = 1
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/guard")).WallNS = 1
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/guard")).PeakRSSBytes = 1
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/guard")).ReadBytes = 1
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/guard")).WriteBytes = 1
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/impact")).CPUCoreNS = 2
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/impact")).WallNS = 1
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/impact")).PeakRSSBytes = 1
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/impact")).ReadBytes = 1
	findReceipt(higherUtilization.SelectedResourceReceipts, id("command/impact")).WriteBytes = 1
	got := Evaluate(higherUtilization)
	if got.Decision != DecisionSound || got.ResourceVector.Class != ResourceImproved {
		t.Fatalf("higher utilization resources got %#v", got)
	}
	if got.ResourceVector.Full.Utilization != (Utilization{Numerator: 10, Denominator: 10}) || got.ResourceVector.Selected.Utilization != (Utilization{Numerator: 3, Denominator: 2}) {
		t.Fatalf("utilization was not retained exactly: %#v", got.ResourceVector)
	}
}
func TestSemanticEvaluationStage(t *testing.T) {
	cases := []struct {
		name      string
		mutate    func(*Input)
		decision  Decision
		reason    Reason
		evaluated bool
		selected  uint64
	}{
		{"missing full receipt", func(input *Input) { input.FullResourceReceipts = nil }, DecisionUnknown, ReasonFullSuiteRequired, false, 2},
		{"missing selected receipt", func(input *Input) { input.SelectedResourceReceipts = nil }, DecisionUnknown, ReasonFullSuiteRequired, false, 2},
		{"global guard omitted", func(input *Input) { selectOnly(input, []string{id("command/impact")}) }, DecisionUnsound, ReasonGlobalGuardOmitted, false, 1},
		{"resource overflow", func(input *Input) { input.FullResourceReceipts[0].CPUCoreNS = math.MaxInt64 }, DecisionUnknown, ReasonResourceOverflow, true, 2},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := soundInput()
			test.mutate(&input)
			got := Evaluate(input)
			if got.Decision != test.decision || got.Reason != test.reason || got.SemanticEvaluated != test.evaluated {
				t.Fatalf("got %#v", got)
			}
			if got.CommandCount != 3 || got.SelectedCommandCount != test.selected || got.ObligationCount != 2 {
				t.Fatalf("raw counts = %#v", got)
			}
			assertDecisionSemanticStage(t, got)
		})
	}
}
