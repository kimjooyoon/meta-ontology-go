package bidir

type threeWayNodeState struct {
	node    Node
	present bool
}
type threeWayRelationState struct {
	relation Relation
	present  bool
}

func mergeThreeWayNodes(base, left, right Model, fingerprints [3]string) ([]Node, []ThreeWayConflict) {
	baseNodes, leftNodes, rightNodes := nodeMap(base.Nodes), nodeMap(left.Nodes), nodeMap(right.Nodes)
	ids := unionNodeIDs(baseNodes, leftNodes, rightNodes)
	merged := make([]Node, 0, len(ids))
	var conflicts []ThreeWayConflict
	for _, id := range ids {
		state, conflict := mergeNodeState(id, threeWayNodeStateFor(baseNodes, id), threeWayNodeStateFor(leftNodes, id), threeWayNodeStateFor(rightNodes, id), fingerprints)
		if conflict != nil {
			conflicts = append(conflicts, *conflict)
			continue
		}
		if state.present {
			merged = append(merged, state.node)
		}
	}
	return merged, conflicts
}
func mergeNodeState(id ID, base, left, right threeWayNodeState, fingerprints [3]string) (threeWayNodeState, *ThreeWayConflict) {
	if nodeStateEqual(left, right) {
		return chooseNodeState(left, right), nil
	}
	if nodeStateEqual(left, base) {
		return preservePartialNodeAbsence(base, right), nil
	}
	if nodeStateEqual(right, base) {
		return preservePartialNodeAbsence(base, left), nil
	}
	code := ThreeWaySameIdentity
	if base.present && (!left.present || !right.present) {
		code = ThreeWayDeleteVsModify
	}
	return threeWayNodeState{}, conflictFor(code, "node", string(id), fingerprints, "left and right changed the node differently")
}
func preservePartialNodeAbsence(base, observed threeWayNodeState) threeWayNodeState {
	if base.present && !observed.present {
		return base
	}
	return observed
}
func chooseNodeState(left, right threeWayNodeState) threeWayNodeState {
	if left.present {
		return left
	}
	return right
}
func nodeStateEqual(left, right threeWayNodeState) bool {
	return left.present == right.present && (!left.present || nodeSemanticEqual(left.node, right.node))
}
func threeWayNodeStateFor(nodes map[ID]Node, id ID) threeWayNodeState {
	node, present := nodes[id]
	return threeWayNodeState{node: node, present: present}
}
