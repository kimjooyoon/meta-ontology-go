package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestWideRecord10459(t *testing.T) {
	makeSource := func(count int) string {
		var b strings.Builder
		b.WriteString("package wide10459\nnamespace wide10459\nentity Integer id \"urn:gooo:type:integer\"\nentity Wide id \"urn:wide10459:record\" fields {\n")
		for i := range count {
			fmt.Fprintf(&b, "field f%02d id \"urn:wide10459:f%d\" type integer required one\n", i, i)
		}
		b.WriteString("}\nactivity Tail(Wide) -> Integer id \"urn:wide10459:tail\" computes \"let copy=input;let tail=copy.f31+179;return tail\"\n")
		return b.String()
	}
	source := makeSource(32)
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, seed := range []int{227, -307} {
		input := map[string]int{}
		for i := range 32 {
			input[fmt.Sprintf("f%02d", i)] = seed + i
		}
		raw, _ := json.Marshal(input)
		expected, _ := json.Marshal(seed + 31 + 179)
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Tail": raw}, Expected: map[string]json.RawMessage{"Tail": expected}})
	}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "wide10459.json", o)
	p, err := GenerateCompositionWithOptions(context.Background(), "wide10459.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Tail"})
	o["composition"] = p
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "wide10459.gooo", []byte(source), p, suite, "")
	o["runtime"] = r
	if err != nil || r.FinitePassed != 2 || !r.RuntimeReplayed {
		t.Fatal("32 field native behavior", err, r.FinitePassed)
	}
	if _, err := GenerateCompositionWithOptions(context.Background(), "wide33-10459.gooo", []byte(makeSource(33)), suite, CompositionOptions{EntryActivity: "Tail"}); err == nil {
		t.Fatal("33 field source must exceed the declared bound")
	}
	t.Log("32-field native2/2;field31 preserved through record copy;33-field source rejected before execution")
}
