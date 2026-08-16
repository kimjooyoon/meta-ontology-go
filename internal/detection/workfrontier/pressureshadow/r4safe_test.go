package pressureshadow

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

func TestR4SafePassAndV1Evidence(t *testing.T) {
	input := r4SafeInput(t, 10)
	raw, _ := workfrontier.EncodeR4JSON(input.R4Input)
	if baseline := workfrontier.FairBaseline(input.R4Input); !sameB1Values(baseline, []string{"path/root"}) {
		t.Fatalf("baseline selected %v", baseline)
	}
	got := ValidateR4Safe(input)
	if got.Decision != DecisionPass || got.Reason != ReasonNone || got.ExecutionAuthorized ||
		got.EnforcementEffect != EnforcementNoEffect || !sameB1Values(got.SafeSelectedIDs, []string{"path/root"}) ||
		!sameB1Values(got.SafeWorkIDs, []string{"work/root"}) || got.R4Result.Status != workfrontier.R4StatusPass {
		t.Fatalf("safe result = %#v", got)
	}
	if after, _ := workfrontier.EncodeR4JSON(input.R4Input); string(raw) != string(after) {
		t.Fatal("R4 v1 evidence changed")
	}
	relocated := r4SafeInput(t, 10)
	relocated.R4Input.RootObligationIDs = []string{"obligation/other"}
	if got := ValidateR4Safe(relocated); got.InputDigest == ValidateR4Safe(input).InputDigest {
		t.Fatal("root relocation did not change replay input")
	}
}

type r4SafeCase struct {
	name     string
	edit     func(*R4SafeInput)
	decision Decision
	reason   Reason
}

