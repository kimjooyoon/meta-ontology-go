package fullsoundness

import "testing"

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
