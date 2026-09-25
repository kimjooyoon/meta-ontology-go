package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
	"reflect"
	"testing"
)

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
