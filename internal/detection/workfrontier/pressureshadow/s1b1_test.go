package pressureshadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

const (
	s1b1Snapshot    = "sha256:875f886774b6c7127f6b59ee5cb5facaf5825a36f708900ba63235d7db2e9b8f"
	s1b1Policy      = "sha256:a4d888b25b683488a6751d9dcc487043002be60441122fe8a87afc54b809fa49"
	s1b1Registry    = "sha256:e0d1d311e52cc85a3ff82cf7b49b299fa7dfa7bcf876eb0995e838188dd24c57"
	s1b1Toolchain   = "sha256:9f2f29d60c221e56bf389b6721e08875db5f7d6b14b30d9c25b8fc73e6908cb2"
	s1b1A2Input     = "sha256:c6dbf237e8c44a55c79836153a6880fabdfaaa3d8e872dc07e99e337cd2e4fd3"
	s1b1A2Result    = "sha256:620f8ba049427fcc6dff30f4592ee10acbdaf931d5075056b9e35128bb41afa5"
	s1b1A2Replay    = "sha256:de2a612f99485c5711b709dc804e4d289c88a1d283b9e55427691b8ce2a29a78"
	s1b1InputDigest = "sha256:fb0fc08308285457a13d6f100455f2d02e0a90faae554c2803cfdd730b9df2d9"
	s1b1Result      = "sha256:f6abb3a8c03de565f3e7e86642df3c5892209ce94431e7189b0d4ddce11dfcec"
	s1b1Replay      = "sha256:866b49b942fc917714c9a48d9eb6c6aa3c96390c4599196bb38bc9254b05daa3"
)

func TestS1B1PositiveAndPermutation(t *testing.T) {
	input := s1b1Input()
	got := ValidateS1B1(input)
	if got.Decision != DecisionPass || got.Reason != ReasonNone ||
		got.InputDigest != s1b1InputDigest || got.ResultDigest != s1b1Result || got.ReplayDigest != s1b1Replay ||
		!sameB1Values(got.PressureCoveragePassPathIDs, []string{"path/a", "path/b", "path/c"}) {
		t.Fatalf("positive result = %#v", got)
	}
	if got.A2Observations[0].Result.InputDigest != s1b1A2Input ||
		got.A2Observations[0].Result.ResultDigest != s1b1A2Result ||
		got.A2Observations[0].Result.ReplayDigest != s1b1A2Replay {
		t.Fatalf("path/a A2 result = %#v", got.A2Observations[0])
	}
	input.Selector.Paths[0], input.Selector.Paths[2] = input.Selector.Paths[2], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[2] = input.PathCoverage[2], input.PathCoverage[0]
	for index := range input.PathCoverage {
		coverage := &input.PathCoverage[index].Coverage
		coverage.RequiredPressureIDs[0], coverage.RequiredPressureIDs[2] =
			coverage.RequiredPressureIDs[2], coverage.RequiredPressureIDs[0]
		coverage.PressureRecords[0], coverage.PressureRecords[2] = coverage.PressureRecords[2], coverage.PressureRecords[0]
	}
	if replay := ValidateS1B1(input); !reflect.DeepEqual(replay, got) {
		t.Fatalf("permutation changed result: %#v", replay)
	}
}

var s1b1Unknown = ReasonPressureCoverageUnknown
var s1b1Fail = ReasonPressureCoverageFailClosed

var s1b1A2Cases = []struct {
	name  string
	edit  func(*Input)
	wantD Decision
	wantR Reason
}{
	{"same group", func(input *Input) { setGroups(input, "group-a", "path/b") }, DecisionUnknown, s1b1Unknown},
	{"blank group", func(input *Input) { setGroups(input, "", "path/b") }, DecisionUnknown, s1b1Unknown},
	{"blank applicability", blankApplicability, DecisionUnknown, s1b1Unknown},
	{"stale inner binding", staleCoverage, DecisionUnknown, s1b1Unknown},
	{"policy floor", func(input *Input) {
		input.PathCoverage[1].Coverage.MinimumIndependent = 1
		rebindCoverage(input, "path/b")
	}, DecisionFailClosed, s1b1Fail},
	{"cardinality shortfall", func(input *Input) {
		setK(input, 4)
		for _, id := range []string{"path/a", "path/b", "path/c"} {
			rebindCoverage(input, id)
		}
	}, DecisionUnknown, s1b1Unknown},
	{"empty required", func(input *Input) {
		for index := range input.Selector.Paths {
			input.Selector.Paths[index].RequiredPressureIDs = nil
		}
		for index := range input.PathCoverage {
			input.PathCoverage[index].Coverage.RequiredPressureIDs = nil
		}
		for _, id := range []string{"path/a", "path/b", "path/c"} {
			rebindCoverage(input, id)
		}
	}, DecisionUnknown, s1b1Unknown},
}

func TestS1B1A2Vectors(t *testing.T) {
	for _, test := range s1b1A2Cases {
		t.Run(test.name, func(t *testing.T) {
			input := s1b1Input()
			test.edit(&input)
			got := ValidateS1B1(input)
			if got.Decision != test.wantD || got.Reason != test.wantR || len(got.A2Observations) != 3 {
				t.Fatalf("vector result = %#v", got)
			}
		})
	}
}

