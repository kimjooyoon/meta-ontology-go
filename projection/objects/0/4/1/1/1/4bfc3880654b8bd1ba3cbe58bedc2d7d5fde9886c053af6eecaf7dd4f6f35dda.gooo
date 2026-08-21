package selectiveci

import (
	"strings"
)

func indexCommands(items []Command) (map[string]Command, Reason) {
	commands := make(map[string]Command, len(items))
	for _, command := range items {
		if command.ID == "" {
			return nil, invalidGraphReason
		}
		if _, exists := commands[command.ID]; exists {
			return nil, duplicateIDReason
		}
		if len(command.Argv) == 0 {
			return nil, emptyArgvReason
		}
		for _, arg := range command.Argv {
			if strings.IndexByte(arg, 0) >= 0 {
				return nil, nulArgvReason
			}
		}
		commands[command.ID] = command
	}
	return commands, ""
}
func indexEdges(edges []Edge, commands map[string]Command) (map[string][]string, Reason) {
	adjacency := make(map[string][]string, len(commands))
	for _, edge := range edges {
		if _, exists := commands[edge.From]; !exists {
			return nil, danglingEdgeReason
		}
		if _, exists := commands[edge.To]; !exists {
			return nil, danglingEdgeReason
		}
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	return adjacency, ""
}
func indexPaths(items []PathEvidence) (map[string]PathEvidence, Reason) {
	paths := make(map[string]PathEvidence, len(items))
	for _, path := range items {
		if path.Path == "" {
			return nil, missingPathReason
		}
		if _, exists := paths[path.Path]; exists {
			return nil, ambiguousReason
		}
		paths[path.Path] = path
	}
	return paths, ""
}
