package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"reflect"
	"strings"
	"testing"
)

func TestS1B2StructuralAndStrictBytes(t *testing.T) {
	malformed := s1b2Input()
	malformed.Selector.Paths[0].StableID = "path a"
	assertStructuralS1B2(t, ValidateS1B2(malformed), DecisionFailClosed, ReasonUpstreamFailClosed)
	missing := s1b2Input()
	missing.PathCoverage = missing.PathCoverage[:2]
	assertStructuralS1B2(t, ValidateS1B2(missing), DecisionUnknown, ReasonUpstreamUnknown)
	raw := b2RawInput
	mutations := []string{
		strings.Replace(raw, `"schema":`, `"expected_label":"PASS", "schema":`, 1),
		strings.Replace(raw, `"schema":`, `"schema":"duplicate", "schema":`, 1),
		raw + `{}`,
		strings.Replace(raw, `"path/a"`, `"path a"`, 1),
	}
	for _, data := range mutations {
		assertStructuralS1B2(t, ValidateS1B2Bytes([]byte(data)), DecisionFailClosed, ReasonUpstreamFailClosed)
	}
}
func TestS1B2SelectorOnlyMutations(t *testing.T) {
	base := s1b2Input()
	wantA2 := ValidateS1B1(base).A2Observations
	want := ValidateS1B2(base)
	mutations := []func(*Input){
		func(input *Input) { input.Selector.Capacity.CPUCoreNS = 0 },
		func(input *Input) { input.Selector.Paths[0].PolicyPriority = 99 },
		func(input *Input) { input.Selector.States[0].Status = "PASS" },
	}
	for _, edit := range mutations {
		input := s1b2Input()
		edit(&input)
		got := ValidateS1B2(input)
		if !reflect.DeepEqual(ValidateS1B1(input).A2Observations, wantA2) ||
			got.SelectorResultDigest == want.SelectorResultDigest {
			t.Fatalf("selector mutation changed wrong evidence: %#v", got)
		}
	}
}
func TestS1B2HistoricalFalseAcceptance(t *testing.T) {
	input := historicalS1B2Input()
	selector := workfrontier.Select(input.Selector)
	if selector.Status != workfrontier.DecisionPass || !sameB1Values(selector.SelectedIDs, []string{"path/a"}) {
		t.Fatalf("historical selector = %#v", selector)
	}
	got := ValidateS1B2(input)
	if got.Decision != DecisionUnknown || got.Reason != ReasonSelectorSelectedUnknownPressureCoverage ||
		!sameB1Values(got.UnsafeSelectedUnknownPathIDs, []string{"path/a"}) {
		t.Fatalf("historical shadow = %#v", got)
	}
}
func assertStructuralS1B2(t *testing.T, got S1B2Result, decision Decision, reason Reason) {
	t.Helper()
	if got.Decision != decision || got.Reason != reason || got.SelectorObserved || got.SelectorResult != nil ||
		got.SelectorResultDigest != "" || len(got.UnsafeSelectedFailPathIDs) != 0 ||
		len(got.UnsafeSelectedUnknownPathIDs) != 0 {
		t.Fatalf("structural result = %#v", got)
	}
}
func s1b2Input() Input {
	input := s1b1Input()
	input.Selector = s1b2Selector()
	return input
}
