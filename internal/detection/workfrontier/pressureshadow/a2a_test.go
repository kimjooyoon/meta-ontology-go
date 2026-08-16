package pressureshadow

import (
	"reflect"
	"strings"
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

func TestValidateRequiredVectors(t *testing.T) {
	cases := []pathVector{
		{
			name: "empty selector paths", mutate: func(input *Input) {
				input.Selector.Paths, input.PathCoverage = nil, nil
			}, decision: DecisionUnknown, reason: ReasonRequiredInputMissing,
		},
		{
			name: "missing row", mutate: func(input *Input) {
				input.PathCoverage = input.PathCoverage[:1]
			}, decision: DecisionUnknown, reason: ReasonMissingPathCoverage,
			missing: []string{"path/a"},
		},
		{
			name: "orphan row", mutate: func(input *Input) {
				row := input.PathCoverage[0]
				row.PathID = "path/orphan"
				input.PathCoverage = append(input.PathCoverage, row)
			}, decision: DecisionFailClosed, reason: ReasonOrphanPathCoverage,
			orphan: []string{"path/orphan"},
		},
		{
			name: "blank selector tuple component", mutate: func(input *Input) {
				input.Selector.SnapshotDigest = ""
			}, decision: DecisionUnknown, reason: ReasonRequiredInputMissing,
			missingBinding: []string{"path/a", "path/b"},
		},
		{
			name: "blank row tuple component", mutate: func(input *Input) {
				input.PathCoverage[0].PolicyDigest = ""
			}, decision: DecisionUnknown, reason: ReasonRequiredInputMissing,
			missingBinding: []string{"path/b"},
		},
	}
	runPathVectors(t, cases)
}

func TestValidatePrecedenceVectors(t *testing.T) {
	cases := []pathVector{
		{
			name: "binding mismatch", mutate: func(input *Input) {
				input.PathCoverage[0].RegistryDigest = "stale"
			}, decision: DecisionUnknown, reason: ReasonBindingMismatch,
			mismatch: []string{"path/b"},
		},
		{
			name: "mixed precedence", mutate: func(input *Input) {
				input.PathCoverage = input.PathCoverage[:1]
				row := input.PathCoverage[0]
				row.PathID = "path/orphan"
				input.PathCoverage = append(input.PathCoverage, row)
				input.PathCoverage[0].PolicyDigest = ""
				input.PathCoverage[0].RegistryDigest = "stale"
			}, decision: DecisionFailClosed, reason: ReasonOrphanPathCoverage,
			missing: []string{"path/a"}, orphan: []string{"path/orphan"},
			missingBinding: []string{"path/b"}, mismatch: []string{"path/b"},
		},
		{
			name: "invalid A1 syntax", mutate: func(input *Input) {
				input.Selector.Paths[0].StableID = "path a"
			}, decision: DecisionFailClosed, reason: ReasonInvalidInput,
		},
	}
	runPathVectors(t, cases)
}

func runPathVectors(t *testing.T, cases []pathVector) {
	t.Helper()
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := a2aInput(t)
			test.mutate(&input)
			assertPathVector(t, Validate(input), test.decision, test.reason,
				test.missing, test.orphan, test.missingBinding, test.mismatch)
		})
	}
}

func TestValidateBytesRejectsInvalidWire(t *testing.T) {
	cases := []struct {
		name, raw, inputDigest, resultDigest, replayDigest string
	}{
		{
			name: "unknown key",
			raw: strings.Replace(a2aRawInput, `"schema":`,
				`"expected_reason":"PASS", "schema":`, 1),
			inputDigest:  "sha256:80daab8a5314b24fb407b75927933558f264f4a23ea51c8ec8890da4efa0a67f",
			resultDigest: "sha256:c65d1f50658330fb78b346e3d1ced7ca1c54e107b6325cc139337532b2250984",
			replayDigest: "sha256:9c75f7d8607039b95c5d9a9368537f7f4926a9be84ec81d834b2c4fe13ffb816",
		},
		{
			name: "duplicate key",
			raw: strings.Replace(a2aRawInput, `"schema":`,
				`"schema":"duplicate", "schema":`, 1),
			inputDigest:  "sha256:4b223fa4b6fafeec641faabaaa1b1c4aa044c2988c84223f20528bda2fa4b82a",
			resultDigest: "sha256:6de7a3f27c582df3027c8da8c7d9781b867ffe4c3a0c8fc7e9a6ce8cd4e7a202",
			replayDigest: "sha256:feb305dfe88becf7dce326b3bd68d5951fda147520ce827e13e1f8f312b0c77b",
		},
		{
			name: "trailing value", raw: a2aRawInput + `{}`,
			inputDigest:  "sha256:e68ff81fcb90db91e96019370d3299ef9f6a37f9cae22ab285ca60b098515c3c",
			resultDigest: "sha256:668a6a0d1c74e6095b4248cfb20c66ef61e6c5ac780c3974fa973aabf08134c9",
			replayDigest: "sha256:f3b3c7d91f25658bb0ec8e1763f5b6a79eafa3ab031e73467dffac845e56524b",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			want := Result{
				Schema: SchemaVersion, InputDigest: test.inputDigest,
				Decision: DecisionFailClosed, Reason: ReasonInvalidInput,
				MissingPathIDs: []string{}, OrphanPathIDs: []string{},
				MissingBindingPathIDs: []string{}, BindingMismatchPathIDs: []string{},
				EnforcementEffect: EnforcementNoEffect,
				ResultDigest:      test.resultDigest, ReplayDigest: test.replayDigest,
			}
			if got := ValidateBytes([]byte(test.raw)); !reflect.DeepEqual(got, want) {
				t.Fatalf("invalid wire result = %#v, want %#v", got, want)
			}
		})
	}
}

func TestResultDigestBindsExpectedLabels(t *testing.T) {
	result := Validate(a2aInput(t))
	mutated := result
	mutated.Decision, mutated.Reason = DecisionUnknown, ReasonBindingMismatch
	if mutated.InputDigest != result.InputDigest || CanonicalResultDigest(mutated) == result.ResultDigest {
		t.Fatal("expected-label mutation was not bound to result digest")
	}
}

func assertPathVector(t *testing.T, got Result, decision Decision, reason Reason,
	missing, orphan, missingBinding, mismatch []string) {
	t.Helper()
	if got.Decision != decision || got.Reason != reason {
		t.Fatalf("decision/reason = %s/%s, want %s/%s", got.Decision, got.Reason, decision, reason)
	}
	if !sameIDs(got.MissingPathIDs, missing) || !sameIDs(got.OrphanPathIDs, orphan) ||
		!sameIDs(got.MissingBindingPathIDs, missingBinding) || !sameIDs(got.BindingMismatchPathIDs, mismatch) {
		t.Fatalf("path sets = %+v", got)
	}
	if got.EnforcementEffect != EnforcementNoEffect || got.InputDigest == "" ||
		got.ResultDigest == "" || got.ReplayDigest == "" {
		t.Fatalf("incomplete digest/effect result: %+v", got)
	}
}

func sameIDs(got, want []string) bool {
	return len(got) == 0 && len(want) == 0 || reflect.DeepEqual(got, want)
}

func a2aInput(t *testing.T) Input {
	t.Helper()
	input, err := DecodeInput([]byte(a2aRawInput))
	if err != nil {
		t.Fatal(err)
	}
	return input
}
