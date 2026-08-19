package resourcevector

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
	var cpu, memory, applicable, peak, work uint64
	resourceKnown, pressureKnown := true, true
	for _, id := range ids {
		command := commands[id]
		if !baselineCommandResources(command, &cpu, &memory, &work, &peak) {
			resourceKnown = false
		}
		if !baselinePressures(command, &applicable, groups) {
			pressureKnown = false
		}
	}
	if resourceKnown {
		vector.CPUCoreNS, vector.MemoryBytes, vector.PeakMemoryBytes = U64(cpu), U64(memory), U64(peak)
		vector.WorkUnits = U64(work)
	}
	if pressureKnown {
		vector.ApplicablePressures, vector.IndependentGroups = U64(applicable), U64(uint64(len(groups)))
	}
	return vector, resourceKnown, pressureKnown
}
func baselineCommandResources(command CommandRecord, cpu, memory, work, peak *uint64) bool {
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
	if *command.PeakMemoryBytes > *peak {
		*peak = *command.PeakMemoryBytes
	}
	return true
}
