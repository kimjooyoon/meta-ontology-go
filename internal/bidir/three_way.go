package bidir

import (
	"fmt"
	"strings"
)

// ThreeWayConflictCode identifies a deterministic semantic merge rejection.
type ThreeWayConflictCode string

const (
	ThreeWayInvalidModel       ThreeWayConflictCode = "invalid-model"
	ThreeWayEndpointInvalid    ThreeWayConflictCode = "endpoint-invalid"
	ThreeWaySameIdentity       ThreeWayConflictCode = "same-identity-changed-differently"
	ThreeWayDeleteVsModify     ThreeWayConflictCode = "delete-vs-modify"
	ThreeWayRelationAttributes ThreeWayConflictCode = "incompatible-relation-attributes"
)

// ThreeWayConflict is a stable, presentation-insensitive merge diagnostic.
type ThreeWayConflict struct {
	Code             ThreeWayConflictCode
	Scope            string
	Identity         string
	Message          string
	BaseFingerprint  string
	LeftFingerprint  string
	RightFingerprint string
}

// ThreeWayConflictError exposes all conflicts in deterministic order.
type ThreeWayConflictError struct {
	Conflicts []ThreeWayConflict
}

func (e *ThreeWayConflictError) Error() string {
	if e == nil || len(e.Conflicts) == 0 {
		return "three-way reconciliation failed"
	}
	parts := make([]string, len(e.Conflicts))
	for index, conflict := range e.Conflicts {
		parts[index] = fmt.Sprintf("%s[%s:%s]: %s", conflict.Code, conflict.Scope, conflict.Identity, conflict.Message)
	}
	return strings.Join(parts, "; ")
}

// ThreeWayResult is the normalized output of a successful three-way merge.
// On conflict, Model is the detached normalized base and Conflicts is filled.
type ThreeWayResult struct {
	Model     Model
	Delta     Delta
	Locality  Locality
	Conflicts []ThreeWayConflict
}

// Succeeded reports whether the merge produced no structured conflicts.
func (r ThreeWayResult) Succeeded() bool { return len(r.Conflicts) == 0 }

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
