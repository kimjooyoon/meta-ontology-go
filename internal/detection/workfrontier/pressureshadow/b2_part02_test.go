package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
	"testing"
)

func TestB2RecordVectors(t *testing.T) {
	runB2Vectors(t, []b2Vector{
		{name: "record missing", mutate: func(input *Input) {
			missingRow := b2Coverage(input, "path/b")
			missingRow.Coverage.PressureRecords = missingRow.Coverage.PressureRecords[:1]
		}, decision: DecisionUnknown, reason: ReasonRequiredPressureRecordMissing,
			missingRecords: []RequiredPressureSetIssue{{"path/b", []string{"pressure/a"}}}},
		{name: "selector missing and record absent", mutate: func(input *Input) {
			input.Selector.Pressures = input.Selector.Pressures[1:]
			row := b2Coverage(input, "path/b")
			row.Coverage.PressureRecords = row.Coverage.PressureRecords[:1]
		}, decision: DecisionFailClosed, reason: ReasonUnregisteredPressureRecord,
			missingRecords:  []RequiredPressureSetIssue{{"path/b", []string{"pressure/a"}}},
			missingSelector: []RequiredPressureSetIssue{{"path/a", []string{"pressure/b"}}, {"path/b", []string{"pressure/b"}}},
			rogue: []RequiredPressureSetIssue{{"path/a", []string{"pressure/b"}},
				{"path/b", []string{"pressure/b"}}}},
		{name: "selector missing and record present", mutate: func(input *Input) {
			input.Selector.Pressures = input.Selector.Pressures[1:]
		}, decision: DecisionFailClosed, reason: ReasonUnregisteredPressureRecord,
			missingSelector: []RequiredPressureSetIssue{{"path/a", []string{"pressure/b"}}, {"path/b", []string{"pressure/b"}}},
			rogue:           []RequiredPressureSetIssue{{"path/a", []string{"pressure/b"}}, {"path/b", []string{"pressure/b"}}}},
	})
}
func TestB2RegistrationVectors(t *testing.T) {
	runB2Vectors(t, []b2Vector{
		{name: "rogue record", mutate: func(input *Input) {
			row := b2Coverage(input, "path/a")
			row.Coverage.PressureRecords = append(row.Coverage.PressureRecords,
				pressurecoverage.PressureRecord{PressureID: "pressure/rogue", CategoryID: "category/rogue"})
		}, decision: DecisionFailClosed, reason: ReasonUnregisteredPressureRecord,
			rogue: []RequiredPressureSetIssue{{"path/a", []string{"pressure/rogue"}}}},
		{name: "mixed fail and unknown", mutate: func(input *Input) {
			input.Selector.Pressures = input.Selector.Pressures[1:]
			missingRow := b2Coverage(input, "path/b")
			missingRow.Coverage.PressureRecords = missingRow.Coverage.PressureRecords[:1]
			row := b2Coverage(input, "path/a")
			row.Coverage.PressureRecords = append(row.Coverage.PressureRecords,
				pressurecoverage.PressureRecord{PressureID: "pressure/rogue", CategoryID: "category/rogue"})
		}, decision: DecisionFailClosed, reason: ReasonUnregisteredPressureRecord,
			missingRecords:  []RequiredPressureSetIssue{{"path/b", []string{"pressure/a"}}},
			missingSelector: []RequiredPressureSetIssue{{"path/a", []string{"pressure/b"}}, {"path/b", []string{"pressure/b"}}},
			rogue: []RequiredPressureSetIssue{{"path/a", []string{"pressure/b", "pressure/rogue"}},
				{"path/b", []string{"pressure/b"}}}},
		{name: "registered non-required row record", mutate: func(input *Input) {
			row := b2Coverage(input, "path/b")
			row.Coverage.PressureRecords = append(row.Coverage.PressureRecords,
				pressurecoverage.PressureRecord{PressureID: "pressure/global", CategoryID: "category/global"})
		}, decision: DecisionPass, reason: ReasonNone},
		{name: "empty mapping", mutate: func(input *Input) {
			for index := range input.Selector.Paths {
				input.Selector.Paths[index].RequiredPressureIDs = nil
			}
			for index := range input.PathCoverage {
				input.PathCoverage[index].Coverage.RequiredPressureIDs = nil
				input.PathCoverage[index].Coverage.PressureRecords = nil
			}
		}, decision: DecisionPass, reason: ReasonNone},
	})
}
