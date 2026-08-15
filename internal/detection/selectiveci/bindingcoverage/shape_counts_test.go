package bindingcoverage

import "testing"

func TestNilAndExplicitEmptyLists(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
	}{
		{"nil required bindings", func(input *Input) { input.RequiredBindings = nil }},
		{"nil partitions", func(input *Input) { input.Partitions = nil }},
		{"nil precedence registry", func(input *Input) { input.PrecedenceRegistry = nil }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := fixtureInput()
			test.mutate(&input)
			got := Observe(input)
			if got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
				t.Fatalf("got %s/%s, want UNKNOWN/MISSING_INPUT", got.Decision, got.Reason)
			}
		})
	}
	empty := fixtureInput()
	empty.RequiredBindings = []RequiredBinding{}
	empty.Partitions = []Partition{}
	empty.PrecedenceRegistry = []PrecedenceEntry{}
	got := Observe(empty)
	if got.Decision != DecisionIncomplete || got.Reason != ReasonZeroDenominator {
		t.Fatalf("explicit empty input got %s/%s, want INCOMPLETE/ZERO_DENOMINATOR", got.Decision, got.Reason)
	}
	assertShapeCounts(t, got, empty)
}

func TestUnknownSchemaRetainsShapeCounts(t *testing.T) {
	input := fixtureInput()
	input.SchemaVersion = "gooo/other/v1"
	got := Observe(input)
	if got.Decision != DecisionUnknown || got.Reason != ReasonUnknownSchema {
		t.Fatalf("got %s/%s, want UNKNOWN/UNKNOWN_SCHEMA", got.Decision, got.Reason)
	}
	assertShapeCounts(t, got, input)
	if len(got.MissingMatchBindingIDs) != 0 || len(got.MissingMismatchBindingIDs) != 0 {
		t.Fatalf("UNKNOWN output reported coverage gaps: %#v", got)
	}
}

func assertShapeCounts(t *testing.T, got Output, input Input) {
	t.Helper()
	required := uint64(len(input.RequiredBindings))
	partitions := uint64(len(input.Partitions))
	endpointReferences := required * 2
	work := required + partitions + endpointReferences
	if got.RequiredBindingCount != required || got.PartitionCount != partitions || got.EndpointReferenceCount != endpointReferences || got.DeterministicWorkUnits != work {
		t.Fatalf("shape counts = %d/%d/%d/%d, want %d/%d/%d/%d", got.RequiredBindingCount, got.PartitionCount, got.EndpointReferenceCount, got.DeterministicWorkUnits, required, partitions, endpointReferences, work)
	}
}
