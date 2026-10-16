package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/viewlens"
)

func TestAdvisoryInputLayoutBoundsFresh1812(t *testing.T) {
	for _, port := range []int{-1, 16} {
		report := bodycodegen.TypedValueOrigins{Status: "RESOLVED", Inputs: []bodycodegen.TypedValueInput{
			{Declaration: &bodycodegen.TypedInputDeclaration{InputIndex: port}},
		}}
		selection := viewlens.Selection{Choice: viewlens.Choice{ViewID: viewlens.NamedInputs}, Record: &viewlens.Event{}}
		got := makeBodyViewOutput(report, selection, bodyViewOptions{}, "layout1812.gooo")
		if got.Status != "RESOLVED" || got.Projection.ViewID != viewlens.Full || got.Selection.Record != nil ||
			got.Selection.Choice.Reason != "HINT_INPUT_LAYOUT_BOUND" || len(got.Projection.InputRefs) != 0 {
			t.Fatalf("unsupported hint layout damaged canonical result: %+v", got)
		}
	}
	inputs := []bodycodegen.TypedValueInput{
		{Declaration: &bodycodegen.TypedInputDeclaration{InputIndex: 15}},
		{Declaration: &bodycodegen.TypedInputDeclaration{InputIndex: 0}},
		{Declaration: &bodycodegen.TypedInputDeclaration{InputIndex: 15}},
	}
	refs, ok := namedBodyViewRefs(bodycodegen.TypedValueOrigins{Inputs: inputs})
	if !ok || len(refs) != 2 || refs[0] != 1 || refs[1] != 0 {
		t.Fatalf("flat references not sorted/deduplicated: %v %t", refs, ok)
	}
}
