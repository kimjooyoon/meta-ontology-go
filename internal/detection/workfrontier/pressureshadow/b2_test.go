package pressureshadow

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

const b2RawInput = `{
  "schema": "gooo/workfrontier-pressure-shadow/v1",
  "selector": {
    "schema_version": "gooo/work-frontier/v1",
    "snapshot_digest": "snapshot",
    "policy_digest": "policy",
    "registry_digest": "registry",
    "minimum_selected_pressures": 2,
    "pressures": [
      {"stable_id": "pressure/b"}, {"stable_id": "pressure/a"},
      {"stable_id": "pressure/global"}
    ],
    "paths": [
      {"stable_id": "path/b", "required_pressure_ids": ["pressure/b", "pressure/a"]},
      {"stable_id": "path/a", "required_pressure_ids": ["pressure/a", "pressure/b"]}
    ]
  },
  "path_coverage": [
    {"path_id": "path/b", "snapshot_digest": "snapshot", "policy_digest": "policy",
     "registry_digest": "registry", "coverage": {"schema": "gooo/workfrontier-pressure-coverage/v1",
     "requested_K": 2, "required_pressure_ids": ["pressure/b", "pressure/a"],
     "pressure_records": [{"pressure_id": "pressure/b", "category_id": "category/b"},
     {"pressure_id": "pressure/a", "category_id": "category/a"}]}},
    {"path_id": "path/a", "snapshot_digest": "snapshot", "policy_digest": "policy",
     "registry_digest": "registry", "coverage": {"schema": "gooo/workfrontier-pressure-coverage/v1",
     "requested_K": 2, "required_pressure_ids": ["pressure/a", "pressure/b"],
     "pressure_records": [{"pressure_id": "pressure/a", "category_id": "category/a"},
     {"pressure_id": "pressure/b", "category_id": "category/b"}]}}
  ]
}`

func TestB2PositivePermutationAndRootReplay(t *testing.T) {
	want := positiveB2Result()
	if got := ValidateB2Bytes([]byte(b2RawInput)); !reflect.DeepEqual(got, want) {
		t.Fatalf("positive result = %#v", got)
	}
	input := b2Input(t)
	input.Selector.Paths[0], input.Selector.Paths[1] = input.Selector.Paths[1], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[1] = input.PathCoverage[1], input.PathCoverage[0]
	for index := range input.PathCoverage {
		coverage := &input.PathCoverage[index].Coverage
		coverage.RequiredPressureIDs = []string{"pressure/a", "pressure/b"}
		coverage.PressureRecords[0], coverage.PressureRecords[1] = coverage.PressureRecords[1], coverage.PressureRecords[0]
	}
	if got := ValidateB2(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("permutation changed result = %#v", got)
	}
	wire, err := json.Marshal(input)
	if err != nil || !reflect.DeepEqual(ValidateB2Bytes(wire), want) {
		t.Fatalf("root replay changed result: %v", err)
	}
}

type b2Vector struct {
	name                                   string
	mutate                                 func(*Input)
	decision                               Decision
	reason                                 Reason
	missingRecords, missingSelector, rogue []RequiredPressureSetIssue
}

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

func TestB2UpstreamPropagationAndStrictWire(t *testing.T) {
	for _, test := range []struct {
		name     string
		mutate   func(*Input)
		decision Decision
		reason   Reason
	}{
		{name: "unknown", mutate: func(input *Input) {
			input.Selector.SnapshotDigest = ""
		}, decision: DecisionUnknown, reason: ReasonUpstreamUnknown},
		{name: "fail closed", mutate: func(input *Input) {
			input.Selector.Paths[0].StableID = "path a"
		}, decision: DecisionFailClosed, reason: ReasonUpstreamFailClosed},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := b2Input(t)
			test.mutate(&input)
			upstream := ValidateB1(input)
			got := ValidateB2(input)
			if got.Decision != test.decision || got.Reason != test.reason ||
				got.InputDigest != upstream.InputDigest || got.UpstreamResultDigest != upstream.ResultDigest ||
				len(got.MissingRequiredPressureRecordIDs) != 0 || len(got.MissingSelectorPressureIDs) != 0 ||
				len(got.UnregisteredPressureRecordIDs) != 0 {
				t.Fatalf("upstream propagation = %#v", got)
			}
		})
	}
	for _, raw := range []string{
		strings.Replace(b2RawInput, `"schema":`, `"expected_label":"PASS", "schema":`, 1),
		strings.Replace(b2RawInput, `"schema":`, `"schema":"duplicate", "schema":`, 1),
		b2RawInput + `{}`,
	} {
		got := ValidateB2Bytes([]byte(raw))
		if got.Decision != DecisionFailClosed || got.Reason != ReasonUpstreamFailClosed ||
			len(got.MissingRequiredPressureRecordIDs) != 0 {
			t.Fatalf("strict wire result = %#v", got)
		}
	}
}

