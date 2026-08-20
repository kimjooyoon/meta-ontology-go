package resourcevector

import "testing"

func TestAffectedIDsAreExplicitUnion(t *testing.T) {
	input := R4F01()
	got := Evaluate(input)
	if got.Selected == nil || got.Full == nil || got.Selected.AffectedStableIDs != 2 || got.Full.AffectedStableIDs != 2 {
		t.Fatalf("unaffected full command invented an ID: %#v", got)
	}

	overlap := R4F01()
	overlap.Commands[1].AffectedStableIDs = []string{"s-order"}
	overlap.Commands[2].AffectedStableIDs = []string{"s-payment"}
	got = Evaluate(overlap)
	if got.Decision != DecisionPass || got.Selected == nil || got.Full == nil || got.Selected.AffectedStableIDs != 1 || got.Full.AffectedStableIDs != 2 {
		t.Fatalf("same affected ID was counted more than once: %#v", got)
	}
}

func TestAffectedBindingMutations(t *testing.T) {
	duplicate := R4F01()
	duplicate.Commands[0].AffectedStableIDs = []string{"s-order", "s-order"}
	if got := Evaluate(duplicate); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicateAffectedBinding {
		t.Fatalf("duplicate affected binding = %#v", got)
	}

	dangling := R4F01()
	dangling.Commands[0].AffectedStableIDs = []string{"s-missing"}
	if got := Evaluate(dangling); got.Decision != DecisionFailClosed || got.Reason != ReasonDanglingAffectedBinding {
		t.Fatalf("dangling affected binding = %#v", got)
	}

	missing := R4F01()
	missing.Commands[0].AffectedStableIDs = nil
	got := Evaluate(missing)
	if got.Decision != DecisionUnknown || got.Reason != ReasonMissingAffectedBinding || got.Selected != nil || got.Full != nil {
		t.Fatalf("missing affected binding was zeroed: %#v", got)
	}
}

func TestEachResourceLimitFailsClosedWithoutCompensation(t *testing.T) {
	cases := []struct {
		name  string
		limit func(*CeilingSet)
	}{
		{name: "cpu", limit: func(ceiling *CeilingSet) { ceiling.CPUCoreNS = U64(32) }},
		{name: "memory", limit: func(ceiling *CeilingSet) { ceiling.MemoryBytes = U64(223) }},
		{name: "work", limit: func(ceiling *CeilingSet) { ceiling.WorkUnits = U64(17) }},
		{name: "prov", limit: func(ceiling *CeilingSet) { ceiling.UniquePROVRecords = U64(11) }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := R4F01()
			generous(&input.Ceilings.Selected)
			test.limit(&input.Ceilings.Selected)
			got := Evaluate(input)
			if got.Decision != DecisionFailClosed || got.Reason != ReasonResourceLimitExceeded || len(got.LimitFailures) != 1 {
				t.Fatalf("limit result = %#v", got)
			}
		})
	}
}

func generous(ceiling *CeilingSet) {
	value := ^uint64(0)
	ceiling.CPUCoreNS, ceiling.MemoryBytes, ceiling.PeakMemoryBytes = new(value), new(value), new(value)
	ceiling.WorkUnits, ceiling.AffectedStableIDs = new(value), new(value)
	ceiling.ApplicablePressures, ceiling.IndependentGroups = new(value), new(value)
	ceiling.UniquePROVRecords, ceiling.FinitePROVPaths = new(value), new(value)
	ceiling.ClosureNumerator, ceiling.ClosureDenominator = new(value), new(value)
}
