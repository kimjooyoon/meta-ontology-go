package selectiveci

import (
	"sort"
)

func canonicalGraph(graph Graph) Graph {
	graph.Commands = append([]Command(nil), graph.Commands...)
	graph.Edges = append([]Edge(nil), graph.Edges...)
	graph.Roots = append([]string(nil), graph.Roots...)
	sort.Slice(graph.Commands, func(i, j int) bool { return graph.Commands[i].ID < graph.Commands[j].ID })
	sort.Slice(graph.Edges, func(i, j int) bool {
		if graph.Edges[i].From != graph.Edges[j].From {
			return graph.Edges[i].From < graph.Edges[j].From
		}
		return graph.Edges[i].To < graph.Edges[j].To
	})
	sort.Strings(graph.Roots)
	return graph
}
func canonicalEvidence(evidence Evidence) Evidence {
	evidence.Paths = append([]PathEvidence(nil), evidence.Paths...)
	evidence.Changes = append([]PathChange(nil), evidence.Changes...)
	for i := range evidence.Paths {
		evidence.Paths[i].Owners = append([]string(nil), evidence.Paths[i].Owners...)
		sort.Strings(evidence.Paths[i].Owners)
	}
	sort.Slice(evidence.Paths, func(i, j int) bool { return evidence.Paths[i].Path < evidence.Paths[j].Path })
	sort.Slice(evidence.Changes, func(i, j int) bool {
		if evidence.Changes[i].Path != evidence.Changes[j].Path {
			return evidence.Changes[i].Path < evidence.Changes[j].Path
		}
		if evidence.Changes[i].Kind != evidence.Changes[j].Kind {
			return evidence.Changes[i].Kind < evidence.Changes[j].Kind
		}
		return evidence.Changes[i].Blob < evidence.Changes[j].Blob
	})
	return evidence
}
