package bodycodegen

import (
	"strings"
	"testing"
)

func TestLexicalValueOrigins5011(t *testing.T) {
	const body = "let value = input + 37; if input > 11 { let value = input + 101; if value > 307 { return value } }; if value > 71 { return value }; return 0"
	flow := InspectBodyValueOrigins(body, 1)
	outer := strings.Index(body, "value > 71")
	link := flow.conditionSlice(FillOriginRange{outer, outer + len("value > 71")}, flow.BodySHA256)
	if link == nil {
		t.Fatal("missing condition")
	}
	found := false
	for _, node := range link.Nodes {
		text := body[node.Decoded.Start:node.Decoded.End]
		if text == "input + 101" {
			t.Fatal("inner shadow leaked into outer variable")
		}
		if text == "input + 37" {
			found = true
		}
	}
	if !found {
		t.Fatal("outer assignment missing", link)
	}
	t.Log("5011: lexical identity preserves outer assignment despite same-name inner variable")
}

func TestJoinedValuesAndReturns5039(t *testing.T) {
	const body = "let amount = 0; if input > 23 { amount = input + 43 } else { amount = input - 19 }; if amount > 127 { return amount }; return input"
	flow := InspectBodyValueOrigins(body, 1)
	start := strings.Index(body, "amount > 127")
	link := flow.conditionSlice(FillOriginRange{start, start + len("amount > 127")}, flow.BodySHA256)
	wants := map[string]bool{"input + 43": false, "input - 19": false, "input > 23": false}
	returns := 0
	for _, node := range flow.Nodes {
		if node.Kind == "return" {
			returns++
		}
	}
	for _, node := range link.Nodes {
		text := body[node.Decoded.Start:node.Decoded.End]
		if _, ok := wants[text]; ok {
			wants[text] = true
		}
	}
	for expression, found := range wants {
		if !found {
			t.Fatalf("missing joined origin %s", expression)
		}
	}
	if returns != 2 {
		t.Fatal("return roots", returns)
	}
	// Views are detached: mutating an inspection cannot alter the saved graph.
	for i := range link.Nodes {
		if len(link.Nodes[i].Inputs) > 0 {
			link.Nodes[i].Inputs[0] = "changed"
		}
	}
	for _, node := range flow.Nodes {
		for _, id := range node.Inputs {
			if id == "changed" {
				t.Fatal("aliased graph")
			}
		}
	}
	t.Log("5039: both branch definitions and controlling input retained; two return roots; detached graph")
}

func TestLoopAndUnsupportedValueOrigins5077(t *testing.T) {
	const body = "let total = input; for i := 0; i < 3; i++ { total = total + 29 }; if total > 211 { return total }; return input"
	flow := InspectBodyValueOrigins(body, 1)
	start := strings.Index(body, "total > 211")
	link := flow.conditionSlice(FillOriginRange{start, start + len("total > 211")}, flow.BodySHA256)
	if flow.Status != "PARTIAL" || len(link.Nodes) == 0 {
		t.Fatal("loop scope must remain explicit", flow)
	}
	found := false
	for _, node := range link.Nodes {
		if body[node.Decoded.Start:node.Decoded.End] == "total + 29" {
			found = true
		}
	}
	if !found {
		t.Fatal("loop carried origin absent")
	}
	unsupported := InspectBodyValueOrigins("let n = input; switch n { case 5: n = 109 }; return n", 1)
	if unsupported.Status != "PARTIAL" {
		t.Fatal("unknown form claimed complete")
	}
	t.Log("5077: bounded loop dependencies terminate; unsupported control forms explicitly partial")
}
