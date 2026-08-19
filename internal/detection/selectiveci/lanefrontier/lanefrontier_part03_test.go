package lanefrontier

import (
	"testing"
)

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
