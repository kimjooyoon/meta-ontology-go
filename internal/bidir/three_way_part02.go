package bidir

// ReconcileThreeWay merges base, left, and right without mutating any input.
// Presentation-only changes are ignored; one-sided absent observations do not
// delete facts unless both changed views agree on the deletion.
func ReconcileThreeWay(base, left, right Model) (ThreeWayResult, error) {
	base, left, right, conflicts := normalizeThreeWayInputs(base, left, right)
	result := ThreeWayResult{Model: base, Conflicts: conflicts}
	if len(conflicts) > 0 {
		return result, &ThreeWayConflictError{Conflicts: conflicts}
	}
	semantic, conflicts := mergeThreeWaySemantic(base, left, right)
	if len(conflicts) > 0 {
		result.Conflicts = conflicts
		return result, &ThreeWayConflictError{Conflicts: conflicts}
	}
	merged, err := base.Apply(Diff(base, semantic))
	if err != nil {
		conflict := threeWayApplyConflict(base, left, right, err)
		result.Conflicts = []ThreeWayConflict{conflict}
		return result, &ThreeWayConflictError{Conflicts: result.Conflicts}
	}
	merged.Candidates = mergeThreeWayCandidates(base, left, right)
	for _, relation := range merged.Relations {
		merged.Candidates = merged.Candidates.withoutSemanticKey(relationKey(relation.Kind, relation.Source, relation.Target))
	}
	merged.Normalize()
	result.Model = merged
	result.Delta = Diff(base, merged)
	result.Locality = LocalityForDelta(base, result.Delta)
	return result, nil
}
func normalizeThreeWayInputs(base, left, right Model) (Model, Model, Model, []ThreeWayConflict) {
	models := []Model{base.Normalized(), left.Normalized(), right.Normalized()}
	fingerprints := []string{SemanticFingerprint(models[0]), SemanticFingerprint(models[1]), SemanticFingerprint(models[2])}
	conflicts := make([]ThreeWayConflict, 0, 3)
	for index, model := range models {
		if err := model.Validate(); err != nil {
			code := ThreeWayInvalidModel
			if hasInvalidEndpoint(model) {
				code = ThreeWayEndpointInvalid
			}
			conflicts = append(conflicts, ThreeWayConflict{
				Code: code, Scope: "model", Identity: modelSide(index), Message: err.Error(),
				BaseFingerprint: fingerprints[0], LeftFingerprint: fingerprints[1], RightFingerprint: fingerprints[2],
			})
		}
	}
	sortThreeWayConflicts(conflicts)
	return models[0], models[1], models[2], conflicts
}
func hasInvalidEndpoint(model Model) bool {
	ids := make(map[ID]struct{}, len(model.Nodes))
	for _, node := range model.Nodes {
		ids[node.ID] = struct{}{}
	}
	for _, relation := range model.Relations {
		if _, exists := ids[relation.Source]; !exists {
			return true
		}
		if _, exists := ids[relation.Target]; !exists {
			return true
		}
	}
	return false
}
func modelSide(index int) string {
	return []string{"base", "left", "right"}[index]
}
