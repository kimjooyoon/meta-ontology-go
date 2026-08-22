package predecessorresolution

import "testing"

func TestBuildResolvesNearestExactAncestor(t *testing.T) {
	current := fixtureSHA("d")
	first, second, selected := fixtureSHA("a"), fixtureSHA("b"), fixtureSHA("c")
	input := Input{Repository: "owner/repo", CurrentHeadSHA: current,
		ImmediatePredecessorSHA: first, SearchLimit: SearchLimit,
		Attempts: []Attempt{
			{Depth: 0, AncestorSHA: first, ParentSHA: second,
				Selection: missingSelection(current, first)},
			{Depth: 1, AncestorSHA: second, ParentSHA: selected,
				Selection: missingSelection(current, second)},
			{Depth: 2, AncestorSHA: selected,
				Selection: selectedSelection(current, selected)},
		}}
	report, err := Build(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionResolved || report.Selected.Depth != 2 ||
		report.Summary.ObservedAttempts != 3 || report.Summary.MissingAttempts != 2 ||
		report.Summary.CoordinatesCompleted != 10 ||
		report.Summary.CoordinatesTotal != 10 || report.Summary.BasisPoints != 10000 {
		t.Fatalf("unexpected exact resolution: %+v", report.Summary)
	}
	if report.Selected.Baseline.Completed != 10 || report.Selected.Baseline.Total != 24 {
		t.Fatalf("unexpected baseline: %+v", report.Selected.Baseline)
	}
}
