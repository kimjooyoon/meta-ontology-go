package bodycodegen

func (o *BodyValueOrigins) conditionSlice(at FillOriginRange, bodySHA string) *BodyValueOrigins {
	if o == nil {
		return nil
	}
	out := &BodyValueOrigins{Schema: o.Schema, Status: o.Status, BodySHA256: o.BodySHA256,
		Reasons: append([]string(nil), o.Reasons...)}
	if o.BodySHA256 != bodySHA {
		out.Status, out.Reasons = "UNMAPPED", []string{"BODY_HASH_DIFFERS"}
		return out
	}
	nodes := map[string]BodyValueOrigin{}
	var pending []string
	for _, node := range o.Nodes {
		nodes[node.ID] = node
		if node.Kind == "condition" && node.Decoded == at {
			pending = append(pending, node.ID)
		}
	}
	out.Roots = append([]string(nil), pending...)
	if len(pending) == 0 {
		out.Status = "PARTIAL"
		out.Reasons = joinValueOriginIDs(out.Reasons, []string{"CONDITION_ROOT_UNAVAILABLE"})
	}
	seen := map[string]bool{}
	for len(pending) > 0 {
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[id] {
			continue
		}
		seen[id] = true
		node, found := nodes[id]
		if !found {
			out.Status = "PARTIAL"
			out.Reasons = joinValueOriginIDs(out.Reasons, []string{"DEPENDENCY_NODE_UNAVAILABLE"})
			continue
		}
		pending = append(pending, node.Inputs...)
		pending = append(pending, node.Controls...)
	}
	for _, node := range o.Nodes {
		if !seen[node.ID] {
			continue
		}
		node.Inputs = append([]string(nil), node.Inputs...)
		node.Controls = append([]string(nil), node.Controls...)
		out.Nodes = append(out.Nodes, node)
	}
	return out
}

func fillFragmentRelation(fragment, condition FillOriginRange, flow *BodyValueOrigins) string {
	direct := fillRangesOverlap(fragment, condition)
	dependency := false
	if flow != nil {
		for _, node := range flow.Nodes {
			if node.Kind != "condition" && fillRangesOverlap(fragment, node.Decoded) {
				dependency = true
			}
			if node.Kind == "condition" && node.Decoded != condition && fillRangesOverlap(fragment, node.Decoded) {
				dependency = true
			}
		}
	}
	if direct && dependency {
		return "direct_and_static_dependency"
	}
	if direct {
		return "direct"
	}
	if dependency {
		return "static_dependency"
	}
	return ""
}

func fillRangesOverlap(a, b FillOriginRange) bool { return a.Start < b.End && a.End > b.Start }
