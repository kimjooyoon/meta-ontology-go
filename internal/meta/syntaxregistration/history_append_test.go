package syntaxregistration

import (
	"go/ast"
	"maps"
	"reflect"
	"testing"
)

func TestHistoryAppendPreservesPriorHelperOwners(t *testing.T) {
	source := historyNormalizationLayout(t, 5)
	prior, err := readHistoryOrigins(source)
	if err != nil {
		t.Fatal(err)
	}
	unchanged := make([]string, len(prior.helpers)-1)
	for index, helper := range prior.helpers[:len(unchanged)] {
		unchanged[index] = source.text(helper)
	}
	terminal := prior.helpers[len(prior.helpers)-1].Body.List[0].(*ast.SwitchStmt)
	oldCases := make([]string, len(terminal.Body.List))
	for index, clause := range terminal.Body.List {
		oldCases[index] = source.text(clause)
	}
	if err := pinPreviousSyntaxTarget(source, 174, 985); err != nil {
		t.Fatal(err)
	}
	changes, err := source.finish()
	if err != nil || len(changes) != 1 {
		t.Fatalf("append changed %d source units: %v", len(changes), err)
	}
	updated := maps.Clone(source.units)
	maps.Copy(updated, changes)
	after, err := parseSourceUnits(updated, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	next, err := readHistoryOrigins(after)
	if err != nil || next.targets[174] != 985 || len(next.helpers) != len(prior.helpers) {
		t.Fatal("new target is not bound to the terminal helper", err)
	}
	delete(next.targets, 174)
	if !reflect.DeepEqual(prior.targets, next.targets) {
		t.Fatal("prior historical targets changed")
	}
	for index, helper := range next.helpers[:len(unchanged)] {
		if after.text(helper) != unchanged[index] {
			t.Fatalf("prior helper %d was rewritten", index)
		}
	}
	terminal = next.helpers[len(next.helpers)-1].Body.List[0].(*ast.SwitchStmt)
	for index, text := range oldCases {
		if after.text(terminal.Body.List[index]) != text {
			t.Fatalf("terminal helper case %d was rewritten", index)
		}
	}
	if err := pinPreviousSyntaxTarget(after, 64, 98); err == nil || len(after.edits) != 0 {
		t.Fatal("historical version gap accepted an edit", err)
	}
}
