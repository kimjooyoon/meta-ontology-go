package pressurecoverage

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestObserveSelectsLexicographicRepresentatives(t *testing.T) {
	input := fixture()
	got := Observe(input)
	if got.Decision != PASS || got.RequiredPressureCount != 4 || got.DistinctGroupCount != 3 {
		t.Fatalf("result = %#v", got)
	}
	if !reflect.DeepEqual(got.SelectedIDs, []string{"pressure-a", "pressure-b"}) ||
		!reflect.DeepEqual(got.UnselectedIDs, []string{"pressure-aa", "pressure-z"}) || len(got.UnknownIDs) != 0 {
		t.Fatalf("partition = %#v", got)
	}
	if got.CPUCoreNS != 2 || got.DeterministicWorkUnits != 16 || got.MemoryBytes != 4096 || got.ProvRecords != 5 || got.ProvPaths != 3 {
		t.Fatalf("receipt = %#v", got)
	}
	if got.OutputDigest == "" || got.ReplayDigest == "" || got.OutputDigest != CanonicalOutputDigest(got) {
		t.Fatalf("digests = %#v", got)
	}
}

func TestObserveRejectsUnsafeOrUnprovenMutations(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*Input)
		decision Decision
		reason   Reason
	}{
		{"same group", func(input *Input) {
			input.PressureRecords[0].IndependenceGroupID, input.PressureRecords[2].IndependenceGroupID = "group-a", "group-a"
		}, UNKNOWN, ReasonIndependentGroupShortfall},
		{"missing group", func(input *Input) { input.PressureRecords[0].IndependenceGroupID = "" }, UNKNOWN, ReasonApplicabilityUnproven},
		{"missing applicability", func(input *Input) { input.PressureRecords[0].ApplicabilityRuleID = "" }, UNKNOWN, ReasonApplicabilityUnproven},
		{"ambiguous applicability", func(input *Input) { input.PressureRecords[0].ApplicabilityRuleID = "rule-2" }, UNKNOWN, ReasonInputAmbiguous},
		{"duplicate", func(input *Input) { input.PressureRecords = append(input.PressureRecords, input.PressureRecords[0]) }, FAIL_CLOSED, ReasonDuplicatePressureID},
		{"conflicting binding", func(input *Input) {
			input.PressureRecords = append(input.PressureRecords, PressureRecord{PressureID: "pressure-a", CategoryID: "category-a", IndependenceGroupID: "group-x", ApplicabilityRuleID: "rule-1"})
		}, FAIL_CLOSED, ReasonConflictingGroupBinding},
		{"malformed path", func(input *Input) { input.FinitePathIDs[0] = "" }, FAIL_CLOSED, ReasonMalformedFinitePath},
		{"stale digest", func(input *Input) { input.PolicyDigest = "stale" }, UNKNOWN, ReasonStaleDigest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := fixture()
			test.mutate(&input)
			got := Observe(input)
			if got.Decision != test.decision || got.Reason != test.reason || !got.FullSuiteRequired {
				t.Fatalf("result = %#v", got)
			}
		})
	}
}

func TestObserveChecksEachCeilingIndependently(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ResourceCeilings)
		reason Reason
	}{
		{"cpu", func(value *ResourceCeilings) { value.CPUCoreNS = 1 }, ReasonCPUCeilingExceeded},
		{"memory", func(value *ResourceCeilings) { value.MemoryBytes = 4095 }, ReasonMemoryCeilingExceeded},
		{"work", func(value *ResourceCeilings) { value.WorkUnits = 15 }, ReasonWorkCeilingExceeded},
		{"prov records", func(value *ResourceCeilings) { value.ProvRecords = 4 }, ReasonProvRecordCeilingExceeded},
		{"prov paths", func(value *ResourceCeilings) { value.ProvPaths = 2 }, ReasonProvPathCeilingExceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := fixture()
			test.mutate(&input.ResourceCeilings)
			got := Observe(input)
			if got.Decision != UNKNOWN || got.Reason != test.reason || !got.FullSuiteRequired {
				t.Fatalf("result = %#v", got)
			}
		})
	}
	input := fixture()
	input.ResourceCeilings.WorkUnits = 0
	if got := Observe(input); got.Decision != UNKNOWN || got.Reason != ReasonResourceCeilingMissing {
		t.Fatalf("missing ceiling = %#v", got)
	}
}

