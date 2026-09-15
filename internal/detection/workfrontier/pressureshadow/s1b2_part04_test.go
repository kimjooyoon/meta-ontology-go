package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

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
