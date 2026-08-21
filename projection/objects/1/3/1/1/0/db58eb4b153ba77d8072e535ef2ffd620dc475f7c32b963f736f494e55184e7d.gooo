package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
	"strings"
	"testing"
)

func TestS1B1StrictBytes(t *testing.T) {
	raw := b2RawInput
	mutations := []string{
		strings.Replace(raw, `"schema":`, `"expected_label":"PASS", "schema":`, 1),
		strings.Replace(raw, `"schema":`, `"schema":"duplicate", "schema":`, 1),
		raw + `{}`,
		strings.Replace(raw, `"path/a"`, `"path a"`, 1),
	}
	for _, data := range mutations {
		got := ValidateS1B1Bytes([]byte(data))
		if got.Decision != DecisionFailClosed || got.Reason != ReasonUpstreamFailClosed || len(got.A2Observations) != 0 {
			t.Fatalf("strict result = %#v", got)
		}
	}
}
func s1b1Input() Input {
	selector := workfrontier.Input{
		SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: "selector-snapshot",
		PolicyDigest: "selector-policy", RegistryDigest: "selector-registry",
		MinimumSelectedPressures: 2,
		Pressures:                []workfrontier.Pressure{{StableID: "p-a"}, {StableID: "p-b"}, {StableID: "p-c"}},
		Paths: []workfrontier.RepairPath{
			{StableID: "path/a", RequiredPressureIDs: ids()},
			{StableID: "path/b", RequiredPressureIDs: ids()},
			{StableID: "path/c", RequiredPressureIDs: ids()},
		},
	}
	return Input{Schema: SchemaVersion, Selector: selector, PathCoverage: []PathCoverage{
		s1b1Row("path/a", selector), s1b1Row("path/b", selector), s1b1Row("path/c", selector),
	}}
}
func s1b1Row(id string, selector workfrontier.Input) PathCoverage {
	return PathCoverage{PathID: id, SnapshotDigest: selector.SnapshotDigest,
		PolicyDigest: selector.PolicyDigest, RegistryDigest: selector.RegistryDigest, Coverage: coverageInput()}
}
func coverageInput() pressurecoverage.Input {
	return pressurecoverage.Input{
		Schema:                  pressurecoverage.SchemaVersion,
		AuthoritySnapshotDigest: s1b1Snapshot, PolicyDigest: s1b1Policy,
		RegistryDigest: s1b1Registry, ToolchainOptionsDigest: s1b1Toolchain,
		RequestedK: 2, MinimumIndependent: 2,
		PressureRecords: []pressurecoverage.PressureRecord{
			{PressureID: "p-c", CategoryID: "category-c", IndependenceGroupID: "group-c", ApplicabilityRuleID: "rule-1"},
			{PressureID: "p-a", CategoryID: "category-a", IndependenceGroupID: "group-a", ApplicabilityRuleID: "rule-1"},
			{PressureID: "p-b", CategoryID: "category-b", IndependenceGroupID: "group-b", ApplicabilityRuleID: "rule-1"},
		},
		RequiredPressureIDs: ids(),
	}
}
func ids() []string { return []string{"p-c", "p-a", "p-b"} }
func setGroups(input *Input, group, pathID string) {
	row := b2Coverage(input, pathID)
	for index := range row.Coverage.PressureRecords {
		row.Coverage.PressureRecords[index].IndependenceGroupID = group
	}
	rebindCoverage(input, pathID)
}
func blankApplicability(input *Input) {
	row := b2Coverage(input, "path/c")
	row.Coverage.PressureRecords[0].ApplicabilityRuleID = ""
	rebindCoverage(input, "path/c")
}
func staleCoverage(input *Input) { input.PathCoverage[0].Coverage.PolicyDigest = "stale" }
