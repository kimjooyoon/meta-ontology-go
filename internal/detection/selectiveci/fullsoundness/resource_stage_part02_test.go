package fullsoundness

import (
	"math"
	"testing"
)

func TestResourceAggregateOverflowIsEvaluated(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*ResourceReceipt)
	}{
		{"CPU sum", func(receipt *ResourceReceipt) { receipt.CPUCoreNS = math.MaxInt64 }},
		{"denominator product", func(receipt *ResourceReceipt) {
			receipt.WallNS = math.MaxInt64
			receipt.AllocatedCPUCount = 2
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := soundInput()
			test.mutate(findReceipt(input.FullResourceReceipts, id("command/guard")))
			got := Evaluate(input)
			if got.Decision != DecisionUnknown || got.Reason != ReasonResourceOverflow || !got.SemanticEvaluated {
				t.Fatalf("got %s/%s evaluated=%t", got.Decision, got.Reason, got.SemanticEvaluated)
			}
		})
	}
}
