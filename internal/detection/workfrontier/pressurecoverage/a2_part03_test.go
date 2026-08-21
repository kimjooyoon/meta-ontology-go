package pressurecoverage

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEvaluatePositiveResult(t *testing.T) {
	got := Evaluate(a2Input())
	if got.Decision != DecisionPass || got.Reason != ReasonNone ||
		got.RequiredPressureCount != 3 || got.DistinctGroupCount != 3 {
		t.Fatalf("result = %#v", got)
	}
	if got.InputDigest != a2InputHash || got.ResultDigest != a2ResultHash || got.ReplayDigest != a2ReplayHash {
		t.Fatalf("digests = %#v", got)
	}
	if !reflect.DeepEqual(got.RequiredPressureIDs, []string{"p-a", "p-b", "p-c"}) ||
		!reflect.DeepEqual(got.RequiredGroupIDs, []string{"group-a", "group-b", "group-c"}) ||
		!reflect.DeepEqual(got.MissingPressureIDs, []string{}) {
		t.Fatalf("canonical sets = %#v", got)
	}
}
func TestEvaluatePrecedence(t *testing.T) {
	for _, test := range a2PrecedenceCases {
		t.Run(test.name, func(t *testing.T) {
			input := a2Input()
			test.edit(&input)
			if test.bind {
				testBind(&input)
			}
			got := Evaluate(input)
			if got.Decision != test.wantD || got.Reason != test.wantR {
				t.Fatalf("result = %#v", got)
			}
		})
	}
}
func TestEvaluateK21IsInputDriven(t *testing.T) {
	input := a2Input()
	input.RequestedK = 21
	for number := 4; number <= 21; number++ {
		id := fmt.Sprintf("p-%02d", number)
		input.PressureRecords = append(input.PressureRecords,
			PressureRecord{id, "category-" + id, "group-" + id, "rule-1"})
		input.RequiredPressureIDs = append(input.RequiredPressureIDs, id)
	}
	testBind(&input)
	got := Evaluate(input)
	if got.Decision != DecisionPass || got.RequiredPressureCount != 21 || got.DistinctGroupCount != 21 {
		t.Fatalf("K=21 result = %#v", got)
	}
}
func TestEvaluatePermutationReplay(t *testing.T) {
	base := Evaluate(a2Input())
	input := a2Input()
	input.RequiredPressureIDs[0], input.RequiredPressureIDs[2] = input.RequiredPressureIDs[2], input.RequiredPressureIDs[0]
	input.PressureRecords[0], input.PressureRecords[2] = input.PressureRecords[2], input.PressureRecords[0]
	got := Evaluate(input)
	if !reflect.DeepEqual(got, base) {
		t.Fatalf("permutation changed result: %#v != %#v", got, base)
	}
}
