package workfrontier

func validateR4Rules(graph r4Graph, rules []R4Rule) string {
	byDigest := make(map[string][]R4Rule, len(rules))
	for _, rule := range rules {
		if rule.SCCDigest == "" {
			return R4ReasonUnboundedFrontier
		}
		byDigest[rule.SCCDigest] = append(byDigest[rule.SCCDigest], rule)
	}
	for _, entries := range byDigest {
		if len(entries) < 2 {
			continue
		}
		first := entries[0]
		for _, entry := range entries[1:] {
			if entry.MaxIterations != first.MaxIterations || entry.IterationsUsed != first.IterationsUsed {
				return R4ReasonConflictingSCCRule
			}
		}
		return R4ReasonDuplicateSCCRule
	}
	cyclic := make(map[string]r4Component)
	for _, component := range graph.components {
		if component.Cyclic {
			cyclic[component.Digest] = component
		}
	}
	if len(rules) != len(cyclic) {
		return R4ReasonUnboundedFrontier
	}
	for digest, component := range cyclic {
		entries := byDigest[digest]
		if len(entries) != 1 {
			return R4ReasonUnboundedFrontier
		}
		rule := entries[0]
		if rule.MaxIterations == 0 {
			return R4ReasonUnboundedFrontier
		}
		if rule.IterationsUsed >= rule.MaxIterations {
			return R4ReasonIterationExhausted
		}
		_ = component
	}
	for digest := range byDigest {
		if _, ok := cyclic[digest]; !ok {
			return R4ReasonUnboundedFrontier
		}
	}
	return ""
}
func r4ResultFromGraph(graph r4Graph) R4Result {
	return R4Result{
		SchemaVersion:      R4SchemaVersion,
		GraphDigest:        graph.graphDigest,
		SCCDigest:          graph.sccDigest,
		CondensationDigest: graph.condensationDigest,
		RuleDigest:         graph.ruleDigest,
		WorkReceipt:        graph.receipt,
	}
}
func r4Unknown(graph r4Graph, reason string) R4Result {
	result := r4ResultFromGraph(graph)
	for _, path := range graph.reachablePaths {
		result.Unknown = append(result.Unknown, path.StableID)
	}
	return r4UnknownWithResult(result, reason)
}
