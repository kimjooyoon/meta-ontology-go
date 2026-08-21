package bidir

func threeWayApplyConflict(base, left, right Model, err error) ThreeWayConflict {
	code := ThreeWayInvalidModel
	if hasInvalidEndpoint(base) || hasInvalidEndpoint(left) || hasInvalidEndpoint(right) {
		code = ThreeWayEndpointInvalid
	}
	return ThreeWayConflict{
		Code: code, Scope: "merge", Identity: "model", Message: err.Error(),
		BaseFingerprint: SemanticFingerprint(base), LeftFingerprint: SemanticFingerprint(left), RightFingerprint: SemanticFingerprint(right),
	}
}
func mergeThreeWaySemantic(base, left, right Model) (Model, []ThreeWayConflict) {
	fingerprints := [3]string{SemanticFingerprint(base), SemanticFingerprint(left), SemanticFingerprint(right)}
	nodes, conflicts := mergeThreeWayNodes(base, left, right, fingerprints)
	relations, relationConflicts := mergeThreeWayRelations(base, left, right, fingerprints)
	conflicts = append(conflicts, relationConflicts...)
	sortThreeWayConflicts(conflicts)
	return Model{Package: base.Package, Namespace: base.Namespace, Nodes: nodes, Relations: relations}.Normalized(), conflicts
}
func mergeThreeWayCandidates(base, left, right Model) FactSet {
	candidates := append([]Fact{}, base.Candidates...)
	candidates = append(candidates, left.Candidates...)
	candidates = append(candidates, right.Candidates...)
	return FactSet(candidates).Normalized()
}
