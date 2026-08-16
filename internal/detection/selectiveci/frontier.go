package selectiveci

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

func selectedCommands(registry Registry, obligations []string) ([]Command, []Command, error) {
	commands := commandIndex(registry.Commands)
	guards := append([]Command(nil), registry.GlobalGuardCommands...)
	bindings := bindingIndex(registry.Obligations)
	selected := map[string]Command{}
	for _, obligationID := range obligations {
		binding, ok := bindings[obligationID]
		if !ok {
			return nil, nil, failure(ReasonMissingBinding, "required obligation is not bound")
		}
		for _, commandID := range binding.CommandIDs {
			command, ok := commands[commandID]
			if !ok {
				return nil, nil, failure(ReasonDanglingCommand, "obligation command is not registered")
			}
			selected[command.ID] = command
		}
	}
	return sortedCommands(selected), guards, nil
}

func commandIndex(commands []Command) map[string]Command {
	result := make(map[string]Command, len(commands))
	for _, command := range commands {
		result[command.ID] = command
	}
	return result
}

func bindingIndex(bindings []ObligationBinding) map[string]ObligationBinding {
	result := make(map[string]ObligationBinding, len(bindings))
	for _, binding := range bindings {
		result[binding.ID] = binding
	}
	return result
}

func sortedCommands(commands map[string]Command) []Command {
	result := make([]Command, 0, len(commands))
	for _, command := range commands {
		result = append(result, command)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

type selectedPath struct {
	command    Command
	obligation string
	guard      bool
}

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
		for _, candidate := range binding.CommandIDs {
			if candidate == commandID {
				return binding.ID
			}
		}
	}
	return ""
}

func fillSelection(result PlanResult, selected []selectedPath, frontier workfrontier.Result) PlanResult {
	workByCommand := map[string]string{}
	for i, pathID := range frontier.SelectedIDs {
		if i < len(frontier.WorkIDs) {
			workByCommand[pathID] = frontier.WorkIDs[i]
		}
	}
	for _, entry := range selected {
		if entry.guard {
			result.SelectedGuardCommandIDs = append(result.SelectedGuardCommandIDs, entry.command.ID)
		} else {
			result.SelectedCommandIDs = append(result.SelectedCommandIDs, entry.command.ID)
		}
		result.SelectedWorkIDs = append(result.SelectedWorkIDs, workByCommand[entry.command.ID])
	}
	return result
}
