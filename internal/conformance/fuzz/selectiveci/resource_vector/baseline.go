package resourcevector

import "sort"

// EvaluateTypedConfigFullSuite is a fair baseline that runs every canonical
// command. It has its own aggregation loop and does not consume Output.
func EvaluateTypedConfigFullSuite(input Input) BaselineResult {
	return evaluateBaseline(input, input.FullCommandIDs, true, "typed-config+full-suite")
}

// EvaluatePlainDAGRetry is the plain impacted-DAG/retry baseline. For this
// fixture it sees the same selected IDs, but it remains a separate evaluator
// so an oracle cannot claim a benefit that the plain strategy already has.
func EvaluatePlainDAGRetry(input Input) BaselineResult {
	return evaluateBaseline(input, input.SelectedCommandIDs, false, "plain-dag/retry")
}

func Compare(input Input) Comparison {
	oracle := Evaluate(input)
	typed := EvaluateTypedConfigFullSuite(input)
	plain := EvaluatePlainDAGRetry(input)
	comparison := Comparison{Oracle: oracle, TypedConfigFullSuite: typed, PlainDAGRetry: plain, Finding: NoUniqueBenefit}
	if oracle.Decision != DecisionPass || typed.Decision != DecisionPass || plain.Decision != DecisionPass ||
		oracle.Selected == nil || typed.Vector == nil || plain.Vector == nil {
		return comparison
	}
	if strictlyBetter(*oracle.Selected, typed.Vector) && strictlyBetter(*oracle.Selected, plain.Vector) {
		comparison.Finding = UniqueBenefitNotEstablished
	}
	return comparison
}

func evaluateBaseline(input Input, selectedIDs []string, fullSuite bool, name string) BaselineResult {
	result := BaselineResult{Name: name, Decision: DecisionUnknown, Reason: ReasonMissingInput, FullSuite: fullSuite}
	commands, ids, selected, failure := baselineSelection(input, selectedIDs)
	if failure.reason != "" {
		result.Decision, result.Reason = failure.decision, failure.reason
		return result
	}
	vector, resourceKnown, pressureKnown := baselineResources(commands, ids)
	prov, provKnown := baselinePROV(input.Paths, selected)
	mergePartial(vector, prov)
	result.Vector = vector
	if !resourceKnown || !pressureKnown {
		result.Reason = ReasonMissingResource
		return result
	}
	if !provKnown {
		result.Reason = ReasonMissingPROV
		return result
	}
	result.Decision, result.Reason = DecisionPass, ReasonNone
	return result
}

func baselineSelection(input Input, selectedIDs []string) (map[string]CommandRecord, []string, map[string]struct{}, validationFailure) {
	if input.Commands == nil || input.Paths == nil || selectedIDs == nil || len(selectedIDs) == 0 {
		return nil, nil, nil, validationFailure{DecisionUnknown, ReasonMissingInput}
	}
	commands := map[string]CommandRecord{}
	for _, command := range input.Commands {
		if _, exists := commands[command.ID]; exists {
			return nil, nil, nil, validationFailure{DecisionFailClosed, ReasonDuplicateID}
		}
		commands[command.ID] = command
	}
	ids := sortedStrings(selectedIDs)
	selected := map[string]struct{}{}
	for _, id := range ids {
		if _, exists := selected[id]; exists {
			return nil, nil, nil, validationFailure{DecisionFailClosed, ReasonDuplicateID}
		}
		if _, exists := commands[id]; !exists {
			return nil, nil, nil, validationFailure{DecisionFailClosed, ReasonDanglingID}
		}
		selected[id] = struct{}{}
	}
	return commands, ids, selected, validationFailure{}
}

func baselineResources(commands map[string]CommandRecord, ids []string) (*PartialVector, bool, bool) {
	vector := &PartialVector{}
	groups := map[string]struct{}{}
	var cpu, memory, work, affected, applicable, peak uint64
	resourceKnown, pressureKnown := true, true
	for _, id := range ids {
		command := commands[id]
		if !baselineCommandResources(command, &cpu, &memory, &work, &affected, &peak) {
			resourceKnown = false
		}
		if !baselinePressures(command, &applicable, groups) {
			pressureKnown = false
		}
	}
	if resourceKnown {
		vector.CPUCoreNS, vector.MemoryBytes, vector.PeakMemoryBytes = U64(cpu), U64(memory), U64(peak)
		vector.WorkUnits, vector.AffectedStableIDs = U64(work), U64(affected)
	}
	if pressureKnown {
		vector.ApplicablePressures, vector.IndependentGroups = U64(applicable), U64(uint64(len(groups)))
	}
	return vector, resourceKnown, pressureKnown
}