func TestB2OpaqueFieldsKeepSemanticResult(t *testing.T) {
	want := positiveB2Result()
	input := b2Input(t)
	input.Selector.MinimumSelectedPressures = 21
	for index := range input.PathCoverage {
		coverage := &input.PathCoverage[index].Coverage
		coverage.RequestedK = 21
		coverage.MinimumIndependent = 99
		coverage.AuthoritySnapshotDigest = "opaque-snapshot"
		coverage.PolicyDigest = "opaque-policy"
		coverage.RegistryDigest = "opaque-registry"
		coverage.ToolchainOptionsDigest = "opaque-toolchain"
		for record := range coverage.PressureRecords {
			coverage.PressureRecords[record].CategoryID = "opaque-category"
			coverage.PressureRecords[record].IndependenceGroupID = "opaque-group"
			coverage.PressureRecords[record].ApplicabilityRuleID = "opaque-rule"
		}
	}
	got := ValidateB2(input)
	if got.InputDigest == want.InputDigest || got.UpstreamResultDigest == want.UpstreamResultDigest ||
		got.ResultDigest == want.ResultDigest || got.ReplayDigest == want.ReplayDigest {
		t.Fatalf("opaque fields did not change bound digests: %#v", got)
	}
	for _, result := range []*B2Result{&got, &want} {
		result.InputDigest, result.UpstreamResultDigest = "", ""
		result.ResultDigest, result.ReplayDigest = "", ""
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("opaque fields changed semantic result: %#v", got)
	}
}

func runB2Vectors(t *testing.T, cases []b2Vector) {
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := b2Input(t)
			test.mutate(&input)
			got := ValidateB2(input)
			if got.Decision != test.decision || got.Reason != test.reason ||
				!sameB1Values(got.MissingRequiredPressureRecordIDs, test.missingRecords) ||
				!sameB1Values(got.MissingSelectorPressureIDs, test.missingSelector) ||
				!sameB1Values(got.UnregisteredPressureRecordIDs, test.rogue) {
				t.Fatalf("B2 vector = %#v", got)
			}
		})
	}
}

func b2Input(t *testing.T) Input {
	input, err := DecodeInput([]byte(b2RawInput))
	if err != nil {
		t.Fatal(err)
	}
	return input
}

func b2Coverage(input *Input, pathID string) *PathCoverage {
	for index := range input.PathCoverage {
		if input.PathCoverage[index].PathID == pathID {
			return &input.PathCoverage[index]
		}
	}
	panic("missing test coverage row")
}

func positiveB2Result() B2Result {
	return B2Result{
		Schema:                           SchemaVersion,
		InputDigest:                      "sha256:a2517624fa346a2dad078c9b13d5f66ac6ca78b6ff75260003d82d0128cffc92",
		UpstreamResultDigest:             "sha256:cff0ff1a908890c9ab568d19a8d4a3c2a401781b1ae5aaf4e5effaf079a443f3",
		Decision:                         DecisionPass,
		Reason:                           ReasonNone,
		MissingRequiredPressureRecordIDs: []RequiredPressureSetIssue{},
		MissingSelectorPressureIDs:       []RequiredPressureSetIssue{},
		UnregisteredPressureRecordIDs:    []RequiredPressureSetIssue{},
		EnforcementEffect:                EnforcementNoEffect,
		ResultDigest:                     "sha256:16431ed57dd4f71911a15aa7f34ce13e326848a89d0dc28c0a077f62e5171c7d",
		ReplayDigest:                     "sha256:ade5d3c5ba3ad18be9aa95d207f5df525e31b446ec61f09f1fb604d19675380e",
	}
}
