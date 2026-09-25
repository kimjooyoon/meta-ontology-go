package bidir

import (
	"slices"
)

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
	slices.Sort(result)
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
