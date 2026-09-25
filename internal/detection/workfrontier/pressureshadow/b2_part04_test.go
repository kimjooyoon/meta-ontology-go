package pressureshadow

import (
	"reflect"
	"testing"
)

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
