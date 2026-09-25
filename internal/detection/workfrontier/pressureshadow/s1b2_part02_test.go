package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"reflect"
	"testing"
)

func TestS1B2UnselectedAndSelectorStates(t *testing.T) {
	cases := []struct {
		name     string
		edit     func(*Input)
		status   workfrontier.Decision
		decision Decision
		reason   Reason
	}{
		{"unknown blocked", func(input *Input) {
			setGroups(input, "group-a", "path/b")
			input.Selector.States[1].Status = "PASS"
		}, workfrontier.DecisionPass, DecisionUnknown, ReasonPressureCoverageUnknown},
		{"fail blocked", func(input *Input) {
			b2Coverage(input, "path/b").Coverage.MinimumIndependent = 1
			rebindCoverage(input, "path/b")
			input.Selector.States[1].Status = "PASS"
		}, workfrontier.DecisionPass, DecisionFailClosed, ReasonPressureCoverageFailClosed},
		{"selector unknown", func(input *Input) { input.Selector.States = nil },
			workfrontier.DecisionUnknown, DecisionUnknown, ReasonSelectorUnknown},
		{"selector blocked", func(input *Input) {
			for index := range input.Selector.States {
				input.Selector.States[index].Status = "PASS"
			}
		}, workfrontier.DecisionBlocked, DecisionPass, ReasonNone},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := s1b2Input()
			test.edit(&input)
			got := ValidateS1B2(input)
			if got.Decision != test.decision || got.Reason != test.reason ||
				got.SelectorResult == nil || got.SelectorResult.Status != test.status ||
				len(got.UnsafeSelectedFailPathIDs) != 0 || len(got.UnsafeSelectedUnknownPathIDs) != 0 {
				t.Fatalf("state result = %#v", got)
			}
		})
	}
}
func TestS1B2PermutationAndK21(t *testing.T) {
	input := s1b2Input()
	want := ValidateS1B2(input)
	permuteS1B2(&input)
	if got := ValidateS1B2(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("permutation changed result: %#v", got)
	}
	input = s1b2Input()
	setK21(&input)
	got := ValidateS1B2(input)
	if got.Decision != DecisionPass || got.SelectorResult == nil ||
		got.SelectorResult.Status != workfrontier.DecisionPass ||
		!sameB1Values(got.SelectorResult.SelectedIDs, []string{"path/a", "path/b", "path/c"}) {
		t.Fatalf("K=21 result = %#v", got)
	}
}
