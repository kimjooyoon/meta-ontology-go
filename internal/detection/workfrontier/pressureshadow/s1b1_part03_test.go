package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"reflect"
	"testing"
)

func TestS1B1MixedAndOpaqueSelectorChanges(t *testing.T) {
	input := s1b1Input()
	input.PathCoverage[1].Coverage.MinimumIndependent = 1
	rebindCoverage(&input, "path/b")
	input.Selector.Paths[2].RequiredPressureIDs = nil
	input.PathCoverage[2].Coverage.RequiredPressureIDs = nil
	rebindCoverage(&input, "path/c")
	got := ValidateS1B1(input)
	if got.Decision != DecisionFailClosed || got.Reason != ReasonPressureCoverageFailClosed ||
		len(got.A2Observations) != 3 ||
		!sameB1Values(got.PressureCoverageFailPathIDs, []string{"path/b"}) ||
		!sameB1Values(got.PressureCoverageUnknownPathIDs, []string{"path/c"}) {
		t.Fatalf("mixed result = %#v", got)
	}
	base := ValidateS1B1(s1b1Input())
	input = s1b1Input()
	input.Selector.Capacity.CPUCoreNS = 99
	input.Selector.Paths[0].PolicyPriority = 99
	input.Selector.States = []workfrontier.ObligationState{{ObligationID: "state/a", Status: "BLOCKED"}}
	changed := ValidateS1B1(input)
	if changed.Decision != base.Decision || changed.Reason != base.Reason ||
		!reflect.DeepEqual(changed.A2Observations, base.A2Observations) {
		t.Fatalf("selector-only mutation changed A2 semantics: %#v", changed)
	}
}
func TestS1B1K21AndStrictUpstream(t *testing.T) {
	input := s1b1Input()
	setK21(&input)
	if got := ValidateS1B1(input); got.Decision != DecisionPass ||
		!sameB1Values(got.PressureCoveragePassPathIDs, []string{"path/a", "path/b", "path/c"}) {
		t.Fatalf("K=21 result = %#v", got)
	}
	unknown, fail := ReasonUpstreamUnknown, ReasonUpstreamFailClosed
	for _, test := range []struct {
		name string
		edit func(*Input)
		want Decision
		why  Reason
	}{
		{"upstream unknown", func(input *Input) { input.Selector.SnapshotDigest = "" }, DecisionUnknown, unknown},
		{"upstream fail", func(input *Input) { input.Selector.Paths[0].StableID = "path a" }, DecisionFailClosed, fail},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := s1b1Input()
			test.edit(&input)
			got := ValidateS1B1(input)
			if got.Decision != test.want || got.Reason != test.why || len(got.A2Observations) != 0 {
				t.Fatalf("upstream result = %#v", got)
			}
		})
	}
}
