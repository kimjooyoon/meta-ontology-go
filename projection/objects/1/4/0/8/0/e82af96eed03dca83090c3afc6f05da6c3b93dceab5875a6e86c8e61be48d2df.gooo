package pressureshadow

import (
	"testing"
)

func TestB1MappingVectors(t *testing.T) {
	runB1Vectors(t, []b1Vector{
		{name: "missing set", mutate: func(input *Input) {
			input.Selector.Paths[0].RequiredPressureIDs = append(
				input.Selector.Paths[0].RequiredPressureIDs, "pressure/c")
		}, decision: DecisionUnknown, reason: ReasonRequiredSetMissing,
			missing: []RequiredPressureSetIssue{{"path/b", []string{"pressure/c"}}}},
		{name: "extra set", mutate: func(input *Input) {
			input.PathCoverage[0].Coverage.RequiredPressureIDs = append(
				input.PathCoverage[0].Coverage.RequiredPressureIDs, "pressure/c")
		}, decision: DecisionFailClosed, reason: ReasonRequiredSetExtra,
			extra: []RequiredPressureSetIssue{{"path/b", []string{"pressure/c"}}}},
		{name: "missing K", mutate: func(input *Input) {
			input.Selector.MinimumSelectedPressures = 0
		}, decision: DecisionUnknown, reason: ReasonRequestedKMissing,
			missingK: []string{"path/a", "path/b"}},
		{name: "K mismatch", mutate: func(input *Input) {
			input.PathCoverage[1].Coverage.RequestedK = 3
		}, decision: DecisionFailClosed, reason: ReasonRequestedKMismatch,
			mismatches: []RequestedKIssue{{"path/a", 2, 3}}},
		{name: "mixed precedence", mutate: func(input *Input) {
			input.Selector.Paths[0].RequiredPressureIDs = append(
				input.Selector.Paths[0].RequiredPressureIDs, "pressure/c")
			input.PathCoverage[0].Coverage.RequiredPressureIDs = append(
				input.PathCoverage[0].Coverage.RequiredPressureIDs, "pressure/d")
			input.PathCoverage[1].Coverage.RequestedK = 3
		}, decision: DecisionFailClosed, reason: ReasonRequiredSetExtra,
			missing:    []RequiredPressureSetIssue{{"path/b", []string{"pressure/c"}}},
			extra:      []RequiredPressureSetIssue{{"path/b", []string{"pressure/d"}}},
			mismatches: []RequestedKIssue{{"path/a", 2, 3}}},
	})
}
func TestB1UpstreamVectors(t *testing.T) {
	runB1Vectors(t, []b1Vector{
		{name: "empty equal sets", mutate: func(input *Input) {
			for index := range input.Selector.Paths {
				input.Selector.Paths[index].RequiredPressureIDs = nil
			}
			for index := range input.PathCoverage {
				input.PathCoverage[index].Coverage.RequiredPressureIDs = nil
			}
		}, decision: DecisionPass, reason: ReasonNone},
		{name: "upstream unknown", mutate: func(input *Input) {
			input.Selector.SnapshotDigest = ""
		}, decision: DecisionUnknown, reason: ReasonUpstreamUnknown},
		{name: "upstream fail closed", mutate: func(input *Input) {
			input.Selector.Paths[0].StableID = "path a"
		}, decision: DecisionFailClosed, reason: ReasonUpstreamFailClosed},
	})
}
func runB1Vectors(t *testing.T, cases []b1Vector) {
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := b1Input(t)
			test.mutate(&input)
			assertB1Vector(t, ValidateB1(input), test)
		})
	}
}
