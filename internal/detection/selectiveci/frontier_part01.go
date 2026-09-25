package selectiveci

import (
	"sort"
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
