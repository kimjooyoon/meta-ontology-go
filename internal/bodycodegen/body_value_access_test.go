package bodycodegen

import (
	"strings"
	"testing"
)

func TestSourceAccessOccurrences8753(t *testing.T) {
	cases := []struct {
		body string
		want map[string][]string
	}{
		{"let 위치 = input.pick; return input.measurements[위치]", map[string][]string{
			"field_access": {"input.pick", "input.measurements"}, "index_operand": {"위치"},
			"index_access": {"input.measurements[위치]"}, "access_base": {"input.measurements"}}},
		{"let 창 = input.values[1:input.stop]; return 창[0]", map[string][]string{
			"slice_low": {"1"}, "slice_high": {"input.stop"}, "index_operand": {"0"},
			"slice_access": {"input.values[1:input.stop]"}, "index_access": {"창[0]"}}},
		{"return input.values[:3:5]", map[string][]string{"slice_high": {"3"}, "slice_max": {"5"}}},
		{"return []int64{101, 103, 107}[2]", map[string][]string{
			"access_base": {"[]int64{101, 103, 107}"}, "index_operand": {"2"}}},
	}
	for _, c := range cases {
		view := InspectBodyValueTransfers(c.body, 1)
		if view.Graph.Status != "MAPPED" {
			t.Fatal("fresh static access body", c.body, view.Graph)
		}
		found := map[string]map[string]bool{}
		ids := map[string]bool{}
		for _, node := range view.Graph.Nodes {
			ids[node.ID] = true
			if node.Decoded.Start < 0 || node.Decoded.End > len(c.body) {
				t.Fatal("original UTF-8 range", node)
			}
			if found[node.Kind] == nil {
				found[node.Kind] = map[string]bool{}
			}
			found[node.Kind][c.body[node.Decoded.Start:node.Decoded.End]] = true
		}
		for kind, expressions := range c.want {
			for _, expression := range expressions {
				if !found[kind][expression] {
					t.Fatal("access occurrence missing", kind, expression, view.Graph)
				}
			}
		}
		for _, node := range view.Graph.Nodes {
			for _, id := range append(append([]string(nil), node.Inputs...), node.Controls...) {
				if !ids[id] {
					t.Fatal("dangling access dependency", node, id)
				}
			}
		}
		if strings.Contains(c.body, "[:3:5]") && len(found["slice_low"]) != 0 {
			t.Fatal("omitted lower bound was assigned an invented source occurrence")
		}
		legacy := InspectBodyValueOrigins(c.body, 1)
		for _, node := range legacy.Nodes {
			if strings.Contains(node.Kind, "access") || strings.HasPrefix(node.Kind, "slice_") || node.Kind == "index_operand" {
				t.Fatal("current inspection changed retained origin graph shape", node)
			}
		}
	}
	t.Log("8753: four fresh static bodies retain fields, literal/dynamic indices, explicit bounds and Korean byte offsets; legacy graph shape unchanged; native0/model0")
}

func TestAccessNodeBudget8761(t *testing.T) {
	body := "return " + strings.Repeat("input.values[19] + ", 350) + "input.values[23]"
	view := InspectBodyValueTransfers(body, 1)
	if len(view.Graph.Nodes) > 256 || view.Graph.Status != "PARTIAL" {
		t.Fatal("bounded access graph", len(view.Graph.Nodes), view.Graph.Status)
	}
	found := false
	for _, reason := range view.Graph.Reasons {
		found = found || reason == "NODE_LIMIT"
	}
	if !found {
		t.Fatal("missing access graph limit reason", view.Graph.Reasons)
	}
	t.Log("8761: 351 fresh indexed source occurrences stop at 256 graph nodes with PARTIAL/NODE_LIMIT")
}
