package pressureshadow

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
	"reflect"
	"testing"
)

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
	input.R4Input, _ = workfrontier.BindR4Payloads(input.R4Input)
	rebindSafeCoverage(input)
}
