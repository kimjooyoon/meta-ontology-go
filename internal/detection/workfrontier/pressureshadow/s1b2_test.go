package pressureshadow

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
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

func s1b2Selector() workfrontier.Input {
	return workfrontier.Input{
		SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: "selector-snapshot",
		PolicyDigest: "selector-policy", RegistryDigest: "selector-registry",
		MinimumSelectedPressures: 2, Capacity: workfrontier.Capacity{CPUCoreNS: 10},
		Pressures: []workfrontier.Pressure{{StableID: "p-a"}, {StableID: "p-b"}, {StableID: "p-c"}},
		States: []workfrontier.ObligationState{
			{ObligationID: "obligation/a", Status: "PENDING"},
			{ObligationID: "obligation/b", Status: "PENDING"},
			{ObligationID: "obligation/c", Status: "PENDING"},
		},
		Paths: []workfrontier.RepairPath{
			s1b2Path("path/a", "obligation/a", "p-a"),
			s1b2Path("path/b", "obligation/b", "p-b"),
			s1b2Path("path/c", "obligation/c", "p-c"),
		},
	}
}

func s1b2Path(id, obligation, read string) workfrontier.RepairPath {
	return workfrontier.RepairPath{StableID: id, ObligationID: obligation,
		ReadSet: []string{read}, RequiredPressureIDs: ids(), CPUCoreNSUpperBound: 1}
}

func permuteS1B2(input *Input) {
	input.Selector.Paths[0], input.Selector.Paths[2] = input.Selector.Paths[2], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[2] = input.PathCoverage[2], input.PathCoverage[0]
	for index := range input.PathCoverage {
		coverage := &input.PathCoverage[index].Coverage
		coverage.RequiredPressureIDs[0], coverage.RequiredPressureIDs[2] =
			coverage.RequiredPressureIDs[2], coverage.RequiredPressureIDs[0]
		coverage.PressureRecords[0], coverage.PressureRecords[2] = coverage.PressureRecords[2], coverage.PressureRecords[0]
	}
}

func historicalS1B2Input() Input {
	input := s1b2Input()
	input.Selector.Pressures = []workfrontier.Pressure{{StableID: "p-a"}, {StableID: "p-b"}}
	input.Selector.States = []workfrontier.ObligationState{{ObligationID: "obligation/a", Status: "PENDING"}}
	input.Selector.Paths = []workfrontier.RepairPath{{
		StableID: "path/a", ObligationID: "obligation/a", ReadSet: []string{"p-a"},
		WriteSet: []string{"p-b"}, RequiredPressureIDs: []string{"p-a", "p-b"}, CPUCoreNSUpperBound: 1,
	}}
	input.PathCoverage = input.PathCoverage[:1]
	coverage := &input.PathCoverage[0].Coverage
	coverage.RequiredPressureIDs = []string{"p-a", "p-b"}
	coverage.PressureRecords = []pressurecoverage.PressureRecord{
		{PressureID: "p-a", CategoryID: "category-a", IndependenceGroupID: "group-a", ApplicabilityRuleID: "rule-1"},
		{PressureID: "p-b", CategoryID: "category-b", IndependenceGroupID: "group-a", ApplicabilityRuleID: "rule-1"},
	}
	rebindCoverage(&input, "path/a")
	return input
}
