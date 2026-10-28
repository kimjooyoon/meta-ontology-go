package semantic

import "fmt"

// ValidateRecordReferences checks already resolved nominal field references.
// The caller supplies the source profile's depth limit and declaration order;
// failures retain the exact field and target identities for source projection.
func ValidateRecordReferences(nodes []Node, depthLimit int) error {
	if depthLimit < 1 {
		return recordReferenceIssue("record-reference-depth", "record depth limit must be positive", "", "")
	}
	graph := recordReferenceGraph{records: map[ID]Node{}, heights: map[ID]int{}, visiting: map[ID]bool{}, limit: depthLimit}
	for _, node := range nodes {
		if node.Kind != Entity || len(node.Fields) == 0 {
			continue
		}
		if _, exists := graph.records[node.ID]; exists {
			return recordReferenceIssue("record-reference-definition", "record identity is duplicated", node.ID, node.ID)
		}
		graph.records[node.ID] = node
	}
	for _, node := range nodes {
		if _, exists := graph.records[node.ID]; exists {
			if _, err := graph.visit(node.ID, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

type recordReferenceGraph struct {
	records  map[ID]Node
	heights  map[ID]int
	visiting map[ID]bool
	limit    int
}

func (graph *recordReferenceGraph) visit(id ID, depth int) (int, error) {
	if height := graph.heights[id]; height != 0 {
		return height, nil
	}
	graph.visiting[id] = true
	height := 1
	for _, field := range graph.records[id].Fields {
		target := field.TypeRef.ID
		if target == BuiltinStringTypeID || target == BuiltinBooleanTypeID || target == BuiltinIntegerTypeID {
			continue
		}
		if _, exists := graph.records[target]; !exists {
			return 0, recordReferenceIssue("record-reference-target", "record field requires a declared record type", field.ID, target)
		}
		if graph.visiting[target] {
			return 0, recordReferenceIssue("record-reference-cycle", "cyclic record reference", field.ID, target)
		}
		if depth >= graph.limit {
			return 0, recordReferenceIssue("record-reference-depth", fmt.Sprintf("record depth exceeds %d", graph.limit), field.ID, target)
		}
		childHeight, err := graph.visit(target, depth+1)
		if err != nil {
			return 0, err
		}
		if depth+childHeight > graph.limit {
			return 0, recordReferenceIssue("record-reference-depth", fmt.Sprintf("record depth exceeds %d", graph.limit), field.ID, target)
		}
		height = max(height, childHeight+1)
	}
	graph.visiting[id] = false
	graph.heights[id] = height
	return height, nil
}

func recordReferenceIssue(code, message string, field, target ID) error {
	return &ValidationErrors{Issues: []ValidationIssue{{Code: code, Message: message, Subject: field, Object: target}}}
}
