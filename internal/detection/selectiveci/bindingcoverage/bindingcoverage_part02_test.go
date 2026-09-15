package bindingcoverage

import (
	"testing"
)

func TestIncompleteCoverage(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		reason Reason
		check  func(*testing.T, Output)
	}{
		{"omit lane registry mismatch", func(input *Input) {
			input.Partitions = withoutPartition(input.Partitions, bindingID("lane-registry"), PolarityMismatch)
		}, ReasonMissingMismatch, func(t *testing.T, got Output) {
			assertMissing(t, got.MissingMismatchBindingIDs, bindingID("lane-registry"))
			if len(got.MissingMatchBindingIDs) != 0 {
				t.Fatalf("unexpected missing MATCH IDs: %v", got.MissingMatchBindingIDs)
			}
		}},
		{"missing match", func(input *Input) {
			input.Partitions = withoutPartition(input.Partitions, bindingID("base-head"), PolarityMatch)
		}, ReasonMissingMatch, func(t *testing.T, got Output) {
			assertMissing(t, got.MissingMatchBindingIDs, bindingID("base-head"))
		}},
		{"zero denominator", func(input *Input) {
			input.RequiredBindings = []RequiredBinding{}
			input.Partitions = []Partition{}
		}, ReasonZeroDenominator, func(t *testing.T, got Output) {
			if got.RequiredBindingCount != 0 || got.PartitionCount != 0 || got.EndpointReferenceCount != 0 || got.DeterministicWorkUnits != 0 || len(got.MissingMatchBindingIDs) != 0 || len(got.MissingMismatchBindingIDs) != 0 {
				t.Fatalf("zero denominator output = %#v", got)
			}
		}},
		{"missing match and mismatch", func(input *Input) {
			input.Partitions = withoutPartition(input.Partitions, bindingID("base-head"), PolarityMatch)
			input.Partitions = withoutPartition(input.Partitions, bindingID("base-head"), PolarityMismatch)
		}, ReasonMissingMatchAndMismatch, func(t *testing.T, got Output) {
			assertMissing(t, got.MissingMatchBindingIDs, bindingID("base-head"))
			assertMissing(t, got.MissingMismatchBindingIDs, bindingID("base-head"))
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := fixtureInput()
			test.mutate(&input)
			got := Observe(input)
			if got.Decision != DecisionIncomplete || got.Reason != test.reason {
				t.Fatalf("got %s/%s, want INCOMPLETE/%s", got.Decision, got.Reason, test.reason)
			}
			test.check(t, got)
		})
	}
}
