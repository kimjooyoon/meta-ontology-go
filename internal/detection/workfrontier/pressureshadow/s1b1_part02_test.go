package pressureshadow

import (
	"testing"
)

var s1b1A2Cases = []struct {
	name  string
	edit  func(*Input)
	wantD Decision
	wantR Reason
}{
	{"same group", func(input *Input) { setGroups(input, "group-a", "path/b") }, DecisionUnknown, s1b1Unknown},
	{"blank group", func(input *Input) { setGroups(input, "", "path/b") }, DecisionUnknown, s1b1Unknown},
	{"blank applicability", blankApplicability, DecisionUnknown, s1b1Unknown},
	{"stale inner binding", staleCoverage, DecisionUnknown, s1b1Unknown},
	{"policy floor", func(input *Input) {
		input.PathCoverage[1].Coverage.MinimumIndependent = 1
		rebindCoverage(input, "path/b")
	}, DecisionFailClosed, s1b1Fail},
	{"cardinality shortfall", func(input *Input) {
		setK(input, 4)
		for _, id := range []string{"path/a", "path/b", "path/c"} {
			rebindCoverage(input, id)
		}
	}, DecisionUnknown, s1b1Unknown},
	{"empty required", func(input *Input) {
		for index := range input.Selector.Paths {
			input.Selector.Paths[index].RequiredPressureIDs = nil
		}
		for index := range input.PathCoverage {
			input.PathCoverage[index].Coverage.RequiredPressureIDs = nil
		}
		for _, id := range []string{"path/a", "path/b", "path/c"} {
			rebindCoverage(input, id)
		}
	}, DecisionUnknown, s1b1Unknown},
}

func TestS1B1A2Vectors(t *testing.T) {
	for _, test := range s1b1A2Cases {
		t.Run(test.name, func(t *testing.T) {
			input := s1b1Input()
			test.edit(&input)
			got := ValidateS1B1(input)
			if got.Decision != test.wantD || got.Reason != test.wantR || len(got.A2Observations) != 3 {
				t.Fatalf("vector result = %#v", got)
			}
		})
	}
}
