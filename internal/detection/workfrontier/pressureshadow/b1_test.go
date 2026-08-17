package pressureshadow

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
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

func TestB1MappingVectors(t *testing.T) {
	runB1Vectors(t, []b1Vector{
		{name: "missing set", mutate: func(input *Input) {
			input.Selector.Paths[0].RequiredPressureIDs = append(
				input.Selector.Paths[0].RequiredPressureIDs, "pressure/c")
		}, decision: DecisionUnknown, reason: ReasonRequiredSetMissing,
			missing: []RequiredPressureSetIssue{{"path/b", []string{"pressure/c"}}}},
		{name: "extra set", mutate: func(input *Input) {
			input.PathCoverage[0].Coverage.RequiredPressureIDs = append(
				input.PathCoverage[0].Coverage.RequiredPressureIDs, "pressure/c")
		}, decision: DecisionFailClosed, reason: ReasonRequiredSetExtra,
			extra: []RequiredPressureSetIssue{{"path/b", []string{"pressure/c"}}}},
		{name: "missing K", mutate: func(input *Input) {
			input.Selector.MinimumSelectedPressures = 0
		}, decision: DecisionUnknown, reason: ReasonRequestedKMissing,
			missingK: []string{"path/a", "path/b"}},
		{name: "K mismatch", mutate: func(input *Input) {
			input.PathCoverage[1].Coverage.RequestedK = 3
		}, decision: DecisionFailClosed, reason: ReasonRequestedKMismatch,
			mismatches: []RequestedKIssue{{"path/a", 2, 3}}},
		{name: "mixed precedence", mutate: func(input *Input) {
			input.Selector.Paths[0].RequiredPressureIDs = append(
				input.Selector.Paths[0].RequiredPressureIDs, "pressure/c")
			input.PathCoverage[0].Coverage.RequiredPressureIDs = append(
				input.PathCoverage[0].Coverage.RequiredPressureIDs, "pressure/d")
			input.PathCoverage[1].Coverage.RequestedK = 3
		}, decision: DecisionFailClosed, reason: ReasonRequiredSetExtra,
			missing:    []RequiredPressureSetIssue{{"path/b", []string{"pressure/c"}}},
			extra:      []RequiredPressureSetIssue{{"path/b", []string{"pressure/d"}}},
			mismatches: []RequestedKIssue{{"path/a", 2, 3}}},
	})
}

func TestB1UpstreamVectors(t *testing.T) {
	runB1Vectors(t, []b1Vector{
		{name: "empty equal sets", mutate: func(input *Input) {
			for index := range input.Selector.Paths {
				input.Selector.Paths[index].RequiredPressureIDs = nil
			}
			for index := range input.PathCoverage {
				input.PathCoverage[index].Coverage.RequiredPressureIDs = nil
			}
		}, decision: DecisionPass, reason: ReasonNone},
		{name: "upstream unknown", mutate: func(input *Input) {
			input.Selector.SnapshotDigest = ""
		}, decision: DecisionUnknown, reason: ReasonUpstreamUnknown},
		{name: "upstream fail closed", mutate: func(input *Input) {
			input.Selector.Paths[0].StableID = "path a"
		}, decision: DecisionFailClosed, reason: ReasonUpstreamFailClosed},
	})
}

func runB1Vectors(t *testing.T, cases []b1Vector) {
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := b1Input(t)
			test.mutate(&input)
			assertB1Vector(t, ValidateB1(input), test)
		})
	}
}

func TestB1InputDrivenKAndOpaqueFields(t *testing.T) {
	input := b1Input(t)
	input.Selector.MinimumSelectedPressures = 21
	for index := range input.PathCoverage {
		input.PathCoverage[index].Coverage.RequestedK = 21
	}
	gotK := ValidateB1(input)
	if gotK.Decision != DecisionPass || gotK.Reason != ReasonNone {
		t.Fatalf("K=21 result = %#v", gotK)
	}
	opaque := b1Input(t)
	opaque.PathCoverage[0].Coverage.MinimumIndependent = 99
	opaque.PathCoverage[0].Coverage.AuthoritySnapshotDigest = "opaque"
	opaque.PathCoverage[0].Coverage.PressureRecords = append(opaque.PathCoverage[0].Coverage.PressureRecords,
		pressurecoverage.PressureRecord{PressureID: "opaque-pressure", CategoryID: "opaque-category"})
	got := ValidateB1(opaque)
	semanticWant := positiveB1Result()
	got.InputDigest, got.UpstreamResultDigest, got.ResultDigest, got.ReplayDigest = "", "", "", ""
	semanticWant.InputDigest, semanticWant.UpstreamResultDigest = "", ""
	semanticWant.ResultDigest, semanticWant.ReplayDigest = "", ""
	if !reflect.DeepEqual(got, semanticWant) {
		t.Fatalf("opaque fields changed mapping semantics: %#v", got)
	}
}

func assertB1Vector(t *testing.T, got B1Result, want b1Vector) {
	t.Helper()
	if got.Decision != want.decision || got.Reason != want.reason ||
		!sameB1Values(got.MissingRequiredPressureIDs, want.missing) ||
		!sameB1Values(got.ExtraRequiredPressureIDs, want.extra) ||
		!sameIDs(got.MissingKPathIDs, want.missingK) ||
		!sameB1Values(got.RequestedKIssues, want.mismatches) {
		t.Fatalf("B1 vector = %#v", got)
	}
	if got.EnforcementEffect != EnforcementNoEffect || got.UpstreamResultDigest == "" || got.ResultDigest == "" {
		t.Fatalf("incomplete B1 result: %#v", got)
	}
}

func sameB1Values[T any](got, want []T) bool {
	return len(got) == 0 && len(want) == 0 || reflect.DeepEqual(got, want)
}

func b1Input(t *testing.T) Input {
	input, err := DecodeInput([]byte(b1RawInput))
	if err != nil {
		t.Fatal(err)
	}
	return input
}

func positiveB1Result() B1Result {
	return B1Result{
		Schema:                     SchemaVersion,
		InputDigest:                "sha256:862f71ef52ff1f70779a93878c6bbacef820c5c9ef80efdfec6353e6c5b84761",
		UpstreamResultDigest:       "sha256:5c805ba5d543ab1339b4f3e4cf9643b306c718fe3300e41a1387d266ee66c35a",
		Decision:                   DecisionPass,
		Reason:                     ReasonNone,
		MissingRequiredPressureIDs: []RequiredPressureSetIssue{},
		ExtraRequiredPressureIDs:   []RequiredPressureSetIssue{},
		MissingKPathIDs:            []string{},
		RequestedKIssues:           []RequestedKIssue{},
		EnforcementEffect:          EnforcementNoEffect,
		ResultDigest:               "sha256:83e420e62327163be6fcfe546ab6f5fb091bb7bfd690b8a7e8628aa3043ce078",
		ReplayDigest:               "sha256:50fade105ff7c2aa729636152f75ee393056671cfabb23236421d7262d3e91bc",
	}
}
