package workfrontier

import (
	"sort"
)

func populateR4Reachability(graph *r4Graph, paths []RepairPath, roots []string) {
	adjacency := make(map[string][]string, len(graph.nodes))
	for _, node := range graph.nodes {
		adjacency[node] = nil
	}
	for _, edge := range graph.edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	for node := range adjacency {
		adjacency[node] = sortedUnique(adjacency[node])
	}
	reachable := reachableFrom(roots, adjacency)
	graph.reachableNodes = sortedKeys(reachable)
	for _, edge := range graph.edges {
		if _, fromOK := reachable[edge.From]; fromOK {
			if _, toOK := reachable[edge.To]; toOK {
				graph.reachableEdges = append(graph.reachableEdges, edge)
			}
		}
	}
	for _, path := range paths {
		if _, ok := reachable[path.ObligationID]; ok {
			graph.reachablePaths = append(graph.reachablePaths, path)
		}
	}
	sort.Slice(graph.reachablePaths, func(i, j int) bool {
		return r4PathKey(graph.reachablePaths[i]) < r4PathKey(graph.reachablePaths[j])
	})
}
func sealR4Graph(graph *r4Graph, input R4Input, roots []string, edges []r4Edge) {
	graph.components = deriveR4Components(graph.reachableNodes, graph.reachableEdges)
	graph.condensationEdges = deriveR4Condensation(graph.components, graph.reachableEdges)
	graph.graphDigest = digestR4(r4GraphPayload{
		SchemaVersion: R4SchemaVersion,
		Roots:         roots,
		Nodes:         graph.reachableNodes,
		Edges:         graph.reachableEdges,
	})
	componentDigests := make([]string, 0, len(graph.components))
	for _, component := range graph.components {
		componentDigests = append(componentDigests, component.Digest)
	}
	sort.Strings(componentDigests)
	graph.sccDigest = digestR4(r4SCCDigestPayload{SchemaVersion: R4SchemaVersion, Components: componentDigests})
	graph.condensationDigest = digestR4(r4CondensationPayload{
		SchemaVersion: R4SchemaVersion,
		Components:    componentDigests,
		Edges:         graph.condensationEdges,
	})
	graph.ruleDigest = digestR4(r4RuleDigestPayload{SchemaVersion: R4SchemaVersion, Rules: normalizeR4Rules(input.Rules)})
	graph.receipt = r4Receipt(graph, edges)
}
func r4Receipt(graph *r4Graph, edges []r4Edge) R4WorkReceipt {
	cyclic := countCyclicR4Components(graph.components)
	return R4WorkReceipt{
		GraphNodes: uint64(len(graph.nodes)), GraphEdges: uint64(len(edges)),
		ReachableNodes: uint64(len(graph.reachableNodes)), ReachableEdges: uint64(len(graph.reachableEdges)),
		SCCs: uint64(len(graph.components)), CyclicSCCs: uint64(cyclic),
		CondensationEdges: uint64(len(graph.condensationEdges)), RuleChecks: uint64(len(graph.components)),
		IterationChecks: uint64(cyclic), PathChecks: uint64(len(graph.reachablePaths)),
	}
}
