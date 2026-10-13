package bodycodegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestValueOriginsNodeBudgetAndControlLinks5297(t *testing.T) {
	var source strings.Builder
	for i := range 320 {
		fmt.Fprintf(&source, "let v%d = input + %d; ", i, 197+i)
	}
	source.WriteString("return v319")
	bounded := InspectBodyValueOrigins(source.String(), 1)
	if bounded.Status != "PARTIAL" || len(bounded.Nodes) > 256 {
		t.Fatal("node budget not visible", bounded.Status, len(bounded.Nodes))
	}
	found := false
	for _, reason := range bounded.Reasons {
		if reason == "NODE_LIMIT" {
			found = true
		}
	}
	if !found {
		t.Fatal("node limit reason absent")
	}
	const body = "let chosen = input > 83; let result = 0; if chosen { result = 151 } else { result = 157 }; if result > 153 { return result }; return 0"
	flow := InspectBodyValueOrigins(body, 1)
	start := strings.Index(body, "result > 153")
	view := flow.conditionSlice(FillOriginRange{start, start + len("result > 153")}, flow.BodySHA256)
	conditions := 0
	origin := false
	for _, node := range view.Nodes {
		if node.Kind == "condition" {
			conditions++
		}
		if body[node.Decoded.Start:node.Decoded.End] == "input > 83" {
			origin = true
		}
	}
	if conditions != 2 || !origin {
		t.Fatal("control ancestry lost", view)
	}
	t.Log("5297: 320 declarations stop at bounded graph; dependency traversal retains earlier controlling condition and its variable origin")
}
