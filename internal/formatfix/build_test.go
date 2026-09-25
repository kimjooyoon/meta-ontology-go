package formatfix

import "testing"

const unformattedFixture = "package billing\nnamespace billing\nentity Payment id \"billing://entity/payment\"\nentity Order id \"billing://entity/order\"\nactivity PayOrder(Order) -> Payment\n"

func TestBuildApplyAndRebuildReachFixedPoint(t *testing.T) {
	plan := Build("billing.gooo", unformattedFixture)
	if err := Validate(plan); err != nil {
		t.Fatal(err)
	}
	if plan.Decision != DecisionChangePlanned || plan.Resolution != ResolutionExact ||
		!plan.SemanticEqual || len(plan.Edits) != 1 || plan.DirectWrites != 0 {
		t.Fatalf("plan = %#v", plan)
	}
	result, err := Apply(unformattedFixture, plan)
	if err != nil {
		t.Fatal(err)
	}
	fixed := Build("billing.gooo", result)
	if err := Validate(fixed); err != nil {
		t.Fatal(err)
	}
	if fixed.Decision != DecisionFixedPoint || fixed.Changed || fixed.DirectWrites != 0 {
		t.Fatalf("fixed = %#v", fixed)
	}
}