func TestObserveIsPermutationRootAndPresentationInvariant(t *testing.T) {
	base := Observe(fixture())
	permuted := fixture()
	permuted.PresentationRoot = "/relocated/worktree"
	permuted.RequiredPressureIDs[0], permuted.RequiredPressureIDs[3] = permuted.RequiredPressureIDs[3], permuted.RequiredPressureIDs[0]
	permuted.FinitePathIDs[0], permuted.FinitePathIDs[2] = permuted.FinitePathIDs[2], permuted.FinitePathIDs[0]
	permuted.PressureRecords[0].DisplayName = "renamed display"
	permuted.PressureRecords[0].PresentationMetadata = map[string]string{"color": "blue"}
	permuted.PressureRecords[0], permuted.PressureRecords[3] = permuted.PressureRecords[3], permuted.PressureRecords[0]
	got := Observe(permuted)
	if !reflect.DeepEqual(base, got) {
		t.Fatalf("invariant changed: base=%#v got=%#v", base, got)
	}
	encoded, _ := json.Marshal(struct {
		Input  Input
		Output Output
	}{permuted, got})
	if strings.Contains(strings.ToLower(string(encoded)), "expected") {
		t.Fatalf("unexpected expected-label echo: %s", encoded)
	}
}

func TestObserveKIsInputDriven(t *testing.T) {
	input := fixture()
	for group := 4; group <= 24; group++ {
		id := "pressure-" + string(rune('a'+group))
		input.RequiredPressureIDs = append(input.RequiredPressureIDs, id)
		input.PressureRecords = append(input.PressureRecords, PressureRecord{PressureID: id, CategoryID: "category-x", IndependenceGroupID: id, ApplicabilityRuleID: "rule-1"})
	}
	input.RequestedK, input.MinimumIndependent = 21, 2
	input.FinitePathIDs = append(input.FinitePathIDs, "path-4")
	input.ResourceCeilings = ResourceCeilings{CPUCoreNS: 100, MemoryBytes: 100000, WorkUnits: 1000, ProvRecords: 1000, ProvPaths: 1000}
	got := Observe(input)
	if got.Decision != PASS || len(got.SelectedIDs) != 21 || got.CPUCoreNS != 21 {
		t.Fatalf("result = %#v", got)
	}
}

func fixture() Input {
	return Input{Schema: SchemaVersion, AuthoritySnapshotDigest: digest("snapshot"), PolicyDigest: digest("policy"), RegistryDigest: digest("registry"), ToolchainOptionsDigest: digest("toolchain"), RequestedK: 2, MinimumIndependent: 2,
		PressureRecords: []PressureRecord{{"pressure-z", "category-z", "group-z", "rule-1", "", nil}, {"pressure-a", "category-a", "group-a", "rule-1", "", nil}, {"pressure-b", "category-b", "group-b", "rule-1", "", nil}, {"pressure-aa", "category-a", "group-a", "rule-1", "", nil}}, RequiredPressureIDs: []string{"pressure-z", "pressure-a", "pressure-b", "pressure-aa"}, FinitePathIDs: []string{"path-1", "path-2", "path-3"}, GuardIDs: []string{"guard-1"}, ResourceCeilings: ResourceCeilings{CPUCoreNS: 100, MemoryBytes: 10000, WorkUnits: 100, ProvRecords: 100, ProvPaths: 100}}
}

func digest(_ string) string { return "sha256:" + strings.Repeat("a", 64) }
