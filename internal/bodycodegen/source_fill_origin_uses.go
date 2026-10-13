package bodycodegen

import (
	"fmt"
	"sort"
)

// FillOriginUse records one shortest dependency path from a condition/return
// root to a candidate fragment. Paths follow use -> input, not runtime order.
type FillOriginUse struct {
	RootID       string                `json:"root_id"`
	RootKind     string                `json:"root_kind"`
	HoleID       string                `json:"hole_id"`
	FragmentKind string                `json:"fragment_kind"`
	Relation     string                `json:"relation"`
	Path         []ValueOriginPathStep `json:"path"`
}

type ValueOriginPathStep struct {
	Node BodyValueOrigin `json:"node"`
	Via  string          `json:"via,omitempty"`
}

// SourceFillUses reads only recorded source dependencies. The bounded graph may
// contain cycles. Limit is a returned-row limit, not a functional coverage claim.
func SourceFillUses(origin SourceFillOrigin, limit int) ([]FillOriginUse, bool, error) {
	if limit < 1 || limit > 128 {
		return nil, false, fmt.Errorf("origin use limit must be 1..128")
	}
	flow := origin.DataFlow
	if flow == nil {
		return nil, false, nil
	}
	if flow.Schema != "gooo/body-value-origins/v1" {
		return nil, false, fmt.Errorf("unsupported recorded value-origin schema %q", flow.Schema)
	}
	if len(flow.Nodes) > 256 || len(flow.Roots) > 256 || len(flow.Reasons) > 32 || len(origin.Fragments) > 256 || flow.BodySHA256 != origin.SelectedBodySHA256 {
		return nil, false, fmt.Errorf("recorded origin graph exceeds its scope or body hash differs")
	}
	nodes := map[string]BodyValueOrigin{}
	for _, node := range flow.Nodes {
		if node.ID == "" || nodes[node.ID].ID != "" || len(node.Inputs) > 256 || len(node.Controls) > 256 {
			return nil, false, fmt.Errorf("recorded origin node is empty, duplicated or unbounded")
		}
		nodes[node.ID] = node
	}
	for _, node := range flow.Nodes {
		for _, id := range append(append([]string(nil), node.Inputs...), node.Controls...) {
			if nodes[id].ID == "" {
				return nil, false, fmt.Errorf("recorded dependency %q is absent", id)
			}
		}
	}
	roots := append([]string(nil), flow.Roots...)
	sort.Strings(roots)
	var uses []FillOriginUse
	steps := 0
	rootSeen := map[string]bool{}
	for _, root := range roots {
		if rootSeen[root] {
			return nil, false, fmt.Errorf("recorded root is duplicated")
		}
		rootSeen[root] = true
		node, ok := nodes[root]
		if !ok {
			return nil, false, fmt.Errorf("recorded root %q is absent", root)
		}
		for _, fragment := range origin.Fragments {
			if fragment.HoleID == "" {
				continue
			}
			path := shortestOriginPath(root, fragment.SelectedDecoded, nodes)
			if len(path) == 0 {
				continue
			}
			if len(uses) >= limit || steps+len(path) > 2048 {
				return uses, true, nil
			}
			relation := "static_dependency"
			if len(path) == 1 {
				relation = "direct"
			}
			uses = append(uses, FillOriginUse{RootID: root, RootKind: node.Kind, HoleID: fragment.HoleID,
				FragmentKind: fragment.Kind, Relation: relation, Path: path})
			steps += len(path)
		}
	}
	return uses, false, nil
}

func shortestOriginPath(root string, fragment FillOriginRange, nodes map[string]BodyValueOrigin) []ValueOriginPathStep {
	type predecessor struct{ id, via string }
	previous := map[string]predecessor{root: {}}
	queue := []string{root}
	for index := 0; index < len(queue); index++ {
		id := queue[index]
		node := nodes[id]
		if fillRangesOverlap(node.Decoded, fragment) {
			var path []ValueOriginPathStep
			for current := id; current != ""; current = previous[current].id {
				n := nodes[current]
				n.Inputs = nil
				n.Controls = nil
				path = append(path, ValueOriginPathStep{Node: n, Via: previous[current].via})
			}
			for a, b := 0, len(path)-1; a < b; a, b = a+1, b-1 {
				path[a], path[b] = path[b], path[a]
			}
			return path
		}
		for _, edge := range []struct {
			via string
			ids []string
		}{{"input", node.Inputs}, {"control", node.Controls}} {
			ordered := append([]string(nil), edge.ids...)
			sort.Strings(ordered)
			for _, next := range ordered {
				if _, seen := previous[next]; !seen {
					previous[next] = predecessor{id, edge.via}
					queue = append(queue, next)
				}
			}
		}
	}
	return nil
}
