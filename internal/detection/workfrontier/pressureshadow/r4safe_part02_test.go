package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"testing"
)

func TestR4SafePrecedenceAndSafeIDKills(t *testing.T) {
	cases := []r4SafeCase{
		{"same group", func(input *R4SafeInput) {
			for i := range input.PathCoverage[0].Coverage.PressureRecords {
				input.PathCoverage[0].Coverage.PressureRecords[i].IndependenceGroupID = "group-a"
			}
			rebindSafeCoverage(input)
		}, DecisionUnknown, ReasonPressureCoverageUnknown},
		{"policy floor", func(input *R4SafeInput) {
			input.PathCoverage[0].Coverage.MinimumIndependent = 1
			rebindSafeCoverage(input)
		}, DecisionFailClosed, ReasonPressureCoverageFailClosed},
		{"blank group or applicability", func(input *R4SafeInput) {
			input.PathCoverage[0].Coverage.PressureRecords[0].IndependenceGroupID = ""
			input.PathCoverage[0].Coverage.PressureRecords[0].ApplicabilityRuleID = ""
			rebindSafeCoverage(input)
		}, DecisionUnknown, ReasonPressureCoverageUnknown},
		{"missing row", func(input *R4SafeInput) {
			input.PathCoverage = nil
		}, DecisionUnknown, ReasonRequiredInputMissing},
		{"orphan row", func(input *R4SafeInput) {
			input.PathCoverage[0].PathID = "path/orphan"
		}, DecisionFailClosed, ReasonInvalidInput},
		{"stale binding", func(input *R4SafeInput) {
			input.PathCoverage[0].Coverage.PolicyDigest = "stale"
		}, DecisionUnknown, ReasonPressureCoverageUnknown},
		{"R4 fail", func(input *R4SafeInput) {
			input.R4Input.Paths[0].PrerequisiteObligationIDs = []string{"obligation/root"}
			input.R4Input.Rules = []workfrontier.R4Rule{{SCCDigest: "scc", MaxIterations: 2},
				{SCCDigest: "scc", MaxIterations: 2}}
			bindSafeR4(input)
		}, DecisionFailClosed, ReasonUpstreamFailClosed},
		{"blocked", func(input *R4SafeInput) {
			input.R4Input.Capacity.CPUCoreNS = 0
			bindSafeR4(input)
		}, DecisionPass, ReasonNone},
	}
	runR4SafeCases(t, cases)
}
func TestR4SafePressureFailurePrecedesMissing(t *testing.T) {
	input := r4SafeInput(t, 10)
	input.PathCoverage[0].Coverage.MinimumIndependent = 1
	input.R4Input.Pressures = append(input.R4Input.Pressures,
		workfrontier.Pressure{StableID: "p-global"})
	bindSafeR4(&input)
	got := ValidateR4Safe(input)
	if got.Decision != DecisionFailClosed || got.Reason != ReasonPressureCoverageFailClosed {
		t.Fatalf("mixed result = %#v", got)
	}
}
