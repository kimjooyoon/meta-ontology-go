package workfrontier

import "sort"

// EvaluateR4 is the bounded detector entry point. It never consults the
// legacy selector's implicit defaults.
func EvaluateR4(input R4Input) R4Result {
	input = normalizeR4Input(input)
	if !r4RequiredInputKnown(input) {
		return r4RequiredInputResult()
	}
	if reason := validateR4Bindings(input); reason != "" {
		if reason == R4ReasonMalformedBinding {
			return r4FailClosed(r4Graph{}, reason)
		}
		return r4Unknown(r4Graph{}, reason)
	}
	graph, graphReason := buildR4Graph(input)
	if graphReason != "" {
		return r4FailClosed(graph, graphReason)
	}
	if !r4StableDeclarationsKnown(input) {
		return r4Unknown(graph, R4ReasonRequiredInputMissing)
	}
	if reason := validateR4Rules(graph, input.Rules); reason != "" {
		if reason == R4ReasonDuplicateSCCRule || reason == R4ReasonConflictingSCCRule {
			return r4FailClosed(graph, reason)
		}
		return r4Unknown(graph, reason)
	}
	legacy := r4LegacyInput(input, graph)
	result := r4ResultFromGraph(graph)
	ready, result := classifyR4Paths(legacy, graph, result)
	result = selectR4Paths(input, legacy, ready, result)
	return finishR4Result(result)
}

func r4RequiredInputResult() R4Result {
	return R4Result{
		SchemaVersion:     R4SchemaVersion,
		Status:            R4StatusUnknown,
		Reason:            R4ReasonRequiredInputMissing,
		Quality:           R4StatusUnknown,
		FullSuiteRequired: true,
	}
}

func r4LegacyInput(input R4Input, graph r4Graph) Input {
	return Input{
		SchemaVersion:            SchemaVersion,
		SnapshotDigest:           input.SnapshotDigest,
		PolicyDigest:             input.PolicyDigest,
		RegistryDigest:           input.RegistryDigest,
		MinimumSelectedPressures: input.MinimumSelectedPressures,
		Capacity:                 input.Capacity,
		Pressures:                append([]Pressure(nil), input.Pressures...),
		States:                   append([]ObligationState(nil), input.States...),
		Paths:                    append([]RepairPath(nil), graph.reachablePaths...),
	}
}

func classifyR4Paths(legacy Input, graph r4Graph, result R4Result) ([]RepairPath, R4Result) {
	indexes := buildIndexes(legacy)
	ready := make([]RepairPath, 0, len(graph.reachablePaths))
	if indexes.invalid {
		for _, path := range graph.reachablePaths {
			result.Unknown = append(result.Unknown, path.StableID)
		}
		return ready, result
	}
	for _, path := range graph.reachablePaths {
		switch classifyPath(legacy, indexes, path) {
		case pathReady:
			ready = append(ready, path)
		case pathUnknown:
			result.Unknown = append(result.Unknown, path.StableID)
		case pathBlocked:
			result.Blocked = append(result.Blocked, path.StableID)
		case pathShortfall:
			result.Shortfall = append(result.Shortfall, path.StableID)
		}
	}
	return ready, result
}

func selectR4Paths(input R4Input, legacy Input, ready []RepairPath, result R4Result) R4Result {
	sort.Slice(ready, func(i, j int) bool { return selectionKey(ready[i]) < selectionKey(ready[j]) })
	var usedCPU uint64
	selected := make([]RepairPath, 0, len(ready))
	for _, path := range ready {
		conflict := false
		for _, prior := range selected {
			result.WorkReceipt.ConflictChecks++
			if conflicts(path, prior) {
				conflict = true
				break
			}
		}
		if conflict || path.CPUCoreNSUpperBound > input.Capacity.CPUCoreNS-usedCPU {
			result.Blocked = append(result.Blocked, path.StableID)
			continue
		}
		selected = append(selected, path)
		usedCPU += path.CPUCoreNSUpperBound
		workID := WorkIDFor(legacy, path)
		result.Selected = append(result.Selected, workID)
		result.SelectedIDs = append(result.SelectedIDs, path.StableID)
		result.WorkIDs = append(result.WorkIDs, workID)
	}
	return result
}

func finishR4Result(result R4Result) R4Result {
	if len(result.Unknown) != 0 {
		return r4UnknownWithResult(result, R4ReasonRequiredInputMissing)
	}
	if len(result.Shortfall) != 0 {
		return r4UnknownWithResult(result, R4ReasonSelectionShortfall)
	}
	if len(result.Selected) == 0 && len(result.Blocked) != 0 {
		result.Status = R4StatusBlocked
		result.Quality = R4StatusBlocked
	} else {
		result.Status = R4StatusPass
		result.Quality = "MAXIMAL"
	}
	result.Reason = R4ReasonNone
	return normalizeR4Result(result)
}

func r4RequiredInputKnown(input R4Input) bool {
	return input.SchemaVersion == R4SchemaVersion && input.SnapshotDigest != "" &&
		input.SnapshotPayload != "" && input.PolicyDigest != "" &&
		input.PolicyPayload != "" && input.RegistryDigest != "" &&
		input.RegistryPayload != "" &&
		input.MinimumSelectedPressures >= 2 && input.Pressures != nil &&
		input.States != nil && input.Paths != nil && input.RootObligationIDs != nil &&
		len(input.RootObligationIDs) != 0 && input.Rules != nil
}

func r4StableDeclarationsKnown(input R4Input) bool {
	for _, pressure := range input.Pressures {
		if pressure.StableID == "" {
			return false
		}
	}
	for _, state := range input.States {
		if state.ObligationID == "" || state.Status == "" {
			return false
		}
	}
	for _, path := range input.Paths {
		if path.StableID == "" || path.ObligationID == "" {
			return false
		}
	}
	return true
}

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

func r4UnknownWithResult(result R4Result, reason string) R4Result {
	result.Status = R4StatusUnknown
	result.Reason = reason
	result.Quality = R4StatusUnknown
	result.FullSuiteRequired = true
	result.Selected = nil
	result.SelectedIDs = nil
	result.WorkIDs = nil
	return normalizeR4Result(result)
}

func r4FailClosed(graph r4Graph, reason string) R4Result {
	result := r4Unknown(graph, reason)
	result.Status = R4StatusFailClosed
	result.Quality = R4StatusFailClosed
	return result
}
