package roundtrip

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func diffFacts(delta *Delta, before, after []semantic.Fact) {
	left, right := factMap(before), factMap(after)
	for key, fact := range left {
		other, exists := right[key]
		if !exists || !sameFact(fact, other) {
			delta.RemovedFacts = append(delta.RemovedFacts, fact)
		}
	}
	for key, fact := range right {
		other, exists := left[key]
		if !exists || !sameFact(fact, other) {
			delta.AddedFacts = append(delta.AddedFacts, fact)
		}
	}
	sort.Slice(delta.AddedFacts, func(i, j int) bool { return factIdentity(delta.AddedFacts[i]) < factIdentity(delta.AddedFacts[j]) })
	sort.Slice(delta.RemovedFacts, func(i, j int) bool { return factIdentity(delta.RemovedFacts[i]) < factIdentity(delta.RemovedFacts[j]) })
}
func nodeMap(nodes []semantic.Node) map[semantic.ID]semantic.Node {
	result := make(map[semantic.ID]semantic.Node, len(nodes))
	for _, node := range nodes {
		result[node.ID] = node
	}
	return result
}
func factMap(facts []semantic.Fact) map[semantic.FactKey]semantic.Fact {
	result := make(map[semantic.FactKey]semantic.Fact, len(facts))
	for _, fact := range facts {
		result[fact.Key()] = fact
	}
	return result
}
func sameNode(left, right semantic.Node) bool {
	return left.SemanticCanonical() == right.SemanticCanonical()
}
func sameFact(left, right semantic.Fact) bool {
	return left.SemanticCanonical() == right.SemanticCanonical()
}
func factIdentity(fact semantic.Fact) string {
	key := fact.Key()
	return key.Subject.String() + "\x00" + key.Predicate.String() + "\x00" + key.Object.String()
}
func touchedIDs(delta Delta) []semantic.ID {
	values := make(map[semantic.ID]struct{})
	for _, node := range append(delta.AddedNodes, delta.RemovedNodes...) {
		values[node.ID] = struct{}{}
	}
	for _, fact := range append(delta.AddedFacts, delta.RemovedFacts...) {
		values[fact.Subject] = struct{}{}
		values[fact.Object] = struct{}{}
	}
	return sortedIDs(values)
}
