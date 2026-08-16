package pressureshadow

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

func TestStrictWire(t *testing.T) {
	base, _ := json.Marshal(s1Input())
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
	duplicates := map[string]func(*Input){
		"pressure": func(input *Input) {
			input.Selector.Pressures = append(input.Selector.Pressures, input.Selector.Pressures[0])
		},
		"state": func(input *Input) {
			input.Selector.States = append(input.Selector.States, input.Selector.States[0])
		},
	}
	for name, mutate := range duplicates {
		input := s1Input()
		mutate(&input)
		cases["duplicate "+name], _ = json.Marshal(input)
		if _, err := CanonicalInputBytes(input); err == nil {
			t.Fatalf("CanonicalInputBytes accepted duplicate %s ID", name)
		}
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
	if CanonicalInputDigest(input) != "sha256:5903d8e3b48faa88e77064ab394c5265cb395494fed8cac5be47a727d13da825" {
		t.Fatalf("digest = %q", CanonicalInputDigest(input))
	}
	input.Selector.Pressures[0], input.Selector.Pressures[1] = input.Selector.Pressures[1], input.Selector.Pressures[0]
	input.Selector.Paths[0], input.Selector.Paths[1] = input.Selector.Paths[1], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[1] = input.PathCoverage[1], input.PathCoverage[0]
	for index := range input.PathCoverage {
		coverage := &input.PathCoverage[index].Coverage
		coverage.PressureRecords[0], coverage.PressureRecords[1] = coverage.PressureRecords[1], coverage.PressureRecords[0]
		coverage.RequiredPressureIDs = []string{"pressure/a", "pressure/b"}
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
	input.PathCoverage[0].PolicyDigest = ""
	input.PathCoverage[0].Coverage.RequiredPressureIDs = []string{"pressure/other"}
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
