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
	if input.Commands == nil || input.Paths == nil || selectedIDs == nil || len(selectedIDs) == 0 {
		return result
	}
	commands := map[string]CommandRecord{}
	for _, command := range input.Commands {
		if _, exists := commands[command.ID]; exists {
			result.Decision, result.Reason = DecisionFailClosed, ReasonDuplicateID
			return result
		}
		commands[command.ID] = command
	}
	ids := sortedStrings(selectedIDs)
	selected := map[string]struct{}{}
	for _, id := range ids {
		if _, exists := selected[id]; exists {
			result.Decision, result.Reason = DecisionFailClosed, ReasonDuplicateID
			return result
		}
		if _, exists := commands[id]; !exists {
			result.Decision, result.Reason = DecisionFailClosed, ReasonDanglingID
			return result
		}
		selected[id] = struct{}{}
	}
	vector := &PartialVector{}
	groups := map[string]struct{}{}
	var resourceKnown, pressureKnown bool = true, true
	var cpu, memory, work, affected, applicable uint64
	var peak uint64
	for _, id := range ids {
		command := commands[id]
		if command.CPUCoreNS == nil || command.MemoryBytes == nil || command.PeakMemoryBytes == nil || command.WorkUnits == nil {
			resourceKnown = false
			continue
		}
		var ok bool
		cpu, ok = add(cpu, *command.CPUCoreNS)
		if !ok {
			resourceKnown = false
		}
		memory, ok = add(memory, *command.MemoryBytes)
		if !ok {
			resourceKnown = false
		}
		work, ok = add(work, *command.WorkUnits)
		if !ok {
			resourceKnown = false
		}
		if *command.PeakMemoryBytes > peak {
			peak = *command.PeakMemoryBytes
		}
		affected, ok = add(affected, 1)
		if !ok {
			resourceKnown = false
		}
		if command.Pressures == nil {
			pressureKnown = false
			continue
		}
		for _, pressure := range command.Pressures {
			if pressure.Applicable == nil || pressure.IndependenceGroupID == "" {
				pressureKnown = false
				continue
			}
			if *pressure.Applicable {
				applicable, ok = add(applicable, 1)
				if !ok {
					pressureKnown = false
				}
				groups[pressure.IndependenceGroupID] = struct{}{}
			}
		}
	}
	if resourceKnown {
		vector.CPUCoreNS, vector.MemoryBytes, vector.PeakMemoryBytes, vector.WorkUnits, vector.AffectedStableIDs = U64(cpu), U64(memory), U64(peak), U64(work), U64(affected)
	}
	if pressureKnown {
		vector.ApplicablePressures, vector.IndependentGroups = U64(applicable), U64(uint64(len(groups)))
	}

	provKnown := true
	var records, finitePaths, numerator, denominator uint64
	pathCount := 0
	for _, path := range input.Paths {
		if _, exists := selected[path.CommandID]; !exists {
			continue
		}
		pathCount++
		if path.RecordIDs == nil || path.Finite == nil || path.ClosureNumerator == nil || path.ClosureDenominator == nil {
			provKnown = false
			continue
		}
		var ok bool
		records, ok = add(records, uint64(len(path.RecordIDs)))
		if !ok {
			provKnown = false
		}
		if *path.Finite {
			finitePaths, ok = add(finitePaths, 1)
			if !ok {
				provKnown = false
			}
			numerator, ok = add(numerator, *path.ClosureNumerator)
			if !ok {
				provKnown = false
			}
			denominator, ok = add(denominator, *path.ClosureDenominator)
			if !ok {
				provKnown = false
			}
		}
	}
	if pathCount == 0 {
		provKnown = false
	}
	if provKnown {
		vector.UniquePROVRecords, vector.FinitePROVPaths = U64(records), U64(finitePaths)
		vector.ClosureNumerator, vector.ClosureDenominator = U64(numerator), U64(denominator)
	}
	result.Vector = vector
	if !resourceKnown {
		result.Reason = ReasonMissingResource
		return result
	}
	if !pressureKnown {
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
