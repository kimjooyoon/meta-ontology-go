package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCollectionMapNestedInputOrigins(t *testing.T) {
	source := strings.Replace(nativeMapSource, "map_values(input.values, Quote)", "map_values(map_values(input.values,Quote),Quote)", 1)
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":["a","a","한글"]}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`"[[a]][[a]][[한글]]"`)}}}}
	p, err := GenerateCompositionWithOptions(context.Background(), "nested.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Read"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "nested.gooo", []byte(source), p, suite, "")
	if err != nil || r.FinitePassed != 1 || !r.RuntimeReplayed {
		t.Fatal(err, r.Failure)
	}
	if r.ReadObservation.ResolvedReads != 3 || r.ReadObservation.ConstructedReads != 6 || r.ReadObservation.RecordedConstructions != 2 || r.ReadObservation.Status != "COMPLETE" {
		t.Fatal(r.ReadObservation)
	}
	trace := r.Traces[0]
	for i := range 3 {
		inner, outer := trace.Reads[3+i].Construction, trace.Reads[6+i].Construction
		if inner == nil || outer == nil || inner.Instance != 0 || outer.Instance != 1 || inner.Index != i || outer.Index != i {
			t.Fatal("nested creation order", inner, outer)
		}
	}
	if err := validateRecordedNativeOrigins(r); err != nil {
		t.Fatal(err)
	}
}
