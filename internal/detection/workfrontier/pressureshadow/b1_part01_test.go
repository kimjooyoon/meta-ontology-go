package pressureshadow

import (
	"reflect"
	"strings"
	"testing"
)

const b1RawInput = `{
  "schema": "gooo/workfrontier-pressure-shadow/v1",
  "selector": {
    "schema_version": "gooo/work-frontier/v1",
    "snapshot_digest": "snapshot",
    "policy_digest": "policy",
    "registry_digest": "registry",
    "minimum_selected_pressures": 2,
    "paths": [
      {"stable_id": "path/b", "required_pressure_ids": ["pressure/b", "pressure/a"]},
      {"stable_id": "path/a", "required_pressure_ids": ["pressure/a", "pressure/b"]}
    ]
  },
  "path_coverage": [
    {"path_id": "path/b", "snapshot_digest": "snapshot", "policy_digest": "policy",
     "registry_digest": "registry", "coverage": {"schema": "gooo/workfrontier-pressure-coverage/v1",
     "requested_K": 2, "required_pressure_ids": ["pressure/b", "pressure/a"]}},
    {"path_id": "path/a", "snapshot_digest": "snapshot", "policy_digest": "policy",
     "registry_digest": "registry", "coverage": {"schema": "gooo/workfrontier-pressure-coverage/v1",
     "requested_K": 2, "required_pressure_ids": ["pressure/a", "pressure/b"]}}
  ]
}`

func TestB1PositiveBytesAndPermutation(t *testing.T) {
	want := positiveB1Result()
	got := ValidateB1Bytes([]byte(b1RawInput))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("positive result = %#v", got)
	}
	input := b1Input(t)
	input.Selector.Paths[0], input.Selector.Paths[1] = input.Selector.Paths[1], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[1] = input.PathCoverage[1], input.PathCoverage[0]
	for index := range input.PathCoverage {
		input.PathCoverage[index].Coverage.RequiredPressureIDs = []string{"pressure/a", "pressure/b"}
	}
	if replay := ValidateB1(input); !reflect.DeepEqual(replay, want) {
		t.Fatalf("permutation changed result: %#v", replay)
	}
	raw := strings.Replace(b1RawInput, `"schema":`, `"expected_reason":"PASS", "schema":`, 1)
	strict := ValidateB1Bytes([]byte(raw))
	if strict.Decision != DecisionFailClosed || strict.Reason != ReasonUpstreamFailClosed ||
		len(strict.MissingRequiredPressureIDs) != 0 {
		t.Fatalf("expected strict upstream failure: %#v", strict)
	}
	mutated := want
	mutated.Decision, mutated.Reason = DecisionUnknown, ReasonRequestedKMissing
	if CanonicalB1ResultDigest(mutated) == want.ResultDigest {
		t.Fatal("decision mutation was not bound to B1 result digest")
	}
}

type b1Vector struct {
	name           string
	mutate         func(*Input)
	decision       Decision
	reason         Reason
	missing, extra []RequiredPressureSetIssue
	missingK       []string
	mismatches     []RequestedKIssue
}
