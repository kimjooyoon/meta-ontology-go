package resourcevector

import (
	"reflect"
	"testing"
)

func TestPermutationAndRootRelocation(t *testing.T) {
	input := R4F01()
	permuted := mustClone(t, input)
	for left, right := 0, len(permuted.Commands)-1; left < right; left, right = left+1, right-1 {
		permuted.Commands[left], permuted.Commands[right] = permuted.Commands[right], permuted.Commands[left]
	}
	for left, right := 0, len(permuted.Paths)-1; left < right; left, right = left+1, right-1 {
		permuted.Paths[left], permuted.Paths[right] = permuted.Paths[right], permuted.Paths[left]
	}
	reverse(permuted.SelectedCommandIDs)
	reverse(permuted.FullCommandIDs)
	for index := range permuted.Commands {
		reversePressures(permuted.Commands[index].Pressures)
	}
	for index := range permuted.Paths {
		reverse(permuted.Paths[index].RecordIDs)
	}
	permuted.Root = "/relocated/r4-f-01"
	for index := range permuted.Commands {
		permuted.Commands[index].Path = relocate(permuted.Commands[index].Path, "/workspace/r4-f-01", permuted.Root)
	}
	for index := range permuted.Paths {
		permuted.Paths[index].Path = relocate(permuted.Paths[index].Path, "/workspace/r4-f-01", permuted.Root)
	}
	left, right := Evaluate(input), Evaluate(permuted)
	if left.InputDigest != right.InputDigest || left.ReplayDigest != right.ReplayDigest || !reflect.DeepEqual(left.Selected, right.Selected) || !reflect.DeepEqual(left.Full, right.Full) {
		t.Fatalf("permutation/root relocation changed replay left=%#v right=%#v", left, right)
	}
}
func TestComponentwiseCeilingsDoNotCompensate(t *testing.T) {
	input := R4F01()
	input.Ceilings.Selected.MemoryBytes = new(uint64(223))
	input.Ceilings.Selected.CPUCoreNS = new(uint64(1_000_000))
	got := Evaluate(input)
	if got.Decision != DecisionFailClosed || got.Reason != ReasonResourceLimitExceeded || !contains(got.LimitFailures, "selected:memory_bytes") {
		t.Fatalf("memory ceiling result = %#v", got)
	}
	if got.Selected == nil || got.Selected.MemoryBytes != 224 {
		t.Fatalf("selected vector lost computed value = %#v", got.Selected)
	}
}
