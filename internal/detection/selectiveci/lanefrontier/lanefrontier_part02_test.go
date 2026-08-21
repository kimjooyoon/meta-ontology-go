package lanefrontier

import (
	"bytes"
	"testing"
)

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
func TestPathValidationAndScope(t *testing.T) {
	cases := []struct {
		name     string
		owned    bool
		path     string
		decision Decision
		reason   Reason
	}{
		{"absolute-path-invalid", false, "/absolute.go", DecisionUnknown, ReasonMissingInput},
		{"traversal-path-invalid", false, "../outside.go", DecisionUnknown, ReasonMissingInput},
		{"empty-path-invalid", false, "", DecisionUnknown, ReasonMissingInput},
		{"normalized-path-out-of-scope", false, "outside/file.go", DecisionIneligible, ReasonPathOutOfScope},
		{"owned-absolute-path-invalid", true, "/internal", DecisionUnknown, ReasonMissingInput},
		{"owned-traversal-path-invalid", true, "../internal", DecisionUnknown, ReasonMissingInput},
		{"owned-empty-path-invalid", true, "", DecisionUnknown, ReasonMissingInput},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := baseInput()
			if test.owned {
				input.OwnedPathPrefixes = []string{test.path}
			} else {
				input.ChangedPaths = []string{test.path}
			}
			got := Classify(input)
			if got.Decision != test.decision || got.Reason != test.reason {
				t.Fatalf("path %q: got %s/%s, want %s/%s", test.path, got.Decision, got.Reason, test.decision, test.reason)
			}
		})
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
