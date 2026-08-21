package selectiveci

import "github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"

func reverseImpactNodes(nodes []impactgraph.Node) []impactgraph.Node {
	result := append([]impactgraph.Node{}, nodes...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseImpactEdges(edges []impactgraph.Edge) []impactgraph.Edge {
	result := append([]impactgraph.Edge{}, edges...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseObligationsForCoverage(values []ObligationBinding) []ObligationBinding {
	result := append([]ObligationBinding{}, values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseCommandsForCoverage(values []Command) []Command {
	result := append([]Command{}, values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
