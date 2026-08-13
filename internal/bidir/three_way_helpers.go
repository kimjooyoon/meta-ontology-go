package bidir

import "sort"

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

func unionNodeIDs(groups ...map[ID]Node) []ID {
	ids := make(map[ID]struct{})
	for _, group := range groups {
		for id := range group {
			ids[id] = struct{}{}
		}
	}
	result := make([]ID, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func mergeThreeWayRelations(base, left, right Model, fingerprints [3]string) ([]Relation, []ThreeWayConflict) {
	baseRelations, leftRelations, rightRelations := relationMap(base.Relations), relationMap(left.Relations), relationMap(right.Relations)
	keys := unionRelationKeys(baseRelations, leftRelations, rightRelations)
	merged := make([]Relation, 0, len(keys))
	var conflicts []ThreeWayConflict
	for _, key := range keys {
		state, conflict := mergeRelationState(key, threeWayRelationStateFor(baseRelations, key), threeWayRelationStateFor(leftRelations, key), threeWayRelationStateFor(rightRelations, key), fingerprints)
		if conflict != nil {
			conflicts = append(conflicts, *conflict)
			continue
		}
		if state.present {
			merged = append(merged, state.relation)
		}
	}
	return merged, conflicts
}

func mergeRelationState(key string, base, left, right threeWayRelationState, fingerprints [3]string) (threeWayRelationState, *ThreeWayConflict) {
	if relationStateEqual(left, right) {
		return chooseRelationState(left, right), nil
	}
	if relationStateEqual(left, base) {
		return preservePartialRelationAbsence(base, right), nil
	}
	if relationStateEqual(right, base) {
		return preservePartialRelationAbsence(base, left), nil
	}
	code := ThreeWayRelationAttributes
	if base.present && (!left.present || !right.present) {
		code = ThreeWayDeleteVsModify
	}
	return threeWayRelationState{}, conflictFor(code, "relation", key, fingerprints, "left and right changed the relation incompatibly")
}

func preservePartialRelationAbsence(base, observed threeWayRelationState) threeWayRelationState {
	if base.present && !observed.present {
		return base
	}
	return observed
}

func chooseRelationState(left, right threeWayRelationState) threeWayRelationState {
	if left.present {
		return left
	}
	return right
}

func relationStateEqual(left, right threeWayRelationState) bool {
	return left.present == right.present && (!left.present || relationSemanticEqual(left.relation, right.relation))
}

func threeWayRelationStateFor(relations map[string]Relation, key string) threeWayRelationState {
	relation, present := relations[key]
	return threeWayRelationState{relation: relation, present: present}
}

func unionRelationKeys(groups ...map[string]Relation) []string {
	keys := make(map[string]struct{})
	for _, group := range groups {
		for key := range group {
			keys[key] = struct{}{}
		}
	}
	result := make([]string, 0, len(keys))
	for key := range keys {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func conflictFor(code ThreeWayConflictCode, scope, identity string, fingerprints [3]string, message string) *ThreeWayConflict {
	return &ThreeWayConflict{Code: code, Scope: scope, Identity: identity, Message: message, BaseFingerprint: fingerprints[0], LeftFingerprint: fingerprints[1], RightFingerprint: fingerprints[2]}
}

func sortThreeWayConflicts(conflicts []ThreeWayConflict) {
	sort.SliceStable(conflicts, func(i, j int) bool {
		left, right := conflicts[i], conflicts[j]
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		if left.Scope != right.Scope {
			return left.Scope < right.Scope
		}
		if left.Identity != right.Identity {
			return left.Identity < right.Identity
		}
		return left.Message < right.Message
	})
}
