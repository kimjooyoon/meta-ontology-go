package selectiveci

import (
	"math"
	"sort"
	"strings"
)

type graphIndex struct {
	commands  map[string]Command
	adjacency map[string][]string
}

type changeSelection struct {
	owners       map[string]struct{}
	changedPaths map[string]struct{}
}

// Evaluate applies the reference rules. Any failure returns a full-suite
// fallback with no partial selection or resource totals.
func Evaluate(c Case) Result {
	digest := CanonicalDigest(c)
	if reason := validateEvidence(c); reason != "" {
		return fallbackResult(digest, reason)
	}
	graph, reason := indexGraph(c.Graph)
	if reason != "" {
		return fallbackResult(digest, reason)
	}
	paths, reason := indexPaths(c.Evidence.Paths)
	if reason != "" {
		return fallbackResult(digest, reason)
	}
	selection, reason := selectChangedOwners(c.Evidence, graph.commands, paths)
	if reason != "" {
		return fallbackResult(digest, reason)
	}
	expandSelection(selection.owners, graph.adjacency)
	return buildResult(digest, selection, graph.commands)
}

func fallbackResult(digest string, reason Reason) Result {
	return Result{
		Decision:        FullSuiteFallback,
		Reason:          reason,
		CommandIDs:      []string{},
		Argv:            map[string][]string{},
		CanonicalDigest: digest,
	}
}

func validateEvidence(c Case) Reason {
	if !c.Evidence.Complete {
		return incompleteReason
	}
	if c.Evidence.GlobalGuards {
		return globalGuardReason
	}
	if c.Graph.SnapshotID == "" || c.Evidence.SnapshotID == "" || c.Graph.SnapshotID != c.Evidence.SnapshotID {
		return snapshotReason
	}
	return ""
}

func indexGraph(graph Graph) (graphIndex, Reason) {
	commands, reason := indexCommands(graph.Commands)
	if reason != "" {
		return graphIndex{}, reason
	}
	if len(commands) == 0 || len(graph.Roots) == 0 {
		return graphIndex{}, invalidGraphReason
	}
	for _, root := range graph.Roots {
		if _, exists := commands[root]; !exists {
			return graphIndex{}, danglingCmdReason
		}
	}
	adjacency, reason := indexEdges(graph.Edges, commands)
	if reason != "" {
		return graphIndex{}, reason
	}
	return graphIndex{commands: commands, adjacency: adjacency}, ""
}

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

func selectChangedOwners(evidence Evidence, commands map[string]Command, paths map[string]PathEvidence) (changeSelection, Reason) {
	if len(evidence.Changes) == 0 {
		return changeSelection{}, noChangesReason
	}
	selection := changeSelection{
		owners:       make(map[string]struct{}),
		changedPaths: make(map[string]struct{}, len(evidence.Changes)),
	}
	for _, change := range evidence.Changes {
		path, reason := validateChange(change, paths)
		if reason != "" {
			return changeSelection{}, reason
		}
		selection.changedPaths[change.Path] = struct{}{}
		for _, owner := range path.Owners {
			if _, exists := commands[owner]; !exists {
				return changeSelection{}, danglingCmdReason
			}
			selection.owners[owner] = struct{}{}
		}
	}
	return selection, ""
}

func validateChange(change PathChange, paths map[string]PathEvidence) (PathEvidence, Reason) {
	path, exists := paths[change.Path]
	if !exists {
		return PathEvidence{}, missingPathReason
	}
	if path.Stale {
		return PathEvidence{}, staleReason
	}
	if !path.Connected {
		return PathEvidence{}, disconnectedReason
	}
	if path.Ambiguous || len(path.Owners) > 1 && change.Kind != ChangeDelete {
		return PathEvidence{}, ambiguousReason
	}
	if path.Authority != AuthorityAuthoritative {
		return PathEvidence{}, nonAuthorityReason
	}
	if len(path.Owners) == 0 {
		return PathEvidence{}, unknownPathReason
	}
	if !changeMatchesEvidence(change, path) {
		return PathEvidence{}, blobMismatchReason
	}
	return path, ""
}

func expandSelection(selected map[string]struct{}, adjacency map[string][]string) {
	queue := make([]string, 0, len(selected))
	for id := range selected {
		queue = append(queue, id)
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, downstream := range adjacency[id] {
			if _, exists := selected[downstream]; exists {
				continue
			}
			selected[downstream] = struct{}{}
			queue = append(queue, downstream)
		}
	}
}

func buildResult(digest string, selection changeSelection, commands map[string]Command) Result {
	ids := make([]string, 0, len(selection.owners))
	for id := range selection.owners {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	argv := make(map[string][]string, len(ids))
	var cpu, memory uint64
	for _, id := range ids {
		command := commands[id]
		if math.MaxUint64-cpu < command.CPUUnits {
			return fallbackResult(digest, cpuOverflowReason)
		}
		cpu += command.CPUUnits
		if math.MaxUint64-memory < command.MemoryCeiling {
			return fallbackResult(digest, memoryOverflowReason)
		}
		memory += command.MemoryCeiling
		argv[id] = append([]string(nil), command.Argv...)
	}
	return Result{
		Decision:        Selective,
		Reason:          completeReason,
		CommandIDs:      ids,
		Argv:            argv,
		CPUUnits:        cpu,
		MemoryCeiling:   memory,
		PathCount:       len(selection.changedPaths),
		CanonicalDigest: digest,
	}
}

func changeMatchesEvidence(change PathChange, path PathEvidence) bool {
	switch change.Kind {
	case ChangeDelete:
		return !path.Present && change.Blob == ""
	case ChangeAdd, ChangeModify, ChangeRelocate:
		return path.Present && change.Blob != "" && change.Blob == path.Blob
	default:
		return false
	}
}
