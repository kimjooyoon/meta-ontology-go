package selectiveci

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
