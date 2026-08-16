package pressureshadow

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

func TestStrictWire(t *testing.T) {
	base, err := json.Marshal(s1Input())
	if err != nil {
		t.Fatal(err)
	}
	nested := bytes.Replace(base, []byte(`"coverage":{`), []byte(`"coverage":{"unknown":1,`), 1)
	duplicate := s1Input()
	duplicate.PathCoverage = append(duplicate.PathCoverage, duplicate.PathCoverage[0])
	duplicateWire, _ := json.Marshal(duplicate)
	cases := map[string][]byte{
		"unknown top-level": addRootField(base, `"unknown":1`),
		"unknown nested":    nested,
		"duplicate key":     addRootField(base, `"schema":"other"`),
		"trailing value":    append(base, []byte(`{}`)...),
		"schema":            bytes.Replace(base, []byte(`"schema":"`+SchemaVersion+`"`), []byte(`"schema":"bad"`), 1),
		"invalid ID":        bytes.Replace(base, []byte(`path-a`), []byte(`path a`), 1),
		"duplicate row":     duplicateWire,
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeInput(data); err == nil {
				t.Fatal("DecodeInput accepted malformed wire")
			}
		})
	}
}

func TestCanonicalReplayUsesIndependentDigestLiteral(t *testing.T) {
	input := s1Input()
	data, err := CanonicalInputBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	if got := CanonicalInputDigest(input); got !=
		"sha256:75b9953b2a959fe851e817ea1af796fa5ba06f60fd0d7ca5408ed39477500e3a" {
		t.Fatalf("digest = %q", got)
	}
	input.Selector.Pressures[0], input.Selector.Pressures[1] = input.Selector.Pressures[1], input.Selector.Pressures[0]
	input.Selector.Paths[0], input.Selector.Paths[1] = input.Selector.Paths[1], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[1] = input.PathCoverage[1], input.PathCoverage[0]
	for index := range input.PathCoverage {
		coverage := &input.PathCoverage[index].Coverage
		coverage.PressureRecords[0], coverage.PressureRecords[1] = coverage.PressureRecords[1], coverage.PressureRecords[0]
		coverage.RequiredPressureIDs[0], coverage.RequiredPressureIDs[1] =
			coverage.RequiredPressureIDs[1], coverage.RequiredPressureIDs[0]
	}
	permuted, err := CanonicalInputBytes(input)
	if err != nil || !bytes.Equal(data, permuted) {
		t.Fatalf("permutation changed canonical bytes: %v", err)
	}
	replayedWire, _ := json.Marshal(input)
	replayed, err := DecodeInput(replayedWire)
	if err != nil || CanonicalInputDigest(replayed) != CanonicalInputDigest(input) {
		t.Fatalf("root replay changed canonical digest: %v", err)
	}
}

func TestCanonicalDataDoesNotDecideCompleteness(t *testing.T) {
	input := s1Input()
	input.PathCoverage[0].SnapshotDigest = ""
	input.PathCoverage[0].Coverage.RequiredPressureIDs = []string{"pressure/other"}
	input.Selector.SnapshotDigest = ""
	if _, err := CanonicalInputBytes(input); err != nil {
		t.Fatalf("canonical data rejected: %v", err)
	}
	input.PathCoverage = nil
	if _, err := CanonicalInputBytes(input); err != nil {
		t.Fatalf("zero-row canonical data rejected: %v", err)
	}
}

func addRootField(data []byte, field string) []byte {
	return append(append(append([]byte{}, data[:len(data)-1]...), ','), append([]byte(field), '}')...)
}

func s1Input() Input {
	selector := workfrontier.Input{
		SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: "snapshot",
		PolicyDigest: "policy", RegistryDigest: "registry", MinimumSelectedPressures: 2,
		Capacity:  workfrontier.Capacity{CPUCoreNS: 10},
		Pressures: []workfrontier.Pressure{{StableID: "pressure/b"}, {StableID: "pressure/a"}},
		States: []workfrontier.ObligationState{{
			ObligationID: "obligation/a", Status: "PENDING",
		}},
		Paths: []workfrontier.RepairPath{
			{StableID: "path-b", ObligationID: "obligation/a", ReadSet: []string{"b", "a"},
				WriteSet: []string{"d", "c"}, RequiredPressureIDs: []string{"pressure/b", "pressure/a"}},
			{StableID: "path-a", ObligationID: "obligation/a", ReadSet: []string{"a", "b"},
				WriteSet: []string{"c", "d"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"}},
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
		{PathID: "path-b", SnapshotDigest: "snapshot", PolicyDigest: "policy",
			RegistryDigest: "registry", Coverage: coverage},
		{PathID: "path-a", SnapshotDigest: "snapshot", PolicyDigest: "policy",
			RegistryDigest: "registry", Coverage: coverage},
	}}
}
