package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

func s1Input() Input {
	selector := workfrontier.Input{
		SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: "snapshot", PolicyDigest: "policy",
		MinimumSelectedPressures: 2,
		Pressures:                []workfrontier.Pressure{{StableID: "pressure/b"}, {StableID: "pressure/a"}},
		States:                   []workfrontier.ObligationState{{ObligationID: "obligation/a", Status: "PENDING"}},
		Paths: []workfrontier.RepairPath{
			{StableID: "path-b", ObligationID: "obligation/a", ReadSet: []string{"b", "a"},
				WriteSet: []string{"d", "c"}, RequiredPressureIDs: []string{"pressure/b", "pressure/a"}},
			{StableID: "path-a", ObligationID: "obligation/a", RequiredPressureIDs: []string{"pressure/a", "pressure/b"}},
		},
	}
	coverage := pressurecoverage.Input{
		Schema: pressurecoverage.SchemaVersion, AuthoritySnapshotDigest: "a", PolicyDigest: "b",
		RegistryDigest: "c", ToolchainOptionsDigest: "d", RequestedK: 2, MinimumIndependent: 2,
		PressureRecords: []pressurecoverage.PressureRecord{
			{PressureID: "pressure/b", CategoryID: "category/b", IndependenceGroupID: "group/b", ApplicabilityRuleID: "rule"},
			{PressureID: "pressure/a", CategoryID: "category/a", IndependenceGroupID: "group/a", ApplicabilityRuleID: "rule"},
		},
		RequiredPressureIDs: []string{"pressure/b", "pressure/a"},
	}
	return Input{Schema: SchemaVersion, Selector: selector, PathCoverage: []PathCoverage{
		{PathID: "path-b", PolicyDigest: "policy", RegistryDigest: "registry", Coverage: coverage},
		{PathID: "path-a", PolicyDigest: "policy", RegistryDigest: "registry", Coverage: coverage},
	}}
}
