package pressureshadow

import (
	"reflect"
	"testing"
)

const a2aRawInput = `{
  "schema": "gooo/workfrontier-pressure-shadow/v1",
  "selector": {
    "schema_version": "gooo/work-frontier/v1",
    "snapshot_digest": "snapshot",
    "policy_digest": "policy",
    "registry_digest": "registry",
    "minimum_selected_pressures": 2,
    "paths": [
      {"stable_id": "path/b"},
      {"stable_id": "path/a"}
    ]
  },
  "path_coverage": [
    {"path_id": "path/b", "snapshot_digest": "snapshot", "policy_digest": "policy",
     "registry_digest": "registry", "coverage": {"schema": "gooo/workfrontier-pressure-coverage/v1"}},
    {"path_id": "path/a", "snapshot_digest": "snapshot", "policy_digest": "policy",
     "registry_digest": "registry", "coverage": {"schema": "gooo/workfrontier-pressure-coverage/v1"}}
  ]
}`

func TestValidatePositiveVector(t *testing.T) {
	input := a2aInput(t)
	want := Result{
		Schema: SchemaVersion, InputDigest: "sha256:44f6d8efb0b7ef85621c81853f2ac28132fdd0a2c7384424f38d65cee2c7f6e9",
		Decision: DecisionPass, Reason: ReasonNone,
		MissingPathIDs: []string{}, OrphanPathIDs: []string{},
		MissingBindingPathIDs: []string{}, BindingMismatchPathIDs: []string{},
		EnforcementEffect: EnforcementNoEffect,
		ResultDigest:      "sha256:b749379945c8bf7d89505184798b1da16980558a5120398ff31cb5d890d5930b",
		ReplayDigest:      "sha256:e3590e1b8398e8276c4d8faceeb705f3d03b1d7e402f0c3c26bf9615d15e22d7",
	}
	got := Validate(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("positive result mismatch: got=%+v", got)
	}
	input.Selector.Paths[0], input.Selector.Paths[1] = input.Selector.Paths[1], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[1] = input.PathCoverage[1], input.PathCoverage[0]
	if replay := Validate(input); !reflect.DeepEqual(replay, want) {
		t.Fatalf("permutation changed result: got=%+v", replay)
	}
}

type pathVector struct {
	name                                      string
	mutate                                    func(*Input)
	decision                                  Decision
	reason                                    Reason
	missing, orphan, missingBinding, mismatch []string
}
