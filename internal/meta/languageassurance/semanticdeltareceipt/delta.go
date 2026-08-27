package semanticdeltareceipt

import (
	"bytes"
	"sort"
)

func textualDelta(before, after []byte) TextualDelta {
	changedBytes := 0
	for index := 0; index < len(before) && index < len(after); index++ {
		if before[index] != after[index] {
			changedBytes++
		}
	}
	if len(before) > len(after) {
		changedBytes += len(before) - len(after)
	} else {
		changedBytes += len(after) - len(before)
	}
	decision := RawFixedPoint
	if !bytes.Equal(before, after) {
		decision = RawChanged
	}
	return TextualDelta{Changed: decision == RawChanged, Decision: decision, BeforeBytes: len(before), AfterBytes: len(after), PositionalByteMismatches: changedBytes, BeforeDigest: digestBytes(before), AfterDigest: digestBytes(after)}
}

func structuralDelta(before, after projectedSource) StructuralDelta {
	leftNodes, rightNodes := nodeMap(before.nodes), nodeMap(after.nodes)
	leftFacts, rightFacts := factMap(before.facts), factMap(after.facts)
	result := StructuralDelta{Status: "KNOWN"}
	for key, node := range rightNodes {
		if _, ok := leftNodes[key]; !ok {
			result.AddedNodes = append(result.AddedNodes, node)
		}
	}
	for key, node := range leftNodes {
		if _, ok := rightNodes[key]; !ok {
			result.RemovedNodes = append(result.RemovedNodes, node)
		}
	}
	for key, fact := range rightFacts {
		if _, ok := leftFacts[key]; !ok {
			result.AddedFacts = append(result.AddedFacts, fact)
		}
	}
	for key, fact := range leftFacts {
		if _, ok := rightFacts[key]; !ok {
			result.RemovedFacts = append(result.RemovedFacts, fact)
		}
	}
	sort.Slice(result.AddedNodes, func(i, j int) bool { return result.AddedNodes[i].ID < result.AddedNodes[j].ID })
	sort.Slice(result.RemovedNodes, func(i, j int) bool { return result.RemovedNodes[i].ID < result.RemovedNodes[j].ID })
	sort.Slice(result.AddedFacts, func(i, j int) bool { return factLess(result.AddedFacts[i], result.AddedFacts[j]) })
	sort.Slice(result.RemovedFacts, func(i, j int) bool { return factLess(result.RemovedFacts[i], result.RemovedFacts[j]) })
	return result
}

func nodeMap(values []Node) map[string]Node {
	result := make(map[string]Node, len(values))
	for _, value := range values {
		result[value.ID+"\x00"+value.Kind] = value
	}
	return result
}

func factMap(values []Fact) map[string]Fact {
	result := make(map[string]Fact, len(values))
	for _, value := range values {
		result[value.Subject+"\x00"+value.Predicate+"\x00"+value.Object] = value
	}
	return result
}
