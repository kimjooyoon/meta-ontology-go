package resourcevector

import (
	"encoding/json"
	"testing"
)

func TestMissingFieldsAndDuplicateRecordsAreNotZeroed(t *testing.T) {
	missing := R4F01()
	missing.Commands[0].CPUCoreNS = nil
	missing.Ceilings.Selected.CPUCoreNS = nil
	got := Evaluate(missing)
	if got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput || got.Selected != nil || got.Full != nil {
		t.Fatalf("missing resource/ceiling result = %#v", got)
	}
	missingPROV := R4F01()
	missingPROV.Paths[0].RecordIDs = nil
	prov := Evaluate(missingPROV)
	if prov.Decision != DecisionUnknown || prov.Reason != ReasonMissingPROV || prov.Selected != nil || prov.Full != nil {
		t.Fatalf("missing PROV result = %#v", prov)
	}
	baseline := EvaluatePlainDAGRetry(missingPROV)
	if baseline.Decision != DecisionUnknown || baseline.Reason != ReasonMissingPROV || baseline.Vector == nil || baseline.Vector.UniquePROVRecords != nil || baseline.Vector.FinitePROVPaths != nil {
		t.Fatalf("baseline missing PROV result = %#v", baseline)
	}
	duplicateID := R4F01()
	duplicateID.Commands = append(duplicateID.Commands, duplicateID.Commands[0])
	if got := Evaluate(duplicateID); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicateID {
		t.Fatalf("duplicate command ID = %#v", got)
	}
	duplicateRecord := R4F01()
	duplicateRecord.Paths[1].RecordIDs[0] = duplicateRecord.Paths[0].RecordIDs[0]
	if got := Evaluate(duplicateRecord); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicateRecord {
		t.Fatalf("duplicate PROV record = %#v", got)
	}
	duplicatePressure := R4F01()
	duplicatePressure.Commands[1].Pressures = append(duplicatePressure.Commands[1].Pressures, duplicatePressure.Commands[0].Pressures[0])
	if got := Evaluate(duplicatePressure); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicateID {
		t.Fatalf("duplicate pressure ID = %#v", got)
	}
	duplicatePath := R4F01()
	duplicatePath.Paths[1].ID = duplicatePath.Paths[0].ID
	if got := Evaluate(duplicatePath); got.Decision != DecisionFailClosed || got.Reason != ReasonDuplicatePath {
		t.Fatalf("duplicate path = %#v", got)
	}
}
func TestFairBaselinesAndFiniteProof(t *testing.T) {
	input := R4F01()
	comparison := Compare(input)
	if comparison.Finding != NoUniqueBenefit || comparison.TypedConfigFullSuite.Decision != DecisionPass || comparison.PlainDAGRetry.Decision != DecisionPass {
		t.Fatalf("comparison = %#v", comparison)
	}
	nonFinite := R4F01()
	nonFinite.Paths[0].Finite = Bool(false)
	got := Evaluate(nonFinite)
	if got.Decision != DecisionPass || got.ProofValid || got.Selected == nil || got.Selected.FinitePROVPaths != 2 {
		t.Fatalf("non-finite proof result = %#v", got)
	}
}
func mustClone(t *testing.T, input Input) Input {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	copy, err := DecodeInput(data)
	if err != nil {
		t.Fatal(err)
	}
	return copy
}
