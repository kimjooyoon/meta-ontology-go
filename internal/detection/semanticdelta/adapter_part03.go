package semanticdelta

import (
	"fmt"
)

// DiffSnapshots computes the semantic delta between two adapter-neutral
// snapshots. Node identity includes the immutable kind; a kind change is a
// remove followed by an add for the same stable ID.
func DiffSnapshots(before, after Snapshot) (Delta, error) {
	left, err := before.Normalized()
	if err != nil {
		return Delta{}, fmt.Errorf("normalize before snapshot: %w", err)
	}
	right, err := after.Normalized()
	if err != nil {
		return Delta{}, fmt.Errorf("normalize after snapshot: %w", err)
	}
	delta := Delta{}
	leftNodes := nodeMap(left.Nodes)
	rightNodes := nodeMap(right.Nodes)
	for id, node := range leftNodes {
		if other, exists := rightNodes[id]; !exists || other != node {
			delta.RemovedNodes = append(delta.RemovedNodes, node)
		}
	}
	for id, node := range rightNodes {
		if other, exists := leftNodes[id]; !exists || other != node {
			delta.AddedNodes = append(delta.AddedNodes, node)
		}
	}
	leftFacts := factMap(left.Facts)
	rightFacts := factMap(right.Facts)
	for key, fact := range leftFacts {
		if _, exists := rightFacts[key]; !exists {
			delta.RemovedFacts = append(delta.RemovedFacts, fact)
		}
	}
	for key, fact := range rightFacts {
		if _, exists := leftFacts[key]; !exists {
			delta.AddedFacts = append(delta.AddedFacts, fact)
		}
	}
	return delta.Normalized()
}
func nodeMap(nodes []Node) map[string]Node {
	result := make(map[string]Node, len(nodes))
	for _, node := range nodes {
		result[node.ID] = node
	}
	return result
}
func factMap(facts []Fact) map[factIdentity]Fact {
	result := make(map[factIdentity]Fact, len(facts))
	for _, fact := range facts {
		result[factIdentityOf(fact)] = fact
	}
	return result
}
