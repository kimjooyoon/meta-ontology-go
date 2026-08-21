package selectiveci

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

func commandFrontier(input Input, commands, guards []Command) (workfrontier.Result, []selectedPath, error) {
	paths := make([]workfrontier.RepairPath, 0, len(commands)+len(guards))
	selected := make([]selectedPath, 0, len(commands)+len(guards))
	for _, command := range commands {
		selected = append(selected, selectedPath{command: command})
	}
	for _, command := range guards {
		selected = append(selected, selectedPath{command: command, guard: true})
	}
	states := make([]workfrontier.ObligationState, 0, len(selected))
	pressures := make([]workfrontier.Pressure, 0, len(selected)*2)
	for index := range selected {
		entry := &selected[index]
		entry.obligation = frontierObligation(input.Registry.Obligations, *entry)
		pressures = append(pressures, workfrontier.Pressure{StableID: entry.obligation}, workfrontier.Pressure{StableID: entry.command.ID})
		states = append(states, workfrontier.ObligationState{ObligationID: entry.obligation, Status: "PENDING"})
		paths = append(paths, frontierPath(*entry))
	}
	frontier := workfrontier.Select(frontierInput(input, pressures, states, paths))
	if frontier.Status == workfrontier.DecisionUnknown {
		return frontier, nil, failure(ReasonEvaluatorError, "work frontier returned UNKNOWN")
	}
	if frontier.Status == workfrontier.DecisionBlocked || len(frontier.SelectedIDs) != len(paths) {
		return frontier, nil, failure(ReasonFrontierBlocked, "work frontier did not select every command")
	}
	return frontier, orderSelected(frontier.SelectedIDs, selected), nil
}
func frontierObligation(bindings []ObligationBinding, entry selectedPath) string {
	if entry.guard {
		return "guard/" + entry.command.ID
	}
	return firstObligationForCommand(bindings, entry.command.ID)
}
func frontierPath(entry selectedPath) workfrontier.RepairPath {
	return workfrontier.RepairPath{StableID: entry.command.ID, ObligationID: entry.obligation, ReadSet: []string{entry.obligation}, WriteSet: []string{entry.command.ID}, RequiredPressureIDs: []string{entry.obligation, entry.command.ID}, CPUCoreNSUpperBound: entry.command.CPUWorkUnits}
}
func frontierInput(input Input, pressures []workfrontier.Pressure, states []workfrontier.ObligationState, paths []workfrontier.RepairPath) workfrontier.Input {
	return workfrontier.Input{SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: input.Head.Digest, PolicyDigest: input.Registry.PolicyDigest, RegistryDigest: input.Registry.Digest, MinimumSelectedPressures: 2, Capacity: workfrontier.Capacity{CPUCoreNS: input.CPUCapacity}, Pressures: pressures, States: states, Paths: paths}
}
func orderSelected(ids []string, selected []selectedPath) []selectedPath {
	byID := map[string]selectedPath{}
	for _, entry := range selected {
		byID[entry.command.ID] = entry
	}
	result := make([]selectedPath, 0, len(ids))
	for _, id := range ids {
		result = append(result, byID[id])
	}
	return result
}
func firstObligationForCommand(bindings []ObligationBinding, commandID string) string {
	for _, binding := range bindings {
		if slices.Contains(binding.CommandIDs, commandID) {
			return binding.ID
		}
	}
	return ""
}
