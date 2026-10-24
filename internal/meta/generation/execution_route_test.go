package generation

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestObservedPlanSkipsUnimplementedConsumerAndRetainsItsDebt(t *testing.T) {
	head := strings.Repeat("b", 40)
	report := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{
			metric("expression", sourcepolicy.OperationCollapseAssign, false, false),
			metric("unconnected.gooo", sourcepolicy.OperationSplitGooo, false, false),
			metric("later.go:3:Large", sourcepolicy.OperationExtractFunction, false, false),
		}}
	var visited []string
	observe := func(action Action) CandidatePrerequisite {
		visited = append(visited, action.Subject)
		if action.Operation == sourcepolicy.OperationExtractFunction {
			observed := NewCandidatePrerequisite(action, head, strings.Repeat("c", 64), strings.Repeat("d", 64), "EXTRACTION_PLAN_READY")
			observed.Stage, observed.Step = "prepare-extraction", "stage-without-writing"
			return observed
		}
		return NewCandidatePrerequisite(action, head, "", strings.Repeat("d", 64), "EXECUTION_CONSUMER_UNAVAILABLE")
	}
	plan := BuildWithPrerequisites(head, head, report, nil, observe)
	if plan.Decision != DecisionPlan || len(plan.Selected) != 2 ||
		plan.Selected[1].Operation != sourcepolicy.OperationExtractFunction ||
		!reflect.DeepEqual(visited, []string{"unconnected.gooo", "later.go:3:Large"}) || !receiptPlanKnown(plan) {
		t.Fatalf("unimplemented consumer hid the later native operation: %+v %v", plan, visited)
	}
	if plan.IndicatorDecisionLedger.IndicatorCount != 3 || plan.IndicatorDecisionLedger.DeferredCount != 1 ||
		!slices.ContainsFunc(plan.Prerequisites, func(p CandidatePrerequisite) bool { return p.Reason == "EXECUTION_CONSUMER_UNAVAILABLE" }) {
		t.Fatal("unimplemented operation lost its indicator or prerequisite", plan)
	}
	forged := BuildWithPrerequisites(head, head, report, nil, func(action Action) CandidatePrerequisite {
		return NewCandidatePrerequisite(action, head, strings.Repeat("c", 64), strings.Repeat("d", 64), "SPLIT_PLAN_READY")
	})
	if forged.Decision != DecisionUnknown || forged.Reason != ReasonInvalidInput || len(forged.Selected) != 0 {
		t.Fatal("split readiness substituted for a missing native consumer", forged)
	}
}

func TestNativeRouteCannotBeAddedByResealingAPlan(t *testing.T) {
	plan := Build(strings.Repeat("a", 40), strings.Repeat("b", 40), prerequisiteMetrics())
	plan.Registry = append([]Binding{}, plan.Registry...)
	for index := range plan.Registry {
		if plan.Registry[index].Operation == sourcepolicy.OperationSplitGooo {
			plan.Registry[index].ExecutionRoute = RouteSplitGo
		}
	}
	plan.RegistryDigest = digestJSON(plan.Registry)
	if receiptPlanKnown(finish(plan)) {
		t.Fatal("caller-controlled native route was accepted")
	}
}
