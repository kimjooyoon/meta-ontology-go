package lanefrontier

import (
	"testing"
)

func TestReasonPartition(t *testing.T) {
	cases := []struct {
		name     string
		mutate   func(*Input)
		decision Decision
		reason   Reason
	}{
		{"unknown schema", func(input *Input) { input.SchemaVersion = "v0" }, DecisionUnknown, ReasonUnknownSchema},
		{"missing input", func(input *Input) { input.RegistryDigest = "" }, DecisionUnknown, ReasonMissingInput},
		{"invalid count", func(input *Input) { input.AheadCount = -1 }, DecisionUnknown, ReasonInvalidCount},
		{"ambiguous owner", func(input *Input) { input.OwnedPathPrefixes = []string{"internal", "internal/billing"} }, DecisionUnknown, ReasonAmbiguousOwner},
		{"path out of scope", func(input *Input) { input.ChangedPaths = []string{"internal/other/file.go"} }, DecisionIneligible, ReasonPathOutOfScope},
		{"active lease", func(input *Input) { input.ActiveLeaseCount = 1 }, DecisionIneligible, ReasonActiveLease},
		{"active PR", func(input *Input) { input.OpenPRCount = 1 }, DecisionIneligible, ReasonActivePR},
		{"diverged branch", func(input *Input) { input.AheadCount, input.BehindCount = 91, 1 }, DecisionIneligible, ReasonDivergedBranch},
		{"stale branch", func(input *Input) { input.BehindCount = 1 }, DecisionIneligible, ReasonStaleBranch},
		{"branch ahead", func(input *Input) { input.AheadCount = 1 }, DecisionIneligible, ReasonBranchAhead},
		{"eligible", func(input *Input) {}, DecisionEligible, ReasonEligible},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := baseInput()
			test.mutate(&input)
			got := Classify(input)
			if got.Decision != test.decision || got.Reason != test.reason {
				t.Fatalf("got %s/%s, want %s/%s", got.Decision, got.Reason, test.decision, test.reason)
			}
			if got.CanonicalDigest == "" || got.CanonicalDigest != got.StableDigest() {
				t.Fatalf("output digest is not canonical: %q", got.CanonicalDigest)
			}
		})
	}
}
func TestReasonOrder(t *testing.T) {
	input := baseInput()
	input.RegistryDigest = ""
	input.AheadCount, input.BehindCount = 2, 1
	input.ActiveLeaseCount = 1
	got := Classify(input)
	if got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
		t.Fatalf("got %s/%s, want UNKNOWN/MISSING_INPUT", got.Decision, got.Reason)
	}

	input = baseInput()
	input.OwnedPathPrefixes = []string{"internal", "internal/billing"}
	input.ChangedPaths = []string{"outside/file.go"}
	got = Classify(input)
	if got.Decision != DecisionUnknown || got.Reason != ReasonAmbiguousOwner {
		t.Fatalf("got %s/%s, want UNKNOWN/AMBIGUOUS_OWNER", got.Decision, got.Reason)
	}

	input = baseInput()
	input.ChangedPaths = []string{"outside/file.go"}
	input.ActiveLeaseCount = 1
	got = Classify(input)
	if got.Decision != DecisionIneligible || got.Reason != ReasonPathOutOfScope {
		t.Fatalf("got %s/%s, want INELIGIBLE/PATH_OUT_OF_SCOPE", got.Decision, got.Reason)
	}
}
