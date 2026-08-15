package lanefrontier

import (
	"bytes"
	"strings"
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

func TestOwnerAmbiguity(t *testing.T) {
	for _, owners := range [][]string{
		{"internal/billing", "internal/billing"},
		{"internal/billing/", "internal/billing"},
		{"internal", "internal/billing"},
		{".", "internal/billing"},
	} {
		input := baseInput()
		input.OwnedPathPrefixes = owners
		got := Classify(input)
		if got.Decision != DecisionUnknown || got.Reason != ReasonAmbiguousOwner {
			t.Fatalf("owners %v: got %s/%s", owners, got.Decision, got.Reason)
		}
	}
}

func TestPathTraversalAndAbsolutePathRejection(t *testing.T) {
	for _, changed := range []string{"../outside.go", "/absolute.go", "internal/../outside.go", "internal\\outside.go"} {
		input := baseInput()
		input.ChangedPaths = []string{changed}
		got := Classify(input)
		if got.Decision != DecisionIneligible || got.Reason != ReasonPathOutOfScope {
			t.Errorf("path %q: got %s/%s", changed, got.Decision, got.Reason)
		}
	}
}

func TestPermutationByteEquality(t *testing.T) {
	first := baseInput()
	first.OwnedPathPrefixes = []string{"internal/other", "internal/billing"}
	first.ChangedPaths = []string{"internal/other/file.go", "internal/billing/order.go", "internal/billing/order.go"}
	second := first
	second.OwnedPathPrefixes = []string{"internal/billing", "internal/other"}
	second.ChangedPaths = []string{"internal/billing/order.go", "internal/other/file.go"}

	left, err := EncodeJSON(Classify(first))
	if err != nil {
		t.Fatal(err)
	}
	right, err := EncodeJSON(Classify(second))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("permutations differ:\n%s\n%s", left, right)
	}
}

func TestObservedShapes(t *testing.T) {
	cases := []struct {
		name          string
		ahead, behind int64
		wantDecision  Decision
		wantReason    Reason
	}{
		{"clean lane 0/0", 0, 0, DecisionEligible, ReasonEligible},
		{"diverged lane 91/1", 91, 1, DecisionIneligible, ReasonDivergedBranch},
		{"diverged lane 55/30", 55, 30, DecisionIneligible, ReasonDivergedBranch},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := baseInput()
			input.AheadCount, input.BehindCount = test.ahead, test.behind
			got := Classify(input)
			if got.Decision != test.wantDecision || got.Reason != test.wantReason {
				t.Fatalf("got %s/%s, want %s/%s", got.Decision, got.Reason, test.wantDecision, test.wantReason)
			}
		})
	}
}

func TestStrictJSON(t *testing.T) {
	input, err := EncodeInputJSON(baseInput())
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyJSON(input); got.Reason != ReasonEligible {
		t.Fatalf("encoded input got %s", got.Reason)
	}
	duplicate := []byte(`{"schema_version":"` + SchemaVersion + `","schema_version":"` + SchemaVersion + `"}`)
	if got := ClassifyJSON(duplicate); got.Reason != ReasonMissingInput || got.Decision != DecisionUnknown {
		t.Fatalf("duplicate fields got %s/%s", got.Decision, got.Reason)
	}
	if got := ClassifyJSON([]byte(`{"schema_version":"v0"}`)); got.Reason != ReasonUnknownSchema || got.Decision != DecisionUnknown {
		t.Fatalf("unsupported schema got %s/%s", got.Decision, got.Reason)
	}
	fixture := baseInput()
	fixture.OwnedPathPrefixes = []string{"internal/billing", "internal/billing"}
	encoded, err := EncodeInputJSON(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyJSON(encoded); got.Reason != ReasonAmbiguousOwner {
		t.Fatalf("duplicate owners got %s/%s", got.Decision, got.Reason)
	}
}

func TestCanonicalFixtureDigest(t *testing.T) {
	got := Classify(baseInput())
	t.Logf("canonical fixture digest: %s", got.CanonicalDigest)
	if got.CanonicalDigest != got.StableDigest() {
		t.Fatal("fixture digest changed after canonicalization")
	}
}

func FuzzClassifyJSONNeverPanics(f *testing.F) {
	seed, err := EncodeInputJSON(baseInput())
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte(`{"schema_version":"bad"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		got := ClassifyJSON(data)
		if got.CanonicalDigest == "" || got.CanonicalDigest != got.StableDigest() {
			t.Fatal("result is not sealed")
		}
	})
}

func baseInput() Input {
	return Input{
		SchemaVersion:     SchemaVersion,
		RegistryDigest:    strings.Repeat("a", 64),
		BaseSHA:           strings.Repeat("b", 40),
		LaneHeadSHA:       strings.Repeat("c", 40),
		LaneID:            "lane://billing",
		RegisteredBranch:  "agent/billing",
		OwnedPathPrefixes: []string{"internal/billing"},
		ChangedPaths:      []string{"internal/billing/order.go"},
	}
}
