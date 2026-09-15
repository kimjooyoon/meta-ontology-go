package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"testing"
)

func TestS1B2PassAndEvidence(t *testing.T) {
	got := ValidateS1B2(s1b2Input())
	if got.Decision != DecisionPass || got.Reason != ReasonNone || !got.SelectorObserved ||
		got.SelectorResult == nil || got.SelectorResultDigest == "" || got.ResultDigest == "" ||
		got.ReplayDigest == "" || got.ExecutionAuthorized || got.EnforcementEffect != EnforcementNoEffect {
		t.Fatalf("positive result = %#v", got)
	}
	selector := got.SelectorResult
	if selector.Status != workfrontier.DecisionPass || selector.Quality != "MAXIMAL" ||
		selector.FullSuiteRequired || len(selector.Selected) != 3 ||
		!sameB1Values(selector.SelectedIDs, []string{"path/a", "path/b", "path/c"}) ||
		len(selector.Unknown) != 0 || len(selector.Blocked) != 0 || len(selector.Shortfall) != 0 {
		t.Fatalf("selector evidence = %#v", selector)
	}
	if len(got.UnsafeSelectedFailPathIDs) != 0 || len(got.UnsafeSelectedUnknownPathIDs) != 0 {
		t.Fatalf("positive conflicts = %#v", got)
	}
	t.Logf("base S1B2 digests input=%s upstream=%s selector=%s result=%s replay=%s",
		got.InputDigest, got.UpstreamResultDigest, got.SelectorResultDigest, got.ResultDigest, got.ReplayDigest)
}
func TestS1B2SelectedPrecedence(t *testing.T) {
	cases := []struct {
		name          string
		edit          func(*Input)
		decision      Decision
		reason        Reason
		fail, unknown []string
	}{
		{"selected unknown", func(input *Input) { setGroups(input, "group-a", "path/b") },
			DecisionUnknown, ReasonSelectorSelectedUnknownPressureCoverage, nil, []string{"path/b"}},
		{"selected fail", func(input *Input) {
			b2Coverage(input, "path/b").Coverage.MinimumIndependent = 1
			rebindCoverage(input, "path/b")
		}, DecisionFailClosed, ReasonSelectorSelectedFailedPressureCoverage, []string{"path/b"}, nil},
		{"selected fail wins", func(input *Input) {
			b2Coverage(input, "path/b").Coverage.MinimumIndependent = 1
			rebindCoverage(input, "path/b")
			setGroups(input, "group-a", "path/c")
		}, DecisionFailClosed, ReasonSelectorSelectedFailedPressureCoverage, []string{"path/b"}, []string{"path/c"}},
		{"unselected fail wins", func(input *Input) {
			setGroups(input, "group-a", "path/b")
			b2Coverage(input, "path/c").Coverage.MinimumIndependent = 1
			rebindCoverage(input, "path/c")
			input.Selector.States[2].Status = "PASS"
		}, DecisionFailClosed, ReasonPressureCoverageFailClosed, nil, []string{"path/b"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := s1b2Input()
			test.edit(&input)
			got := ValidateS1B2(input)
			if got.Decision != test.decision || got.Reason != test.reason ||
				!sameB1Values(got.UnsafeSelectedFailPathIDs, test.fail) ||
				!sameB1Values(got.UnsafeSelectedUnknownPathIDs, test.unknown) {
				t.Fatalf("precedence result = %#v", got)
			}
		})
	}
}
