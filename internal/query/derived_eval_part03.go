package query

func findDerivedTarget(
	root, target ID, maxDepth int, selection FactSelection, wantCandidate bool,
	index map[ID][]Fact, quota *queryWorkQuota,
) (DerivedFact, bool, bool) {
	frontier := []derivedState{{node: target}}
	visited := make(map[derivedVisitKey]struct{})
	visited[derivedVisitKey{node: target}] = struct{}{}
	for head := 0; head < len(frontier); head++ {
		state := frontier[head]
		if state.depth == maxDepth {
			continue
		}
		for _, fact := range index[state.node] {
			if !quota.take() {
				return DerivedFact{}, false, false
			}
			candidate := state.candidate || fact.Status == FactCandidate
			depth := state.depth + 1
			if fact.Subject == root {
				if candidate != wantCandidate {
					continue
				}
				return newDerivedFact(
					derivedRule{id: RuleDependsOn, predicate: DerivedDependsOn},
					root, target, depth, statusForCandidate(candidate),
				), true, true
			}
			nextState := derivedState{node: fact.Subject, depth: depth, candidate: candidate}
			visitKey := derivedVisitKey{node: nextState.node, candidate: candidate}
			if _, exists := visited[visitKey]; exists {
				continue
			}
			visited[visitKey] = struct{}{}
			frontier = append(frontier, nextState)
		}
	}
	return DerivedFact{}, false, true
}
func statusForCandidate(candidate bool) FactStatus {
	if candidate {
		return FactCandidate
	}
	return FactDeterministic
}
func newDerivedFact(rule derivedRule, subject, object ID, depth int, status FactStatus) DerivedFact {
	return DerivedFact{
		Subject: subject, Predicate: rule.predicate, Object: object,
		RuleID: rule.id, Depth: depth, Status: DerivedFactStatus,
		SourceLayer: status.String(),
	}
}
func derivedKeys(rows []DerivedFact) map[derivedKey]struct{} {
	keys := make(map[derivedKey]struct{}, len(rows))
	for _, row := range rows {
		keys[derivedKey{row.Subject, row.Predicate, row.Object}] = struct{}{}
	}
	return keys
}