func TestS1B1MixedAndOpaqueSelectorChanges(t *testing.T) {
	input := s1b1Input()
	input.PathCoverage[1].Coverage.MinimumIndependent = 1
	rebindCoverage(&input, "path/b")
	input.Selector.Paths[2].RequiredPressureIDs = nil
	input.PathCoverage[2].Coverage.RequiredPressureIDs = nil
	rebindCoverage(&input, "path/c")
	got := ValidateS1B1(input)
	if got.Decision != DecisionFailClosed || got.Reason != ReasonPressureCoverageFailClosed ||
		len(got.A2Observations) != 3 ||
		!sameB1Values(got.PressureCoverageFailPathIDs, []string{"path/b"}) ||
		!sameB1Values(got.PressureCoverageUnknownPathIDs, []string{"path/c"}) {
		t.Fatalf("mixed result = %#v", got)
	}
	base := ValidateS1B1(s1b1Input())
	input = s1b1Input()
	input.Selector.Capacity.CPUCoreNS = 99
	input.Selector.Paths[0].PolicyPriority = 99
	input.Selector.States = []workfrontier.ObligationState{{ObligationID: "state/a", Status: "BLOCKED"}}
	changed := ValidateS1B1(input)
	if changed.Decision != base.Decision || changed.Reason != base.Reason ||
		!reflect.DeepEqual(changed.A2Observations, base.A2Observations) {
		t.Fatalf("selector-only mutation changed A2 semantics: %#v", changed)
	}
}

func TestS1B1K21AndStrictUpstream(t *testing.T) {
	input := s1b1Input()
	setK21(&input)
	if got := ValidateS1B1(input); got.Decision != DecisionPass ||
		!sameB1Values(got.PressureCoveragePassPathIDs, []string{"path/a", "path/b", "path/c"}) {
		t.Fatalf("K=21 result = %#v", got)
	}
	unknown, fail := ReasonUpstreamUnknown, ReasonUpstreamFailClosed
	for _, test := range []struct {
		name string
		edit func(*Input)
		want Decision
		why  Reason
	}{
		{"upstream unknown", func(input *Input) { input.Selector.SnapshotDigest = "" }, DecisionUnknown, unknown},
		{"upstream fail", func(input *Input) { input.Selector.Paths[0].StableID = "path a" }, DecisionFailClosed, fail},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := s1b1Input()
			test.edit(&input)
			got := ValidateS1B1(input)
			if got.Decision != test.want || got.Reason != test.why || len(got.A2Observations) != 0 {
				t.Fatalf("upstream result = %#v", got)
			}
		})
	}
}

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

func setK(input *Input, k uint64) {
	input.Selector.MinimumSelectedPressures = uint32(k)
	for index := range input.PathCoverage {
		input.PathCoverage[index].Coverage.RequestedK = k
	}
}

func setK21(input *Input) {
	for number := 4; number <= 21; number++ {
		id := fmt.Sprintf("p-%02d", number)
		input.Selector.Pressures = append(input.Selector.Pressures, workfrontier.Pressure{StableID: id})
		for index := range input.Selector.Paths {
			input.Selector.Paths[index].RequiredPressureIDs = append(input.Selector.Paths[index].RequiredPressureIDs, id)
		}
		for index := range input.PathCoverage {
			coverage := &input.PathCoverage[index].Coverage
			coverage.RequiredPressureIDs = append(coverage.RequiredPressureIDs, id)
			coverage.PressureRecords = append(coverage.PressureRecords, pressurecoverage.PressureRecord{
				PressureID: id, CategoryID: "category-" + id, IndependenceGroupID: "group-" + id, ApplicabilityRuleID: "rule-1",
			})
		}
	}
	setK(input, 21)
	for _, id := range []string{"path/a", "path/b", "path/c"} {
		rebindCoverage(input, id)
	}
}

func rebindCoverage(input *Input, pathID string) {
	row := b2Coverage(input, pathID)
	unsigned := row.Coverage
	unsigned.AuthoritySnapshotDigest, unsigned.PolicyDigest = "", ""
	unsigned.RegistryDigest, unsigned.ToolchainOptionsDigest = "", ""
	unsigned.PressureRecords = append([]pressurecoverage.PressureRecord{}, unsigned.PressureRecords...)
	unsigned.RequiredPressureIDs = append([]string{}, unsigned.RequiredPressureIDs...)
	sort.Slice(unsigned.PressureRecords, func(left, right int) bool {
		return unsigned.PressureRecords[left].PressureID < unsigned.PressureRecords[right].PressureID
	})
	sort.Strings(unsigned.RequiredPressureIDs)
	data, _ := json.Marshal(unsigned)
	inputDigest := testDigest(data)
	row.Coverage.AuthoritySnapshotDigest = testRoleDigest("authority-snapshot", inputDigest)
	row.Coverage.PolicyDigest = testRoleDigest("policy", inputDigest)
	row.Coverage.RegistryDigest = testRoleDigest("registry", inputDigest)
	row.Coverage.ToolchainOptionsDigest = testRoleDigest("toolchain-options", inputDigest)
}

func testRoleDigest(role, inputDigest string) string {
	return testDigest([]byte(role + "\x00" + inputDigest))
}

func testDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
