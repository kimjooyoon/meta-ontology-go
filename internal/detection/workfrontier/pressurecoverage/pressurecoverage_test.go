package pressurecoverage

import "testing"

func TestPositiveK2(t *testing.T) {
	input := fixture()
	got := Observe(input)
	if got.Decision != DecisionPass || got.RequiredPressureCount != 4 || got.DistinctGroupCount != 3 {
		t.Fatalf("result = %#v", got)
	}
	if len(got.SelectedIDs) != 2 || got.SelectedIDs[0] != "pressure-a" || got.SelectedIDs[1] != "pressure-b" {
		t.Fatalf("selected = %#v", got.SelectedIDs)
	}
	if len(got.UnselectedIDs) != 2 || got.UnselectedIDs[0] != "pressure-aa" || got.UnselectedIDs[1] != "pressure-z" {
		t.Fatalf("partition = %#v", got)
	}
	if got.DeterministicWorkUnits != 16 || got.CPUCoreNS != 2 ||
		got.MemoryBytes != 4096 || got.ProvRecords != 5 || got.ProvPaths != 3 {
		t.Fatalf("cost vector = %#v", got)
	}
	if got.InputDigest != CanonicalInputDigest(input) ||
		got.OutputDigest != CanonicalOutputDigest(got) || got.ReplayDigest == "" {
		t.Fatalf("digests = %#v", got)
	}
}

func TestMutationDecisions(t *testing.T) {
	cases := []struct {
		name     string
		decision Decision
		reason   Reason
	}{
		{"same group", DecisionUnknown, ReasonIndependentGroupShortfall},
		{"missing group", DecisionUnknown, ReasonApplicabilityUnproven},
		{"missing applicability", DecisionUnknown, ReasonApplicabilityUnproven},
		{"duplicate", DecisionFailClosed, ReasonDuplicatePressureID},
		{"conflicting binding", DecisionFailClosed, ReasonConflictingGroupBinding},
		{"stale digest", DecisionUnknown, ReasonStaleDigest},
		{"empty required", DecisionUnknown, ReasonRequiredInputMissing},
		{"empty guards", DecisionUnknown, ReasonRequiredInputMissing},
		{"empty paths", DecisionUnknown, ReasonRequiredInputMissing},
		{"internal whitespace", DecisionUnknown, ReasonCatalogMismatch},
		{"control ID", DecisionUnknown, ReasonCatalogMismatch},
	}
	for _, test := range cases {
		input := fixture()
		mutate(&input, test.name)
		got := Observe(input)
		if got.Decision != test.decision || got.Reason != test.reason || !got.FullSuiteRequired {
			t.Fatalf("%s: result = %#v", test.name, got)
		}
	}
}

func mutate(input *Input, name string) {
	switch name {
	case "same group":
		input.PressureRecords[0].IndependenceGroupID, input.PressureRecords[2].IndependenceGroupID = "group-a", "group-a"
	case "missing group":
		input.PressureRecords[0].IndependenceGroupID = ""
	case "missing applicability":
		input.PressureRecords[0].ApplicabilityRuleID = ""
	case "duplicate":
		input.PressureRecords = append(input.PressureRecords, input.PressureRecords[0])
	case "conflicting binding":
		input.PressureRecords = append(input.PressureRecords, PressureRecord{"pressure-a", "category-a", "group-x", "rule-1"})
	case "stale digest":
		input.PolicyDigest = digestBytes([]byte("stale-policy"))
	case "empty required":
		input.RequiredPressureIDs = []string{}
	case "empty guards":
		input.GuardIDs = []string{}
	case "empty paths":
		input.FinitePathIDs = []string{}
	case "internal whitespace":
		input.RequiredPressureIDs[0] = "pressure z"
	case "control ID":
		input.RequiredPressureIDs[0] = "pressure-z\x00"
	}
	if name != "stale digest" {
		bindDigests(input)
	}
}

func TestPermutationInvariance(t *testing.T) {
	base := Observe(fixture())
	input := fixture()
	input.RequiredPressureIDs[0], input.RequiredPressureIDs[3] = input.RequiredPressureIDs[3], input.RequiredPressureIDs[0]
	input.PressureRecords[0], input.PressureRecords[3] = input.PressureRecords[3], input.PressureRecords[0]
	if got := Observe(input); got.OutputDigest != base.OutputDigest {
		t.Fatalf("invariant changed: %#v != %#v", base, got)
	}
}

func fixture() Input {
	records := []PressureRecord{
		{"pressure-z", "category-z", "group-z", "rule-1"},
		{"pressure-a", "category-a", "group-a", "rule-1"},
		{"pressure-b", "category-b", "group-b", "rule-1"},
		{"pressure-aa", "category-a", "group-a", "rule-1"},
	}
	input := Input{
		Schema:              SchemaVersion,
		RequestedK:          2,
		MinimumIndependent:  2,
		PressureRecords:     records,
		RequiredPressureIDs: []string{"pressure-z", "pressure-a", "pressure-b", "pressure-aa"},
		FinitePathIDs:       []string{"path-1", "path-2", "path-3"},
		GuardIDs:            []string{"guard-1"},
	}
	bindDigests(&input)
	return input
}

func bindDigests(input *Input) {
	input.AuthoritySnapshotDigest = authorityBindingDigest(*input, "authority-snapshot")
	input.PolicyDigest = authorityBindingDigest(*input, "policy")
	input.RegistryDigest = authorityBindingDigest(*input, "registry")
	input.ToolchainOptionsDigest = authorityBindingDigest(*input, "toolchain-options")
}
