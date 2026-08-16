package fullsoundness

import (
	"math"
	"testing"
)

func TestStaleResourceReceiptsAreFullSuiteRequired(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
	}{
		{"full snapshot", staleFullSnapshot},
		{"full toolchain", staleFullToolchain},
		{"full runner", staleFullRunner},
		{"selected snapshot", staleSelectedSnapshot},
		{"selected toolchain", staleSelectedToolchain},
		{"selected runner", staleSelectedRunner},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := soundInput()
			test.mutate(&input)
			got := Evaluate(input)
			if got.Decision != DecisionUnknown || got.Reason != ReasonFullSuiteRequired || got.SemanticEvaluated {
				t.Fatalf("got %s/%s evaluated=%t", got.Decision, got.Reason, got.SemanticEvaluated)
			}
		})
	}
}

func staleFullSnapshot(input *Input) {
	findReceipt(input.FullResourceReceipts, id("command/impact")).SnapshotDigest = digest("0")
}

func staleFullToolchain(input *Input) {
	findReceipt(input.FullResourceReceipts, id("command/impact")).ToolchainDigest = digest("0")
}

func staleFullRunner(input *Input) {
	findReceipt(input.FullResourceReceipts, id("command/impact")).RunnerDigest = digest("0")
}

func staleSelectedSnapshot(input *Input) {
	findReceipt(input.SelectedResourceReceipts, id("command/impact")).SnapshotDigest = digest("0")
}

func staleSelectedToolchain(input *Input) {
	findReceipt(input.SelectedResourceReceipts, id("command/impact")).ToolchainDigest = digest("0")
}

func staleSelectedRunner(input *Input) {
	findReceipt(input.SelectedResourceReceipts, id("command/impact")).RunnerDigest = digest("0")
}

func TestMalformedResourceNumbersAreFullSuiteRequired(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*ResourceReceipt)
	}{
		{"negative CPU", func(receipt *ResourceReceipt) { receipt.CPUCoreNS = -1 }},
		{"zero allocation", func(receipt *ResourceReceipt) { receipt.AllocatedCPUCount = 0 }},
		{"zero wall", func(receipt *ResourceReceipt) { receipt.WallNS = 0 }},
		{"negative RSS", func(receipt *ResourceReceipt) { receipt.PeakRSSBytes = -1 }},
		{"negative read", func(receipt *ResourceReceipt) { receipt.ReadBytes = -1 }},
		{"negative write", func(receipt *ResourceReceipt) { receipt.WriteBytes = -1 }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := soundInput()
			test.mutate(findReceipt(input.FullResourceReceipts, id("command/impact")))
			got := Evaluate(input)
			if got.Decision != DecisionUnknown || got.Reason != ReasonFullSuiteRequired || got.SemanticEvaluated {
				t.Fatalf("got %s/%s evaluated=%t", got.Decision, got.Reason, got.SemanticEvaluated)
			}
		})
	}
}

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
