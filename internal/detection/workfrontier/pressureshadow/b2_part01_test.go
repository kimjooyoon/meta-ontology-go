package pressureshadow

import (
	"encoding/json"
	"reflect"
	"testing"
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