func baselineCommandResources(command CommandRecord, cpu, memory, work, affected, peak *uint64) bool {
	if command.CPUCoreNS == nil || command.MemoryBytes == nil || command.PeakMemoryBytes == nil || command.WorkUnits == nil {
		return false
	}
	var ok bool
	*cpu, ok = add(*cpu, *command.CPUCoreNS)
	if !ok {
		return false
	}
	*memory, ok = add(*memory, *command.MemoryBytes)
	if !ok {
		return false
	}
	*work, ok = add(*work, *command.WorkUnits)
	if !ok {
		return false
	}
	*affected, ok = add(*affected, 1)
	if !ok {
		return false
	}
	if *command.PeakMemoryBytes > *peak {
		*peak = *command.PeakMemoryBytes
	}
	return true
}

func baselinePressures(command CommandRecord, applicable *uint64, groups map[string]struct{}) bool {
	if command.Pressures == nil {
		return false
	}
	known := true
	for _, pressure := range command.Pressures {
		if pressure.Applicable == nil || pressure.IndependenceGroupID == "" {
			known = false
			continue
		}
		if *pressure.Applicable {
			var ok bool
			*applicable, ok = add(*applicable, 1)
			if !ok {
				known = false
			}
			groups[pressure.IndependenceGroupID] = struct{}{}
		}
	}
	return known
}

func baselinePROV(paths []PathRecord, selected map[string]struct{}) (*PartialVector, bool) {
	vector := &PartialVector{}
	var records, finitePaths, numerator, denominator uint64
	pathCount := 0
	known := true
	for _, path := range paths {
		if _, exists := selected[path.CommandID]; !exists {
			continue
		}
		pathCount++
		if path.RecordIDs == nil || path.Finite == nil || path.ClosureNumerator == nil || path.ClosureDenominator == nil {
			known = false
			continue
		}
		var ok bool
		records, ok = add(records, uint64(len(path.RecordIDs)))
		if !ok {
			known = false
		}
		if *path.Finite {
			finitePaths, ok = add(finitePaths, 1)
			if !ok {
				known = false
			}
			numerator, ok = add(numerator, *path.ClosureNumerator)
			if !ok {
				known = false
			}
			denominator, ok = add(denominator, *path.ClosureDenominator)
			if !ok {
				known = false
			}
		}
	}
	if pathCount == 0 {
		known = false
	}
	if known {
		vector.UniquePROVRecords, vector.FinitePROVPaths = U64(records), U64(finitePaths)
		vector.ClosureNumerator, vector.ClosureDenominator = U64(numerator), U64(denominator)
	}
	return vector, known
}

func mergePartial(left, right *PartialVector) {
	left.UniquePROVRecords, left.FinitePROVPaths = right.UniquePROVRecords, right.FinitePROVPaths
	left.ClosureNumerator, left.ClosureDenominator = right.ClosureNumerator, right.ClosureDenominator
}

func strictlyBetter(oracle Vector, baseline *PartialVector) bool {
	values := []struct {
		left  uint64
		right *uint64
	}{
		{oracle.CPUCoreNS, baseline.CPUCoreNS}, {oracle.MemoryBytes, baseline.MemoryBytes},
		{oracle.PeakMemoryBytes, baseline.PeakMemoryBytes}, {oracle.WorkUnits, baseline.WorkUnits},
		{oracle.AffectedStableIDs, baseline.AffectedStableIDs}, {oracle.ApplicablePressures, baseline.ApplicablePressures},
		{oracle.IndependentGroups, baseline.IndependentGroups}, {oracle.UniquePROVRecords, baseline.UniquePROVRecords},
		{oracle.FinitePROVPaths, baseline.FinitePROVPaths}, {oracle.ClosureNumerator, baseline.ClosureNumerator},
		{oracle.ClosureDenominator, baseline.ClosureDenominator},
	}
	strict := false
	for _, value := range values {
		if value.right == nil {
			return false
		}
		if value.left > *value.right {
			return false
		}
		if value.left < *value.right {
			strict = true
		}
	}
	return strict
}

func PartialVectorValues(vector *PartialVector) []uint64 {
	if vector == nil {
		return nil
	}
	values := make([]uint64, 0, 11)
	for _, value := range []*uint64{vector.CPUCoreNS, vector.MemoryBytes, vector.PeakMemoryBytes, vector.WorkUnits, vector.AffectedStableIDs, vector.ApplicablePressures, vector.IndependentGroups, vector.UniquePROVRecords, vector.FinitePROVPaths, vector.ClosureNumerator, vector.ClosureDenominator} {
		if value != nil {
			values = append(values, *value)
		}
	}
	sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
	return values
}