func runR4SafeCases(t *testing.T, cases []r4SafeCase) {
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := r4SafeInput(t, 10)
			test.edit(&input)
			got := ValidateR4Safe(input)
			if got.Decision != test.decision || got.Reason != test.reason ||
				got.SafeSelectedIDs != nil || got.SafeWorkIDs != nil {
				t.Fatalf("result = %#v", got)
			}
		})
	}
}

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
func TestR4SafeStrictPermutationK21AndOpaqueMutation(t *testing.T) {
	base := r4SafeInput(t, 10)
	raw, _ := CanonicalR4SafeInputBytes(base)
	for _, data := range [][]byte{
		[]byte(strings.Replace(string(raw), `"schema":`, `"expected_label":"PASS","schema":`, 1)),
		[]byte(strings.Replace(string(raw), `"schema":`, `"schema":"duplicate","schema":`, 1)),
		append(raw, []byte(`{}`)...),
		[]byte(strings.Replace(string(raw), `"path/root"`, `"path root"`, 1)),
	} {
		got := ValidateR4SafeBytes(data)
		if got.Decision != DecisionFailClosed || got.Reason != ReasonInvalidInput || !got.FullSuiteRequired {
			t.Fatalf("strict result = %#v", got)
		}
	}
	duplicateRow := base
	duplicateRow.PathCoverage = append(duplicateRow.PathCoverage, duplicateRow.PathCoverage[0])
	duplicatePressure := r4SafeInput(t, 10)
	conflict := duplicatePressure.PathCoverage[0].Coverage.PressureRecords[0]
	conflict.CategoryID = "conflicting-category"
	duplicatePressure.PathCoverage[0].Coverage.PressureRecords = append(
		duplicatePressure.PathCoverage[0].Coverage.PressureRecords, conflict)
	for _, input := range []R4SafeInput{duplicateRow, duplicatePressure} {
		if got := ValidateR4Safe(input); got.Decision != DecisionFailClosed || got.Reason != ReasonInvalidInput {
			t.Fatalf("duplicate identity result = %#v", got)
		}
	}
}
func TestR4SafePermutationK21AndOpaqueMutation(t *testing.T) {
	base := r4SafeInput(t, 10)
	permuted := r4SafeInput(t, 10)
	permuted.R4Input.Pressures[0], permuted.R4Input.Pressures[2] =
		permuted.R4Input.Pressures[2], permuted.R4Input.Pressures[0]
	row := &permuted.PathCoverage[0].Coverage
	row.RequiredPressureIDs[0], row.RequiredPressureIDs[2] = row.RequiredPressureIDs[2], row.RequiredPressureIDs[0]
	row.PressureRecords[0], row.PressureRecords[2] = row.PressureRecords[2], row.PressureRecords[0]
	if got, want := ValidateR4Safe(permuted), ValidateR4Safe(base); !reflect.DeepEqual(got, want) {
		t.Fatal("permutation changed result")
	}
	k21 := r4SafeInput(t, 10)
	for n := 4; n <= 21; n++ {
		id := fmt.Sprintf("p-%02d", n)
		k21.R4Input.Pressures = append(k21.R4Input.Pressures, workfrontier.Pressure{StableID: id})
		k21.R4Input.Paths[0].RequiredPressureIDs = append(k21.R4Input.Paths[0].RequiredPressureIDs, id)
		k21.PathCoverage[0].Coverage.RequiredPressureIDs = append(k21.PathCoverage[0].Coverage.RequiredPressureIDs, id)
		k21.PathCoverage[0].Coverage.PressureRecords = append(
			k21.PathCoverage[0].Coverage.PressureRecords,
			pressurecoverage.PressureRecord{PressureID: id, IndependenceGroupID: "group-" + id,
				CategoryID: id, ApplicabilityRuleID: "rule-1"})
	}
	k21.R4Input.MinimumSelectedPressures, k21.PathCoverage[0].Coverage.RequestedK = 21, 21
	bindSafeR4(&k21)
	if got := ValidateR4Safe(k21); got.Decision != DecisionPass || len(got.SafeSelectedIDs) != 1 {
		t.Fatalf("K=21 result = %#v", got)
	}
	mutated := r4SafeInput(t, 10)
	mutated.R4Input.Capacity.CPUCoreNS = 99
	bindSafeR4(&mutated)
	if got, want := ValidateR4Safe(mutated), ValidateR4Safe(base); got.InputDigest == want.InputDigest ||
		got.PressureResult.Decision != want.PressureResult.Decision ||
		got.PressureResult.Reason != want.PressureResult.Reason ||
		!reflect.DeepEqual(got.PressureResult.A2Observations, want.PressureResult.A2Observations) {
		t.Fatal("opaque R4 mutation changed pressure semantics")
	}
}
func r4SafeInput(t *testing.T, capacity uint64) R4SafeInput {
	input := R4SafeInput{Schema: R4SafeSchemaVersion, R4Input: workfrontier.R4Input{
		SchemaVersion: workfrontier.R4SchemaVersion, MinimumSelectedPressures: 2,
		Capacity:  workfrontier.Capacity{CPUCoreNS: capacity},
		Pressures: []workfrontier.Pressure{{StableID: "p-a"}, {StableID: "p-b"}, {StableID: "p-c"}},
		States:    []workfrontier.ObligationState{{ObligationID: "obligation/root", Status: "PENDING"}},
		Paths: []workfrontier.RepairPath{{StableID: "path/root", WorkID: "work/root",
			ObligationID: "obligation/root", ReadSet: []string{"p-a"}, WriteSet: []string{"p-a"},
			RequiredPressureIDs: ids(), CPUCoreNSUpperBound: 1}},
		RootObligationIDs: []string{"obligation/root"}, Rules: []workfrontier.R4Rule{},
	}, PathCoverage: []PathCoverage{{PathID: "path/root", Coverage: coverageInput()}}}
	bindSafeR4(&input)
	return input
}
func bindSafeR4(input *R4SafeInput) {
	bound, _ := workfrontier.BindR4Payloads(input.R4Input)
	input.R4Input = bound
	rebindSafeCoverage(input)
}
func rebindSafeCoverage(input *R4SafeInput) {
	selector := projectR4SafeInput(*input).Selector
	shadow := Input{Schema: SchemaVersion, Selector: selector, PathCoverage: input.PathCoverage}
	for i := range shadow.PathCoverage {
		shadow.PathCoverage[i].SnapshotDigest = selector.SnapshotDigest
		shadow.PathCoverage[i].PolicyDigest = selector.PolicyDigest
		shadow.PathCoverage[i].RegistryDigest = selector.RegistryDigest
	}
	for i := range shadow.PathCoverage {
		rebindCoverage(&shadow, shadow.PathCoverage[i].PathID)
	}
	input.PathCoverage = shadow.PathCoverage
}
